$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$go = 'C:\Program Files\Go\bin\go.exe'
$dotnet = (Get-Command dotnet).Source
$run = Join-Path $env:TEMP ('conveyance-security-interop-' + [guid]::NewGuid().ToString('N'))
$goBin = Join-Path $run 'security-interop-go.exe'
$dotnetOut = Join-Path $run 'dotnet'
$serverOut = Join-Path $run 'server.stdout.log'
$serverErr = Join-Path $run 'server.stderr.log'
$server = $null
$registered = $false
$unknown = $false
$mTlsStatus = 'NOT-RUN'
$cleanupStatus = 'NOT-RUN'
$diagnostic = $null
$diagnosticReopen = 'NOT-RUN'

New-Item -ItemType Directory -Force $run, $dotnetOut | Out-Null
try {
    if (!(Test-Path $go)) { throw "expected Go toolchain not found: $go" }
    $env:GOCACHE = Join-Path $run 'go-cache'
    $env:GOPATH = Join-Path $run 'go-path'
    $env:GOMODCACHE = Join-Path $run 'go-mod'
    & $go version
    & $go -C (Join-Path $root 'go') build -buildvcs=false -o $goBin .
    if ($LASTEXITCODE -ne 0) { throw 'Go spike build failed' }

    & $dotnet restore (Join-Path $root 'windows\SecurityInterop.csproj') -p:RestoreLockedMode=false
    & $dotnet build (Join-Path $root 'windows\SecurityInterop.csproj') -c Release -o $dotnetOut -p:RestoreLockedMode=false
    if ($LASTEXITCODE -ne 0) { throw '.NET spike build failed' }
    $dll = Join-Path $dotnetOut 'SecurityInterop.dll'

    $aesGo = (& $goBin aes-fixture | Out-String | ConvertFrom-Json)
    $hpkeGo = (& $goBin hpke | Out-String | ConvertFrom-Json)
    $vectorGo = (& $goBin vector | Out-String | ConvertFrom-Json)
    $aesTamperGo = (& $goBin aes-tamper | Out-String | ConvertFrom-Json)
    if ($hpkeGo.round_trip -ne 'PASS' -or $hpkeGo.tamper -ne 'PASS') { throw 'Go HPKE checks failed' }
    if ($vectorGo.selected_vector_material -ne 'PASS' -or $vectorGo.aad_round_trip -ne 'PASS') { throw 'Go vector checks failed' }
    if ($aesTamperGo.result -ne 'PASS' -or $aesTamperGo.passed -ne 10) { throw 'Go AES tamper checks failed' }

    & $dotnet $dll aes-fixture $run | Out-Null
    & $dotnet $dll aes-tamper | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'Windows AES tamper checks failed' }
    $aesTamperWindows = (& $dotnet $dll aes-tamper | Out-String | ConvertFrom-Json)
    if ($aesTamperWindows.result -ne 'PASS' -or $aesTamperWindows.passed -ne 10) { throw 'Windows AES tamper count failed' }
    $aesWindows = Get-Content (Join-Path $run 'windows-aes.json') -Raw | ConvertFrom-Json
    foreach ($field in @('nonce','ciphertext','tag','aad','plaintext')) {
        if ($aesGo.$field -cne $aesWindows.$field) { throw "AES fixture mismatch: $field" }
    }
    $canonical = Get-Content (Join-Path $root 'testdata\aes-gcm-project-fixture.json') -Raw | ConvertFrom-Json
    foreach ($field in @('nonce','ciphertext','tag','aad','plaintext')) {
        if ($aesGo.$field -cne $canonical.$field -or $aesWindows.$field -cne $canonical.$field) { throw "canonical AES fixture mismatch: $field" }
    }
    $aesGo | ConvertTo-Json | Set-Content (Join-Path $run 'go-aes.json')
    & $goBin aes-open -input (Join-Path $run 'windows-aes.json') | Out-Null
    & $dotnet $dll aes-open (Join-Path $run 'go-aes.json') | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'AES cross-runtime open failed' }

    & $dotnet $dll diagnostic $run | Out-Null
    if (Test-Path (Join-Path $run 'diagnostic.json')) { $diagnostic = Get-Content (Join-Path $run 'diagnostic.json') -Raw | ConvertFrom-Json }
    $diagnosticReopen = (& $dotnet $dll diagnostic-reopen $run | Out-String).Trim()
    & $goBin make-server -cert (Join-Path $run 'server.pem') -key (Join-Path $run 'server.key')
    & $dotnet $dll create $run registered
    if ($LASTEXITCODE -eq 0) {
        $registered = $true
        & $dotnet $dll create $run unknown
        $unknown = ($LASTEXITCODE -eq 0)
        $server = Start-Process -FilePath $goBin -ArgumentList @('serve','-cert',(Join-Path $run 'server.pem'),'-key',(Join-Path $run 'server.key'),'-client-cert',(Join-Path $run 'registered.cer')) -RedirectStandardOutput $serverOut -RedirectStandardError $serverErr -PassThru -WindowStyle Hidden
        $ready = $null
        for ($n = 0; $n -lt 50 -and !$ready; $n++) { Start-Sleep -Milliseconds 100; if (Test-Path $serverOut) { $ready = Select-String -Path $serverOut -Pattern '^READY ' | Select-Object -Last 1 } }
        if (!$ready) { throw 'Go TLS server did not become ready' }
        $address = ($ready.Line -split ' ')[1]
        $url = "https://$address/spike/mTLS"
        & $dotnet $dll request-no-cert (Join-Path $run 'server.pem.der') $url
        $noCert = $LASTEXITCODE
        & $dotnet $dll request $run registered (Join-Path $run 'server.pem.der') $url
        $positive = $LASTEXITCODE
        & $dotnet $dll request $run unknown (Join-Path $run 'server.pem.der') $url
        $unknownResult = $LASTEXITCODE
        & $dotnet $dll request $run registered (Join-Path $run 'server.pem.der') $url '00000000-0000-0000-0000-000000000999'
        $mismatch = $LASTEXITCODE
        $mTlsStatus = "no_certificate=$noCert; registered=$positive; unknown=$unknownResult; mismatched_installation_ref=$mismatch; tls=TLS1.3; unknown_credential_created=$unknown"
    } else {
        $mTlsStatus = 'BLOCKED-MTLS-WINDOWS-ENVIRONMENT: persisted credential creation unavailable'
    }

    $stage = 'BLOCKED-HPKE'
    $summary = [ordered]@{
        stage = $stage; go = (& $go version); dotnet = (& $dotnet --version)
        go_hpke = $hpkeGo; go_vector = $vectorGo; aes = 'AES-WINDOWS-INTEROP-PASS'; go_aes_tamper = $aesTamperGo; windows_aes_tamper = $aesTamperWindows; canonical_fixture = 'PASS'; mtls = $mTlsStatus
        cng_diagnostic = $diagnostic; cng_reopen = $diagnosticReopen
        hpke_candidate = 'BLOCKED: no acceptable managed C# RFC 9180 candidate identified'
        real_iphone = 'OUTSTANDING'; cleanup = 'finally block removes run-scoped directory'
    }
    $summary | ConvertTo-Json -Depth 8 | Set-Content (Join-Path $run 'summary.json')
    Get-Content (Join-Path $run 'summary.json')
    Write-Output "STAGE=$stage"
}
finally {
    if ($server -and !$server.HasExited) { Stop-Process -Id $server.Id -Force -ErrorAction SilentlyContinue }
    if ($registered -and (Test-Path (Join-Path $run 'registered.json'))) { & $dotnet (Join-Path $dotnetOut 'SecurityInterop.dll') cleanup $run registered | Out-Null }
    if ($unknown -and (Test-Path (Join-Path $run 'unknown.json'))) { & $dotnet (Join-Path $dotnetOut 'SecurityInterop.dll') cleanup $run unknown | Out-Null }
    & $dotnet (Join-Path $dotnetOut 'SecurityInterop.dll') diagnostic-cleanup $run | Out-Null
    $cleanupStatus = 'attempted; run-scoped directory removed'
    Remove-Item -LiteralPath $run -Recurse -Force -ErrorAction SilentlyContinue
}

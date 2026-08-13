[CmdletBinding()]
param(
    [string]$SourceRoot = '',
    [switch]$DownloadOfficialSource,
    [switch]$RunBridgeProof
)

$ErrorActionPreference = 'Stop'
$officialUrl = 'https://github.com/openssl/openssl/releases/download/openssl-3.5.7/openssl-3.5.7.tar.gz'
$checksumUrl = 'https://github.com/openssl/openssl/releases/download/openssl-3.5.7/openssl-3.5.7.tar.gz.sha256'
$expected = 'a8c0d28a529ca480f9f36cf5792e2cd21984552a3c8e4aa11a24aa31aeac98e8'

if ([string]::IsNullOrWhiteSpace($SourceRoot)) {
    $SourceRoot = Join-Path ([System.IO.Path]::GetTempPath()) ('conveyance-openssl-' + [guid]::NewGuid().ToString('N'))
}
$archive = Join-Path $SourceRoot 'openssl-3.5.7.tar.gz'
$source = Join-Path $SourceRoot 'openssl-3.5.7'

New-Item -ItemType Directory -Force -Path $SourceRoot | Out-Null
if ($DownloadOfficialSource) {
    Invoke-WebRequest -Uri $officialUrl -OutFile $archive
    $checksumFile = Join-Path $SourceRoot 'openssl-3.5.7.tar.gz.sha256'
    Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumFile
    $published = Get-Content -LiteralPath $checksumFile -Raw
    if ([string]::IsNullOrWhiteSpace($published)) { throw 'Official checksum file was empty' }
}
if (-not (Test-Path -LiteralPath $archive)) { throw "Missing official archive: $archive" }
$actual = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
if ($actual -ne $expected) { throw "SHA-256 mismatch: $actual" }
Write-Output "OpenSSL 3.5.7 archive SHA-256: $actual (PASS)"

if (-not (Test-Path -LiteralPath (Join-Path $source 'include\openssl\hpke.h'))) {
    tar -xzf $archive -C $SourceRoot
}
if (-not (Test-Path -LiteralPath (Join-Path $source 'include\openssl\hpke.h'))) {
    throw 'OpenSSL 3.5.7 source extraction did not produce include\openssl\hpke.h'
}

$nasm = Get-Command nasm.exe -ErrorAction SilentlyContinue
if ($null -eq $nasm -and (Test-Path -LiteralPath 'C:\Program Files\NASM\nasm.exe')) {
    $nasm = Get-Item -LiteralPath 'C:\Program Files\NASM\nasm.exe'
}
if ($null -eq $nasm) {
    Write-Output 'Stage: BLOCKED-OPENSSL-BUILD-ENVIRONMENT'
    Write-Output 'Missing prerequisite: NASM (not installed or discoverable)'
    exit 0
}

Write-Output "NASM: $($nasm.FullName)"
$vs = 'C:\Program Files\Microsoft Visual Studio\18\Community\Common7\Tools\VsDevCmd.bat'
$perl = 'C:\Program Files\MATLAB\R2023a\sys\perl\win32\bin\perl.exe'
if (-not (Test-Path -LiteralPath (Join-Path $source 'libcrypto.lib'))) {
    cmd /c "set `"PATH=C:\Program Files\NASM;%PATH%`" && call `"$vs`" -arch=x64 && cd /d `"$source`" && `"$perl`" Configure VC-WIN64A no-shared no-tests && nmake"
    if ($LASTEXITCODE -ne 0) { throw 'OpenSSL Configure/nmake failed' }
}
$nativeRoot = Join-Path $PSScriptRoot 'openssl'
$output = Join-Path $SourceRoot 'shim'
& (Join-Path $nativeRoot 'build-windows.ps1') -OpenSslRoot $source -OutputDirectory $output
if ($LASTEXITCODE -ne 0) { throw 'Native shim build failed' }
Write-Output 'OpenSSL build and native shim: PASS'
if ($RunBridgeProof) {
    $project = Join-Path $PSScriptRoot '..\windows\HpkeBridge.csproj'
    dotnet build $project --nologo
    if ($LASTEXITCODE -ne 0) { throw '.NET HPKE bridge build failed' }
    $managed = Join-Path (Split-Path $project) 'bin\Debug\net10.0-windows\HpkeBridge.dll'
    & dotnet $managed (Join-Path $output 'conveyance_hpke.dll') proof
    if ($LASTEXITCODE -ne 0) { throw 'Native/.NET proof failed' }

    $go = (Get-Command go.exe -ErrorAction Stop).Source
    $goDir = Join-Path $PSScriptRoot '..\go'
    $env:GOCACHE = Join-Path $SourceRoot 'go-cache'
    $env:GOMODCACHE = Join-Path $SourceRoot 'go-mod'
    $env:GOPATH = Join-Path $SourceRoot 'go-path'
    $goBinary = Join-Path $SourceRoot 'security-interop-go.exe'
    & $go -C $goDir build -buildvcs=false -o $goBinary .
    if ($LASTEXITCODE -ne 0) { throw 'Go spike build failed' }

    $goWrapperPath = Join-Path $SourceRoot 'go-wrapper.json'
    $goGrantPath = Join-Path $SourceRoot 'go-grant.json'
    & $goBinary hpke | Set-Content -LiteralPath $goWrapperPath
    if ($LASTEXITCODE -ne 0) { throw 'Go grant generation failed' }
    $goWrapper = Get-Content -LiteralPath $goWrapperPath -Raw | ConvertFrom-Json
    [pscustomobject]@{ grant = $goWrapper.grant; recipient_private_key = $goWrapper.recipient_private_key; recipient_public_key = $goWrapper.recipient_public_key } |
        ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $goGrantPath
    & dotnet $managed (Join-Path $output 'conveyance_hpke.dll') open-grant $goGrantPath
    if ($LASTEXITCODE -ne 0) { throw 'Go-to-OpenSSL grant proof failed' }
    & dotnet $managed (Join-Path $output 'conveyance_hpke.dll') grant-tamper $goGrantPath
    if ($LASTEXITCODE -ne 0) { throw 'OpenSSL recipient tamper proof failed' }

    $opensslGrantPath = Join-Path $SourceRoot 'openssl-grant.json'
    & dotnet $managed (Join-Path $output 'conveyance_hpke.dll') grant $opensslGrantPath
    if ($LASTEXITCODE -ne 0) { throw 'OpenSSL grant generation failed' }
    & $goBinary open-grant $opensslGrantPath
    if ($LASTEXITCODE -ne 0) { throw 'OpenSSL-to-Go grant proof failed' }
    & $goBinary grant-tamper $opensslGrantPath
    if ($LASTEXITCODE -ne 0) { throw 'Go recipient tamper proof failed' }
    Write-Output 'Grant interop and exact six-case tamper proof: PASS both directions'
}

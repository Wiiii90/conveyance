[CmdletBinding()]
param(
    [string]$SourceRoot = '',
    [switch]$DownloadOfficialSource
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
    $published = (Invoke-WebRequest -Uri $checksumUrl).Content.Trim()
    $published | Set-Content -LiteralPath (Join-Path $SourceRoot 'openssl-3.5.7.tar.gz.sha256') -NoNewline
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
if ($null -eq $nasm) {
    Write-Output 'Stage: BLOCKED-OPENSSL-BUILD-ENVIRONMENT'
    Write-Output 'Missing prerequisite: NASM (not installed or discoverable)'
    exit 0
}

Write-Output "NASM: $($nasm.Source)"
Write-Output 'All discovery gates passed; native build and bridge harness are not included in this run.'

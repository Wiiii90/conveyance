[CmdletBinding()]
param([Parameter(Mandatory)][string]$OpenSslRoot, [Parameter(Mandatory)][string]$OutputDirectory)
$ErrorActionPreference = 'Stop'
$vs = 'C:\Program Files\Microsoft Visual Studio\18\Community\Common7\Tools\VsDevCmd.bat'
$nasm = 'C:\Program Files\NASM\nasm.exe'
$cl = 'C:\Program Files\Microsoft Visual Studio\18\Community\VC\Tools\MSVC\14.51.36231\bin\Hostx64\x64\cl.exe'
if (!(Test-Path $nasm) -or !(Test-Path $cl) -or !(Test-Path (Join-Path $OpenSslRoot 'libcrypto.lib'))) { throw 'Expected OpenSSL/MSVC/NASM inputs are missing' }
New-Item -ItemType Directory -Force -Path $OutputDirectory | Out-Null
$env:Path = 'C:\Program Files\NASM;' + $env:Path
$include = Join-Path $OpenSslRoot 'include'
$source = Join-Path $PSScriptRoot 'conveyance_hpke.c'
$header = Join-Path $PSScriptRoot 'conveyance_hpke.h'
$object = Join-Path $OutputDirectory 'conveyance_hpke.obj'
$dll = Join-Path $OutputDirectory 'conveyance_hpke.dll'
$lib = Join-Path $OutputDirectory 'conveyance_hpke.lib'
cmd /c "call `"$vs`" -arch=x64 && cd /d `"$PSScriptRoot`" && cl /nologo /LD /O2 /W4 /WX /I`"$include`" /c `"$source`" /Fo`"$object`" && link /DLL /OUT:`"$dll`" /IMPLIB:`"$lib`" `"$object`" `"$(Join-Path $OpenSslRoot 'libcrypto.lib')`" ws2_32.lib gdi32.lib advapi32.lib crypt32.lib user32.lib"
if ($LASTEXITCODE -ne 0) { throw 'Native shim build failed' }
Get-Item $dll,$lib | Select-Object FullName,Length

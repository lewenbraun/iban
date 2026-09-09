param([string]$Tag)
$ErrorActionPreference = 'Stop'
$repo = 'lewenbraun/iban'
$gh = (Get-Command gh -ErrorAction Stop).Source
$installDir = Join-Path $env:LOCALAPPDATA 'Iban'
$stage = Join-Path ([IO.Path]::GetTempPath()) ('iban-install-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $stage | Out-Null
if (-not $Tag) {
    $Tag = & $gh release view --repo $repo --json tagName --jq .tagName
    if ($LASTEXITCODE -ne 0) { throw 'Cannot read latest release. Run gh auth login.' }
}
if ($Tag -notmatch '^v[0-9]+\.[0-9]+\.[0-9]+$') { throw "Unsupported release tag: $Tag" }
$version = $Tag.Substring(1)
$archive = "iban_${version}_windows_amd64.zip"
$checksums = "iban_${version}_checksums.txt"
& $gh release download $Tag --repo $repo --pattern $archive --pattern $checksums --dir $stage
if ($LASTEXITCODE -ne 0) { throw 'Release download failed; installed version unchanged.' }
$line = @(Get-Content -LiteralPath (Join-Path $stage $checksums) | Where-Object { ($_ -split '\s+')[-1] -eq $archive })
if ($line.Count -ne 1) { throw 'Expected exactly one archive checksum.' }
$expected = ($line[0] -split '\s+')[0]
if ($expected -notmatch '^[a-fA-F0-9]{64}$') { throw 'Invalid SHA256 checksum.' }
if ((Get-FileHash -LiteralPath (Join-Path $stage $archive) -Algorithm SHA256).Hash -ne $expected) { throw 'SHA256 mismatch.' }
$unpacked = Join-Path $stage 'unpacked'
Expand-Archive -LiteralPath (Join-Path $stage $archive) -DestinationPath $unpacked
$newExe = Join-Path $unpacked 'iban.exe'
if (-not (Test-Path -LiteralPath $newExe)) { throw 'Archive contains no iban.exe.' }
& $newExe help | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'Downloaded executable failed its help check.' }
if (-not (Get-Command ffmpeg -ErrorAction SilentlyContinue)) {
    & scoop install ffmpeg
    if ($LASTEXITCODE -ne 0) { throw 'FFmpeg installation failed.' }
}
New-Item -ItemType Directory -Path $installDir -Force | Out-Null
$exe = Join-Path $installDir 'iban.exe'
Get-Process iban -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq $exe } | Stop-Process
$backup = Join-Path $stage 'previous-iban.exe'
if (Test-Path -LiteralPath $exe) { Move-Item -LiteralPath $exe -Destination $backup }
try { Copy-Item -LiteralPath $newExe -Destination $exe }
catch {
    if (Test-Path -LiteralPath $backup) { Copy-Item -LiteralPath $backup -Destination $exe -Force }
    throw
}
$userKey = [Environment]::GetEnvironmentVariable('ELEVENLABS_API_KEY', 'User')
if (-not [string]::IsNullOrWhiteSpace($userKey)) { $env:ELEVENLABS_API_KEY = $userKey }
Start-Process -FilePath $exe -ArgumentList tray -WindowStyle Hidden
Write-Host "Installed $Tag to $installDir; SHA256 verified; tray started."
Write-Host "Recovery files: $stage"

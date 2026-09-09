# Windows installation and updates

Requires PowerShell, Scoop and GitHub CLI signed in to an account with access to this private repository (`gh auth login`). The installer uses that authentication to download releases and installs FFmpeg through Scoop if missing.

Run the same PowerShell command for installation and every update:

```powershell
& { $script = gh api repos/lewenbraun/eban/contents/install.ps1 -H 'Accept: application/vnd.github.raw+json'; if ($LASTEXITCODE -ne 0) { throw 'Cannot download installer' }; & ([scriptblock]::Create(($script -join "`n"))) }
```

This executes the installer maintained in this repository's default branch. It downloads the latest stable release, verifies SHA256 before replacing the executable, retains the previous executable in the printed recovery directory, and restarts the tray. Installation path: `%LOCALAPPDATA%\Eban`. An explicit version can be supplied as `-Tag v0.0.3` when invoking the script block.

Set your own ElevenLabs key once (never commit it):

```powershell
setx ELEVENLABS_API_KEY "YOUR_REAL_KEY"
```

Then run the installation/update command again. It reads the saved user variable and restarts the tray with the key, so reopening PowerShell is unnecessary for this command. A tray already running before `setx` does not receive the new value.

PowerShell ExecutionPolicy is not needed for `eban.exe`. If Scoop requires RemoteSigned, set it only for your user; do not change it for each update.

Press Alt+Space once to start recording; press Alt+Space again to stop, paste and press Enter. Hold V as part of the press for paste-only delivery; hold B for clipboard-only.

# Release maintenance

Pushing a `v*` tag builds and publishes the release, then updates `eban.json` on the default branch using authenticated GitHub CLI downloads. To repair a manifest for an already published release without creating another version:

```powershell
gh workflow run release.yml --repo lewenbraun/eban -f tag=v0.0.3
```

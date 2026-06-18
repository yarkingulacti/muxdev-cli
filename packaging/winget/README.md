# Submit to microsoft/winget-pkgs

Manifests live in `packaging/winget/`. Update versions and SHA256 values after each release, then open a PR to [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs).

Suggested target path in winget-pkgs:

```
manifests/y/ya/yarkingulacti/muxdev/1.0.0/
```

Files to copy:

- `yarkingulacti.muxdev.yaml` (merged) or split installer/locale files
- `yarkingulacti.muxdev.installer.yaml`
- `yarkingulacti.muxdev.locale.en-US.yaml`

Install after merge:

```powershell
winget install yarkingulacti.muxdev
```

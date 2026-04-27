# Free Quick Share Native

[Chinese](README.md) | [English](README.en.md)

`fq-share-native` is the native-side repository for Free Quick Share. It contains:

- the Go Native Messaging host
- the local HTTP file sharing service
- install/uninstall scripts published through GitHub Releases

## Layout

- `go/`: Go source code
- `install_host.*` / `uninstall_host.*`: install and uninstall scripts
- `.github/workflows/release-native.yml`: release workflow

## Prerequisites

- Go 1.22+
- Chrome or Chromium

## Local Build

```bash
cd go
go build -o free-quick-share-host .
```

## Install as Native Host

Linux / macOS:

```bash
curl -fsSL "https://github.com/mingkao92/fq-share-native/releases/latest/download/install_host.sh" | bash -s -- "<extension-id>"
```

Windows (PowerShell):

```powershell
$tmp = Join-Path $env:TEMP 'fq_install_host.ps1'
Invoke-WebRequest -Uri "https://github.com/mingkao92/fq-share-native/releases/latest/download/install_host.ps1" -OutFile $tmp
powershell -ExecutionPolicy Bypass -File $tmp -ExtensionId "<extension-id>"
```

## Uninstall

Linux / macOS:

```bash
curl -fsSL "https://github.com/mingkao92/fq-share-native/releases/latest/download/uninstall_host.sh" | bash
```

Windows (PowerShell):

```powershell
$tmp = Join-Path $env:TEMP 'fq_uninstall_host.ps1'
Invoke-WebRequest -Uri "https://github.com/mingkao92/fq-share-native/releases/latest/download/uninstall_host.ps1" -OutFile $tmp
powershell -ExecutionPolicy Bypass -File $tmp
```

## Features

- bidirectional transfer between phone and computer

## License

MIT, see `LICENSE`.

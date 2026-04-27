# 飞快分享 本机端

[简体中文](README.md) | [English](README.en.md)

`fq-share-native` 是 `Free Quick Share` 的本机端仓库，提供：

- Go 实现的 Native Messaging Host
- 本地 HTTP 文件分享服务
- GitHub Release 安装/卸载脚本

## 目录结构

- `go/`：Go 源码
- `install_host.*` / `uninstall_host.*`：安装/卸载脚本
- `.github/workflows/release-native.yml`：发布 workflow

## 先决条件

- Go 1.22+
- Chrome 或 Chromium

## 本地构建

```bash
cd go
go build -o free-quick-share-host .
```

## 作为 Native Host 安装

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

## 卸载

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

## 功能说明

- 手机与电脑双向传输

## 许可证

MIT，见 `LICENSE`。

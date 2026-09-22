# 鼠标保活（MouseKeeper）

一个仅面向 Windows 10/11 amd64 的轻量桌面工具。启用时会立即移动鼠标，之后按固定周期左右移动 50 像素，以保持系统处于活动状态；不会点击鼠标。

## 功能

- 主窗口、系统托盘菜单或全局快捷键控制启停
- 启用后通过 Windows 电源请求阻止自动息屏和系统休眠，并立即移动一次鼠标，之后按周期移动 50 像素
- 5、15、30、60 秒四档移动间隔，默认 30 秒
- 可配置 `Ctrl`、`Alt`、`Shift` 加 `A-Z`、`0-9` 或 `F1-F11`
- 默认快捷键 `Ctrl+F1`
- 关闭窗口后隐藏到系统托盘；托盘菜单可完全退出
- 保存间隔和快捷键，但每次启动默认停止
- 单实例运行和 1 MiB 日志轮转

## Windows 构建

安装 Go 1.26、MinGW-w64，并确保 `go`、`gcc` 可从 PowerShell 的 `PATH` 找到，然后运行：

```powershell
./scripts/build.ps1
```

输出文件为 `dist/MouseKeeper.exe`。该便携版没有代码签名，Windows SmartScreen 可能提示“未知发布者”。

## 开发验证

```powershell
go test -tags win ./...
go vet -tags win ./...
```

macOS/Linux 可运行平台无关单元测试：

```bash
go test ./...
```

## 使用说明

1. 启动 `MouseKeeper.exe`。
2. 选择移动间隔和快捷键，点击“应用快捷键”。
3. 点击“启动”，或使用全局快捷键切换状态。
4. 关闭主窗口会隐藏到托盘；通过托盘菜单选择“退出”才能完全结束程序。

运行日志位于 `%APPDATA%\MouseKeeper\mousekeeper.log`，超过 1 MiB 后轮转为 `mousekeeper.log.1`。

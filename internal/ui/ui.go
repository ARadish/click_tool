package ui

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"mousekeeper/internal/controller"
	"mousekeeper/internal/keeper"
	"mousekeeper/internal/shortcut"
)

type Controller interface {
	Start()
	Stop()
	SetInterval(time.Duration) error
	SetShortcut(shortcut.Shortcut) error
	State() controller.State
	Close() error
}

type UI struct {
	app         fyne.App
	controller  Controller
	window      fyne.Window
	status      *widget.Label
	toggle      *widget.Button
	interval    *widget.RadioGroup
	hotkeyLabel *widget.Label
	ctrl        *widget.Check
	alt         *widget.Check
	shift       *widget.Check
	key         *widget.Select
	applyHotkey *widget.Button
	errorLabel  *widget.Label
	trayToggle  *fyne.MenuItem
	trayMenu    *fyne.Menu
	closeOnce   sync.Once
}

func New(app fyne.App, control Controller, initialError string) *UI {
	state := control.State()
	view := &UI{app: app, controller: control}
	view.window = app.NewWindow("鼠标保活")
	view.window.Resize(fyne.NewSize(430, 390))
	view.window.SetFixedSize(true)

	view.status = widget.NewLabel("")
	view.status.Alignment = fyne.TextAlignCenter
	view.status.TextStyle = fyne.TextStyle{Bold: true}
	view.toggle = widget.NewButton("", view.toggleKeeper)

	view.interval = widget.NewRadioGroup([]string{"5 秒", "15 秒", "30 秒", "60 秒"}, nil)
	view.interval.Horizontal = true
	view.interval.SetSelected(formatInterval(state.Settings.Interval))
	view.interval.OnChanged = view.changeInterval

	view.hotkeyLabel = widget.NewLabel("当前快捷键：" + state.Settings.Shortcut.String())
	view.ctrl = widget.NewCheck("Ctrl", nil)
	view.alt = widget.NewCheck("Alt", nil)
	view.shift = widget.NewCheck("Shift", nil)
	view.ctrl.SetChecked(state.Settings.Shortcut.Ctrl)
	view.alt.SetChecked(state.Settings.Shortcut.Alt)
	view.shift.SetChecked(state.Settings.Shortcut.Shift)
	view.key = widget.NewSelect(shortcut.SupportedKeys(), nil)
	view.key.SetSelected(state.Settings.Shortcut.Key)
	view.applyHotkey = widget.NewButton("应用快捷键", view.applyShortcut)

	view.errorLabel = widget.NewLabel(initialError)
	view.errorLabel.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		widget.NewCard("状态", "", container.NewVBox(view.status, view.toggle)),
		widget.NewCard("移动间隔", "", view.interval),
		widget.NewCard("全局快捷键", "至少选择 Ctrl / Alt / Shift 中的一项",
			container.NewVBox(
				view.hotkeyLabel,
				container.NewHBox(view.ctrl, view.alt, view.shift, view.key, view.applyHotkey),
			)),
		view.errorLabel,
	)
	view.window.SetContent(container.NewPadded(content))
	view.window.SetCloseIntercept(view.window.Hide)
	view.configureTray()
	view.Update(state.Keeper)
	return view
}

func (u *UI) Run() {
	u.window.Show()
	u.app.Run()
}

func (u *UI) Update(state keeper.State) {
	if state.Running {
		u.status.SetText("运行中")
		u.toggle.SetText("停止")
		if u.trayToggle != nil {
			u.trayToggle.Label = "停用"
		}
	} else {
		u.status.SetText("已停止")
		u.toggle.SetText("启动")
		if u.trayToggle != nil {
			u.trayToggle.Label = "启用"
		}
	}
	if state.LastError != "" {
		u.errorLabel.SetText("错误：" + state.LastError)
	}
	if u.trayMenu != nil {
		u.trayMenu.Refresh()
	}
}

func (u *UI) toggleKeeper() {
	if u.controller.State().Keeper.Running {
		u.controller.Stop()
	} else {
		u.controller.Start()
	}
	u.Update(u.controller.State().Keeper)
}

func (u *UI) changeInterval(label string) {
	secondsText := strings.TrimSpace(strings.TrimSuffix(label, "秒"))
	seconds, err := strconv.Atoi(secondsText)
	if err != nil {
		u.showError(fmt.Errorf("无效的移动间隔"))
		return
	}
	if err := u.controller.SetInterval(time.Duration(seconds) * time.Second); err != nil {
		u.showError(err)
		return
	}
	u.errorLabel.SetText("")
}

func (u *UI) applyShortcut() {
	value, err := shortcut.New(u.ctrl.Checked, u.alt.Checked, u.shift.Checked, u.key.Selected)
	if err != nil {
		u.showError(err)
		return
	}
	if err := u.controller.SetShortcut(value); err != nil {
		u.showError(err)
		return
	}
	u.hotkeyLabel.SetText("当前快捷键：" + value.String())
	u.errorLabel.SetText("")
}

func (u *UI) showError(err error) {
	u.errorLabel.SetText("错误：" + err.Error())
}

func (u *UI) configureTray() {
	desktopApp, ok := u.app.(desktop.App)
	if !ok {
		return
	}
	u.trayToggle = fyne.NewMenuItem("启用", u.toggleKeeper)
	u.trayMenu = fyne.NewMenu("鼠标保活",
		u.trayToggle,
		fyne.NewMenuItem("显示主窗口", func() {
			u.window.Show()
			u.window.RequestFocus()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("退出", u.quit),
	)
	icon := u.app.Icon()
	if icon == nil {
		icon = theme.ComputerIcon()
	}
	desktopApp.SetSystemTrayIcon(icon)
	desktopApp.SetSystemTrayMenu(u.trayMenu)
	desktopApp.SetSystemTrayWindow(u.window)
}

func (u *UI) quit() {
	u.closeOnce.Do(func() {
		if err := u.controller.Close(); err != nil {
			u.showError(err)
		}
		u.app.Quit()
	})
}

func formatInterval(interval time.Duration) string {
	return fmt.Sprintf("%d 秒", int(interval/time.Second))
}

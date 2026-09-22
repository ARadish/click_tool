package ui

import (
	"errors"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"mousekeeper/internal/config"
	"mousekeeper/internal/controller"
	"mousekeeper/internal/keeper"
	"mousekeeper/internal/shortcut"
)

type trayRecorder struct {
	iconCalls int
	menu      *fyne.Menu
	windowSet bool
}

func (t *trayRecorder) SetSystemTrayIcon(fyne.Resource) { t.iconCalls++ }
func (t *trayRecorder) SetSystemTrayMenu(menu *fyne.Menu) {
	t.menu = menu
}
func (t *trayRecorder) SetSystemTrayWindow(fyne.Window) { t.windowSet = true }

func TestConfigureDesktopTrayDefersIconUntilFyneIsReady(t *testing.T) {
	tray := &trayRecorder{}
	menu := fyne.NewMenu("鼠标保活")

	configureDesktopTray(tray, menu, nil)

	if tray.iconCalls != 0 {
		t.Fatalf("icon calls = %d, want 0 before systray is ready", tray.iconCalls)
	}
	if tray.menu != menu {
		t.Fatal("tray menu was not configured")
	}
	if !tray.windowSet {
		t.Fatal("tray window was not configured")
	}
}

type fakeController struct {
	state         controller.State
	startCount    int
	stopCount     int
	interval      time.Duration
	shortcut      shortcut.Shortcut
	shortcutError error
}

func (f *fakeController) Start() { f.startCount++; f.state.Keeper.Running = true }
func (f *fakeController) Stop()  { f.stopCount++; f.state.Keeper.Running = false }
func (f *fakeController) SetInterval(interval time.Duration) error {
	f.interval = interval
	f.state.Keeper.Interval = interval
	return nil
}
func (f *fakeController) SetShortcut(value shortcut.Shortcut) error {
	if f.shortcutError != nil {
		return f.shortcutError
	}
	f.shortcut = value
	f.state.Settings.Shortcut = value
	return nil
}
func (f *fakeController) State() controller.State { return f.state }
func (f *fakeController) Close() error            { return nil }

func defaultController() *fakeController {
	settings := config.DefaultSettings()
	return &fakeController{state: controller.State{
		Keeper:   keeper.State{Interval: settings.Interval},
		Settings: settings,
	}}
}

func TestNewShowsStoppedStateAndDefaults(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	view := New(app, defaultController(), "")

	if view.status.Text != "已停止" {
		t.Fatalf("status = %q, want 已停止", view.status.Text)
	}
	if view.toggle.Text != "启动" {
		t.Fatalf("toggle = %q, want 启动", view.toggle.Text)
	}
	if view.interval.Selected != "30 秒" {
		t.Fatalf("interval = %q, want 30 秒", view.interval.Selected)
	}
	if view.hotkeyLabel.Text != "当前快捷键：Ctrl+F1" {
		t.Fatalf("hotkey label = %q", view.hotkeyLabel.Text)
	}
}

func TestToggleButtonStartsAndStopsController(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	control := defaultController()
	view := New(app, control, "")

	test.Tap(view.toggle)
	if control.startCount != 1 {
		t.Fatalf("start count = %d, want 1", control.startCount)
	}
	view.Update(control.state.Keeper)
	test.Tap(view.toggle)
	if control.stopCount != 1 {
		t.Fatalf("stop count = %d, want 1", control.stopCount)
	}
}

func TestSelectingIntervalUpdatesController(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	control := defaultController()
	view := New(app, control, "")

	view.interval.SetSelected("15 秒")
	if control.interval != 15*time.Second {
		t.Fatalf("interval = %v, want 15s", control.interval)
	}
}

func TestApplyingShortcutShowsRegistrationError(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	control := defaultController()
	control.shortcutError = errors.New("快捷键已被占用")
	view := New(app, control, "")
	view.ctrl.SetChecked(false)
	view.alt.SetChecked(true)
	view.key.SetSelected("K")

	test.Tap(view.applyHotkey)
	if view.errorLabel.Text == "" {
		t.Fatal("error label is empty after registration failure")
	}
	if view.hotkeyLabel.Text != "当前快捷键：Ctrl+F1" {
		t.Fatalf("hotkey changed after failure: %q", view.hotkeyLabel.Text)
	}
}

//go:build windows

package main

import (
	"log"
	"os"
	"path/filepath"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"mousekeeper/internal/config"
	"mousekeeper/internal/controller"
	"mousekeeper/internal/desktop"
	"mousekeeper/internal/keeper"
	applog "mousekeeper/internal/logging"
	"mousekeeper/internal/ui"
)

const instanceName = "Global\\MouseKeeper-4f716fa2-0e9a-4de7-90fe-047a579347c8"

func main() {
	instance, alreadyRunning, err := desktop.AcquireSingleInstance(instanceName)
	if err != nil {
		log.Printf("acquire single instance: %v", err)
		return
	}
	defer instance.Close()
	if alreadyRunning {
		desktop.ShowAlreadyRunningMessage()
		return
	}

	fyneApp := app.NewWithID("com.local.mousekeeper")
	configureLogging()
	store := config.NewStore(fyneApp.Preferences())
	settings := store.Load()

	var view atomic.Pointer[ui.UI]
	mover := desktop.NewRobotMouse()
	engine := keeper.New(mover, settings.Interval, nil, func(state keeper.State) {
		current := view.Load()
		if current == nil {
			return
		}
		fyne.Do(func() { current.Update(state) })
	})
	hotkeys := desktop.NewSystemHotkeys()
	control := controller.New(engine, hotkeys, store, settings)

	initialError := ""
	if err := control.Initialize(); err != nil {
		initialError = "快捷键注册失败：" + err.Error()
		log.Print(initialError)
	}
	mainUI := ui.New(fyneApp, control, initialError)
	view.Store(mainUI)
	defer func() {
		if err := control.Close(); err != nil {
			log.Printf("close application: %v", err)
		}
	}()
	mainUI.Run()
}

func configureLogging() {
	directory, err := os.UserConfigDir()
	if err != nil {
		log.Printf("resolve config directory: %v", err)
		return
	}
	file, err := applog.Open(filepath.Join(directory, "MouseKeeper"))
	if err != nil {
		log.Printf("open log: %v", err)
		return
	}
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
}

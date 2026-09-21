package controller

import (
	"fmt"
	"sync"
	"time"

	"mousekeeper/internal/config"
	"mousekeeper/internal/keeper"
	"mousekeeper/internal/shortcut"
)

type Keeper interface {
	Start()
	Stop()
	Toggle()
	SetInterval(time.Duration) error
	State() keeper.State
	Close()
}

type Hotkeys interface {
	Replace(shortcut.Shortcut, func()) error
	Close() error
}

type SettingsStore interface {
	Save(config.Settings)
}

type State struct {
	Keeper   keeper.State
	Settings config.Settings
}

type Controller struct {
	keeper   Keeper
	hotkeys  Hotkeys
	store    SettingsStore
	mu       sync.RWMutex
	settings config.Settings
}

func New(k Keeper, hotkeys Hotkeys, store SettingsStore, settings config.Settings) *Controller {
	return &Controller{keeper: k, hotkeys: hotkeys, store: store, settings: settings}
}

func (c *Controller) Initialize() error {
	c.mu.RLock()
	value := c.settings.Shortcut
	c.mu.RUnlock()
	if err := c.hotkeys.Replace(value, c.keeper.Toggle); err != nil {
		return fmt.Errorf("register %s: %w", value.String(), err)
	}
	return nil
}

func (c *Controller) Start()  { c.keeper.Start() }
func (c *Controller) Stop()   { c.keeper.Stop() }
func (c *Controller) Toggle() { c.keeper.Toggle() }

func (c *Controller) SetInterval(interval time.Duration) error {
	if !config.ValidInterval(interval) {
		return fmt.Errorf("unsupported interval %s", interval)
	}
	if err := c.keeper.SetInterval(interval); err != nil {
		return err
	}
	c.mu.Lock()
	c.settings.Interval = interval
	settings := c.settings
	c.mu.Unlock()
	c.store.Save(settings)
	return nil
}

func (c *Controller) SetShortcut(value shortcut.Shortcut) error {
	if err := value.Validate(); err != nil {
		return err
	}
	if err := c.hotkeys.Replace(value, c.keeper.Toggle); err != nil {
		return fmt.Errorf("register %s: %w", value.String(), err)
	}
	c.mu.Lock()
	c.settings.Shortcut = value
	settings := c.settings
	c.mu.Unlock()
	c.store.Save(settings)
	return nil
}

func (c *Controller) State() State {
	c.mu.RLock()
	settings := c.settings
	c.mu.RUnlock()
	return State{Keeper: c.keeper.State(), Settings: settings}
}

func (c *Controller) Close() error {
	c.mu.RLock()
	settings := c.settings
	c.mu.RUnlock()
	c.store.Save(settings)
	err := c.hotkeys.Close()
	c.keeper.Close()
	return err
}

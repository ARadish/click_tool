package controller

import (
	"errors"
	"testing"
	"time"

	"mousekeeper/internal/config"
	"mousekeeper/internal/keeper"
	"mousekeeper/internal/shortcut"
)

type fakeKeeper struct {
	state       keeper.State
	toggleCount int
	closed      bool
}

func (f *fakeKeeper) Start()                            { f.state.Running = true }
func (f *fakeKeeper) Stop()                             { f.state.Running = false }
func (f *fakeKeeper) Toggle()                           { f.state.Running = !f.state.Running; f.toggleCount++ }
func (f *fakeKeeper) SetInterval(d time.Duration) error { f.state.Interval = d; return nil }
func (f *fakeKeeper) State() keeper.State               { return f.state }
func (f *fakeKeeper) Close()                            { f.closed = true }

type fakeHotkeys struct {
	shortcut shortcut.Shortcut
	callback func()
	err      error
	closed   bool
	replaces int
}

func (f *fakeHotkeys) Replace(value shortcut.Shortcut, callback func()) error {
	f.replaces++
	if f.err != nil {
		return f.err
	}
	f.shortcut = value
	f.callback = callback
	return nil
}

func TestSetShortcutDoesNotReregisterUnchangedCombination(t *testing.T) {
	k := &fakeKeeper{state: keeper.State{Interval: 30 * time.Second}}
	h := &fakeHotkeys{}
	c := New(k, h, &fakeStore{}, config.DefaultSettings())
	if err := c.Initialize(); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	if err := c.SetShortcut(shortcut.Default()); err != nil {
		t.Fatalf("SetShortcut returned error: %v", err)
	}
	if h.replaces != 1 {
		t.Fatalf("Replace call count = %d, want 1", h.replaces)
	}
}

func (f *fakeHotkeys) Close() error { f.closed = true; return nil }

type fakeStore struct {
	saved []config.Settings
}

func (f *fakeStore) Save(settings config.Settings) {
	f.saved = append(f.saved, settings)
}

func TestInitializeRegistersShortcutThatTogglesKeeper(t *testing.T) {
	k := &fakeKeeper{state: keeper.State{Interval: 30 * time.Second}}
	h := &fakeHotkeys{}
	c := New(k, h, &fakeStore{}, config.DefaultSettings())

	if err := c.Initialize(); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}
	h.callback()

	if k.toggleCount != 1 || !k.state.Running {
		t.Fatalf("hotkey toggle count = %d, running = %v", k.toggleCount, k.state.Running)
	}
}

func TestSetShortcutKeepsOldSettingWhenRegistrationFails(t *testing.T) {
	k := &fakeKeeper{state: keeper.State{Interval: 30 * time.Second}}
	h := &fakeHotkeys{err: errors.New("already registered")}
	store := &fakeStore{}
	c := New(k, h, store, config.DefaultSettings())
	newShortcut, _ := shortcut.Parse("Alt+K")

	if err := c.SetShortcut(newShortcut); err == nil {
		t.Fatal("SetShortcut succeeded, want registration error")
	}
	if got := c.State().Settings.Shortcut.String(); got != "Ctrl+F1" {
		t.Fatalf("shortcut after failure = %q, want Ctrl+F1", got)
	}
	if len(store.saved) != 0 {
		t.Fatalf("saved settings count = %d, want 0", len(store.saved))
	}
}

func TestSetIntervalUpdatesKeeperAndPersists(t *testing.T) {
	k := &fakeKeeper{state: keeper.State{Interval: 30 * time.Second}}
	store := &fakeStore{}
	c := New(k, &fakeHotkeys{}, store, config.DefaultSettings())

	if err := c.SetInterval(15 * time.Second); err != nil {
		t.Fatalf("SetInterval returned error: %v", err)
	}

	if k.state.Interval != 15*time.Second {
		t.Fatalf("keeper interval = %v, want 15s", k.state.Interval)
	}
	if len(store.saved) != 1 || store.saved[0].Interval != 15*time.Second {
		t.Fatalf("saved settings = %#v, want interval 15s", store.saved)
	}
}

func TestSetIntervalRejectsNonPreset(t *testing.T) {
	k := &fakeKeeper{state: keeper.State{Interval: 30 * time.Second}}
	c := New(k, &fakeHotkeys{}, &fakeStore{}, config.DefaultSettings())
	if err := c.SetInterval(17 * time.Second); err == nil {
		t.Fatal("SetInterval(17s) succeeded, want validation error")
	}
}

func TestCloseReleasesHotkeyAndKeeper(t *testing.T) {
	k := &fakeKeeper{state: keeper.State{Interval: 30 * time.Second}}
	h := &fakeHotkeys{}
	c := New(k, h, &fakeStore{}, config.DefaultSettings())

	if err := c.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if !h.closed || !k.closed {
		t.Fatalf("closed hotkeys=%v keeper=%v, want both true", h.closed, k.closed)
	}
}

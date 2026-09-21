package config

import (
	"testing"
	"time"
)

type memoryPreferences struct {
	ints    map[string]int
	strings map[string]string
}

func (m *memoryPreferences) IntWithFallback(key string, fallback int) int {
	if value, ok := m.ints[key]; ok {
		return value
	}
	return fallback
}

func (m *memoryPreferences) StringWithFallback(key, fallback string) string {
	if value, ok := m.strings[key]; ok {
		return value
	}
	return fallback
}

func (m *memoryPreferences) SetInt(key string, value int) {
	m.ints[key] = value
}

func (m *memoryPreferences) SetString(key, value string) {
	m.strings[key] = value
}

func TestLoadUsesDefaultsWhenPreferencesAreMissing(t *testing.T) {
	store := NewStore(&memoryPreferences{ints: map[string]int{}, strings: map[string]string{}})
	settings := store.Load()
	if settings.Interval != 30*time.Second {
		t.Fatalf("Interval = %v, want 30s", settings.Interval)
	}
	if got := settings.Shortcut.String(); got != "Ctrl+F1" {
		t.Fatalf("Shortcut = %q, want Ctrl+F1", got)
	}
}

func TestLoadFallsBackFromInvalidPreferences(t *testing.T) {
	prefs := &memoryPreferences{
		ints:    map[string]int{"interval_seconds": 17},
		strings: map[string]string{"hotkey": "F12"},
	}
	settings := NewStore(prefs).Load()
	if settings.Interval != 30*time.Second || settings.Shortcut.String() != "Ctrl+F1" {
		t.Fatalf("Load() = %#v, want default settings", settings)
	}
}

func TestSaveWritesCanonicalSettings(t *testing.T) {
	prefs := &memoryPreferences{ints: map[string]int{}, strings: map[string]string{}}
	store := NewStore(prefs)
	settings := DefaultSettings()
	settings.Interval = 60 * time.Second
	store.Save(settings)
	if prefs.ints["interval_seconds"] != 60 {
		t.Fatalf("saved interval = %d, want 60", prefs.ints["interval_seconds"])
	}
	if prefs.strings["hotkey"] != "Ctrl+F1" {
		t.Fatalf("saved hotkey = %q, want Ctrl+F1", prefs.strings["hotkey"])
	}
}

func TestValidIntervalAcceptsOnlyPresets(t *testing.T) {
	for _, seconds := range []int{5, 15, 30, 60} {
		if !ValidInterval(time.Duration(seconds) * time.Second) {
			t.Errorf("ValidInterval(%ds) = false", seconds)
		}
	}
	if ValidInterval(17 * time.Second) {
		t.Fatal("ValidInterval(17s) = true, want false")
	}
}

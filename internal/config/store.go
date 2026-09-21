package config

import (
	"time"

	"mousekeeper/internal/shortcut"
)

const (
	intervalKey = "interval_seconds"
	hotkeyKey   = "hotkey"
)

type Preferences interface {
	IntWithFallback(key string, fallback int) int
	StringWithFallback(key, fallback string) string
	SetInt(key string, value int)
	SetString(key, value string)
}

type Settings struct {
	Interval time.Duration
	Shortcut shortcut.Shortcut
}

type Store struct {
	preferences Preferences
}

func NewStore(preferences Preferences) *Store {
	return &Store{preferences: preferences}
}

func DefaultSettings() Settings {
	return Settings{Interval: 30 * time.Second, Shortcut: shortcut.Default()}
}

func ValidInterval(interval time.Duration) bool {
	switch interval {
	case 5 * time.Second, 15 * time.Second, 30 * time.Second, 60 * time.Second:
		return true
	default:
		return false
	}
}

func (s *Store) Load() Settings {
	defaults := DefaultSettings()
	interval := time.Duration(s.preferences.IntWithFallback(intervalKey, int(defaults.Interval/time.Second))) * time.Second
	if !ValidInterval(interval) {
		interval = defaults.Interval
	}
	parsed, err := shortcut.Parse(s.preferences.StringWithFallback(hotkeyKey, defaults.Shortcut.String()))
	if err != nil {
		parsed = defaults.Shortcut
	}
	return Settings{Interval: interval, Shortcut: parsed}
}

func (s *Store) Save(settings Settings) {
	s.preferences.SetInt(intervalKey, int(settings.Interval/time.Second))
	s.preferences.SetString(hotkeyKey, settings.Shortcut.String())
}

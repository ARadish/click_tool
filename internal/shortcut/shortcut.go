package shortcut

import (
	"fmt"
	"strconv"
	"strings"
)

// Shortcut is a normalized Windows global shortcut.
type Shortcut struct {
	Ctrl  bool
	Alt   bool
	Shift bool
	Key   string
}

func Default() Shortcut {
	return Shortcut{Ctrl: true, Key: "F1"}
}

func New(ctrl, alt, shift bool, key string) (Shortcut, error) {
	s := Shortcut{Ctrl: ctrl, Alt: alt, Shift: shift, Key: strings.ToUpper(strings.TrimSpace(key))}
	if err := s.Validate(); err != nil {
		return Shortcut{}, err
	}
	return s, nil
}

func Parse(value string) (Shortcut, error) {
	var s Shortcut
	parts := strings.Split(value, "+")
	for _, raw := range parts {
		part := strings.ToUpper(strings.TrimSpace(raw))
		switch part {
		case "CTRL":
			if s.Ctrl {
				return Shortcut{}, fmt.Errorf("duplicate Ctrl modifier")
			}
			s.Ctrl = true
		case "ALT":
			if s.Alt {
				return Shortcut{}, fmt.Errorf("duplicate Alt modifier")
			}
			s.Alt = true
		case "SHIFT":
			if s.Shift {
				return Shortcut{}, fmt.Errorf("duplicate Shift modifier")
			}
			s.Shift = true
		case "", "WIN", "WINDOWS", "META", "SUPER":
			return Shortcut{}, fmt.Errorf("unsupported shortcut part %q", raw)
		default:
			if s.Key != "" {
				return Shortcut{}, fmt.Errorf("shortcut must contain exactly one key")
			}
			s.Key = part
		}
	}
	if err := s.Validate(); err != nil {
		return Shortcut{}, err
	}
	return s, nil
}

func (s Shortcut) Validate() error {
	if !s.Ctrl && !s.Alt && !s.Shift {
		return fmt.Errorf("shortcut requires Ctrl, Alt, or Shift")
	}
	if !supportedKey(s.Key) {
		return fmt.Errorf("unsupported shortcut key %q", s.Key)
	}
	return nil
}

func (s Shortcut) String() string {
	parts := make([]string, 0, 4)
	if s.Ctrl {
		parts = append(parts, "Ctrl")
	}
	if s.Alt {
		parts = append(parts, "Alt")
	}
	if s.Shift {
		parts = append(parts, "Shift")
	}
	if s.Key != "" {
		parts = append(parts, strings.ToUpper(s.Key))
	}
	return strings.Join(parts, "+")
}

func SupportedKeys() []string {
	keys := make([]string, 0, 47)
	for c := 'A'; c <= 'Z'; c++ {
		keys = append(keys, string(c))
	}
	for c := '0'; c <= '9'; c++ {
		keys = append(keys, string(c))
	}
	for n := 1; n <= 11; n++ {
		keys = append(keys, "F"+strconv.Itoa(n))
	}
	return keys
}

func supportedKey(key string) bool {
	key = strings.ToUpper(key)
	if len(key) == 1 && ((key[0] >= 'A' && key[0] <= 'Z') || (key[0] >= '0' && key[0] <= '9')) {
		return true
	}
	if !strings.HasPrefix(key, "F") {
		return false
	}
	n, err := strconv.Atoi(strings.TrimPrefix(key, "F"))
	return err == nil && n >= 1 && n <= 11
}

//go:build windows

package desktop

import (
	"strconv"
	"strings"
	"sync"

	systemhotkey "golang.design/x/hotkey"
	"mousekeeper/internal/shortcut"
)

type systemBinding struct {
	hotkey *systemhotkey.Hotkey
	down   chan struct{}
	up     chan struct{}
	stop   chan struct{}
	once   sync.Once
}

func NewSystemHotkeys() *HotkeyManager {
	return newHotkeyManager(func(value shortcut.Shortcut) hotkeyBinding {
		return newSystemBinding(value)
	})
}

func newSystemBinding(value shortcut.Shortcut) *systemBinding {
	modifiers := make([]systemhotkey.Modifier, 0, 3)
	if value.Ctrl {
		modifiers = append(modifiers, systemhotkey.ModCtrl)
	}
	if value.Alt {
		modifiers = append(modifiers, systemhotkey.ModAlt)
	}
	if value.Shift {
		modifiers = append(modifiers, systemhotkey.ModShift)
	}
	return &systemBinding{
		hotkey: systemhotkey.New(modifiers, windowsHotkeyCode(value.Key)),
		down:   make(chan struct{}, 32),
		up:     make(chan struct{}, 4),
		stop:   make(chan struct{}),
	}
}

func (b *systemBinding) Register() error {
	if err := b.hotkey.Register(); err != nil {
		return err
	}
	go b.forwardEvents()
	return nil
}

func (b *systemBinding) Unregister() error {
	b.once.Do(func() { close(b.stop) })
	return b.hotkey.Unregister()
}

func (b *systemBinding) Keydown() <-chan struct{} { return b.down }
func (b *systemBinding) Keyup() <-chan struct{}   { return b.up }

func (b *systemBinding) forwardEvents() {
	for {
		select {
		case <-b.stop:
			return
		case _, ok := <-b.hotkey.Keydown():
			if !ok {
				return
			}
			select {
			case b.down <- struct{}{}:
			default:
			}
		case _, ok := <-b.hotkey.Keyup():
			if !ok {
				return
			}
			select {
			case b.up <- struct{}{}:
			default:
			}
		}
	}
}

func windowsHotkeyCode(key string) systemhotkey.Key {
	key = strings.ToUpper(key)
	if len(key) == 1 {
		return systemhotkey.Key(key[0])
	}
	n, _ := strconv.Atoi(strings.TrimPrefix(key, "F"))
	return systemhotkey.Key(0x6F + n)
}

package desktop

import (
	"context"
	"fmt"
	"sync"

	"mousekeeper/internal/shortcut"
)

type hotkeyBinding interface {
	Register() error
	Unregister() error
	Keydown() <-chan struct{}
	Keyup() <-chan struct{}
}

type hotkeyFactory func(shortcut.Shortcut) hotkeyBinding

type HotkeyManager struct {
	mu      sync.Mutex
	factory hotkeyFactory
	binding hotkeyBinding
	cancel  context.CancelFunc
}

func newHotkeyManager(factory hotkeyFactory) *HotkeyManager {
	return &HotkeyManager{factory: factory}
}

func (m *HotkeyManager) Replace(value shortcut.Shortcut, callback func()) error {
	if err := value.Validate(); err != nil {
		return err
	}
	candidate := m.factory(value)
	if err := candidate.Register(); err != nil {
		return fmt.Errorf("register hotkey: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go listenForHotkey(ctx, candidate, callback)

	m.mu.Lock()
	oldBinding := m.binding
	oldCancel := m.cancel
	m.binding = candidate
	m.cancel = cancel
	m.mu.Unlock()

	if oldCancel != nil {
		oldCancel()
	}
	if oldBinding != nil {
		_ = oldBinding.Unregister()
	}
	return nil
}

func (m *HotkeyManager) Close() error {
	m.mu.Lock()
	binding := m.binding
	cancel := m.cancel
	m.binding = nil
	m.cancel = nil
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if binding != nil {
		return binding.Unregister()
	}
	return nil
}

func listenForHotkey(ctx context.Context, binding hotkeyBinding, callback func()) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-binding.Keydown():
			if !ok {
				return
			}
			callback()
			select {
			case <-ctx.Done():
				return
			case _, ok := <-binding.Keyup():
				if !ok {
					return
				}
			}
			for {
				select {
				case <-binding.Keydown():
					continue
				default:
					goto drained
				}
			}
		drained:
		}
	}
}

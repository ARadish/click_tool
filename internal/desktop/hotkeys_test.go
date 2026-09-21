package desktop

import (
	"errors"
	"sync"
	"testing"
	"time"

	"mousekeeper/internal/shortcut"
)

type fakeBinding struct {
	down         chan struct{}
	up           chan struct{}
	registerErr  error
	unregistered bool
}

func newFakeBinding() *fakeBinding {
	return &fakeBinding{down: make(chan struct{}, 8), up: make(chan struct{}, 8)}
}
func (f *fakeBinding) Register() error          { return f.registerErr }
func (f *fakeBinding) Unregister() error        { f.unregistered = true; return nil }
func (f *fakeBinding) Keydown() <-chan struct{} { return f.down }
func (f *fakeBinding) Keyup() <-chan struct{}   { return f.up }

func waitDesktop(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition was not satisfied before timeout")
}

func TestHotkeyInvokesCallbackOnceUntilKeyIsReleased(t *testing.T) {
	binding := newFakeBinding()
	m := newHotkeyManager(func(shortcut.Shortcut) hotkeyBinding { return binding })
	defer m.Close()
	var mu sync.Mutex
	count := 0
	value := shortcut.Default()
	if err := m.Replace(value, func() { mu.Lock(); count++; mu.Unlock() }); err != nil {
		t.Fatalf("Replace returned error: %v", err)
	}

	binding.down <- struct{}{}
	binding.down <- struct{}{}
	waitDesktop(t, func() bool { mu.Lock(); defer mu.Unlock(); return count == 1 })
	binding.up <- struct{}{}
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 1 {
		t.Fatalf("callback count = %d, want 1", count)
	}
}

func TestReplaceKeepsOldBindingWhenNewRegistrationFails(t *testing.T) {
	oldBinding := newFakeBinding()
	badBinding := newFakeBinding()
	badBinding.registerErr = errors.New("conflict")
	bindings := []*fakeBinding{oldBinding, badBinding}
	index := 0
	m := newHotkeyManager(func(shortcut.Shortcut) hotkeyBinding {
		binding := bindings[index]
		index++
		return binding
	})
	defer m.Close()
	if err := m.Replace(shortcut.Default(), func() {}); err != nil {
		t.Fatalf("first Replace returned error: %v", err)
	}
	altK, _ := shortcut.Parse("Alt+K")

	if err := m.Replace(altK, func() {}); err == nil {
		t.Fatal("second Replace succeeded, want conflict")
	}
	if oldBinding.unregistered {
		t.Fatal("old binding was unregistered after replacement failed")
	}
}

func TestReplaceUnregistersOldBindingAfterNewRegistration(t *testing.T) {
	first := newFakeBinding()
	second := newFakeBinding()
	bindings := []*fakeBinding{first, second}
	index := 0
	m := newHotkeyManager(func(shortcut.Shortcut) hotkeyBinding {
		binding := bindings[index]
		index++
		return binding
	})
	defer m.Close()
	m.Replace(shortcut.Default(), func() {})
	altK, _ := shortcut.Parse("Alt+K")

	if err := m.Replace(altK, func() {}); err != nil {
		t.Fatalf("second Replace returned error: %v", err)
	}
	if !first.unregistered {
		t.Fatal("old binding was not unregistered")
	}
}

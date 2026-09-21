package shortcut

import "testing"

func TestParseCanonicalizesShortcut(t *testing.T) {
	s, err := Parse("shift+ctrl+f1")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if got := s.String(); got != "Ctrl+Shift+F1" {
		t.Fatalf("String() = %q, want %q", got, "Ctrl+Shift+F1")
	}
}

func TestParseRejectsShortcutWithoutModifier(t *testing.T) {
	if _, err := Parse("F1"); err == nil {
		t.Fatal("Parse(F1) succeeded, want modifier validation error")
	}
}

func TestParseRejectsF12(t *testing.T) {
	if _, err := Parse("Ctrl+F12"); err == nil {
		t.Fatal("Parse(Ctrl+F12) succeeded, want reserved-key error")
	}
}

func TestParseAcceptsSupportedKeys(t *testing.T) {
	for _, value := range []string{"Ctrl+A", "Alt+9", "Ctrl+Alt+F11"} {
		if _, err := Parse(value); err != nil {
			t.Errorf("Parse(%q) returned error: %v", value, err)
		}
	}
}

func TestParseRejectsWindowsModifier(t *testing.T) {
	if _, err := Parse("Win+K"); err == nil {
		t.Fatal("Parse(Win+K) succeeded, want unsupported-modifier error")
	}
}

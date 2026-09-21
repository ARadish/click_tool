package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenRotatesOversizedLog(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, fileName)
	if err := os.WriteFile(path, make([]byte, maxLogSize+1), 0o600); err != nil {
		t.Fatal(err)
	}

	file, err := Open(directory)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	file.Close()

	if _, err := os.Stat(path + ".1"); err != nil {
		t.Fatalf("rotated log missing: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("new log missing: %v", err)
	}
	if info.Size() != 0 {
		t.Fatalf("new log size = %d, want 0", info.Size())
	}
}

func TestOpenCreatesDirectoryAndLog(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "nested", "MouseKeeper")
	file, err := Open(directory)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	file.Close()
	if _, err := os.Stat(filepath.Join(directory, fileName)); err != nil {
		t.Fatalf("log file missing: %v", err)
	}
}

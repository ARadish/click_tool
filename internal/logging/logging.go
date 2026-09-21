package logging

import (
	"os"
	"path/filepath"
)

const (
	fileName   = "mousekeeper.log"
	maxLogSize = 1024 * 1024
)

func Open(directory string) (*os.File, error) {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(directory, fileName)
	if info, err := os.Stat(path); err == nil && info.Size() > maxLogSize {
		_ = os.Remove(path + ".1")
		if err := os.Rename(path, path+".1"); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
}

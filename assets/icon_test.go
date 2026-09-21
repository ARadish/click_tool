package assets

import (
	"bytes"
	"testing"
)

func TestIconIsEmbeddedPNG(t *testing.T) {
	content := Icon().Content()
	if !bytes.HasPrefix(content, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		t.Fatal("embedded icon is not a PNG resource")
	}
}

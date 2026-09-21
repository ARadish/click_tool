package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed Icon.png
var iconBytes []byte

// Icon returns the application icon embedded in the executable.
func Icon() fyne.Resource {
	return fyne.NewStaticResource("Icon.png", iconBytes)
}

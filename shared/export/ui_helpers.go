package export

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CreateExportButton creates a consistently-styled export button.
func CreateExportButton(label string, action func(), window fyne.Window) *widget.Button {
	_ = window // reserved for future dialog standardization hooks
	return widget.NewButtonWithIcon(label, iconForExportLabel(label), action)
}

func iconForExportLabel(label string) fyne.Resource {
	lower := strings.ToLower(label)
	switch {
	case strings.Contains(lower, "image") || strings.Contains(lower, "png"):
		return theme.MediaPhotoIcon()
	case strings.Contains(lower, "repro"):
		return theme.StorageIcon()
	default:
		return theme.DocumentSaveIcon()
	}
}

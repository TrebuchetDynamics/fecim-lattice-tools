//go:build legacy_fyne

package widgets

import (
	"time"

	"fecim-lattice-tools/shared/widgets/interaction"
)

// ThrottledSlider provides lightweight preview updates during drag and a full commit on release.
type ThrottledSlider = interaction.ThrottledSlider

// NewThrottledSlider creates a slider with throttled preview callback and full commit callback.
func NewThrottledSlider(min, max float64, previewInterval time.Duration, onPreview, onCommit func(value float64)) *ThrottledSlider {
	return interaction.NewThrottledSlider(min, max, previewInterval, onPreview, onCommit)
}

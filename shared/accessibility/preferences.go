//go:build cgo

package accessibility

import "fyne.io/fyne/v2"

const (
	PreferenceKeyLargeTextMode = "accessibility.large_text_mode"
	PreferenceKeyReducedMotion = "accessibility.reduced_motion"
)

const (
	NormalTextScale float32 = 1.0
	LargeTextScale  float32 = 1.35
)

func IsLargeTextModeEnabled(app fyne.App) bool {
	if app == nil {
		return false
	}
	return app.Preferences().BoolWithFallback(PreferenceKeyLargeTextMode, false)
}

func SetLargeTextMode(app fyne.App, enabled bool) {
	if app == nil {
		return
	}
	app.Preferences().SetBool(PreferenceKeyLargeTextMode, enabled)
}

func TextScale(app fyne.App) float32 {
	if IsLargeTextModeEnabled(app) {
		return LargeTextScale
	}
	return NormalTextScale
}

func IsReducedMotionEnabled(app fyne.App) bool {
	if app == nil {
		return false
	}
	return app.Preferences().BoolWithFallback(PreferenceKeyReducedMotion, false)
}

func SetReducedMotion(app fyne.App, enabled bool) {
	if app == nil {
		return
	}
	app.Preferences().SetBool(PreferenceKeyReducedMotion, enabled)
}

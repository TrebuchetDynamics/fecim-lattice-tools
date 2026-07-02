package themes

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"fecim-lattice-tools/shared/accessibility"
)

func TestManagerApplyAccessibilityPreferences_LargeTextScale(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	m := NewManager(app)
	m.SetTheme(ThemeDark)

	baseSize := app.Settings().Theme().Size("text")
	accessibility.SetLargeTextMode(app, true)
	m.ApplyAccessibilityPreferences()

	scaledSize := app.Settings().Theme().Size("text")
	if scaledSize <= baseSize {
		t.Fatalf("expected scaled text size > base size, got base=%f scaled=%f", baseSize, scaledSize)
	}
}

func TestScaledThemeLargeTextDoesNotScaleLayoutPadding(t *testing.T) {
	base := GetTheme(ThemeHighContrast)
	scaled := NewScaledTheme(base, accessibility.LargeTextScale)

	if got, want := scaled.Size(theme.SizeNameText), base.Size(theme.SizeNameText)*accessibility.LargeTextScale; got != want {
		t.Fatalf("text size = %f, want %f", got, want)
	}
	if got, want := scaled.Size(theme.SizeNamePadding), base.Size(theme.SizeNamePadding); got != want {
		t.Fatalf("padding size = %f, want unscaled %f", got, want)
	}
	if got, want := scaled.Size(theme.SizeNameInnerPadding), base.Size(theme.SizeNameInnerPadding); got != want {
		t.Fatalf("inner padding size = %f, want unscaled %f", got, want)
	}
}

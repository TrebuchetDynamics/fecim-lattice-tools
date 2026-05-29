//go:build legacy_fyne

package visual

import "testing"

func TestVCellOverlayColor_NeutralAtZero(t *testing.T) {
	c := VCellOverlayColor(0, 2.0)
	if !(c.R >= 220 && c.G >= 220 && c.B >= 220) {
		t.Fatalf("expected near-neutral color at 0V, got RGB=(%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestVCellOverlayColor_WarmAndCoolAtExtremes(t *testing.T) {
	warm := VCellOverlayColor(+2.0, 2.0)
	cool := VCellOverlayColor(-2.0, 2.0)

	if !(warm.R > warm.B && warm.R > warm.G) {
		t.Fatalf("expected warm color at +max, got RGB=(%d,%d,%d)", warm.R, warm.G, warm.B)
	}
	if !(cool.B > cool.R && cool.B > cool.G) {
		t.Fatalf("expected cool color at -max, got RGB=(%d,%d,%d)", cool.R, cool.G, cool.B)
	}
}

//go:build legacy_fyne

package visual

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewVCellLegendSpec_ShowsUnitString(t *testing.T) {
	spec := NewVCellLegendSpec(1.0)
	if !strings.Contains(spec.Title, "(V)") {
		t.Fatalf("legend title missing unit V: %q", spec.Title)
	}
}

func TestNewVCellLegendSpec_RangeSymmetricAroundZero(t *testing.T) {
	spec := NewVCellLegendSpec(2.5)
	if got := spec.Min + spec.Max; got > 1e-9 || got < -1e-9 {
		t.Fatalf("expected symmetric range around 0, min=%.6f max=%.6f", spec.Min, spec.Max)
	}
	if !reflect.DeepEqual(spec.TickText, []string{"-Vmax", "0", "+Vmax"}) {
		t.Fatalf("unexpected VC tick labels: %#v", spec.TickText)
	}
	if spec.SignText != "+ = BL>WL" {
		t.Fatalf("unexpected VC sign semantics label: %q", spec.SignText)
	}
}

func TestNewVCellLegendSpec_DefaultsInvalidRange(t *testing.T) {
	spec := NewVCellLegendSpec(0)
	if spec.Min != -1 || spec.Max != 1 {
		t.Fatalf("expected default +/-1V range, got min=%.6f max=%.6f", spec.Min, spec.Max)
	}
}

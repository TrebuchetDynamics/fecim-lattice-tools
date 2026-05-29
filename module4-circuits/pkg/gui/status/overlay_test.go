//go:build legacy_fyne

package status

import (
	"strings"
	"testing"
)

func TestFormatOverlayCellInfo_LevelAndVoltageLabels(t *testing.T) {
	info := FormatOverlayCellInfo(7, -0.1234, "Vcell")
	if info.TopLabel != "L: 7" {
		t.Fatalf("unexpected top label: %q", info.TopLabel)
	}
	if !strings.HasPrefix(info.BottomLabel, "V:") {
		t.Fatalf("bottom label missing voltage prefix: %q", info.BottomLabel)
	}
}

func TestFormatOverlayCellInfo_CurrentLabel(t *testing.T) {
	info := FormatOverlayCellInfo(3, -1.234e-6, "Icell")
	if info.BottomLabel != "I: -1.23 µA" {
		t.Fatalf("unexpected current bottom label: %q", info.BottomLabel)
	}
}

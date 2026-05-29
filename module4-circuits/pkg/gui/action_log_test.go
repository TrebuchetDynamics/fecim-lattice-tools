//go:build legacy_fyne

package gui

import "testing"

func TestModeLabelHelpers(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"op read", opModeLabel(OpModeRead), "READ"},
		{"op write", opModeLabel(OpModeWrite), "WRITE"},
		{"op compute", opModeLabel(OpModeCompute), "COMPUTE"},
		{"op unknown", opModeLabel(OpMode(99)), "IDLE"},
		{"dac manual", dacModeLabel(DACManual), "MANUAL"},
		{"dac input", dacModeLabel(DACInputVector), "INPUT_VECTOR"},
		{"dac unknown", dacModeLabel(DACMode(99)), "UNKNOWN"},
		{"range read", dacRangeLabel(DACRangeRead), "READ"},
		{"range write", dacRangeLabel(DACRangeWrite), "WRITE"},
		{"range unknown", dacRangeLabel(DACRangeMode(99)), "UNKNOWN"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Fatalf("%s: got %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

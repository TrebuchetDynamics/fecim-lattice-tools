//go:build legacy_fyne

package status

import "testing"

func TestModeLabels(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"op read", OpModeLabel(OperationRead), "READ"},
		{"op write", OpModeLabel(OperationWrite), "WRITE"},
		{"op compute", OpModeLabel(OperationCompute), "COMPUTE"},
		{"op unknown", OpModeLabel(99), "IDLE"},
		{"dac manual", DACModeLabel(DACManualMode), "MANUAL"},
		{"dac input", DACModeLabel(DACInputVectorMode), "INPUT_VECTOR"},
		{"dac unknown", DACModeLabel(99), "UNKNOWN"},
		{"range read", DACRangeLabel(DACRangeReadMode), "READ"},
		{"range write", DACRangeLabel(DACRangeWriteMode), "WRITE"},
		{"range unknown", DACRangeLabel(99), "UNKNOWN"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Fatalf("%s: got %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

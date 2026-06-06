package levels

import "testing"

func TestQuantizeToDefaultLevels(t *testing.T) {
	cases := []struct {
		name  string
		input float64
		want  float64
	}{
		{name: "clamps low", input: -0.2, want: 0},
		{name: "clamps high", input: 1.2, want: 1},
		{name: "midpoint maps to nearest 30-level bin", input: 0.5, want: 15.0 / 29.0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := QuantizeToDefaultLevels(tc.input)
			if got != tc.want {
				t.Fatalf("QuantizeToDefaultLevels(%v) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestDefaultLevelForPreservesLegacyCalculation(t *testing.T) {
	if got := DefaultLevelFor(15.0 / 29.0); got != 15 {
		t.Fatalf("DefaultLevelFor(15/29) = %d, want 15", got)
	}
	if got := DefaultLevelFor(QuantizeToDefaultLevels(-1)); got != 0 {
		t.Fatalf("DefaultLevelFor(quantized -1) = %d, want 0", got)
	}
	if got := DefaultLevelFor(QuantizeToDefaultLevels(2)); got != 29 {
		t.Fatalf("DefaultLevelFor(quantized 2) = %d, want 29", got)
	}
}

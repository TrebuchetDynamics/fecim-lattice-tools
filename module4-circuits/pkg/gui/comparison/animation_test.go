//go:build legacy_fyne

package comparison

import "testing"

func TestAnimationSteps(t *testing.T) {
	steps := AnimationSteps()
	if len(steps) != 6 {
		t.Fatalf("expected 6 steps, got %d", len(steps))
	}
	if steps[0] != "Step 1: CPU loads data from DRAM (250ns)..." {
		t.Fatalf("unexpected first step: %q", steps[0])
	}
	if steps[len(steps)-1] != "Animation complete: FeFET ≈6.6x faster than CPU (latency model)" {
		t.Fatalf("unexpected final step: %q", steps[len(steps)-1])
	}
}

func TestNextScaleSize(t *testing.T) {
	cases := map[int]int{0: 8, 8: 16, 16: 32, 32: 64, 64: 8}
	for in, want := range cases {
		if got := NextScaleSize(in); got != want {
			t.Fatalf("NextScaleSize(%d) got %d, want %d", in, got, want)
		}
	}
}

package timeline

import (
	"strings"
	"testing"
)

func TestAnimationSteps_KnownOperations(t *testing.T) {
	cases := []struct {
		op        string
		wantFinal string
	}{
		{"WRITE", "Write complete: Total 203ns"},
		{"READ", "Read complete: Total 76ns"},
		{"COMPUTE", "Compute complete: Total 76ns for full MVM"},
	}
	for _, tc := range cases {
		steps := AnimationSteps(tc.op)
		if len(steps) != 6 {
			t.Fatalf("%s: expected 6 steps, got %d", tc.op, len(steps))
		}
		if got := steps[len(steps)-1]; got != tc.wantFinal {
			t.Fatalf("%s: final step got %q, want %q", tc.op, got, tc.wantFinal)
		}
		if !strings.HasPrefix(steps[0], "Phase 1:") {
			t.Fatalf("%s: first step should start with phase: %q", tc.op, steps[0])
		}
	}
}

func TestAnimationSteps_UnknownOperation(t *testing.T) {
	steps := AnimationSteps("BAD")
	if len(steps) != 1 || steps[0] != "Select an operation to animate" {
		t.Fatalf("unexpected fallback steps: %#v", steps)
	}
}

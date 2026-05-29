//go:build legacy_fyne

package writeseq

import "testing"

func TestCellKey(t *testing.T) {
	if got := CellKey(2, 3); got != "2,3" {
		t.Fatalf("CellKey got %q, want 2,3", got)
	}
}

func TestPhaseName(t *testing.T) {
	cases := map[int]string{
		PhaseIdle:   "IDLE",
		PhaseReset:  "RESET",
		PhaseHold1:  "HOLD",
		PhaseWrite:  "WRITE",
		PhaseHold2:  "HOLD",
		PhaseVerify: "VERIFY",
		99:          "UNKNOWN",
	}
	for phase, want := range cases {
		if got := PhaseName(phase); got != want {
			t.Fatalf("PhaseName(%d) got %q, want %q", phase, got, want)
		}
	}
}

func TestPhaseDuration(t *testing.T) {
	cases := map[int]int{
		PhaseIdle:   0,
		PhaseReset:  PhaseResetDurationNs,
		PhaseHold1:  PhaseHold1DurationNs,
		PhaseWrite:  PhaseWriteDurationNs,
		PhaseHold2:  PhaseHold2DurationNs,
		PhaseVerify: PhaseVerifyDurationNs,
		99:          0,
	}
	for phase, want := range cases {
		if got := PhaseDuration(phase); got != want {
			t.Fatalf("PhaseDuration(%d) got %d, want %d", phase, got, want)
		}
	}
}

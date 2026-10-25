package calibration

import (
	"strings"
	"testing"
)

func TestDefaultKilohertzTargetIsSimulationDefault(t *testing.T) {
	for _, target := range DefaultTargets().FrequencySweep {
		if target.FrequencyHz == 1e3 && !strings.Contains(target.Citation, "simulation-default") {
			t.Fatalf("unsupported Park extraction attribution: %s", target.Citation)
		}
	}
}

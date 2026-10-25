package physics

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParkPresetNamesDeclareSyntheticReference(t *testing.T) {
	for _, mat := range []*HZOMaterial{Park2015Fig2aHZO10nm(), CalibratedPark2015HZO()} {
		if !strings.Contains(strings.ToLower(mat.Name), "synthetic") {
			t.Errorf("fixture-derived preset name must say synthetic, got %q", mat.Name)
		}
	}
}

func TestLoadPELoopCSVSupportsSyntheticProvenanceHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.csv")
	if err := os.WriteFile(path, []byte("# SYNTHETIC, not experimental\n# generator: model sweep\nE_MV_cm,P_uC_cm2\n-3,-19.4\n0,-15.8\n3,19.4\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	points, err := LoadPELoopCSV(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 3 || points[0].Field_Vm != -3e8 || math.Abs(points[0].Polarization_Cm+0.194) > 1e-12 {
		t.Fatalf("commented fixture parsed incorrectly: %+v", points)
	}
}

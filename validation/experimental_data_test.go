package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

func TestExperimentalDataValidation(t *testing.T) {
	root := filepath.Join("..", "experimental-data")

	var jsonFiles []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			jsonFiles = append(jsonFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk experimental-data: %v", err)
	}
	if len(jsonFiles) == 0 {
		t.Fatalf("no JSON datasets found under %s", root)
	}

	allowedUnits := map[string]bool{
		"uC/cm2": true,
		"MV/cm":  true,
		"ns":     true,
		"Hz":     true,
		"K":      true,
	}

	var peCount int
	var sqErrSum float64
	var nErr int

	for _, f := range jsonFiles {
		ds, err := LoadLiteratureDataset(f)
		if err != nil {
			t.Fatalf("load %s: %v", f, err)
		}

		if strings.TrimSpace(ds.Reference.DOI) == "" {
			t.Fatalf("%s: DOI is empty", f)
		}
		if strings.TrimSpace(ds.Reference.Authors) == "" || ds.Reference.Year == 0 || strings.TrimSpace(ds.Reference.Title) == "" || strings.TrimSpace(ds.Reference.Journal) == "" {
			t.Fatalf("%s: incomplete citation metadata", f)
		}
		if strings.TrimSpace(ds.Reference.Figure) == "" && strings.TrimSpace(ds.Reference.Table) == "" {
			t.Fatalf("%s: figure/table anchor required", f)
		}
		if len(ds.DataPoints) == 0 {
			t.Fatalf("%s: no data points", f)
		}

		for _, dp := range ds.DataPoints {
			if strings.TrimSpace(dp.Unit) == "" || !allowedUnits[dp.Unit] {
				t.Fatalf("%s: invalid or missing unit %q", f, dp.Unit)
			}
		}

		if strings.Contains(f, string(filepath.Separator)+"pe-loops"+string(filepath.Separator)) {
			peCount++
			for _, dp := range ds.DataPoints {
				pred, ok := predictPE(dp)
				if !ok {
					continue
				}
				delta := pred - dp.Value
				sqErrSum += delta * delta
				nErr++
			}
		}
	}

	if peCount == 0 {
		t.Fatal("no pe-loop datasets found")
	}
	if nErr == 0 {
		t.Fatal("no comparable pe-loop points for simulation-vs-experiment")
	}
	rmse := sqrt(sqErrSum / float64(nErr))
	if rmse > 3.0 {
		t.Fatalf("P-E comparison RMSE too high: %.3f (limit 3.0)", rmse)
	}
	t.Logf("validated %d datasets (%d P-E), compared %d points, RMSE=%.3f", len(jsonFiles), peCount, nErr, rmse)
}

func predictPE(dp LiteratureDataPoint) (float64, bool) {
	freq := 1e3
	temp := 300.0
	if v, ok := dp.Conditions["frequency_hz"]; ok && v > 0 {
		freq = v
	}
	if v, ok := dp.Conditions["temperature_k"]; ok && v > 0 {
		temp = v
	}

	base := sharedphysics.HysteresisMetrics{FrequencyHz: 1e3, Pr_Cm2: 21e-2, Ec_Vm: 1.02e8, LoopArea_Jm3: 2.35e8}
	cfg := sharedphysics.FrequencyDispersionConfig{
		ReferenceHz:        1e3,
		EcLogSlope:         0.030,
		PrLogSlope:         -0.035,
		LoopAreaLogSlope:   -0.045,
		MinMultiplierClamp: 0.4,
	}
	m, _ := sharedphysics.ApplyFrequencyDispersion(base, freq, cfg)

	prTempFactor := 1.0 - 0.0012*(temp-300.0)
	ecTempFactor := 1.0 - 0.0008*(temp-300.0)

	switch dp.Unit {
	case "uC/cm2":
		return m.Pr_Cm2 * 100.0 * prTempFactor, true
	case "MV/cm":
		return (m.Ec_Vm / 1e8) * ecTempFactor, true
	default:
		return 0, false
	}
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

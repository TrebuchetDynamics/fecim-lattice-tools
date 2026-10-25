package literature

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHZOForcReferenceDeclaresSyntheticTargets(t *testing.T) {
	ref := HZO10nm_FORCReference()
	if !strings.Contains(strings.ToLower(ref.Source), "synthetic") || !strings.Contains(strings.ToLower(ref.MaterialName), "synthetic") {
		t.Fatalf("model-derived FORC targets presented as measured data: %+v", ref)
	}
}

func TestParkFixtureDeclaresSyntheticProvenance(t *testing.T) {
	base := filepath.Join("data", "park2015_synthetic_shape")
	body, err := os.ReadFile(base + ".csv")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(body), "# SYNTHETIC") || !strings.Contains(string(body), "not experimental") {
		t.Fatal("Park fixture must carry an explicit synthetic, not experimental header")
	}
	var prov struct {
		DatasetID  string `json:"dataset_id"`
		Status     string `json:"status"`
		Evidence   string `json:"evidence"`
		Generation struct {
			Method             string `json:"method"`
			HistoricalRevision string `json:"historical_revision"`
		} `json:"generation"`
	}
	raw, err := os.ReadFile(base + ".provenance.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &prov); err != nil {
		t.Fatal(err)
	}
	if prov.DatasetID != "park2015_synthetic_shape" || prov.Status != "synthetic_regression" || prov.Evidence != "simulation-default" || prov.Generation.Method == "" || prov.Generation.HistoricalRevision != "unknown" {
		t.Fatalf("incorrect synthetic provenance: %+v", prov)
	}
	for _, loader := range []struct {
		name string
		load func(string) ([]float64, []float64, error)
	}{{"loop", loadCSVLoop}, {"diagnostic", loadPECSV}} {
		t.Run(loader.name, func(t *testing.T) {
			e, p, err := loader.load(base + ".csv")
			if err != nil {
				t.Fatal(err)
			}
			if len(e) != 61 || len(p) != 61 || e[0] != -3 || e[30] != 3 || e[60] != -3 {
				t.Fatalf("historical regression sweep changed: E=%v, len(P)=%d", e, len(p))
			}
		})
	}
}

func TestExcludedBTODigitizationCannotPassProvenanceGate(t *testing.T) {
	err := validateStrictProvenance(peLoopDataset{
		MaterialID: "bto2021_cryst11101192_hysteresis_digitized",
		Provenance: filepath.Join("data", "bto2021_cryst11101192_hysteresis_digitized.provenance.json"),
		DOI:        "10.3390/cryst11101192",
		FieldUnit:  "MV/cm", PolarUnit: "uC/cm2",
	})
	if err == nil || !strings.Contains(err.Error(), "excluded") {
		t.Fatalf("unreviewed BTO extraction must be explicitly excluded, got %v", err)
	}
}

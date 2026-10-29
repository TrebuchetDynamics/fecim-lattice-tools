package comparison

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestComparisonWebUnits(t *testing.T) {
	// 10^9 MACs at 10^6 fJ/MAC is exactly one J/inference.
	use, err := EnergyAtRate(1_000_000_000, 1_000_000, 2)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(use.JPerInference-1) > 1e-15 || use.DailyInferences != 172800 || math.Abs(use.DailyKWh-0.048) > 1e-15 || use.PowerW != 2 || math.Abs(use.MonthlyUSD-0.146) > 1e-15 {
		t.Fatalf("dimension identities: %+v", use)
	}
	result, err := CalculateScenario("GPT-2", 2, ScenarioBaseline)
	if err != nil {
		t.Fatal(err)
	}
	want := (result.Energy[1].DailyKWh - result.Energy[2].DailyKWh) * 365 * 0.10 / 1e6
	if result.AnnualSavingsMillionUSD != want {
		t.Fatalf("million USD: %g != %g", result.AnnualSavingsMillionUSD, want)
	}
	data, err := MarshalScenario(result, "json")
	if err != nil {
		t.Fatal(err)
	}
	var artifact struct {
		SchemaVersion string           `json:"schema_version"`
		ModelVersion  string           `json:"model_version"`
		Sources       []string         `json:"sources"`
		Result        ScenarioEstimate `json:"result"`
	}
	if err = json.Unmarshal(data, &artifact); err != nil {
		t.Fatal(err)
	}
	if artifact.SchemaVersion != "fecim-comparison/2" || artifact.ModelVersion == "" || len(artifact.Sources) == 0 || artifact.Result.Energy[1] != result.Energy[1] {
		t.Fatalf("versioned canonical export: %s", data)
	}
	if strings.Contains(string(data), `"co2_reduction_pct","Value":0`) {
		t.Fatal("unavailable CO2 serialized as zero")
	}
	csvData, err := MarshalScenario(result, "csv")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(bytes.NewReader(csvData)).ReadAll()
	if err != nil || len(rows) != 4 || rows[1][0] != ScenarioArtifactVersion || rows[1][8] != "172800" {
		t.Fatalf("canonical CSV dimensions: %s", csvData)
	}
	invalid := result
	invalid.Energy[0].DailyKWh = math.NaN()
	if _, err := MarshalScenario(invalid, "csv"); err == nil {
		t.Fatal("CSV accepted nonfinite canonical metric")
	}
	for _, v := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if _, err := EnergyAtRate(1, v, 1); err == nil {
			t.Errorf("accepted invalid energy %g", v)
		}
		if _, err := EnergyAtRate(1, 1, v); err == nil {
			t.Errorf("accepted invalid rate %g", v)
		}
	}
}

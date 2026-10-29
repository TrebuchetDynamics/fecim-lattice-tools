package hysteresis

import (
	"encoding/json"
	"reflect"
	"strings"

	"fecim-lattice-tools/shared/constants"
	"testing"
)

func TestScientificScalarPresentation(t *testing.T) {
	s, _ := NewSession(DefaultConfig())
	s.Apply(Command{Kind: Run})
	snap, _ := s.Advance(Budget{Steps: 1, DT: 0.01})
	page, _ := s.Page(PageQuery{Revision: snap.Revision})
	// Decode the public JSON seam, so RED is a missing presentation rather than
	// a compiler failure. Canonical producer values must remain unchanged.
	check := func(value any, field, unit, speech string, canonical float64) {
		t.Helper()
		data, _ := json.Marshal(value)
		var fields map[string]json.RawMessage
		json.Unmarshal(data, &fields)
		var quantity Quantity
		json.Unmarshal(fields[field], &quantity)
		if quantity.SI != canonical || quantity.Unit != unit || quantity.Presentation.Availability != "available" || !strings.Contains(quantity.Presentation.Speakable, speech) {
			t.Fatalf("%s lost canonical/unit/Go speech: %+v", field, quantity)
		}
	}
	check(page.Rows[0], "TimeValue", "s", "seconds", page.Rows[0].Time)
	// Every numeric material parameter exposed by the metadata view needs a
	// Go-owned unit and speech, not only the loop's E and P columns.
	data, _ := json.Marshal(snap)
	var snapshot struct{ ScientificValues map[string]Quantity }
	json.Unmarshal(data, &snapshot)
	material := reflect.ValueOf(snap.Material.Parameters)
	for i := 0; i < material.NumField(); i++ {
		if material.Field(i).Kind() != reflect.Float64 {
			continue
		}
		name := material.Type().Field(i).Name
		q := snapshot.ScientificValues["Material / Parameters / "+name]
		want := material.Field(i).Float()
		if name == "Ea" {
			want *= constants.ElectronChargeC
		}
		if name == "StressGPa" {
			want *= 1e9
		}
		if q.SI != want || q.Unit == "" || q.Presentation.Speakable == "" || q.Presentation.Availability != "available" {
			t.Fatalf("material %s lacks canonical SI/unit/Go speech: %+v", name, q)
		}
	}
	if q := snapshot.ScientificValues["Material / Parameters / Ec"]; q.Unit != "V/m" || !strings.Contains(q.Presentation.Speakable, "megavolts per centimeter") {
		t.Fatalf("coercivity lacks Go field units/speech: %+v", q)
	}
	job, _ := NewAnalysis(AnalysisRequest{Input: snap, Kind: AnalysisPUND})
	result := finishAnalysis(t, job)
	traces, _ := job.Page(AnalysisPageQuery{Revision: result.Revision, Table: PUNDTraces})
	check(traces.Rows[0], "CurrentValue", "A", "amperes", traces.Rows[0].CurrentA)
	data, _ = json.Marshal(result)
	var wire struct{ ScientificValues map[string]Quantity }
	json.Unmarshal(data, &wire)
	q := wire.ScientificValues["PUNDCharges / QP_C"]
	if q.SI != result.PUND.Charges.QP_C || !strings.Contains(q.Presentation.Speakable, "coulombs") {
		t.Fatalf("charge has no Go speech: %+v", q)
	}
	job, _ = NewAnalysis(AnalysisRequest{Input: snap, Kind: AnalysisFORC, FORC: FORCParameters{EmaxVm: 1, Reversals: 3, Resolution: 3}})
	result = finishAnalysis(t, job)
	density, _ := job.Page(AnalysisPageQuery{Revision: result.Revision, Table: FORCDensity})
	check(density.Rows[0], "DensityValue", "C/V²", "coulombs per square volt", density.Rows[0].Density)
}

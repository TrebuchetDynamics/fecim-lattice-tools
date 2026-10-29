package hysteresis

import (
	"bytes"
	"context"
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/physics"
)

func analysisInput(t *testing.T, model Model, material string) Snapshot {
	t.Helper()
	c := DefaultConfig()
	c.Model = model
	c.MaterialID = material
	if model == Landau {
		c.Temperature = 450
		c.Stress = 2e9
	}
	s, err := NewSession(c)
	if err != nil {
		t.Fatal(err)
	}
	return s.Snapshot()
}
func finishAnalysis(t *testing.T, j *AnalysisJob) AnalysisResult {
	t.Helper()
	for i := 0; i < 200; i++ {
		p, err := j.Advance(AnalysisBudget{Steps: 1})
		if err != nil {
			t.Fatal(err)
		}
		if p.State == "complete" {
			r, err := j.Result()
			if err != nil {
				t.Fatal(err)
			}
			return r
		}
	}
	t.Fatal("analysis never completed")
	return AnalysisResult{}
}

func TestAnalysisPUND(t *testing.T) {
	ec := 3e7
	want, traces, err := physics.RunPUNDSimulation(physics.NewPreisachStack(ec, &physics.TanhEverett{Ec: ec, Ps: 0.25, Delta: ec * 0.3}), 5*ec, 100e-9, 5e-9, 1e-12)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range Catalogue() {
		for _, engine := range []Model{Preisach, Landau} {
			input := analysisInput(t, engine, m.ID)
			j, err := NewAnalysis(AnalysisRequest{Input: input, Kind: AnalysisPUND})
			if err != nil {
				t.Fatal(err)
			}
			r := finishAnalysis(t, j)
			if r.Diagnostic.Profile != "legacy-pund-demo" || r.Diagnostic.Version != 1 || r.Diagnostic.Model != Preisach || r.Diagnostic.MaterialID != "hzo-demo" || r.Context.Config != input.Config || r.Context.Revision != input.Revision {
				t.Fatalf("wrong identity: %+v", r)
			}
			p := r.Diagnostic.Applied
			if p.EcVm != ec || p.PsCm2 != 0.25 || p.DeltaVm != 0.3*ec || p.AmplitudeVm != 5*ec || p.PulseWidthS != 100e-9 || p.SampleIntervalS != 5e-9 || p.AreaM2 != 1e-12 {
				t.Fatalf("wrong applied profile: %+v", p)
			}
			if r.PUND.Charges != want || !reflect.DeepEqual(r.PUND.Traces, traces) || r.ExperimentalValidation || r.Evidence != "simulation_only" {
				t.Fatal("PUND oracle/evidence mismatch")
			}
			page, err := j.Page(AnalysisPageQuery{Revision: r.Revision, Table: PUNDTraces})
			if err != nil || page.Diagnostic != r.Diagnostic || page.Context != r.Context || page.Evidence != "simulation_only" || page.Total != 84 || len(page.Rows) != 32 || page.Rows[0].Pulse != "P" || page.Rows[21].Pulse != "U" || page.Rows[0].CurrentA != traces[0][0].CurrentA {
				t.Fatalf("trace page: %+v, %v", page, err)
			}
			r.PUND.Traces[0][0].CurrentA = 42
			again, _ := j.Result()
			if again.PUND.Traces[0][0].CurrentA != traces[0][0].CurrentA {
				t.Fatal("result aliases job")
			}
			data, err := j.Artifact(PUNDCSV)
			if err != nil || !bytes.Contains(data, []byte("legacy-pund-demo")) || !bytes.Contains(data, []byte("simulation_only")) {
				t.Fatal("PUND CSV loses identity", err)
			}
		}
	}
}

func TestAnalysisFORC(t *testing.T) {
	input := analysisInput(t, Landau, DefaultConfig().MaterialID)
	j, err := NewAnalysis(AnalysisRequest{Input: input, Kind: AnalysisFORC})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = j.Result(); !errors.Is(err, ErrAnalysisIncomplete) {
		t.Fatal("partial result called complete", err)
	}
	p, err := j.Advance(AnalysisBudget{Steps: 1})
	if err != nil || p.State != "running" {
		t.Fatal(p, err)
	}
	if _, err = j.Page(AnalysisPageQuery{Revision: p.Revision - 1, Table: FORCSweep}); !errors.Is(err, ErrStaleRevision) {
		t.Fatal("accepted stale page", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p, err = j.Advance(AnalysisBudget{Steps: 1, Context: ctx})
	if !errors.Is(err, context.Canceled) || p.State != "cancelled" {
		t.Fatal(p, err)
	}
	if _, err = j.Result(); !errors.Is(err, ErrAnalysisIncomplete) {
		t.Fatal("cancelled result called complete")
	}
	if _, err = j.Artifact(AnalysisJSON); !errors.Is(err, ErrAnalysisIncomplete) {
		t.Fatal("exported cancelled matrix")
	}
	j, _ = NewAnalysis(AnalysisRequest{Input: input, Kind: AnalysisFORC})
	input.Config.Temperature = 200 // request is an immutable copy
	r := finishAnalysis(t, j)
	if r.Context.Config.Temperature != 450 || r.Diagnostic.Model != Preisach || r.Diagnostic.Profile != "gui-forc" || r.Diagnostic.Applied.EmaxVm != 1 || r.Diagnostic.Applied.PsCm2 != input.Material.Parameters.Ps {
		t.Fatalf("wrong FORC identity: %+v", r)
	}
	if len(r.FORC.Sweep.Curves) != 21 || r.FORC.Matrix.Rows != 21 || r.FORC.Matrix.Cols != 21 || len(r.FORC.Sweep.ReversalPairs) == 0 {
		t.Fatal("lost sweep/grid metadata")
	}
	page, err := j.Page(AnalysisPageQuery{Revision: r.Revision, Table: FORCSweep})
	if err != nil || page.Total != 231 || page.Rows[0].Branch != "descending" {
		t.Fatal(page, err)
	}
	density, err := j.Page(AnalysisPageQuery{Revision: r.Revision, Table: FORCDensity})
	if err != nil || density.Total != 441 || density.Rows[0].Ea.SI != -1 || density.Rows[0].Eb.SI != -1 {
		t.Fatal(density, err)
	}
	for _, args := range []FORCParameters{{EmaxVm: math.NaN(), Reversals: 21, Resolution: 21}, {EmaxVm: math.Inf(1), Reversals: 21, Resolution: 21}, {EmaxVm: math.MaxFloat64, Reversals: 21, Resolution: 21}, {EmaxVm: math.SmallestNonzeroFloat64, Reversals: 21, Resolution: 21}, {EmaxVm: 1, Reversals: 65, Resolution: 21}, {EmaxVm: 1, Reversals: 21, Resolution: 257}, {EmaxVm: 1, Reversals: 2, Resolution: 21}, {EmaxVm: 1, Reversals: 21, Resolution: 2}} {
		if _, err := NewAnalysis(AnalysisRequest{Input: input, Kind: AnalysisFORC, FORC: args}); err == nil {
			t.Fatalf("accepted invalid bounds %+v", args)
		}
	}
	for _, steps := range []int{0, 257} {
		if _, err := j.Advance(AnalysisBudget{Steps: steps}); err == nil {
			t.Fatal("accepted invalid budget")
		}
	}
	for _, kind := range []AnalysisArtifact{AnalysisJSON, FORCSweepCSV, FORCDensityCSV} {
		if b, err := j.Artifact(kind); err != nil || len(b) == 0 {
			t.Fatal(kind, err)
		}
	}
}

func TestAnalysisEvidence(t *testing.T) {
	cases := []struct {
		format, data, branch, eunit, punit string
		e, p                               float64
	}{
		{"csv", "E_MV_cm,P_uC_cm2\n-1,-20\n1,20\n", "ascending", "MV/cm", "uC/cm2", -1e8, -0.2},
		{"json", `{"source":"synthetic cited example","material":"synthetic","E_V_m":[1,-1],"P_C_m2":[0.2,-0.2]}`, "descending", "V/m", "C/m2", 1, 0.2},
	}
	for _, c := range cases {
		ref, err := ParseReference(ReferenceInput{Format: c.format, Source: "synthetic cited example", Data: []byte(c.data)})
		if err != nil {
			t.Fatal(err)
		}
		if ref.Source != "synthetic cited example" || ref.Branch != c.branch || ref.Evidence != "imported_unverified" || ref.ExperimentalValidation || ref.SourceUnits.EFieldUnit != c.eunit || ref.SourceUnits.PolarizationUnit != c.punit || ref.Points[0].E.SI != c.e || ref.Points[0].P.SI != c.p || ref.Comparison.Available || ref.Comparison.Reason == "" {
			t.Fatalf("reference loses context: %+v", ref)
		}
	}
	input := analysisInput(t, Preisach, DefaultConfig().MaterialID)
	for _, bad := range []ReferenceInput{
		{Format: "csv", Data: []byte("E_MV_cm,P_uC_cm2\nNaN,2\n1,2")},
		{Format: "csv", Data: []byte("E_MV_cm,P_uC_cm2\n1e308,2\n1,2")},
		{Format: "json", Data: []byte(`{"E_V_m":[1,1],"P_C_m2":[1,2]}`)},
		{Format: "json", Data: []byte(`{"E_V_m":[null,1],"P_C_m2":[1,2]}`)},
		{Format: "json", Data: []byte(`{"E_V_m":[1,2,1],"P_C_m2":[1,2,1]}`)},
		{Format: "json", Data: []byte(`{"E_V_m":[1,2],"P_C_m2":[1]}`)},
		{Format: "json", Data: []byte(`{"E_V_m":[1,2],"P_C_m2":[1,2],"experimental_validation":true}`)},
		{Format: "json", Data: []byte(`{"E_V_m":[1],"E_V_m":[2],"P_C_m2":[3,4]}`)},
		{Format: "json", Data: []byte(`{"E_V_m":[1,2],"P_C_m2":[3],"p_c_m2":[4]}`)},
		{Format: "json", Data: []byte(`{"source":"first","SOURCE":"second","E_V_m":[1,2],"P_C_m2":[3,4]}`)},
		{Format: "json", Data: []byte(`{"material":"first","material":"second","E_V_m":[1,2],"P_C_m2":[3,4]}`)},
		{Format: "json", Data: []byte(`{"branch":"descending","branch":"ascending","E_V_m":[1,2],"P_C_m2":[3,4]}`)},
		{Format: "json", Data: []byte(`{} {}`)}, {Format: "unknown", Data: []byte("x")},
		{Format: "csv", Data: []byte(strings.Repeat("x", MaxReferenceBytes+1))},
	} {
		if _, err := ParseReference(bad); err == nil {
			t.Fatal("accepted malformed reference", bad.Format)
		}
		if _, err := NewAnalysis(AnalysisRequest{Input: input, Kind: AnalysisPUND, References: []ReferenceInput{bad}}); err == nil {
			t.Fatal("analysis accepted malformed reference", bad.Format)
		}
	}
	req := AnalysisRequest{Input: input, Kind: AnalysisPUND, References: []ReferenceInput{{Format: "json", Data: []byte(cases[1].data)}}}
	j, err := NewAnalysis(req)
	if err != nil {
		t.Fatal(err)
	}
	req.References[0].Data[0] = '!'
	r := finishAnalysis(t, j)
	if len(r.References) != 1 || r.ReferenceContext.Evidence != "simulation_default" || r.ReferenceContext.PrBand.Min.SI != input.Material.Parameters.Pr*0.8 || r.ReferenceContext.EcBand.Max.SI != input.Material.Parameters.Ec*1.3 || r.ReferenceContext.Comparison.Available || r.ReferenceContext.Comparison.Reason == "" || len(r.ReferenceContext.Citations) == 0 {
		t.Fatal("illustrative band promoted to evidence")
	}
	b, err := j.Artifact(AnalysisJSON)
	if err != nil || !bytes.Contains(b, []byte("imported_unverified")) || !bytes.Contains(b, []byte("synthetic cited example")) || !bytes.Contains(b, []byte("descending")) || bytes.Contains(b, []byte(`"ExperimentalValidation":true`)) {
		t.Fatal("serialization loses evidence", err)
	}
	// Reference work is capped across branches, not just by raw byte size.
	csvBranch := func(n int) ReferenceInput {
		var data strings.Builder
		data.WriteString("E_MV_cm,P_uC_cm2\n")
		for i := 0; i < n; i++ {
			data.WriteString(strconv.Itoa(i) + ",0\n")
		}
		return ReferenceInput{Format: "csv", Source: "synthetic allocation-bound input", Data: []byte(data.String())}
	}
	part := csvBranch(2049)
	if _, err := NewAnalysis(AnalysisRequest{Input: input, Kind: AnalysisPUND, References: []ReferenceInput{part, part}}); err == nil {
		t.Fatal("accepted more than 4096 aggregate reference points")
	}
	if _, err := ParseReference(csvBranch(4097)); err == nil {
		t.Fatal("accepted oversized reference branch")
	}
	maxJob, err := NewAnalysis(AnalysisRequest{Input: input, Kind: AnalysisFORC, FORC: FORCParameters{EmaxVm: 1, Reversals: 64, Resolution: 256}, References: []ReferenceInput{csvBranch(4096)}})
	if err != nil {
		t.Fatal(err)
	}
	finishAnalysis(t, maxJob)
	output, err := maxJob.Artifact(AnalysisJSON)
	if err != nil || len(output) > 8<<20 {
		t.Fatalf("maximum admitted result exceeds 8 MiB: %d bytes, %v", len(output), err)
	}
}

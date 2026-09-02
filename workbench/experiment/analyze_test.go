package experiment

import (
	"testing"

	"fecim-lattice-tools/workbench/project"
)

func analysisRun(id string, status Status, metrics ...Metric) RunRecord {
	return RunRecord{
		Manifest: RunManifest{RunID: id, Status: status},
		Result:   Result{Status: status, Metrics: metrics},
	}
}

func metric(name string, value float64, unit string) Metric {
	return Metric{Name: name, Value: value, Unit: unit, Model: "test", Assumptions: []string{"test"}, Evidence: EvidenceDerived}
}

func analysisBundle() project.Bundle {
	bundle := fixtureBundle()
	bundle.Project.Objectives = []project.Objective{
		{Metric: "energy_pj", Direction: project.Minimize},
		{Metric: "latency_ns", Direction: project.Minimize},
	}
	bundle.Project.Constraints = []project.Constraint{{Metric: "area_um2", Operator: "<=", Value: 5000, Unit: "um2"}}
	return bundle
}

func TestAnalyzeClassifiesFailuresConstraintsAndPareto(t *testing.T) {
	runs := []RunRecord{
		analysisRun("a", StatusSuccess, metric("energy_pj", 10, "pJ"), metric("latency_ns", 10, "ns"), metric("area_um2", 100, "um2")),
		analysisRun("b", StatusSuccess, metric("energy_pj", 8, "pJ"), metric("latency_ns", 12, "ns"), metric("area_um2", 100, "um2")),
		analysisRun("c", StatusSuccess, metric("energy_pj", 12, "pJ"), metric("latency_ns", 12, "ns"), metric("area_um2", 100, "um2")),
		analysisRun("d", StatusSuccess, metric("energy_pj", 5, "pJ"), metric("latency_ns", 5, "ns"), metric("area_um2", 6000, "um2")),
		{Manifest: RunManifest{RunID: "e", Status: StatusFailed}, Result: Result{Status: StatusFailed, Failure: &Failure{Kind: "model-domain", Message: "failed"}}},
	}
	got := Analyze(analysisBundle(), runs)
	if len(got.Runs) != 5 {
		t.Fatalf("runs=%d", len(got.Runs))
	}
	if !got.Runs[0].Feasible || !got.Runs[0].Pareto || !got.Runs[1].Pareto {
		t.Fatalf("expected a and b on Pareto front: %+v", got.Runs)
	}
	if got.Runs[2].Pareto || got.Runs[3].Feasible || got.Runs[4].Feasible {
		t.Fatalf("unexpected classifications: %+v", got.Runs)
	}
	if got.Counts["pareto"] != 2 || got.Counts["failed"] != 1 || got.Counts["infeasible"] != 1 {
		t.Fatalf("counts=%v", got.Counts)
	}
}

func TestAnalyzeMarksMissingAndMismatchedMetricsUnusable(t *testing.T) {
	runs := []RunRecord{
		analysisRun("missing", StatusSuccess, metric("energy_pj", 10, "pJ"), metric("area_um2", 10, "um2")),
		analysisRun("unit", StatusSuccess, metric("energy_pj", 10, "pJ"), metric("latency_ns", 10, "ns"), metric("area_um2", 10, "mm2")),
	}
	got := Analyze(analysisBundle(), runs)
	for _, run := range got.Runs {
		if run.UnusableReason == "" || run.Feasible || run.Pareto {
			t.Fatalf("run should be unusable: %+v", run)
		}
	}
	if got.Counts["unusable"] != 2 {
		t.Fatalf("counts=%v", got.Counts)
	}
}

func TestAnalyzeHonorsMaximizeAndKeepsEqualPoints(t *testing.T) {
	bundle := analysisBundle()
	bundle.Project.Objectives = []project.Objective{{Metric: "score", Direction: project.Maximize}}
	bundle.Project.Constraints = nil
	runs := []RunRecord{
		analysisRun("a", StatusSuccess, metric("score", 2, "ratio")),
		analysisRun("b", StatusSuccess, metric("score", 2, "ratio")),
		analysisRun("c", StatusSuccess, metric("score", 1, "ratio")),
	}
	got := Analyze(bundle, runs)
	if !got.Runs[0].Pareto || !got.Runs[1].Pareto || got.Runs[2].Pareto {
		t.Fatalf("maximize Pareto=%+v", got.Runs)
	}
}

func TestAnalyzeConstraintOperators(t *testing.T) {
	for _, test := range []struct {
		op    string
		value float64
		pass  bool
	}{
		{"<", 11, true}, {"<=", 10, true}, {">", 9, true}, {">=", 10, true}, {"==", 10, true}, {"<", 10, false},
	} {
		bundle := analysisBundle()
		bundle.Project.Objectives = []project.Objective{{Metric: "energy_pj", Direction: project.Minimize}}
		bundle.Project.Constraints = []project.Constraint{{Metric: "area_um2", Operator: test.op, Value: test.value, Unit: "um2"}}
		got := Analyze(bundle, []RunRecord{analysisRun("x", StatusSuccess, metric("energy_pj", 1, "pJ"), metric("area_um2", 10, "um2"))})
		if got.Runs[0].Feasible != test.pass {
			t.Fatalf("operator %s value %g feasible=%v", test.op, test.value, got.Runs[0].Feasible)
		}
	}
}

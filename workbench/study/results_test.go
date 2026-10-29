package study

import (
	"reflect"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/experiment"
)

func TestQueryResults(t *testing.T) {
	metric := func(value float64, unit string) experiment.Metric {
		return experiment.Metric{Name: "energy", Value: value, Unit: unit, Model: "fixture", Evidence: experiment.EvidenceDerived, Assumptions: []string{"synthetic"}}
	}
	row := func(id string, status experiment.Status, feasible, pareto bool, metrics ...experiment.Metric) experiment.AnalyzedRun {
		return experiment.AnalyzedRun{Run: experiment.RunRecord{Manifest: experiment.RunManifest{RunID: id}, Result: experiment.Result{Status: status, Metrics: metrics}}, Feasible: feasible, Pareto: pareto}
	}
	analysis := experiment.Analysis{Runs: []experiment.AnalyzedRun{
		row("c", experiment.StatusSuccess, true, false, metric(2, "pJ")),
		row("b", experiment.StatusSuccess, true, true, metric(1, "pJ")),
		row("a", experiment.StatusSuccess, false, false, metric(1, "pJ")),
		row("d", experiment.StatusFailed, false, false),
	}}
	t.Run("full filtered supplementary plot", func(t *testing.T) {
		x := func(v float64) experiment.Metric { m := metric(v, "pJ"); m.Name = "energy_pj"; return m }
		y := func(v float64) experiment.Metric { m := metric(v, "ns"); m.Name = "latency_ns"; return m }
		a := experiment.Analysis{Runs: []experiment.AnalyzedRun{row("b", experiment.StatusSuccess, true, true, x(2), y(4)), row("a", experiment.StatusSuccess, true, false, x(1), y(2)), row("missing", experiment.StatusSuccess, true, false, x(3)), row("failed", experiment.StatusFailed, false, false)}}
		query := ResultsQuery{Page: 2, PageSize: 1, Filter: "feasible"}
		page, err := QueryResults(a, query)
		if err != nil {
			t.Fatal(err)
		}
		plot, err := PlotResults(a, query)
		if err != nil {
			t.Fatal(err)
		}
		if plot.Revision != page.Revision || plot.FilteredTotal != 3 || plot.DatasetTotal != 4 || plot.Plotted != 2 || plot.Omitted != 1 || len(plot.Points) != 2 || plot.Points[0].RunID != "a" || plot.Points[1].RunID != "b" || plot.Points[0].X != 0 || plot.Points[0].Y != 600 || plot.Points[1].X != 1000 || plot.Points[1].Y != 0 || plot.XAxis.Unit != "pJ" || plot.YAxis.Unit != "ns" || !strings.Contains(plot.XAxis.MinimumPresentation.Speakable, "picojoules") || plot.Caption == "" || plot.Scope != "all-filtered-runs" {
			t.Fatalf("plot/table snapshot mismatch: %+v", plot)
		}
		absent, err := PlotResults(analysis, ResultsQuery{})
		if err != nil || absent.Availability != "unavailable" || absent.Omitted != 4 || absent.Reason == "" {
			t.Fatalf("missing metrics should not become zero: %+v %v", absent, err)
		}
	})
	page, err := QueryResults(analysis, ResultsQuery{Sort: "energy", PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if page.Page != 1 || page.PageSize != 2 || page.PageCount != 2 || page.Total != 4 || page.Rows[0].RunID != "a" || page.Rows[1].RunID != "b" || len(page.Revision) != 64 || page.Rows[0].Metrics[0].Metric.Value != 1 || page.Rows[0].Metrics[0].Presentation.Speakable != "1 picojoules" {
		t.Fatalf("page = %+v", page)
	}
	again, err := QueryResults(analysis, ResultsQuery{Page: 2, PageSize: 2, Sort: "energy", Revision: page.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if again.Reset || again.Rows[0].RunID != "c" || again.Rows[1].RunID != "d" || again.Rows[1].Metrics[0].Metric != nil || again.Rows[1].Metrics[0].Presentation.Availability != "unavailable" {
		t.Fatalf("page 2 = %+v", again)
	}
	desc, err := QueryResults(analysis, ResultsQuery{Sort: "energy", Descending: true})
	if err != nil {
		t.Fatal(err)
	}
	if desc.PageSize != 32 || desc.Rows[0].RunID != "c" || desc.Rows[1].RunID != "a" || desc.Rows[2].RunID != "b" || desc.Rows[3].RunID != "d" {
		t.Fatalf("descending=%+v", desc)
	}
	for _, tt := range []struct {
		filter string
		want   int
	}{{"all", 4}, {"success", 3}, {"failed", 1}, {"feasible", 2}, {"pareto", 1}} {
		got, err := QueryResults(analysis, ResultsQuery{Filter: tt.filter})
		if err != nil || got.Total != tt.want {
			t.Fatalf("%s: %+v %v", tt.filter, got, err)
		}
	}
	reordered := analysis
	reordered.Runs = append([]experiment.AnalyzedRun(nil), analysis.Runs...)
	reordered.Runs[0], reordered.Runs[3] = reordered.Runs[3], reordered.Runs[0]
	stable, err := QueryResults(reordered, ResultsQuery{})
	if err != nil || stable.Revision != page.Revision {
		t.Fatalf("unstable revision: %+v %v", stable, err)
	}
	changed := analysis
	changed.Runs = append([]experiment.AnalyzedRun(nil), analysis.Runs[:3]...)
	stale, err := QueryResults(changed, ResultsQuery{Page: 2, PageSize: 2, Revision: page.Revision})
	if err != nil || !stale.Reset || stale.Page != 1 || stale.ResetReason == "" || stale.Revision == page.Revision {
		t.Fatalf("stale=%+v %v", stale, err)
	}
	for _, q := range []ResultsQuery{{Filter: "unknown"}, {Sort: "unknown"}, {Page: -1}, {PageSize: 101}, {PageSize: -1}} {
		if _, err := QueryResults(analysis, q); err == nil {
			t.Fatalf("invalid query accepted: %+v", q)
		}
	}
	empty, err := QueryResults(experiment.Analysis{}, ResultsQuery{Page: 100})
	if err != nil || empty.Page != 1 || empty.PageCount != 1 || len(empty.Rows) != 0 {
		t.Fatalf("empty: %+v %v", empty, err)
	}
	compared, err := CompareResults(analysis, []string{"b", "a"})
	if err != nil {
		t.Fatal(err)
	}
	if compared.Revision != page.Revision || !reflect.DeepEqual(compared.RunIDs, []string{"b", "a"}) || len(compared.Metrics) != 1 || compared.Metrics[0].Availability != "available" || compared.Metrics[0].Values[0].Metric.Value != 1 {
		t.Fatalf("comparison=%+v", compared)
	}
	missing, err := CompareResults(analysis, []string{"a", "d"})
	if err != nil || missing.Metrics[0].Availability != "unavailable" || missing.Metrics[0].Values[1].Metric != nil || missing.Metrics[0].Values[1].Presentation.Reason == "" {
		t.Fatalf("missing comparison=%+v %v", missing, err)
	}
	allMissing := analysis
	allMissing.Runs = append(append([]experiment.AnalyzedRun(nil), analysis.Runs...), row("e", experiment.StatusFailed, false, false))
	missingPair, err := CompareResults(allMissing, []string{"d", "e"})
	if err != nil || len(missingPair.Metrics) != 1 || missingPair.Metrics[0].Availability != "unavailable" {
		t.Fatalf("failed selections lost the dataset's metric rows: %+v %v", missingPair, err)
	}
	mismatch := analysis
	mismatch.Runs = append([]experiment.AnalyzedRun(nil), analysis.Runs...)
	mismatch.Runs[1].Run.Result.Metrics = []experiment.Metric{metric(1, "J")}
	incompatible, err := CompareResults(mismatch, []string{"b", "a"})
	if err != nil || incompatible.Metrics[0].Availability != "unavailable" || !strings.Contains(incompatible.Metrics[0].Reason, "units") || incompatible.Metrics[0].Values[0].Presentation.Availability != "unavailable" {
		t.Fatalf("unit mismatch=%+v %v", incompatible, err)
	}
	for _, ids := range [][]string{{"a"}, {"a", "a"}, {"a", "absent"}, {"a", "b", "c", "d", "e"}} {
		if _, err := CompareResults(analysis, ids); err == nil {
			t.Fatalf("invalid selection accepted: %v", ids)
		}
	}
	if analysis.Runs[0].Run.Manifest.RunID != "c" || analysis.Runs[1].Run.Result.Metrics[0].Unit != "pJ" {
		t.Fatal("source mutated")
	}
}

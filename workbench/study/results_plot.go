package study

import (
	"fmt"
	"math"
	"sort"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/presentation"
)

type ResultsPlotAxis struct {
	Name                string             `json:"name"`
	Label               string             `json:"label"`
	Unit                string             `json:"unit"`
	Minimum             *float64           `json:"minimum,omitempty"`
	Maximum             *float64           `json:"maximum,omitempty"`
	MinimumPresentation presentation.Value `json:"minimum_presentation"`
	MaximumPresentation presentation.Value `json:"maximum_presentation"`
}
type ResultsPlotPoint struct {
	RunID string `json:"run_id"`
	// X/Y are presentation-only SVG coordinates in the Go-owned 1000 by 600
	// viewBox, not scientific values. Y is inverted for SVG. Canonical metrics,
	// model, assumptions, evidence and Go display/speech remain alongside them.
	X       float64     `json:"x"`
	Y       float64     `json:"y"`
	XMetric MetricValue `json:"x_metric"`
	YMetric MetricValue `json:"y_metric"`
}
type ResultsPlotOmission struct {
	RunID  string `json:"run_id"`
	Reason string `json:"reason"`
}
type ResultsPlot struct {
	Revision      string                `json:"revision"`
	Filter        string                `json:"filter"`
	Scope         string                `json:"scope"`
	Caption       string                `json:"caption"`
	Availability  string                `json:"availability"`
	Reason        string                `json:"reason,omitempty"`
	DatasetTotal  int                   `json:"dataset_total"`
	FilteredTotal int                   `json:"filtered_total"`
	Plotted       int                   `json:"plotted"`
	Omitted       int                   `json:"omitted"`
	Width         int                   `json:"width"`
	Height        int                   `json:"height"`
	XAxis         ResultsPlotAxis       `json:"x_axis"`
	YAxis         ResultsPlotAxis       `json:"y_axis"`
	Points        []ResultsPlotPoint    `json:"points"`
	Omissions     []ResultsPlotOmission `json:"omissions"`
}

// PlotResults is a supplementary representation of the complete filtered set,
// never just the visible table page. QueryResults supplies identical validation,
// filtering vocabulary and dataset revision. Missing/failed/incomparable metrics
// are explicit omissions, not zeros or implicit unit conversions. The result
// retains the entire stored filtered dataset, including historical runs.
func PlotResults(a experiment.Analysis, q ResultsQuery) (ResultsPlot, error) {
	page, err := QueryResults(a, q)
	if err != nil {
		return ResultsPlot{}, err
	}
	out := ResultsPlot{Revision: page.Revision, Filter: page.Filter, Scope: "all-filtered-runs", DatasetTotal: page.DatasetTotal, FilteredTotal: page.Total, Width: 1000, Height: 600, Availability: "unavailable", Points: []ResultsPlotPoint{}, Omissions: []ResultsPlotOmission{}, XAxis: ResultsPlotAxis{Name: "energy_pj", Label: "Energy", Unit: "pJ"}, YAxis: ResultsPlotAxis{Name: "latency_ns", Label: "Latency", Unit: "ns"}}
	for _, run := range a.Runs {
		if !includeResult(run, page.Filter) {
			continue
		}
		x, y := resultMetric(run, "energy_pj"), resultMetric(run, "latency_ns")
		reason := ""
		if x.Presentation.Availability != "available" || y.Presentation.Availability != "available" {
			reason = "Energy or latency is unavailable for this run"
		} else if x.Metric.Unit != "pJ" || y.Metric.Unit != "ns" {
			reason = "Plot requires energy in picojoules and latency in nanoseconds; no implicit conversion"
		}
		if reason != "" {
			out.Omissions = append(out.Omissions, ResultsPlotOmission{RunID: run.Run.Manifest.RunID, Reason: reason})
			continue
		}
		out.Points = append(out.Points, ResultsPlotPoint{RunID: run.Run.Manifest.RunID, XMetric: x, YMetric: y})
	}
	sort.Slice(out.Points, func(i, j int) bool { return out.Points[i].RunID < out.Points[j].RunID })
	sort.Slice(out.Omissions, func(i, j int) bool { return out.Omissions[i].RunID < out.Omissions[j].RunID })
	out.Plotted = len(out.Points)
	out.Omitted = len(out.Omissions)
	if out.Plotted == 0 {
		out.Reason = "No filtered runs have comparable finite energy and latency"
		out.XAxis.MinimumPresentation = presentation.FormatMetric(experiment.Metric{}, out.Reason)
		out.XAxis.MaximumPresentation = out.XAxis.MinimumPresentation
		out.YAxis.MinimumPresentation = out.XAxis.MinimumPresentation
		out.YAxis.MaximumPresentation = out.XAxis.MinimumPresentation
	} else {
		out.Availability = "available"
		xmin, xmax := out.Points[0].XMetric.Metric.Value, out.Points[0].XMetric.Metric.Value
		ymin, ymax := out.Points[0].YMetric.Metric.Value, out.Points[0].YMetric.Metric.Value
		for _, point := range out.Points {
			xmin = math.Min(xmin, point.XMetric.Metric.Value)
			xmax = math.Max(xmax, point.XMetric.Metric.Value)
			ymin = math.Min(ymin, point.YMetric.Metric.Value)
			ymax = math.Max(ymax, point.YMetric.Metric.Value)
		}
		out.XAxis.Minimum = &xmin
		out.XAxis.Maximum = &xmax
		out.YAxis.Minimum = &ymin
		out.YAxis.Maximum = &ymax
		out.XAxis.MinimumPresentation = presentation.FormatMetric(experiment.Metric{Value: xmin, Unit: "pJ"}, "")
		out.XAxis.MaximumPresentation = presentation.FormatMetric(experiment.Metric{Value: xmax, Unit: "pJ"}, "")
		out.YAxis.MinimumPresentation = presentation.FormatMetric(experiment.Metric{Value: ymin, Unit: "ns"}, "")
		out.YAxis.MaximumPresentation = presentation.FormatMetric(experiment.Metric{Value: ymax, Unit: "ns"}, "")
		for i := range out.Points {
			p := &out.Points[i]
			p.X = 1000 * plotFraction(p.XMetric.Metric.Value, xmin, xmax)
			p.Y = 600 * (1 - plotFraction(p.YMetric.Metric.Value, ymin, ymax))
		}
	}
	out.Caption = fmt.Sprintf("Energy versus latency: %d of %d filtered runs plotted, %d omitted; filter %s. Supplementary to the run table.", out.Plotted, out.FilteredTotal, out.Omitted, out.Filter)
	return out, nil
}
func includeResult(r experiment.AnalyzedRun, filter string) bool {
	return filter == "all" || filter == "success" && r.Run.Result.Status == experiment.StatusSuccess || filter == "failed" && r.Run.Result.Status == experiment.StatusFailed || filter == "feasible" && r.Feasible || filter == "pareto" && r.Pareto
}
func plotFraction(value, low, high float64) float64 {
	if low == high {
		return 0.5
	}
	if low < 0 && high > 0 {
		scale := math.Max(math.Abs(low), math.Abs(high))
		return (value/scale - low/scale) / (high/scale - low/scale)
	}
	return (value - low) / (high - low)
}

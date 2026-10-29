package study

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/presentation"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/report"
)

// ResultsQuery uses one-based pages, a default page size of 32 (maximum 100),
// filters all/success/failed/feasible/pareto and run_id or a canonical metric name
// for Sort. Missing/unavailable metrics sort last in either direction. Ties
// always use ascending immutable run IDs, never formatted numeric text.
// Revision is the dataset revision observed by the caller, not a file path.
type ResultsQuery struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Filter     string `json:"filter"`
	Sort       string `json:"sort"`
	Descending bool   `json:"descending"`
	Revision   string `json:"revision"`
}

// MetricValue retains canonical provenance and Go presentation. Nil Metric
// means missing, not zero. Incomparable units retain the original metric but
// explicitly mark Presentation unavailable; no implicit conversion is made.
type MetricValue struct {
	Name         string             `json:"name"`
	Metric       *experiment.Metric `json:"metric,omitempty"`
	Presentation presentation.Value `json:"presentation"`
}

type ResultRow struct {
	RunID          string            `json:"run_id"`
	Status         experiment.Status `json:"status"`
	Feasible       bool              `json:"feasible"`
	Pareto         bool              `json:"pareto"`
	UnusableReason string            `json:"unusable_reason,omitempty"`
	Metrics        []MetricValue     `json:"metrics"`
}

type ResultsPage struct {
	Revision     string      `json:"revision"`
	Reset        bool        `json:"reset"`
	ResetReason  string      `json:"reset_reason,omitempty"`
	Page         int         `json:"page"`
	PageSize     int         `json:"page_size"`
	PageCount    int         `json:"page_count"`
	Total        int         `json:"total"`
	DatasetTotal int         `json:"dataset_total"`
	Filter       string      `json:"filter"`
	Sort         string      `json:"sort"`
	Descending   bool        `json:"descending"`
	Caption      string      `json:"caption"`
	MetricNames  []string    `json:"metric_names"`
	Rows         []ResultRow `json:"rows"`
}

type ComparisonMetric struct {
	Name         string        `json:"name"`
	Availability string        `json:"availability"`
	Reason       string        `json:"reason,omitempty"`
	Values       []MetricValue `json:"values"`
}

type ResultsComparison struct {
	Revision string             `json:"revision"`
	RunIDs   []string           `json:"run_ids"`
	Metrics  []ComparisonMetric `json:"metrics"`
}

// QueryResults is storage-independent and never mutates analysis. The revision
// hashes the original scientific report projection in ID order, excluding
// timestamps, disk integrity metadata and presentation-only text. A changed
// requested revision resets to page 1 with a readable explanation.
func QueryResults(analysis experiment.Analysis, query ResultsQuery) (ResultsPage, error) {
	page := ResultsPage{}
	if query.Page < 0 || query.PageSize < 0 || query.PageSize > 100 {
		return page, fmt.Errorf("invalid page or page size")
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 32
	}
	if query.Filter == "" {
		query.Filter = "all"
	}
	if query.Sort == "" {
		query.Sort = "run_id"
	}
	switch query.Filter {
	case "all", "success", "failed", "feasible", "pareto":
	default:
		return page, fmt.Errorf("unsupported result filter")
	}
	revision, names, err := resultDataset(analysis)
	if err != nil {
		return page, err
	}
	if query.Sort != "run_id" {
		found := false
		for _, name := range names {
			found = found || query.Sort == name
		}
		if !found {
			return page, fmt.Errorf("unsupported result sort")
		}
		unit := ""
		for _, run := range analysis.Runs {
			cell := resultMetric(run, query.Sort)
			if cell.Metric != nil && cell.Presentation.Availability == "available" {
				if unit != "" && unit != cell.Metric.Unit {
					return page, fmt.Errorf("sort metric has incompatible units")
				}
				unit = cell.Metric.Unit
			}
		}
	}
	rows := make([]experiment.AnalyzedRun, 0, len(analysis.Runs))
	for _, run := range analysis.Runs {
		include := includeResult(run, query.Filter)
		if include {
			rows = append(rows, run)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if query.Sort != "run_id" {
			av, bv := resultMetric(a, query.Sort), resultMetric(b, query.Sort)
			aa, ba := av.Presentation.Availability == "available", bv.Presentation.Availability == "available"
			if aa != ba {
				return aa
			}
			if aa && av.Metric.Value != bv.Metric.Value {
				if query.Descending {
					return av.Metric.Value > bv.Metric.Value
				}
				return av.Metric.Value < bv.Metric.Value
			}
		} else if query.Descending {
			return a.Run.Manifest.RunID > b.Run.Manifest.RunID
		}
		return a.Run.Manifest.RunID < b.Run.Manifest.RunID
	})
	page = ResultsPage{Revision: revision, Page: query.Page, PageSize: query.PageSize, PageCount: 1, Total: len(rows), DatasetTotal: len(analysis.Runs), Filter: query.Filter, Sort: query.Sort, Descending: query.Descending, MetricNames: names, Rows: []ResultRow{}}
	if query.Revision != "" && query.Revision != revision {
		page.Reset = true
		page.ResetReason = "Results changed; returned to the first page"
		page.Page = 1
	}
	if len(rows) > 0 {
		page.PageCount = 1 + (len(rows)-1)/query.PageSize
	}
	if page.Page > page.PageCount {
		page.Page = page.PageCount
	}
	start := (page.Page - 1) * page.PageSize
	end := start + page.PageSize
	if end > len(rows) {
		end = len(rows)
	}
	for _, run := range rows[start:end] {
		row := ResultRow{RunID: run.Run.Manifest.RunID, Status: run.Run.Result.Status, Feasible: run.Feasible, Pareto: run.Pareto, UnusableReason: run.UnusableReason, Metrics: make([]MetricValue, 0, len(names))}
		for _, name := range names {
			row.Metrics = append(row.Metrics, resultMetric(run, name))
		}
		page.Rows = append(page.Rows, row)
	}
	direction := "ascending"
	if query.Descending {
		direction = "descending"
	}
	page.Caption = fmt.Sprintf("Runs, page %d of %d, sorted by %s %s, filter %s, %d results", page.Page, page.PageCount, page.Sort, direction, page.Filter, page.Total)
	return page, nil
}

// CompareResults selects 2 to 4 distinct immutable IDs in caller order. Metrics
// are rows and selected runs columns. Failed/missing/unsupported values and unit
// mismatches remain explicit unavailable cells, never fabricated zeros.
func CompareResults(analysis experiment.Analysis, ids []string) (ResultsComparison, error) {
	out := ResultsComparison{}
	if len(ids) < 2 || len(ids) > 4 {
		return out, fmt.Errorf("comparison requires 2 to 4 distinct runs")
	}
	revision, ordered, err := resultDataset(analysis)
	if err != nil {
		return out, err
	}
	selected := make([]experiment.AnalyzedRun, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			return out, fmt.Errorf("comparison requires distinct runs")
		}
		seen[id] = true
		found := false
		for _, run := range analysis.Runs {
			if run.Run.Manifest.RunID == id {
				selected = append(selected, run)
				found = true
				break
			}
		}
		if !found {
			return out, fmt.Errorf("selected run is not in the current results")
		}
	}
	out = ResultsComparison{Revision: revision, RunIDs: append([]string(nil), ids...), Metrics: make([]ComparisonMetric, 0, len(ordered))}
	for _, name := range ordered {
		row := ComparisonMetric{Name: name, Availability: "available", Values: make([]MetricValue, 0, len(ids))}
		unit := ""
		mismatch := false
		for _, run := range selected {
			cell := resultMetric(run, name)
			if cell.Presentation.Availability != "available" {
				row.Availability = "unavailable"
				row.Reason = "Metric is unavailable for one or more runs"
			}
			if cell.Metric != nil {
				if unit != "" && unit != cell.Metric.Unit {
					mismatch = true
				}
				unit = cell.Metric.Unit
			}
			row.Values = append(row.Values, cell)
		}
		if mismatch {
			row.Availability = "unavailable"
			row.Reason = "Metric units do not match"
			for i := range row.Values {
				row.Values[i].Presentation = presentation.FormatMetric(experiment.Metric{}, row.Reason)
			}
		}
		out.Metrics = append(out.Metrics, row)
	}
	return out, nil
}

func resultMetric(run experiment.AnalyzedRun, name string) MetricValue {
	cell := MetricValue{Name: name}
	reason := "Metric was not produced"
	if run.Run.Result.Status != experiment.StatusSuccess {
		reason = "Run did not succeed"
	} else {
		for _, metric := range run.Run.Result.Metrics {
			if metric.Name == name {
				copy := metric
				copy.Assumptions = append([]string(nil), metric.Assumptions...)
				cell.Metric = &copy
				cell.Presentation = presentation.FormatMetric(copy, "")
				return cell
			}
		}
	}
	cell.Presentation = presentation.FormatMetric(experiment.Metric{}, reason)
	return cell
}

func resultDataset(analysis experiment.Analysis) (string, []string, error) {
	ordered := analysis
	ordered.Runs = append([]experiment.AnalyzedRun(nil), analysis.Runs...)
	sort.Slice(ordered.Runs, func(i, j int) bool { return ordered.Runs[i].Run.Manifest.RunID < ordered.Runs[j].Run.Manifest.RunID })
	names := map[string]bool{}
	for i, run := range ordered.Runs {
		if run.Run.Manifest.RunID == "" || i > 0 && ordered.Runs[i-1].Run.Manifest.RunID == run.Run.Manifest.RunID {
			return "", nil, fmt.Errorf("results require distinct nonempty run IDs")
		}
		seen := map[string]bool{}
		for _, metric := range run.Run.Result.Metrics {
			if metric.Name == "" || seen[metric.Name] {
				return "", nil, fmt.Errorf("results require distinct metric names")
			}
			seen[metric.Name] = true
			names[metric.Name] = true
		}
	}
	var data bytes.Buffer
	if err := report.WriteJSON(&data, project.Bundle{}, ordered); err != nil {
		return "", nil, fmt.Errorf("results cannot be serialized")
	}
	sum := sha256.Sum256(data.Bytes())
	metrics := make([]string, 0, len(names))
	for name := range names {
		metrics = append(metrics, name)
	}
	sort.Strings(metrics)
	return fmt.Sprintf("%x", sum), metrics, nil
}

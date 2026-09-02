package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/project"
)

type reportRun struct {
	RunID          string                         `json:"run_id"`
	Status         experiment.Status              `json:"status"`
	Design         project.Design                 `json:"design"`
	Result         experiment.Result              `json:"result"`
	Feasible       bool                           `json:"feasible"`
	Pareto         bool                           `json:"pareto"`
	UnusableReason string                         `json:"unusable_reason,omitempty"`
	Constraints    []experiment.ConstraintOutcome `json:"constraints"`
}

type reportData struct {
	Counts map[string]int `json:"counts"`
	Runs   []reportRun    `json:"runs"`
}

func WriteJSON(writer io.Writer, analysis experiment.Analysis) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(projectAnalysis(analysis))
}

func WriteCSV(writer io.Writer, analysis experiment.Analysis) error {
	metrics := metricNames(analysis)
	csvWriter := csv.NewWriter(writer)
	header := append([]string{"run_id", "status", "feasible", "pareto", "unusable_reason"}, metrics...)
	header = append(header, "warnings")
	if err := csvWriter.Write(header); err != nil {
		return err
	}
	for _, analyzed := range analysis.Runs {
		lookup := make(map[string]float64, len(analyzed.Run.Result.Metrics))
		for _, metric := range analyzed.Run.Result.Metrics {
			lookup[metric.Name] = metric.Value
		}
		record := []string{
			analyzed.Run.Manifest.RunID,
			string(analyzed.Run.Result.Status),
			strconv.FormatBool(analyzed.Feasible),
			strconv.FormatBool(analyzed.Pareto),
			analyzed.UnusableReason,
		}
		for _, name := range metrics {
			if value, ok := lookup[name]; ok {
				record = append(record, strconv.FormatFloat(value, 'g', -1, 64))
			} else {
				record = append(record, "")
			}
		}
		record = append(record, strings.Join(analyzed.Run.Result.Warnings, " | "))
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}
	csvWriter.Flush()
	return csvWriter.Error()
}

func WriteMarkdown(writer io.Writer, bundle project.Bundle, analysis experiment.Analysis) error {
	if _, err := fmt.Fprintln(writer, "> Simulation-only, literature-calibrated pre-silicon estimates. Not measured silicon or foundry sign-off."); err != nil {
		return err
	}
	fmt.Fprintf(writer, "\n# %s Trade-off Report\n\n", bundle.Project.Name)
	fmt.Fprintf(writer, "**Hypothesis:** %s\n\n", bundle.Project.Hypothesis)
	fmt.Fprintln(writer, "## Summary")
	fmt.Fprintln(writer)
	for _, key := range []string{"success", "failed", "feasible", "infeasible", "unusable", "pareto"} {
		fmt.Fprintf(writer, "- %s: %d\n", key, analysis.Counts[key])
	}
	fmt.Fprintln(writer, "\n## Objectives")
	fmt.Fprintln(writer)
	for _, objective := range bundle.Project.Objectives {
		fmt.Fprintf(writer, "- `%s`: %s\n", objective.Metric, objective.Direction)
	}
	fmt.Fprintln(writer, "\n## Runs")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "| Run | Status | Feasible | Pareto | Metrics | Evidence | Warnings |")
	fmt.Fprintln(writer, "|---|---|---:|---:|---|---|---|")
	for _, analyzed := range analysis.Runs {
		metricText := make([]string, 0, len(analyzed.Run.Result.Metrics))
		evidence := make([]string, 0, len(analyzed.Run.Result.Metrics))
		for _, metric := range analyzed.Run.Result.Metrics {
			metricText = append(metricText, fmt.Sprintf("%s=%s %s", metric.Name, strconv.FormatFloat(metric.Value, 'g', -1, 64), metric.Unit))
			evidence = append(evidence, string(metric.Evidence))
		}
		fmt.Fprintf(writer, "| `%s` | %s | %t | %t | %s | %s | %s |\n",
			shortID(analyzed.Run.Manifest.RunID), analyzed.Run.Result.Status, analyzed.Feasible, analyzed.Pareto,
			strings.Join(metricText, "<br>"), strings.Join(evidence, "<br>"), strings.Join(analyzed.Run.Result.Warnings, "<br>"))
	}
	return nil
}

func Generate(root string, bundle project.Bundle, analysis experiment.Analysis) error {
	dir := filepath.Join(root, "reports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	outputs := []struct {
		name  string
		write func(io.Writer) error
	}{
		{"results.json", func(w io.Writer) error { return WriteJSON(w, analysis) }},
		{"results.csv", func(w io.Writer) error { return WriteCSV(w, analysis) }},
		{"report.md", func(w io.Writer) error { return WriteMarkdown(w, bundle, analysis) }},
	}
	for _, output := range outputs {
		if err := writeAtomic(dir, output.name, output.write); err != nil {
			return err
		}
	}
	return nil
}

func projectAnalysis(analysis experiment.Analysis) reportData {
	data := reportData{Counts: analysis.Counts, Runs: make([]reportRun, len(analysis.Runs))}
	for index, analyzed := range analysis.Runs {
		data.Runs[index] = reportRun{
			RunID:          analyzed.Run.Manifest.RunID,
			Status:         analyzed.Run.Result.Status,
			Design:         analyzed.Run.Design,
			Result:         analyzed.Run.Result,
			Feasible:       analyzed.Feasible,
			Pareto:         analyzed.Pareto,
			UnusableReason: analyzed.UnusableReason,
			Constraints:    analyzed.Constraints,
		}
	}
	return data
}

func metricNames(analysis experiment.Analysis) []string {
	seen := map[string]struct{}{}
	for _, analyzed := range analysis.Runs {
		for _, metric := range analyzed.Run.Result.Metrics {
			seen[metric.Name] = struct{}{}
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func writeAtomic(dir, name string, write func(io.Writer) error) error {
	file, err := os.CreateTemp(dir, ".tmp-"+name+"-")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer os.Remove(temp)
	if err := write(file); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temp, filepath.Join(dir, name))
}

func shortID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

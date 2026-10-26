package report

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/json"
	"errors"
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
	RunID              string                         `json:"run_id"`
	ModelVersion       string                         `json:"model_version,omitempty"`
	EvaluatorVersion   string                         `json:"evaluator_version"`
	RepositoryRevision string                         `json:"repository_revision,omitempty"`
	Inputs             []project.ResolvedInput        `json:"inputs,omitempty"`
	Status             experiment.Status              `json:"status"`
	Design             project.Design                 `json:"design"`
	Result             experiment.Result              `json:"result"`
	Feasible           bool                           `json:"feasible"`
	Pareto             bool                           `json:"pareto"`
	UnusableReason     string                         `json:"unusable_reason,omitempty"`
	Constraints        []experiment.ConstraintOutcome `json:"constraints"`
}

type reportData struct {
	ModelVersion string         `json:"model_version"`
	Citations    []string       `json:"citations,omitempty"`
	Counts       map[string]int `json:"counts"`
	Runs         []reportRun    `json:"runs"`
}

func WriteJSON(writer io.Writer, bundle project.Bundle, analysis experiment.Analysis) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(projectAnalysis(bundle, analysis))
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
	if _, err := fmt.Fprintln(writer, "> Simplified simulation-only pre-silicon estimates. Not measured silicon or foundry sign-off."); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "\n# %s Trade-off Report\n\n", bundle.Project.Name); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "**Hypothesis:** %s\n\n", bundle.Project.Hypothesis); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "**Model version:** `%s`\n\n", bundle.Project.ModelVersion); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "## Summary"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	for _, key := range []string{"success", "failed", "feasible", "infeasible", "unusable", "pareto"} {
		if _, err := fmt.Fprintf(writer, "- %s: %d\n", key, analysis.Counts[key]); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer, "\n## Objectives"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	for _, objective := range bundle.Project.Objectives {
		if _, err := fmt.Fprintf(writer, "- `%s`: %s\n", objective.Metric, objective.Direction); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer, "\n## Inputs and citations"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	for _, input := range bundle.Inputs {
		if _, err := fmt.Fprintf(writer, "- `%s` — sha256 `%s`; evidence `%s`; citation `%s`\n", input.Path, input.SHA256, input.Evidence, input.Citation); err != nil {
			return err
		}
	}
	for _, citation := range bundle.Project.Citations {
		if _, err := fmt.Fprintf(writer, "- Citation: `%s`\n", citation); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer, "\n## Runs"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "| Run | Status | Feasible | Pareto | Metrics | Evidence | Warnings |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "|---|---|---:|---:|---|---|---|"); err != nil {
		return err
	}
	for _, analyzed := range analysis.Runs {
		metricText := make([]string, 0, len(analyzed.Run.Result.Metrics))
		evidence := make([]string, 0, len(analyzed.Run.Result.Metrics))
		for _, metric := range analyzed.Run.Result.Metrics {
			metricText = append(metricText, fmt.Sprintf("%s=%s %s (model: %s; assumptions: %s)", metric.Name, strconv.FormatFloat(metric.Value, 'g', -1, 64), metric.Unit, metric.Model, strings.Join(metric.Assumptions, ", ")))
			evidence = append(evidence, string(metric.Evidence))
		}
		if _, err := fmt.Fprintf(writer, "| `%s`<br>model `%s`<br>evaluator `%s`<br>revision `%s` | %s | %t | %t | %s | %s | %s |\n",
			shortID(analyzed.Run.Manifest.RunID), analyzed.Run.Manifest.ModelVersion, analyzed.Run.Manifest.EvaluatorVersion, analyzed.Run.Manifest.RepositoryRevision,
			analyzed.Run.Result.Status, analyzed.Feasible, analyzed.Pareto,
			strings.Join(metricText, "<br>"), strings.Join(evidence, "<br>"), strings.Join(analyzed.Run.Result.Warnings, "<br>")); err != nil {
			return err
		}
	}
	return nil
}

// Generate opens and owns a project root, preserving creation of missing parents.
func Generate(root string, bundle project.Bundle, analysis experiment.Analysis) error {
	root = filepath.Clean(root)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	handle, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	defer handle.Close()
	return GenerateRoot(handle, bundle, analysis)
}

// GenerateRoot borrows root and ignores bundle.Root as filesystem authority.
// Each report is replaced separately after sync/close; this is not a transaction
// across all three reports or a power-loss durability guarantee.
func GenerateRoot(root *os.Root, bundle project.Bundle, analysis experiment.Analysis) error {
	if root == nil {
		return errors.New("report root is required")
	}
	if err := root.MkdirAll("reports", 0o755); err != nil {
		return err
	}
	// Pin the output directory as well: retargeting reports while writing must
	// not redirect rename or cleanup to a different directory.
	dir, err := root.OpenRoot("reports")
	if err != nil {
		return err
	}
	defer dir.Close()
	outputs := []struct {
		name  string
		write func(io.Writer) error
	}{
		{"results.json", func(w io.Writer) error { return WriteJSON(w, bundle, analysis) }},
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

func projectAnalysis(bundle project.Bundle, analysis experiment.Analysis) reportData {
	data := reportData{ModelVersion: bundle.Project.ModelVersion, Citations: append([]string(nil), bundle.Project.Citations...), Counts: analysis.Counts, Runs: make([]reportRun, len(analysis.Runs))}
	for index, analyzed := range analysis.Runs {
		data.Runs[index] = reportRun{
			RunID:              analyzed.Run.Manifest.RunID,
			ModelVersion:       analyzed.Run.Manifest.ModelVersion,
			EvaluatorVersion:   analyzed.Run.Manifest.EvaluatorVersion,
			RepositoryRevision: analyzed.Run.Manifest.RepositoryRevision,
			Inputs:             analyzed.Run.Manifest.Inputs,
			Status:             analyzed.Run.Result.Status,
			Design:             analyzed.Run.Design,
			Result:             analyzed.Run.Result,
			Feasible:           analyzed.Feasible,
			Pareto:             analyzed.Pareto,
			UnusableReason:     analyzed.UnusableReason,
			Constraints:        analyzed.Constraints,
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

// Private seam for deterministic precommit faults/interleavings in tests.
// Production always uses the real file method; never change it during use.
var syncReportFile = (*os.File).Sync

func writeAtomic(dir *os.Root, name string, write func(io.Writer) error) error {
	temp := ".tmp-" + name + "-" + rand.Text()
	file, err := dir.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err // Do not remove a path we did not create.
	}
	defer file.Close()
	owned, err := file.Stat()
	committed := false
	defer func() {
		if !committed {
			if owned == nil {
				owned, _ = file.Stat()
			}
			if current, err := dir.Lstat(temp); err == nil && owned != nil && os.SameFile(owned, current) {
				_ = dir.Remove(temp)
			}
		}
	}()
	if err != nil {
		return err
	}
	if err := write(file); err != nil {
		return errors.Join(err, file.Close())
	}
	if err := syncReportFile(file); err != nil {
		return errors.Join(err, file.Close())
	}
	if err := file.Close(); err != nil {
		return err
	}
	current, err := dir.Lstat(temp)
	if err != nil {
		return err
	}
	if !os.SameFile(owned, current) {
		return errors.New("report temporary changed before replacement")
	}
	if err := dir.Rename(temp, name); err != nil {
		return err
	}
	committed = true
	return nil
}

func shortID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

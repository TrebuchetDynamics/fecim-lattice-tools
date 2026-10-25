// Package study coordinates project, experiment, analysis, and report operations.
// CLI and GUI adapters should render its typed outcomes instead of invoking each other.
package study

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"fecim-lattice-tools/workbench/edaexport"
	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/fecim"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/report"
)

type StarterSweepOptions struct {
	ConductanceLevels []int
	Rows              []int
	ADCBits           []int
}

func DefaultStarterSweepOptions() StarterSweepOptions {
	return StarterSweepOptions{ConductanceLevels: []int{16, 30}, Rows: []int{32, 64}, ADCBits: []int{4, 6}}
}

type RunOptions struct {
	Workers            int
	CitationDir        string
	RepositoryRevision string
}

// OutcomeResultsState describes collection completeness, independently of operation errors.
// Consumers must treat zero and unknown values as unavailable.
type OutcomeResultsState string

const (
	ResultsUnavailable OutcomeResultsState = "unavailable" // No trusted result set.
	ResultsPartial     OutcomeResultsState = "partial"     // A trusted subset, possibly empty.
	ResultsComplete    OutcomeResultsState = "complete"    // Every current-sweep point collected.
)

type RunOutcome struct {
	Bundle            project.Bundle
	Summary           experiment.Summary
	Analysis          experiment.Analysis
	ResultsState      OutcomeResultsState
	AnalysisAvailable bool // Partial analysis describes only the collected subset, not the full sweep.
}

func CreateStarterProject(root string, options StarterSweepOptions) error {
	if err := validateStarterValues("conductance levels", options.ConductanceLevels, 2, 64); err != nil {
		return err
	}
	if err := validateStarterValues("array rows", options.Rows, 1, 4096); err != nil {
		return err
	}
	if err := validateStarterValues("ADC bits", options.ADCBits, 1, 16); err != nil {
		return err
	}

	config := project.Project{
		SchemaVersion: 1,
		ID:            "hzo-study",
		Name:          "HZO Device-Array-Circuit Study",
		Hypothesis:    "Increasing ADC resolution trades area and energy for readout fidelity.",
		ModelVersion:  fecim.EvaluatorVersion,
		Objectives:    []project.Objective{{Metric: "energy_pj", Direction: project.Minimize}, {Metric: "latency_ns", Direction: project.Minimize}},
		Constraints:   []project.Constraint{{Metric: "area_um2", Operator: "<=", Value: 5000, Unit: "um2"}},
		Citations:     []string{"park2015_advmat_hzo"},
	}
	design := project.Design{
		SchemaVersion: 1,
		Device:        project.Device{Material: "HZO", ConductanceLevels: 30, GMinS: 1e-6, GMaxS: 30e-6},
		Array:         project.Array{Rows: 32, Cols: 32, ReadVoltageV: 0.2},
		Circuit:       project.Circuit{ADCBits: 6, DACBits: 4, TIAGainOhm: 10000, TechNode: "65nm"},
	}
	sweep := project.Sweep{
		SchemaVersion: 1,
		Seed:          17,
		MaxPoints:     10000,
		Parameters: []project.Parameter{
			{Path: "device.conductance_levels", Values: intValues(options.ConductanceLevels)},
			{Path: "array.rows", Values: intValues(options.Rows)},
			{Path: "circuit.adc_bits", Values: intValues(options.ADCBits)},
		},
	}
	bundle := project.Bundle{Project: config, Design: design, Sweep: sweep}
	if _, err := project.Expand(bundle); err != nil {
		return fmt.Errorf("starter sweep: %w", err)
	}

	contents := make([]struct{ name, body string }, 0, 3)
	for _, item := range []struct {
		name  string
		value any
	}{{"project.yaml", config}, {"design.yaml", design}, {"sweep.yaml", sweep}} {
		data, err := yaml.Marshal(item.value)
		if err != nil {
			return err
		}
		contents = append(contents, struct{ name, body string }{item.name, string(data)})
	}

	info, err := os.Stat(root)
	createdDir := false
	switch {
	case errors.Is(err, os.ErrNotExist):
		if err := os.MkdirAll(root, 0o755); err != nil {
			return err
		}
		createdDir = true
	case err != nil:
		return err
	case !info.IsDir():
		return fmt.Errorf("project path %s is not a directory", root)
	default:
		entries, err := os.ReadDir(root)
		if err != nil {
			return err
		}
		if len(entries) != 0 {
			return fmt.Errorf("project directory %s is not empty", root)
		}
	}
	created := make([]string, 0, len(contents))
	for _, item := range contents {
		path := filepath.Join(root, item.name)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			rollbackStarterProject(created, root, createdDir)
			return err
		}
		_, writeErr := file.WriteString(item.body)
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			removeErr := os.Remove(path)
			rollbackStarterProject(created, root, createdDir)
			return errors.Join(writeErr, closeErr, removeErr)
		}
		created = append(created, path)
	}
	return nil
}

func validateStarterValues(name string, values []int, min, max int) error {
	if len(values) == 0 || len(values) > 16 {
		return fmt.Errorf("%s requires 1 to 16 values", name)
	}
	seen := make(map[int]struct{}, len(values))
	for _, value := range values {
		if value < min || value > max {
			return fmt.Errorf("%s value %d outside [%d,%d]", name, value, min, max)
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("%s contains duplicate value %d", name, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func intValues(values []int) []float64 {
	out := make([]float64, len(values))
	for i, value := range values {
		out[i] = float64(value)
	}
	return out
}

func rollbackStarterProject(paths []string, root string, createdDir bool) {
	for _, path := range paths {
		_ = os.Remove(path)
	}
	if createdDir {
		_ = os.Remove(root)
	}
}

func ValidateProject(root, citationDir string) (project.Bundle, int, error) {
	bundle, err := project.Load(root, project.LoadOptions{CitationDir: citationDir})
	if err != nil {
		return project.Bundle{}, 0, err
	}
	points, err := project.Expand(bundle)
	if err != nil {
		return project.Bundle{}, 0, err
	}
	return bundle, len(points), nil
}

func Run(ctx context.Context, root string, opts RunOptions) (RunOutcome, error) {
	return run(ctx, root, opts, fecim.Evaluate, report.Generate)
}

func run(ctx context.Context, root string, opts RunOptions, evaluate experiment.Evaluator, writeReport func(string, project.Bundle, experiment.Analysis) error) (RunOutcome, error) {
	outcome := RunOutcome{ResultsState: ResultsUnavailable}
	if opts.Workers < 1 {
		return outcome, errors.New("workers must be at least 1")
	}
	bundle, err := project.Load(root, project.LoadOptions{CitationDir: opts.CitationDir})
	if err != nil {
		return outcome, err
	}
	outcome.Bundle = bundle
	points, err := experiment.Expand(bundle)
	if err != nil {
		return outcome, err
	}
	// Snapshot only verified active-sweep IDs. Old or corrupt unrelated runs
	// neither contribute results nor prevent collection of this study.
	ids := make([]string, len(points))
	reused := make(map[string]bool, len(points))
	for i, point := range points {
		id, err := experiment.ID(point, fecim.EvaluatorVersion, bundle.Inputs)
		if err != nil {
			return outcome, err
		}
		_, ok, err := experiment.LoadRunRecord(bundle.Root, id)
		if err != nil {
			return outcome, err
		}
		ids[i] = id
		reused[id] = ok
	}

	summary, runErr := experiment.Run(ctx, bundle, experiment.RunOptions{
		Evaluator:          evaluate,
		EvaluatorVersion:   fecim.EvaluatorVersion,
		Workers:            opts.Workers,
		RepositoryRevision: opts.RepositoryRevision,
	})
	// Run has joined its workers. Its delivered records are independently
	// trusted; reload expected IDs to recover commits lost during cancellation.
	collected := make(map[string]experiment.RunRecord, len(summary.Runs))
	for _, record := range summary.Runs {
		if _, expected := reused[record.Manifest.RunID]; expected {
			collected[record.Manifest.RunID] = record
		}
	}
	var reconciliationErr error
	outcome.Summary.Runs = make([]experiment.RunRecord, 0, len(points))
	for _, id := range ids {
		record, ok, err := experiment.LoadRunRecord(bundle.Root, id)
		if err != nil {
			reconciliationErr = errors.Join(reconciliationErr, err)
		} else if ok {
			record.Reused = reused[id] || collected[id].Reused
			collected[id] = record
		}
		if record, ok := collected[id]; ok {
			record.Reused = record.Reused || reused[id]
			outcome.Summary.Runs = append(outcome.Summary.Runs, record)
		} else if runErr == nil && err == nil {
			reconciliationErr = errors.Join(reconciliationErr, fmt.Errorf("%w: expected run %s is missing after execution", experiment.ErrUnverifiableRun, id))
		}
	}
	if reconciliationErr != nil {
		if len(outcome.Summary.Runs) > 0 {
			outcome.ResultsState = ResultsPartial
		}
		return outcome, errors.Join(runErr, reconciliationErr)
	}
	outcome.ResultsState = ResultsPartial
	if len(outcome.Summary.Runs) == len(points) {
		outcome.ResultsState = ResultsComplete
	}
	outcome.Analysis = experiment.Analyze(bundle, outcome.Summary.Runs)
	outcome.AnalysisAvailable = true
	if runErr != nil {
		return outcome, runErr
	}
	if err := writeReport(bundle.Root, bundle, outcome.Analysis); err != nil {
		return outcome, err
	}
	return outcome, nil
}

func GenerateReport(root, citationDir string) (experiment.Analysis, error) {
	bundle, err := project.Load(root, project.LoadOptions{CitationDir: citationDir})
	if err != nil {
		return experiment.Analysis{}, err
	}
	runs, err := experiment.LoadRuns(bundle.Root)
	if err != nil {
		return experiment.Analysis{}, err
	}
	if len(runs) == 0 {
		return experiment.Analysis{}, errors.New("project has no committed runs")
	}
	analysis := experiment.Analyze(bundle, runs)
	if err := report.Generate(bundle.Root, bundle, analysis); err != nil {
		return experiment.Analysis{}, fmt.Errorf("generate report: %w", err)
	}
	return analysis, nil
}

func ExportRun(root, runID string) (string, error) {
	return edaexport.ExportSelectedRun(root, runID)
}

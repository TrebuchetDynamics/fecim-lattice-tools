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

	// Progress forwards best-effort, nonblocking experiment count snapshots.
	// Nil disables notifications; even the final snapshot may be dropped.
	// The caller owns the channel and must not close it during Run; Run never
	// closes it. The returned outcome/error and committed runs are authoritative.
	Progress chan<- experiment.ProgressSnapshot

	// PointYield is a memory-adapter scheduling hook, called before each point.
	// It may yield to a Worker event loop; it never changes numerical inputs.
	// Native persisted runners do not use this hook.
	PointYield func(context.Context) error
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
	AnalysisAvailable bool   // Partial analysis describes only the collected subset, not the full sweep.
	ResultsJSON       []byte // RunMemory only: original report.WriteJSON bytes for the collected subset.
}

type starterDocument struct{ name, body string }

func starterDocuments(options StarterSweepOptions) ([]starterDocument, error) {
	if err := validateStarterValues("conductance levels", options.ConductanceLevels, 2, 64); err != nil {
		return nil, err
	}
	if err := validateStarterValues("array rows", options.Rows, 1, 4096); err != nil {
		return nil, err
	}
	if err := validateStarterValues("ADC bits", options.ADCBits, 1, 16); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("starter sweep: %w", err)
	}

	contents := make([]starterDocument, 0, 3)
	for _, item := range []struct {
		name  string
		value any
	}{{"project.yaml", config}, {"design.yaml", design}, {"sweep.yaml", sweep}} {
		data, err := yaml.Marshal(item.value)
		if err != nil {
			return nil, err
		}
		contents = append(contents, starterDocument{item.name, string(data)})
	}

	return contents, nil
}

func CreateStarterProject(root string, options StarterSweepOptions) error {
	contents, err := starterDocuments(options)
	if err != nil {
		return err
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
	handle, citations, closeRoots, err := openRoots(root, citationDir)
	if err != nil {
		return project.Bundle{}, 0, err
	}
	defer closeRoots()
	return ValidateProjectRoot(handle, citations)
}

func Run(ctx context.Context, root string, opts RunOptions) (RunOutcome, error) {
	return run(ctx, root, opts, fecim.Evaluate, report.GenerateRoot)
}

func run(ctx context.Context, root string, opts RunOptions, evaluate experiment.Evaluator, writeReport func(*os.Root, project.Bundle, experiment.Analysis) error) (RunOutcome, error) {
	if opts.Workers < 1 {
		return RunOutcome{ResultsState: ResultsUnavailable}, errors.New("workers must be at least 1")
	}
	handle, citations, closeRoots, err := openRoots(root, opts.CitationDir)
	if err != nil {
		return RunOutcome{ResultsState: ResultsUnavailable}, err
	}
	defer closeRoots()
	opts.CitationDir = ""
	return runRoot(ctx, handle, citations, opts, evaluate, writeReport)
}

func GenerateReport(root, citationDir string) (experiment.Analysis, error) {
	handle, citations, closeRoots, err := openRoots(root, citationDir)
	if err != nil {
		return experiment.Analysis{}, err
	}
	defer closeRoots()
	return GenerateReportRoot(handle, citations)
}

// Legacy path adapters own these handles; rooted operations only borrow them.
func openRoots(path, citationDir string) (*os.Root, *os.Root, func(), error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, nil, err
	}
	root, err := os.OpenRoot(abs)
	if err != nil {
		return nil, nil, nil, err
	}
	var citations *os.Root
	if citationDir != "" {
		citations, err = os.OpenRoot(citationDir)
		if err != nil {
			_ = root.Close()
			return nil, nil, nil, err
		}
	}
	return root, citations, func() {
		if citations != nil {
			_ = citations.Close()
		}
		_ = root.Close()
	}, nil
}

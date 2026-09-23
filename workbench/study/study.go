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

type RunOutcome struct {
	Bundle   project.Bundle
	Summary  experiment.Summary
	Analysis experiment.Analysis
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
	if opts.Workers < 1 {
		return RunOutcome{}, errors.New("workers must be at least 1")
	}
	bundle, err := project.Load(root, project.LoadOptions{CitationDir: opts.CitationDir})
	if err != nil {
		return RunOutcome{}, err
	}
	summary, err := experiment.Run(ctx, bundle, experiment.RunOptions{
		Evaluator:          fecim.Evaluate,
		EvaluatorVersion:   fecim.EvaluatorVersion,
		Workers:            opts.Workers,
		RepositoryRevision: opts.RepositoryRevision,
	})
	if err != nil {
		return RunOutcome{}, err
	}
	analysis := experiment.Analyze(bundle, summary.Runs)
	if err := report.Generate(bundle.Root, bundle, analysis); err != nil {
		return RunOutcome{}, err
	}
	return RunOutcome{Bundle: bundle, Summary: summary, Analysis: analysis}, nil
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

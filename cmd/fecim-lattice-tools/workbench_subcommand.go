package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"

	"fecim-lattice-tools/workbench/experiment"
	workbenchfecim "fecim-lattice-tools/workbench/fecim"
	workbenchproject "fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/report"
)

const defaultProjectYAML = `schema_version: 1
id: hzo-study
name: HZO Device-Array-Circuit Study
hypothesis: Increasing ADC resolution trades area and energy for readout fidelity.
model_version: fecim-system-v1
objectives:
  - metric: energy_pj
    direction: minimize
  - metric: latency_ns
    direction: minimize
constraints:
  - metric: area_um2
    operator: <=
    value: 5000
    unit: um2
citations:
  - park2015_advmat_hzo
`

const defaultDesignYAML = `schema_version: 1
device:
  material: HZO
  conductance_levels: 30
  g_min_s: 0.000001
  g_max_s: 0.00003
array:
  rows: 32
  cols: 32
  read_voltage_v: 0.2
circuit:
  adc_bits: 6
  dac_bits: 4
  tia_gain_ohm: 10000
  tech_node: 65nm
`

const defaultSweepYAML = `schema_version: 1
seed: 17
max_points: 32
parameters:
  - path: device.conductance_levels
    values: [16, 30]
  - path: array.rows
    values: [32, 64]
  - path: circuit.adc_bits
    values: [4, 6]
`

func runProjectSubcommand(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("project requires init or validate")
	}
	switch args[0] {
	case "init":
		if len(args) != 2 {
			return errors.New("usage: fecim-lattice-tools project init DIRECTORY")
		}
		if err := initProject(args[1]); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "initialized: %s\n", filepath.Clean(args[1]))
		return nil
	case "validate":
		if len(args) < 2 {
			return errors.New("usage: fecim-lattice-tools project validate DIRECTORY [-citation-dir PATH]")
		}
		fs := flag.NewFlagSet("project validate", flag.ContinueOnError)
		fs.SetOutput(stderr)
		citationDir := fs.String("citation-dir", "", "citation paper-record directory")
		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		bundle, err := workbenchproject.Load(args[1], workbenchproject.LoadOptions{CitationDir: *citationDir})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "valid: %s\n", bundle.Project.ID)
		return nil
	default:
		return fmt.Errorf("unknown project subcommand %q", args[0])
	}
}

func runExperimentSubcommand(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] != "run" {
		if len(args) > 0 {
			return fmt.Errorf("unknown experiment subcommand %q", args[0])
		}
		return errors.New("experiment requires run")
	}
	if len(args) < 2 {
		return errors.New("usage: fecim-lattice-tools experiment run DIRECTORY [-workers N] [-citation-dir PATH]")
	}
	fs := flag.NewFlagSet("experiment run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	workers := fs.Int("workers", 1, "maximum concurrent design points")
	citationDir := fs.String("citation-dir", "", "citation paper-record directory")
	if err := fs.Parse(args[2:]); err != nil {
		return err
	}
	if *workers < 1 {
		return errors.New("workers must be at least 1")
	}
	bundle, err := workbenchproject.Load(args[1], workbenchproject.LoadOptions{CitationDir: *citationDir})
	if err != nil {
		return err
	}
	summary, err := experiment.Run(context.Background(), bundle, experiment.RunOptions{
		Evaluator:          workbenchfecim.Evaluate,
		EvaluatorVersion:   workbenchfecim.EvaluatorVersion,
		Workers:            *workers,
		RepositoryRevision: buildRevision(),
	})
	if err != nil {
		return err
	}
	analysis := experiment.Analyze(bundle, summary.Runs)
	if err := report.Generate(bundle.Root, bundle, analysis); err != nil {
		return err
	}
	reused := 0
	for _, run := range summary.Runs {
		if run.Reused {
			reused++
		}
	}
	fmt.Fprintf(stdout, "total=%d reused=%d failed=%d feasible=%d pareto=%d\n", len(summary.Runs), reused, analysis.Counts["failed"], analysis.Counts["feasible"], analysis.Counts["pareto"])
	return nil
}

func runReportSubcommand(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] != "generate" {
		if len(args) > 0 {
			return fmt.Errorf("unknown report subcommand %q", args[0])
		}
		return errors.New("report requires generate")
	}
	if len(args) < 2 {
		return errors.New("usage: fecim-lattice-tools report generate DIRECTORY [-citation-dir PATH]")
	}
	fs := flag.NewFlagSet("report generate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	citationDir := fs.String("citation-dir", "", "citation paper-record directory")
	if err := fs.Parse(args[2:]); err != nil {
		return err
	}
	bundle, err := workbenchproject.Load(args[1], workbenchproject.LoadOptions{CitationDir: *citationDir})
	if err != nil {
		return err
	}
	runs, err := experiment.LoadRuns(bundle.Root)
	if err != nil {
		return err
	}
	if len(runs) == 0 {
		return errors.New("project has no committed runs")
	}
	analysis := experiment.Analyze(bundle, runs)
	if err := report.Generate(bundle.Root, bundle, analysis); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "generated: %s\n", filepath.Join(bundle.Root, "reports"))
	return nil
}

func initProject(dir string) error {
	info, err := os.Stat(dir)
	createdDir := false
	switch {
	case errors.Is(err, os.ErrNotExist):
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		createdDir = true
	case err != nil:
		return err
	case !info.IsDir():
		return fmt.Errorf("project path %s is not a directory", dir)
	default:
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		if len(entries) != 0 {
			return fmt.Errorf("project directory %s is not empty", dir)
		}
	}

	files := []struct {
		name string
		body string
	}{{"project.yaml", defaultProjectYAML}, {"design.yaml", defaultDesignYAML}, {"sweep.yaml", defaultSweepYAML}}
	created := make([]string, 0, len(files))
	for _, file := range files {
		path := filepath.Join(dir, file.name)
		handle, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			rollbackProjectInit(created, dir, createdDir)
			return err
		}
		_, writeErr := io.WriteString(handle, file.body)
		closeErr := handle.Close()
		if writeErr != nil || closeErr != nil {
			os.Remove(path)
			rollbackProjectInit(created, dir, createdDir)
			if writeErr != nil {
				return writeErr
			}
			return closeErr
		}
		created = append(created, path)
	}
	return nil
}

func rollbackProjectInit(paths []string, dir string, createdDir bool) {
	for _, path := range paths {
		_ = os.Remove(path)
	}
	if createdDir {
		_ = os.Remove(dir)
	}
}

func buildRevision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return ""
}

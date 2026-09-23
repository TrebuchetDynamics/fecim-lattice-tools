package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"runtime/debug"

	"fecim-lattice-tools/workbench/study"
)

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
		_, err := fmt.Fprintf(stdout, "initialized: %s\n", filepath.Clean(args[1]))
		return err
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
		bundle, _, err := study.ValidateProject(args[1], *citationDir)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "valid: %s\n", bundle.Project.ID)
		return err
	default:
		return fmt.Errorf("unknown project subcommand %q", args[0])
	}
}

func runExperimentSubcommand(args []string, stdout, stderr io.Writer) error {
	return runExperimentSubcommandContext(context.Background(), args, stdout, stderr)
}

func runExperimentSubcommandContext(ctx context.Context, args []string, stdout, stderr io.Writer) error {
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
	outcome, err := study.Run(ctx, args[1], study.RunOptions{
		Workers:            *workers,
		CitationDir:        *citationDir,
		RepositoryRevision: buildRevision(),
	})
	if err != nil {
		return err
	}
	summary, analysis := outcome.Summary, outcome.Analysis
	reused := 0
	for _, run := range summary.Runs {
		if run.Reused {
			reused++
		}
	}
	_, err = fmt.Fprintf(stdout, "total=%d reused=%d failed=%d feasible=%d pareto=%d\n", len(summary.Runs), reused, analysis.Counts["failed"], analysis.Counts["feasible"], analysis.Counts["pareto"])
	return err
}

func runExportSubcommand(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: fecim-lattice-tools export DIRECTORY -run RUN_ID")
	}
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(stderr)
	runID := fs.String("run", "", "successful immutable run ID to export")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected export argument %q", fs.Arg(0))
	}
	if *runID == "" {
		return errors.New("-run is required")
	}
	outputDir, err := study.ExportRun(args[0], *runID)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "exported: %s\n", outputDir)
	return err
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
	if _, err := study.GenerateReport(args[1], *citationDir); err != nil {
		return err
	}
	root, err := filepath.Abs(args[1])
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "generated: %s\n", filepath.Join(root, "reports"))
	return err
}

func initProject(dir string) error {
	return study.CreateStarterProject(dir, study.DefaultStarterSweepOptions())
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

// Package main provides a CLI tool for validating FeCIM configuration files.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fecim-lattice-tools/validation/configvalidator"
)

func main() {
	os.Exit(runValidateConfig(os.Args[1:], os.Stdout, os.Stderr))
}

func runValidateConfig(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("validate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var (
		recursive    = flags.Bool("r", false, "Recursively validate all JSON files in directories")
		showWarnings = flags.Bool("w", false, "Show warnings (not just errors)")
		summary      = flags.Bool("s", false, "Show summary only (no individual file results)")
		quiet        = flags.Bool("q", false, "Quiet mode (only exit code)")
	)

	flags.Usage = func() {
		fmt.Fprintf(stderr, "Usage: %s [options] <file.json|directory> ...\n\n", flags.Name())
		fmt.Fprintf(stderr, "Validates FeCIM configuration JSON files.\n\n")
		fmt.Fprintf(stderr, "Options:\n")
		flags.PrintDefaults()
		fmt.Fprintf(stderr, "\nSupported config types:\n")
		fmt.Fprintf(stderr, "  - calibration:    Ferroelectric calibration data\n")
		fmt.Fprintf(stderr, "  - preisach_state: Preisach hysteron states\n")
		fmt.Fprintf(stderr, "  - array_design:   Crossbar array designs\n")
		fmt.Fprintf(stderr, "  - weight_matrix:  Neural network weight matrices\n")
		fmt.Fprintf(stderr, "  - openlane:       OpenLane ASIC flow configs\n")
		fmt.Fprintf(stderr, "\nExamples:\n")
		fmt.Fprintf(stderr, "  %s data/calibrations/fecim_hzo.json\n", flags.Name())
		fmt.Fprintf(stderr, "  %s -r data/\n", flags.Name())
		fmt.Fprintf(stderr, "  %s -w -s .\n", flags.Name())
	}

	if err := flags.Parse(args); err != nil {
		return 2
	}

	if flags.NArg() == 0 {
		flags.Usage()
		return 1
	}

	var allResults []*configvalidator.ValidationResult
	processingError := false

	// Process each argument
	for _, arg := range flags.Args() {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(stderr, "Error: cannot access %s: %v\n", arg, err)
			processingError = true
			continue
		}

		if info.IsDir() {
			if *recursive {
				results, err := configvalidator.ValidateDirectory(arg)
				if err != nil {
					fmt.Fprintf(stderr, "Error validating directory %s: %v\n", arg, err)
					processingError = true
					continue
				}
				allResults = append(allResults, results...)
			} else {
				// Just validate JSON files in the immediate directory
				entries, err := os.ReadDir(arg)
				if err != nil {
					fmt.Fprintf(stderr, "Error reading directory %s: %v\n", arg, err)
					processingError = true
					continue
				}
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
						path := filepath.Join(arg, entry.Name())
						result, err := configvalidator.ValidateFile(path)
						if err != nil {
							fmt.Fprintf(stderr, "Error validating %s: %v\n", path, err)
							processingError = true
							continue
						}
						allResults = append(allResults, result)
					}
				}
			}
		} else {
			result, err := configvalidator.ValidateFile(arg)
			if err != nil {
				fmt.Fprintf(stderr, "Error validating %s: %v\n", arg, err)
				processingError = true
				continue
			}
			allResults = append(allResults, result)
		}
	}

	// Process results
	var totalFiles, validFiles, invalidFiles int
	var totalErrors, totalWarnings int

	for _, result := range allResults {
		totalFiles++
		if result.Valid {
			validFiles++
		} else {
			invalidFiles++
		}
		totalErrors += len(result.Errors)
		totalWarnings += len(result.Warnings)

		// Print individual results unless in quiet or summary mode
		if !*quiet && !*summary {
			if !result.Valid || (*showWarnings && len(result.Warnings) > 0) {
				fmt.Fprintln(stdout, result.String())
				fmt.Fprintln(stdout, strings.Repeat("-", 60))
			}
		}
	}

	// Print summary
	if !*quiet {
		if *summary || totalFiles > 1 {
			fmt.Fprintf(stdout, "\n=== Validation Summary ===\n")
			fmt.Fprintf(stdout, "Total files:  %d\n", totalFiles)
			fmt.Fprintf(stdout, "Valid:        %d\n", validFiles)
			fmt.Fprintf(stdout, "Invalid:      %d\n", invalidFiles)
			fmt.Fprintf(stdout, "Total errors: %d\n", totalErrors)
			if *showWarnings {
				fmt.Fprintf(stdout, "Total warnings: %d\n", totalWarnings)
			}
		}
	}

	// Exit with appropriate code
	if invalidFiles > 0 || processingError {
		return 1
	}
	return 0
}

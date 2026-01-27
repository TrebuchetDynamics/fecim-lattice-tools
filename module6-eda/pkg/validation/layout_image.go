// pkg/validation/layout_image.go
// Layout image generation using KLayout for proper EDA visualization
// Generates PNG images from DEF/LEF files using industry-standard tools

package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fecim-lattice-tools/module6-eda/pkg/openlane"
	"fecim-lattice-tools/shared/logging"
)

// Package-level logger for validation
var log *logging.Logger

func init() {
	log = logging.NewLogger("eda-validation")
}

// LayoutImageResult contains the result of layout image generation
type LayoutImageResult struct {
	Success   bool
	ImagePath string
	RawOutput string
	Error     string
}

// klayoutScript is the Ruby script for KLayout to generate layout image
// Uses -rd variables: $lef_file, $def_file, $output_png
// Pattern from: https://www.klayout.de/forum/discussion/812/get-screenshot-of-a-design-from-command-line
const klayoutScript = `# layout_export.rb - KLayout script to export DEF/LEF as PNG
# Variables passed via -rd: lef_file, def_file, output_png
# Requires -z flag (not -zz) for main window access

puts "=== KLayout Layout Export ==="
puts "LEF: #{$lef_file}"
puts "DEF: #{$def_file}"
puts "Output: #{$output_png}"

# Get main window (requires -z flag)
mw = RBA::Application::instance.main_window

# Load LEF first (cell definitions)
puts "Reading LEF..."
begin
  mw.load_layout($lef_file, 0)
  puts "  LEF loaded"
rescue => e
  puts "  LEF warning: #{e.message}"
end

# Load DEF (placement) into same view
puts "Reading DEF..."
begin
  mw.load_layout($def_file, 1)
  puts "  DEF loaded"
rescue => e
  puts "  DEF error: #{e.message}"
  exit 1
end

# Get current view
view = mw.current_view
if view.nil?
  puts "Error: No view available"
  exit 1
end

# Show full hierarchy and fit to view
view.max_hier
view.zoom_fit

# Configure view for clean export
view.set_config("background-color", "#001020")  # Dark background
view.set_config("grid-visible", "false")        # Hide grid

# Export to PNG (1600x1200 for good resolution)
puts "Exporting PNG..."
view.save_image($output_png, 1600, 1200)
puts "  Saved: #{$output_png}"

puts "=== Layout Export Complete ==="

# Exit cleanly
RBA::Application::instance.exit(0)
`

// GenerateLayoutImage creates a layout image using KLayout
func GenerateLayoutImage(defPath string, lefPath string, outputPath string, manager *openlane.Manager, config *openlane.Config) (*LayoutImageResult, error) {
	result := &LayoutImageResult{
		Success:   false,
		ImagePath: outputPath,
	}

	log.Info("=== KLayout Image Generation ===")
	log.Info("  DEF: %s", defPath)
	log.Info("  LEF: %s", lefPath)
	log.Info("  Output: %s", outputPath)

	// Check if files exist
	if _, err := os.Stat(defPath); os.IsNotExist(err) {
		result.Error = fmt.Sprintf("DEF file not found: %s", defPath)
		log.Printf("KLayout: %s", result.Error)
		return result, nil
	}
	if _, err := os.Stat(lefPath); os.IsNotExist(err) {
		result.Error = fmt.Sprintf("LEF file not found: %s", lefPath)
		log.Printf("KLayout: %s", result.Error)
		return result, nil
	}

	// Check mode
	mode := manager.DetectMode()
	log.Info("  Mode: %s", mode)
	if mode == openlane.ModeNone {
		result.Error = "KLayout not available (install Docker with OpenLane image or native KLayout)"
		log.Printf("KLayout: %s", result.Error)
		return result, nil
	}

	// Create work directory with script
	workDir := filepath.Dir(defPath)
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get absolute path: %v", err)
		return result, nil
	}

	// Write KLayout script
	scriptPath := filepath.Join(absWorkDir, "layout_export.rb")
	if err := os.WriteFile(scriptPath, []byte(klayoutScript), 0644); err != nil {
		result.Error = fmt.Sprintf("failed to write KLayout script: %v", err)
		return result, nil
	}
	defer os.Remove(scriptPath)

	// Copy LEF to work directory if needed
	lefName := filepath.Base(lefPath)
	lefDst := filepath.Join(absWorkDir, lefName)
	if lefPath != lefDst {
		if lefData, err := os.ReadFile(lefPath); err == nil {
			os.WriteFile(lefDst, lefData, 0644)
			defer os.Remove(lefDst)
		}
	}

	// Set up KLayout -rd variables (lowercase with underscores)
	var rdVars map[string]string
	outputName := filepath.Base(outputPath)
	if mode == openlane.ModeDocker {
		rdVars = map[string]string{
			"def_file":   "/design/" + filepath.Base(defPath),
			"lef_file":   "/design/" + lefName,
			"output_png": "/design/" + outputName,
		}
	} else {
		rdVars = map[string]string{
			"def_file":   filepath.Join(absWorkDir, filepath.Base(defPath)),
			"lef_file":   lefDst,
			"output_png": filepath.Join(absWorkDir, outputName),
		}
	}

	// Run KLayout
	log.Info("  Running KLayout...")
	runner := openlane.NewRunner(manager, config)
	runResult, err := runner.RunKLayout(scriptPath, absWorkDir, rdVars)

	if runResult != nil {
		result.RawOutput = runResult.Stdout + "\n" + runResult.Stderr
		// Log output for debugging
		if runResult.Stdout != "" {
			for _, line := range strings.Split(runResult.Stdout, "\n") {
				if line != "" {
					log.Info("  [KLayout stdout] %s", line)
				}
			}
		}
		if runResult.Stderr != "" {
			for _, line := range strings.Split(runResult.Stderr, "\n") {
				if line != "" {
					log.Info("  [KLayout stderr] %s", line)
				}
			}
		}
		log.Info("  [KLayout] Exit code: %d, Duration: %v", runResult.ExitCode, runResult.Duration)
	}

	if err != nil {
		result.Error = fmt.Sprintf("KLayout execution failed: %v", err)
		log.Printf("KLayout error: %v", err)
		log.Printf("KLayout raw output:\n%s", result.RawOutput)
		return result, nil
	}

	// Check if output file was created
	expectedOutput := filepath.Join(absWorkDir, outputName)
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		result.Error = "KLayout did not produce output image"
		log.Printf("KLayout: %s (expected: %s)", result.Error, expectedOutput)
		return result, nil
	}

	// Move to final output path if different
	if expectedOutput != outputPath {
		os.Rename(expectedOutput, outputPath)
	}

	result.Success = true
	result.ImagePath = outputPath
	log.Info("  KLayout image generated: %s", outputPath)
	return result, nil
}

// IsKLayoutAvailable checks if KLayout is available
func IsKLayoutAvailable(manager *openlane.Manager) bool {
	return manager.DetectMode() != openlane.ModeNone
}

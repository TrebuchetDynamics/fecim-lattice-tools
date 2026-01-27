// pkg/validation/layout_image.go
// Layout image generation using KLayout for proper EDA visualization
// Generates PNG images from DEF/LEF files using industry-standard tools

package validation

import (
	"fmt"
	"os"
	"path/filepath"

	"fecim-lattice-tools/module6-eda/pkg/openlane"
)

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

	// Check if files exist
	if _, err := os.Stat(defPath); os.IsNotExist(err) {
		result.Error = fmt.Sprintf("DEF file not found: %s", defPath)
		return result, nil
	}
	if _, err := os.Stat(lefPath); os.IsNotExist(err) {
		result.Error = fmt.Sprintf("LEF file not found: %s", lefPath)
		return result, nil
	}

	// Check mode
	mode := manager.DetectMode()
	if mode == openlane.ModeNone {
		result.Error = "KLayout not available (install Docker with OpenLane image or native KLayout)"
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
	runner := openlane.NewRunner(manager, config)
	runResult, err := runner.RunKLayout(scriptPath, absWorkDir, rdVars)

	if runResult != nil {
		result.RawOutput = runResult.Stdout + "\n" + runResult.Stderr
	}

	if err != nil {
		result.Error = fmt.Sprintf("KLayout execution failed: %v", err)
		return result, nil
	}

	// Check if output file was created
	expectedOutput := filepath.Join(absWorkDir, outputName)
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		result.Error = "KLayout did not produce output image"
		return result, nil
	}

	// Move to final output path if different
	if expectedOutput != outputPath {
		os.Rename(expectedOutput, outputPath)
	}

	result.Success = true
	result.ImagePath = outputPath
	return result, nil
}

// IsKLayoutAvailable checks if KLayout is available
func IsKLayoutAvailable(manager *openlane.Manager) bool {
	return manager.DetectMode() != openlane.ModeNone
}

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
// Uses environment variables: LEF_FILE, DEF_FILE, OUTPUT_PNG
const klayoutScript = `# layout_export.rb - KLayout script to export DEF as PNG
# Environment: LEF_FILE, DEF_FILE, OUTPUT_PNG

require 'fileutils'

lef_file = ENV['LEF_FILE'] || 'fecim_bitcell.lef'
def_file = ENV['DEF_FILE'] || 'placement.def'
output_png = ENV['OUTPUT_PNG'] || 'layout.png'

puts "=== KLayout Layout Export ==="
puts "LEF: #{lef_file}"
puts "DEF: #{def_file}"
puts "Output: #{output_png}"

# Create layout view
main_window = RBA::Application.instance.main_window
layout_view = main_window.create_layout(1)
layout = layout_view.active_cellview.layout

# Set up DEF/LEF reader options
opt = RBA::LoadLayoutOptions.new
lef_map = RBA::LEFDEFReaderConfiguration.new
opt.set_lefdef_config(lef_map)

# Read LEF first (cell definitions)
puts "Reading LEF..."
begin
  layout.read(lef_file, opt)
  puts "  LEF loaded successfully"
rescue => e
  puts "  LEF warning: #{e.message}"
end

# Read DEF (placement)
puts "Reading DEF..."
begin
  layout.read(def_file, opt)
  puts "  DEF loaded successfully"
rescue => e
  puts "  DEF error: #{e.message}"
  exit 1
end

# Get the top cell
top_cell = layout.top_cell
if top_cell.nil?
  puts "Error: No top cell found"
  exit 1
end
puts "Top cell: #{top_cell.name}"

# Configure view for export
layout_view.select_cell(top_cell.cell_index, 0)
layout_view.zoom_fit

# Set up layers with colors
layer_colors = {
  'met1' => 0x4444ff,  # Blue for metal1
  'met2' => 0xff4444,  # Red for metal2
  'via1' => 0x44ff44,  # Green for via1
  'nwell' => 0xffff44, # Yellow for nwell
  'pwell' => 0xff44ff, # Magenta for pwell
}

# Apply colors to layers
layout.layer_indices.each do |li|
  info = layout.get_info(li)
  layer_name = info.name.downcase
  layer_colors.each do |pattern, color|
    if layer_name.include?(pattern)
      lp = layout_view.find_layer_iter(li, 0).current
      lp.fill_color = color if lp
      lp.frame_color = color if lp
    end
  end
end

# Export to PNG
puts "Exporting PNG..."
layout_view.save_image(output_png, 1200, 900)
puts "  Saved: #{output_png}"

puts "=== Layout Export Complete ==="
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

	// Set up environment variables
	var envVars map[string]string
	outputName := filepath.Base(outputPath)
	if mode == openlane.ModeDocker {
		envVars = map[string]string{
			"DEF_FILE":   "/design/" + filepath.Base(defPath),
			"LEF_FILE":   "/design/" + lefName,
			"OUTPUT_PNG": "/design/" + outputName,
		}
	} else {
		envVars = map[string]string{
			"DEF_FILE":   filepath.Join(absWorkDir, filepath.Base(defPath)),
			"LEF_FILE":   lefDst,
			"OUTPUT_PNG": filepath.Join(absWorkDir, outputName),
		}
	}

	// Run KLayout
	runner := openlane.NewRunner(manager, config)
	runResult, err := runner.RunKLayout(scriptPath, absWorkDir, envVars)

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

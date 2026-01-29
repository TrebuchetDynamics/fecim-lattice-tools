// Package gui provides CLI calibration functionality for hysteresis module.
package gui

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

// CLICalibrationOptions configures CLI calibration behavior
type CLICalibrationOptions struct {
	MaterialName string  // Material to calibrate (empty = all materials)
	NumLevels    int     // Number of discrete levels (default: 30)
	Temperature  float64 // Temperature in Kelvin (default: 300)
	Force        bool    // Force recalibration even if file exists
	Verbose      bool    // Print progress messages
}

// RunCLICalibration performs calibration without GUI and saves results to file.
// Returns nil on success, error on failure.
func RunCLICalibration(opts CLICalibrationOptions) error {
	// Set defaults (NumLevels=0 means use material's native level count)
	if opts.Temperature == 0 {
		opts.Temperature = 300
	}

	materials := ferroelectric.AllMaterials()
	var materialsToCalibrate []*ferroelectric.HZOMaterial

	if opts.MaterialName == "" || opts.MaterialName == "all" {
		// Calibrate all materials
		materialsToCalibrate = materials
	} else {
		// Find specific material
		for _, m := range materials {
			if m.Name == opts.MaterialName {
				materialsToCalibrate = []*ferroelectric.HZOMaterial{m}
				break
			}
		}
		if len(materialsToCalibrate) == 0 {
			return fmt.Errorf("material not found: %s\nAvailable materials: %v", opts.MaterialName, getMaterialNames(materials))
		}
	}

	// Ensure calibration directory exists
	if err := os.MkdirAll(calibrationDir, 0755); err != nil {
		return fmt.Errorf("failed to create calibration directory: %w", err)
	}

	for _, mat := range materialsToCalibrate {
		calibFile := calibrationFileForMaterial(mat.Name)

		// Skip if file exists and not forcing
		if !opts.Force {
			if _, err := os.Stat(calibFile); err == nil {
				if opts.Verbose {
					fmt.Printf("Skipping %s (calibration exists: %s)\n", mat.Name, calibFile)
				}
				continue
			}
		}

		// Calculate actual level count for this material
		actualLevels := mat.GetNumLevels()
		if opts.NumLevels > 0 {
			actualLevels = opts.NumLevels
		}

		if opts.Verbose {
			fmt.Printf("Calibrating %s at %.0fK with %d levels...\n", mat.Name, opts.Temperature, actualLevels)
		}

		start := time.Now()
		if err := calibrateMaterial(mat, opts); err != nil {
			return fmt.Errorf("calibration failed for %s: %w", mat.Name, err)
		}

		if opts.Verbose {
			fmt.Printf("  Saved: %s (%.2fs)\n", calibFile, time.Since(start).Seconds())
		}
	}

	return nil
}

// calibrateMaterial performs calibration for a single material
func calibrateMaterial(mat *ferroelectric.HZOMaterial, opts CLICalibrationOptions) error {
	// Use material's native level count, or override if explicitly specified
	numLevels := mat.GetNumLevels()
	if opts.NumLevels > 0 {
		numLevels = opts.NumLevels
	}
	tempK := opts.Temperature

	// Create Preisach model with high resolution
	preisachGridSize := 200
	preisach := ferroelectric.NewMayergoyzPreisach(mat, preisachGridSize)
	preisach.SetTemperature(tempK)

	// Get temperature-corrected Ec
	Ec := preisach.GetEffectiveEc()
	if Ec == 0 {
		Ec = mat.Ec
	}
	Emax := 1.5 * Ec
	Ps := mat.Ps

	maxLevel := numLevels - 1
	if maxLevel < 1 {
		maxLevel = 1
	}

	// Initialize calibration arrays
	calibrationUp := make([]float64, numLevels)
	calibrationDown := make([]float64, numLevels)
	calibUpLow := make([]float64, numLevels)
	calibUpHigh := make([]float64, numLevels)
	calibDownLow := make([]float64, numLevels)
	calibDownHigh := make([]float64, numLevels)
	lastErrorUp := make([]int, numLevels)
	lastErrorDown := make([]int, numLevels)
	relaxCompUp := make([]float64, numLevels)
	relaxCompDown := make([]float64, numLevels)

	// Initialize bounds
	for i := 0; i < numLevels; i++ {
		calibUpLow[i] = Ec * 0.3
		calibUpHigh[i] = Ec * 2.0
		calibDownLow[i] = -Ec * 2.0
		calibDownHigh[i] = -Ec * 0.3

		// Initialize relaxation compensation with parabolic profile
		normalizedPos := float64(i) / float64(maxLevel)
		relaxCompUp[i] = 0.05 * 4 * normalizedPos * (1 - normalizedPos)
		relaxCompDown[i] = 0.05 * 4 * normalizedPos * (1 - normalizedPos)
	}

	// Calibrate ascending (from -Ps to each level)
	for level := 1; level < numLevels; level++ {
		targetP := -Ps + float64(level)/float64(maxLevel)*2*Ps

		// Binary search for field
		lowE := Ec * 0.3
		highE := Ec * 2.0
		var bestE float64

		for iter := 0; iter < 20; iter++ {
			midE := (lowE + highE) / 2

			// Reset to -Ps
			preisach.Reset()
			preisach.Update(-Emax)
			preisach.Update(0)

			// Apply test field
			preisach.Update(midE)
			preisach.Update(0)

			P := preisach.Polarization()
			bestE = midE

			if P < targetP {
				lowE = midE
			} else {
				highE = midE
			}

			if highE-lowE < Ec*0.01 {
				break
			}
		}

		calibrationUp[level] = bestE
		calibUpLow[level] = lowE
		calibUpHigh[level] = highE
	}

	// Calibrate descending (from +Ps to each level)
	for level := numLevels - 2; level >= 0; level-- {
		targetP := Ps - float64(numLevels-1-level)/float64(maxLevel)*2*Ps

		// Binary search for field (negative)
		lowE := -Ec * 2.0
		highE := -Ec * 0.3
		var bestE float64

		for iter := 0; iter < 20; iter++ {
			midE := (lowE + highE) / 2

			// Reset to +Ps
			preisach.Reset()
			preisach.Update(Emax)
			preisach.Update(0)

			// Apply test field
			preisach.Update(midE)
			preisach.Update(0)

			P := preisach.Polarization()
			bestE = midE

			if P > targetP {
				highE = midE
			} else {
				lowE = midE
			}

			if highE-lowE < Ec*0.01 {
				break
			}
		}

		calibrationDown[level] = bestE
		calibDownLow[level] = lowE
		calibDownHigh[level] = highE
	}

	// Build calibration data structure
	tempKRounded := int(math.Round(tempK))
	calData := &CalibrationData{
		Version:      calibrationVersion,
		MaterialName: mat.Name,
		NumLevels:    numLevels,
		Calibrations: map[int]*TempCalibration{
			tempKRounded: {
				Temperature:     tempK,
				CalibrationUp:   calibrationUp,
				CalibrationDown: calibrationDown,
				CalibUpLow:      calibUpLow,
				CalibUpHigh:     calibUpHigh,
				CalibDownLow:    calibDownLow,
				CalibDownHigh:   calibDownHigh,
				LastErrorUp:     lastErrorUp,
				LastErrorDown:   lastErrorDown,
				RelaxCompUp:     relaxCompUp,
				RelaxCompDown:   relaxCompDown,
			},
		},
		SavedAt: time.Now().Format(time.RFC3339),
	}

	// Save to file
	return saveCalibrationData(mat.Name, calData)
}

// saveCalibrationData writes calibration to JSON file
func saveCalibrationData(materialName string, data *CalibrationData) error {
	filePath := calibrationFileForMaterial(materialName)

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// getMaterialNames returns list of material names for error messages
func getMaterialNames(materials []*ferroelectric.HZOMaterial) []string {
	names := make([]string, len(materials))
	for i, m := range materials {
		names[i] = m.Name
	}
	return names
}

// ListMaterials returns available material names for CLI help
func ListMaterials() []string {
	return getMaterialNames(ferroelectric.AllMaterials())
}

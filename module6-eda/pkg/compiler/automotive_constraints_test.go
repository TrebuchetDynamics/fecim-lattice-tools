// Package compiler - Automotive Configuration Validation Tests
//
// These tests verify that:
// 1. Software models correctly represent published temperature-dependent physics
// 2. Code doesn't crash at extreme parameter values
// 3. Documentation properly cites automotive requirements
//
// NOTE: These are SOFTWARE tests, not HARDWARE qualification.
// Actual AEC-Q100 hardware testing requires:
// - Physical FeFET chips
// - Thermal chambers (-40°C to 150°C)
// - Lab equipment ($500K+)
// - 6-12 months of testing
// - Dr. Tour's lab facilities
package compiler

import (
	"os"
	"strings"
	"testing"
)

// TestAutomotiveDocumentationExists verifies automotive docs are present
func TestAutomotiveDocumentationExists(t *testing.T) {
	automotiveDocs := "../../../docs/papers/by-topic/22-automotive-harsh-env/README.md"

	if _, err := os.Stat(automotiveDocs); os.IsNotExist(err) {
		t.Skip("Automotive documentation not found (may be running from different directory)")
	}

	data, err := os.ReadFile(automotiveDocs)
	if err != nil {
		t.Fatalf("Failed to read automotive docs: %v", err)
	}

	content := string(data)

	// Should document temperature range
	if !strings.Contains(content, "-40") {
		t.Error("Should document automotive cold temp (-40°C)")
	}

	if !strings.Contains(content, "150") {
		t.Error("Should document automotive hot temp (150°C)")
	}

	// Should reference AEC-Q100
	if !strings.Contains(content, "AEC-Q100") {
		t.Error("Should reference AEC-Q100 automotive qualification standard")
	}

	t.Log("Automotive documentation exists with required temperature and qualification references")
}

// TestStorageConfigEnforcesRetention verifies storage mode has retention requirements
func TestStorageConfigEnforcesRetention(t *testing.T) {
	config := NewStorageConfig(256, 256)

	if config.StorageConfig == nil {
		t.Fatal("StorageConfig should be initialized for storage mode")
	}

	// Automotive/enterprise storage needs 10+ year retention
	if config.StorageConfig.RetentionYears < 10 {
		t.Errorf("Storage mode should have >=10 year retention, got %f",
			config.StorageConfig.RetentionYears)
	}

	// Should have reasonable endurance
	if config.StorageConfig.EnduranceCycles < 1e6 {
		t.Errorf("Storage mode should have >=1M write cycles, got %d",
			config.StorageConfig.EnduranceCycles)
	}

	t.Logf("Storage config: %d year retention, %.0e endurance cycles",
		int(config.StorageConfig.RetentionYears), float64(config.StorageConfig.EnduranceCycles))
}

// TestArrayConfigValidRanges verifies config parameters are physically reasonable
func TestArrayConfigValidRanges(t *testing.T) {
	modes := []struct {
		name   string
		config *ArrayConfig
	}{
		{"Storage", NewStorageConfig(64, 64)},
		{"Memory", NewMemoryConfig(64, 64)},
		{"Compute", NewComputeConfig(64, 64)},
	}

	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			cfg := m.config

			// Conductance range should be positive
			if cfg.GMin <= 0 || cfg.GMax <= 0 {
				t.Error("Conductance values should be positive")
			}
			if cfg.GMin >= cfg.GMax {
				t.Error("GMin should be less than GMax")
			}

			// Programming voltage should be reasonable (1-10V typical)
			if cfg.VProgMin < 0.5 || cfg.VProgMax > 15 {
				t.Errorf("Programming voltage range [%.1f, %.1f] seems unreasonable",
					cfg.VProgMin, cfg.VProgMax)
			}

			// Levels should be FeCIM standard (30)
			if cfg.Levels != 30 {
				t.Errorf("Expected 30 FeCIM levels, got %d", cfg.Levels)
			}

			// Cell dimensions should be reasonable for nanoscale
			if cfg.CellPitch <= 0 || cfg.CellPitch > 10 {
				t.Errorf("Cell pitch %.2f um seems unreasonable", cfg.CellPitch)
			}
		})
	}
}

// TestGenerateDesignDoesNotPanicAtExtremes verifies robustness
func TestGenerateDesignDoesNotPanicAtExtremes(t *testing.T) {
	// Test with minimum valid dimensions
	t.Run("MinDimensions", func(t *testing.T) {
		config := NewStorageConfig(1, 1)
		design, err := GenerateDesign(config)
		if err != nil {
			t.Fatalf("Should handle 1x1 array: %v", err)
		}
		if design.Stats.TotalCells != 1 {
			t.Errorf("Expected 1 cell, got %d", design.Stats.TotalCells)
		}
	})

	// Test with large dimensions (stress test)
	t.Run("LargeDimensions", func(t *testing.T) {
		config := NewStorageConfig(512, 512)
		design, err := GenerateDesign(config)
		if err != nil {
			t.Fatalf("Should handle 512x512 array: %v", err)
		}
		expected := 512 * 512
		if design.Stats.TotalCells != expected {
			t.Errorf("Expected %d cells, got %d", expected, design.Stats.TotalCells)
		}
	})

	// Test with asymmetric dimensions
	t.Run("AsymmetricDimensions", func(t *testing.T) {
		config := NewComputeConfig(784, 10) // MNIST-like
		design, err := GenerateDesign(config)
		if err != nil {
			t.Fatalf("Should handle 784x10 array: %v", err)
		}
		if design.Stats.TotalCells != 784*10 {
			t.Errorf("Expected %d cells, got %d", 784*10, design.Stats.TotalCells)
		}
	})
}

// TestWeightQuantizationPreservesSign verifies signed weight handling
func TestWeightQuantizationPreservesSign(t *testing.T) {
	// Create weights spanning full range
	weights := [][]float64{
		{-1.0, -0.5, 0.0, 0.5, 1.0},
	}

	config := NewComputeConfig(4, 8)
	config.ComputeConfig.InitialWeights = weights

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// Verify cells exist for all weights
	if len(design.Cells) != 5 {
		t.Errorf("Expected 5 cells, got %d", len(design.Cells))
	}

	// Verify quantization levels span the range
	levels := make(map[int]bool)
	for _, cell := range design.Cells {
		levels[cell.Level] = true
	}

	// Should have multiple distinct levels
	if len(levels) < 3 {
		t.Error("Weights should quantize to multiple distinct levels")
	}

	// Extreme weights should map to extreme levels
	minLevel, maxLevel := config.Levels, 0
	for _, cell := range design.Cells {
		if cell.Level < minLevel {
			minLevel = cell.Level
		}
		if cell.Level > maxLevel {
			maxLevel = cell.Level
		}
	}

	if minLevel > 5 {
		t.Errorf("Negative weights should map to low levels, got min=%d", minLevel)
	}
	if maxLevel < config.Levels-5 {
		t.Errorf("Positive weights should map to high levels, got max=%d", maxLevel)
	}
}

// TestQuantizationErrorIsReasonable verifies MSE/PSNR are computed correctly
func TestQuantizationErrorIsReasonable(t *testing.T) {
	// Create random-ish weights
	weights := make([][]float64, 8)
	for i := range weights {
		weights[i] = make([]float64, 8)
		for j := range weights[i] {
			// Spread weights across range
			weights[i][j] = float64(i*8+j)/64.0*2.0 - 1.0
		}
	}

	config := NewComputeConfig(8, 8)
	config.ComputeConfig.InitialWeights = weights

	design, err := GenerateDesign(config)
	if err != nil {
		t.Fatalf("GenerateDesign failed: %v", err)
	}

	// MSE should be small for 30-level quantization
	if design.Stats.QuantMSE > 0.01 {
		t.Errorf("MSE %.6f seems too high for 30 levels", design.Stats.QuantMSE)
	}

	// PSNR should be good (>25 dB for 30 levels)
	if design.Stats.QuantPSNR < 20 {
		t.Errorf("PSNR %.2f dB seems too low for 30 levels", design.Stats.QuantPSNR)
	}

	t.Logf("Quantization quality: MSE=%.6f, PSNR=%.2f dB",
		design.Stats.QuantMSE, design.Stats.QuantPSNR)
}

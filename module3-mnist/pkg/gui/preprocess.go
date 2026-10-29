// Package gui provides Fyne-based GUI components for MNIST visualization.
package gui

import "fecim-lattice-tools/shared/neural"

// InputSource tracks where the current digit input came from.
type InputSource int

const (
	InputProgrammatic InputSource = iota
	InputUser
)

const (
	preprocessThreshold = 0.05
	preprocessTarget    = 20
	preprocessSize      = 28
)

func preprocessIfUserInput(dc *DigitCanvas, pixels []float64) []float64 {
	if dc == nil || dc.LastInputSource() != InputUser {
		return pixels
	}
	return PreprocessDigit(pixels)
}

// PreprocessDigit shares the unchanged native drawing normalization with web jobs.
func PreprocessDigit(pixels []float64) []float64 { return neural.PreprocessDigit(pixels) }

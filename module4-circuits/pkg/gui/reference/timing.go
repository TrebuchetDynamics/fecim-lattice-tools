//go:build legacy_fyne

// Package reference contains compatibility wrappers for module 4 reference-tab content.
package reference

import "fecim-lattice-tools/module4-circuits/pkg/gui/reference/timeline"

// TimingAnimationSteps returns the status messages used to animate a timing diagram.
func TimingAnimationSteps(operation string) []string {
	return timeline.AnimationSteps(operation)
}

//go:build legacy_fyne

package comparison

// AnimationSteps returns the status messages for the CPU/GPU/FeFET comparison animation.
func AnimationSteps() []string {
	return []string{
		"Step 1: CPU loads data from DRAM (250ns)...",
		"Step 2: CPU computes MVM (250ns)...",
		"Step 3: GPU loads data from HBM (25ns)...",
		"Step 4: GPU computes MVM (25ns)...",
		"Step 5: FeFET performs in-memory compute (76ns)...",
		"Animation complete: FeFET ≈6.6x faster than CPU (latency model)",
	}
}

// NextScaleSize cycles through supported comparison array sizes.
func NextScaleSize(current int) int {
	sizes := []int{8, 16, 32, 64}
	for i, size := range sizes {
		if size == current {
			return sizes[(i+1)%len(sizes)]
		}
	}
	return sizes[0]
}

//go:build legacy_fyne

// Package reference contains pure helpers for module 4 reference-tab content.
package reference

// TimingAnimationSteps returns the status messages used to animate a timing diagram.
func TimingAnimationSteps(operation string) []string {
	switch operation {
	case "WRITE":
		return []string{
			"Phase 1: DAC settle (0-10ns)...",
			"Phase 2: Charge pump rise (10-98ns)...",
			"Phase 3: V_PROG write pulse (98-198ns)...",
			"Phase 4: Array settle (198-203ns)...",
			"Phase 5: DONE asserted (203ns)...",
			"Write complete: Total 203ns",
		}
	case "READ":
		return []string{
			"Phase 1: DAC settle (0-10ns)...",
			"Phase 2: Array settle (10-15ns)...",
			"Phase 3: TIA settle (15-26ns)...",
			"Phase 4: ADC convert (26-76ns)...",
			"Phase 5: DATA_OUT valid (76ns)...",
			"Read complete: Total 76ns",
		}
	case "COMPUTE":
		return []string{
			"Phase 1: INPUT_VALID asserted (0ns)...",
			"Phase 2: DAC_ALL converts inputs (0-10ns)...",
			"Phase 3: ARRAY_SETTLE (10-15ns)...",
			"Phase 4: TIA+ADC digitizes summed currents (15-76ns)...",
			"Phase 5: OUTPUT_VALID - MVM result ready (76ns)...",
			"Compute complete: Total 76ns for full MVM",
		}
	default:
		return []string{"Select an operation to animate"}
	}
}

package arraysim

// SenseMeasurementPreset is the native GUI's measurement preset. Rf is in
// kiloohms, references in volts. Return a fresh slice so callers cannot mutate
// the scientific catalogue for another session.
type SenseMeasurementPreset struct {
	Name           string
	Rf, Vmin, Vmax float64
}

func NativeSensePresets() []SenseMeasurementPreset {
	return []SenseMeasurementPreset{
		{"Balanced", 10, 0, 1},
		{"High Sensitivity", 50, 0, 1},
		{"Wide Current Range", 5, 0, 1.2},
		{"Low-Current Focus", 100, 0.1, 0.9},
		{"Ultra-Low Current", 500, 0.2, 0.8},
	}
}

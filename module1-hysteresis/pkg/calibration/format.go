// Package calibration defines the existing legacy calibration JSON format.
// File access and model validation belong to the caller.
package calibration

// TempCalibration holds calibration data for a specific temperature
type TempCalibration struct {
	Temperature     float64   `json:"temperature_k"`    // Temperature in Kelvin
	CalibrationUp   []float64 `json:"calibration_up"`   // Ascending calibration values
	CalibrationDown []float64 `json:"calibration_down"` // Descending calibration values
	CalibUpLow      []float64 `json:"calib_up_low"`     // Binary search lower bounds (ascending)
	CalibUpHigh     []float64 `json:"calib_up_high"`    // Binary search upper bounds (ascending)
	CalibDownLow    []float64 `json:"calib_down_low"`   // Binary search lower bounds (descending)
	CalibDownHigh   []float64 `json:"calib_down_high"`  // Binary search upper bounds (descending)
	LastErrorUp     []int     `json:"last_error_up"`    // Last error for oscillation detection
	LastErrorDown   []int     `json:"last_error_down"`  // Last error for oscillation detection
	RelaxCompUp     []float64 `json:"relax_comp_up"`    // Relaxation compensation factors (ascending)
	RelaxCompDown   []float64 `json:"relax_comp_down"`  // Relaxation compensation factors (descending)
}

// CalibrationData holds persistent calibration state (v2+: multi-temperature support)
type CalibrationData struct {
	Version         int                      `json:"version"`           // Schema version (4 = Everett clamp fix)
	MaterialName    string                   `json:"material_name"`     // Material these calibrations are for
	NumLevels       int                      `json:"num_levels"`        // Number of discrete levels
	TargetRangeFrac float64                  `json:"target_range_frac"` // Effective Ps range used for level mapping
	Calibrations    map[int]*TempCalibration `json:"calibrations"`      // Key: temperature in Kelvin (rounded)
	SavedAt         string                   `json:"saved_at"`          // Timestamp

	// Legacy v1 fields (for migration, not used in v2)
	CalibrationUp   []float64 `json:"calibration_up,omitempty"`
	CalibrationDown []float64 `json:"calibration_down,omitempty"`
	CalibUpLow      []float64 `json:"calib_up_low,omitempty"`
	CalibUpHigh     []float64 `json:"calib_up_high,omitempty"`
	CalibDownLow    []float64 `json:"calib_down_low,omitempty"`
	CalibDownHigh   []float64 `json:"calib_down_high,omitempty"`
	LastErrorUp     []int     `json:"last_error_up,omitempty"`
	LastErrorDown   []int     `json:"last_error_down,omitempty"`
	RelaxCompUp     []float64 `json:"relax_comp_up,omitempty"`
	RelaxCompDown   []float64 `json:"relax_comp_down,omitempty"`
}

// pkg/config/types.go
package config

// CellConfig defines configuration for a single FeCIM bitcell
type CellConfig struct {
	Name         string  // e.g., "fecim_bitcell"
	Width        float64 // Cell width in μm (e.g., 0.46)
	Height       float64 // Cell height in μm (e.g., 2.72)
	CellType     string  // "passive" or "1t1r"
	Technology   string  // e.g., "sky130"
	
	// Timing parameters (placeholder values for Liberty generation)
	RiseTime     float64 // Rise time in ns (e.g., 0.1)
	FallTime     float64 // Fall time in ns (e.g., 0.1)
	InputCap     float64 // Input capacitance in pF (e.g., 0.002)
	LeakagePower float64 // Leakage power in nW (e.g., 0.001)
}

// ArrayConfig defines configuration for a FeCIM crossbar array
type ArrayConfig struct {
	Rows         int     // Number of rows (e.g., 4, 8, 16, 32)
	Cols         int     // Number of columns
	Mode         string  // "storage", "memory", or "compute"
	Architecture string  // "passive" or "1t1r"
	Technology   string  // e.g., "sky130"
	CellWidth    float64 // From CellConfig, in μm
	CellHeight   float64 // From CellConfig, in μm
}

// DefaultCellConfig returns a default cell configuration for FeCIM bitcell
func DefaultCellConfig() CellConfig {
	return CellConfig{
		Name:         "fecim_bitcell",
		Width:        0.46,
		Height:       2.72,
		CellType:     "passive",
		Technology:   "sky130",
		RiseTime:     0.1,
		FallTime:     0.1,
		InputCap:     0.002,
		LeakagePower: 0.001,
	}
}

// DefaultArrayConfig returns a default array configuration
func DefaultArrayConfig() ArrayConfig {
	return ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}
}

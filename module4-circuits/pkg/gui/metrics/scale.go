//go:build legacy_fyne

package metrics

import (
	"fmt"
	"math"
)

// ScaledUnit describes a display unit and its conversion scale from the base unit.
type ScaledUnit struct {
	Unit  string
	Scale float64
}

// FormatSignedScaled formats a signed value using the largest applicable display unit.
func FormatSignedScaled(value float64, units []ScaledUnit) string {
	absValue := math.Abs(value)
	if absValue < 1e-12 {
		return fmt.Sprintf("0 %s", units[0].Unit)
	}

	chosen := units[len(units)-1]
	for _, candidate := range units {
		if absValue >= candidate.Scale {
			chosen = candidate
			break
		}
	}
	scaled := value / chosen.Scale
	absScaled := math.Abs(scaled)
	format := "%+.3f"
	switch {
	case absScaled >= 100:
		format = "%+.0f"
	case absScaled >= 10:
		format = "%+.1f"
	case absScaled >= 1:
		format = "%+.2f"
	}
	return fmt.Sprintf(format+" %s", scaled, chosen.Unit)
}

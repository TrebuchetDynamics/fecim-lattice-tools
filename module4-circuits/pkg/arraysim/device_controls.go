package arraysim

import (
	physics "fecim-lattice-tools/shared/physics"
	"math"
)

// Native GUI technology-node selector assumptions (ohms, siemens).
func NodeSelectorParams(node int) (float64, float64) {
	switch node {
	case 14:
		return 0.8e3, 5e-11
	case 28:
		return 1.5e3, 2e-11
	case 65:
		return 3.5e3, 5e-12
	}
	return 7e3, 1e-12
}

// DeviceVoltageLimits preserves embedded calibration ratios and practical caps.
func DeviceVoltageLimits(mat *physics.HZOMaterial) (readMax, writeMax float64) {
	vc := mat.CoerciveVoltage()
	readMax = 0.7 * vc
	if readMax > 1 {
		readMax = 1
	}
	if readMax < 0.1 {
		readMax = 0.1
	}
	writeMax = 2.5 * vc
	if writeMax > 3 {
		writeMax = 3
	}
	return
}
func SafeDeviceReadVoltage(readMax float64) float64 {
	v := readMax * 0.4
	if v < 0.1 {
		return 0.2
	}
	return v
}
func QuantizeDeviceWrite(s DeviceSettings, target float64) (float64, int) {
	d := s.DAC
	if d == nil {
		return target, -1
	}
	target = math.Max(d.VrefLow, math.Min(d.VrefHigh, target))
	span := d.VrefHigh - d.VrefLow
	if span == 0 {
		return target, 0
	}
	code := int(math.Round((target - d.VrefLow) / span * float64(d.Levels()-1)))
	if s.DACNonlinearity {
		return d.ConvertWithCondition(code, s.Temperature, s.Corner), code
	}
	return d.Convert(code), code
}
func ClampProgramLevel(next, current, target int, direction physics.HysteresisDirection) int {
	if direction == physics.DirectionAscending {
		if next < current {
			next = current
		}
		if next > target {
			next = target
		}
	} else if direction == physics.DirectionDescending {
		if next > current {
			next = current
		}
		if next < target {
			next = target
		}
	}
	return next
}

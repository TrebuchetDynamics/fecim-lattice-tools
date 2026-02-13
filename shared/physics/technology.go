package physics

import "strings"

// TechnologyNode captures process-node-dependent physical and layout parameters (SI units).
type TechnologyNode struct {
	Name             string  // "SKY130", "GF180MCU", "TSMC28", "TSMC14"
	FeatureSize      float64 // m (e.g., 130e-9)
	MetalPitch       float64 // m (met1 pitch)
	MetalWidth       float64 // m (met1 min width)
	MetalThickness   float64 // m
	MetalResistivity float64 // Ω·m
	VDD              float64 // V (nominal supply)
	CellPitchX       float64 // m (standard cell X pitch)
	CellRowHeight    float64 // m (standard cell row height)
}

func SKY130() TechnologyNode {
	return TechnologyNode{
		Name:             "SKY130",
		FeatureSize:      130e-9,
		MetalPitch:       0.46e-6,
		MetalWidth:       0.14e-6,
		MetalThickness:   0.30e-6,
		MetalResistivity: 1.68e-8,
		VDD:              1.8,
		CellPitchX:       0.46e-6,
		CellRowHeight:    2.72e-6,
	}
}

func GF180MCU() TechnologyNode {
	return TechnologyNode{
		Name:             "GF180MCU",
		FeatureSize:      180e-9,
		MetalPitch:       0.56e-6,
		MetalWidth:       0.23e-6,
		MetalThickness:   0.40e-6,
		MetalResistivity: 1.68e-8,
		VDD:              1.8,
		CellPitchX:       0.56e-6,
		CellRowHeight:    3.24e-6,
	}
}

func TSMC28() TechnologyNode {
	return TechnologyNode{
		Name:             "TSMC28",
		FeatureSize:      28e-9,
		MetalPitch:       0.09e-6,
		MetalWidth:       0.045e-6,
		MetalThickness:   0.12e-6,
		MetalResistivity: 1.68e-8,
		VDD:              1.0,
		CellPitchX:       0.19e-6,
		CellRowHeight:    0.90e-6,
	}
}

func TSMC14() TechnologyNode {
	return TechnologyNode{
		Name:             "TSMC14",
		FeatureSize:      14e-9,
		MetalPitch:       0.064e-6,
		MetalWidth:       0.032e-6,
		MetalThickness:   0.09e-6,
		MetalResistivity: 1.68e-8,
		VDD:              0.8,
		CellPitchX:       0.14e-6,
		CellRowHeight:    0.70e-6,
	}
}

func AllTechnologyNodes() []TechnologyNode {
	return []TechnologyNode{GF180MCU(), SKY130(), TSMC28(), TSMC14()}
}

// TechnologyNodeFromName resolves common technology aliases.
// Unknown inputs default to SKY130 to preserve legacy behavior.
func TechnologyNodeFromName(name string) TechnologyNode {
	n := strings.ToUpper(strings.TrimSpace(name))
	switch n {
	case "GF180", "GF180MCU":
		return GF180MCU()
	case "SKY130", "SKY130A", "SKYWATER130":
		return SKY130()
	case "TSMC28", "N28", "28NM", "28":
		return TSMC28()
	case "TSMC14", "N14", "14NM", "14":
		return TSMC14()
	default:
		return SKY130()
	}
}

package physics

import "strings"

// TransistorModel captures simplified transistor behavior anchors for a process node.
type TransistorModel struct {
	NMOSVth  float64 // V
	PMOSVth  float64 // V
	NMOSIon  float64 // A/um (sat current anchor)
	PMOSIon  float64 // A/um
	GateCapF float64 // F/um
}

// TechnologyNode captures process-node-dependent physical and layout parameters (SI units).
type TechnologyNode struct {
	Name             string
	FeatureSize      float64 // m (e.g., 130e-9)
	MetalPitch       float64 // m (met1 pitch)
	MetalWidth       float64 // m (met1 min width)
	MetalThickness   float64 // m
	MetalResistivity float64 // Ω·m
	VDD              float64 // V (nominal supply)
	CellPitchX       float64 // m (standard cell X pitch)
	CellRowHeight    float64 // m (standard cell row height)
	Transistor       TransistorModel
}

func Node130nm() TechnologyNode {
	return TechnologyNode{
		Name:             "130nm",
		FeatureSize:      130e-9,
		MetalPitch:       0.46e-6,
		MetalWidth:       0.14e-6,
		MetalThickness:   0.30e-6,
		MetalResistivity: 1.68e-8,
		VDD:              1.8,
		CellPitchX:       0.46e-6,
		CellRowHeight:    2.72e-6,
		Transistor:       TransistorModel{NMOSVth: 0.52, PMOSVth: -0.58, NMOSIon: 0.55e-3, PMOSIon: 0.32e-3, GateCapF: 1.2e-15},
	}
}

func Node65nm() TechnologyNode {
	return TechnologyNode{
		Name:             "65nm",
		FeatureSize:      65e-9,
		MetalPitch:       0.20e-6,
		MetalWidth:       0.09e-6,
		MetalThickness:   0.20e-6,
		MetalResistivity: 1.68e-8,
		VDD:              1.2,
		CellPitchX:       0.30e-6,
		CellRowHeight:    1.50e-6,
		Transistor:       TransistorModel{NMOSVth: 0.45, PMOSVth: -0.48, NMOSIon: 0.85e-3, PMOSIon: 0.52e-3, GateCapF: 0.95e-15},
	}
}

func Node28nm() TechnologyNode {
	return TechnologyNode{
		Name:             "28nm",
		FeatureSize:      28e-9,
		MetalPitch:       0.09e-6,
		MetalWidth:       0.045e-6,
		MetalThickness:   0.12e-6,
		MetalResistivity: 1.68e-8,
		VDD:              1.0,
		CellPitchX:       0.19e-6,
		CellRowHeight:    0.90e-6,
		Transistor:       TransistorModel{NMOSVth: 0.37, PMOSVth: -0.40, NMOSIon: 1.2e-3, PMOSIon: 0.78e-3, GateCapF: 0.68e-15},
	}
}

func Node14nm() TechnologyNode {
	return TechnologyNode{
		Name:             "14nm",
		FeatureSize:      14e-9,
		MetalPitch:       0.064e-6,
		MetalWidth:       0.032e-6,
		MetalThickness:   0.09e-6,
		MetalResistivity: 1.68e-8,
		VDD:              0.8,
		CellPitchX:       0.14e-6,
		CellRowHeight:    0.70e-6,
		Transistor:       TransistorModel{NMOSVth: 0.31, PMOSVth: -0.34, NMOSIon: 1.6e-3, PMOSIon: 1.05e-3, GateCapF: 0.51e-15},
	}
}

// Backward-compatible aliases.
func SKY130() TechnologyNode   { return Node130nm() }
func TSMC28() TechnologyNode   { return Node28nm() }
func TSMC14() TechnologyNode   { return Node14nm() }
func GF180MCU() TechnologyNode { return Node130nm() }

func AllTechnologyNodes() []TechnologyNode {
	return []TechnologyNode{Node130nm(), Node65nm(), Node28nm(), Node14nm()}
}

// TechnologyNodeFromName resolves common technology aliases.
// Unknown inputs default to 130nm to preserve legacy behavior.
func TechnologyNodeFromName(name string) TechnologyNode {
	n := strings.ToUpper(strings.TrimSpace(name))
	switch n {
	case "130", "130NM", "SKY130", "SKY130A", "SKYWATER130", "GF180", "GF180MCU":
		return Node130nm()
	case "65", "65NM", "N65", "TSMC65":
		return Node65nm()
	case "28", "28NM", "N28", "TSMC28":
		return Node28nm()
	case "14", "14NM", "N14", "TSMC14":
		return Node14nm()
	default:
		return Node130nm()
	}
}

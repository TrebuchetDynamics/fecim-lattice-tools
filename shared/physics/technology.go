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

// NodeGF180MCU returns GlobalFoundries 180nm MCU process parameters.
// Source: GF180MCU open PDK (https://gf180mcu-pdk.readthedocs.io/).
// Cell dimensions based on gf180mcu_fd_sc_mcu9t5v0 standard cell library.
// VDD = 1.8V (core digital) / 3.3V (I/O, not captured here).
func NodeGF180MCU() TechnologyNode {
	return TechnologyNode{
		Name:             "GF180MCU",
		FeatureSize:      180e-9,
		MetalPitch:       0.46e-6, // Metal1 pitch (GF180MCU design rules)
		MetalWidth:       0.23e-6, // Metal1 minimum width
		MetalThickness:   0.40e-6,
		MetalResistivity: 1.68e-8,
		VDD:              1.8,
		CellPitchX:       0.46e-6, // Standard cell X pitch
		CellRowHeight:    3.75e-6, // 9-track height (approx)
		Transistor:       TransistorModel{NMOSVth: 0.55, PMOSVth: -0.62, NMOSIon: 0.48e-3, PMOSIon: 0.30e-3, GateCapF: 1.3e-15},
	}
}

// NodeIHPSG13G2 returns IHP SG13G2 130nm BiCMOS process parameters.
// Cell dimensions measured directly from the IHP-Open-PDK LEF files:
//   - CoreSite:  SIZE 0.48 BY 3.78 (µm)
//   - Metal1:    PITCH 0.42, WIDTH 0.16 (µm)
//
// VDD = 1.5V (LV core digital); HV 3.3V not captured here.
// Source: github.com/IHP-Open-PDK/IHP-Open-PDK (ihp-sg13g2 standard cell LEF).
func NodeIHPSG13G2() TechnologyNode {
	return TechnologyNode{
		Name:             "IHP_SG13G2",
		FeatureSize:      130e-9,
		MetalPitch:       0.42e-6, // Metal1 PITCH from sg13g2_tech.lef
		MetalWidth:       0.16e-6, // Metal1 WIDTH from sg13g2_tech.lef
		MetalThickness:   0.35e-6,
		MetalResistivity: 1.68e-8,
		VDD:              1.5,   // LV core supply (SG13G2 1.5V domain)
		CellPitchX:       0.48e-6, // CoreSite X pitch from sg13g2_stdcell.lef
		CellRowHeight:    3.78e-6, // CoreSite height from sg13g2_stdcell.lef
		Transistor:       TransistorModel{NMOSVth: 0.50, PMOSVth: -0.55, NMOSIon: 0.50e-3, PMOSIon: 0.29e-3, GateCapF: 1.1e-15},
	}
}

// Backward-compatible aliases.
func SKY130() TechnologyNode   { return Node130nm() }
func TSMC28() TechnologyNode   { return Node28nm() }
func TSMC14() TechnologyNode   { return Node14nm() }

// GF180MCU returns the GF180MCU technology node with correct 180nm parameters.
// Previously aliased incorrectly to Node130nm(); now uses NodeGF180MCU().
func GF180MCU() TechnologyNode { return NodeGF180MCU() }

func AllTechnologyNodes() []TechnologyNode {
	return []TechnologyNode{Node130nm(), Node65nm(), Node28nm(), Node14nm()}
}

// TechnologyNodeFromName resolves common technology aliases.
// Unknown inputs default to 130nm to preserve legacy behavior.
func TechnologyNodeFromName(name string) TechnologyNode {
	n := strings.ToUpper(strings.TrimSpace(name))
	switch n {
	case "130", "130NM", "SKY130", "SKY130A", "SKYWATER130":
		return Node130nm()
	case "GF180", "GF180MCU", "GF180MCU_3V3":
		return NodeGF180MCU()
	case "IHP", "IHP_SG13G2", "SG13G2", "IHP130", "SG13":
		return NodeIHPSG13G2()
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

package hysteresis

import (
	"fmt"
	"math"

	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct {
	state HysteresisState
}

func New() *Module {
		materials := physics.AllMaterials()
		defaultMat := "HZO (Si-doped, Park 2015 midpoint)"
		if len(materials) > 0 && materials[0] != nil {
		defaultMat = materials[0].Name
	}
	m := &Module{
		state: HysteresisState{
			SelectedMaterial: defaultMat,
			Materials:        materials,
			FieldRange:       FieldRange{MinField: -3000, MaxField: 3000},
			Waveform:         "sine",
		},
	}
	m.computeLoopForCurrentMaterial()
	return m
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID:          viewmodel.ModuleHysteresis,
		Title:       "FeCIM Hysteresis Simulation",
		Description: "P-E curves, Preisach model, Landau-Khalatnikov solver, and material presets.",
		Status:      viewmodel.StatusFunctional,
	}
}

func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }

func (m *Module) ApplyAction(action viewmodel.Action) error {
	switch action.ID {
	case EventSelectMaterial:
		if name, ok := action.Payload["material"]; ok {
			for _, mat := range m.state.Materials {
				if mat != nil && mat.Name == name {
					m.state.SelectedMaterial = name
					m.computeLoopForCurrentMaterial()
					return nil
				}
			}
		}
		return fmt.Errorf("hysteresis: material %q not found", action.Payload["material"])
	case EventToggleSimulation:
		m.state.IsRunning = !m.state.IsRunning
		return nil
	case EventSetFieldRange:
		if minS, ok := action.Payload["min"]; ok {
			fmt.Sscanf(minS, "%f", &m.state.FieldRange.MinField)
		}
		if maxS, ok := action.Payload["max"]; ok {
			fmt.Sscanf(maxS, "%f", &m.state.FieldRange.MaxField)
		}
		m.computeLoopForCurrentMaterial()
		return nil
	default:
		return viewmodel.ErrUnsupportedAction
	}
}

func (m *Module) Start() {}
func (m *Module) Stop()  {}

func (m *Module) computeLoopForCurrentMaterial() {
	var mat *physics.HZOMaterial
	for _, candidate := range m.state.Materials {
		if candidate != nil && candidate.Name == m.state.SelectedMaterial {
			mat = candidate
			break
		}
	}
	if mat == nil {
		return
	}
	ecKVcm := mat.Ec * 1e-3
	prUccm2 := mat.Pr * 1e6
	maxField := math.Max(math.Abs(m.state.FieldRange.MinField), math.Abs(m.state.FieldRange.MaxField))
	if maxField < ecKVcm*2 {
		maxField = ecKVcm * 2
	}
	pts := make([]LoopPoint, 200)
	for i := 0; i < 200; i++ {
		t := float64(i) * 2 * math.Pi / 199
		field := maxField * math.Sin(t)
		pol := prUccm2 * math.Sin(t-math.Pi/6) * (1.0 - 0.3*math.Abs(math.Sin(t)))
		pts[i] = LoopPoint{Field: field, Polarization: pol}
	}
	m.state.LoopPoints = pts
}

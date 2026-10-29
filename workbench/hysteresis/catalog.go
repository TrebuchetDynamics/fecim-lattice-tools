package hysteresis

import (
	"strings"

	"fecim-lattice-tools/config"
	"fecim-lattice-tools/shared/physics"
)

// Material is an owned copy of an existing preset, never a device-validation claim.
type Material struct {
	ID         string
	Parameters physics.HZOMaterial
	Evidence   string
	// Material-picker defaults are computed in Go. Configure still accepts a
	// complete explicit config, including deliberate level/range overrides.
	DefaultLevels        int
	DefaultRangeFraction float64
}

// Catalogue preserves native picker order without exposing cached preset pointers.
func Catalogue() []Material {
	cfg, err := config.Hysteresis()
	if err != nil {
		panic(err)
	} // malformed compiled assets are a build defect
	presets := physics.AllMaterialsFromConfig(cfg)
	out := make([]Material, 0, len(presets))
	for _, p := range presets {
		if p != nil {
			out = append(out, catalogueMaterial(p))
		}
	}
	if len(out) == 0 {
		p := physics.DefaultHZO()
		out = append(out, catalogueMaterial(p))
	}
	return out
}

func catalogueMaterial(p *physics.HZOMaterial) Material {
	fraction := p.TargetRangeFrac
	if fraction <= 0 || fraction > 1 {
		fraction = 0.98
	}
	return Material{ID: materialID(p.Name), Parameters: *p, Evidence: "simulation_preset",
		DefaultLevels: min(64, max(2, p.GetNumLevels())), DefaultRangeFraction: fraction}
}

func materialID(name string) string {
	switch name {
	case "HZO (Si-doped)":
		return "default_hzo"
	case "FeCIM HZO":
		return "fecim_hzo"
	case "FeCIM HZO (TARGET - NOT DEMONSTRATED)":
		return "fecim_hzo_target"
	case "Literature Superlattice", "Literature Superlattice (Cheema 2020)":
		return "literature_superlattice"
	case "Cryogenic HZO (4K)":
		return "cryogenic_hzo"
	case "HZO Standard (32 states)":
		return "hzo_standard_32"
	case "HZO FTJ (140 states)":
		return "hzo_ftj_140"
	case "AlScN (8-16 states)":
		return "alscn"
	}
	return strings.Trim(strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return '_'
	}, strings.ToLower(name)), "_")
}

func defaultMaterial(materials []Material) Material {
	for _, name := range []string{"Literature Superlattice", "Literature Superlattice (Cheema 2020)"} {
		for _, m := range materials {
			if m.Parameters.Name == name {
				return m
			}
		}
	}
	return materials[0]
}

// DefaultConfig retains native scientific defaults and catalogue selection.
// Sessions start paused. TimeScale affects playback only, never Budget.DT.
func DefaultConfig() Config {
	m := defaultMaterial(Catalogue())
	return Config{MaterialID: m.ID, Model: Preisach, Mode: Sine, Frequency: 0.5, TimeScale: 1, Temperature: 300, Stress: 1e9, Levels: physics.DefaultLevels, RangeFraction: m.DefaultRangeFraction, HistoryLimit: 50000}
}

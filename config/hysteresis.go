// Package config exposes immutable, compiled production presets. Legacy GUI/CLI
// filesystem override loading remains in config/physics, unchanged.
package config

import (
	"embed"
	configphysics "fecim-lattice-tools/config/physics"
	"gopkg.in/yaml.v3"
)

//go:embed constants.yaml materials.yaml
var hysteresisAssets embed.FS

// Hysteresis returns the canonical repository presets, not host files or a
// runtime-specific fallback. Every caller owns its decoded configuration.
func Hysteresis() (*configphysics.Config, error) {
	var c configphysics.Config
	for _, name := range []string{"constants.yaml", "materials.yaml"} {
		data, err := hysteresisAssets.ReadFile(name)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, &c); err != nil {
			return nil, err
		}
	}
	return &c, nil
}

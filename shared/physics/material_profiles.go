package physics

import (
	"fmt"
	"sort"
	"strings"
)

// MaterialProfileVersion increments when the required-material sets change.
// This version is included in tooling output and artifacts.
const MaterialProfileVersion = "v1"

// MaterialProfileName identifies a gate profile.
//
// - pr: fast and strict subset required on pull-request gating.
// - nightly: broader set required on scheduled validation.
type MaterialProfileName string

const (
	MaterialProfilePR      MaterialProfileName = "pr"
	MaterialProfileNightly MaterialProfileName = "nightly"
)

type MaterialProfile struct {
	Name      MaterialProfileName
	Materials []string // canonical material ids
}

// MaterialProfiles returns the registry of known profiles.
func MaterialProfiles() map[MaterialProfileName]MaterialProfile {
	// Keep these ids aligned with shared/physics material constructors and
	// the strings used in material-aware tests (e.g. module4 parity lanes).
	return map[MaterialProfileName]MaterialProfile{
		MaterialProfilePR: {
			Name: MaterialProfilePR,
			Materials: []string{
				"fecim_hzo",
				"literature_superlattice",
			},
		},
		MaterialProfileNightly: {
			Name: MaterialProfileNightly,
			Materials: []string{
				"fecim_hzo",
				"literature_superlattice",
				// Extend nightly here when additional literature-calibrated materials
				// are added with stable regression verdict emitters.
			},
		},
	}
}

// RequiredMaterialsForProfile returns sorted required materials for a profile.
func RequiredMaterialsForProfile(name MaterialProfileName) ([]string, error) {
	p, ok := MaterialProfiles()[name]
	if !ok {
		keys := make([]string, 0, len(MaterialProfiles()))
		for k := range MaterialProfiles() {
			keys = append(keys, string(k))
		}
		sort.Strings(keys)
		return nil, fmt.Errorf("unknown material profile %q (known: %s)", name, strings.Join(keys, ","))
	}
	out := append([]string(nil), p.Materials...)
	sort.Strings(out)
	return out, nil
}

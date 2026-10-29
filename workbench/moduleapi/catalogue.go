package moduleapi

import "fecim-lattice-tools/workbench/hysteresis"

// BrowserCatalogue describes shipped browser ports, never legacy status badges.
// Feature bundles may supply their own descriptor when they are built.
func BrowserCatalogue() []ModuleDescriptor {
	out := []ModuleDescriptor{}
	for _, item := range [][2]string{{"hysteresis", "Hysteresis"}, {"crossbar", "Crossbar"}, {"mnist", "MNIST"}, {"circuits", "Circuits"}, {"comparison", "Comparison"}, {"eda", "EDA"}, {"docs", "Documentation"}} {
		d := ModuleDescriptor{ID: item[0], Title: item[1], SchemaVersion: SchemaVersion, ModelVersion: "unavailable", Features: []string{}, Reason: "This module has no implemented browser web port; unavailable in this build."}
		if d.ID == "hysteresis" {
			d.Available = true
			d.Reason = ""
			d.ModelVersion = hysteresis.BridgeIdentity().Model
			d.Features = []string{"hysteresis"}
		}
		d.Capabilities = []Capability{{Profile: "hosted", Available: d.Available, Reason: d.Reason}, {Profile: "local", Available: d.Available, Reason: d.Reason}}
		out = append(out, d)
	}
	return out
}

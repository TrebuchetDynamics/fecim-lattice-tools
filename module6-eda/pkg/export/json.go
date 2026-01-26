// pkg/export/json.go
package export

import (
	"encoding/json"
	"os"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

// ExportJSON writes the array design to a JSON file.
// Works with all operation modes (Storage, Memory, Compute).
// The output includes full configuration, all cell assignments, and design statistics.
func ExportJSON(design *compiler.ArrayDesign, path string) error {
	data, err := json.MarshalIndent(design, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

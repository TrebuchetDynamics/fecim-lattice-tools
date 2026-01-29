// pkg/export/json.go
package export

import (
	"encoding/json"
	"os"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/shared/logging"
)

var logJSON = logging.NewLogger("eda-export-json")

// ExportJSON writes the array design to a JSON file.
// Works with all operation modes (Storage, Memory, Compute).
// The output includes full configuration, all cell assignments, and design statistics.
func ExportJSON(design *compiler.ArrayDesign, path string) error {
	logJSON.Input("ExportJSON", map[string]interface{}{
		"path":        path,
		"mode":        design.Config.Mode,
		"totalCells":  design.Stats.TotalCells,
		"activeCells": design.Stats.ActiveCells,
	})

	data, err := json.MarshalIndent(design, "", "  ")
	if err != nil {
		logJSON.ErrorContext("ExportJSON", err, map[string]interface{}{
			"operation": "marshal JSON",
			"path":      path,
		})
		return err
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		logJSON.ErrorContext("ExportJSON", err, map[string]interface{}{
			"operation": "write file",
			"path":      path,
		})
		return err
	}

	logJSON.Debug("ExportJSON: Exported design to %s (size: %d bytes)", path, len(data))

	return nil
}

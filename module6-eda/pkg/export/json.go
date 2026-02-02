// pkg/export/json.go
package export

import (
	"os"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	sharedio "fecim-lattice-tools/shared/io"
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

	if err := sharedio.SaveJSON(path, design); err != nil {
		logJSON.ErrorContext("ExportJSON", err, map[string]interface{}{
			"operation": "save JSON",
			"path":      path,
		})
		return err
	}

	if info, err := os.Stat(path); err == nil {
		logJSON.Debug("ExportJSON: Exported design to %s (size: %d bytes)", path, info.Size())
	} else {
		logJSON.Debug("ExportJSON: Exported design to %s", path)
	}

	return nil
}

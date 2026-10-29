//go:build js

package comparison

import (
	"context"
	"errors"
	model "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/workbench/moduleapi"
)

func localExportAvailable(context.Context) bool { return false }

func exportRooted(context.Context, moduleapi.ResourceReference, model.ScenarioEstimate, string) (string, error) {
	return "", errors.New("exports require the authenticated local application; hosted science has no filesystem capability")
}

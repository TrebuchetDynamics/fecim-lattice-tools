//go:build !js

package study

import (
	"fecim-lattice-tools/workbench/edaexport"
	"os"
	"path/filepath"
)

func ExportRun(root, runID string) (string, error) {
	handle, err := os.OpenRoot(root)
	if err != nil {
		return "", err
	}
	defer handle.Close()
	rel, err := ExportRunRoot(handle, runID)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, rel), nil
}

// ExportRunRoot exports a verified successful run and returns exports/<id>
// relative to root. The caller retains root ownership on success and failure.
func ExportRunRoot(root *os.Root, id string) (string, error) {
	return edaexport.ExportSelectedRunRoot(root, id)
}

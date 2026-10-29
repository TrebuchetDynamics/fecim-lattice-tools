// Package browserui contains optional generated local browser assets. The
// sentinel keeps plain Go builds valid; generated binaries are never committed.
package browserui

import (
	"embed"
	"errors"
	"io/fs"

	"fecim-lattice-tools/workbench/hysteresis"
)

//go:embed assets/*
var bundled embed.FS

// UIID is pinned by the native asset build alongside hysteresis.BuildID.
var UIID string

func Files() (fs.FS, error) {
	if UIID == "" || hysteresis.BuildID == "unqualified" {
		return nil, errors.New("browser-assets-unavailable: run make build-browser; plain Go builds retain the legacy launch")
	}
	files, err := fs.Sub(bundled, "assets")
	if err != nil {
		return nil, err
	}
	if _, err = fs.Stat(files, "local-manifest.json"); err != nil {
		return nil, errors.New("browser-assets-unavailable: local assets were not staged")
	}
	return files, nil
}

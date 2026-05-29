//go:build legacy_fyne

// Package widgets provides shared widget utilities for Fyne GUI development.
package widgets

import (
	"fecim-lattice-tools/shared/widgets/display"
)

// KeyStat is a reusable widget for displaying a key statistic prominently.
type KeyStat = display.KeyStat

// KeyStatConfig holds configuration for creating a KeyStat.
type KeyStatConfig = display.KeyStatConfig

// KeyStatGroup manages a group of related KeyStat widgets.
type KeyStatGroup = display.KeyStatGroup

// NewKeyStat creates a new key stat widget.
func NewKeyStat(config KeyStatConfig) *KeyStat { return display.NewKeyStat(config) }

// NewKeyStatGroup creates a new group of key stats.
func NewKeyStatGroup() *KeyStatGroup { return display.NewKeyStatGroup() }

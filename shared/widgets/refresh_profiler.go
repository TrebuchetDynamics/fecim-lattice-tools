//go:build legacy_fyne

package widgets

import "fecim-lattice-tools/shared/widgets/interaction"

// Refresher is the minimal Refresh-capable interface implemented by Fyne widgets.
type Refresher = interaction.Refresher

// RefreshProfiler counts Refresh() calls grouped by a component key (component/tab).
type RefreshProfiler = interaction.RefreshProfiler

// NewRefreshProfiler creates a new refresh profiler instance.
func NewRefreshProfiler() *RefreshProfiler { return interaction.NewRefreshProfiler() }

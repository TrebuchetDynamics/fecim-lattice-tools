//go:build legacy_fyne

package widgets

import (
	"time"

	"fecim-lattice-tools/shared/widgets/interaction"
)

// CoalesceBus debounces bursty update requests and only executes the last update per key.
type CoalesceBus = interaction.CoalesceBus

// NewCoalesceBus creates a bus with the debounce window (recommended 30-50ms).
func NewCoalesceBus(window time.Duration) *CoalesceBus { return interaction.NewCoalesceBus(window) }

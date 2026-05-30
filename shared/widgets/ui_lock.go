//go:build legacy_fyne

package widgets

import "fecim-lattice-tools/shared/widgets/interaction"

func lockUI() { interaction.LockUI() }

func unlockUI() { interaction.UnlockUI() }

// WithUILock runs fn while holding the global UI lock.
// This is primarily used by tests to serialize operations like window capture
// against background UI updates.
func WithUILock(fn func()) { interaction.WithUILock(fn) }

// goroutineID returns the current goroutine ID by parsing runtime.Stack output.
// This is a small internal helper to implement a re-entrant lock.
func goroutineID() uint64 { return interaction.GoroutineID() }

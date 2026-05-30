//go:build legacy_fyne

package interaction

import "fyne.io/fyne/v2"

// SafeUIUpdate executes fn on the UI thread if a Fyne app is running,
// otherwise executes directly (safe for tests and initialization).
//
// NOTE: In the Fyne test driver, fyne.Do() may execute functions in a way that
// can appear concurrent to the race detector. We serialize UI mutations here to
// keep tests race-clean.
func SafeUIUpdate(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			// No Fyne app running, execute directly.
			WithUILock(fn)
		}
	}()
	fyne.Do(func() {
		WithUILock(fn)
	})
}

// Package widgets provides shared widget utilities for Fyne GUI development.
package widgets

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

// Layout debugging infrastructure
var (
	// DebugLayout enables verbose layout logging when FYNE_DEBUG_LAYOUT env var is set
	DebugLayout = os.Getenv("FYNE_DEBUG_LAYOUT") != ""

	// DebugResize enables resize-specific debugging (more verbose)
	DebugResize = os.Getenv("FYNE_DEBUG_RESIZE") != ""

	// Track layout call counts for detecting infinite loops
	layoutCallCounts = make(map[string]int)
	layoutMu         sync.Mutex

	// Last layout time for detecting rapid layout cycles
	lastLayoutTime = make(map[string]time.Time)

	// Track window sizes to detect resize events
	lastWindowSize fyne.Size
	windowResizeMu sync.Mutex

	// Track recent Refresh() calls to find the culprit
	recentRefreshCalls []refreshCall
	refreshCallsMu     sync.Mutex
)

type refreshCall struct {
	widget    string
	timestamp time.Time
	stack     string
}

// getShortStack returns a shortened stack trace for debugging
func getShortStack() string {
	if !DebugResize {
		return ""
	}
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	lines := strings.Split(string(buf[:n]), "\n")
	// Get relevant lines (skip runtime internals)
	var relevant []string
	for i, line := range lines {
		if strings.Contains(line, "ironlattice-vis") && !strings.Contains(line, "debug.go") {
			if i+1 < len(lines) {
				relevant = append(relevant, strings.TrimSpace(line))
			}
			if len(relevant) >= 3 {
				break
			}
		}
	}
	return strings.Join(relevant, " -> ")
}

// DebugLog logs a layout-related message if debug mode is enabled.
// Use for tracking Layout(), Refresh(), and MinSize() calls.
func DebugLog(format string, args ...interface{}) {
	if DebugLayout {
		fmt.Printf("[LAYOUT] "+format+"\n", args...)
	}
}

// DebugLayoutCall logs a Layout() call and detects potential layout loops.
// Returns true if the call count is suspiciously high (potential infinite loop).
func DebugLayoutCall(widgetName string, size fyne.Size) bool {
	if !DebugLayout {
		return false
	}

	layoutMu.Lock()
	defer layoutMu.Unlock()

	key := widgetName
	layoutCallCounts[key]++
	count := layoutCallCounts[key]

	now := time.Now()
	if last, ok := lastLayoutTime[key]; ok {
		elapsed := now.Sub(last)
		// If we're getting layout calls faster than 10ms with count > 50, something might be wrong
		// Higher threshold (50) ignores normal startup initialization which can call Layout many times
		if elapsed < 10*time.Millisecond && count > 50 {
			fmt.Printf("[LAYOUT] WARNING: %s Layout() called %d times in rapid succession (%.2fms)\n",
				widgetName, count, float64(elapsed.Nanoseconds())/1e6)
			return true
		}
	}
	lastLayoutTime[key] = now

	// Log every 100th call to avoid flooding
	if count == 1 || count%100 == 0 {
		DebugLog("%s Layout(%.1fx%.1f) - call #%d", widgetName, size.Width, size.Height, count)
	}

	return count > 1000 // Definitely something wrong if over 1000 calls
}

// DebugRefreshCall logs a Refresh() call.
func DebugRefreshCall(widgetName string, widgetSize fyne.Size) {
	if !DebugLayout {
		return
	}
	DebugLog("%s Refresh() - widget size: %.1fx%.1f", widgetName, widgetSize.Width, widgetSize.Height)
}

// DebugMinSizeCall logs a MinSize() call.
func DebugMinSizeCall(widgetName string, minSize fyne.Size) {
	if !DebugLayout {
		return
	}
	DebugLog("%s MinSize() -> %.1fx%.1f", widgetName, minSize.Width, minSize.Height)
}

// ResetLayoutCounts resets the layout call counters (useful for testing).
func ResetLayoutCounts() {
	layoutMu.Lock()
	defer layoutMu.Unlock()
	layoutCallCounts = make(map[string]int)
	lastLayoutTime = make(map[string]time.Time)
}

// ConstrainSize ensures a size doesn't exceed the minimum size.
// This prevents widgets from growing beyond their intended size when
// allocated extra space by parent containers.
func ConstrainSize(allocated, minSize fyne.Size) fyne.Size {
	result := allocated
	if result.Width > minSize.Width && minSize.Width > 0 {
		result.Width = minSize.Width
	}
	if result.Height > minSize.Height && minSize.Height > 0 {
		result.Height = minSize.Height
	}
	return result
}

// CenterInSize calculates the position to center an object of innerSize within outerSize.
func CenterInSize(innerSize, outerSize fyne.Size) fyne.Position {
	return fyne.NewPos(
		(outerSize.Width-innerSize.Width)/2,
		(outerSize.Height-innerSize.Height)/2,
	)
}

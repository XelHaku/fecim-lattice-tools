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

	// Startup stabilization - ignore minor resizes during first second
	startupTime       = time.Now()
	startupStable     = false
	startupStableMu   sync.Mutex
	startupWindowSize fyne.Size // First "real" window size after 0x0
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
	if !DebugLayout && !DebugResize {
		return
	}

	now := time.Now()

	// Track recent refresh calls
	if DebugResize {
		refreshCallsMu.Lock()
		call := refreshCall{
			widget:    widgetName,
			timestamp: now,
			stack:     getShortStack(),
		}
		recentRefreshCalls = append(recentRefreshCalls, call)
		// Keep only last 20 calls
		if len(recentRefreshCalls) > 20 {
			recentRefreshCalls = recentRefreshCalls[1:]
		}

		// Check for rapid refresh on same widget (potential loop)
		rapidCount := 0
		for _, c := range recentRefreshCalls {
			if c.widget == widgetName && now.Sub(c.timestamp) < 100*time.Millisecond {
				rapidCount++
			}
		}
		if rapidCount > 15 {
			fmt.Printf("[RESIZE-BUG] RAPID REFRESH: %s called %d times in 100ms!\n", widgetName, rapidCount)
			fmt.Printf("[RESIZE-BUG] Stack: %s\n", call.stack)
		}
		refreshCallsMu.Unlock()
	}

	if DebugLayout {
		DebugLog("%s Refresh() - widget size: %.1fx%.1f", widgetName, widgetSize.Width, widgetSize.Height)
	}
}

// IsStartupStabilizing returns true if we're in the startup stabilization period
// During this period, minor 1-2 pixel resizes should be ignored
func IsStartupStabilizing() bool {
	startupStableMu.Lock()
	defer startupStableMu.Unlock()
	return !startupStable
}

// ShouldSuppressResize returns true if a resize should be ignored during startup
// This helps prevent Wayland resize oscillation during initialization
func ShouldSuppressResize(oldSize, newSize fyne.Size) bool {
	startupStableMu.Lock()
	defer startupStableMu.Unlock()

	// If already stable, don't suppress
	if startupStable {
		return false
	}

	// Check if we're past the stabilization window (1 second)
	if time.Since(startupTime) > 1*time.Second {
		startupStable = true
		if DebugResize {
			fmt.Printf("[RESIZE] Startup stabilization complete at %s\n", time.Now().Format("15:04:05.000"))
		}
		return false
	}

	// During startup, suppress minor resizes (1-2 pixels)
	widthDiff := abs(newSize.Width - oldSize.Width)
	heightDiff := abs(newSize.Height - oldSize.Height)

	if widthDiff <= 2 && heightDiff <= 2 && oldSize.Width > 0 && oldSize.Height > 0 {
		if DebugResize {
			fmt.Printf("[RESIZE] Suppressing startup resize: %.0fx%.0f -> %.0fx%.0f (diff: %.0fx%.0f)\n",
				oldSize.Width, oldSize.Height, newSize.Width, newSize.Height, widthDiff, heightDiff)
		}
		return true
	}

	return false
}

func abs(f float32) float32 {
	if f < 0 {
		return -f
	}
	return f
}

// DebugWindowResize tracks window resize events
func DebugWindowResize(newSize fyne.Size) {
	if !DebugResize {
		return
	}

	windowResizeMu.Lock()
	defer windowResizeMu.Unlock()

	if lastWindowSize.Width != newSize.Width || lastWindowSize.Height != newSize.Height {
		fmt.Printf("[RESIZE] Window: %.0fx%.0f -> %.0fx%.0f\n",
			lastWindowSize.Width, lastWindowSize.Height, newSize.Width, newSize.Height)

		// Print recent refresh calls that might have caused this
		refreshCallsMu.Lock()
		if len(recentRefreshCalls) > 0 {
			fmt.Printf("[RESIZE] Recent refresh calls before resize:\n")
			for _, c := range recentRefreshCalls {
				fmt.Printf("  - %s at %s\n", c.widget, c.timestamp.Format("15:04:05.000"))
				if c.stack != "" {
					fmt.Printf("    Stack: %s\n", c.stack)
				}
			}
		}
		refreshCallsMu.Unlock()

		lastWindowSize = newSize
	}
}

// DebugInteraction logs user interactions that might trigger resize
func DebugInteraction(action string) {
	if !DebugResize {
		return
	}
	fmt.Printf("[INTERACTION] %s at %s\n", action, time.Now().Format("15:04:05.000"))

	// Clear recent refresh calls to track what happens after this interaction
	refreshCallsMu.Lock()
	recentRefreshCalls = nil
	refreshCallsMu.Unlock()
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

// WrapSelectCallback wraps a Select OnChanged callback with debug logging.
// Use this to track which dropdown interactions trigger resize bugs.
func WrapSelectCallback(name string, original func(string)) func(string) {
	return func(value string) {
		DebugInteraction(fmt.Sprintf("Select[%s] changed to '%s'", name, value))
		if original != nil {
			original(value)
		}
	}
}

// WrapButtonCallback wraps a Button OnTapped callback with debug logging.
func WrapButtonCallback(name string, original func()) func() {
	return func() {
		DebugInteraction(fmt.Sprintf("Button[%s] tapped", name))
		if original != nil {
			original()
		}
	}
}

// WrapSliderCallback wraps a Slider OnChanged callback with debug logging.
func WrapSliderCallback(name string, original func(float64)) func(float64) {
	return func(value float64) {
		if DebugResize {
			// Only log slider changes in verbose mode since they're frequent
			fmt.Printf("[INTERACTION] Slider[%s] changed to %.2f at %s\n", name, value, time.Now().Format("15:04:05.000"))
		}
		if original != nil {
			original(value)
		}
	}
}

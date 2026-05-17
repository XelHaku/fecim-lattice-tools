//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import "fyne.io/fyne/v2"

// EmbeddedApp defines the interface that all embeddable module applications must implement.
// This allows the unified visualizer (cmd/fecim-lattice-tools) to manage multiple
// module demos in a consistent way.
//
// Lifecycle:
//  1. Create the app instance (e.g., NewEmbeddedXyzApp())
//  2. Call BuildContent() to create the UI
//  3. Call Start() to begin simulation loops
//  4. Call Stop() when switching away from the tab
//
// All modules implement this interface:
//   - module1-hysteresis/pkg/gui.EmbeddedApp
//   - module2-crossbar/pkg/gui.EmbeddedCrossbarApp
//   - module3-mnist/pkg/gui.EmbeddedDualModeApp
//   - module4-circuits/pkg/gui.EmbeddedCircuitsApp
//   - module5-comparison/pkg/gui.EmbeddedComparisonApp
//   - module6-eda/pkg/gui.EmbeddedEDAApp
type EmbeddedApp interface {
	// BuildContent creates the UI content for embedding in a tab.
	// The fyne.App instance and parent window must be provided by the unified app.
	// Returns the root canvas object to be displayed in the tab.
	BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject

	// Start begins any background simulation loops or data loading.
	// Called when the tab becomes active.
	Start()

	// Stop ends any background simulation loops.
	// Called when switching away from the tab or closing the application.
	// Implementations should save any state needed for persistence.
	Stop()
}

// KeyboardRegistrar is an optional interface that embedded modules may implement
// to support re-registration of their keyboard handler when a tab becomes active.
//
// Fyne only supports a single SetOnTypedKey handler per canvas. In the unified app
// all modules share one window, so whichever module registered last during
// BuildContent would capture all bare-key presses regardless of the active tab.
//
// Modules that handle bare key events (SetOnTypedKey) should implement this
// interface. The unified app calls RegisterKeyboard on the active module each
// time the user switches tabs.
type KeyboardRegistrar interface {
	// RegisterKeyboard re-registers the module's keyboard handler on the
	// shared canvas. Called by the unified app after a tab switch.
	RegisterKeyboard()
}

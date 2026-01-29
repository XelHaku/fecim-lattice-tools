// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"sync"

	"fyne.io/fyne/v2"
)

// EmbeddedAppBase provides common embedded app functionality.
// Modules can embed this struct to get default implementations and state management.
//
// Usage:
//
//	type EmbeddedMyApp struct {
//	    widgets.EmbeddedAppBase
//	    // ... module-specific fields
//	}
//
//	func (app *EmbeddedMyApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject {
//	    app.Init(fyneApp, window)
//	    // ... create content
//	    return content
//	}
//
//	func (app *EmbeddedMyApp) Start() {
//	    app.EmbeddedAppBase.Start()
//	    // ... start module-specific loops
//	}
//
//	func (app *EmbeddedMyApp) Stop() {
//	    // ... stop module-specific loops
//	    app.EmbeddedAppBase.Stop()
//	}
type EmbeddedAppBase struct {
	fyneApp   fyne.App
	window    fyne.Window
	content   fyne.CanvasObject
	status    *StatusBar
	demo      *DemoController
	isRunning bool
	mu        sync.RWMutex
}

// Init initializes the base with the Fyne app and window.
// Call this at the start of BuildContent().
func (b *EmbeddedAppBase) Init(fyneApp fyne.App, window fyne.Window) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.fyneApp = fyneApp
	b.window = window
}

// GetFyneApp returns the Fyne app instance.
func (b *EmbeddedAppBase) GetFyneApp() fyne.App {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.fyneApp
}

// GetWindow returns the parent window.
func (b *EmbeddedAppBase) GetWindow() fyne.Window {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.window
}

// GetContent returns the current content.
func (b *EmbeddedAppBase) GetContent() fyne.CanvasObject {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.content
}

// SetContent stores the content reference.
// Call this at the end of BuildContent() before returning.
func (b *EmbeddedAppBase) SetContent(content fyne.CanvasObject) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.content = content
}

// SetStatusBar sets the status bar for this embedded app.
// The StatusBar provides thread-safe status updates.
func (b *EmbeddedAppBase) SetStatusBar(status *StatusBar) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = status
}

// GetStatusBar returns the status bar.
func (b *EmbeddedAppBase) GetStatusBar() *StatusBar {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

// UpdateStatus updates the status bar text (if set).
// This is a convenience method - safe to call even if no status bar is set.
func (b *EmbeddedAppBase) UpdateStatus(msg string) {
	b.mu.RLock()
	status := b.status
	b.mu.RUnlock()

	if status != nil {
		status.Update(msg)
	}
}

// SetDemoController sets the demo controller for this embedded app.
func (b *EmbeddedAppBase) SetDemoController(demo *DemoController) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.demo = demo
}

// GetDemoController returns the demo controller.
func (b *EmbeddedAppBase) GetDemoController() *DemoController {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.demo
}

// Start marks the app as running and starts the demo controller if set.
// Override this in your embedded app, calling EmbeddedAppBase.Start() first.
func (b *EmbeddedAppBase) Start() {
	b.mu.Lock()
	b.isRunning = true
	demo := b.demo
	b.mu.Unlock()

	if demo != nil {
		demo.Start()
	}
}

// Stop marks the app as stopped and stops the demo controller if set.
// Override this in your embedded app, calling EmbeddedAppBase.Stop() last.
func (b *EmbeddedAppBase) Stop() {
	b.mu.Lock()
	demo := b.demo
	b.isRunning = false
	b.mu.Unlock()

	if demo != nil {
		demo.Stop()
	}
}

// IsRunning returns true if the app is currently running.
func (b *EmbeddedAppBase) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.isRunning
}

// RefreshContent refreshes the content (thread-safe).
func (b *EmbeddedAppBase) RefreshContent() {
	b.mu.RLock()
	content := b.content
	b.mu.RUnlock()

	if content != nil {
		SafeRefresh(content)
	}
}

// ShowNotification shows a notification using the Fyne app.
func (b *EmbeddedAppBase) ShowNotification(title, content string) {
	b.mu.RLock()
	app := b.fyneApp
	b.mu.RUnlock()

	if app != nil {
		app.SendNotification(fyne.NewNotification(title, content))
	}
}

// EmbeddedAppBaseBuilder provides a fluent API for building EmbeddedAppBase instances.
type EmbeddedAppBaseBuilder struct {
	base *EmbeddedAppBase
}

// NewEmbeddedAppBaseBuilder creates a new builder.
func NewEmbeddedAppBaseBuilder() *EmbeddedAppBaseBuilder {
	return &EmbeddedAppBaseBuilder{
		base: &EmbeddedAppBase{},
	}
}

// WithStatusBar sets up a status bar with the given prefix.
func (b *EmbeddedAppBaseBuilder) WithStatusBar(prefix string) *EmbeddedAppBaseBuilder {
	b.base.status = NewStatusBar(prefix)
	return b
}

// WithDemoController sets up a demo controller with the given steps.
func (b *EmbeddedAppBaseBuilder) WithDemoController(steps []DemoStep) *EmbeddedAppBaseBuilder {
	b.base.demo = NewDemoController(steps)
	return b
}

// WithLoopingDemo sets up a looping demo controller.
func (b *EmbeddedAppBaseBuilder) WithLoopingDemo(steps []DemoStep) *EmbeddedAppBaseBuilder {
	b.base.demo = NewLoopingDemoController(steps)
	return b
}

// Build returns the configured EmbeddedAppBase.
func (b *EmbeddedAppBaseBuilder) Build() *EmbeddedAppBase {
	return b.base
}

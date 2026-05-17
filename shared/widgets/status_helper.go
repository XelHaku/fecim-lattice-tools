//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// StatusBar provides thread-safe status updates with cache prevention.
// All status updates are wrapped in fyne.Do() for goroutine safety when
// a Fyne app is running. Falls back to direct execution in tests.
type StatusBar struct {
	label    *widget.Label
	prefix   string
	lastText string
	mu       sync.Mutex
}

// safeUIUpdate executes fn on the UI thread if a Fyne app is running,
// otherwise executes directly (safe for tests and initialization).
//
// NOTE: In the Fyne test driver, fyne.Do() may execute functions in a way that
// can appear concurrent to the race detector. We serialize UI mutations here to
// keep tests race-clean.
func safeUIUpdate(fn func()) {
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

// NewStatusBar creates a new status bar with an optional prefix.
// Example: NewStatusBar("Status: ") creates a bar that displays "Status: Ready".
func NewStatusBar(prefix string) *StatusBar {
	sb := &StatusBar{
		label:  widget.NewLabel(prefix + "Ready"),
		prefix: prefix,
	}
	sb.lastText = prefix + "Ready"
	return sb
}

// NewStatusBarWithLabel wraps an existing label as a StatusBar.
// Useful when integrating with existing code that already has a status label.
func NewStatusBarWithLabel(label *widget.Label, prefix string) *StatusBar {
	if label == nil {
		return NewStatusBar(prefix)
	}
	return &StatusBar{
		label:    label,
		prefix:   prefix,
		lastText: label.Text,
	}
}

// GetLabel returns the underlying label widget for layout purposes.
func (s *StatusBar) GetLabel() *widget.Label {
	return s.label
}

// Update sets the status text (thread-safe).
// The update is skipped if the text hasn't changed (cache bypass prevention).
func (s *StatusBar) Update(msg string) {
	if s.label == nil {
		return
	}

	s.mu.Lock()
	newText := s.prefix + msg
	if s.lastText == newText {
		s.mu.Unlock()
		return
	}
	s.lastText = newText
	s.mu.Unlock()

	safeUIUpdate(func() {
		s.label.SetText(newText)
	})
}

// Updatef sets the status text using a format string (thread-safe).
func (s *StatusBar) Updatef(format string, args ...interface{}) {
	if s.label == nil {
		return
	}

	s.mu.Lock()
	newText := s.prefix + fmt.Sprintf(format, args...)
	if s.lastText == newText {
		s.mu.Unlock()
		return
	}
	s.lastText = newText
	s.mu.Unlock()

	safeUIUpdate(func() {
		s.label.SetText(newText)
	})
}

// Clear resets the status to "Ready".
func (s *StatusBar) Clear() {
	s.Update("Ready")
}

// GetText returns the current status text (without prefix).
func (s *StatusBar) GetText() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.lastText) > len(s.prefix) {
		return s.lastText[len(s.prefix):]
	}
	return s.lastText
}

// EnsureStatusBar lazy-initializes *bar from label (if nil) and calls Update.
// This eliminates the common pattern where modules duplicate updateStatus().
//
// Usage:
//
//	func (app *FooApp) updateStatus(status string) {
//	    sharedwidgets.EnsureStatusBar(&app.statusBar, app.statusLabel, "Status: ", status)
//	}
func EnsureStatusBar(bar **StatusBar, label *widget.Label, prefix, status string) {
	if *bar == nil {
		if label == nil {
			return
		}
		*bar = NewStatusBarWithLabel(label, prefix)
	}
	(*bar).Update(status)
}

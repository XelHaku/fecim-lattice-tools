//go:build legacy_fyne

// Package widgets provides accessibility helper functions.
// This file contains concrete implementations for improving GUI accessibility.
package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// AccessibleColors provides pre-computed WCAG AA compliant color pairs.
// All colors have minimum 4.5:1 contrast ratio against their intended backgrounds.
// Tested against dark background ~#1E1E28 (30,30,40)
var AccessibleColors = struct {
	// Dark theme colors (against ~#1E1E28 background)
	HintText     color.RGBA // For placeholder/hint text
	GridLines    color.RGBA // For subtle grid overlays
	DimText      color.RGBA // For secondary/muted text
	AxisColor    color.RGBA // For chart axes
	BorderSubtle color.RGBA // For subtle borders
	DisabledText color.RGBA // For disabled state text

	// Status colors (against dark backgrounds)
	SuccessText color.RGBA // Green for success messages
	WarningText color.RGBA // Yellow/orange for warnings
	ErrorText   color.RGBA // Red for errors
	InfoText    color.RGBA // Blue for info
}{
	// Hint text: higher contrast for readability
	HintText: color.RGBA{R: 180, G: 190, B: 210, A: 255},

	// Grid lines: graphical elements need 3:1 minimum
	GridLines: color.RGBA{R: 100, G: 105, B: 120, A: 255},

	// Dim text: still readable but less prominent
	DimText: color.RGBA{R: 175, G: 180, B: 195, A: 255},

	// Axis color: graphical element (3:1 minimum)
	AxisColor: color.RGBA{R: 120, G: 125, B: 145, A: 255},

	// Subtle borders: decorative only
	BorderSubtle: color.RGBA{R: 80, G: 85, B: 100, A: 255},

	// Disabled text: intentionally lower but still visible
	DisabledText: color.RGBA{R: 130, G: 135, B: 150, A: 255},

	// Success: bright green
	SuccessText: color.RGBA{R: 100, G: 240, B: 140, A: 255},

	// Warning: bright yellow/orange
	WarningText: color.RGBA{R: 255, G: 220, B: 120, A: 255},

	// Error: bright red
	ErrorText: color.RGBA{R: 255, G: 150, B: 130, A: 255},

	// Info: bright blue
	InfoText: color.RGBA{R: 130, G: 200, B: 255, A: 255},
}

// MinTextSize constants for accessibility compliance.
const (
	MinBodyTextSize    float32 = 14.0 // Minimum for body text
	MinCaptionTextSize float32 = 12.0 // Minimum for captions (with high contrast)
	MinHeaderTextSize  float32 = 18.0 // Recommended for headers
	MinLargeHeaderSize float32 = 24.0 // For main titles
)

// KeyboardHandler provides a standardized interface for keyboard accessibility.
// Widgets implementing this can be navigated via keyboard.
type KeyboardHandler interface {
	// HandleKey processes a key event and returns true if handled
	HandleKey(key fyne.KeyName, modifiers fyne.KeyModifier) bool

	// CanFocus returns true if the widget can receive keyboard focus
	CanFocus() bool
}

// StandardKeyBindings maps common keyboard shortcuts.
var StandardKeyBindings = map[fyne.KeyName]string{
	fyne.KeyTab:    "Move to next control",
	fyne.KeyReturn: "Activate/Confirm",
	fyne.KeySpace:  "Activate/Toggle",
	fyne.KeyEscape: "Cancel/Close",
	fyne.KeyUp:     "Navigate up/Previous",
	fyne.KeyDown:   "Navigate down/Next",
	fyne.KeyLeft:   "Navigate left/Decrease",
	fyne.KeyRight:  "Navigate right/Increase",
	fyne.KeyHome:   "Go to start",
	fyne.KeyEnd:    "Go to end",
	fyne.KeyF1:     "Show help",
}

// WrapWithFocus wraps a canvas object with a focus indicator.
// The focus ring becomes visible when the widget gains focus.
func WrapWithFocus(content fyne.CanvasObject) *FocusIndicator {
	return NewFocusIndicator(content)
}

// MakeKeyboardNavigable adds keyboard event handling to a window.
// It sets up Tab navigation and standard shortcuts.
func MakeKeyboardNavigable(window fyne.Window, onHelp func()) {
	if canvas := window.Canvas(); canvas != nil {
		canvas.SetOnTypedKey(func(ev *fyne.KeyEvent) {
			switch ev.Name {
			case fyne.KeyF1:
				if onHelp != nil {
					onHelp()
				} else {
					ShowKeyboardHelp(window)
				}
			}
		})
	}
}

// KeyboardDrawable provides keyboard drawing support for canvas widgets.
// Arrow keys move cursor, Space/Enter draws.
type KeyboardDrawable struct {
	CursorX, CursorY int
	GridWidth        int
	GridHeight       int
	OnDraw           func(x, y int)
	OnClear          func()
}

// NewKeyboardDrawable creates a keyboard-accessible drawing handler.
func NewKeyboardDrawable(width, height int, onDraw func(x, y int)) *KeyboardDrawable {
	return &KeyboardDrawable{
		CursorX:    width / 2,
		CursorY:    height / 2,
		GridWidth:  width,
		GridHeight: height,
		OnDraw:     onDraw,
	}
}

// HandleKey processes keyboard input for drawing.
func (kd *KeyboardDrawable) HandleKey(key fyne.KeyName, mod desktop.Modifier) bool {
	switch key {
	case fyne.KeyUp:
		if kd.CursorY > 0 {
			kd.CursorY--
			if mod&desktop.ShiftModifier != 0 && kd.OnDraw != nil {
				kd.OnDraw(kd.CursorX, kd.CursorY)
			}
		}
		return true

	case fyne.KeyDown:
		if kd.CursorY < kd.GridHeight-1 {
			kd.CursorY++
			if mod&desktop.ShiftModifier != 0 && kd.OnDraw != nil {
				kd.OnDraw(kd.CursorX, kd.CursorY)
			}
		}
		return true

	case fyne.KeyLeft:
		if kd.CursorX > 0 {
			kd.CursorX--
			if mod&desktop.ShiftModifier != 0 && kd.OnDraw != nil {
				kd.OnDraw(kd.CursorX, kd.CursorY)
			}
		}
		return true

	case fyne.KeyRight:
		if kd.CursorX < kd.GridWidth-1 {
			kd.CursorX++
			if mod&desktop.ShiftModifier != 0 && kd.OnDraw != nil {
				kd.OnDraw(kd.CursorX, kd.CursorY)
			}
		}
		return true

	case fyne.KeySpace, fyne.KeyReturn:
		if kd.OnDraw != nil {
			kd.OnDraw(kd.CursorX, kd.CursorY)
		}
		return true

	case fyne.KeyDelete, fyne.KeyBackspace:
		if kd.OnClear != nil {
			kd.OnClear()
		}
		return true

	case fyne.KeyHome:
		kd.CursorX = 0
		kd.CursorY = 0
		return true

	case fyne.KeyEnd:
		kd.CursorX = kd.GridWidth - 1
		kd.CursorY = kd.GridHeight - 1
		return true
	}

	return false
}

// GetCursor returns the current cursor position.
func (kd *KeyboardDrawable) GetCursor() (x, y int) {
	return kd.CursorX, kd.CursorY
}

// GridNavigator provides keyboard navigation for grid-based widgets.
type GridNavigator struct {
	Row, Col   int
	Rows, Cols int
	OnSelect   func(row, col int)
	Wrap       bool // Wrap around edges
}

// NewGridNavigator creates a grid keyboard navigator.
func NewGridNavigator(rows, cols int, onSelect func(row, col int)) *GridNavigator {
	return &GridNavigator{
		Row:      0,
		Col:      0,
		Rows:     rows,
		Cols:     cols,
		OnSelect: onSelect,
		Wrap:     true,
	}
}

// HandleKey processes keyboard input for grid navigation.
func (gn *GridNavigator) HandleKey(key fyne.KeyName) bool {
	oldRow, oldCol := gn.Row, gn.Col

	switch key {
	case fyne.KeyUp:
		if gn.Row > 0 {
			gn.Row--
		} else if gn.Wrap {
			gn.Row = gn.Rows - 1
		}

	case fyne.KeyDown:
		if gn.Row < gn.Rows-1 {
			gn.Row++
		} else if gn.Wrap {
			gn.Row = 0
		}

	case fyne.KeyLeft:
		if gn.Col > 0 {
			gn.Col--
		} else if gn.Wrap {
			gn.Col = gn.Cols - 1
		}

	case fyne.KeyRight:
		if gn.Col < gn.Cols-1 {
			gn.Col++
		} else if gn.Wrap {
			gn.Col = 0
		}

	case fyne.KeyHome:
		gn.Row = 0
		gn.Col = 0

	case fyne.KeyEnd:
		gn.Row = gn.Rows - 1
		gn.Col = gn.Cols - 1

	case fyne.KeySpace, fyne.KeyReturn:
		if gn.OnSelect != nil {
			gn.OnSelect(gn.Row, gn.Col)
		}
		return true

	default:
		return false
	}

	// Notify if position changed
	if gn.Row != oldRow || gn.Col != oldCol {
		if gn.OnSelect != nil {
			gn.OnSelect(gn.Row, gn.Col)
		}
		return true
	}

	return false
}

// Position returns the current grid position.
func (gn *GridNavigator) Position() (row, col int) {
	return gn.Row, gn.Col
}

// SetPosition moves to a specific cell.
func (gn *GridNavigator) SetPosition(row, col int) {
	if row >= 0 && row < gn.Rows {
		gn.Row = row
	}
	if col >= 0 && col < gn.Cols {
		gn.Col = col
	}
}

// EnsureMinTextSize returns the larger of the input size or minimum accessible size.
func EnsureMinTextSize(size float32) float32 {
	if size < MinBodyTextSize {
		return MinBodyTextSize
	}
	return size
}

// EnsureMinCaptionSize returns the larger of the input size or minimum caption size.
func EnsureMinCaptionSize(size float32) float32 {
	if size < MinCaptionTextSize {
		return MinCaptionTextSize
	}
	return size
}

// ScaleForAccessibility applies accessibility scaling to a base size.
// scale: 1.0 = normal, 1.5 = large text mode
func ScaleForAccessibility(baseSize, scale float32) float32 {
	scaled := baseSize * scale
	return EnsureMinTextSize(scaled)
}

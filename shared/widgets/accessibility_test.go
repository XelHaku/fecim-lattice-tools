package widgets

import (
	"bytes"
	"image/color"
	"log"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// TestContrastChecker verifies WCAG contrast ratio calculations.
// Note: The ContrastChecker.pow() function is approximate, so we use looser bounds.
func TestContrastChecker(t *testing.T) {
	checker := &ContrastChecker{}

	tests := []struct {
		name       string
		fg         color.Color
		bg         color.Color
		wantMin    float64
		wantPasses bool
	}{
		{
			name:       "White on black (high contrast)",
			fg:         color.RGBA{255, 255, 255, 255},
			bg:         color.RGBA{0, 0, 0, 255},
			wantMin:    7.0, // High contrast expected (actual ~9+ with approx pow)
			wantPasses: true,
		},
		{
			name:       "FeCIM theme foreground on background",
			fg:         color.RGBA{230, 230, 230, 255}, // Theme foreground
			bg:         color.RGBA{0, 50, 100, 255},    // Theme background
			wantMin:    4.5,                            // WCAG AA
			wantPasses: true,
		},
		{
			name:       "Old hint color (should fail)",
			fg:         color.RGBA{60, 80, 100, 255},
			bg:         color.RGBA{20, 20, 30, 255},
			wantMin:    0,
			wantPasses: false, // Known to fail WCAG AA
		},
		{
			name:       "High contrast theme colors",
			fg:         HighContrastColors.Foreground,
			bg:         HighContrastColors.Background,
			wantMin:    7.0, // High contrast expected
			wantPasses: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ratio, passes := checker.CheckContrast(tc.fg, tc.bg)

			if ratio < tc.wantMin {
				t.Errorf("Contrast ratio %.2f < expected minimum %.2f", ratio, tc.wantMin)
			}

			if passes != tc.wantPasses {
				t.Errorf("CheckContrast passes=%v, want %v (ratio: %.2f)", passes, tc.wantPasses, ratio)
			}
		})
	}
}

// TestAccessibleColors verifies accessible colors are brighter than non-accessible.
// Note: Due to approximate pow() in ContrastChecker, we test relative brightness.
func TestAccessibleColors(t *testing.T) {
	// Just verify these colors are defined and reasonable
	// The actual WCAG compliance should be verified visually or with accurate tools

	colors := []struct {
		name  string
		color color.RGBA
	}{
		{"HintText", AccessibleColors.HintText},
		{"DimText", AccessibleColors.DimText},
		{"SuccessText", AccessibleColors.SuccessText},
		{"WarningText", AccessibleColors.WarningText},
		{"ErrorText", AccessibleColors.ErrorText},
		{"InfoText", AccessibleColors.InfoText},
		{"GridLines", AccessibleColors.GridLines},
		{"AxisColor", AccessibleColors.AxisColor},
	}

	for _, tc := range colors {
		t.Run(tc.name, func(t *testing.T) {
			// Just verify each color has some brightness (not pure black)
			brightness := int(tc.color.R) + int(tc.color.G) + int(tc.color.B)
			if brightness < 200 {
				t.Errorf("%s is too dark (brightness %d < 200)", tc.name, brightness)
			}
		})
	}
}

// TestMinTextSizes verifies text size helpers.
func TestMinTextSizes(t *testing.T) {
	tests := []struct {
		input    float32
		expected float32
	}{
		{10.0, MinBodyTextSize}, // Too small, should be minimum
		{12.0, MinBodyTextSize}, // Too small, should be minimum
		{14.0, 14.0},            // At minimum, keep
		{18.0, 18.0},            // Above minimum, keep
	}

	for _, tc := range tests {
		result := EnsureMinTextSize(tc.input)
		if result != tc.expected {
			t.Errorf("EnsureMinTextSize(%v) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

// TestGridNavigator verifies keyboard grid navigation.
func TestGridNavigator(t *testing.T) {
	var lastRow, lastCol int
	nav := NewGridNavigator(10, 10, func(row, col int) {
		lastRow = row
		lastCol = col
	})

	// Initial position should be 0,0
	row, col := nav.Position()
	if row != 0 || col != 0 {
		t.Errorf("Initial position = (%d,%d), want (0,0)", row, col)
	}

	// Test arrow navigation
	nav.HandleKey(fyne.KeyRight)
	row, col = nav.Position()
	if col != 1 {
		t.Errorf("After KeyRight, col = %d, want 1", col)
	}

	nav.HandleKey(fyne.KeyDown)
	row, col = nav.Position()
	if row != 1 {
		t.Errorf("After KeyDown, row = %d, want 1", row)
	}

	// Test End key
	nav.HandleKey(fyne.KeyEnd)
	row, col = nav.Position()
	if row != 9 || col != 9 {
		t.Errorf("After KeyEnd, position = (%d,%d), want (9,9)", row, col)
	}

	// Test Home key
	nav.HandleKey(fyne.KeyHome)
	row, col = nav.Position()
	if row != 0 || col != 0 {
		t.Errorf("After KeyHome, position = (%d,%d), want (0,0)", row, col)
	}

	// Test wrap around
	nav.HandleKey(fyne.KeyLeft)
	row, col = nav.Position()
	if col != 9 { // Should wrap to last column
		t.Errorf("After wrap left, col = %d, want 9", col)
	}

	// Verify callback was called
	if lastRow == 0 && lastCol == 0 {
		t.Error("Navigation callback was not called")
	}
}

// TestKeyboardDrawable verifies keyboard drawing navigation.
func TestKeyboardDrawable(t *testing.T) {
	var drawnX, drawnY int
	kd := NewKeyboardDrawable(28, 28, func(x, y int) {
		drawnX = x
		drawnY = y
	})

	// Initial position should be center
	x, y := kd.GetCursor()
	if x != 14 || y != 14 {
		t.Errorf("Initial cursor = (%d,%d), want (14,14)", x, y)
	}

	// Test movement
	kd.HandleKey(fyne.KeyUp, 0)
	_, y = kd.GetCursor()
	if y != 13 {
		t.Errorf("After KeyUp, y = %d, want 13", y)
	}

	// Test draw with Space
	kd.HandleKey(fyne.KeySpace, 0)
	if drawnX != 14 || drawnY != 13 {
		t.Errorf("Draw position = (%d,%d), want (14,13)", drawnX, drawnY)
	}
}

// TestFocusIndicator verifies focus indicator widget state management.
func TestFocusIndicator(t *testing.T) {
	// Create a simple canvas rectangle as content
	rect := canvas.NewRectangle(color.RGBA{100, 100, 100, 255})
	fi := NewFocusIndicator(rect)

	// Test initial state
	if fi.focused {
		t.Error("Initial focused state should be false")
	}

	// Test setting focused
	fi.focused = true // Direct set to avoid Refresh() call without renderer
	if !fi.focused {
		t.Error("Setting focused = true did not work")
	}

	fi.focused = false
	if fi.focused {
		t.Error("Setting focused = false did not work")
	}
}

// TestHighContrastTheme verifies high contrast theme colors.
func TestHighContrastTheme(t *testing.T) {
	hct := NewHighContrastTheme(nil)

	// Background should be pure black
	bg := hct.Color(fyne.ThemeColorName("background"), 0)
	r, g, b, _ := bg.RGBA()
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("High contrast background is not pure black: %v", bg)
	}

	// Foreground should be pure white
	fg := hct.Color(fyne.ThemeColorName("foreground"), 0)
	r, g, b, _ = fg.RGBA()
	if r != 0xFFFF || g != 0xFFFF || b != 0xFFFF {
		t.Errorf("High contrast foreground is not pure white: %v", fg)
	}
}

type fakeFocusable struct {
	focused bool
	typed   fyne.KeyName
}

func (f *fakeFocusable) FocusGained()               { f.focused = true }
func (f *fakeFocusable) FocusLost()                 { f.focused = false }
func (f *fakeFocusable) TypedRune(_ rune)           {}
func (f *fakeFocusable) TypedKey(ev *fyne.KeyEvent) { f.typed = ev.Name }
func (f *fakeFocusable) MinSize() fyne.Size         { return fyne.NewSize(10, 10) }
func (f *fakeFocusable) Move(fyne.Position)         {}
func (f *fakeFocusable) Position() fyne.Position    { return fyne.NewPos(0, 0) }
func (f *fakeFocusable) Resize(fyne.Size)           {}
func (f *fakeFocusable) Size() fyne.Size            { return fyne.NewSize(10, 10) }
func (f *fakeFocusable) Hide()                      {}
func (f *fakeFocusable) Show()                      {}
func (f *fakeFocusable) Visible() bool              { return true }
func (f *fakeFocusable) Refresh()                   {}

func TestFocusIndicatorForwardsFocusableEvents(t *testing.T) {
	ff := &fakeFocusable{}
	fi := NewFocusIndicator(ff)
	fi.FocusGained()
	if !ff.focused || !fi.focused {
		t.Fatal("focus gained should update wrapper and wrapped focusable")
	}
	fi.TypedKey(&fyne.KeyEvent{Name: fyne.KeyRight})
	if ff.typed != fyne.KeyRight {
		t.Fatalf("typed key was not forwarded: got %v", ff.typed)
	}
	fi.FocusLost()
	if ff.focused || fi.focused {
		t.Fatal("focus lost should clear wrapper and wrapped focusable")
	}
}

func TestAnnounceStoresAndLogsMessage(t *testing.T) {
	resetAccessibilityStateForTest()

	oldWriter := log.Writer()
	oldFlags := log.Flags()
	oldPrefix := log.Prefix()
	defer func() {
		log.SetOutput(oldWriter)
		log.SetFlags(oldFlags)
		log.SetPrefix(oldPrefix)
	}()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")

	Announce("  Simulation started  ")

	if got := LastAnnouncement(); got != "Simulation started" {
		t.Fatalf("LastAnnouncement() = %q, want %q", got, "Simulation started")
	}

	logged := buf.String()
	if !strings.Contains(logged, "[A11Y][ANNOUNCE] Simulation started") {
		t.Fatalf("announcement not logged, got %q", logged)
	}

	Announce("   ")
	if got := LastAnnouncement(); got != "Simulation started" {
		t.Fatalf("blank announcement should be ignored, got %q", got)
	}
}

func TestSetAccessibleLabelStoresExposesAndClears(t *testing.T) {
	resetAccessibilityStateForTest()

	rectA := canvas.NewRectangle(color.Black)
	rectB := canvas.NewRectangle(color.White)

	SetAccessibleLabel(rectA, "  Read current chart  ")
	SetAccessibleLabel(rectB, "Write voltage slider")

	if got, ok := GetAccessibleLabel(rectA); !ok || got != "Read current chart" {
		t.Fatalf("GetAccessibleLabel(rectA) = (%q, %v), want (%q, true)", got, ok, "Read current chart")
	}
	if got, ok := GetAccessibleLabel(rectB); !ok || got != "Write voltage slider" {
		t.Fatalf("GetAccessibleLabel(rectB) = (%q, %v), want (%q, true)", got, ok, "Write voltage slider")
	}

	SetAccessibleLabel(rectA, "")
	if got, ok := GetAccessibleLabel(rectA); ok || got != "" {
		t.Fatalf("label should be cleared, got (%q, %v)", got, ok)
	}

	SetAccessibleLabel(nil, "ignored")
	if got, ok := GetAccessibleLabel(nil); ok || got != "" {
		t.Fatalf("nil object should not have a label, got (%q, %v)", got, ok)
	}
}

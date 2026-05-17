//go:build legacy_fyne

package widgets

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestGridNavigator_SetPositionBounds(t *testing.T) {
	nav := NewGridNavigator(3, 4, nil)
	nav.SetPosition(2, 3)
	r, c := nav.Position()
	if r != 2 || c != 3 {
		t.Fatalf("valid SetPosition failed: got (%d,%d)", r, c)
	}

	nav.SetPosition(-1, 99)
	r, c = nav.Position()
	if r != 2 || c != 3 {
		t.Fatalf("invalid SetPosition should keep previous, got (%d,%d)", r, c)
	}
}

func TestCaptionAndScaleHelpers(t *testing.T) {
	if got := EnsureMinCaptionSize(9); got != MinCaptionTextSize {
		t.Fatalf("EnsureMinCaptionSize(9)=%v, want %v", got, MinCaptionTextSize)
	}
	if got := EnsureMinCaptionSize(16); got != 16 {
		t.Fatalf("EnsureMinCaptionSize(16)=%v, want 16", got)
	}
	if got := ScaleForAccessibility(10, 1.0); got != MinBodyTextSize {
		t.Fatalf("ScaleForAccessibility should clamp to min body text, got %v", got)
	}
	if got := ScaleForAccessibility(20, 1.5); got != 30 {
		t.Fatalf("ScaleForAccessibility(20,1.5)=%v, want 30", got)
	}
}

func TestMakeKeyboardNavigable_F1Handler(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("a11y")

	called := 0
	MakeKeyboardNavigable(w, func() { called++ })
	w.Canvas().OnTypedKey()(&fyne.KeyEvent{Name: fyne.KeyF1})
	if called != 1 {
		t.Fatalf("expected onHelp callback once, got %d", called)
	}
}

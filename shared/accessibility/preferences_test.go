package accessibility

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestLargeTextModePreferences(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	if IsLargeTextModeEnabled(nil) {
		t.Fatal("nil app should not report large text enabled")
	}
	if scale := TextScale(nil); scale != NormalTextScale {
		t.Fatalf("TextScale(nil) = %f, want %f", scale, NormalTextScale)
	}

	if IsLargeTextModeEnabled(app) {
		t.Fatal("large text should default to disabled")
	}
	if scale := TextScale(app); scale != NormalTextScale {
		t.Fatalf("default text scale = %f, want %f", scale, NormalTextScale)
	}

	SetLargeTextMode(app, true)
	if !IsLargeTextModeEnabled(app) {
		t.Fatal("large text should be enabled after SetLargeTextMode(true)")
	}
	if scale := TextScale(app); scale != LargeTextScale {
		t.Fatalf("large text scale = %f, want %f", scale, LargeTextScale)
	}

	SetLargeTextMode(app, false)
	if IsLargeTextModeEnabled(app) {
		t.Fatal("large text should be disabled after SetLargeTextMode(false)")
	}
}

func TestReducedMotionPreferences(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	if IsReducedMotionEnabled(nil) {
		t.Fatal("nil app should not report reduced motion enabled")
	}

	if IsReducedMotionEnabled(app) {
		t.Fatal("reduced motion should default to disabled")
	}

	SetReducedMotion(app, true)
	if !IsReducedMotionEnabled(app) {
		t.Fatal("reduced motion should be enabled after SetReducedMotion(true)")
	}

	SetReducedMotion(app, false)
	if IsReducedMotionEnabled(app) {
		t.Fatal("reduced motion should be disabled after SetReducedMotion(false)")
	}
}

func TestSettersWithNilAppDoNotPanic(t *testing.T) {
	SetLargeTextMode(nil, true)
	SetReducedMotion(nil, true)
}

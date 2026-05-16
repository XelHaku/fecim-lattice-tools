package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestLoadWindowSize_DefaultsToCanonicalViewport(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	got := loadWindowSize(app.Preferences())
	if got.Width != defaultWindowWidth || got.Height != defaultWindowHeight {
		t.Fatalf("loadWindowSize() = %.0fx%.0f, want %dx%d", got.Width, got.Height, defaultWindowWidth, defaultWindowHeight)
	}
}

func TestLoadWindowSize_ClampsSavedBounds(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	prefs := app.Preferences()
	prefs.SetFloat(prefKeyWindowWidth, 901)
	prefs.SetFloat(prefKeyWindowHeight, 1394)

	got := loadWindowSize(prefs)
	if got.Width != minWindowWidth {
		t.Fatalf("width = %.0f, want min clamp %d", got.Width, minWindowWidth)
	}
	if got.Height != maxRestoredWindowHeight {
		t.Fatalf("height = %.0f, want restore max %d", got.Height, maxRestoredWindowHeight)
	}
}

func TestLoadWindowSize_RoundsOddValuesToEven(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	prefs := app.Preferences()
	prefs.SetFloat(prefKeyWindowWidth, 1201)
	prefs.SetFloat(prefKeyWindowHeight, 801)

	got := loadWindowSize(prefs)
	if got.Width != 1202 || got.Height != 802 {
		t.Fatalf("loadWindowSize() = %.0fx%.0f, want 1202x802", got.Width, got.Height)
	}
}

func TestSaveWindowSize_RoundsValuesToEven(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	prefs := app.Preferences()
	saveWindowSize(prefs, fyne.NewSize(1201, 801))

	if got := prefs.FloatWithFallback(prefKeyWindowWidth, 0); got != 1202 {
		t.Fatalf("saved width = %.0f, want 1202", got)
	}
	if got := prefs.FloatWithFallback(prefKeyWindowHeight, 0); got != 802 {
		t.Fatalf("saved height = %.0f, want 802", got)
	}
}

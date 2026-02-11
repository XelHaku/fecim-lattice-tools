package gui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestShowPhysicsEquationsDialog_NoWindow(t *testing.T) {
	a := &App{}
	a.ShowPhysicsEquationsDialog()
}

func TestShowPhysicsEquationsDialog_AddsOverlay(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(widget.NewLabel("host"))
	a := &App{
		mainWindow: win,
		pauseBtn:   widget.NewButton("Pause", nil),
	}

	before := len(win.Canvas().Overlays().List())
	a.ShowPhysicsEquationsDialog()
	after := len(win.Canvas().Overlays().List())

	if after != before+1 {
		t.Fatalf("expected overlays to increase by 1, got before=%d after=%d", before, after)
	}
	if !a.paused.Load() {
		t.Fatal("expected equation dialog to pause simulation")
	}
}

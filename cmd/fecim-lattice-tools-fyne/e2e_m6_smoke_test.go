package main

import (
	"os"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
)

func TestE2E_M6GUISmoke(t *testing.T) {
	if os.Getenv("FECIM_RUN_E2E_M6_SMOKE") != "1" {
		t.Skip("Set FECIM_RUN_E2E_M6_SMOKE=1 to enable module6 GUI smoke test")
	}
	ensureDisplayForGraphicalTests(t)
	app := test.NewApp()
	defer app.Quit()

	var (
		w       fyne.Window
		demo    = demo6gui.NewEmbeddedEDAApp()
		content fyne.CanvasObject
	)

	uiDo(func() {
		w = app.NewWindow("E2E M6 Smoke")
		w.Resize(fyne.NewSize(1280, 900))
		content = demo.BuildContent(app, w)
		w.SetContent(content)
		w.Show()
		demo.Start()
	})
	defer uiDo(func() {
		demo.Stop()
		w.Close()
	})

	if content == nil {
		t.Fatal("module 6 BuildContent returned nil")
	}

	var gen fyne.Tappable
	for i := 0; i < 20 && gen == nil; i++ {
		if btn := findButton(content, "Generate All"); btn != nil {
			gen = btn
		}
		if gen == nil {
			time.Sleep(100 * time.Millisecond)
		}
	}
	if gen == nil {
		t.Fatal("Generate All button not found")
	}
	uiDo(func() { test.Tap(gen) })
	time.Sleep(120 * time.Millisecond)

	var exportPkg fyne.Tappable
	for i := 0; i < 20 && exportPkg == nil; i++ {
		if btn := findButton(content, "Export Package"); btn != nil {
			exportPkg = btn
		}
		if exportPkg == nil {
			time.Sleep(100 * time.Millisecond)
		}
	}
	if exportPkg == nil {
		t.Fatal("Export Package button not found")
	}
	uiDo(func() { test.Tap(exportPkg) })
	time.Sleep(120 * time.Millisecond)

	uiDo(func() { saveGUISmokeScreenshot(t, w, "m6/smoke.png") })
}

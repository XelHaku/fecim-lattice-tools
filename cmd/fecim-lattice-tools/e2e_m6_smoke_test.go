package main

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	demo6gui "fecim-lattice-tools/module6-eda/pkg/gui"
)

func TestE2E_M6GUISmoke(t *testing.T) {
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

	gen := findButton(content, "Generate All")
	if gen == nil {
		t.Fatal("Generate All button not found")
	}
	uiDo(func() { test.Tap(gen) })
	time.Sleep(120 * time.Millisecond)

	exportPkg := findButton(content, "Export Package")
	if exportPkg == nil {
		t.Fatal("Export Package button not found")
	}
	uiDo(func() { test.Tap(exportPkg) })
	time.Sleep(120 * time.Millisecond)

	uiDo(func() { saveGUISmokeScreenshot(t, w, "m6/smoke.png") })
}

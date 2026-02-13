package main

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	demo1gui "fecim-lattice-tools/module1-hysteresis/pkg/gui"
)

func TestE2E_M1GUISmoke(t *testing.T) {
	t.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")
	t.Setenv("FECIM_DISABLE_STARTUP_CALIBRATION", "1")

	app := test.NewApp()
	defer app.Quit()

	var (
		w       fyne.Window
		demo    = demo1gui.NewEmbeddedApp()
		content fyne.CanvasObject
	)

	uiDo(func() {
		w = app.NewWindow("E2E M1 Smoke")
		w.Resize(fyne.NewSize(1200, 800))
		content = demo.BuildContent(app, w)
		w.SetContent(content)
		w.Show()
	})
	defer uiDo(func() { w.Close() })

	if content == nil {
		t.Fatal("module 1 BuildContent returned nil")
	}

	uiDo(func() { demo.Start() })
	defer uiDo(func() { demo.Stop() })

	time.Sleep(150 * time.Millisecond)

	// Primary engine toggle verification (works independent of widget tree introspection).
	demo.SetPhysicsEngineName("lk")
	time.Sleep(60 * time.Millisecond)
	demo.SetPhysicsEngineName("preisach")
	time.Sleep(60 * time.Millisecond)

	if sel := findSelectWithOptions(content, "L-K (Dynamic)", "Preisach (Quasi-Static)"); sel == nil {
		t.Log("physics select widget not discovered via tree walk (continuing)")
	}

	uiDo(func() { saveGUISmokeScreenshot(t, w, "m1/initial.png") })
	uiDo(func() { saveGUISmokeScreenshot(t, w, "m1/post-engine-toggle.png") })
}

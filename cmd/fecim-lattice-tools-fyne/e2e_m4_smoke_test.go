package main

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	demo4gui "fecim-lattice-tools/module4-circuits/pkg/gui"
)

func TestE2E_M4GUISmoke(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	var (
		w       fyne.Window
		demo    = demo4gui.NewEmbeddedCircuitsApp()
		content fyne.CanvasObject
	)

	uiDo(func() {
		w = app.NewWindow("E2E M4 Smoke")
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
		t.Fatal("module 4 BuildContent returned nil")
	}

	modeSel := findSelectWithOptions(content, "WRITE", "READ", "COMPUTE")
	if modeSel != nil {
		for _, mode := range []string{"WRITE", "READ", "COMPUTE"} {
			uiDo(func() { modeSel.SetSelected(mode) })
			time.Sleep(40 * time.Millisecond)
		}
	} else {
		for _, mode := range []string{"READ", "WRITE", "COMPUTE"} {
			btn := findButton(content, mode)
			if btn == nil {
				btn = findButtonContains(content, mode)
			}
			if btn != nil {
				uiDo(func() { test.Tap(btn) })
				time.Sleep(40 * time.Millisecond)
			}
		}
		t.Log("mode controls not fully discoverable via tree walk (continuing)")
	}

	if arrSize := findSelectWithOptions(content, "1x1", "8x8", "128x128"); arrSize != nil {
		uiDo(func() { arrSize.SetSelected("16x16") })
		time.Sleep(40 * time.Millisecond)
	} else {
		t.Log("array size selector not discovered via tree walk (continuing)")
	}

	for _, arch := range []string{"0T1R", "1T1R", "2T1R"} {
		btn := findButton(content, arch)
		if btn == nil {
			btn = findButtonContains(content, arch)
		}
		if btn != nil {
			uiDo(func() { test.Tap(btn) })
			time.Sleep(40 * time.Millisecond)
		}
	}

	if tier := findSelectWithOptions(content, "Ideal", "Tier-A", "Tier-B"); tier != nil {
		for _, opt := range []string{"Ideal", "Tier-A", "Tier-B"} {
			uiDo(func() { tier.SetSelected(opt) })
			time.Sleep(40 * time.Millisecond)
		}
	} else {
		t.Log("tier selector not discovered via tree walk (continuing)")
	}

	uiDo(func() { saveGUISmokeScreenshot(t, w, "m4/smoke.png") })
}

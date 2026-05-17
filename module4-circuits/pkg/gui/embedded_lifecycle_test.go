//go:build legacy_fyne

package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestEmbeddedCircuitsBuildContentLeavesUIRefreshIdle(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	win := app.NewWindow("circuits")
	defer win.Close()

	embedded := NewEmbeddedCircuitsApp()
	content := embedded.BuildContent(app, win)
	if content == nil {
		t.Fatal("BuildContent returned nil")
	}

	embedded.uiUpdateMu.Lock()
	defer embedded.uiUpdateMu.Unlock()

	if embedded.suspendUIRefresh {
		t.Fatal("UI refresh should be resumed after BuildContent")
	}
	if embedded.pendingUIUpd {
		t.Fatal("BuildContent should not leave a deferred UI refresh pending")
	}
	if embedded.uiUpdateTimer != nil {
		t.Fatal("BuildContent should not leave a deferred UI refresh timer active")
	}
}

func TestEmbeddedCircuitsStopCancelsPendingUIRefresh(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	win := app.NewWindow("circuits")
	defer win.Close()

	embedded := NewEmbeddedCircuitsApp()
	embedded.BuildContent(app, win)

	timer := time.NewTimer(time.Minute)
	defer timer.Stop()

	embedded.uiUpdateMu.Lock()
	embedded.pendingUIUpd = true
	embedded.uiUpdateTimer = timer
	embedded.uiUpdateMu.Unlock()

	embedded.Stop()

	embedded.uiUpdateMu.Lock()
	defer embedded.uiUpdateMu.Unlock()

	if embedded.pendingUIUpd {
		t.Fatal("Stop should clear pending UI refresh state")
	}
	if embedded.uiUpdateTimer != nil {
		t.Fatal("Stop should clear the deferred UI refresh timer")
	}
}

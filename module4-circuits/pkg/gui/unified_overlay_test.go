//go:build legacy_fyne

package gui

import (
	"image"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func imageChecksum(img image.Image) uint64 {
	bounds := img.Bounds()
	var sum uint64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			sum += uint64(r) + uint64(g) + uint64(b) + uint64(a)
		}
	}
	return sum
}

func TestReadOverlaySelectorUpdatesMode(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca.readOverlaySelect == nil {
		t.Fatal("expected read overlay selector")
	}
	if got := ca.readOverlaySelect.Selected; got != "Off" {
		t.Fatalf("default overlay: got %q, want Off", got)
	}

	ca.readOverlaySelect.SetSelected("Icell")
	waitFor(t, 500*time.Millisecond, "overlay Icell", func() bool {
		ca.mu.RLock()
		defer ca.mu.RUnlock()
		return ca.readOverlayMode == "Icell"
	})
}

func TestReadOverlayAffectsAllOperationModes(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp

	// READ mode default. Off vs Vcell should render differently.
	if ca.deviceState.GetOperationMode() != OpModeRead {
		t.Fatalf("expected default read mode")
	}

	ca.readOverlaySelect.SetSelected("Off")
	offRead := imageChecksum(ca.drawUnifiedArray(900, 700))
	ca.readOverlaySelect.SetSelected("Vcell")
	vcellRead := imageChecksum(ca.drawUnifiedArray(900, 700))
	if offRead == vcellRead {
		t.Fatalf("expected overlay to change read canvas rendering")
	}

	// WRITE mode should also reflect per-cell overlay toggles.
	test.Tap(ca.modeWriteBtn)
	waitFor(t, 500*time.Millisecond, "write mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeWrite
	})

	ca.readOverlaySelect.SetSelected("Off")
	offWrite := imageChecksum(ca.drawUnifiedArray(900, 700))
	ca.readOverlaySelect.SetSelected("Vcell")
	vcellWrite := imageChecksum(ca.drawUnifiedArray(900, 700))
	if offWrite == vcellWrite {
		t.Fatalf("expected overlay selection to affect WRITE rendering")
	}

	// COMPUTE mode should also keep overlays available.
	test.Tap(ca.modeComputeBtn)
	waitFor(t, 500*time.Millisecond, "compute mode", func() bool {
		return ca.deviceState.GetOperationMode() == OpModeCompute
	})
	ca.readOverlaySelect.SetSelected("Off")
	offCompute := imageChecksum(ca.drawUnifiedArray(900, 700))
	ca.readOverlaySelect.SetSelected("Icell")
	iCellCompute := imageChecksum(ca.drawUnifiedArray(900, 700))
	if offCompute == iCellCompute {
		t.Fatalf("expected overlay selection to affect COMPUTE rendering")
	}
}

func TestFormatSignedScaled(t *testing.T) {
	if got := formatSignedScaled(0, []scaledUnit{{unit: "V", scale: 1}, {unit: "mV", scale: 1e-3}}); got != "0 V" {
		t.Fatalf("zero format: got %q", got)
	}
	if got := formatSignedScaled(1.2e-3, []scaledUnit{{unit: "A", scale: 1}, {unit: "mA", scale: 1e-3}, {unit: "uA", scale: 1e-6}}); got != "+1.20 mA" {
		t.Fatalf("milli scaling: got %q", got)
	}
	if got := formatSignedScaled(-4.56e-6, []scaledUnit{{unit: "A", scale: 1}, {unit: "mA", scale: 1e-3}, {unit: "uA", scale: 1e-6}}); got != "-4.56 uA" {
		t.Fatalf("sign handling: got %q", got)
	}
}

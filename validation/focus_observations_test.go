//go:build legacy_fyne

package validation

import (
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	circuitsgui "fecim-lattice-tools/module4-circuits/pkg/gui"
)

func TestFocus91_WriteRangeIsBipolarAndCode0IsNegativeEdge(t *testing.T) {
	ds := circuitsgui.NewDeviceState(4, 4, nil, nil)
	wr := ds.GetWriteRange()
	if wr.Min >= 0 || wr.Max <= 0 {
		t.Fatalf("write range must be bipolar, got [%.6f, %.6f]", wr.Min, wr.Max)
	}
	if math.Abs(wr.Max+wr.Min) > 1e-9 {
		t.Fatalf("write range must be symmetric around 0, got [%.6f, %.6f]", wr.Min, wr.Max)
	}
	applied, code := ds.DACWriteVoltage(wr.Min)
	if code != 0 {
		t.Fatalf("DAC code 0 should map to negative edge; got code=%d applied=%.6f", code, applied)
	}
}

func TestFocus92_Module4LayoutRemovesViewDropdownLabel(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	win := test.NewWindow(nil)
	defer win.Close()

	embedded := circuitsgui.NewEmbeddedCircuitsApp()
	content := embedded.BuildContent(app, win)
	defer embedded.Stop()

	if hasLabelText(content, "View:") {
		t.Fatalf("expected View dropdown header label to be removed (OPERATIONS-only mode)")
	}
}

func hasLabelText(obj fyne.CanvasObject, want string) bool {
	switch o := obj.(type) {
	case *widget.Label:
		return o.Text == want
	case *fyne.Container:
		for _, child := range o.Objects {
			if hasLabelText(child, want) {
				return true
			}
		}
	case *container.Split:
		if hasLabelText(o.Leading, want) || hasLabelText(o.Trailing, want) {
			return true
		}
	}
	return false
}

func TestFocus98_ActiveArchitecturesClampTinyResidualVoltagesToZero(t *testing.T) {
	ds := circuitsgui.NewDeviceState(2, 2, nil, nil)
	ds.SetPassiveMode(false) // 1T1R/2T1R-style isolation path
	ds.SetDACVoltage(0, 5e-13)
	if got := ds.GetEffectiveCellVoltage(0, 0); got != 0 {
		t.Fatalf("expected tiny residual to clamp to 0V, got %.3eV", got)
	}
}

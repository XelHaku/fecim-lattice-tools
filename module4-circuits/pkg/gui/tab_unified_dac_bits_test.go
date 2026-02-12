package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func TestDACBitsSelectorUpdatesDeviceState(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	if ca == nil {
		t.Fatal("expected circuits app")
	}

	obj := ca.createDACBitsSelector()
	hbox, ok := obj.(*fyne.Container)
	if !ok {
		t.Fatal("expected HBox container")
	}

	var selector *widget.Select
	for _, child := range hbox.Objects {
		if s, ok := child.(*widget.Select); ok {
			selector = s
			break
		}
	}
	if selector == nil {
		t.Fatal("expected DAC selector widget")
	}

	selector.SetSelected("8-bit (256)")
	if got := ca.deviceState.GetDACBits(); got != 8 {
		t.Fatalf("expected device DAC bits=8, got %d", got)
	}

	selector.SetSelected("4-bit (16)")
	if got := ca.deviceState.GetDACBits(); got != 4 {
		t.Fatalf("expected device DAC bits=4, got %d", got)
	}
	if ca.dacBits != 4 {
		t.Fatalf("expected app dacBits=4, got %d", ca.dacBits)
	}
}

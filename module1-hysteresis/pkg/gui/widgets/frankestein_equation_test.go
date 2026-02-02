package widgets

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestNewFrankesteinEquationWidget(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(widget.NewLabel("host"))

	obj := NewFrankesteinEquationWidget(win)
	if obj == nil {
		t.Fatal("NewFrankesteinEquationWidget should return non-nil")
	}

	container, ok := obj.(*fyne.Container)
	if !ok {
		t.Fatalf("expected *fyne.Container, got %T", obj)
	}

	if len(container.Objects) < 2 {
		t.Errorf("expected at least 2 top-level objects, got %d", len(container.Objects))
	}
}

func TestTermChipSelection(t *testing.T) {
	test.NewApp()
	_ = test.NewWindow(widget.NewLabel("host"))

	var gotID, gotFallback string
	chip := NewTermChip("rho_eff_main", "\\rho_{eff}", "Effective viscosity tooltip", func(id, fallback string) {
		gotID = id
		gotFallback = fallback
	})
	chip.Tapped(nil)
	if gotID != "rho_eff_main" {
		t.Fatalf("expected term ID, got %q", gotID)
	}
	if gotFallback == "" {
		t.Fatal("expected fallback tooltip to be passed")
	}
}

func TestTermChipSelectionNoCallback(t *testing.T) {
	test.NewApp()
	_ = test.NewWindow(widget.NewLabel("host"))

	chip := NewTermChip("alpha", "\\alpha", "", nil)
	chip.Tapped(nil)
}

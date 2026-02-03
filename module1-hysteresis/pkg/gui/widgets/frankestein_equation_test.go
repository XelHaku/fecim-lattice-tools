package widgets

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

	root, ok := obj.(*fyne.Container)
	if !ok {
		t.Fatalf("expected *fyne.Container, got %T", obj)
	}

	if len(root.Objects) < 1 {
		t.Errorf("expected at least 1 top-level object, got %d", len(root.Objects))
	}

	var tabs *container.AppTabs
	for _, child := range root.Objects {
		if t, ok := child.(*container.AppTabs); ok {
			tabs = t
			break
		}
	}
	if tabs == nil {
		t.Fatal("expected AppTabs in equation widget")
	}
	if len(tabs.Items) < 2 {
		t.Fatalf("expected at least 2 tabs, got %d", len(tabs.Items))
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

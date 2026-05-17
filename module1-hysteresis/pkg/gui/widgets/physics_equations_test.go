//go:build legacy_fyne

package widgets

import (
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestNewPhysicsEquationsWidget(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(widget.NewLabel("host"))

	obj := NewPhysicsEquationsWidget(win)
	if obj == nil {
		t.Fatal("NewPhysicsEquationsWidget should return non-nil")
	}

	tabs, ok := obj.(*container.AppTabs)
	if !ok {
		t.Fatalf("expected *container.AppTabs, got %T", obj)
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

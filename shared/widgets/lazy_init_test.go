package widgets

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func TestLazyTabItem_DeferredCreation(t *testing.T) {
	lazy := NewLazyTabItem("Heavy", func() fyne.CanvasObject {
		return widget.NewLabel("loaded")
	})

	if lazy.CurrentContent() != nil {
		t.Fatalf("expected nil content before first show")
	}

	_ = lazy.Content()
	if lazy.CurrentContent() == nil {
		t.Fatalf("expected non-nil content after first show")
	}
}

func TestApplyToTabs_LoadsOnSelection(t *testing.T) {
	lazy := NewLazyTabItem("Heavy", func() fyne.CanvasObject {
		return widget.NewLabel("loaded")
	})
	tabs := container.NewAppTabs(
		container.NewTabItem("Light", widget.NewLabel("light")),
		container.NewTabItem("Heavy", widget.NewLabel("placeholder")),
	)
	ApplyToTabs(tabs, map[string]*LazyTabItem{"Heavy": lazy})

	if lazy.CurrentContent() != nil {
		t.Fatalf("expected nil content before selection")
	}

	tabs.SelectIndex(1)
	if lazy.CurrentContent() == nil {
		t.Fatalf("expected content initialized after selection")
	}
}

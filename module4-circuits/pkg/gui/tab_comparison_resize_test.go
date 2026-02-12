package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func TestComparisonTab_HasScrollGuardsForResize(t *testing.T) {
	ca := &CircuitsApp{}
	tab := ca.createComparisonTab()

	var hScrolls, vScrolls int
	walkCanvas(tab, func(obj fyne.CanvasObject) {
		s, ok := obj.(*container.Scroll)
		if !ok {
			return
		}
		switch s.Direction {
		case fyne.ScrollHorizontalOnly:
			hScrolls++
		case fyne.ScrollVerticalOnly:
			vScrolls++
		}
	})

	if hScrolls == 0 {
		t.Fatal("expected at least one horizontal scroll container to prevent clipping on narrow widths")
	}
	if vScrolls == 0 {
		t.Fatal("expected at least one vertical scroll container to prevent clipping on short heights")
	}
}

func walkCanvas(obj fyne.CanvasObject, visit func(fyne.CanvasObject)) {
	if obj == nil {
		return
	}
	visit(obj)

	switch v := obj.(type) {
	case *fyne.Container:
		for _, child := range v.Objects {
			walkCanvas(child, visit)
		}
	case *container.Scroll:
		walkCanvas(v.Content, visit)
	}
}

// Package tabs provides individual tab components for the Demo 2 crossbar GUI.
package tabs

import (
	"fmt"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/crossbar"
)

// IdealTab provides the ideal MVM crossbar visualization.
type IdealTab struct {
	array       *crossbar.Array
	heatmapView fyne.CanvasObject
	statsLabel  *widget.Label
	statusLabel *widget.Label

	// Callbacks
	onArrayChange func()
}

// NewIdealTab creates a new ideal MVM tab.
func NewIdealTab(array *crossbar.Array, onArrayChange func()) *IdealTab {
	tab := &IdealTab{
		array:         array,
		statsLabel:    widget.NewLabel(""),
		statusLabel:   widget.NewLabel("Ready"),
		onArrayChange: onArrayChange,
	}
	return tab
}

// SetArray updates the array reference.
func (t *IdealTab) SetArray(array *crossbar.Array) {
	t.array = array
}

// Content returns the tab content.
func (t *IdealTab) Content() fyne.CanvasObject {
	// Left side: Crossbar heatmap placeholder
	heatmapPlaceholder := widget.NewLabel("Crossbar Array Heatmap\n(Click cells to program weights)")
	heatmapPlaceholder.Alignment = fyne.TextAlignCenter

	heatmapContainer := container.NewCenter(heatmapPlaceholder)

	// Right side: Controls
	arraySizeSelect := widget.NewSelect([]string{"8x8", "16x16", "32x32", "64x64"}, func(s string) {
		t.statusLabel.SetText(fmt.Sprintf("Array size: %s", s))
		if t.onArrayChange != nil {
			t.onArrayChange()
		}
	})
	arraySizeSelect.SetSelected("16x16")

	programRandomBtn := widget.NewButton("Program Random Weights", func() {
		if t.array != nil {
			rows := t.array.Rows()
			cols := t.array.Cols()
			for i := 0; i < rows; i++ {
				for j := 0; j < cols; j++ {
					t.array.ProgramWeight(i, j, rand.Float64())
				}
			}
			t.statusLabel.SetText("Random weights programmed")
		}
	})

	clearBtn := widget.NewButton("Clear Array", func() {
		if t.array != nil {
			rows := t.array.Rows()
			cols := t.array.Cols()
			for i := 0; i < rows; i++ {
				for j := 0; j < cols; j++ {
					t.array.ProgramWeight(i, j, 0)
				}
			}
			t.statusLabel.SetText("Array cleared")
		}
	})

	runMVMBtn := widget.NewButton("Run MVM", func() {
		t.statusLabel.SetText("MVM operation complete")
	})

	statsText := `Ideal MVM Mode
==============

In this mode, the crossbar array
performs matrix-vector multiplication
without any non-idealities.

Each cell stores a conductance value
(demo baseline: 30 levels, conference claim).

Output = G × V
(Kirchhoff's current law)

This is the theoretical best case.`

	t.statsLabel.SetText(statsText)
	t.statsLabel.Wrapping = fyne.TextWrapOff
	t.statsLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Wrap stats in scroll to prevent layout resize
	statsScroll := container.NewVScroll(t.statsLabel)
	statsScroll.SetMinSize(fyne.NewSize(240, 150))

	controls := container.NewVBox(
		widget.NewLabel("Array Size:"),
		arraySizeSelect,
		widget.NewSeparator(),
		programRandomBtn,
		clearBtn,
		widget.NewSeparator(),
		runMVMBtn,
		widget.NewSeparator(),
		statsScroll,
	)

	content := container.NewHSplit(
		container.NewBorder(
			widget.NewLabelWithStyle("Crossbar Array (Ideal)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			t.statusLabel,
			nil, nil,
			heatmapContainer,
		),
		controls,
	)
	content.SetOffset(0.65)

	return content
}

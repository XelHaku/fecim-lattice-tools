//go:build legacy_fyne

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

func (a *App) showMaterialPickerDialog() {
	if a == nil || a.mainWindow == nil {
		return
	}

	resume := a.pauseSimulationForModal()
	picker := sharedwidgets.NewMaterialPicker(nil)

	// Pre-select current material.
	currentID := a.getCurrentMaterialID()
	if currentID != "" {
		picker.SetSelected(currentID)
	}

	d := dialog.NewCustomConfirm(
		"Select Ferroelectric Material",
		"Select",
		"Cancel",
		picker,
		func(confirmed bool) {
			defer resume()
			if !confirmed {
				return
			}
			id, mat := picker.GetSelected()
			if id == "" || mat == nil {
				return
			}
			a.onMaterialPickerSelected(id, mat)
		},
		a.mainWindow,
	)
	// Also resume if the dialog is dismissed by escape/window close.
	d.SetOnClosed(resume)

	// Responsive sizing: never exceed the window; keep a sensible minimum.
	canvasSize := a.mainWindow.Canvas().Size()
	w := float32(1150)
	h := float32(520)
	if canvasSize.Width > 0 {
		maxW := canvasSize.Width * 0.95
		if w > maxW {
			w = maxW
		}
		if w < 680 {
			w = 680
		}
	}
	if canvasSize.Height > 0 {
		maxH := canvasSize.Height * 0.90
		if h > maxH {
			h = maxH
		}
		if h < 420 {
			h = 420
		}
	}
	if w <= 0 {
		w = 900
	}
	if h <= 0 {
		h = 500
	}

	d.Resize(fyne.NewSize(w, h))
	d.Show()
}

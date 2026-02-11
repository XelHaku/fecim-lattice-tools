package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module1-hysteresis/pkg/gui/widgets"
)

// showPhysicsEquationsDialog opens the equations modal. Kept unexported for normal UI wiring.
func (a *App) showPhysicsEquationsDialog() {
	if a.mainWindow == nil {
		return
	}

	wasPaused := a.paused.Load()
	if !wasPaused {
		a.paused.Store(true)
		if a.pauseBtn != nil {
			a.pauseBtn.SetText("Resume")
		}
	}

	initialTab := 0 // L-K
	if a.physicsEngine == PhysicsPreisach {
		initialTab = 1 // Preisach
	}
	content := widgets.NewPhysicsEquationsWidget(a.mainWindow, initialTab)
	canvasSize := a.mainWindow.Canvas().Size()
	width := canvasSize.Width * 0.92
	height := canvasSize.Height * 0.95
	if width <= 0 {
		width = 900
	}
	if height <= 0 {
		height = 700
	}
	framed := container.NewPadded(content)

	var dialog *widget.PopUp
	closeBtn := widget.NewButton("Close", func() {
		if dialog != nil {
			dialog.Hide()
		}
		if !wasPaused && a.paused.Load() {
			a.paused.Store(false)
			if a.pauseBtn != nil {
				a.pauseBtn.SetText("Pause")
			}
		}
	})

	dialog = widget.NewModalPopUp(
		container.NewBorder(nil, closeBtn, nil, nil, framed),
		a.mainWindow.Canvas(),
	)
	dialog.Resize(fyne.NewSize(width, height))

	dialog.Show()
}

// ShowPhysicsEquationsDialogForCapture is a small exported wrapper used by the
// headless screenshot harness (scripts/capture_fyne_window.go).
func (a *App) ShowPhysicsEquationsDialogForCapture() {
	a.showPhysicsEquationsDialog()
}

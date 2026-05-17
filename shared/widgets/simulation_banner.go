//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
//
// simulation_banner.go provides a thin amber banner reminding users that
// all results are simulation-only and not validated against fabricated devices.
package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// NewSimulationBanner returns a thin horizontal bar with an amber background
// and white text reading "Simulation Only -- Not Validated Against Fabricated Devices".
func NewSimulationBanner() *fyne.Container {
	bg := canvas.NewRectangle(color.RGBA{180, 130, 20, 255})
	bg.SetMinSize(fyne.NewSize(0, 24))
	icon := canvas.NewText("\u26A0", color.RGBA{255, 230, 160, 255}) // Warning triangle
	icon.TextSize = 12
	text := canvas.NewText(" Simulation Only \u2014 Not Validated Against Fabricated Devices", color.RGBA{255, 245, 220, 255})
	text.TextSize = 11
	row := container.NewHBox(icon, text)
	return container.NewStack(bg, container.NewCenter(row))
}

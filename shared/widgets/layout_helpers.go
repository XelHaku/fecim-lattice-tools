//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CreateLabeledBox creates a styled box with a title and static value text.
// Useful for displaying static information in a visually distinct container.
//
// Parameters:
//   - title: Bold text displayed at the top of the box
//   - value: Regular text displayed below the title
//   - bgColor: Background color for the box
//
// Returns a container with centered content over a colored background.
func CreateLabeledBox(title, value string, bgColor color.Color) *fyne.Container {
	titleLbl := widget.NewLabel(title)
	titleLbl.TextStyle = fyne.TextStyle{Bold: true}
	titleLbl.Alignment = fyne.TextAlignCenter

	valueLbl := widget.NewLabel(value)
	valueLbl.Alignment = fyne.TextAlignCenter

	bg := canvas.NewRectangle(bgColor)
	bg.SetMinSize(fyne.NewSize(100, 60))
	bg.CornerRadius = 5

	content := container.NewVBox(titleLbl, valueLbl)

	return container.NewStack(bg, container.NewCenter(content))
}

// CreateLabeledBoxWithLabel creates a styled box with a title and a widget.Label
// for dynamic updates. The caller provides the value label so it can be updated
// programmatically.
//
// Parameters:
//   - title: Bold text displayed at the top of the box
//   - valueLbl: Label widget for the value (caller manages content)
//   - bgColor: Background color for the box
//
// Returns a container with centered content over a colored background.
func CreateLabeledBoxWithLabel(title string, valueLbl *widget.Label, bgColor color.Color) *fyne.Container {
	titleLbl := widget.NewLabel(title)
	titleLbl.TextStyle = fyne.TextStyle{Bold: true}
	titleLbl.Alignment = fyne.TextAlignCenter

	valueLbl.Alignment = fyne.TextAlignCenter

	bg := canvas.NewRectangle(bgColor)
	bg.SetMinSize(fyne.NewSize(100, 60))
	bg.CornerRadius = 5

	content := container.NewVBox(titleLbl, valueLbl)

	return container.NewStack(bg, container.NewCenter(content))
}

// CreateSectionDivider creates a subtle horizontal divider line.
// Useful for visually separating sections in a UI.
//
// Parameters:
//   - dividerColor: Color for the divider line (use subtle/transparent colors)
//
// Returns a padded container with a thin rectangle.
func CreateSectionDivider(dividerColor color.Color) fyne.CanvasObject {
	line := canvas.NewRectangle(dividerColor)
	line.SetMinSize(fyne.NewSize(0, 1))
	return container.NewPadded(line)
}

// CreateSectionDividerDefault creates a divider with a default blue-tinted color.
// Convenient shortcut for common usage.
func CreateSectionDividerDefault() fyne.CanvasObject {
	return CreateSectionDivider(color.RGBA{0, 100, 180, 100})
}

// CreateSectionHeader creates a bold, centered header with a separator below it.
// Useful for creating titled sections in a UI.
//
// Parameters:
//   - title: The header text to display
//
// Returns a container with the title and a separator.
func CreateSectionHeader(title string) *fyne.Container {
	titleLbl := widget.NewLabelWithStyle(
		title,
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	return container.NewVBox(titleLbl, widget.NewSeparator())
}

//go:build legacy_fyne

package widgets

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewModuleErrorContent returns a consistent fallback view for embeddable
// modules that fail during initialization.
func NewModuleErrorContent(moduleName string, err error) fyne.CanvasObject {
	if moduleName == "" {
		moduleName = "FeCIM"
	}

	message := "The module could not be initialized."
	if err != nil {
		message = err.Error()
	}

	detail := widget.NewLabel(message)
	detail.Wrapping = fyne.TextWrapWord

	return container.NewPadded(
		widget.NewCard(
			fmt.Sprintf("%s module unavailable", moduleName),
			"Initialization failed before the interface could be created.",
			detail,
		),
	)
}

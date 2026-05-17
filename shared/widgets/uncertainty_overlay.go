//go:build legacy_fyne

package widgets

import (
	"fmt"

	"fyne.io/fyne/v2/widget"
)

// FormatUncertaintyAnnotation returns text in the required form:
// "value ± uncertainty (confidence%)".
func FormatUncertaintyAnnotation(value, uncertainty float64, confidencePct float64) string {
	return fmt.Sprintf("%.4g ± %.4g (%.1f%%)", value, uncertainty, confidencePct)
}

// NewUncertaintyOverlay creates a readout annotation label for UI overlays.
// Text format is fixed as: "value ± uncertainty (confidence%)".
func NewUncertaintyOverlay(value, uncertainty float64, confidencePct float64) *widget.Label {
	return widget.NewLabel(FormatUncertaintyAnnotation(value, uncertainty, confidencePct))
}

//go:build legacy_fyne

package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/physics"
)

// ProvenanceLabel is a composite widget that displays a formatted value,
// an inline confidence badge (colored dot + provenance level), and an
// optional source-reference subtitle.
type ProvenanceLabel struct {
	widget.BaseWidget

	// FormattedValue is the display string, e.g. "24.5 uC/cm^2".
	FormattedValue string
	// Provenance indicates the data-quality tier.
	Provenance physics.Provenance
	// SourceRef is an optional citation shown as a subtitle,
	// e.g. "Park et al., Adv. Mater. 2015".
	SourceRef string
}

// NewProvenanceLabel creates a ProvenanceLabel and initialises its base widget.
func NewProvenanceLabel(value string, prov physics.Provenance, sourceRef string) *ProvenanceLabel {
	if prov == "" {
		prov = physics.ProvenancePlaceholder
	}
	pl := &ProvenanceLabel{
		FormattedValue: value,
		Provenance:     prov,
		SourceRef:      sourceRef,
	}
	pl.ExtendBaseWidget(pl)
	return pl
}

// provenanceColor returns the badge colour for a given provenance tier.
//
//	Measured    -> green
//	Calibrated  -> yellow
//	Estimated   -> orange
//	Placeholder -> red
func provenanceColor(p physics.Provenance) color.Color {
	switch p {
	case physics.ProvenanceMeasured:
		return color.RGBA{R: 60, G: 180, B: 75, A: 255}
	case physics.ProvenanceCalibrated:
		return color.RGBA{R: 240, G: 190, B: 60, A: 255}
	case physics.ProvenanceEstimated:
		return color.RGBA{R: 230, G: 130, B: 40, A: 255}
	default: // Placeholder or unknown
		return color.RGBA{R: 220, G: 80, B: 80, A: 255}
	}
}

// provenanceDisplayName returns a human-readable label for a provenance tier.
func provenanceDisplayName(p physics.Provenance) string {
	switch p {
	case physics.ProvenanceMeasured:
		return "Measured"
	case physics.ProvenanceCalibrated:
		return "Calibrated"
	case physics.ProvenanceEstimated:
		return "Estimated"
	default:
		return "Placeholder"
	}
}

// CreateRenderer implements fyne.Widget.
func (pl *ProvenanceLabel) CreateRenderer() fyne.WidgetRenderer {
	// Value label (bold).
	valLabel := widget.NewLabel(pl.FormattedValue)
	valLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Coloured dot.
	dot := canvas.NewCircle(provenanceColor(pl.Provenance))
	dotWrap := container.NewGridWrap(fyne.NewSize(10, 10), dot)

	// Provenance-level text.
	levelLabel := widget.NewLabel(provenanceDisplayName(pl.Provenance))
	levelLabel.TextStyle = fyne.TextStyle{Italic: true}

	badge := container.NewHBox(dotWrap, levelLabel)
	row := container.NewHBox(valLabel, badge)

	// Optional source reference subtitle.
	if pl.SourceRef != "" {
		srcLabel := widget.NewLabel(pl.SourceRef)
		srcLabel.TextStyle = fyne.TextStyle{Italic: true}
		srcLabel.Importance = widget.LowImportance
		return widget.NewSimpleRenderer(container.NewVBox(row, srcLabel))
	}
	return widget.NewSimpleRenderer(row)
}

// Package widgets provides custom Fyne widgets for the hysteresis visualization.
package widgets

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// SimVsExpComparison displays a comparison between simulated and experimental
// hysteresis loop characteristics. This provides transparency about model accuracy.
type SimVsExpComparison struct {
	widget.BaseWidget

	// Simulated values (from Preisach model)
	simPr         float64 // Remanent polarization (C/m²)
	simEc         float64 // Coercive field (V/m)
	simSquareness float64 // Pr/Ps ratio

	// Experimental/literature values
	expPrMin     float64
	expPrMax     float64
	expEcMin     float64
	expEcMax     float64
	expSquareMin float64
	expSquareMax float64
	litSource    string

	// Mutable UI elements — stored as struct fields so Refresh() can update
	// them in-place without recreating the entire content tree.
	// NewSimpleRenderer captures a single content reference at CreateRenderer()
	// time; rebuilding s.content in Refresh() would leave the renderer pointing
	// at the original (stale) object. Instead we mutate individual labels.
	prSimLabel   *widget.Label
	prRangeLabel *widget.Label
	prStatusText *canvas.Text
	ecSimLabel   *widget.Label
	ecRangeLabel *widget.Label
	ecStatusText *canvas.Text
	sqSimLabel   *widget.Label
	sqRangeLabel *widget.Label
	sqStatusText *canvas.Text
	sourceLabel  *widget.Label

	content fyne.CanvasObject
}

// NewSimVsExpComparison creates a new simulation vs experiment comparison widget.
func NewSimVsExpComparison() *SimVsExpComparison {
	s := &SimVsExpComparison{
		// Default literature values from Park 2015 HZO
		expPrMin:     15e-2, // 15 µC/cm²
		expPrMax:     34e-2, // 34 µC/cm²
		expEcMin:     0.6e8, // 0.6 MV/cm
		expEcMax:     1.5e8, // 1.5 MV/cm
		expSquareMin: 0.60,
		expSquareMax: 0.95,
		litSource:    "HZO literature envelope (e.g., Park et al., Nature 2015; follow-up reports)",
	}
	s.ExtendBaseWidget(s)
	return s
}

// SetSimulatedValues updates the simulated values from the Preisach model.
func (s *SimVsExpComparison) SetSimulatedValues(pr, ec, squareness float64) {
	s.simPr = pr
	s.simEc = ec
	s.simSquareness = squareness
	s.Refresh()
}

// SetLiteratureSource sets the literature reference being compared against.
func (s *SimVsExpComparison) SetLiteratureSource(source string) {
	s.litSource = source
	s.Refresh()
}

// SetExperimentalBounds sets the experimental/literature bounds for comparison.
func (s *SimVsExpComparison) SetExperimentalBounds(prMin, prMax, ecMin, ecMax, sqMin, sqMax float64) {
	s.expPrMin = prMin
	s.expPrMax = prMax
	s.expEcMin = ecMin
	s.expEcMax = ecMax
	s.expSquareMin = sqMin
	s.expSquareMax = sqMax
	s.Refresh()
}

// checkInRange returns a status indicator based on whether value is in range.
func checkInRange(value, min, max float64) (string, color.Color) {
	if value >= min && value <= max {
		return "✓", color.RGBA{0, 200, 0, 255} // Green check
	}
	// Calculate how far out of range
	var deviation float64
	if value < min {
		deviation = (min - value) / min
	} else {
		deviation = (value - max) / max
	}
	if deviation < 0.2 {
		return "~", color.RGBA{255, 200, 0, 255} // Yellow warning (within 20%)
	}
	return "✗", color.RGBA{255, 80, 80, 255} // Red X (>20% out of range)
}

// CreateRenderer implements fyne.Widget. Called once by Fyne when the widget
// is first laid out. Subsequent data changes go through Refresh().
func (s *SimVsExpComparison) CreateRenderer() fyne.WidgetRenderer {
	s.content = s.createContent()
	return widget.NewSimpleRenderer(s.content)
}

func (s *SimVsExpComparison) createContent() fyne.CanvasObject {
	// Header
	headerBg := canvas.NewRectangle(color.RGBA{30, 60, 90, 255})
	headerText := canvas.NewText("SIMULATION vs EXPERIMENT", color.RGBA{0, 212, 255, 255})
	headerText.TextSize = 14
	headerText.TextStyle = fyne.TextStyle{Bold: true}
	headerText.Alignment = fyne.TextAlignCenter

	header := container.NewStack(
		headerBg,
		container.NewCenter(headerText),
	)

	// Context banner
	disclaimerBg := canvas.NewRectangle(color.RGBA{35, 55, 40, 255})
	disclaimerText := canvas.NewText("Model output compared against literature ranges (not device-specific lab data)", color.RGBA{170, 235, 180, 255})
	disclaimerText.TextSize = 13
	disclaimerText.Alignment = fyne.TextAlignCenter
	disclaimer := container.NewStack(
		disclaimerBg,
		container.NewCenter(disclaimerText),
	)

	// Column headers
	colParam := widget.NewLabelWithStyle("Parameter", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	colSim := widget.NewLabelWithStyle("Simulation", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	colExp := widget.NewLabelWithStyle("Literature range", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	colStatus := widget.NewLabelWithStyle("Status", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	headerRow := container.NewGridWithColumns(4, colParam, colSim, colExp, colStatus)

	// Data rows — create mutable widgets and store refs for Refresh()
	rows := container.NewVBox()

	// Pr row
	prStatus, prColor := checkInRange(s.simPr, s.expPrMin, s.expPrMax)
	s.prStatusText = canvas.NewText(prStatus, prColor)
	s.prStatusText.TextStyle = fyne.TextStyle{Bold: true}
	s.prStatusText.Alignment = fyne.TextAlignCenter

	s.prSimLabel = widget.NewLabel(fmt.Sprintf("%.1f", s.simPr*100))
	s.prRangeLabel = widget.NewLabel(fmt.Sprintf("%.1f - %.1f", s.expPrMin*100, s.expPrMax*100))

	prRow := container.NewGridWithColumns(4,
		widget.NewLabel("Pr (µC/cm²)"),
		s.prSimLabel,
		s.prRangeLabel,
		container.NewCenter(s.prStatusText),
	)
	rows.Add(prRow)

	// Ec row
	ecStatus, ecColor := checkInRange(s.simEc, s.expEcMin, s.expEcMax)
	s.ecStatusText = canvas.NewText(ecStatus, ecColor)
	s.ecStatusText.TextStyle = fyne.TextStyle{Bold: true}
	s.ecStatusText.Alignment = fyne.TextAlignCenter

	s.ecSimLabel = widget.NewLabel(fmt.Sprintf("%.2f", s.simEc/1e8))
	s.ecRangeLabel = widget.NewLabel(fmt.Sprintf("%.1f - %.1f", s.expEcMin/1e8, s.expEcMax/1e8))

	ecRow := container.NewGridWithColumns(4,
		widget.NewLabel("Ec (MV/cm)"),
		s.ecSimLabel,
		s.ecRangeLabel,
		container.NewCenter(s.ecStatusText),
	)
	rows.Add(ecRow)

	// Squareness row
	sqStatus, sqColor := checkInRange(s.simSquareness, s.expSquareMin, s.expSquareMax)
	s.sqStatusText = canvas.NewText(sqStatus, sqColor)
	s.sqStatusText.TextStyle = fyne.TextStyle{Bold: true}
	s.sqStatusText.Alignment = fyne.TextAlignCenter

	s.sqSimLabel = widget.NewLabel(fmt.Sprintf("%.2f", s.simSquareness))
	s.sqRangeLabel = widget.NewLabel(fmt.Sprintf("%.2f - %.2f", s.expSquareMin, s.expSquareMax))

	sqRow := container.NewGridWithColumns(4,
		widget.NewLabel("Squareness"),
		s.sqSimLabel,
		s.sqRangeLabel,
		container.NewCenter(s.sqStatusText),
	)
	rows.Add(sqRow)

	// Source citation
	s.sourceLabel = widget.NewLabel("Source: " + s.litSource)
	s.sourceLabel.TextStyle = fyne.TextStyle{Italic: true}
	s.sourceLabel.Wrapping = fyne.TextWrapWord

	// Legend
	legendGreen := canvas.NewText("✓ Within range", color.RGBA{0, 200, 0, 255})
	legendGreen.TextSize = 14
	legendYellow := canvas.NewText("~ Close (<20% off)", color.RGBA{255, 200, 0, 255})
	legendYellow.TextSize = 14
	legendRed := canvas.NewText("✗ Out of range", color.RGBA{255, 80, 80, 255})
	legendRed.TextSize = 14

	legend := container.NewHBox(legendGreen, legendYellow, legendRed)

	// Placeholder for future experimental data upload
	uploadPlaceholder := widget.NewLabel("[ Future: Upload your experimental P-E loop data for comparison ]")
	uploadPlaceholder.TextStyle = fyne.TextStyle{Italic: true}
	uploadPlaceholder.Alignment = fyne.TextAlignCenter

	return container.NewVBox(
		header,
		disclaimer,
		widget.NewSeparator(),
		headerRow,
		widget.NewSeparator(),
		rows,
		widget.NewSeparator(),
		s.sourceLabel,
		legend,
		widget.NewSeparator(),
		uploadPlaceholder,
	)
}

// Refresh updates the widget display. Updates mutable labels in-place so the
// renderer's content reference stays valid (avoids the NewSimpleRenderer stale
// reference problem).
func (s *SimVsExpComparison) Refresh() {
	if s.prSimLabel != nil {
		// Pr row
		s.prSimLabel.SetText(fmt.Sprintf("%.1f", s.simPr*100))
		s.prRangeLabel.SetText(fmt.Sprintf("%.1f - %.1f", s.expPrMin*100, s.expPrMax*100))
		prStatus, prColor := checkInRange(s.simPr, s.expPrMin, s.expPrMax)
		s.prStatusText.Text = prStatus
		s.prStatusText.Color = prColor
		s.prStatusText.Refresh()

		// Ec row
		s.ecSimLabel.SetText(fmt.Sprintf("%.2f", s.simEc/1e8))
		s.ecRangeLabel.SetText(fmt.Sprintf("%.1f - %.1f", s.expEcMin/1e8, s.expEcMax/1e8))
		ecStatus, ecColor := checkInRange(s.simEc, s.expEcMin, s.expEcMax)
		s.ecStatusText.Text = ecStatus
		s.ecStatusText.Color = ecColor
		s.ecStatusText.Refresh()

		// Squareness row
		s.sqSimLabel.SetText(fmt.Sprintf("%.2f", s.simSquareness))
		s.sqRangeLabel.SetText(fmt.Sprintf("%.2f - %.2f", s.expSquareMin, s.expSquareMax))
		sqStatus, sqColor := checkInRange(s.simSquareness, s.expSquareMin, s.expSquareMax)
		s.sqStatusText.Text = sqStatus
		s.sqStatusText.Color = sqColor
		s.sqStatusText.Refresh()

		// Source
		s.sourceLabel.SetText("Source: " + s.litSource)
	}
	s.BaseWidget.Refresh()
}

// MinSize returns the minimum size for the widget.
func (s *SimVsExpComparison) MinSize() fyne.Size {
	return fyne.NewSize(300, 250)
}

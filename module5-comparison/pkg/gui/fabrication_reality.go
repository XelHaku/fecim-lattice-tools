// Package gui provides Fyne-based GUI components for architecture comparison.
// This file implements H08: Fabrication Reality section showing realistic
// chip development timelines and costs.
package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FabricationReality shows realistic chip development expectations.
// Per Dr. Tour critique - provide honest timeline and cost expectations.
type FabricationReality struct {
	widget.BaseWidget
}

// NewFabricationReality creates a new fabrication reality widget.
func NewFabricationReality() *FabricationReality {
	f := &FabricationReality{}
	f.ExtendBaseWidget(f)
	return f
}

// CreateRenderer implements fyne.Widget.
func (f *FabricationReality) CreateRenderer() fyne.WidgetRenderer {
	// Header
	headerBg := canvas.NewRectangle(color.RGBA{60, 40, 30, 255})
	headerText := canvas.NewText("⚠ FABRICATION REALITY CHECK", color.RGBA{255, 180, 100, 255})
	headerText.TextSize = 14
	headerText.TextStyle = fyne.TextStyle{Bold: true}
	headerText.Alignment = fyne.TextAlignCenter

	header := container.NewStack(
		headerBg,
		container.NewCenter(headerText),
	)

	// Timeline section
	timelineTitle := widget.NewLabelWithStyle("Development Timeline", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	timelineItems := container.NewVBox(
		createTimelineItem("Design & Verification", "6-9 months", color.RGBA{100, 150, 200, 255}),
		createTimelineItem("First Tape-out", "3-4 months", color.RGBA{150, 150, 100, 255}),
		createTimelineItem("Fab Processing", "3-4 months", color.RGBA{200, 150, 100, 255}),
		createTimelineItem("Testing & Characterization", "3-6 months", color.RGBA{150, 100, 150, 255}),
		createTimelineItem("Revision (if needed)", "+6-12 months", color.RGBA{200, 100, 100, 255}),
	)

	totalTimeline := widget.NewLabelWithStyle(
		"TOTAL: 18-36 months to production silicon",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true, Italic: true},
	)

	// Cost section
	costTitle := widget.NewLabelWithStyle("Development Costs (Typical)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	costItems := container.NewVBox(
		createCostItem("First silicon (28nm)", "$1-2M"),
		createCostItem("First silicon (7nm)", "$10-30M"),
		createCostItem("EDA tool licenses", "$500K-2M/year"),
		createCostItem("Engineering team (10 engineers)", "$2-4M/year"),
		createCostItem("Testing equipment", "$500K-2M"),
	)

	totalCost := widget.NewLabelWithStyle(
		"TOTAL: $5-50M+ to first working chip",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true, Italic: true},
	)

	// Risk factors
	riskTitle := widget.NewLabelWithStyle("Risk Factors", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	riskItems := container.NewVBox(
		widget.NewLabel("• First silicon success rate: ~50% (industry average)"),
		widget.NewLabel("• Novel memory tech adds additional risk"),
		widget.NewLabel("• FeCIM: TRL 4 (lab demo) → TRL 7 (prototype) gap"),
		widget.NewLabel("• Yield challenges for analog multi-level cells"),
	)

	// Disclaimer
	disclaimer := canvas.NewText(
		"Note: These are industry-typical figures. Actual costs vary by process node and complexity.",
		color.RGBA{150, 150, 150, 255},
	)
	disclaimer.TextSize = 10
	disclaimer.TextStyle = fyne.TextStyle{Italic: true}
	disclaimer.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		timelineTitle,
		timelineItems,
		totalTimeline,
		widget.NewSeparator(),
		costTitle,
		costItems,
		totalCost,
		widget.NewSeparator(),
		riskTitle,
		riskItems,
		widget.NewSeparator(),
		container.NewCenter(disclaimer),
	)

	return widget.NewSimpleRenderer(content)
}

// createTimelineItem creates a timeline entry with color indicator.
func createTimelineItem(phase, duration string, c color.Color) fyne.CanvasObject {
	indicator := canvas.NewRectangle(c)
	indicator.SetMinSize(fyne.NewSize(10, 10))

	phaseLabel := widget.NewLabel(phase)
	durationLabel := widget.NewLabel(duration)
	durationLabel.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewHBox(
		indicator,
		phaseLabel,
		widget.NewLabel("→"),
		durationLabel,
	)
}

// createCostItem creates a cost entry.
func createCostItem(item, cost string) fyne.CanvasObject {
	itemLabel := widget.NewLabel(item)
	costLabel := widget.NewLabel(cost)
	costLabel.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewHBox(
		widget.NewLabel("•"),
		itemLabel,
		widget.NewLabel(":"),
		costLabel,
	)
}

// MinSize returns the minimum size for the widget.
func (f *FabricationReality) MinSize() fyne.Size {
	return fyne.NewSize(350, 450)
}

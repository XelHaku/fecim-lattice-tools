//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/config/physics"
)

// MaterialCard displays a compact summary of a material for list/grid display.
type MaterialCard struct {
	widget.BaseWidget

	material   *physics.Material
	materialID string
	selected   bool
	onTapped   func(materialID string)

	// UI components (created in CreateRenderer)
	background  *canvas.Rectangle
	nameLabel   *widget.Label
	statesLabel *widget.Label
	prLabel     *widget.Label
	ecLabel     *widget.Label
	refLabel    *widget.Label
}

// NewMaterialCard creates a new material card widget.
func NewMaterialCard(materialID string, material *physics.Material, onTapped func(string)) *MaterialCard {
	mc := &MaterialCard{
		material:   material,
		materialID: materialID,
		onTapped:   onTapped,
		selected:   false,
	}
	mc.ExtendBaseWidget(mc)
	return mc
}

// Tapped handles tap events on the card.
func (mc *MaterialCard) Tapped(_ *fyne.PointEvent) {
	if mc.onTapped != nil {
		mc.onTapped(mc.materialID)
	}
}

// TappedSecondary handles secondary tap (right-click).
func (mc *MaterialCard) TappedSecondary(_ *fyne.PointEvent) {}

// SetSelected updates the selection state.
func (mc *MaterialCard) SetSelected(selected bool) {
	mc.selected = selected
	mc.Refresh()
}

// IsSelected returns the selection state.
func (mc *MaterialCard) IsSelected() bool {
	return mc.selected
}

// GetMaterialID returns the material ID.
func (mc *MaterialCard) GetMaterialID() string {
	return mc.materialID
}

// MinSize returns the minimum size for the card.
func (mc *MaterialCard) MinSize() fyne.Size {
	return fyne.NewSize(280, 90)
}

// CreateRenderer creates the widget renderer.
func (mc *MaterialCard) CreateRenderer() fyne.WidgetRenderer {
	// Background with selection highlight
	mc.background = canvas.NewRectangle(color.RGBA{30, 35, 45, 255})
	mc.background.StrokeColor = color.RGBA{60, 70, 90, 255}
	mc.background.StrokeWidth = 1
	mc.background.CornerRadius = 6

	// Material name (bold, larger)
	mc.nameLabel = widget.NewLabelWithStyle(
		mc.material.Name,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	// Analog states with bits calculation
	statesText := "N/A"
	if mc.material.AnalogStates > 0 {
		bits := math.Log2(float64(mc.material.AnalogStates))
		statesText = fmt.Sprintf("%d states (%.1f bits/cell)", mc.material.AnalogStates, bits)
	}
	mc.statesLabel = widget.NewLabel(statesText)
	mc.statesLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Pr and Ec on same line
	prText := FormatPolarization(mc.material.PrCM2)
	ecText := FormatField(mc.material.EcVM)
	mc.prLabel = widget.NewLabel(fmt.Sprintf("Pr: %s", prText))
	mc.ecLabel = widget.NewLabel(fmt.Sprintf("Ec: %s", ecText))

	// Reference (truncated)
	refText := TruncateString(mc.material.Reference, 50)
	mc.refLabel = widget.NewLabel(refText)
	mc.refLabel.Wrapping = fyne.TextWrapOff
	mc.refLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Properties row
	propsRow := container.NewHBox(
		mc.prLabel,
		widget.NewLabel(" | "),
		mc.ecLabel,
	)

	// Content layout
	content := container.NewVBox(
		mc.nameLabel,
		mc.statesLabel,
		propsRow,
		mc.refLabel,
	)

	paddedContent := container.NewPadded(content)

	return &materialCardRenderer{
		card:       mc,
		background: mc.background,
		content:    paddedContent,
		objects:    []fyne.CanvasObject{mc.background, paddedContent},
	}
}

// materialCardRenderer handles the rendering of MaterialCard.
type materialCardRenderer struct {
	card       *MaterialCard
	background *canvas.Rectangle
	content    *fyne.Container
	objects    []fyne.CanvasObject
}

func (r *materialCardRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.content.Resize(size)
}

func (r *materialCardRenderer) MinSize() fyne.Size {
	return r.card.MinSize()
}

func (r *materialCardRenderer) Refresh() {
	// Update selection styling
	if r.card.selected {
		r.background.FillColor = color.RGBA{40, 60, 100, 255}
		r.background.StrokeColor = color.RGBA{0, 150, 255, 255}
		r.background.StrokeWidth = 2
	} else {
		r.background.FillColor = color.RGBA{30, 35, 45, 255}
		r.background.StrokeColor = color.RGBA{60, 70, 90, 255}
		r.background.StrokeWidth = 1
	}
	r.background.Refresh()

	// Update labels
	r.card.nameLabel.SetText(r.card.material.Name)

	if r.card.material.AnalogStates > 0 {
		bits := math.Log2(float64(r.card.material.AnalogStates))
		r.card.statesLabel.SetText(fmt.Sprintf("%d states (%.1f bits/cell)", r.card.material.AnalogStates, bits))
	} else {
		r.card.statesLabel.SetText("N/A")
	}

	r.card.prLabel.SetText(fmt.Sprintf("Pr: %s", FormatPolarization(r.card.material.PrCM2)))
	r.card.ecLabel.SetText(fmt.Sprintf("Ec: %s", FormatField(r.card.material.EcVM)))
	r.card.refLabel.SetText(TruncateString(r.card.material.Reference, 50))
}

func (r *materialCardRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *materialCardRenderer) Destroy() {}

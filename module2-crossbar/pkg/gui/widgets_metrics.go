// Package gui provides custom widgets for crossbar visualization.
package gui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// MetricsPanel displays accuracy and energy metrics.
type MetricsPanel struct {
	widget.BaseWidget

	// Accuracy metrics
	idealAccuracy  float64
	actualAccuracy float64
	accuracyDelta  float64

	// Energy metrics
	fecimEnergy   float64
	gpuEnergy     float64
	efficiency    float64

	// Performance
	macOps  int
	latency float64

	// Labels
	labels map[string]*widget.Label
}

// NewMetricsPanel creates a new metrics panel.
func NewMetricsPanel() *MetricsPanel {
	m := &MetricsPanel{
		labels: make(map[string]*widget.Label),
	}
	m.ExtendBaseWidget(m)
	return m
}

// UpdateMetrics updates all metrics.
func (m *MetricsPanel) UpdateMetrics(idealAcc, actualAcc, fecimE, gpuE float64, macs int, lat float64) {
	m.idealAccuracy = idealAcc
	m.actualAccuracy = actualAcc
	m.accuracyDelta = idealAcc - actualAcc
	m.fecimEnergy = fecimE
	m.gpuEnergy = gpuE
	m.efficiency = gpuE / fecimE
	m.macOps = macs
	m.latency = lat
	if sharedwidgets.IsStartupStabilizing() {
		return
	}
	fyne.Do(func() {
		m.Refresh()
	})
}

// CreateRenderer implements fyne.Widget.
func (m *MetricsPanel) CreateRenderer() fyne.WidgetRenderer {
	// Create labels
	headerLabel := widget.NewLabelWithStyle("Live Metrics", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	accHeader := widget.NewLabelWithStyle("Accuracy", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	idealLabel := widget.NewLabel("Ideal: --")
	actualLabel := widget.NewLabel("Actual: --")
	deltaLabel := widget.NewLabel("Δ: --")

	energyHeader := widget.NewLabelWithStyle("Energy", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	fecimELabel := widget.NewLabel("FeCIM: --")
	gpuELabel := widget.NewLabel("GPU: --")
	effLabel := widget.NewLabel("Efficiency: --")

	perfHeader := widget.NewLabelWithStyle("Performance", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	macLabel := widget.NewLabel("MACs: --")
	latLabel := widget.NewLabel("Latency: --")

	m.labels["ideal"] = idealLabel
	m.labels["actual"] = actualLabel
	m.labels["delta"] = deltaLabel
	m.labels["fecim"] = fecimELabel
	m.labels["gpu"] = gpuELabel
	m.labels["eff"] = effLabel
	m.labels["mac"] = macLabel
	m.labels["lat"] = latLabel

	content := container.NewVBox(
		headerLabel,
		widget.NewSeparator(),
		accHeader,
		idealLabel,
		actualLabel,
		deltaLabel,
		widget.NewSeparator(),
		energyHeader,
		fecimELabel,
		gpuELabel,
		effLabel,
		widget.NewSeparator(),
		perfHeader,
		macLabel,
		latLabel,
	)

	return widget.NewSimpleRenderer(content)
}

// Refresh updates the display.
func (m *MetricsPanel) Refresh() {
	if len(m.labels) == 0 {
		m.BaseWidget.Refresh()
		return
	}

	// Update accuracy labels
	if m.idealAccuracy > 0 {
		m.labels["ideal"].SetText(fmt.Sprintf("Ideal: %.1f%%", m.idealAccuracy))
		m.labels["actual"].SetText(fmt.Sprintf("Actual: %.1f%%", m.actualAccuracy))
		deltaStr := fmt.Sprintf("Δ: %.1f%%", m.accuracyDelta)
		if m.accuracyDelta > 0 {
			deltaStr = fmt.Sprintf("Δ: -%.1f%%", math.Abs(m.accuracyDelta))
		}
		m.labels["delta"].SetText(deltaStr)
	}

	// Update energy labels
	if m.fecimEnergy > 0 {
		m.labels["fecim"].SetText(fmt.Sprintf("FeCIM: %.2f pJ", m.fecimEnergy))
		m.labels["gpu"].SetText(fmt.Sprintf("GPU: %.0f pJ", m.gpuEnergy))
		m.labels["eff"].SetText(fmt.Sprintf("%.0f× better", m.efficiency))
	}

	// Update performance labels
	if m.macOps > 0 {
		m.labels["mac"].SetText(fmt.Sprintf("%d MACs", m.macOps))
		m.labels["lat"].SetText(fmt.Sprintf("%.0f ns", m.latency))
	}

	m.BaseWidget.Refresh()
}

// ComparisonBadge displays a visual comparison.
type ComparisonBadge struct {
	widget.BaseWidget

	metric      string
	fecimValue  string
	gpuValue    string
	improvement string

	raster *fyne.CanvasObject
}

// NewComparisonBadge creates a new comparison badge.
func NewComparisonBadge(metric string) *ComparisonBadge {
	b := &ComparisonBadge{
		metric:      metric,
		fecimValue:  "--",
		gpuValue:    "--",
		improvement: "--",
	}
	b.ExtendBaseWidget(b)
	return b
}

// UpdateValues updates the comparison values.
func (b *ComparisonBadge) UpdateValues(fecimVal, gpuVal string, improvement string) {
	b.fecimValue = fecimVal
	b.gpuValue = gpuVal
	b.improvement = improvement
	if sharedwidgets.IsStartupStabilizing() {
		return
	}
	fyne.Do(func() {
		b.Refresh()
	})
}

// CreateRenderer implements fyne.Widget.
func (b *ComparisonBadge) CreateRenderer() fyne.WidgetRenderer {
	// Simple implementation - can be enhanced with canvas.Raster if needed
	content := container.NewVBox(
		widget.NewLabel(b.metric),
		widget.NewLabel(fmt.Sprintf("FeCIM: %s", b.fecimValue)),
		widget.NewLabel(fmt.Sprintf("GPU: %s", b.gpuValue)),
		widget.NewLabel(fmt.Sprintf("%s improvement", b.improvement)),
	)
	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (b *ComparisonBadge) MinSize() fyne.Size {
	return fyne.NewSize(200, 100)
}

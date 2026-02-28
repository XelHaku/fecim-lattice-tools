// Package tabs - 3D multi-layer stack visualization tab.
//
// L10: Renders stacked crossbar layers with isometric projection using a
// pure-Go software renderer. Supports up to 512 layers with automatic
// subsampling for performance.
package tabs

import (
	"fmt"
	"math"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/render3d"
)

// Stack3DTab provides the 3D multi-layer stack visualization.
type Stack3DTab struct {
	stackWidget *render3d.StackWidget
	statusLabel *widget.Label
	infoLabel   *widget.Label

	numLayers     int
	cellsPerLayer int // cells per side (square layers)
}

// NewStack3DTab creates a 3D layer stack visualization tab.
func NewStack3DTab() *Stack3DTab {
	tab := &Stack3DTab{
		stackWidget:   render3d.NewStackWidget(),
		statusLabel:   widget.NewLabel("3D Stack: drag to rotate, scroll to zoom, click to select layer"),
		infoLabel:     widget.NewLabel("Layers: 4 | Grid: 8x8 | Colormap: viridis"),
		numLayers:     4,
		cellsPerLayer: 8,
	}

	// Wire up callbacks
	tab.stackWidget.OnLayerSelected = func(layer int) {
		fyne.Do(func() {
			tab.statusLabel.SetText(fmt.Sprintf("Selected: Layer %d", layer))
		})
	}
	tab.stackWidget.OnRotated = func(az, el float64) {
		fyne.Do(func() {
			tab.statusLabel.SetText(fmt.Sprintf("Camera: azimuth=%.0f° elevation=%.0f°",
				az*180/math.Pi, el*180/math.Pi))
		})
	}

	// Generate initial layers
	tab.regenerateLayers()

	return tab
}

// Content returns the tab content.
func (t *Stack3DTab) Content() fyne.CanvasObject {
	// Layer count slider (1-512)
	layerCountLabel := widget.NewLabel("Layers: 4")
	layerCountSlider := widget.NewSlider(1, 512)
	layerCountSlider.Value = 4
	layerCountSlider.Step = 1
	layerCountSlider.OnChanged = func(v float64) {
		n := int(v)
		t.numLayers = n
		layerCountLabel.SetText(fmt.Sprintf("Layers: %d", n))
		t.regenerateLayers()
		t.updateInfo()
	}

	// Grid size slider (2-64 per side)
	gridSizeLabel := widget.NewLabel("Grid: 8x8")
	gridSizeSlider := widget.NewSlider(2, 64)
	gridSizeSlider.Value = 8
	gridSizeSlider.Step = 2
	gridSizeSlider.OnChanged = func(v float64) {
		n := int(v)
		t.cellsPerLayer = n
		gridSizeLabel.SetText(fmt.Sprintf("Grid: %dx%d", n, n))
		t.regenerateLayers()
		t.updateInfo()
	}

	// Camera controls
	azimuthLabel := widget.NewLabel("Azimuth: 30°")
	azimuthSlider := widget.NewSlider(0, 360)
	azimuthSlider.Value = 30
	azimuthSlider.Step = 5
	azimuthSlider.OnChanged = func(v float64) {
		azimuthLabel.SetText(fmt.Sprintf("Azimuth: %.0f°", v))
		t.stackWidget.SetCamera(
			v*math.Pi/180,
			t.stackWidget.Elevation(),
			t.stackWidget.ZoomLevel(),
		)
	}

	elevationLabel := widget.NewLabel("Elevation: 30°")
	elevationSlider := widget.NewSlider(5, 85)
	elevationSlider.Value = 30
	elevationSlider.Step = 5
	elevationSlider.OnChanged = func(v float64) {
		elevationLabel.SetText(fmt.Sprintf("Elevation: %.0f°", v))
		t.stackWidget.SetCamera(
			t.stackWidget.AzimuthAngle(),
			v*math.Pi/180,
			t.stackWidget.ZoomLevel(),
		)
	}

	// Layer gap slider
	gapLabel := widget.NewLabel("Gap: 15%")
	gapSlider := widget.NewSlider(0, 100)
	gapSlider.Value = 15
	gapSlider.Step = 5
	gapSlider.OnChanged = func(v float64) {
		gapLabel.SetText(fmt.Sprintf("Gap: %.0f%%", v))
		t.stackWidget.SetLayerGap(v / 100.0)
	}

	// Colormap selector
	colormapSelect := widget.NewSelect([]string{"viridis", "plasma", "coolwarm"}, func(s string) {
		t.stackWidget.SetColormap(s)
		t.updateInfo()
	})
	colormapSelect.SetSelected("viridis")

	// Show wires checkbox
	wiresCheck := widget.NewCheck("Show Wires", func(b bool) {
		t.stackWidget.SetShowWires(b)
	})

	// Randomize button
	randomizeBtn := widget.NewButton("Randomize Data", func() {
		t.regenerateLayers()
	})
	randomizeBtn.Importance = widget.MediumImportance

	// Max visible layers (performance cap)
	maxVisLabel := widget.NewLabel("Max Visible: 32")
	maxVisSlider := widget.NewSlider(4, 128)
	maxVisSlider.Value = 32
	maxVisSlider.Step = 4
	maxVisSlider.OnChanged = func(v float64) {
		n := int(v)
		maxVisLabel.SetText(fmt.Sprintf("Max Visible: %d", n))
		t.stackWidget.SetMaxVisibleLayers(n)
	}
	t.stackWidget.SetMaxVisibleLayers(32)

	// Controls panel
	controls := container.NewVBox(
		widget.NewLabelWithStyle("3D Layer Stack", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		layerCountLabel,
		layerCountSlider,
		gridSizeLabel,
		gridSizeSlider,
		widget.NewSeparator(),
		azimuthLabel,
		azimuthSlider,
		elevationLabel,
		elevationSlider,
		gapLabel,
		gapSlider,
		widget.NewSeparator(),
		container.NewBorder(nil, nil, widget.NewLabel("Colormap:"), nil, colormapSelect),
		wiresCheck,
		maxVisLabel,
		maxVisSlider,
		widget.NewSeparator(),
		randomizeBtn,
	)
	controlsScroll := container.NewVScroll(controls)
	controlsScroll.SetMinSize(fyne.NewSize(220, 200))

	// Main layout: 3D view in center, controls on right
	content := container.NewBorder(
		nil,
		container.NewVBox(t.infoLabel, t.statusLabel),
		nil,
		controlsScroll,
		t.stackWidget,
	)

	return content
}

// regenerateLayers creates randomized layer data and updates the widget.
func (t *Stack3DTab) regenerateLayers() {
	layers := make([]render3d.LayerData, t.numLayers)
	for l := 0; l < t.numLayers; l++ {
		data := make([]float64, t.cellsPerLayer*t.cellsPerLayer)
		for i := range data {
			// Gradient with noise: base value depends on layer, noise makes it interesting
			base := float64(l) / math.Max(float64(t.numLayers-1), 1)
			data[i] = clamp01(base + (rand.Float64()-0.5)*0.3)
		}
		layers[l] = render3d.LayerData{
			Values: data,
			Rows:   t.cellsPerLayer,
			Cols:   t.cellsPerLayer,
			Label:  fmt.Sprintf("L%d", l),
		}
	}
	t.stackWidget.SetLayers(layers)
}

// updateInfo refreshes the info label.
func (t *Stack3DTab) updateInfo() {
	fyne.Do(func() {
		t.infoLabel.SetText(fmt.Sprintf("Layers: %d | Grid: %dx%d | Cells: %d",
			t.numLayers, t.cellsPerLayer, t.cellsPerLayer,
			t.numLayers*t.cellsPerLayer*t.cellsPerLayer))
	})
}

// clamp01 clamps v to [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

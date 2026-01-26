// Package gui provides Fyne-based GUI components for MNIST visualization.
// dualmode_weights.go provides weight visualization functionality.
package gui

import (
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// createWeightZone creates the weight visualization zone (Zone 4).
func (app *DualModeApp) createWeightZone() fyne.CanvasObject {
	fmt.Println("[MNIST] createWeightZone: start")
	label := widget.NewLabel("Weights")
	label.TextStyle = fyne.TextStyle{Bold: true}

	// Layer selector as horizontal radio
	// NOTE: Don't set callback initially - SetSelected triggers callback which uses fyne.Do() and deadlocks during startup
	fmt.Println("[MNIST] createWeightZone: creating radio group...")
	layerSelect := widget.NewRadioGroup(
		[]string{"Layer1 (784x128)", "Layer2 (128x10)"},
		nil, // No callback initially to avoid fyne.Do() deadlock during startup
	)
	layerSelect.Horizontal = true
	fmt.Println("[MNIST] createWeightZone: setting selected...")
	layerSelect.SetSelected("Layer1 (784x128)")
	// Now set the callback after initial selection
	layerSelect.OnChanged = func(s string) {
		if s == "Layer1 (784x128)" {
			app.weightLayer = 0
		} else {
			app.weightLayer = 1
		}
		app.updateWeightHeatmap()
		app.updateWeightComparison()
	}
	fmt.Println("[MNIST] createWeightZone: radio done")

	// Info labels (combined into one)
	fmt.Println("[MNIST] createWeightZone: creating labels...")
	app.weightDimLabel = widget.NewLabel("")
	app.weightRangeLabel = widget.NewLabel("")
	app.weightLevelsLabel = widget.NewLabel("")

	// Create heatmap raster (quantized weights)
	fmt.Println("[MNIST] createWeightZone: creating heatmap...")
	app.hoverableWeightHeatmap = NewHoverableHeatmap(app)
	app.hoverableWeightHeatmap.SetMinSize(fyne.NewSize(600, 400))
	app.weightHeatmap = app.hoverableWeightHeatmap.raster // Keep reference for updateWeightHeatmap

	// Create color legend for weight heatmap (blue-white-red)
	fmt.Println("[MNIST] createWeightZone: creating color legend...")
	weightLegend := sharedwidgets.NewColorLegend(-1.0, 1.0, "", true, sharedwidgets.BlueWhiteRedColor)

	// Create FP vs Quantized comparison widget
	fmt.Println("[MNIST] createWeightZone: creating comparison widget...")
	app.weightComparisonWidget = NewWeightComparisonWidget()

	// Create dual heatmap (side-by-side FP vs Quantized)
	fmt.Println("[MNIST] createWeightZone: creating dual heatmap...")
	app.dualWeightHeatmap = NewDualWeightHeatmap()

	fmt.Println("[MNIST] createWeightZone: creating zoom button...")
	zoomBtn := widget.NewButtonWithIcon("", theme.ZoomFitIcon(), func() {
		app.showZoomedHeatmap()
	})
	fmt.Println("[MNIST] createWeightZone: done with widgets")

	// Header with controls
	fmt.Println("[MNIST] createWeightZone: creating header container...")
	header := container.NewVBox(
		container.NewHBox(label, layout.NewSpacer(), layerSelect, zoomBtn),
		container.NewHBox(
			app.weightDimLabel,
			widget.NewSeparator(),
			app.weightRangeLabel,
			widget.NewSeparator(),
			app.weightLevelsLabel,
		),
	)

	// P1 Enhancement Widgets (moved from controls zone for better space utilization)
	fmt.Println("[MNIST] createWeightZone: creating P1 widgets...")
	app.quantizationWidget = NewQuantizationWidget()
	app.energyWidget = NewEnergyWidget(MNISTInputSize, MNISTHiddenSize, MNISTOutputSize)

	// Create tabbed view for different weight visualizations
	fmt.Println("[MNIST] createWeightZone: creating tabs...")
	// Add legend to the left of the heatmap
	quantizedTab := container.NewBorder(nil, nil, weightLegend, nil, app.hoverableWeightHeatmap)
	comparisonTab := container.NewMax(app.weightComparisonWidget)
	sideBySideTab := container.NewMax(app.dualWeightHeatmap)
	quantizationTab := container.NewMax(app.quantizationWidget)
	energyTab := container.NewMax(app.energyWidget)

	weightTabs := container.NewAppTabs(
		container.NewTabItem("Quantized", quantizedTab),
		container.NewTabItem("FP vs Quant", comparisonTab),
		container.NewTabItem("Side-by-Side", sideBySideTab),
		container.NewTabItem("Quantization", quantizationTab),
		container.NewTabItem("Energy", energyTab),
	)

	fmt.Println("[MNIST] createWeightZone: returning...")
	return container.NewBorder(
		header,
		nil, nil, nil,
		weightTabs,
	)
}

// updateWeightComparison updates the FP vs Quantized comparison widgets.
func (app *DualModeApp) updateWeightComparison() {
	fpW1, fpW2, _, _ := app.network.GetFPWeights()
	quantW1, quantW2, _, _ := app.network.GetQuantWeights()

	// Update comparison widget based on selected layer
	if app.weightComparisonWidget != nil {
		if app.weightLayer == 0 {
			app.weightComparisonWidget.SetWeights(fpW1, quantW1)
		} else {
			app.weightComparisonWidget.SetWeights(fpW2, quantW2)
		}
	}

	// Update dual heatmap
	if app.dualWeightHeatmap != nil {
		app.dualWeightHeatmap.SetWeights(fpW1, fpW2, quantW1, quantW2)
		app.dualWeightHeatmap.SetLayer(app.weightLayer)
	}
}

// drawWeightHeatmap draws the weight heatmap.
func (app *DualModeApp) drawWeightHeatmap(w, h int) image.Image {
	_, w2, _, _ := app.network.GetQuantWeights()

	var weights [][]float64
	if app.weightLayer == 0 {
		weights, _, _, _ = app.network.GetQuantWeights()
	} else {
		weights = w2
	}

	if len(weights) == 0 || len(weights[0]) == 0 {
		return image.NewRGBA(image.Rect(0, 0, w, h))
	}

	rows := len(weights)
	cols := len(weights[0])

	// Find range
	wMin, wMax := weights[0][0], weights[0][0]
	for i := range weights {
		for j := range weights[i] {
			if weights[i][j] < wMin {
				wMin = weights[i][j]
			}
			if weights[i][j] > wMax {
				wMax = weights[i][j]
			}
		}
	}

	// Create image with proper scaling
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	scaleX := float64(cols) / float64(w)
	scaleY := float64(rows) / float64(h)

	wRange := wMax - wMin
	if wRange == 0 {
		wRange = 1
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Map to weight matrix
			row := int(float64(y) * scaleY)
			col := int(float64(x) * scaleX)

			if row >= rows {
				row = rows - 1
			}
			if col >= cols {
				col = cols - 1
			}

			val := weights[row][col]
			normalized := (val - wMin) / wRange

			// Blue-White-Red colormap
			c := weightToColor(normalized)
			img.Set(x, y, c)
		}
	}

	return img
}

// weightToColor converts normalized value [0,1] to color.
func weightToColor(normalized float64) color.Color {
	if normalized < 0.5 {
		// Blue to white
		t := normalized * 2
		r := uint8(t * 255)
		g := uint8(t * 255)
		b := uint8(255)
		return color.RGBA{r, g, b, 255}
	}
	// White to red
	t := (normalized - 0.5) * 2
	r := uint8(255)
	g := uint8((1 - t) * 255)
	b := uint8((1 - t) * 255)
	return color.RGBA{r, g, b, 255}
}

// HoverableHeatmap wraps a weight heatmap with hover tooltip functionality.
type HoverableHeatmap struct {
	widget.BaseWidget
	raster  *canvas.Raster
	tooltip *widget.PopUp
	app     *DualModeApp
}

// NewHoverableHeatmap creates a new hoverable heatmap widget.
func NewHoverableHeatmap(app *DualModeApp) *HoverableHeatmap {
	h := &HoverableHeatmap{
		raster: canvas.NewRaster(app.drawWeightHeatmap),
		app:    app,
	}
	h.ExtendBaseWidget(h)
	return h
}

// CreateRenderer creates the renderer for the hoverable heatmap.
func (h *HoverableHeatmap) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.raster)
}

// SetMinSize sets the minimum size of the heatmap.
func (h *HoverableHeatmap) SetMinSize(size fyne.Size) {
	h.raster.SetMinSize(size)
}

// Refresh updates the heatmap display.
func (h *HoverableHeatmap) Refresh() {
	h.raster.Refresh()
	h.BaseWidget.Refresh()
}

// MouseIn is called when the mouse enters the widget.
func (h *HoverableHeatmap) MouseIn(e *desktop.MouseEvent) {
	// Nothing to do on mouse in
}

// MouseMoved is called when the mouse moves over the widget.
func (h *HoverableHeatmap) MouseMoved(e *desktop.MouseEvent) {
	// Get the current weight matrix
	_, w2, _, _ := h.app.network.GetQuantWeights()
	var weights [][]float64
	if h.app.weightLayer == 0 {
		weights, _, _, _ = h.app.network.GetQuantWeights()
	} else {
		weights = w2
	}

	if len(weights) == 0 || len(weights[0]) == 0 {
		return
	}

	rows := len(weights)
	cols := len(weights[0])

	// Get widget size
	size := h.Size()
	if size.Width == 0 || size.Height == 0 {
		return
	}

	// Calculate which cell was hovered
	row := int(e.Position.Y / size.Height * float32(rows))
	col := int(e.Position.X / size.Width * float32(cols))

	// Bounds check
	if row < 0 || row >= rows || col < 0 || col >= cols {
		if h.tooltip != nil {
			h.tooltip.Hide()
			h.tooltip = nil
		}
		return
	}

	// Get weight value
	value := weights[row][col]

	// Create or update tooltip
	tooltipText := fmt.Sprintf("Weight[%d,%d] = %.4f", row, col, value)
	label := widget.NewLabel(tooltipText)

	if h.tooltip != nil {
		h.tooltip.Hide()
	}

	// Create new tooltip at cursor position
	h.tooltip = widget.NewPopUp(label, h.app.window.Canvas())

	// Position tooltip near the cursor
	tooltipPos := fyne.NewPos(
		e.AbsolutePosition.X+10,
		e.AbsolutePosition.Y+10,
	)
	h.tooltip.ShowAtPosition(tooltipPos)
}

// MouseOut is called when the mouse leaves the widget.
func (h *HoverableHeatmap) MouseOut() {
	if h.tooltip != nil {
		h.tooltip.Hide()
		h.tooltip = nil
	}
}

// updateWeightHeatmap refreshes the weight visualization.
func (app *DualModeApp) updateWeightHeatmap() {
	if !app.initialized {
		return
	}
	if app.weightHeatmap != nil {
		fyne.Do(func() {
			app.weightHeatmap.Refresh()
		})
	}

	// Update info labels
	var weights [][]float64
	if app.weightLayer == 0 {
		weights, _, _, _ = app.network.GetQuantWeights()
		if len(weights) > 0 && len(weights[0]) > 0 {
			msg := fmt.Sprintf("Dimensions: %d rows x %d cols", len(weights), len(weights[0]))
			fyne.Do(func() {
				app.weightDimLabel.SetText(msg)
			})
		}
	} else {
		_, weights, _, _ = app.network.GetQuantWeights()
		if len(weights) > 0 && len(weights[0]) > 0 {
			msg := fmt.Sprintf("Dimensions: %d rows x %d cols", len(weights), len(weights[0]))
			fyne.Do(func() {
				app.weightDimLabel.SetText(msg)
			})
		}
	}

	if len(weights) > 0 && len(weights[0]) > 0 {
		// Find range
		wMin, wMax := weights[0][0], weights[0][0]
		distinctMap := make(map[float64]bool)
		for i := range weights {
			for j := range weights[i] {
				if weights[i][j] < wMin {
					wMin = weights[i][j]
				}
				if weights[i][j] > wMax {
					wMax = weights[i][j]
				}
				distinctMap[weights[i][j]] = true
			}
		}
		rangeMsg := fmt.Sprintf("Range: [%.3f, %.3f]", wMin, wMax)
		levelsMsg := fmt.Sprintf("Levels: %d/30", len(distinctMap))
		fyne.Do(func() {
			app.weightRangeLabel.SetText(rangeMsg)
			app.weightLevelsLabel.SetText(levelsMsg)
		})
	}

	// Also update FP vs Quantized comparison
	app.updateWeightComparison()
}

// showZoomedHeatmap opens a new window with a larger view of the weight heatmap.
func (app *DualModeApp) showZoomedHeatmap() {
	w := app.fyneApp.NewWindow("Weight Matrix Detail")
	w.Resize(fyne.NewSize(800, 600))

	// Create a new hoverable heatmap for the zoomed view
	zoomedApp := &DualModeApp{
		fyneApp:     app.fyneApp,
		window:      w,
		network:     app.network,
		weightLayer: app.weightLayer,
	}
	heatmap := NewHoverableHeatmap(zoomedApp)

	// Add a close button
	closeBtn := widget.NewButton("Close", func() {
		w.Close()
	})

	content := container.NewBorder(
		nil,
		container.NewHBox(layout.NewSpacer(), closeBtn),
		nil, nil,
		container.NewMax(heatmap),
	)

	w.SetContent(content)
	w.Show()
}

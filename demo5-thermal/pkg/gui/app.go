// Package gui provides Fyne-based GUI for thermal analysis visualization.
package gui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/demo5-thermal/pkg/thermal"
)

// FeCIM theme colors
var (
	colorBackground = color.RGBA{0, 50, 100, 255}  // FeCIM blue #003264
	colorPrimary    = color.RGBA{0, 212, 255, 255} // Cyan
)

// feCIMTheme implements fyne.Theme for consistent FeCIM branding
type feCIMTheme struct{}

func (t *feCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBackground
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255}
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *feCIMTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *feCIMTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *feCIMTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// ThermalApp is the main application for the thermal analysis demo.
type ThermalApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Simulation
	multiSim  *thermal.MultiLayerSim
	singleSim *thermal.ThermalSim
	running   bool
	stopChan  chan bool

	// UI Components
	heatmapView      *fyne.Container
	layerTabs        *container.AppTabs
	statsLabel       *widget.Label
	statusLabel      *widget.Label
	comparisonPanel  *fyne.Container
	educationalPanel *widget.Label
	operationLog     *widget.List
	logEntries       []string

	// Settings
	selectedLayer  int
	powerLevel     float64
	simulationMode string
}

// NewThermalApp creates a new thermal visualization app.
func NewThermalApp() *ThermalApp {
	return &ThermalApp{
		selectedLayer:  1, // Middle layer (crossbar)
		powerLevel:     1.0,
		simulationMode: "FeCIM (Low Power)",
		logEntries:     make([]string, 0, 20),
	}
}

// Run starts the application.
func (ta *ThermalApp) Run() {
	ta.fyneApp = app.NewWithID("com.fecim.thermal-demo")
	ta.fyneApp.Settings().SetTheme(&feCIMTheme{})

	ta.window = ta.fyneApp.NewWindow("FeCIM Demo 5: Thermal Analysis")
	ta.window.Resize(fyne.NewSize(1300, 850))

	// Initialize simulation
	ta.initSimulation()

	content := ta.createMainLayout()
	ta.window.SetContent(content)

	// Initial updates
	ta.updateHeatmap()
	ta.updateStats()

	ta.window.ShowAndRun()
}

// initSimulation sets up the thermal simulation.
func (ta *ThermalApp) initSimulation() {
	ta.multiSim = thermal.DefaultMultiLayerSim()
	ta.singleSim = ta.multiSim.Layers[ta.selectedLayer]

	// Set initial power distribution (FeCIM-like)
	ta.setPowerDistribution("FeCIM (Low Power)")
}

// createMainLayout builds the main application layout.
func (ta *ThermalApp) createMainLayout() fyne.CanvasObject {
	// Header
	header := ta.createHeader()

	// Left panel: Controls
	leftPanel := ta.createControlPanel()

	// Center: Heatmap visualization
	centerPanel := ta.createCenterPanel()

	// Right panel: Stats and comparison
	rightPanel := ta.createRightPanel()

	// Status bar
	ta.statusLabel = widget.NewLabel("Ready - Run simulation to see thermal evolution")

	// Main content
	mainSplit := container.NewHSplit(
		container.NewHSplit(
			container.NewPadded(leftPanel),
			centerPanel,
		),
		container.NewPadded(rightPanel),
	)
	mainSplit.SetOffset(0.7)

	return container.NewBorder(
		header,
		ta.statusLabel,
		nil,
		nil,
		mainSplit,
	)
}

func (ta *ThermalApp) createHeader() fyne.CanvasObject {
	title := canvas.NewText("FeCIM Demo 5: Thermal Analysis", color.White)
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}

	quote := widget.NewLabel("\"1000x lower energy means 1000x less heat\" — Dr. external research group")
	quote.TextStyle = fyne.TextStyle{Italic: true}

	subtitle := widget.NewLabel("Compare FeCIM thermal profile vs GPU/DRAM - See the cooling advantage")
	subtitle.Alignment = fyne.TextAlignCenter

	return container.NewVBox(
		container.NewCenter(title),
		container.NewCenter(quote),
		container.NewCenter(subtitle),
		widget.NewSeparator(),
	)
}

func (ta *ThermalApp) createControlPanel() fyne.CanvasObject {
	// Simulation mode selector
	modeSelect := widget.NewSelect(
		[]string{"FeCIM (Low Power)", "GPU (High Power)", "DRAM (Medium Power)", "Custom"},
		func(s string) {
			ta.simulationMode = s
			ta.setPowerDistribution(s)
			ta.addLog(fmt.Sprintf("Mode: %s", s))
			ta.updateHeatmap()
			ta.updateStats()
		},
	)
	modeSelect.SetSelected("FeCIM (Low Power)")

	// Layer selector
	layerSelect := widget.NewSelect(
		[]string{"Layer 1 (Substrate)", "Layer 2 (Crossbar)", "Layer 3 (BEOL)"},
		func(s string) {
			switch s {
			case "Layer 1 (Substrate)":
				ta.selectedLayer = 0
			case "Layer 2 (Crossbar)":
				ta.selectedLayer = 1
			case "Layer 3 (BEOL)":
				ta.selectedLayer = 2
			}
			ta.singleSim = ta.multiSim.Layers[ta.selectedLayer]
			ta.addLog(fmt.Sprintf("Layer: %d selected", ta.selectedLayer+1))
			ta.updateHeatmap()
			ta.updateStats()
		},
	)
	layerSelect.SetSelected("Layer 2 (Crossbar)")

	// Power slider
	powerLabel := widget.NewLabel("Power Level: 1.0x")
	powerSlider := widget.NewSlider(0.1, 10.0)
	powerSlider.Value = 1.0
	powerSlider.OnChanged = func(v float64) {
		ta.powerLevel = v
		powerLabel.SetText(fmt.Sprintf("Power Level: %.1fx", v))
		ta.setPowerDistribution(ta.simulationMode)
		ta.updateHeatmap()
		ta.updateStats()
	}

	// Simulation controls
	runBtn := widget.NewButton("Run Simulation", func() {
		ta.startSimulation()
	})
	runBtn.Importance = widget.HighImportance

	stopBtn := widget.NewButton("Stop", func() {
		ta.stopSimulation()
	})

	resetBtn := widget.NewButton("Reset", func() {
		ta.resetSimulation()
	})

	// Add hotspot button
	addHotspotBtn := widget.NewButton("Add Hotspot", func() {
		ta.addHotspot()
	})

	// Educational panel
	ta.educationalPanel = widget.NewLabel(`Thermal Analysis
================

FeCIM operates at 1000x
lower power than GPUs.

Lower power = less heat
= better efficiency
= no cooling needed

This demo shows:
- Temperature distribution
- Hotspot identification
- Multi-layer thermal flow
- Technology comparison`)
	ta.educationalPanel.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		widget.NewLabel("Simulation Mode:"),
		modeSelect,
		widget.NewSeparator(),
		widget.NewLabel("View Layer:"),
		layerSelect,
		widget.NewSeparator(),
		powerLabel,
		powerSlider,
		widget.NewSeparator(),
		container.NewHBox(runBtn, stopBtn),
		container.NewHBox(resetBtn, addHotspotBtn),
		widget.NewSeparator(),
		ta.educationalPanel,
	)
}

func (ta *ThermalApp) createCenterPanel() fyne.CanvasObject {
	// Main heatmap container
	ta.heatmapView = container.NewWithoutLayout()
	ta.heatmapView.Resize(fyne.NewSize(400, 400))

	heatmapScroll := container.NewScroll(ta.heatmapView)
	heatmapScroll.SetMinSize(fyne.NewSize(380, 380))

	// Temperature scale
	scaleContainer := ta.createTemperatureScale()

	// Layer tabs for multi-layer view
	tab1Content := widget.NewLabel("Layer 1: Silicon Substrate\nHigh conductivity, heat sink contact")
	tab2Content := widget.NewLabel("Layer 2: Crossbar Array\nFeFET cells, main computation")
	tab3Content := widget.NewLabel("Layer 3: BEOL Interconnects\nMetal routing, moderate conductivity")

	ta.layerTabs = container.NewAppTabs(
		container.NewTabItem("L1: Substrate", tab1Content),
		container.NewTabItem("L2: Crossbar", tab2Content),
		container.NewTabItem("L3: BEOL", tab3Content),
	)
	ta.layerTabs.OnSelected = func(tab *container.TabItem) {
		switch tab.Text {
		case "L1: Substrate":
			ta.selectedLayer = 0
		case "L2: Crossbar":
			ta.selectedLayer = 1
		case "L3: BEOL":
			ta.selectedLayer = 2
		}
		ta.singleSim = ta.multiSim.Layers[ta.selectedLayer]
		ta.updateHeatmap()
		ta.updateStats()
	}

	heatmapLabel := widget.NewLabel("2D Heat Map (Temperature Distribution)")
	heatmapLabel.TextStyle = fyne.TextStyle{Bold: true}
	heatmapLabel.Alignment = fyne.TextAlignCenter

	return container.NewBorder(
		container.NewVBox(
			heatmapLabel,
			ta.layerTabs,
		),
		scaleContainer,
		nil,
		nil,
		heatmapScroll,
	)
}

func (ta *ThermalApp) createRightPanel() fyne.CanvasObject {
	// Stats display
	ta.statsLabel = widget.NewLabel("Loading stats...")
	ta.statsLabel.Wrapping = fyne.TextWrapWord

	// Comparison panel
	ta.comparisonPanel = container.NewVBox()
	ta.updateComparisonPanel()

	// Operation log
	ta.operationLog = widget.NewList(
		func() int { return len(ta.logEntries) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Log entry")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(ta.logEntries) {
				obj.(*widget.Label).SetText(ta.logEntries[len(ta.logEntries)-1-id])
			}
		},
	)
	ta.operationLog.Resize(fyne.NewSize(200, 100))

	return container.NewVBox(
		widget.NewLabel("Thermal Statistics"),
		widget.NewSeparator(),
		ta.statsLabel,
		widget.NewSeparator(),
		widget.NewLabel("Power/Heat Comparison"),
		ta.comparisonPanel,
		widget.NewSeparator(),
		widget.NewLabel("Operation Log"),
		container.NewGridWrap(fyne.NewSize(250, 120), ta.operationLog),
	)
}

func (ta *ThermalApp) createTemperatureScale() fyne.CanvasObject {
	// Create temperature scale bar
	scaleWidth := float32(300)
	scaleHeight := float32(20)

	scaleContainer := container.NewWithoutLayout()

	// Temperature gradient rectangles
	steps := 10
	stepWidth := scaleWidth / float32(steps)
	for i := 0; i < steps; i++ {
		ratio := float64(i) / float64(steps-1)
		r, g, b := temperatureToRGB(ratio)
		rect := canvas.NewRectangle(color.RGBA{r, g, b, 255})
		rect.Resize(fyne.NewSize(stepWidth+1, scaleHeight))
		rect.Move(fyne.NewPos(float32(i)*stepWidth, 0))
		scaleContainer.Add(rect)
	}

	// Labels
	minLabel := canvas.NewText("25°C", color.White)
	minLabel.TextSize = 10
	minLabel.Move(fyne.NewPos(0, scaleHeight+2))
	scaleContainer.Add(minLabel)

	maxLabel := canvas.NewText("85°C", color.White)
	maxLabel.TextSize = 10
	maxLabel.Move(fyne.NewPos(scaleWidth-30, scaleHeight+2))
	scaleContainer.Add(maxLabel)

	scaleContainer.Resize(fyne.NewSize(scaleWidth, scaleHeight+20))

	return container.NewCenter(scaleContainer)
}

// temperatureToRGB converts a 0-1 temperature ratio to RGB color.
func temperatureToRGB(ratio float64) (uint8, uint8, uint8) {
	// Blue (cool) -> Green -> Yellow -> Red (hot)
	if ratio < 0.25 {
		// Blue to Cyan
		t := ratio / 0.25
		return 0, uint8(t * 200), 255
	} else if ratio < 0.5 {
		// Cyan to Green
		t := (ratio - 0.25) / 0.25
		return 0, 200, uint8(255 * (1 - t))
	} else if ratio < 0.75 {
		// Green to Yellow
		t := (ratio - 0.5) / 0.25
		return uint8(255 * t), 200, 0
	} else {
		// Yellow to Red
		t := (ratio - 0.75) / 0.25
		return 255, uint8(200 * (1 - t)), 0
	}
}

func (ta *ThermalApp) updateHeatmap() {
	ta.heatmapView.Objects = nil

	if ta.singleSim == nil {
		return
	}

	grid := ta.singleSim.GetGridCopy()
	height := len(grid)
	if height == 0 {
		return
	}
	width := len(grid[0])

	cellSize := float32(12)
	padding := float32(1)

	minT := ta.singleSim.AmbientTemp
	maxT := ta.singleSim.MaxTemp

	// Find hotspots
	hotspots := ta.singleSim.FindHotspots(maxT * 0.75)
	hotspotMap := make(map[string]bool)
	for _, h := range hotspots {
		hotspotMap[fmt.Sprintf("%d,%d", h.X, h.Y)] = true
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			temp := grid[y][x]
			ratio := (temp - minT) / (maxT - minT)
			if ratio < 0 {
				ratio = 0
			}
			if ratio > 1 {
				ratio = 1
			}

			r, g, b := temperatureToRGB(ratio)

			rect := canvas.NewRectangle(color.RGBA{r, g, b, 255})
			rect.Resize(fyne.NewSize(cellSize, cellSize))
			rect.Move(fyne.NewPos(float32(x)*(cellSize+padding)+30, float32(y)*(cellSize+padding)+30))
			ta.heatmapView.Add(rect)

			// Mark hotspots
			key := fmt.Sprintf("%d,%d", x, y)
			if hotspotMap[key] {
				mark := canvas.NewText("!", color.White)
				mark.TextSize = 10
				mark.TextStyle = fyne.TextStyle{Bold: true}
				mark.Move(fyne.NewPos(float32(x)*(cellSize+padding)+33, float32(y)*(cellSize+padding)+30))
				ta.heatmapView.Add(mark)
			}
		}
	}

	// Add axis labels
	for i := 0; i < width; i += 8 {
		label := canvas.NewText(fmt.Sprintf("%d", i), color.RGBA{200, 200, 200, 255})
		label.TextSize = 8
		label.Move(fyne.NewPos(float32(i)*(cellSize+padding)+30, 15))
		ta.heatmapView.Add(label)
	}

	ta.heatmapView.Refresh()
}

func (ta *ThermalApp) updateStats() {
	if ta.singleSim == nil {
		return
	}

	globalMax := ta.multiSim.GetGlobalMaxTemp()
	globalMin := ta.multiSim.GetGlobalMinTemp()
	globalAvg := ta.multiSim.GetStackAverageTemp()
	totalPower := ta.multiSim.TotalHeatGeneration()

	layerMax := ta.singleSim.GetMaxTemperature()
	layerAvg := ta.singleSim.GetAverageTemperature()

	hotspots := ta.singleSim.FindHotspots(ta.singleSim.MaxTemp * 0.6)

	stats := fmt.Sprintf(`Layer %d Statistics
==================
Max Temp: %.1f°C
Avg Temp: %.1f°C
Hotspots: %d

Stack Statistics
================
Global Max: %.1f°C
Global Min: %.1f°C
Stack Avg: %.1f°C
Total Power: %.2e W/m²

Thermal Margin
==============
Max Safe: %.1f°C
Headroom: %.1f°C`,
		ta.selectedLayer+1,
		layerMax, layerAvg, len(hotspots),
		globalMax, globalMin, globalAvg, totalPower,
		ta.singleSim.MaxTemp,
		ta.singleSim.MaxTemp-globalMax)

	ta.statsLabel.SetText(stats)

	// Check warning
	warning := ta.multiSim.CheckStackWarning()
	if warning != nil {
		ta.statusLabel.SetText(fmt.Sprintf("[Level %d] %s - Max: %.1f°C",
			warning.Level, warning.Message, warning.MaxTemp))
	} else {
		ta.statusLabel.SetText(fmt.Sprintf("Normal operation - Max: %.1f°C, Avg: %.1f°C",
			globalMax, globalAvg))
	}
}

func (ta *ThermalApp) updateComparisonPanel() {
	ta.comparisonPanel.Objects = nil

	// Power comparison data
	comparisons := []struct {
		name  string
		power float64
		color color.Color
	}{
		{"FeCIM", 1.0, color.RGBA{0, 220, 150, 255}},
		{"DRAM", 100.0, color.RGBA{255, 200, 0, 255}},
		{"GPU", 1000.0, color.RGBA{255, 80, 80, 255}},
	}

	maxPower := 1000.0

	for _, c := range comparisons {
		label := widget.NewLabel(fmt.Sprintf("%-6s: %.0fx", c.name, c.power))
		bar := widget.NewProgressBar()
		bar.SetValue(c.power / maxPower)

		colorRect := canvas.NewRectangle(c.color)
		colorRect.SetMinSize(fyne.NewSize(15, 15))

		row := container.NewHBox(colorRect, label)
		ta.comparisonPanel.Add(row)
		ta.comparisonPanel.Add(bar)
	}

	// Summary
	summary := widget.NewLabel("FeCIM: 1000x less heat!\n→ No active cooling needed")
	summary.TextStyle = fyne.TextStyle{Bold: true}
	ta.comparisonPanel.Add(summary)

	ta.comparisonPanel.Refresh()
}

func (ta *ThermalApp) setPowerDistribution(mode string) {
	ta.multiSim.Reset()

	basePower := 1e6 * ta.powerLevel // W/m² base

	switch mode {
	case "FeCIM (Low Power)":
		basePower *= 0.001 // 1000x lower
	case "GPU (High Power)":
		basePower *= 1.0 // Full power
	case "DRAM (Medium Power)":
		basePower *= 0.1 // 10x lower than GPU
	}

	// Set power for crossbar layer (layer 1)
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			// Create a realistic power pattern
			// Higher power near center (active region)
			centerDist := float64((x-16)*(x-16)+(y-16)*(y-16)) / float64(16*16)
			cellPower := basePower * (1.0 - 0.3*centerDist)
			ta.multiSim.SetCellPower(1, x, y, cellPower)
		}
	}

	// Run initial simulation to reach quasi-steady state
	ta.multiSim.StepMultiple(100, 1e-6)
}

func (ta *ThermalApp) startSimulation() {
	if ta.running {
		return
	}

	ta.running = true
	ta.stopChan = make(chan bool)
	ta.addLog("Simulation started")
	ta.statusLabel.SetText("Simulation running...")

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ta.stopChan:
				return
			case <-ticker.C:
				ta.multiSim.StepMultiple(10, 1e-6)
				fyne.Do(func() {
					ta.updateHeatmap()
					ta.updateStats()
				})
			}
		}
	}()
}

func (ta *ThermalApp) stopSimulation() {
	if !ta.running {
		return
	}

	ta.running = false
	close(ta.stopChan)
	ta.addLog("Simulation stopped")
	ta.statusLabel.SetText("Simulation stopped")
}

func (ta *ThermalApp) resetSimulation() {
	ta.stopSimulation()
	ta.multiSim.Reset()
	ta.setPowerDistribution(ta.simulationMode)
	ta.addLog("Simulation reset")
	ta.updateHeatmap()
	ta.updateStats()
}

func (ta *ThermalApp) addHotspot() {
	// Add a hotspot in the center of the crossbar layer
	hotspotPower := 1e8 // Strong hotspot
	ta.multiSim.SetCellPower(1, 16, 16, hotspotPower)
	ta.multiSim.SetCellPower(1, 15, 16, hotspotPower*0.7)
	ta.multiSim.SetCellPower(1, 17, 16, hotspotPower*0.7)
	ta.multiSim.SetCellPower(1, 16, 15, hotspotPower*0.7)
	ta.multiSim.SetCellPower(1, 16, 17, hotspotPower*0.7)

	ta.multiSim.StepMultiple(50, 1e-6)
	ta.addLog("Hotspot added at center")
	ta.updateHeatmap()
	ta.updateStats()
}

func (ta *ThermalApp) addLog(entry string) {
	timestamp := time.Now().Format("15:04:05")
	ta.logEntries = append(ta.logEntries, fmt.Sprintf("[%s] %s", timestamp, entry))
	if len(ta.logEntries) > 20 {
		ta.logEntries = ta.logEntries[1:]
	}
	if ta.operationLog != nil {
		ta.operationLog.Refresh()
	}
}

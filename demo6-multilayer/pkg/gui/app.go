// Package gui provides Fyne-based GUI for the 3D multilayer stack visualization.
package gui

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/demo6-multilayer/pkg/multilayer"
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

var debug *log.Logger
var logFile *os.File

func init() {
	logsDir := "<local-path>"
	os.MkdirAll(logsDir, 0755)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logsDir, timestamp+"-multilayer-demo06.log")

	var err error
	logFile, err = os.Create(logPath)
	if err != nil {
		debug = log.New(os.Stdout, "[DEBUG] ", log.Ltime|log.Lmicroseconds)
		debug.Printf("Failed to create log file: %v, using stdout", err)
		return
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	debug = log.New(multiWriter, "[DEBUG] ", log.Ltime|log.Lmicroseconds)
	debug.Printf("Logging to: %s", logPath)
}

// MultilayerApp is the main application for the 3D stack demo.
type MultilayerApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Data
	stack       *multilayer.Stack
	stackSelect *widget.Select

	// UI Components
	stackView    *fyne.Container
	metricsLabel *widget.Label
	energyBars   *fyne.Container
	layerList    *widget.List
	statusLabel  *widget.Label
}

// NewMultilayerApp creates a new multilayer visualization app.
func NewMultilayerApp() *MultilayerApp {
	return &MultilayerApp{}
}

// Run starts the application.
func (ma *MultilayerApp) Run() {
	ma.fyneApp = app.NewWithID("com.fecim.multilayer-demo")
	ma.fyneApp.Settings().SetTheme(&feCIMTheme{})

	ma.window = ma.fyneApp.NewWindow("FeCIM Demo 6: 3D Multilayer Stack")
	ma.window.Resize(fyne.NewSize(1200, 800))

	// Initialize with demo stack
	ma.stack = multilayer.SmallStack()

	content := ma.createMainLayout()
	ma.window.SetContent(content)

	ma.updateStackView()
	ma.updateMetrics()

	debug.Printf("MultilayerApp started")
	ma.window.ShowAndRun()
}

func (ma *MultilayerApp) createMainLayout() fyne.CanvasObject {
	// Header
	header := ma.createHeader()

	// Left panel: Stack selector and layer list
	leftPanel := ma.createLeftPanel()

	// Center: 3D Stack visualization
	centerPanel := ma.createCenterPanel()

	// Right panel: Metrics and energy comparison
	rightPanel := ma.createRightPanel()

	// Status bar
	ma.statusLabel = widget.NewLabel("Ready")

	// Main layout
	mainContent := container.NewHSplit(
		leftPanel,
		container.NewHSplit(
			centerPanel,
			rightPanel,
		),
	)
	mainContent.SetOffset(0.2)

	return container.NewBorder(
		header,
		ma.statusLabel,
		nil,
		nil,
		mainContent,
	)
}

func (ma *MultilayerApp) createHeader() fyne.CanvasObject {
	title := canvas.NewText("FeCIM Demo 6: 3D Multilayer Stack Architecture", color.White)
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}

	quote := widget.NewLabel("\"Vertical stacking enables massive parallelism in neuromorphic computing.\"")
	quote.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewCenter(title),
		container.NewCenter(quote),
		widget.NewSeparator(),
	)
}

func (ma *MultilayerApp) createLeftPanel() fyne.CanvasObject {
	// Stack selector
	ma.stackSelect = widget.NewSelect([]string{"Demo Stack (16-8-4)", "MNIST Stack (784-128-64-10)"}, func(s string) {
		if s == "MNIST Stack (784-128-64-10)" {
			ma.stack = multilayer.MNISTStack()
		} else {
			ma.stack = multilayer.SmallStack()
		}
		ma.updateStackView()
		ma.updateMetrics()
		ma.statusLabel.SetText("Loaded: " + s)
		debug.Printf("Stack changed to: %s", s)
	})
	ma.stackSelect.SetSelected("Demo Stack (16-8-4)")

	// Layer list
	ma.layerList = widget.NewList(
		func() int {
			if ma.stack == nil {
				return 0
			}
			return len(ma.stack.Layers)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Layer X"),
				layout.NewSpacer(),
				widget.NewLabel("NxM"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if ma.stack == nil || id >= len(ma.stack.Layers) {
				return
			}
			layer := ma.stack.Layers[id]
			hbox := obj.(*fyne.Container)
			hbox.Objects[0].(*widget.Label).SetText(fmt.Sprintf("L%d: %s", id+1, layer.Name))
			hbox.Objects[2].(*widget.Label).SetText(fmt.Sprintf("%dx%d", layer.Rows, layer.Cols))
		},
	)
	ma.layerList.OnSelected = func(id widget.ListItemID) {
		if ma.stack != nil && id < len(ma.stack.Layers) {
			layer := ma.stack.Layers[id]
			ma.statusLabel.SetText(fmt.Sprintf("Selected: %s (%dx%d = %d cells)",
				layer.Name, layer.Rows, layer.Cols, layer.Rows*layer.Cols))
		}
	}

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Stack Configuration"),
			ma.stackSelect,
			widget.NewSeparator(),
			widget.NewLabel("Layers"),
		),
		nil, nil, nil,
		ma.layerList,
	)
}

func (ma *MultilayerApp) createCenterPanel() fyne.CanvasObject {
	ma.stackView = container.NewWithoutLayout()
	ma.stackView.Resize(fyne.NewSize(400, 500))

	scrollContainer := container.NewScroll(ma.stackView)
	scrollContainer.SetMinSize(fyne.NewSize(400, 400))

	return container.NewBorder(
		widget.NewLabel("3D Stack Visualization"),
		nil, nil, nil,
		scrollContainer,
	)
}

func (ma *MultilayerApp) createRightPanel() fyne.CanvasObject {
	// Metrics display
	ma.metricsLabel = widget.NewLabel("Loading metrics...")
	ma.metricsLabel.Wrapping = fyne.TextWrapWord

	// Energy comparison bars
	ma.energyBars = container.NewVBox()

	return container.NewVBox(
		widget.NewLabel("Stack Metrics"),
		widget.NewSeparator(),
		ma.metricsLabel,
		widget.NewSeparator(),
		widget.NewLabel("Energy Comparison"),
		ma.energyBars,
	)
}

func (ma *MultilayerApp) updateStackView() {
	ma.stackView.Objects = nil

	if ma.stack == nil || len(ma.stack.Layers) == 0 {
		return
	}

	// Find max dimensions for scaling
	maxCols := 0
	for _, layer := range ma.stack.Layers {
		if layer.Cols > maxCols {
			maxCols = layer.Cols
		}
	}

	// Draw layers as stacked rectangles (isometric-like view)
	baseX := float32(50)
	baseY := float32(400)
	layerHeight := float32(60)
	maxWidth := float32(300)

	colors := []color.Color{
		color.RGBA{0, 180, 220, 255},   // Cyan
		color.RGBA{0, 220, 180, 255},   // Teal
		color.RGBA{100, 200, 255, 255}, // Light blue
		color.RGBA{0, 150, 200, 255},   // Blue
	}

	for i := len(ma.stack.Layers) - 1; i >= 0; i-- {
		layer := ma.stack.Layers[i]

		// Scale width based on layer size
		width := maxWidth * float32(layer.Cols) / float32(maxCols)
		if width < 80 {
			width = 80
		}

		y := baseY - float32(len(ma.stack.Layers)-1-i)*layerHeight
		x := baseX + float32(len(ma.stack.Layers)-1-i)*10

		// Layer rectangle
		rect := canvas.NewRectangle(colors[i%len(colors)])
		rect.StrokeColor = colorPrimary
		rect.StrokeWidth = 2
		rect.CornerRadius = 4
		rect.Resize(fyne.NewSize(width, layerHeight-10))
		rect.Move(fyne.NewPos(x, y))
		ma.stackView.Add(rect)

		// Layer label
		label := canvas.NewText(fmt.Sprintf("L%d: %s", i+1, layer.Name), color.White)
		label.TextSize = 12
		label.Move(fyne.NewPos(x+5, y+5))
		ma.stackView.Add(label)

		// Dimensions
		dims := canvas.NewText(fmt.Sprintf("%dx%d", layer.Rows, layer.Cols), color.RGBA{200, 200, 200, 255})
		dims.TextSize = 10
		dims.Move(fyne.NewPos(x+5, y+25))
		ma.stackView.Add(dims)

		// Via connection (if not bottom layer)
		if i > 0 {
			via := canvas.NewLine(colorPrimary)
			via.StrokeWidth = 2
			via.Position1 = fyne.NewPos(x+width/2, y+layerHeight-10)
			via.Position2 = fyne.NewPos(x+width/2+10, y+layerHeight+10)
			ma.stackView.Add(via)

			viaLabel := canvas.NewText(fmt.Sprintf("%d vias", layer.Rows), color.RGBA{150, 150, 150, 255})
			viaLabel.TextSize = 9
			viaLabel.Move(fyne.NewPos(x+width/2+15, y+layerHeight))
			ma.stackView.Add(viaLabel)
		}
	}

	// Input arrow
	inputArrow := canvas.NewText("Input", colorPrimary)
	inputArrow.TextSize = 12
	inputArrow.Move(fyne.NewPos(baseX, baseY+20))
	ma.stackView.Add(inputArrow)

	ma.stackView.Refresh()
}

func (ma *MultilayerApp) updateMetrics() {
	if ma.stack == nil {
		return
	}

	metrics := fmt.Sprintf(`Stack: %s

Layers: %d
Total Cells: %d
Total Parameters: %d
Bits per Cell: %.2f

Physical:
  Stack Height: %.0f nm
  Footprint: %.2f um2
  Areal Density: %.2f bits/um2
  Volume Density: %.2f bits/um3

Total Vias: %d`,
		ma.stack.Name,
		len(ma.stack.Layers),
		ma.stack.TotalCells(),
		ma.stack.TotalParameters(),
		ma.stack.BitsPerCell(),
		ma.stack.StackHeight(),
		ma.stack.FootprintArea(),
		ma.stack.ArealDensity(),
		ma.stack.VolumetricDensity(),
		ma.stack.TotalVias,
	)

	ma.metricsLabel.SetText(metrics)

	// Update energy bars
	ma.energyBars.Objects = nil
	energyEst := ma.stack.EstimateEnergy()

	totalCIM := 0.0
	totalTraditional := 0.0
	for _, e := range energyEst {
		totalCIM += e.TotalEnergy
		totalTraditional += e.TotalEnergy * e.TraditionalComp
	}

	// FeCIM bar
	cimLabel := widget.NewLabel(fmt.Sprintf("FeCIM: %.3f pJ", totalCIM))
	cimBar := widget.NewProgressBar()
	cimBar.SetValue(totalCIM / totalTraditional)
	ma.energyBars.Add(cimLabel)
	ma.energyBars.Add(cimBar)

	// Traditional bar
	tradLabel := widget.NewLabel(fmt.Sprintf("Traditional: %.1f pJ", totalTraditional))
	tradBar := widget.NewProgressBar()
	tradBar.SetValue(1.0)
	ma.energyBars.Add(tradLabel)
	ma.energyBars.Add(tradBar)

	// Advantage
	advLabel := widget.NewLabel(fmt.Sprintf("FeCIM Advantage: %.0fx lower energy!", totalTraditional/totalCIM))
	advLabel.TextStyle = fyne.TextStyle{Bold: true}
	ma.energyBars.Add(advLabel)

	ma.energyBars.Refresh()
}

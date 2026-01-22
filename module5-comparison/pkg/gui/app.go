// Package gui provides Fyne-based GUI components for architecture comparison.
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
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var debug *log.Logger
var logFile *os.File

func init() {
	logsDir := "<local-path>"
	os.MkdirAll(logsDir, 0755)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logsDir, timestamp+"-comparison-module05.log")

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

// FeCIM theme colors
var (
	colorBackground = color.RGBA{0, 50, 100, 255}
	colorPrimary    = color.RGBA{0, 212, 255, 255}
)

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

// EnergySpec holds energy per MAC specifications with sources.
type EnergySpec struct {
	Name          string
	EnergyFJ      float64 // femtojoules per MAC
	Source        string
	Verified      bool
	SourceDetails string
}

// ComparisonApp is the main application for architecture comparison.
type ComparisonApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Energy specs (honest numbers with sources)
	cpuSpec   EnergySpec
	gpuSpec   EnergySpec
	fecimSpec EnergySpec

	// GUI components
	energyChart      *EnergyBarChart
	archDiagram      *ArchitectureDiagram
	calculator       *DataCenterCalculator
	verifiedTable    *VerifiedClaimsTable
	educationalPanel *ComparisonEducationalPanel
	operationLog     *ComparisonOperationLog
	modeIndicator    *ComparisonModeIndicator

	// Controls
	workloadSelect   *widget.Select
	inferencesSlider *widget.Slider
	inferencesLabel  *widget.Label

	// Status
	statusLabel *widget.Label

	// Current settings
	currentWorkload   string
	currentInferences float64
}

// NewComparisonApp creates the comparison demo application.
func NewComparisonApp() *ComparisonApp {
	debug.Println("NewComparisonApp: Creating application")
	ca := &ComparisonApp{
		currentWorkload:   "MNIST",
		currentInferences: 10000,
	}

	ca.fyneApp = app.NewWithID("com.fecim.comparison-demo")
	ca.fyneApp.Settings().SetTheme(&feCIMTheme{})

	// Initialize energy specs with HONEST numbers and sources
	ca.cpuSpec = EnergySpec{
		Name:          "CPU + DRAM",
		EnergyFJ:      1000, // ~1000 fJ/MAC
		Source:        "Intel/AMD published specs",
		Verified:      true,
		SourceDetails: "Includes memory access energy. Intel Xeon specs, AMD EPYC specs.",
	}

	ca.gpuSpec = EnergySpec{
		Name:          "GPU + HBM",
		EnergyFJ:      100, // ~100 fJ/MAC
		Source:        "NVIDIA H100 specifications",
		Verified:      true,
		SourceDetails: "H100 SXM: 700W TDP, ~3958 TFLOPS FP16. ~177 fJ/FLOP.",
	}

	ca.fecimSpec = EnergySpec{
		Name:          "FeCIM",
		EnergyFJ:      10, // ~1-10 fJ/MAC (claimed)
		Source:        "Dr. Tour's presentation (NOT independently verified)",
		Verified:      false,
		SourceDetails: "Claimed: '10M× lower energy than NAND'. TRL 4 - lab only.",
	}

	debug.Println("NewComparisonApp: Initialization complete")
	return ca
}

// Run starts the GUI application.
func (ca *ComparisonApp) Run() {
	debug.Println("App: Creating window")
	ca.window = ca.fyneApp.NewWindow("FeCIM Demo 8: Architecture Comparison")
	ca.window.Resize(fyne.NewSize(1400, 900))

	content := ca.createMainLayout()
	ca.window.SetContent(content)

	ca.updateCalculations()
	ca.updateStatus("Ready. Select workload and adjust parameters.")

	debug.Println("App: ShowAndRun starting")
	ca.window.ShowAndRun()
}

// createMainLayout builds the main application layout.
func (ca *ComparisonApp) createMainLayout() fyne.CanvasObject {
	// Create components
	ca.energyChart = NewEnergyBarChart()
	ca.archDiagram = NewArchitectureDiagram()
	ca.calculator = NewDataCenterCalculator()
	ca.verifiedTable = NewVerifiedClaimsTable()
	ca.educationalPanel = NewComparisonEducationalPanel()
	ca.operationLog = NewComparisonOperationLog()
	ca.modeIndicator = NewComparisonModeIndicator()

	// Set initial energy values
	ca.energyChart.SetValues(ca.cpuSpec, ca.gpuSpec, ca.fecimSpec)

	// Workload selector
	ca.workloadSelect = widget.NewSelect(
		[]string{"MNIST", "ResNet-50", "BERT-Base", "GPT-2", "LLM-70B"},
		ca.onWorkloadChanged,
	)
	ca.workloadSelect.SetSelected("MNIST")

	// Inferences slider
	ca.inferencesLabel = widget.NewLabel("Inferences/sec: 10,000")
	ca.inferencesSlider = widget.NewSlider(100, 100000)
	ca.inferencesSlider.Value = 10000
	ca.inferencesSlider.OnChanged = func(v float64) {
		ca.currentInferences = v
		ca.inferencesLabel.SetText(fmt.Sprintf("Inferences/sec: %.0f", v))
		ca.updateCalculations()
	}

	// Status
	ca.statusLabel = widget.NewLabel("Status: Ready")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Calculate button
	calcBtn := widget.NewButton("Calculate", func() {
		ca.updateCalculations()
	})
	calcBtn.Importance = widget.HighImportance

	// Header
	titleLabel := widget.NewLabel("FeCIM Architecture Comparison")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	header := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
	)

	// Left panel: Controls
	controlsLabel := widget.NewLabel("Configuration")
	controlsLabel.TextStyle = fyne.TextStyle{Bold: true}

	leftPanel := container.NewVBox(
		controlsLabel,
		widget.NewSeparator(),
		widget.NewLabel("Workload:"),
		ca.workloadSelect,
		widget.NewSeparator(),
		ca.inferencesLabel,
		ca.inferencesSlider,
		widget.NewSeparator(),
		calcBtn,
		widget.NewSeparator(),
		ca.verifiedTable,
	)

	// Center panel: Charts
	energyLabel := widget.NewLabel("Energy per MAC Operation (fJ)")
	energyLabel.TextStyle = fyne.TextStyle{Bold: true}
	energyLabel.Alignment = fyne.TextAlignCenter

	archLabel := widget.NewLabel("Architecture Comparison")
	archLabel.TextStyle = fyne.TextStyle{Bold: true}
	archLabel.Alignment = fyne.TextAlignCenter

	calcLabel := widget.NewLabel("Data Center Calculator")
	calcLabel.TextStyle = fyne.TextStyle{Bold: true}
	calcLabel.Alignment = fyne.TextAlignCenter

	centerTop := container.NewVBox(
		energyLabel,
		ca.energyChart,
	)

	centerMid := container.NewVBox(
		widget.NewSeparator(),
		archLabel,
		ca.archDiagram,
	)

	centerBottom := container.NewVBox(
		widget.NewSeparator(),
		calcLabel,
		ca.calculator,
	)

	centerPanel := container.NewVBox(
		centerTop,
		centerMid,
		centerBottom,
	)

	// Right panel: Educational + Log
	rightPanel := container.NewVBox(
		ca.educationalPanel,
		widget.NewSeparator(),
		ca.operationLog,
	)

	// Footer
	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			ca.modeIndicator,
			widget.NewSeparator(),
			ca.statusLabel,
			layout.NewSpacer(),
			widget.NewLabel("TRL 4 | Lab Validation Only | Sources in footnotes"),
		),
	)

	// Main layout using HSplit
	leftCenterSplit := container.NewHSplit(
		container.NewPadded(leftPanel),
		container.NewScroll(centerPanel),
	)
	leftCenterSplit.SetOffset(0.25)

	mainSplit := container.NewHSplit(
		leftCenterSplit,
		container.NewPadded(rightPanel),
	)
	mainSplit.SetOffset(0.75)

	mainContent := container.NewBorder(
		header,
		footer,
		nil,
		nil,
		mainSplit,
	)

	return mainContent
}

// onWorkloadChanged handles workload selection.
func (ca *ComparisonApp) onWorkloadChanged(workload string) {
	ca.currentWorkload = workload
	ca.operationLog.Add(fmt.Sprintf("Workload: %s", workload))
	ca.updateCalculations()
}

// updateCalculations recalculates all values.
func (ca *ComparisonApp) updateCalculations() {
	debug.Printf("updateCalculations: workload=%s, inferences=%.0f", ca.currentWorkload, ca.currentInferences)

	// Get MACs for workload
	macs := ca.getWorkloadMACs()

	// Calculate energy per inference (µJ)
	cpuEnergy := float64(macs) * ca.cpuSpec.EnergyFJ / 1e9 // fJ to µJ
	gpuEnergy := float64(macs) * ca.gpuSpec.EnergyFJ / 1e9
	fecimEnergy := float64(macs) * ca.fecimSpec.EnergyFJ / 1e9

	// Calculate power for target inferences/sec (W)
	cpuPower := cpuEnergy * ca.currentInferences / 1e6 // µJ * inf/s = µW, /1e6 = W
	gpuPower := gpuEnergy * ca.currentInferences / 1e6
	fecimPower := fecimEnergy * ca.currentInferences / 1e6

	// Monthly cost at $0.10/kWh
	hoursPerMonth := 730.0
	cpuCost := cpuPower / 1000 * hoursPerMonth * 0.10
	gpuCost := gpuPower / 1000 * hoursPerMonth * 0.10
	fecimCost := fecimPower / 1000 * hoursPerMonth * 0.10

	// Update calculator
	ca.calculator.SetResults(
		ca.currentWorkload,
		macs,
		ca.currentInferences,
		cpuEnergy, gpuEnergy, fecimEnergy,
		cpuPower, gpuPower, fecimPower,
		cpuCost, gpuCost, fecimCost,
	)

	// Update educational panel
	ca.educationalPanel.SetComparison(
		cpuPower/fecimPower,
		gpuPower/fecimPower,
	)

	// Log
	ca.operationLog.Add(fmt.Sprintf("Calculated: %.0f MACs × %.0f inf/s", float64(macs), ca.currentInferences))
	ca.operationLog.Add(fmt.Sprintf("  CPU: %.1fW, GPU: %.1fW, FeCIM: %.2fW", cpuPower, gpuPower, fecimPower))

	ca.modeIndicator.SetMode(ComparisonModeCalculating)
	ca.updateStatus(fmt.Sprintf("Calculated for %s @ %.0f inf/s", ca.currentWorkload, ca.currentInferences))

	// Reset mode after brief delay
	go func() {
		time.Sleep(500 * time.Millisecond)
		fyne.Do(func() {
			ca.modeIndicator.SetMode(ComparisonModeIdle)
		})
	}()
}

// getWorkloadMACs returns MACs for the current workload.
func (ca *ComparisonApp) getWorkloadMACs() int {
	switch ca.currentWorkload {
	case "MNIST":
		return 101632 // 784*128 + 128*10
	case "ResNet-50":
		return 4000000000 // ~4B MACs
	case "BERT-Base":
		return 11000000000 // ~11B MACs
	case "GPT-2":
		return 35000000000 // ~35B MACs
	case "LLM-70B":
		return 140000000000000 // ~140T MACs
	default:
		return 101632
	}
}

// updateStatus updates the status label.
func (ca *ComparisonApp) updateStatus(status string) {
	if ca.statusLabel == nil {
		return
	}
	ca.statusLabel.SetText("Status: " + status)
}

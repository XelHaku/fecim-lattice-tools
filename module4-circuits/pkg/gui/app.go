// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
package gui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module4-circuits/pkg/peripherals"
)

// FeCIM theme colors - same as demo1
var (
	colorBackground = color.RGBA{0, 50, 100, 255}  // FeCIM blue #003264
	colorPrimary    = color.RGBA{0, 212, 255, 255} // Cyan
)

// feCIMTheme implements fyne.Theme for consistent FeCIM branding
type feCIMTheme struct{}

func (t *feCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBackground // FeCIM blue #003264
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255} // Slightly lighter blue
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255} // Darker blue for inputs
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255} // Separator lines
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

// CircuitsApp is the main application for the peripheral circuits demo.
type CircuitsApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Peripheral components
	dac  *peripherals.DAC
	adc  *peripherals.ADC
	tia  *peripherals.TIA
	pump *peripherals.ChargePump

	// GUI components
	signalFlow    *SignalFlowDiagram
	timingDiagram *TimingDiagram
	levelSlider   *widget.Slider
	levelLabel    *widget.Label

	// Live Slide components
	modeIndicator    *CircuitModeIndicator
	educationalPanel *CircuitEducationalPanel
	operationLog     *CircuitOperationLog
	keyStat          *CircuitKeyStat
	demoModeSelect   *widget.Select

	// Value displays
	dacValueLabel  *widget.Label
	adcValueLabel  *widget.Label
	tiaValueLabel  *widget.Label
	pumpValueLabel *widget.Label

	// Status
	statusLabel *widget.Label

	// Current level
	currentLevel int

	// Auto demo state
	autoDemo      bool
	autoDemoTimer *time.Ticker
	stopAutoDemo  chan bool
	autoDemoPhase int
}

// NewCircuitsApp creates and initializes the circuits demo application.
func NewCircuitsApp() *CircuitsApp {
	ca := &CircuitsApp{
		currentLevel: 15,
	}

	// Create Fyne app
	ca.fyneApp = app.NewWithID("com.fecim.circuits-demo")
	ca.fyneApp.Settings().SetTheme(&feCIMTheme{})

	// Initialize peripheral components
	ca.dac = peripherals.DefaultDAC()
	ca.adc = peripherals.DefaultADC()
	ca.tia = peripherals.DefaultTIA()
	ca.pump = peripherals.DefaultChargePump()

	return ca
}

// Run starts the GUI application.
func (ca *CircuitsApp) Run() {
	ca.window = ca.fyneApp.NewWindow("FeCIM Demo 4: Peripheral Circuits")
	ca.window.Resize(fyne.NewSize(1400, 900))

	// Create main layout
	content := ca.createMainLayout()
	ca.window.SetContent(content)

	// Initialize displays
	ca.updateValues()
	ca.updateStatus("Ready. Select a circuit or run a cycle.")

	ca.window.ShowAndRun()
}

// createMainLayout builds the main application layout.
func (ca *CircuitsApp) createMainLayout() fyne.CanvasObject {
	// Create visualization components
	ca.signalFlow = NewSignalFlowDiagram()
	ca.timingDiagram = NewTimingDiagram()

	// Create Live Slide components
	ca.modeIndicator = NewCircuitModeIndicator()
	ca.educationalPanel = NewCircuitEducationalPanel()
	ca.educationalPanel.SetIdleExplanation()
	ca.operationLog = NewCircuitOperationLog()
	ca.keyStat = NewCircuitKeyStat("CMOS Compatible", "Standard Process")

	// Demo mode selector
	ca.demoModeSelect = widget.NewSelect(
		[]string{"Manual", "Auto Demo", "Step-by-Step"},
		ca.onDemoModeChanged,
	)
	ca.demoModeSelect.SetSelected("Manual")

	// Level slider
	ca.levelLabel = widget.NewLabel("Level: 15")
	ca.levelSlider = widget.NewSlider(0, 29)
	ca.levelSlider.Value = 15
	ca.levelSlider.OnChanged = func(v float64) {
		ca.currentLevel = int(v)
		ca.levelLabel.SetText(fmt.Sprintf("Level: %d", ca.currentLevel))
		ca.updateValues()
	}

	// Value display labels
	ca.dacValueLabel = widget.NewLabel("DAC: -")
	ca.adcValueLabel = widget.NewLabel("ADC: -")
	ca.tiaValueLabel = widget.NewLabel("TIA: -")
	ca.pumpValueLabel = widget.NewLabel("Pump: -")

	// Status label
	ca.statusLabel = widget.NewLabel("● IDLE | Ready for operations")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Control buttons
	dacBtn := widget.NewButton("DAC Convert", func() {
		ca.runDACDemo()
	})
	adcBtn := widget.NewButton("ADC Convert", func() {
		ca.runADCDemo()
	})
	tiaBtn := widget.NewButton("TIA Demo", func() {
		ca.runTIADemo()
	})
	pumpBtn := widget.NewButton("Charge Pump", func() {
		ca.runPumpDemo()
	})
	writeBtn := widget.NewButton("Write Cycle", func() {
		ca.runWriteCycle()
	})
	writeBtn.Importance = widget.HighImportance
	readBtn := widget.NewButton("Read Cycle", func() {
		ca.runReadCycle()
	})
	readBtn.Importance = widget.HighImportance

	// Button container
	buttonBox := container.NewHBox(
		dacBtn, adcBtn, tiaBtn, pumpBtn,
		widget.NewSeparator(),
		writeBtn, readBtn,
	)

	// Title and header
	titleLabel := widget.NewLabel("FeCIM Peripheral Circuits")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	specsLabel := widget.NewLabel("DAC (5-bit) → Charge Pump → FeFET → TIA → ADC (5-bit) | 30 Levels")
	specsLabel.Alignment = fyne.TextAlignCenter

	header := container.NewVBox(
		titleLabel,
		specsLabel,
		widget.NewSeparator(),
	)

	// Left panel: Controls + values
	controlPanel := container.NewVBox(
		widget.NewLabel("Demo Mode:"),
		ca.demoModeSelect,
		widget.NewSeparator(),
		ca.levelLabel,
		ca.levelSlider,
		widget.NewSeparator(),
		widget.NewLabel("Circuit Values:"),
		ca.dacValueLabel,
		ca.pumpValueLabel,
		ca.tiaValueLabel,
		ca.adcValueLabel,
		widget.NewSeparator(),
		buttonBox,
	)

	// Center panel: Visualizations
	signalLabel := widget.NewLabel("Signal Flow: Digital → DAC → Pump → Cell → TIA → ADC → Digital")
	signalLabel.TextStyle = fyne.TextStyle{Bold: true}
	signalLabel.Alignment = fyne.TextAlignCenter

	timingLabel := widget.NewLabel("Timing Diagram")
	timingLabel.TextStyle = fyne.TextStyle{Bold: true}
	timingLabel.Alignment = fyne.TextAlignCenter

	centerPanel := container.NewVBox(
		signalLabel,
		ca.signalFlow,
		widget.NewSeparator(),
		timingLabel,
		ca.timingDiagram,
	)

	// Right panel: Educational + Log + Key Stat
	rightPanel := container.NewVBox(
		ca.educationalPanel,
		widget.NewSeparator(),
		ca.operationLog,
		widget.NewSeparator(),
		ca.keyStat,
	)

	// Footer with mode indicator and status
	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			ca.modeIndicator,
			widget.NewSeparator(),
			ca.statusLabel,
			layout.NewSpacer(),
			widget.NewLabel("FeCIM Ferroelectric CIM | Standard CMOS"),
		),
	)

	// Main content
	mainContent := container.NewBorder(
		header,                            // top
		footer,                            // bottom
		container.NewPadded(controlPanel), // left
		container.NewPadded(rightPanel),   // right
		container.NewPadded(centerPanel),  // center
	)

	return mainContent
}

// updateValues updates all displayed values based on current level.
func (ca *CircuitsApp) updateValues() {
	dacV := ca.dac.Convert(ca.currentLevel)
	pumpV := ca.pump.ActualOutputVoltage()
	tiaCurrent := float64(ca.currentLevel) / 29.0 * ca.tia.MaxInputCurrent
	tiaV := ca.tia.Convert(tiaCurrent)
	adcLevel := ca.adc.Convert(tiaV)

	ca.dacValueLabel.SetText(fmt.Sprintf("DAC: Level %d → %.3fV", ca.currentLevel, dacV))
	ca.pumpValueLabel.SetText(fmt.Sprintf("Pump: 1.0V → %.2fV", pumpV))
	ca.tiaValueLabel.SetText(fmt.Sprintf("TIA: %.1fµA → %.3fV", tiaCurrent*1e6, tiaV))
	ca.adcValueLabel.SetText(fmt.Sprintf("ADC: %.3fV → Level %d", tiaV, adcLevel))

	ca.signalFlow.SetValues(ca.currentLevel, dacV, tiaCurrent, tiaV, adcLevel)
}

// updateStatus updates the status label.
func (ca *CircuitsApp) updateStatus(status string) {
	ca.statusLabel.SetText(status)
}

// runDACDemo demonstrates the DAC.
func (ca *CircuitsApp) runDACDemo() {
	ca.modeIndicator.SetMode(CircuitModeDAC)
	ca.educationalPanel.SetDACExplanation()
	ca.signalFlow.SetActiveStage(1)
	ca.operationLog.Add(fmt.Sprintf("DAC: Level %d → %.3fV", ca.currentLevel, ca.dac.Convert(ca.currentLevel)))
	ca.updateStatus(fmt.Sprintf("DAC | Level %d → %.3fV (settle: %.0fns)",
		ca.currentLevel, ca.dac.Convert(ca.currentLevel), ca.dac.SettleTime))
	ca.keyStat.SetValue(fmt.Sprintf("%.0f fJ/conv", ca.dac.EnergyPerConversion()*1e15))
}

// runADCDemo demonstrates the ADC.
func (ca *CircuitsApp) runADCDemo() {
	ca.modeIndicator.SetMode(CircuitModeADC)
	ca.educationalPanel.SetADCExplanation()
	ca.signalFlow.SetActiveStage(5)
	tiaCurrent := float64(ca.currentLevel) / 29.0 * ca.tia.MaxInputCurrent
	tiaV := ca.tia.Convert(tiaCurrent)
	adcLevel := ca.adc.Convert(tiaV)
	ca.operationLog.Add(fmt.Sprintf("ADC: %.3fV → Level %d", tiaV, adcLevel))
	ca.updateStatus(fmt.Sprintf("ADC | %.3fV → Level %d (conv: %.0fns)",
		tiaV, adcLevel, ca.adc.ConversionTime))
	ca.keyStat.SetValue(fmt.Sprintf("ENOB: %.1f bits", ca.adc.ENOB()))
}

// runTIADemo demonstrates the TIA.
func (ca *CircuitsApp) runTIADemo() {
	ca.modeIndicator.SetMode(CircuitModeTIA)
	ca.educationalPanel.SetTIAExplanation()
	ca.signalFlow.SetActiveStage(4)
	tiaCurrent := float64(ca.currentLevel) / 29.0 * ca.tia.MaxInputCurrent
	tiaV := ca.tia.Convert(tiaCurrent)
	snr := ca.tia.SNR(tiaCurrent)
	ca.operationLog.Add(fmt.Sprintf("TIA: %.1fµA → %.3fV (SNR: %.1fdB)", tiaCurrent*1e6, tiaV, snr))
	ca.updateStatus(fmt.Sprintf("TIA | %.1fµA → %.3fV (gain: %.0fkΩ)",
		tiaCurrent*1e6, tiaV, ca.tia.Gain/1e3))
	ca.keyStat.SetValue(fmt.Sprintf("Gain: %.0fkΩ", ca.tia.Gain/1e3))
}

// runPumpDemo demonstrates the charge pump.
func (ca *CircuitsApp) runPumpDemo() {
	ca.modeIndicator.SetMode(CircuitModePump)
	ca.educationalPanel.SetPumpExplanation()
	ca.signalFlow.SetActiveStage(2)
	ca.operationLog.Add(fmt.Sprintf("Pump: 1.0V → %.2fV (%.0f%% eff)",
		ca.pump.ActualOutputVoltage(), ca.pump.Efficiency*100))
	ca.updateStatus(fmt.Sprintf("PUMP | 1.0V → %.2fV (rise: %.1fµs)",
		ca.pump.ActualOutputVoltage(), ca.pump.RiseTime()*1e6))
	ca.keyStat.SetValue(fmt.Sprintf("%.0f%% efficient", ca.pump.Efficiency*100))
}

// runWriteCycle demonstrates a complete write cycle.
func (ca *CircuitsApp) runWriteCycle() {
	ca.modeIndicator.SetMode(CircuitModeWrite)
	ca.timingDiagram.SetWriteCycle()
	ca.operationLog.Add("Write Cycle: Starting")

	// Phase 1: DAC
	ca.educationalPanel.SetWriteCycleExplanation(1)
	ca.signalFlow.SetActiveStage(1)
	ca.timingDiagram.SetPhase(1)
	ca.operationLog.Add(fmt.Sprintf("  1. DAC: Level %d → %.3fV", ca.currentLevel, ca.dac.Convert(ca.currentLevel)))

	// Phase 2: Pump
	ca.educationalPanel.SetWriteCycleExplanation(2)
	ca.signalFlow.SetActiveStage(2)
	ca.timingDiagram.SetPhase(2)
	ca.operationLog.Add(fmt.Sprintf("  2. Pump: 1.0V → %.2fV", ca.pump.ActualOutputVoltage()))

	// Phase 3: Cell
	ca.educationalPanel.SetWriteCycleExplanation(3)
	ca.signalFlow.SetActiveStage(3)
	ca.timingDiagram.SetPhase(3)
	ca.operationLog.Add("  3. Program: FeFET polarization set")

	// Phase 4: Verify
	ca.educationalPanel.SetWriteCycleExplanation(4)
	ca.timingDiagram.SetPhase(4)
	ca.operationLog.Add("  4. Verify: Read back OK")

	timing := peripherals.AnalyzeTiming(ca.dac, ca.adc, ca.tia, ca.pump)
	ca.updateStatus(fmt.Sprintf("WRITE | Complete: Level %d in %.0fns", ca.currentLevel, timing.WriteTime*1e9))
	ca.keyStat.SetValue(fmt.Sprintf("%.0fns write", timing.WriteTime*1e9))
	ca.modeIndicator.SetMode(CircuitModeIdle)
}

// runReadCycle demonstrates a complete read cycle.
func (ca *CircuitsApp) runReadCycle() {
	ca.modeIndicator.SetMode(CircuitModeRead)
	ca.timingDiagram.SetReadCycle()
	ca.operationLog.Add("Read Cycle: Starting")

	// Phase 1: Apply Vread
	ca.educationalPanel.SetReadCycleExplanation(1)
	ca.signalFlow.SetActiveStage(3)
	ca.timingDiagram.SetPhase(1)
	ca.operationLog.Add("  1. Apply: V_read to cell")

	// Phase 2: TIA
	tiaCurrent := float64(ca.currentLevel) / 29.0 * ca.tia.MaxInputCurrent
	tiaV := ca.tia.Convert(tiaCurrent)
	ca.educationalPanel.SetReadCycleExplanation(2)
	ca.signalFlow.SetActiveStage(4)
	ca.timingDiagram.SetPhase(2)
	ca.operationLog.Add(fmt.Sprintf("  2. TIA: %.1fµA → %.3fV", tiaCurrent*1e6, tiaV))

	// Phase 3: ADC
	adcLevel := ca.adc.Convert(tiaV)
	ca.educationalPanel.SetReadCycleExplanation(3)
	ca.signalFlow.SetActiveStage(5)
	ca.timingDiagram.SetPhase(3)
	ca.operationLog.Add(fmt.Sprintf("  3. ADC: %.3fV → Level %d", tiaV, adcLevel))

	// Phase 4: Output
	ca.educationalPanel.SetReadCycleExplanation(4)
	ca.signalFlow.SetActiveStage(6)
	ca.timingDiagram.SetPhase(4)
	ca.operationLog.Add(fmt.Sprintf("  4. Output: Level %d", adcLevel))

	timing := peripherals.AnalyzeTiming(ca.dac, ca.adc, ca.tia, ca.pump)
	ca.updateStatus(fmt.Sprintf("READ | Complete: Level %d in %.0fns", adcLevel, timing.ReadTime*1e9))
	ca.keyStat.SetValue(fmt.Sprintf("%.0fns read", timing.ReadTime*1e9))
	ca.modeIndicator.SetMode(CircuitModeIdle)
}

// onDemoModeChanged handles demo mode selection changes.
func (ca *CircuitsApp) onDemoModeChanged(mode string) {
	// Stop any existing auto demo
	ca.stopAutoDemoLoop()

	switch mode {
	case "Auto Demo":
		ca.startAutoDemoLoop()
	case "Step-by-Step":
		ca.operationLog.Add("Mode: Step-by-Step (manual)")
		ca.educationalPanel.SetContent("Step-by-Step Mode",
			"Follow the signal path:\n\n"+
				"1. Adjust level slider\n"+
				"2. Click DAC → Pump\n"+
				"3. Click TIA → ADC\n"+
				"4. Or run full Write/Read\n\n"+
				"Watch values update\n"+
				"through the chain.")
	case "Manual":
		ca.operationLog.Add("Mode: Manual")
		ca.educationalPanel.SetIdleExplanation()
	}
}

// startAutoDemoLoop starts the automatic demo loop.
func (ca *CircuitsApp) startAutoDemoLoop() {
	ca.autoDemo = true
	ca.autoDemoPhase = 0
	ca.stopAutoDemo = make(chan bool)
	ca.autoDemoTimer = time.NewTicker(2 * time.Second)

	ca.operationLog.Add("Mode: Auto Demo started")
	ca.educationalPanel.SetContent("Auto Demo Mode",
		"Watch the demo cycle\nthrough all circuits.\n\n"+
			"1. DAC conversion\n"+
			"2. Charge pump boost\n"+
			"3. TIA amplification\n"+
			"4. ADC conversion\n"+
			"5. Full Write cycle\n"+
			"6. Full Read cycle")

	go ca.autoDemoLoop()
}

// stopAutoDemoLoop stops the automatic demo loop.
func (ca *CircuitsApp) stopAutoDemoLoop() {
	if ca.autoDemo {
		ca.autoDemo = false
		if ca.stopAutoDemo != nil {
			close(ca.stopAutoDemo)
		}
		if ca.autoDemoTimer != nil {
			ca.autoDemoTimer.Stop()
		}
		ca.operationLog.Add("Mode: Auto Demo stopped")
	}
}

// autoDemoLoop runs the automatic demonstration.
func (ca *CircuitsApp) autoDemoLoop() {
	// Run first operation immediately
	fyne.Do(func() {
		ca.runAutoDemoStep()
	})

	for {
		select {
		case <-ca.stopAutoDemo:
			return
		case <-ca.autoDemoTimer.C:
			if !ca.autoDemo {
				return
			}
			fyne.Do(func() {
				ca.runAutoDemoStep()
			})
		}
	}
}

// runAutoDemoStep executes one step of the auto demo.
func (ca *CircuitsApp) runAutoDemoStep() {
	switch ca.autoDemoPhase {
	case 0:
		ca.runDACDemo()
	case 1:
		ca.runPumpDemo()
	case 2:
		ca.runTIADemo()
	case 3:
		ca.runADCDemo()
	case 4:
		ca.runWriteCycle()
	case 5:
		ca.runReadCycle()
	}
	ca.autoDemoPhase = (ca.autoDemoPhase + 1) % 6
}

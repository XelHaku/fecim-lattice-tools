// Package gui provides Fyne-based GUI components for architecture comparison.
package gui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	sharedexport "fecim-lattice-tools/shared/export"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/presets"
	sharedtheme "fecim-lattice-tools/shared/theme"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// Package-level logger using shared logging infrastructure
var debug *logging.Logger

func init() {
	debug = logging.NewLogger("comparison-app")
}

// Energy specs (model inputs) derived from docs/videos/ironlattice-youtube-script.md line 205:
// "CPU plus DRAM: 1000 picojoules. GPU plus HBM: 100 picojoules. FeCIM: under 1 picojoule."
const (
	cpuEnergyPJPerMAC   = 1000.0 // 1000 pJ/MAC
	gpuEnergyPJPerMAC   = 100.0  // 100 pJ/MAC
	fecimEnergyPJPerMAC = 1.0    // ~1 pJ/MAC (conservative for claimed "<1 pJ")
)

const (
	hoursPerMonth         = 730.0
	electricityCostPerKWh = 0.10
)

// EnergySpec holds energy per MAC model inputs with source references.
type EnergySpec struct {
	Name          string
	EnergyFJ      float64 // femtojoules per MAC (1 pJ = 1000 fJ)
	Source        string
	Verified      bool // Deprecated: keep false; all values are model inputs, not validated
	SourceDetails string
}

// ComparisonApp is the main application for architecture comparison.
type ComparisonApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Energy specs
	cpuSpec   EnergySpec
	gpuSpec   EnergySpec
	fecimSpec EnergySpec

	// Animation state (protected by animMu)
	animMu            sync.RWMutex
	animWG            sync.WaitGroup
	running           bool
	paused            bool
	foregroundVisible bool
	simTime           float64
	presentationMode  PresentationMode
	currentPhase      AutoDemoPhase
	phaseTimer        float64

	// GUI components - Hero visualizations
	energyRace        *AnimatedEnergyRace
	marketChart       *MarketOpportunityChart
	competitiveMatrix *CompetitiveMatrix
	phasedStrategy    *PhasedStrategyDiagram
	analogStates      *AnalogStatesComparison
	dcTransformation  *DataCenterTransformation

	// GUI components - Calculator
	calculator *DataCenterCalculator

	// Controls
	workloadSelect   *widget.Select
	inferencesSlider *widget.Slider
	inferencesLabel  *widget.Label

	// Status
	// IMPORTANT: All status updates MUST go through updateStatus() method to use cache
	// and prevent redundant SetText calls that bypass Fyne's internal caching.
	statusLabel *widget.Label
	statusBar   *sharedwidgets.StatusBar

	// Current settings
	currentWorkload   string
	currentInferences float64
	uiMode            string
	scenarioProfile   ScenarioProfile
	lastRun           ScenarioRun

	recomputeBus *sharedwidgets.CoalesceBus
}

// NewComparisonApp creates the comparison demo application.
func NewComparisonApp() *ComparisonApp {
	debug.Println("NewComparisonApp: Creating application")
	ca := &ComparisonApp{
		currentWorkload:   "GPT-2",
		currentInferences: 10000,
		uiMode:            "Technical Review",
		scenarioProfile:   ScenarioBaseline,
		foregroundVisible: true,
		recomputeBus:      sharedwidgets.NewCoalesceBus(40 * time.Millisecond),
	}

	ca.fyneApp = app.NewWithID("com.fecim.comparison-demo")
	ca.fyneApp.Settings().SetTheme(&sharedtheme.FeCIMTheme{})

	// Initialize energy specs (convert pJ to fJ: multiply by 1000)
	ca.cpuSpec = EnergySpec{
		Name:          "CPU + DRAM",
		EnergyFJ:      cpuEnergyPJPerMAC * 1000, // 1,000,000 fJ/MAC
		Source:        "Model input (public CPU/DRAM datasheets)",
		Verified:      false,
		SourceDetails: "Model input: includes memory access energy (~640 pJ for DRAM fetch + ~3-5 pJ for MAC).",
	}

	ca.gpuSpec = EnergySpec{
		Name:          "GPU + HBM",
		EnergyFJ:      gpuEnergyPJPerMAC * 1000, // 100,000 fJ/MAC
		Source:        "Model input (public GPU/HBM datasheets)",
		Verified:      false,
		SourceDetails: "Model input: H100 SXM 700W TDP, ~3958 TFLOPS FP16; HBM access dominates.",
	}

	ca.fecimSpec = EnergySpec{
		Name:          "FeCIM",
		EnergyFJ:      fecimEnergyPJPerMAC * 1000, // 1,000 fJ/MAC
		Source:        "Model input (COSM 2025 conference)",
		Verified:      false,
		SourceDetails: "Model input: under 1 picojoule per MAC (conference claim).",
	}

	// Register with global preset manager
	presets.Global().RegisterProvider(NewComparisonPresetProvider(ca))

	debug.Println("NewComparisonApp: Initialization complete")
	return ca
}

// Run starts the GUI application.
func (ca *ComparisonApp) Run() {
	debug.Println("App: Creating window")
	ca.window = ca.fyneApp.NewWindow("FeCIM Technical Briefing: Architecture Comparison")
	ca.window.Resize(fyne.NewSize(1400, 900))

	content := ca.createMainLayout()
	ca.window.SetContent(content)

	// Setup keyboard shortcuts
	ca.setupKeyboard()
	ca.setupVisibilityHooks()

	ca.updateCalculations()
	ca.updateStatus("Ready. Select workload and adjust parameters. Press ? for shortcuts.")

	// Animation goroutine: runs at 30 FPS until ca.running=false on window close.
	ca.animMu.Lock()
	ca.running = true
	ca.animMu.Unlock()
	ca.animWG.Add(1)
	go func() {
		defer ca.animWG.Done()
		ca.animationLoop()
	}()

	debug.Println("App: ShowAndRun starting")
	ca.window.ShowAndRun()
	ca.animMu.Lock()
	ca.running = false
	ca.animMu.Unlock()
	ca.animWG.Wait()
	if ca.recomputeBus != nil {
		ca.recomputeBus.Close()
	}
}

// animationLoop runs the main animation at 30 FPS (reduced from 60 to prevent resize loops on tiling WMs).
func (ca *ComparisonApp) animationLoop() {
	ticker := time.NewTicker(33 * time.Millisecond)
	defer ticker.Stop()

	lastTime := time.Now()

	for {
		<-ticker.C

		// Check if we should stop and update simTime atomically
		ca.animMu.Lock()
		running := ca.running
		paused := ca.paused
		visible := ca.foregroundVisible
		if !running {
			ca.animMu.Unlock()
			return
		}

		if paused || !visible {
			ca.animMu.Unlock()
			lastTime = time.Now()
			continue
		}

		dt := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		ca.simTime += dt
		ca.animMu.Unlock()

		dirtyEnergy := ca.energyRace != nil && ca.energyRace.UpdateAnimation(dt)
		dirtyMarket := ca.marketChart != nil && ca.marketChart.UpdateAnimation(dt)
		dirtyPhased := ca.phasedStrategy != nil && ca.phasedStrategy.UpdateAnimation(dt)
		dirtyAnalog := ca.analogStates != nil && ca.analogStates.UpdateAnimation(dt)
		dirtyTransform := ca.dcTransformation != nil && ca.dcTransformation.UpdateAnimation(dt)

		if !(dirtyEnergy || dirtyMarket || dirtyPhased || dirtyAnalog || dirtyTransform) {
			continue
		}

		// Refresh only dirty widgets.
		sharedwidgets.SafeDo(func() {
			if dirtyEnergy {
				ca.energyRace.Refresh()
			}
			if dirtyMarket {
				ca.marketChart.Refresh()
			}
			if dirtyPhased {
				ca.phasedStrategy.Refresh()
			}
			if dirtyAnalog {
				ca.analogStates.Refresh()
			}
			if dirtyTransform {
				ca.dcTransformation.Refresh()
			}
		})
	}
}

func (ca *ComparisonApp) setupVisibilityHooks() {
	lifecycle := ca.fyneApp.Lifecycle()
	lifecycle.SetOnEnteredForeground(func() {
		ca.animMu.Lock()
		ca.foregroundVisible = true
		ca.animMu.Unlock()
	})
	lifecycle.SetOnExitedForeground(func() {
		ca.animMu.Lock()
		ca.foregroundVisible = false
		ca.animMu.Unlock()
	})
}

// createMainLayout builds the main application layout.
func (ca *ComparisonApp) createMainLayout() fyne.CanvasObject {
	// Create widgets
	ca.energyRace = NewAnimatedEnergyRace()
	ca.marketChart = NewMarketOpportunityChart()
	ca.competitiveMatrix = NewCompetitiveMatrix()
	ca.phasedStrategy = NewPhasedStrategyDiagram()
	ca.analogStates = NewAnalogStatesComparison()
	ca.dcTransformation = NewDataCenterTransformation()
	ca.calculator = NewDataCenterCalculator()

	// Workload selector - default to GPT-2 for illustrative savings display
	ca.workloadSelect = widget.NewSelect(
		[]string{"MNIST", "ResNet-50", "BERT-Base", "GPT-2", "LLM-70B"},
		ca.onWorkloadChanged,
	)
	ca.workloadSelect.SetSelected("GPT-2")

	uiModeSelect := widget.NewSelect([]string{"Technical Review", "Presentation"}, func(v string) {
		if v == "" {
			return
		}
		ca.uiMode = v
		ca.scheduleRecompute()
	})
	uiModeSelect.SetSelected(ca.uiMode)

	scenarioSelect := widget.NewSelect([]string{string(ScenarioConservative), string(ScenarioBaseline), string(ScenarioOptimistic)}, func(v string) {
		ca.scenarioProfile = ScenarioProfileFromString(v)
		ca.scheduleRecompute()
	})
	scenarioSelect.SetSelected(string(ca.scenarioProfile))

	// Inferences slider - wider and reasonable range
	ca.inferencesLabel = widget.NewLabel("Inferences/sec: 10,000")
	throttled := sharedwidgets.NewThrottledSlider(1000, 50000, 40*time.Millisecond,
		func(v float64) {
			ca.currentInferences = v
			ca.inferencesLabel.SetText(fmt.Sprintf("Inferences/sec: %.0f", v))
		},
		func(v float64) {
			ca.currentInferences = v
			ca.inferencesLabel.SetText(fmt.Sprintf("Inferences/sec: %.0f", v))
			ca.scheduleRecompute()
		},
	)
	ca.inferencesSlider = throttled.Slider
	ca.inferencesSlider.Value = 10000
	ca.inferencesSlider.Step = 1000

	// Calculate button
	var calcBtn *widget.Button
	calcBtn = widget.NewButton("Calculate", func() {
		calcBtn.Disable()
		calcBtn.SetText("Calculating...")
		// Calculation goroutine: runs once per button click; exits after updateCalculations completes.
		go func() {
			ca.updateCalculations()
			sharedwidgets.SafeDo(func() {
				calcBtn.Enable()
				calcBtn.SetText("Calculate")
			})
		}()
	})
	calcBtn.Importance = widget.HighImportance

	// Status
	ca.statusLabel = widget.NewLabel("Status: Ready")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	ca.statusLabel.Truncation = fyne.TextTruncateEllipsis
	ca.statusBar = sharedwidgets.NewStatusBarWithLabel(ca.statusLabel, "Status: ")

	// === SIMULATION WARNING BANNER ===
	// CRITICAL: Per Dr. Tour critique - must prominently display TRL status
	warningBanner := widget.NewLabelWithStyle(
		"SIMULATION ONLY - NOT VALIDATED | TRL 4 Lab Prototype | All values are model inputs",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	warningBanner.Wrapping = fyne.TextWrapWord

	// === HEADER (title moved to main navbar) ===
	resetBtn := widget.NewButton("Reset", func() {
		if ca.energyRace != nil {
			ca.energyRace.Reset()
		}
		if ca.marketChart != nil {
			ca.marketChart.Reset()
		}
		if ca.phasedStrategy != nil {
			ca.phasedStrategy.Reset()
		}
		ca.updateStatus("Animation reset")
	})

	header := container.NewHBox(
		widget.NewLabel("Dr. external research group | COSM 2025"),
		layout.NewSpacer(),
		resetBtn,
	)

	// === UNIFIED INVESTOR PITCH VIEW ===
	// All sections in one scrollable view for seamless presentation

	// SECTION 1: THE ENERGY PROBLEM
	sectionEnergyHeader := widget.NewLabelWithStyle(
		"THE ENERGY PROBLEM",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	energySection := container.NewVBox(
		sectionEnergyHeader,
		container.NewPadded(container.NewPadded(ca.energyRace)),
		ca.createPlainTextEvidencePanel("Energy comparison visualization"),
		ca.createEvidenceFirstPanel(),
	)

	// SECTION 2: MARKET OPPORTUNITY
	sectionMarketHeader := widget.NewLabelWithStyle(
		"MARKET OPPORTUNITY",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	marketSection := container.NewVBox(
		sectionMarketHeader,
		container.NewPadded(ca.marketChart),
		ca.createPlainTextEvidencePanel("Market opportunity chart"),
		widget.NewCard(
			"Phased Entry Strategy",
			"De-risking through staged market entry",
			container.NewPadded(ca.phasedStrategy),
		),
		ca.createPlainTextEvidencePanel("Phased entry strategy visualization"),
		container.NewPadded(ca.competitiveMatrix),
		ca.createPlainTextEvidencePanel("Competitive matrix comparison"),
	)

	// SECTION 3: ROI CALCULATOR
	sectionROIHeader := widget.NewLabelWithStyle(
		"ROI CALCULATOR",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	// Wrap slider in fixed-width container for better visibility
	sliderContainer := container.NewGridWrap(fyne.NewSize(200, 30), ca.inferencesSlider)

	// Export buttons
	exportDataBtn := sharedexport.CreateExportButton("Export Data", func() {
		ca.exportComparisonData()
	}, ca.window)
	exportImageBtn := sharedexport.CreateExportButton("Save Image", func() {
		ca.exportVisualization()
	}, ca.window)
	exportReproBtn := sharedexport.CreateExportButton("Export Repro Pack", func() {
		ca.exportReproducibilityPack()
	}, ca.window)

	// Use a wrapping grid so controls don't overflow at 1024px minimum width.
	// Row 1: mode/scenario/workload selectors + calculate
	controlsRow1 := container.NewHBox(
		widget.NewLabel("Mode:"),
		uiModeSelect,
		widget.NewLabel("Scenario:"),
		scenarioSelect,
		widget.NewLabel("Workload:"),
		ca.workloadSelect,
		layout.NewSpacer(),
		calcBtn,
	)
	// Row 2: slider label + slider + export buttons
	ca.inferencesLabel.Truncation = fyne.TextTruncateEllipsis
	controlsRow2 := container.NewHBox(
		ca.inferencesLabel,
		sliderContainer,
		layout.NewSpacer(),
		exportDataBtn,
		exportImageBtn,
		exportReproBtn,
	)
	configRow := container.NewVBox(controlsRow1, controlsRow2)
	roiSection := container.NewVBox(
		sectionROIHeader,
		container.NewPadded(configRow),
		container.NewPadded(ca.calculator),
		ca.createPlainTextEvidencePanel("ROI and calculator panel"),
		ca.createScenarioDiffPanel(),
	)

	// SECTION 4: FABRICATION REALITY (H08)
	// Per Dr. Tour critique - show honest development expectations
	fabricationReality := NewFabricationReality()
	sectionFabHeader := widget.NewLabelWithStyle(
		"FABRICATION REALITY",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	fabSection := container.NewVBox(
		sectionFabHeader,
		container.NewPadded(fabricationReality),
	)

	// UNIFIED VIEW - Single scrollable container
	unifiedContent := container.NewVBox(
		energySection,
		widget.NewSeparator(),
		marketSection,
		widget.NewSeparator(),
		roiSection,
		widget.NewSeparator(),
		fabSection,
	)

	centerPanel := container.NewScroll(unifiedContent)

	// === FOOTER ===
	footer := container.NewHBox(
		ca.statusLabel,
		layout.NewSpacer(),
	)

	// === MAIN LAYOUT ===
	// No side panels - tabs are full width, config is in Calculator tab
	centerContainer := container.NewPadded(centerPanel)

	mainContent := container.NewBorder(
		container.NewVBox(warningBanner, header, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), footer),
		nil, // No left panel - config is in Calculator tab
		nil, // No right panel
		centerContainer,
	)

	return mainContent
}

func (ca *ComparisonApp) createEvidenceFirstPanel() fyne.CanvasObject {
	run := ca.lastRun
	if len(run.Inputs) == 0 {
		run = BuildScenarioRun(ca.scenarioProfile, PresetScenarioConfig(ca.scenarioProfile))
	}

	if ca.uiMode == "Presentation" {
		summary := widget.NewLabel(fmt.Sprintf("Scenario: %s | Energy %.1f%% | Latency %.1f%% | TCO %.1f%% | CO2 %.1f%%",
			run.Profile,
			run.Outputs["energy_reduction_pct"].Value,
			run.Outputs["latency_reduction_pct"].Value,
			run.Outputs["tco_reduction_pct"].Value,
			run.Outputs["co2_reduction_pct"].Value,
		))
		caveat := widget.NewLabel("Caveat: Presentation mode simplifies outputs. Switch to Technical Review for raw assumptions and uncertainty.")
		return widget.NewCard("Evidence Summary", "Presentation", container.NewVBox(summary, caveat))
	}

	assumptions := widget.NewLabel(PlainTextEvidence("Assumptions & outputs", run))
	assumptions.Wrapping = fyne.TextWrapWord
	sensitivity := SensitivityRanking(run, "tco_reduction_pct")
	top := "n/a"
	if len(sensitivity) > 0 {
		top = fmt.Sprintf("Top sensitivity: %s impact %.2f%%", sensitivity[0].InputName, sensitivity[0].ImpactPct)
	}
	caveat := widget.NewLabel("Caveat: all values are model inputs/derived estimates (TRL4).")
	return widget.NewCard("Evidence-first panel", "Assumptions + outputs + caveat", container.NewVBox(assumptions, widget.NewLabel(top), caveat))
}

func (ca *ComparisonApp) createPlainTextEvidencePanel(title string) fyne.CanvasObject {
	run := ca.lastRun
	if len(run.Inputs) == 0 {
		run = BuildScenarioRun(ca.scenarioProfile, PresetScenarioConfig(ca.scenarioProfile))
	}
	text := widget.NewLabel(PlainTextEvidence(title, run))
	text.Wrapping = fyne.TextWrapWord
	// Wrap in a fixed-height scroll so the raw parameter dump doesn't push
	// content below out of view or overlap content above at narrow widths.
	scrolled := container.NewScroll(text)
	scrolled.SetMinSize(fyne.NewSize(0, 120))
	return widget.NewCard("Plain-text evidence", "Screen-reader-first", scrolled)
}

func (ca *ComparisonApp) createScenarioDiffPanel() fyne.CanvasObject {
	baseline := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	current := ca.lastRun
	if len(current.Inputs) == 0 {
		current = BuildScenarioRun(ca.scenarioProfile, PresetScenarioConfig(ca.scenarioProfile))
	}
	diff := DiffScenarios(baseline, current)
	content := widget.NewLabel(fmt.Sprintf("Changed assumptions: %d\nOutput deltas: %v\nAttribution: %v", len(diff.ChangedAssumptions), diff.OutputDeltas, diff.Attribution))
	content.Wrapping = fyne.TextWrapWord
	return widget.NewCard("Scenario Diff (Run A vs Run B)", "A=baseline, B=current", content)
}

// createVerifiedClaimsWidget creates a compact model input / scenario input section.
// Dr. Tour recommendation: Show explicit energy numbers with units and citations (as model inputs)
func (ca *ComparisonApp) createVerifiedClaimsWidget() fyne.CanvasObject {
	verifiedLabel := widget.NewLabelWithStyle("MODEL INPUT REFERENCES:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	verifiedItems := widget.NewLabel("• Analog levels (literature ranges; not validated here)\n• MNIST accuracy (literature ranges; not validated here)\n• CMOS compatibility (assumed)")

	// Explicit energy numbers with units (Dr. Tour recommendation)
	energyLabel := widget.NewLabelWithStyle("ENERGY/MAC (pJ, MODEL INPUTS):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	energyItems := widget.NewLabel(fmt.Sprintf(
		"• CPU+DRAM: %d pJ (model input)\n• GPU+HBM: %d pJ (model input)\n• FeCIM: ~%.1f pJ (model input; TRL 4)",
		int(cpuEnergyPJPerMAC), int(gpuEnergyPJPerMAC), fecimEnergyPJPerMAC))

	claimedLabel := widget.NewLabelWithStyle("SCENARIO INPUTS (NOT VALIDATED):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	claimedItems := widget.NewLabel("• 30 analog levels (conference claim)\n• 25-100× vs NAND (scenario input)\n• 1000× vs DRAM (scenario input)\n• 80-90% DC savings (scenario input)")

	trlLabel := widget.NewLabelWithStyle("Status: TRL 4 (Lab only) — model inputs only", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true})

	return container.NewVBox(
		verifiedLabel,
		verifiedItems,
		widget.NewSeparator(),
		energyLabel,
		energyItems,
		widget.NewSeparator(),
		claimedLabel,
		claimedItems,
		widget.NewSeparator(),
		trlLabel,
	)
}

// onWorkloadChanged handles workload selection.
func (ca *ComparisonApp) onWorkloadChanged(workload string) {
	ca.currentWorkload = workload
	ca.scheduleRecompute()
}

func (ca *ComparisonApp) scheduleRecompute() {
	if ca.recomputeBus == nil {
		ca.updateCalculations()
		return
	}
	ca.recomputeBus.Submit("module5-calculation", ca.updateCalculations)
}

func energyPowerCost(macs int, energyFJ, inferences float64) (energyUJ, powerW, monthlyCost float64) {
	energyUJ = float64(macs) * energyFJ / 1e9
	powerW = energyUJ * inferences / 1e6
	monthlyCost = powerW / 1000 * hoursPerMonth * electricityCostPerKWh
	return energyUJ, powerW, monthlyCost
}

// updateCalculations recalculates all values.
func (ca *ComparisonApp) updateCalculations() {
	debug.Printf("updateCalculations: workload=%s, inferences=%.0f", ca.currentWorkload, ca.currentInferences)

	macs := ca.getWorkloadMACs()

	cfg := PresetScenarioConfig(ca.scenarioProfile)
	ca.lastRun = BuildScenarioRun(ca.scenarioProfile, cfg)
	ca.cpuSpec.EnergyFJ = cfg.CPUEnergyPJPerMAC * 1000
	ca.gpuSpec.EnergyFJ = cfg.GPUEnergyPJPerMAC * 1000
	ca.fecimSpec.EnergyFJ = cfg.FeCIMEnergyPJPerMAC * 1000

	cpuEnergy, cpuPower, cpuCost := energyPowerCost(macs, ca.cpuSpec.EnergyFJ, ca.currentInferences)
	gpuEnergy, gpuPower, gpuCost := energyPowerCost(macs, ca.gpuSpec.EnergyFJ, ca.currentInferences)
	fecimEnergy, fecimPower, fecimCost := energyPowerCost(macs, ca.fecimSpec.EnergyFJ, ca.currentInferences)

	// Update calculator
	ca.calculator.SetResults(
		ca.currentWorkload, macs, ca.currentInferences,
		cpuEnergy, gpuEnergy, fecimEnergy,
		cpuPower, gpuPower, fecimPower,
		cpuCost, gpuCost, fecimCost,
	)

	// Update transformation widget
	ca.dcTransformation.SetValues(gpuPower, fecimPower)

	ca.updateStatus(fmt.Sprintf("Calculated for %s @ %.0f inf/s | mode=%s | scenario=%s", ca.currentWorkload, ca.currentInferences, ca.uiMode, ca.scenarioProfile))
}

// getWorkloadMACs returns MACs per inference for common neural network workloads.
// Model inputs derived from typical architecture descriptions (not validated).
func (ca *ComparisonApp) getWorkloadMACs() int {
	switch ca.currentWorkload {
	case "MNIST":
		// Simple 2-layer MLP: 784 input → 128 hidden → 10 output
		return 101632 // (784×128) + (128×10)
	case "ResNet-50":
		// Deep residual network for image classification
		return 4000000000 // ~4 GMACs
	case "BERT-Base":
		// Transformer for NLP (sequence length 512)
		return 11000000000 // ~11 GMACs
	case "GPT-2":
		// Large language model (117M parameters)
		return 35000000000 // ~35 GMACs
	case "LLM-70B":
		// Llama-2-70B or similar large model
		return 140000000000000 // ~140 TMACs
	default:
		return 101632
	}
}

// updateStatus updates the status label using cache to prevent redundant SetText calls.
// All status updates should go through this method to avoid bypassing Fyne's internal cache.
func (ca *ComparisonApp) updateStatus(status string) {
	sharedwidgets.EnsureStatusBar(&ca.statusBar, ca.statusLabel, "Status: ", status)
}

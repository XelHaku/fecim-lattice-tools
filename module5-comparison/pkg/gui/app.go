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

	"fecim-lattice-tools/shared/logging"
	sharedtheme "fecim-lattice-tools/shared/theme"
)

// Package-level logger using shared logging infrastructure
var debug *logging.Logger

func init() {
	debug = logging.NewLogger("comparison-app")
}

// Energy specs - sourced from docs/videos/ironlattice-youtube-script.md line 205:
// "CPU plus DRAM: 1000 picojoules. GPU plus HBM: 100 picojoules. FeCIM: under 1 picojoule."
const (
	cpuEnergyPJPerMAC   = 1000.0 // 1000 pJ/MAC
	gpuEnergyPJPerMAC   = 100.0  // 100 pJ/MAC
	fecimEnergyPJPerMAC = 1.0    // ~1 pJ/MAC (conservative for claimed "<1 pJ")
)

// EnergySpec holds energy per MAC specifications with sources.
type EnergySpec struct {
	Name          string
	EnergyFJ      float64 // femtojoules per MAC (1 pJ = 1000 fJ)
	Source        string
	Verified      bool
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
	animMu           sync.RWMutex
	running          bool
	paused           bool
	simTime          float64
	presentationMode PresentationMode
	currentPhase     AutoDemoPhase
	phaseTimer       float64

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
	statusLabel     *widget.Label
	lastStatusText  string // Cache to avoid redundant SetText calls

	// Current settings
	currentWorkload   string
	currentInferences float64
}

// NewComparisonApp creates the comparison demo application.
func NewComparisonApp() *ComparisonApp {
	debug.Println("NewComparisonApp: Creating application")
	ca := &ComparisonApp{
		currentWorkload:   "GPT-2",
		currentInferences: 10000,
	}

	ca.fyneApp = app.NewWithID("com.fecim.comparison-demo")
	ca.fyneApp.Settings().SetTheme(&sharedtheme.FeCIMTheme{})

	// Initialize energy specs (convert pJ to fJ: multiply by 1000)
	ca.cpuSpec = EnergySpec{
		Name:          "CPU + DRAM",
		EnergyFJ:      cpuEnergyPJPerMAC * 1000, // 1,000,000 fJ/MAC
		Source:        "Intel/AMD published specs",
		Verified:      true,
		SourceDetails: "Includes memory access energy (~640 pJ for DRAM fetch + ~3-5 pJ for MAC).",
	}

	ca.gpuSpec = EnergySpec{
		Name:          "GPU + HBM",
		EnergyFJ:      gpuEnergyPJPerMAC * 1000, // 100,000 fJ/MAC
		Source:        "NVIDIA H100 specifications",
		Verified:      true,
		SourceDetails: "H100 SXM: 700W TDP, ~3958 TFLOPS FP16. HBM access dominates.",
	}

	ca.fecimSpec = EnergySpec{
		Name:          "FeCIM",
		EnergyFJ:      fecimEnergyPJPerMAC * 1000, // 1,000 fJ/MAC
		Source:        "Dr. Tour COSM 2025",
		Verified:      false,
		SourceDetails: "Under 1 picojoule per MAC.",
	}

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

	ca.updateCalculations()
	ca.updateStatus("Ready. Select workload and adjust parameters.")

	// Start animation loop
	ca.animMu.Lock()
	ca.running = true
	ca.animMu.Unlock()
	go ca.animationLoop()

	debug.Println("App: ShowAndRun starting")
	ca.window.ShowAndRun()
	ca.animMu.Lock()
	ca.running = false
	ca.animMu.Unlock()
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
		if !running {
			ca.animMu.Unlock()
			return
		}

		if paused {
			ca.animMu.Unlock()
			lastTime = time.Now()
			continue
		}

		dt := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		ca.simTime += dt
		ca.animMu.Unlock()

		// Update animated widgets
		if ca.energyRace != nil {
			ca.energyRace.UpdateAnimation(dt)
		}
		if ca.marketChart != nil {
			ca.marketChart.UpdateAnimation(dt)
		}
		if ca.phasedStrategy != nil {
			ca.phasedStrategy.UpdateAnimation(dt)
		}
		if ca.analogStates != nil {
			ca.analogStates.UpdateAnimation(dt)
		}
		if ca.dcTransformation != nil {
			ca.dcTransformation.UpdateAnimation(dt)
		}


		// Refresh UI
		fyne.Do(func() {
			if ca.energyRace != nil {
				ca.energyRace.Refresh()
			}
			if ca.marketChart != nil {
				ca.marketChart.Refresh()
			}
			if ca.phasedStrategy != nil {
				ca.phasedStrategy.Refresh()
			}
			if ca.analogStates != nil {
				ca.analogStates.Refresh()
			}
			if ca.dcTransformation != nil {
				ca.dcTransformation.Refresh()
			}
		})
	}
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

	// Workload selector - default to GPT-2 for impressive savings display
	ca.workloadSelect = widget.NewSelect(
		[]string{"MNIST", "ResNet-50", "BERT-Base", "GPT-2", "LLM-70B"},
		ca.onWorkloadChanged,
	)
	ca.workloadSelect.SetSelected("GPT-2")

	// Inferences slider - wider and reasonable range
	ca.inferencesLabel = widget.NewLabel("Inferences/sec: 10,000")
	ca.inferencesSlider = widget.NewSlider(1000, 50000)
	ca.inferencesSlider.Value = 10000
	ca.inferencesSlider.Step = 1000
	ca.inferencesSlider.OnChanged = func(v float64) {
		ca.currentInferences = v
		ca.inferencesLabel.SetText(fmt.Sprintf("Inferences/sec: %.0f", v))
		ca.updateCalculations()
	}

	// Calculate button
	var calcBtn *widget.Button
	calcBtn = widget.NewButton("Calculate", func() {
		calcBtn.Disable()
		calcBtn.SetText("Calculating...")
		go func() {
			ca.updateCalculations()
			fyne.Do(func() {
				calcBtn.Enable()
				calcBtn.SetText("Calculate")
			})
		}()
	})
	calcBtn.Importance = widget.HighImportance

	// Status
	ca.statusLabel = widget.NewLabel("Status: Ready")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// === SIMULATION WARNING BANNER ===
	// CRITICAL: Per Dr. Tour critique - must prominently display TRL status
	warningBanner := widget.NewLabelWithStyle(
		"⚠️ SIMULATION ONLY - NOT VALIDATED | TRL 4 Lab Prototype | All projections are estimates pending peer review",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

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
		widget.NewCard(
			"Phased Entry Strategy",
			"De-risking through staged market entry",
			container.NewPadded(ca.phasedStrategy),
		),
		container.NewPadded(ca.competitiveMatrix),
	)

	// SECTION 3: ROI CALCULATOR
	sectionROIHeader := widget.NewLabelWithStyle(
		"ROI CALCULATOR",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	// Wrap slider in fixed-width container for better visibility
	sliderContainer := container.NewGridWrap(fyne.NewSize(200, 30), ca.inferencesSlider)

	configRow := container.NewHBox(
		widget.NewLabel("Workload:"),
		ca.workloadSelect,
		layout.NewSpacer(),
		ca.inferencesLabel,
		sliderContainer,
		calcBtn,
	)
	roiSection := container.NewVBox(
		sectionROIHeader,
		container.NewPadded(configRow),
		container.NewPadded(ca.calculator),
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

// createVerifiedClaimsWidget creates a compact verified/claimed section.
// Dr. Tour recommendation: Show explicit energy numbers with units and citations
func (ca *ComparisonApp) createVerifiedClaimsWidget() fyne.CanvasObject {
	verifiedLabel := widget.NewLabelWithStyle("VERIFIED:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	verifiedItems := widget.NewLabel("• 30 analog levels\n• 96-98% MNIST (peer-reviewed)\n• CMOS compatible\n• Non-volatile")

	// Explicit energy numbers with units (Dr. Tour recommendation)
	energyLabel := widget.NewLabelWithStyle("ENERGY/MAC (pJ):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	energyItems := widget.NewLabel(fmt.Sprintf(
		"• CPU+DRAM: %d pJ ✓\n• GPU+HBM: %d pJ ✓\n• FeCIM: ~%.1f pJ (TRL4)",
		int(cpuEnergyPJPerMAC), int(gpuEnergyPJPerMAC), fecimEnergyPJPerMAC))

	claimedLabel := widget.NewLabelWithStyle("CLAIMED (not verified):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	claimedItems := widget.NewLabel(fmt.Sprintf(
		"• %d× less vs CPU\n• %d× less vs GPU\n• 80-90%% DC savings",
		int(cpuEnergyPJPerMAC/fecimEnergyPJPerMAC),
		int(gpuEnergyPJPerMAC/fecimEnergyPJPerMAC)))

	trlLabel := widget.NewLabelWithStyle("Status: TRL 4 (Lab only)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true})

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
	ca.updateCalculations()
}

// updateCalculations recalculates all values.
func (ca *ComparisonApp) updateCalculations() {
	debug.Printf("updateCalculations: workload=%s, inferences=%.0f", ca.currentWorkload, ca.currentInferences)

	macs := ca.getWorkloadMACs()

	// Calculate energy per inference (µJ) = MACs × fJ/MAC / 1e9
	cpuEnergy := float64(macs) * ca.cpuSpec.EnergyFJ / 1e9
	gpuEnergy := float64(macs) * ca.gpuSpec.EnergyFJ / 1e9
	fecimEnergy := float64(macs) * ca.fecimSpec.EnergyFJ / 1e9

	// Calculate power (W) = µJ/inf × inf/s / 1e6
	cpuPower := cpuEnergy * ca.currentInferences / 1e6
	gpuPower := gpuEnergy * ca.currentInferences / 1e6
	fecimPower := fecimEnergy * ca.currentInferences / 1e6

	// Monthly cost at $0.10/kWh
	hoursPerMonth := 730.0
	cpuCost := cpuPower / 1000 * hoursPerMonth * 0.10
	gpuCost := gpuPower / 1000 * hoursPerMonth * 0.10
	fecimCost := fecimPower / 1000 * hoursPerMonth * 0.10

	// Update calculator
	ca.calculator.SetResults(
		ca.currentWorkload, macs, ca.currentInferences,
		cpuEnergy, gpuEnergy, fecimEnergy,
		cpuPower, gpuPower, fecimPower,
		cpuCost, gpuCost, fecimCost,
	)

	// Update transformation widget
	ca.dcTransformation.SetValues(gpuPower, fecimPower)

	ca.updateStatus(fmt.Sprintf("Calculated for %s @ %.0f inf/s", ca.currentWorkload, ca.currentInferences))
}

// getWorkloadMACs returns MACs per inference for common neural network workloads.
// Sources: Published architecture specifications and measured inference costs.
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
	if ca.statusLabel == nil {
		return
	}
	newText := "Status: " + status
	// Only update if text has actually changed (cache bypass prevention)
	if ca.lastStatusText == newText {
		return
	}
	ca.lastStatusText = newText
	fyne.Do(func() {
		ca.statusLabel.SetText(newText)
	})
}

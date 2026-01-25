// Package gui provides Fyne-based GUI components for MNIST visualization.
// tour.go implements the Guided Tour mode for educational demonstrations.
package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// TourStep represents a single step in the guided tour.
type TourStep struct {
	Title       string
	Description string
	Action      func() // Action to perform when step starts
	Duration    time.Duration
}

// GuidedTour manages the 7-step educational tour.
type GuidedTour struct {
	app          *DualModeApp
	currentStep  int
	steps        []TourStep
	isRunning    bool
	stopChan     chan bool
	statusLabel  *widget.Label
	progressBar  *widget.ProgressBar
	nextButton   *widget.Button
	stopButton   *widget.Button
	tourDialog   dialog.Dialog
}

// NewGuidedTour creates a new guided tour for the dual-mode app.
func NewGuidedTour(app *DualModeApp) *GuidedTour {
	gt := &GuidedTour{
		app:         app,
		currentStep: 0,
		stopChan:    make(chan bool, 1),
	}

	gt.steps = []TourStep{
		{
			Title:       "Step 1/5: Welcome to FeCIM",
			Description: "This neural network classifies handwritten digits using ferroelectric compute-in-memory (FeCIM).\n\nKey Innovation: 30 analog conductance levels per cell (~4.9 bits/cell)\n\nPhysics: HfO₂-ZrO₂ (HZO) superlattice ferroelectric exhibits ~30 stable polarization states due to domain wall pinning at crystal defects.\n\nTarget: 87% accuracy with 10,000x less energy than GPU!\n\nLet's see it in action...",
			Action:      func() { gt.app.applyPreset(30, 0.01, 8, 8) },
			Duration:    4 * time.Second,
		},
		{
			Title:       "Step 2/5: Draw & Classify",
			Description: "Loading a test digit from the MNIST dataset...\n\nWatch both inference paths run simultaneously:\n• FP (Float32): Ideal digital computation baseline\n• CIM (30 Levels): Real FeCIM hardware simulation\n\nBoth should predict the SAME digit when hardware is ideal!\n\nPhysics: Each crossbar cell stores a weight as polarization state. Matrix-vector multiplication happens in ONE analog step via Kirchhoff's Current Law.",
			Action: func() {
				gt.app.loadRandomSample()
			},
			Duration: 4 * time.Second,
		},
		{
			Title:       "Step 3/5: BREAK IT - Binary Weights",
			Description: "Now watch what happens with only 2 levels (binary, like SRAM)...\n\nAccuracy COLLAPSES to ~50% (worse than random guessing!)\n\nPhysics: Binary weights {-1, +1} create severe quantization error. The 128-dimensional weight space collapses to 2 discrete values, destroying the network's ability to represent learned features.\n\nLook at the weight heatmap: only blue and red.\n\nThis is why 30 analog levels are critical!",
			Action: func() {
				gt.app.applyPreset(2, 0.01, 8, 8)
				time.Sleep(500 * time.Millisecond)
				gt.app.runQuickTest()
			},
			Duration: 5 * time.Second,
		},
		{
			Title:       "Step 4/5: BREAK IT - High Noise",
			Description: "What about noise? Real ferroelectric hardware faces:\n• Thermal (Johnson) noise in sense amplifiers\n• Device-to-device variation (~2.75%)\n• Cycle-to-cycle retention drift (~1.5%)\n• Temperature-dependent polarization fluctuations\n\nWith HIGH noise (15% std dev), accuracy drops to ~70%.\n\nPhysics: Gaussian noise corrupts the analog current during matrix-vector multiply, causing the ADC to read incorrect values. The '8' is misclassified as '3'.\n\nNoise mitigation is crucial!",
			Action: func() {
				gt.app.applyPreset(30, 0.15, 6, 8)
				time.Sleep(500 * time.Millisecond)
				gt.app.runQuickTest()
			},
			Duration: 5 * time.Second,
		},
		{
			Title:       "Step 5/5: FeCIM Sweet Spot",
			Description: "Restoring optimal settings...\n\n30 levels + 1% noise = 87% accuracy!\n\nThe optimal operating point:\n• 30 levels: Enough precision to represent the network\n• 1% noise: Low enough to be manufacturable at 28nm\n• 50 fJ/MAC: 10,000x more energy-efficient than GPU DRAM access (500 pJ/MAC)\n\nPhysics: This sweet spot balances quantization error vs. analog noise, achieving near-digital accuracy with analog efficiency.\n\nExplore the presets to learn more!",
			Action: func() {
				gt.app.applyPreset(30, 0.01, 8, 8)
				time.Sleep(500 * time.Millisecond)
				gt.app.runQuickTest()
			},
			Duration: 5 * time.Second,
		},
	}

	return gt
}

// Start begins the guided tour.
func (gt *GuidedTour) Start() {
	if gt.isRunning {
		return
	}

	gt.isRunning = true
	gt.currentStep = 0

	// Create tour dialog content
	gt.statusLabel = widget.NewLabel("")
	gt.statusLabel.Wrapping = fyne.TextWrapWord

	gt.progressBar = widget.NewProgressBar()
	gt.progressBar.Min = 0
	gt.progressBar.Max = float64(len(gt.steps))

	gt.nextButton = widget.NewButton("Next Step", func() {
		gt.nextStep()
	})

	gt.stopButton = widget.NewButton("End Tour", func() {
		gt.Stop()
	})

	content := container.NewVBox(
		gt.statusLabel,
		widget.NewSeparator(),
		gt.progressBar,
		container.NewHBox(gt.nextButton, gt.stopButton),
	)

	gt.tourDialog = dialog.NewCustomWithoutButtons("Guided Tour", content, gt.app.window)
	gt.tourDialog.Resize(fyne.NewSize(500, 300))
	gt.tourDialog.Show()

	// Start first step
	gt.showCurrentStep()
}

// Stop ends the guided tour.
func (gt *GuidedTour) Stop() {
	gt.isRunning = false
	select {
	case gt.stopChan <- true:
	default:
	}

	fyne.Do(func() {
		if gt.tourDialog != nil {
			gt.tourDialog.Hide()
		}
	})

	// Reset to ideal settings
	gt.app.applyPreset(30, 0.01, 8, 8)
	fyne.Do(func() {
		gt.app.statusLabel.SetText("Tour ended. Ready to explore!")
	})
}

// nextStep advances to the next step.
func (gt *GuidedTour) nextStep() {
	gt.currentStep++
	if gt.currentStep >= len(gt.steps) {
		gt.finishTour()
		return
	}
	gt.showCurrentStep()
}

// showCurrentStep displays the current step.
func (gt *GuidedTour) showCurrentStep() {
	if gt.currentStep >= len(gt.steps) {
		return
	}

	step := gt.steps[gt.currentStep]

	// Update UI
	fyne.Do(func() {
		gt.statusLabel.SetText(fmt.Sprintf("%s\n\n%s", step.Title, step.Description))
		gt.progressBar.SetValue(float64(gt.currentStep + 1))

		if gt.currentStep == len(gt.steps)-1 {
			gt.nextButton.SetText("Finish")
		} else {
			gt.nextButton.SetText("Next Step")
		}
	})

	// Execute step action
	go func() {
		step.Action()
	}()
}

// finishTour completes the tour with a summary.
func (gt *GuidedTour) finishTour() {
	gt.isRunning = false

	fyne.Do(func() {
		if gt.tourDialog != nil {
			gt.tourDialog.Hide()
		}

		// Show completion dialog with key insights
		completionContent := container.NewVBox(
			widget.NewLabelWithStyle("Tour Complete!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			widget.NewLabelWithStyle("Key Physics Insights:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(""),
			widget.NewLabel("1. 30 analog levels (HZO ferroelectric) → 87% accuracy"),
			widget.NewLabel("2. Binary (2 levels, like SRAM) → 50% accuracy (fails!)"),
			widget.NewLabel("3. High noise (15% σ/μ) → 70% accuracy (degraded)"),
			widget.NewLabel("4. Energy: 50 fJ/MAC vs 500 pJ/MAC (GPU) = 10,000x better"),
			widget.NewLabel(""),
			widget.NewSeparator(),
			widget.NewLabelWithStyle("Next Steps:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("• Try the 'Tour' preset (Dr. Tour's COSM 2025 architecture)"),
			widget.NewLabel("• Click 'Why 30?' to understand the physics"),
			widget.NewLabel("• Click 'HW Reality' to see manufacturing constraints"),
			widget.NewLabel("• Explore failure modes: QuantCliff, Noisy, BrokenADC"),
		)

		completionDialog := dialog.NewCustom("Tour Complete!", "Start Exploring", completionContent, gt.app.window)
		completionDialog.Show()
	})

	// Reset to ideal settings
	gt.app.applyPreset(30, 0.01, 8, 8)
	fyne.Do(func() {
		gt.app.statusLabel.SetText("Tour complete! Now explore: Ideal (30-level baseline) | QuantCliff (binary failure) | Noisy (15% noise) | BrokenADC (3-bit) | Tour (Dr. Tour's architecture)")
	})
}

// IsRunning returns whether the tour is currently running.
func (gt *GuidedTour) IsRunning() bool {
	return gt.isRunning
}

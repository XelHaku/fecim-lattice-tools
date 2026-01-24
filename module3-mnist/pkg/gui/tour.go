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
			Description: "This neural network classifies handwritten digits using ferroelectric compute-in-memory.\n\nKey Innovation: 30 analog conductance levels per cell (~5 bits/cell)\n\nTarget: 87% accuracy with 10,000x less energy than a GPU!\n\nLet's see it in action...",
			Action:      func() { gt.app.applyPreset(30, 0.01, 8, 8) },
			Duration:    4 * time.Second,
		},
		{
			Title:       "Step 2/5: Draw & Classify",
			Description: "Loading a test digit...\n\nWatch both paths run simultaneously:\n• FP (Float32): Ideal digital computation\n• CIM (30 Levels): Real hardware simulation\n\nBoth should predict the SAME digit!",
			Action: func() {
				gt.app.loadRandomSample()
			},
			Duration: 4 * time.Second,
		},
		{
			Title:       "Step 3/5: BREAK IT - Binary Weights",
			Description: "Now watch what happens with only 2 levels (binary)...\n\nAccuracy COLLAPSES to ~50%!\n\nLook at the weight heatmap: only blue and red.\nBinary weights cannot represent this neural network.\n\nThis is why 30 levels matter!",
			Action: func() {
				gt.app.applyPreset(2, 0.01, 8, 8)
				time.Sleep(500 * time.Millisecond)
				gt.app.runQuickTest()
			},
			Duration: 5 * time.Second,
		},
		{
			Title:       "Step 4/5: BREAK IT - High Noise",
			Description: "What about noise? Real hardware has:\n• Read circuit noise\n• Device variation\n• Temperature effects\n\nWith HIGH noise (0.15), accuracy drops to ~70%.\nNoise mitigation is crucial!",
			Action: func() {
				gt.app.applyPreset(30, 0.15, 6, 8)
				time.Sleep(500 * time.Millisecond)
				gt.app.runQuickTest()
			},
			Duration: 5 * time.Second,
		},
		{
			Title:       "Step 5/5: FeCIM Sweet Spot",
			Description: "Restoring optimal settings...\n\n30 levels + low noise = 87% accuracy!\n\nThe magic formula:\n• Enough precision to represent the network\n• Low enough noise to be manufacturable\n• 10,000x more energy-efficient than GPU\n\nExplore the presets to learn more!",
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
			widget.NewLabelWithStyle("Key Insights:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(""),
			widget.NewLabel("1. 30 analog levels → 87% accuracy"),
			widget.NewLabel("2. Binary (2 levels) → 50% (fails!)"),
			widget.NewLabel("3. High noise → 70% (degraded)"),
			widget.NewLabel("4. Energy: 10,000x better than GPU"),
			widget.NewLabel(""),
			widget.NewSeparator(),
			widget.NewLabelWithStyle("Next Steps:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("• Try the 'Tour' preset (Dr. Tour's architecture)"),
			widget.NewLabel("• Click 'Why 30?' for physics details"),
			widget.NewLabel("• Explore failure modes with presets"),
		)

		completionDialog := dialog.NewCustom("Tour Complete!", "Start Exploring", completionContent, gt.app.window)
		completionDialog.Show()
	})

	// Reset to ideal settings
	gt.app.applyPreset(30, 0.01, 8, 8)
	fyne.Do(func() {
		gt.app.statusLabel.SetText("Tour complete! Try the presets: Ideal, QuantCliff, Noisy, BrokenADC, Tour")
	})
}

// IsRunning returns whether the tour is currently running.
func (gt *GuidedTour) IsRunning() bool {
	return gt.isRunning
}

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
			Title:       "Step 1/7: Welcome to FeCIM Demo",
			Description: "This neural network classifies handwritten digits.\n\nBut instead of running on a GPU, it runs on a ferroelectric crossbar chip.\n\nFeCIM hardware achieves 87% accuracy with 30 analog levels.\n\nLet's see how it works!",
			Action:      func() { gt.app.applyPreset(30, 0.01, 8, 8) },
			Duration:    5 * time.Second,
		},
		{
			Title:       "Step 2/7: Draw a Digit",
			Description: "Draw a digit (0-9) on the canvas on the left.\n\nMake it clear, like writing on a whiteboard.\n\nOr click 'Random Sample' to load a test digit.",
			Action: func() {
				gt.app.loadRandomSample()
			},
			Duration: 4 * time.Second,
		},
		{
			Title:       "Step 3/7: FeCIM Classifies It",
			Description: "The network runs in TWO modes simultaneously:\n\n• Digital (FP): Ideal floating-point computation\n• FeCIM (Analog): Realistic hardware simulation\n\nNotice how both predict the same digit, but FeCIM has slightly lower confidence due to analog noise.",
			Action:      func() {},
			Duration:    5 * time.Second,
		},
		{
			Title:       "Step 4/7: The 30 Analog Levels",
			Description: "Look at the weight heatmap (bottom-right).\n\nEach crossbar cell stores ONE of 30 conductance states:\n• Blue = negative weight\n• White = zero\n• Red = positive weight\n\nThis is the key innovation: 30 levels vs binary (2 levels)!",
			Action:      func() {},
			Duration:    5 * time.Second,
		},
		{
			Title:       "Step 5/7: What If We Only Had 2 Levels?",
			Description: "Watch what happens when we reduce to binary weights...\n\nThe accuracy COLLAPSES to ~50% (worse than random guessing!)\n\nLook at the heatmap: only blue and red, no gradients.\nBinary weights cannot represent this network.",
			Action: func() {
				gt.app.applyPreset(2, 0.01, 8, 8)
				time.Sleep(500 * time.Millisecond)
				gt.app.runQuickTest()
			},
			Duration: 6 * time.Second,
		},
		{
			Title:       "Step 6/7: What About Noise?",
			Description: "Real hardware has noise from:\n• Read circuits\n• Device variation\n• Temperature effects\n\nWith 30 levels but HIGH noise (0.15), accuracy drops to ~70%.\n\nNoise mitigation is key to achieving good accuracy!",
			Action: func() {
				gt.app.applyPreset(30, 0.15, 6, 8)
				time.Sleep(500 * time.Millisecond)
				gt.app.runQuickTest()
			},
			Duration: 6 * time.Second,
		},
		{
			Title:       "Step 7/7: FeCIM's Sweet Spot",
			Description: "With 30 levels and low noise, we achieve 87% accuracy!\n\nThe sweet spot: enough precision to represent the network,\nlow enough noise to be manufacturable.\n\nAnd it uses 10,000x less energy than a GPU!",
			Action: func() {
				gt.app.applyPreset(30, 0.01, 8, 8)
				time.Sleep(500 * time.Millisecond)
				gt.app.runQuickTest()
			},
			Duration: 6 * time.Second,
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

	if gt.tourDialog != nil {
		gt.tourDialog.Hide()
	}

	// Reset to ideal settings
	gt.app.applyPreset(30, 0.01, 8, 8)
	gt.app.statusLabel.SetText("Tour ended. Ready to explore!")
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

	if gt.tourDialog != nil {
		gt.tourDialog.Hide()
	}

	// Show completion dialog
	completionContent := container.NewVBox(
		widget.NewLabel("Tour Complete!"),
		widget.NewSeparator(),
		widget.NewLabel("Key Takeaways:"),
		widget.NewLabel("• 30 analog levels enable 87% accuracy"),
		widget.NewLabel("• Binary (2 levels) fails completely (~50%)"),
		widget.NewLabel("• Noise must be controlled for accuracy"),
		widget.NewLabel("• FeCIM uses 10,000x less energy than GPU"),
		widget.NewSeparator(),
		widget.NewLabel("Now explore the presets yourself!"),
	)

	completionDialog := dialog.NewCustom("Tour Complete!", "Close", completionContent, gt.app.window)
	completionDialog.Show()

	// Reset to ideal settings
	gt.app.applyPreset(30, 0.01, 8, 8)
	gt.app.statusLabel.SetText("Tour complete! Explore the presets.")
}

// IsRunning returns whether the tour is currently running.
func (gt *GuidedTour) IsRunning() bool {
	return gt.isRunning
}

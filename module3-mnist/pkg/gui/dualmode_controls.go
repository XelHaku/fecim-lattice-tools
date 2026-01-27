// Package gui provides Fyne-based GUI components for MNIST visualization.
// dualmode_controls.go provides control panel creation and preset functionality.
package gui

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
	"fecim-lattice-tools/module3-mnist/pkg/core"
	"fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/module3-mnist/pkg/training"
)

// createControlsZone creates the hardware control panel (Zone 3).
func (app *DualModeApp) createControlsZone() fyne.CanvasObject {
	label := widget.NewLabel("Hardware Config")
	label.TextStyle = fyne.TextStyle{Bold: true}

	// Levels slider (2-256)
	levelsTitle := widget.NewLabel("Levels:")
	app.levelsLabel = widget.NewLabel(fmt.Sprintf("%d", FeCIMDefaultLevels))
	app.levelsSlider = widget.NewSlider(2, 256)
	app.levelsSlider.Step = 1
	app.levelsSlider.Value = float64(FeCIMDefaultLevels)
	app.levelsSlider.OnChanged = func(v float64) {
		mnistLog.SliderChange("Levels", v)
		levels := int(v)
		app.levelsLabel.SetText(fmt.Sprintf("%d", levels))

		// Check if we should load level-specific QAT weights
		bestLevel := core.GetBestMatchingWeightsLevel(levels)
		app.tryLoadQATWeights(bestLevel)

		app.network.SetNumLevels(levels)
		app.updateWeightHeatmap()
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	}
	levelsRow := container.NewBorder(nil, nil,
		container.NewHBox(levelsTitle, widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
			dialog.ShowInformation("Quantization Levels", "Number of discrete conductance states in the ferroelectric device.\n\nHigher levels = better precision but harder to stabilize.\n\nPhysics: Ferroelectric material domains determine stable polarization states.", app.window)
		})),
		app.levelsLabel, app.levelsSlider)

	// Noise slider
	noiseTitle := widget.NewLabel("Noise:")
	app.noiseLabel = widget.NewLabel(fmt.Sprintf("%.2f", FeCIMDefaultNoise))
	app.noiseSlider = widget.NewSlider(0.0, 0.20)
	app.noiseSlider.Step = 0.01
	app.noiseSlider.Value = FeCIMDefaultNoise
	app.noiseSlider.OnChanged = func(v float64) {
		mnistLog.SliderChange("Noise", v)
		app.noiseLabel.SetText(fmt.Sprintf("%.2f", v))
		app.network.SetNoiseLevel(v)
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	}
	noiseRow := container.NewBorder(nil, nil,
		container.NewHBox(noiseTitle, widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
			dialog.ShowInformation("Read Noise", "Gaussian noise added to analog read operations.\n\nSimulates:\n• Thermal (Johnson) noise in sense amplifiers\n• Device-to-device variability\n• Cycle-to-cycle retention drift\n\n0.01 = 1% standard deviation (typical for production hardware)", app.window)
		})),
		app.noiseLabel, app.noiseSlider)

	// ADC/DAC + Hidden selects on one row
	bitOptions := []string{"3", "4", "5", "6", "7", "8", "10", "12", "16"}
	app.adcSelect = widget.NewSelect(bitOptions, func(s string) {
		mnistLog.Selection("ADC", s)
		var bits int
		fmt.Sscanf(s, "%d", &bits)
		app.network.SetADCBits(bits)
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	})
	app.adcSelect.SetSelected(fmt.Sprintf("%d", FeCIMDefaultADC))

	app.dacSelect = widget.NewSelect(bitOptions, func(s string) {
		mnistLog.Selection("DAC", s)
		var bits int
		fmt.Sscanf(s, "%d", &bits)
		app.network.SetDACBits(bits)
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	})
	app.dacSelect.SetSelected(fmt.Sprintf("%d", FeCIMDefaultDAC))

	app.hiddenSelect = widget.NewSelect([]string{"64", "128", "256"}, func(s string) {
		mnistLog.Selection("Hidden", s)
		var size int
		fmt.Sscanf(s, "%d", &size)
		go app.changeHiddenSize(size)
	})
	app.hiddenSelect.SetSelected("128")

	// ADC/DAC/Hidden selects - use 2 rows for better spacing
	selectsRow1 := container.NewGridWithColumns(4,
		widget.NewLabel("ADC:"), app.adcSelect,
		widget.NewLabel("DAC:"), app.dacSelect,
	)
	selectsRow2 := container.NewGridWithColumns(2,
		widget.NewLabel("Hidden:"), app.hiddenSelect,
	)
	selectsRow := container.NewVBox(selectsRow1, selectsRow2)

	// Preset buttons (first row: standard presets)
	idealBtn := widget.NewButton("Ideal", func() {
		mnistLog.Button("Preset:Ideal")
		app.applyPresetWithMode(FeCIMDefaultLevels, FeCIMDefaultNoise, FeCIMDefaultADC, FeCIMDefaultDAC, false)
	})
	quantCliffBtn := widget.NewButton("QuantCliff", func() {
		mnistLog.Button("Preset:QuantCliff")
		app.applyPresetWithMode(2, FeCIMDefaultNoise, FeCIMDefaultADC, FeCIMDefaultDAC, false)
	})

	// Hardware preset - realistic production hardware
	hw87Btn := widget.NewButton("Hardware", func() {
		mnistLog.Button("Preset:Hardware")
		// Realistic production config: 30 levels, 3% noise, 6-bit ADC, single layer
		app.applyPresetWithMode(FeCIMDefaultLevels, 0.03, 6, FeCIMDefaultDAC, true)
	})
	hw87Btn.Importance = widget.HighImportance // Highlight alongside Calibration button

	noisyBtn := widget.NewButton("Noisy", func() {
		mnistLog.Button("Preset:Noisy")
		app.applyPresetWithMode(FeCIMDefaultLevels, 0.15, 6, FeCIMDefaultDAC, false)
	})
	brokenBtn := widget.NewButton("BrokenADC", func() {
		mnistLog.Button("Preset:BrokenADC")
		app.applyPresetWithMode(FeCIMDefaultLevels, FeCIMDefaultNoise, 3, FeCIMDefaultDAC, false)
	})

	// Calibration Mode button (single-layer 784→10, matching hardware architecture)
	// Achieved accuracy depends on physics simulation, not forced target
	tourBtn := widget.NewButton("Calibration", func() {
		mnistLog.Button("Preset:Calibration")
		app.applyPresetWithMode(FeCIMDefaultLevels, FeCIMDefaultNoise, FeCIMDefaultADC, FeCIMDefaultDAC, true)
	})
	tourBtn.Importance = widget.HighImportance // Highlight this button

	// Quick test with progress bar
	app.testResultLabel = widget.NewLabel("")
	app.testProgressBar = widget.NewProgressBar()
	app.testProgressBar.Hide() // Hidden until test starts
	app.testButton = widget.NewButton("Test (200)", func() {
		mnistLog.Button("Test200")
		app.runQuickTest()
	})

	// Train Weights button - allows user to generate weights for current level
	trainWeightsBtn := widget.NewButton("Train Weights", func() {
		mnistLog.Button("TrainWeights")
		app.showTrainWeightsDialog()
	})

	// Simple Mode presets (shown by default)
	app.simplePresets = container.NewGridWithColumns(3, idealBtn, hw87Btn, noisyBtn)

	// Expert-only controls (hidden by default)
	expertPresetRow := container.NewGridWithColumns(3, quantCliffBtn, brokenBtn, tourBtn)
	testRow := container.NewHBox(app.testButton, trainWeightsBtn, app.testProgressBar, app.testResultLabel)
	app.expertOnlyControls = container.NewVBox(
		selectsRow,
		expertPresetRow,
		testRow,
	)

	// Set initial visibility based on Expert Mode
	if !app.expertMode {
		app.expertOnlyControls.Hide()
	} else {
		app.simplePresets.Hide()
	}

	// Controls layout with header
	controlsHeader := container.NewVBox(
		container.NewHBox(label, layout.NewSpacer()),
		levelsRow,
		noiseRow,
		app.simplePresets,      // Simple mode presets
		app.expertOnlyControls, // Expert-only controls
	)

	return controlsHeader
}

// applyPreset sets hardware parameters to a failure mode preset.
func (app *DualModeApp) applyPreset(levels int, noise float64, adcBits, dacBits int) {
	app.applyPresetWithMode(levels, noise, adcBits, dacBits, false)
}

// applyPresetWithMode sets hardware parameters with optional Calibration Mode (single-layer).
func (app *DualModeApp) applyPresetWithMode(levels int, noise float64, adcBits, dacBits int, singleLayer bool) {
	// Update network parameters (thread-safe, not UI)
	app.network.SetNumLevels(levels)
	app.network.SetNoiseLevel(noise)
	app.network.SetADCBits(adcBits)
	app.network.SetDACBits(dacBits)
	app.network.SetSingleLayer(singleLayer)

	// Update UI elements on main thread
	fyne.Do(func() {
		app.levelsSlider.SetValue(float64(levels))
		app.noiseSlider.SetValue(noise)
		app.adcSelect.SetSelected(fmt.Sprintf("%d", adcBits))
		app.dacSelect.SetSelected(fmt.Sprintf("%d", dacBits))

		// Update status to indicate Calibration Mode
		if singleLayer {
			app.statusLabel.SetText("Calibration Mode: Single-layer network (784→10) matching hardware architecture | Target: Physics-limited accuracy")
		}

		app.updateWeightHeatmap()

		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	})
}

// runQuickTest runs inference on 200 samples and reports accuracy.
func (app *DualModeApp) runQuickTest() {
	fyne.Do(func() {
		app.testButton.Disable()
		app.testResultLabel.SetText("")
		app.testProgressBar.Show()
		app.testProgressBar.SetValue(0)
	})

	go func() {
		if len(app.testImages) == 0 {
			fyne.Do(func() {
				app.testResultLabel.SetText("Loading test data...")
			})
			app.loadTestData()
		}

		n := 200
		if n > len(app.testImages) {
			n = len(app.testImages)
		}

		fpCorrect := 0
		cimCorrect := 0
		agreements := 0

		// Update progress every 10 samples to avoid UI overload
		updateInterval := 10

		for i := 0; i < n; i++ {
			result := app.network.Infer(app.testImages[i])
			if result.FPPrediction == app.testLabels[i] {
				fpCorrect++
			}
			if result.CIMPrediction == app.testLabels[i] {
				cimCorrect++
			}
			if result.Agree {
				agreements++
			}

			// Update progress bar at intervals
			if (i+1)%updateInterval == 0 || i == n-1 {
				progress := float64(i+1) / float64(n)
				current := i + 1
				fyne.Do(func() {
					app.testProgressBar.SetValue(progress)
					app.testResultLabel.SetText(fmt.Sprintf("Testing %d/%d...", current, n))
				})
			}
		}

		fpAcc := float64(fpCorrect) / float64(n) * 100
		cimAcc := float64(cimCorrect) / float64(n) * 100
		agreeRate := float64(agreements) / float64(n) * 100

		fyne.Do(func() {
			app.testProgressBar.Hide()
			app.testResultLabel.SetText(fmt.Sprintf(
				"FP: %.1f%% | CIM: %.1f%% | Agreement: %.1f%%",
				fpAcc, cimAcc, agreeRate))
			app.testButton.Enable()
		})
	}()
}

// showTrainWeightsDialog shows a dialog to train weights for the current quantization level.
func (app *DualModeApp) showTrainWeightsDialog() {
	currentLevels := int(app.levelsSlider.Value)

	// Create progress UI elements
	progressBar := widget.NewProgressBar()
	statusLabel := widget.NewLabel("Ready to train")
	epochLabel := widget.NewLabel("")
	accuracyLabel := widget.NewLabel("")

	// Training parameters (with sensible defaults)
	epochsEntry := widget.NewEntry()
	epochsEntry.SetText("10")
	samplesEntry := widget.NewEntry()
	samplesEntry.SetText("10000")

	// Training state (use atomic.Bool for thread-safe access between UI and training goroutine)
	var isTraining atomic.Bool
	var stopTraining atomic.Bool

	// Create the dialog content
	form := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Train weights optimized for %d quantization levels", currentLevels)),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("Epochs:"), epochsEntry,
			widget.NewLabel("Training samples:"), samplesEntry,
		),
		widget.NewSeparator(),
		progressBar,
		statusLabel,
		epochLabel,
		accuracyLabel,
	)

	// Create buttons
	var trainBtn *widget.Button

	trainBtn = widget.NewButton("Start Training", func() {
		if isTraining.Load() {
			return
		}

		// Parse parameters
		var epochs int
		var samples int
		_, err := fmt.Sscanf(epochsEntry.Text, "%d", &epochs)
		if err != nil || epochs < 1 {
			epochs = 10
		}
		_, err = fmt.Sscanf(samplesEntry.Text, "%d", &samples)
		if err != nil || samples < 100 {
			samples = 10000
		}

		isTraining.Store(true)
		stopTraining.Store(false)
		trainBtn.Disable()
		epochsEntry.Disable()
		samplesEntry.Disable()

		go app.runTraining(currentLevels, epochs, samples, progressBar, statusLabel, epochLabel, accuracyLabel, &stopTraining, func() {
			fyne.Do(func() {
				isTraining.Store(false)
				trainBtn.Enable()
				epochsEntry.Enable()
				samplesEntry.Enable()
			})
		})
	})
	trainBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Stop", func() {
		stopTraining.Store(true)
		statusLabel.SetText("Stopping training...")
	})

	buttonRow := container.NewHBox(layout.NewSpacer(), trainBtn, cancelBtn, layout.NewSpacer())
	content := container.NewVBox(form, buttonRow)

	d := dialog.NewCustom("Train Weights", "Close", content, app.window)
	d.SetOnClosed(func() {
		stopTraining.Store(true) // Stop training when dialog is closed
	})
	d.Resize(fyne.NewSize(450, 350))
	d.Show()
}

// runTraining performs the actual training in a background goroutine.
func (app *DualModeApp) runTraining(levels, epochs, samples int, progressBar *widget.ProgressBar, statusLabel, epochLabel, accuracyLabel *widget.Label, stopTraining *atomic.Bool, onComplete func()) {
	defer onComplete()

	// Use local RNG for shuffling
	// For debugging: set FECIM_DEBUG_SEED=42 (or any integer) for reproducible training
	var rng *rand.Rand
	if seedStr := os.Getenv("FECIM_DEBUG_SEED"); seedStr != "" {
		if seed, err := strconv.ParseInt(seedStr, 10, 64); err == nil {
			rng = rand.New(rand.NewSource(seed))
			mnistLog.Printf("Debug mode: using fixed seed %d for reproducible training", seed)
		} else {
			rng = rand.New(rand.NewSource(rand.Int63()))
		}
	} else {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}

	fyne.Do(func() {
		statusLabel.SetText("Loading MNIST training data...")
		progressBar.SetValue(0)
	})

	// Load MNIST training data
	trainImages, trainLabels, err := mnist.LoadMNIST(app.dataDir, true)
	if err != nil {
		fyne.Do(func() {
			statusLabel.SetText(fmt.Sprintf("Error loading data: %v", err))
		})
		return
	}

	// Limit samples
	if samples > len(trainImages) {
		samples = len(trainImages)
	}
	trainImages = trainImages[:samples]
	trainLabels = trainLabels[:samples]

	fyne.Do(func() {
		statusLabel.SetText(fmt.Sprintf("Loaded %d training samples", samples))
	})

	// Load test data for evaluation
	testImages, testLabels, err := mnist.LoadMNIST(app.dataDir, false)
	if err != nil {
		fyne.Do(func() {
			statusLabel.SetText(fmt.Sprintf("Error loading test data: %v", err))
		})
		return
	}
	// Use subset for evaluation
	if len(testImages) > 1000 {
		testImages = testImages[:1000]
		testLabels = testLabels[:1000]
	}

	// Create crossbar arrays for training
	// Note: Crossbar uses 30-level quantization internally, but the saved weights
	// can be re-quantized to different levels at inference time by DualModeNetwork
	hidden := 128
	layer1, err := crossbar.NewArray(&crossbar.Config{
		Rows: hidden, Cols: 784, NoiseLevel: 0, ADCBits: 16, DACBits: 16,
	})
	if err != nil {
		fyne.Do(func() {
			statusLabel.SetText(fmt.Sprintf("Error creating layer1: %v", err))
		})
		return
	}

	layer2, err := crossbar.NewArray(&crossbar.Config{
		Rows: 10, Cols: hidden, NoiseLevel: 0, ADCBits: 16, DACBits: 16,
	})
	if err != nil {
		fyne.Do(func() {
			statusLabel.SetText(fmt.Sprintf("Error creating layer2: %v", err))
		})
		return
	}

	// Create training network
	net := training.NewMNISTNetwork(layer1, layer2)

	// Training parameters
	learningRate := 1.0
	bestAcc := 0.0

	fyne.Do(func() {
		statusLabel.SetText("Training...")
	})

	for epoch := 1; epoch <= epochs; epoch++ {
		if stopTraining.Load() {
			fyne.Do(func() {
				statusLabel.SetText("Training stopped by user")
			})
			return
		}

		// Shuffle training data
		perm := rng.Perm(len(trainImages))
		shuffledImages := make([][]float64, len(trainImages))
		shuffledLabels := make([]int, len(trainImages))
		for i, p := range perm {
			shuffledImages[i] = trainImages[p]
			shuffledLabels[i] = trainLabels[p]
		}

		// Train one epoch
		loss := net.TrainEpoch(shuffledImages, shuffledLabels, learningRate)

		// Evaluate
		acc := net.Evaluate(testImages, testLabels) * 100
		if acc > bestAcc {
			bestAcc = acc
		}

		// Update UI
		progress := float64(epoch) / float64(epochs)
		currentEpoch := epoch
		currentLoss := loss
		currentAcc := acc
		currentBest := bestAcc
		fyne.Do(func() {
			progressBar.SetValue(progress)
			epochLabel.SetText(fmt.Sprintf("Epoch %d/%d | Loss: %.4f", currentEpoch, epochs, currentLoss))
			accuracyLabel.SetText(fmt.Sprintf("Accuracy: %.1f%% | Best: %.1f%%", currentAcc, currentBest))
		})

		// Learning rate decay
		if epoch%5 == 0 {
			learningRate *= 0.5
		}
	}

	// Save weights
	weightsPath := core.GetWeightsFilename(app.dataDir, levels)
	// For non-standard levels, save with level number
	if levels != 30 {
		weightsPath = filepath.Join(app.dataDir, fmt.Sprintf("pretrained_weights_%d.json", levels))
	}

	fyne.Do(func() {
		statusLabel.SetText(fmt.Sprintf("Saving weights to %s...", filepath.Base(weightsPath)))
	})

	if err := net.SaveWeights(weightsPath); err != nil {
		fyne.Do(func() {
			statusLabel.SetText(fmt.Sprintf("Error saving weights: %v", err))
		})
		return
	}

	// Clear the "warned" flag so the new weights can be loaded (thread-safe access)
	app.warnedMissingLevelsMu.Lock()
	app.warnedMissingLevels[levels] = false
	app.warnedMissingLevelsMu.Unlock()

	// Reload the weights into the current network (LoadWeights is thread-safe internally)
	if err := app.network.LoadWeights(weightsPath); err != nil {
		fyne.Do(func() {
			statusLabel.SetText(fmt.Sprintf("Trained but error loading: %v", err))
		})
		return
	}

	// Update state on main thread to ensure thread safety
	finalBestAcc := bestAcc
	app.currentQATLevelMu.Lock()
	app.currentQATLevel = levels
	app.currentQATLevelMu.Unlock()
	fyne.Do(func() {
		statusLabel.SetText(fmt.Sprintf("Training complete! Accuracy: %.1f%% | Weights saved.", finalBestAcc))
		progressBar.SetValue(1.0)
		app.updateWeightHeatmap()
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	})
}

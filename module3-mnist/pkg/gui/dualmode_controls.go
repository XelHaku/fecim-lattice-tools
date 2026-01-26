// Package gui provides Fyne-based GUI components for MNIST visualization.
// dualmode_controls.go provides control panel creation and preset functionality.
package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module3-mnist/pkg/core"
)

// createControlsZone creates the hardware control panel (Zone 3).
func (app *DualModeApp) createControlsZone() fyne.CanvasObject {
	label := widget.NewLabel("Hardware Config")
	label.TextStyle = fyne.TextStyle{Bold: true}

	// Levels slider (2-31, covers all QAT-trained levels)
	levelsTitle := widget.NewLabel("Levels:")
	app.levelsLabel = widget.NewLabel(fmt.Sprintf("%d", FeCIMDefaultLevels))
	app.levelsSlider = widget.NewSlider(2, 31)
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
			dialog.ShowInformation("Quantization Levels", "Number of discrete conductance states in the ferroelectric device.\n\n30 levels = ~4.9 bits per cell (FeCIM standard)\n\nPhysics: HZO ferroelectric material exhibits ~30 stable polarization states due to domain wall pinning at crystal defects.", app.window)
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
	noisyBtn := widget.NewButton("Noisy", func() {
		mnistLog.Button("Preset:Noisy")
		app.applyPresetWithMode(FeCIMDefaultLevels, 0.15, 6, FeCIMDefaultDAC, false)
	})
	brokenBtn := widget.NewButton("BrokenADC", func() {
		mnistLog.Button("Preset:BrokenADC")
		app.applyPresetWithMode(FeCIMDefaultLevels, FeCIMDefaultNoise, 3, FeCIMDefaultDAC, false)
	})

	// Tour Mode button (single-layer 784→10, matching Dr. Tour's architecture)
	// Achieved 83% accuracy with 30-level quantization (Tour claimed 87% with 88% theoretical max)
	tourBtn := widget.NewButton("Tour", func() {
		mnistLog.Button("Preset:Tour")
		app.applyPresetWithMode(FeCIMDefaultLevels, FeCIMDefaultNoise, FeCIMDefaultADC, FeCIMDefaultDAC, true)
	})
	tourBtn.Importance = widget.HighImportance // Highlight this button

	// Preset buttons - use 2 rows for better spacing (3+2)
	presetRow1 := container.NewGridWithColumns(3, idealBtn, quantCliffBtn, noisyBtn)
	presetRow2 := container.NewGridWithColumns(2, brokenBtn, tourBtn)
	presetRow := container.NewVBox(presetRow1, presetRow2)

	// Quick test with progress bar
	app.testResultLabel = widget.NewLabel("")
	app.testProgressBar = widget.NewProgressBar()
	app.testProgressBar.Hide() // Hidden until test starts
	app.testButton = widget.NewButton("Test (200)", func() {
		mnistLog.Button("Test200")
		app.runQuickTest()
	})

	// Controls header (compact, fixed height)
	controlsHeader := container.NewVBox(
		container.NewHBox(label, layout.NewSpacer(), app.testButton, app.testProgressBar, app.testResultLabel),
		levelsRow,
		noiseRow,
		selectsRow,
		presetRow,
	)

	return controlsHeader
}

// applyPreset sets hardware parameters to a failure mode preset.
func (app *DualModeApp) applyPreset(levels int, noise float64, adcBits, dacBits int) {
	app.applyPresetWithMode(levels, noise, adcBits, dacBits, false)
}

// applyPresetWithMode sets hardware parameters with optional Tour Mode (single-layer).
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

		// Update status to indicate Tour Mode
		if singleLayer {
			app.statusLabel.SetText("Tour Mode: Single-layer network (784→10) matching Dr. Tour's COSM 2025 architecture | Target: 87% accuracy")
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
				"FP: %.1f%% | CIM: %.1f%% | Agreement: %.1f%% | Target: 87%%",
				fpAcc, cimAcc, agreeRate))
			app.testButton.Enable()
		})
	}()
}

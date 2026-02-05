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

	"fecim-lattice-tools/module3-mnist/pkg/core"
)

// createControlsZone creates the hardware control panel (Zone 3).
func (app *DualModeApp) createControlsZone() fyne.CanvasObject {
	label := widget.NewLabel("Hardware Config")
	label.TextStyle = fyne.TextStyle{Bold: true}

	// Levels select (only available QAT levels)
	levelsTitle := widget.NewLabel("Levels:")
	app.levelsLabel = widget.NewLabel(fmt.Sprintf("%d", FeCIMDefaultLevels))

	// Build options from available QAT levels (scanned from data directory)
	availableLevels := core.ScanAvailableQATLevels(app.weightsDir())
	levelOptions := make([]string, len(availableLevels))
	for i, l := range availableLevels {
		levelOptions[i] = fmt.Sprintf("%d", l)
	}

	app.levelsSelect = widget.NewSelect(levelOptions, func(s string) {
		var levels int
		fmt.Sscanf(s, "%d", &levels)
		mnistLog.Selection("Levels", s)
		app.levelsLabel.SetText(fmt.Sprintf("%d", levels))

		// Load weights for this level
		app.tryLoadQATWeights(levels)

		app.network().SetNumLevels(levels)
		app.updateWeightHeatmap()
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	})
	app.levelsSelect.SetSelected(fmt.Sprintf("%d", FeCIMDefaultLevels))

	levelsRow := container.NewBorder(nil, nil,
		container.NewHBox(levelsTitle, widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
			dialog.ShowInformation("Quantization Levels", "Number of discrete conductance states in the ferroelectric device.\n\nOnly levels with trained weights are shown.\n\nPhysics: Ferroelectric material domains determine stable polarization states.", app.window)
		})),
		app.levelsLabel, app.levelsSelect)

	// Noise slider
	noiseTitle := widget.NewLabel("Noise:")
	app.noiseLabel = widget.NewLabel(fmt.Sprintf("%.2f", FeCIMDefaultNoise))
	app.noiseSlider = widget.NewSlider(0.0, 0.20)
	app.noiseSlider.Step = 0.01
	app.noiseSlider.Value = FeCIMDefaultNoise
	app.noiseSlider.OnChanged = func(v float64) {
		mnistLog.SliderChange("Noise", v)
		app.noiseLabel.SetText(fmt.Sprintf("%.2f", v))
		app.network().SetNoiseLevel(v)
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	}
	noiseRow := container.NewBorder(nil, nil,
		container.NewHBox(noiseTitle, widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
			dialog.ShowInformation("Read Noise", "Gaussian noise added to analog read operations.\n\nSimulates:\n• Thermal (Johnson) noise in sense amplifiers\n• Device-to-device variability\n• Cycle-to-cycle retention drift\n\n0.01 = 1% standard deviation (typical for production hardware)", app.window)
		})),
		app.noiseLabel, app.noiseSlider)

	// Preprocess toggle
	preprocessTitle := widget.NewLabel("Preprocess:")
	preprocessInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Input Preprocessing", "Normalizes drawn digits to match MNIST style:\n• Threshold background\n• Crop to bounding box\n• Scale to 20×20\n• Center by mass\n\nRandom test samples are never altered.", app.window)
	})
	app.preprocessToggle = widget.NewCheck("", func(checked bool) {
		app.preprocessEnabled = checked
		mnistLog.Selection("Preprocess", map[bool]string{true: "On", false: "Off"}[checked])
		if len(app.lastRawPixels) > 0 {
			effective := app.lastRawPixels
			if app.preprocessEnabled {
				effective = preprocessIfUserInput(app.digitCanvas, app.lastRawPixels)
			}
			app.lastPixels = effective
			app.runInference(effective)
		}
	})
	app.preprocessToggle.SetChecked(app.preprocessEnabled)

	preprocessRow := container.NewHBox(
		container.NewHBox(preprocessTitle, preprocessInfo),
		layout.NewSpacer(),
		app.preprocessToggle,
	)

	// Preset buttons
	idealBtn := widget.NewButton("Ideal", func() {
		mnistLog.Button("Preset:Ideal")
		app.applyPreset(FeCIMDefaultLevels, FeCIMDefaultNoise, FeCIMDefaultADC, FeCIMDefaultDAC)
	})

	hwBtn := widget.NewButton("Hardware", func() {
		mnistLog.Button("Preset:Hardware")
		// Demo baseline config: 30 levels, 3% noise, 6-bit ADC
		app.applyPreset(FeCIMDefaultLevels, 0.03, 6, FeCIMDefaultDAC)
	})
	hwBtn.Importance = widget.HighImportance

	noisyBtn := widget.NewButton("Noisy", func() {
		mnistLog.Button("Preset:Noisy")
		app.applyPreset(FeCIMDefaultLevels, 0.15, 6, FeCIMDefaultDAC)
	})

	presetRow := container.NewGridWithColumns(3, idealBtn, hwBtn, noisyBtn)

	return container.NewVBox(
		container.NewHBox(label, layout.NewSpacer()),
		levelsRow,
		noiseRow,
		preprocessRow,
		presetRow,
	)
}

// applyPreset sets hardware parameters to a preset configuration.
func (app *DualModeApp) applyPreset(levels int, noise float64, adcBits, dacBits int) {
	// Attempt to load QAT weights for this level (if available)
	app.tryLoadQATWeights(levels)

	// Update network parameters (thread-safe, not UI)
	app.network().SetNumLevels(levels)
	app.network().SetNoiseLevel(noise)
	app.network().SetADCBits(adcBits)
	app.network().SetDACBits(dacBits)

	// Update UI elements on main thread
	fyne.Do(func() {
		if app.levelsSelect != nil {
			selected := fmt.Sprintf("%d", levels)
			hasOption := false
			for _, option := range app.levelsSelect.Options {
				if option == selected {
					hasOption = true
					break
				}
			}
			if hasOption {
				app.levelsSelect.SetSelected(selected)
			} else if app.levelsLabel != nil {
				app.levelsLabel.SetText(selected)
			}
		} else if app.levelsLabel != nil {
			app.levelsLabel.SetText(fmt.Sprintf("%d", levels))
		}
		app.noiseSlider.SetValue(noise)
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
		if len(app.testImages()) == 0 {
			fyne.Do(func() {
				app.testResultLabel.SetText("Loading test data...")
			})
			if err := app.networkCtrl.LoadTestData(); err != nil {
				mnistLog.Printf("Failed to load test data: %v", err)
			}
		}

		n := 200
		if n > len(app.testImages()) {
			n = len(app.testImages())
		}

		fpCorrect := 0
		cimCorrect := 0
		agreements := 0

		// Update progress every 10 samples to avoid UI overload
		updateInterval := 10

		for i := 0; i < n; i++ {
			result := app.network().Infer(app.testImages()[i])
			if result.FPPrediction == app.testLabels()[i] {
				fpCorrect++
			}
			if result.CIMPrediction == app.testLabels()[i] {
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

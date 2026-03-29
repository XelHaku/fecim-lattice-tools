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
	sharedexport "fecim-lattice-tools/shared/export"
)

// createControlsZone creates the hardware control panel (Zone 3).
func (app *DualModeApp) createControlsZone() fyne.CanvasObject {
	label := widget.NewLabel("Hardware Config")
	label.TextStyle = fyne.TextStyle{Bold: true}

	// Tooltip reference
	tips := mnistTooltips()

	// Helper to show tooltip dialogs
	showTooltip := func(tc tooltipContent) {
		content := fmt.Sprintf("%s\n\n📊 Range: %s\n\n🔬 Physics: %s",
			tc.description, tc.rangeInfo, tc.physics)
		if len(tc.tips) > 0 {
			content += "\n\n💡 Tips:\n"
			for _, tip := range tc.tips {
				content += "• " + tip + "\n"
			}
		}
		dialog.ShowInformation(tc.title, content, app.window)
	}

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
			showTooltip(tips.levels)
		})),
		app.levelsLabel, app.levelsSelect)

	// Noise slider
	noiseTitle := widget.NewLabel("Noise:")
	app.noiseLabel = widget.NewLabel(fmt.Sprintf("%.1f%%", FeCIMDefaultNoise*100))
	app.noiseSlider = widget.NewSlider(0.0, core.MaxNoiseLevel)
	app.noiseSlider.Step = 0.005
	app.noiseSlider.Value = FeCIMDefaultNoise
	app.noiseSlider.OnChanged = func(v float64) {
		mnistLog.SliderChange("Noise", v)
		app.noiseLabel.SetText(fmt.Sprintf("%.1f%%", v*100))
		app.network().SetNoiseLevel(v)
		if len(app.lastPixels) > 0 {
			app.runInference(app.lastPixels)
		}
	}
	noiseRow := container.NewBorder(nil, nil,
		container.NewHBox(noiseTitle, widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
			showTooltip(tips.noise)
		})),
		app.noiseLabel, app.noiseSlider)

	// Preprocess toggle
	preprocessTitle := widget.NewLabel("Preprocess:")
	preprocessInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		showTooltip(tips.preprocess)
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

	// Preset buttons with tooltips
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

	// Preset info buttons
	presetInfoRow := container.NewGridWithColumns(3,
		widget.NewButtonWithIcon("", theme.InfoIcon(), func() { showTooltip(tips.presetIdeal) }),
		widget.NewButtonWithIcon("", theme.InfoIcon(), func() { showTooltip(tips.presetHW) }),
		widget.NewButtonWithIcon("", theme.InfoIcon(), func() { showTooltip(tips.presetNoisy) }),
	)

	presetRow := container.NewVBox(
		container.NewGridWithColumns(3, idealBtn, hwBtn, noisyBtn),
		presetInfoRow,
	)

	// Export buttons
	exportLabel := widget.NewLabel("Export")
	exportLabel.TextStyle = fyne.TextStyle{Bold: true}

	exportDataBtn := sharedexport.CreateExportButton("Export Data", func() {
		mnistLog.Button("ExportData")
		app.exportInferenceData()
	}, app.window)

	exportImageBtn := sharedexport.CreateExportButton("Save Image", func() {
		mnistLog.Button("ExportImage")
		app.exportVisualization()
	}, app.window)

	exportRow := container.NewGridWithColumns(2, exportDataBtn, exportImageBtn)

	return container.NewVBox(
		container.NewHBox(label, layout.NewSpacer()),
		levelsRow,
		noiseRow,
		preprocessRow,
		presetRow,
		widget.NewSeparator(),
		container.NewHBox(exportLabel, layout.NewSpacer()),
		exportRow,
	)
}

// tooltipContent holds structured tooltip information for MNIST module.
type tooltipContent struct {
	title       string
	description string
	rangeInfo   string
	physics     string
	tips        []string
}

// mnistTooltipsStruct holds all MNIST-specific tooltips.
type mnistTooltipsStruct struct {
	levels      tooltipContent
	noise       tooltipContent
	preprocess  tooltipContent
	presetIdeal tooltipContent
	presetHW    tooltipContent
	presetNoisy tooltipContent
	confidence  tooltipContent
	comparison  tooltipContent
	energy      tooltipContent
}

// mnistTooltips returns all MNIST module tooltips.
func mnistTooltips() mnistTooltipsStruct {
	return mnistTooltipsStruct{
		levels: tooltipContent{
			title:       "Quantization Levels",
			description: "Number of discrete weight values in the neural network. Affects model accuracy and hardware efficiency.",
			rangeInfo:   "2-256 levels (weights from pretrained QAT models)",
			physics:     "Each level corresponds to a programmable conductance state in the crossbar. Quantization-aware training (QAT) optimizes weights for discrete levels.",
			tips: []string{
				"30 levels ≈ 4.9 bits (conference claim for FeCIM)",
				"8-16 levels often sufficient for MNIST",
				"Only levels with trained weights are available",
			},
		},
		noise: tooltipContent{
			title:       "Read Noise",
			description: "Gaussian noise added to weight reads, simulating analog hardware imperfections.",
			rangeInfo:   "0-20% (standard deviation relative to weight magnitude)",
			physics:     "Combines thermal noise, device variability, and retention drift. The network must be robust to this noise for reliable inference.",
			tips: []string{
				"0% = ideal (floating-point equivalent)",
				"3% = typical production hardware target",
				"10%+ = stress test for noise robustness",
			},
		},
		preprocess: tooltipContent{
			title:       "Input Preprocessing",
			description: "Normalizes hand-drawn digits to match MNIST training distribution.",
			rangeInfo:   "On/Off toggle",
			physics:     "Data preprocessing, not physics. Centers and scales drawn digits to match the 28×28 centered format of MNIST training data.",
			tips: []string{
				"Enable for better recognition of drawn digits",
				"Random test samples are never preprocessed",
				"Preprocessing: threshold → crop → scale → center",
			},
		},
		presetIdeal: tooltipContent{
			title:       "Ideal Preset",
			description: "Maximum precision, no noise—equivalent to floating-point inference.",
			rangeInfo:   "30 levels, 0% noise, 8-bit ADC/DAC",
			physics:     "Best-case scenario: no analog non-idealities. Used as the accuracy reference.",
			tips: []string{
				"Use to establish baseline accuracy",
				"Compare against hardware presets",
				"Should match FP32 closely",
			},
		},
		presetHW: tooltipContent{
			title:       "Hardware Preset",
			description: "Realistic hardware parameters based on published FeCIM specifications.",
			rangeInfo:   "30 levels, 3% noise, 6-bit ADC",
			physics:     "Represents achievable performance with current FeCIM technology. Noise includes read variability and retention drift.",
			tips: []string{
				"This is the target for real hardware",
				"Accuracy should be close to ideal",
				"Small accuracy drop is acceptable",
			},
		},
		presetNoisy: tooltipContent{
			title:       "Noisy Preset",
			description: "Stress test with high noise to evaluate robustness.",
			rangeInfo:   "30 levels, 15% noise, 6-bit ADC",
			physics:     "Represents challenging conditions: high variability, poor retention, or degraded devices.",
			tips: []string{
				"Expect visible accuracy degradation",
				"Some digits become unreliable",
				"Useful for robustness testing",
			},
		},
		confidence: tooltipContent{
			title:       "Prediction Confidence",
			description: "Softmax probability distribution over the 10 digit classes.",
			rangeInfo:   "0-100% for each class (sums to 100%)",
			physics:     "Softmax normalizes raw output scores to probabilities. Higher confidence = more certain prediction.",
			tips: []string{
				"High confidence on correct class = reliable",
				"Similar scores across classes = uncertain",
				"Noise typically reduces confidence margins",
			},
		},
		comparison: tooltipContent{
			title:       "FP32 vs CIM Comparison",
			description: "Side-by-side comparison of floating-point and compute-in-memory inference.",
			rangeInfo:   "Agreement rate typically 95-99%",
			physics:     "FP32 uses 32-bit floating point (ideal). CIM uses quantized weights + analog noise. Differences show hardware effects.",
			tips: []string{
				"Green = both agree on prediction",
				"Red = predictions differ (hardware error)",
				"Disagreement increases with noise/fewer levels",
			},
		},
		energy: tooltipContent{
			title:       "Energy Consumption",
			description: "Estimated energy per inference for different compute platforms.",
			rangeInfo:   "fJ to mJ per inference (varies by 10⁶×)",
			physics:     "CIM energy: E = CV² per cell access. No data movement = orders of magnitude less than Von Neumann. GPU moves data through memory hierarchy = high energy.",
			tips: []string{
				"FeCIM: ~1 fJ/MAC (femtojoules per multiply-accumulate)",
				"GPU: ~1 pJ/MAC (1000× more)",
				"CPU: ~10 pJ/MAC (10000× more)",
			},
		},
	}
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
		if n == 0 {
			fyne.Do(func() {
				app.testProgressBar.Hide()
				app.testResultLabel.SetText("No test data available")
			})
			return
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

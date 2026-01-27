// Package gui provides Fyne-based GUI components for MNIST visualization.
// dualmode_inference.go provides inference execution and result display functionality.
package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"fecim-lattice-tools/module3-mnist/pkg/core"
	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// onDigitChanged handles canvas drawing updates.
func (app *DualModeApp) onDigitChanged(pixels []float64) {
	app.lastPixels = pixels
	if app.animationEnabled {
		go app.runInferenceAnimated(pixels)
	} else {
		app.runInference(pixels)
	}
}

// runInference runs dual-path inference and updates the UI.
func (app *DualModeApp) runInference(pixels []float64) {
	result := app.network.Infer(pixels)

	// Get quantized weights for P1.1 visualization
	quantWeights, _, _, _ := app.network.GetQuantWeights()

	fyne.Do(func() {
		// Update status line (not in updateResultDisplays since it's specific to runInference)
		app.statusLabel.SetText(fmt.Sprintf("FP: %d (%.1f%%) | CIM: %d (%.1f%%) | %s",
			result.FPPrediction, result.FPConfidence*100,
			result.CIMPrediction, result.CIMConfidence*100,
			map[bool]string{true: "MATCH", false: "Prediction Mismatch"}[result.Agree]))

		// Update all result displays
		app.updateResultDisplays(result, quantWeights)
	})
}

// runInferenceAnimated performs inference with visual animation phases.
// This shows the data flow through the network for educational purposes.
func (app *DualModeApp) runInferenceAnimated(pixels []float64) {
	// Phase 1: Input Processing (150ms)
	fyne.Do(func() {
		if app.inferencePhaseLabel != nil {
			app.inferencePhaseLabel.SetText("Phase 1: Processing input (784 pixels)...")
		}
		app.statusLabel.SetText("INFERENCE | Phase 1: Input → 784 neurons")
	})
	time.Sleep(150 * time.Millisecond)

	// Phase 2: Layer 1 MVM (200ms)
	fyne.Do(func() {
		if app.inferencePhaseLabel != nil {
			app.inferencePhaseLabel.SetText("Phase 2: Layer 1 MVM (784×128 = 100,352 MACs)...")
		}
		app.statusLabel.SetText("INFERENCE | Phase 2: MVM 784→128 (100,352 parallel MACs)")
	})
	time.Sleep(200 * time.Millisecond)

	// Phase 3: Layer 2 MVM (150ms)
	fyne.Do(func() {
		if app.inferencePhaseLabel != nil {
			app.inferencePhaseLabel.SetText("Phase 3: Layer 2 MVM (128×10 = 1,280 MACs)...")
		}
		app.statusLabel.SetText("INFERENCE | Phase 3: MVM 128→10 (1,280 parallel MACs)")
	})
	time.Sleep(150 * time.Millisecond)

	// Phase 4: Result - run actual inference and display
	result := app.network.Infer(pixels)
	quantWeights, _, _, _ := app.network.GetQuantWeights()

	fyne.Do(func() {
		if app.inferencePhaseLabel != nil {
			app.inferencePhaseLabel.SetText("Complete! 101,632 MACs in ~5µs")
		}

		// Update all displays (same as runInference)
		app.updateResultDisplays(result, quantWeights)

		// Show dramatic match/mismatch feedback (UI-024 fix: better units, UI-023 fix: clearer wording)
		if result.Agree {
			app.statusLabel.SetText(fmt.Sprintf("MATCH | FP: %d | CIM: %d | Confidence: %.1f%% | Energy Efficiency: 10,000× improvement",
				result.FPPrediction, result.CIMPrediction, result.CIMConfidence*100))
		} else {
			app.statusLabel.SetText(fmt.Sprintf("Prediction Mismatch | FP: %d vs CIM: %d | Weight quantization may need tuning",
				result.FPPrediction, result.CIMPrediction))
		}
	})
}

// updateResultDisplays updates all UI elements with inference results.
// Extracted from runInference to avoid duplication.
func (app *DualModeApp) updateResultDisplays(result *core.InferenceResult, quantWeights [][]float64) {
	// Update FP results (legacy)
	app.fpPredLabel.SetText(fmt.Sprintf("Prediction: %d (%.1f%%)", result.FPPrediction, result.FPConfidence*100))
	app.fpConfBar.SetValue(result.FPConfidence)

	// Update CIM results (legacy)
	app.cimPredLabel.SetText(fmt.Sprintf("Prediction: %d (%.1f%%)", result.CIMPrediction, result.CIMConfidence*100))
	app.cimConfBar.SetValue(result.CIMConfidence)

	// Update agreement (legacy)
	if result.Agree {
		app.agreementLabel.SetText("PREDICTIONS MATCH")
	} else {
		app.agreementLabel.SetText(fmt.Sprintf("DISAGREEMENT (KL=%.3f)", result.Disagreement))
	}

	// Update probability bars (legacy)
	for i := 0; i < 10; i++ {
		app.fpProbBars[i].SetValue(result.FPProbabilities[i])
		app.cimProbBars[i].SetValue(result.CIMProbabilities[i])
	}

	// Update energy (legacy) - UI-024 fix: clearer units and wording
	gpuEnergy := result.EnergyUsed * EnergyRatioGPU
	app.energyLabel.SetText(fmt.Sprintf("Energy: %.2f µJ (FeCIM) vs %.0f mJ (GPU) = %.0f× improvement",
		result.EnergyUsed, gpuEnergy/1000, float64(EnergyRatioGPU)))

	// P1 Enhancements
	if app.quantizationWidget != nil && len(quantWeights) > 0 {
		app.quantizationWidget.SetNumLevels(app.network.GetNumLevels())
		app.quantizationWidget.UpdateWithWeights(quantWeights, 5)
	}

	if app.comparisonCard != nil {
		compResult := &ComparisonResult{
			FPPrediction:     result.FPPrediction,
			FPConfidence:     result.FPConfidence,
			FPProbabilities:  result.FPProbabilities,
			CIMPrediction:    result.CIMPrediction,
			CIMConfidence:    result.CIMConfidence,
			CIMProbabilities: result.CIMProbabilities,
			Match:            result.Agree,
			ConfidenceDelta:  result.FPConfidence - result.CIMConfidence,
			EnergyFeCIM:      result.EnergyUsed * 1e6,
			EnergyGPU:        result.EnergyUsed * 1e6 * EnergyRatioGPU,
			EnergyRatio:      float64(EnergyRatioGPU),
		}
		if compResult.ConfidenceDelta < 0 {
			compResult.ConfidenceDelta = -compResult.ConfidenceDelta
		}
		app.comparisonCard.SetResult(compResult)
	}

	if app.dualProbabilityChart != nil {
		app.dualProbabilityChart.SetProbabilities(
			result.FPProbabilities,
			result.CIMProbabilities,
			result.FPPrediction,
			result.CIMPrediction,
		)
	}

	if app.energyWidget != nil {
		app.energyWidget.RecordInference()
	}
}

// resetResults clears the result displays.
func (app *DualModeApp) resetResults() {
	app.lastPixels = nil
	app.fpPredLabel.SetText("Prediction: -")
	app.fpConfBar.SetValue(0)
	app.cimPredLabel.SetText("Prediction: -")
	app.cimConfBar.SetValue(0)
	app.agreementLabel.SetText("")
	app.energyLabel.SetText("Energy: -")
	for i := 0; i < 10; i++ {
		app.fpProbBars[i].SetValue(0)
		app.cimProbBars[i].SetValue(0)
	}
	app.statusLabel.SetText("Ready. Draw a digit or load a test sample to see FeCIM in action.")

	// Clear P1 widgets
	if app.quantizationWidget != nil {
		app.quantizationWidget.Clear()
	}
	if app.comparisonCard != nil {
		app.comparisonCard.Clear()
	}
	if app.dualProbabilityChart != nil {
		app.dualProbabilityChart.Clear()
	}
	// Note: Energy widget is not cleared to keep session totals
}

// loadRandomSample loads a random test sample.
func (app *DualModeApp) loadRandomSample() {
	if len(app.testImages) == 0 {
		app.loadTestData()
		if len(app.testImages) == 0 {
			fyne.Do(func() {
				app.statusLabel.SetText("No test data available")
			})
			return
		}
	}

	idx := int(time.Now().UnixNano() % int64(len(app.testImages)))
	pixels := app.testImages[idx]
	label := app.testLabels[idx]

	fyne.Do(func() {
		app.digitCanvas.SetPixels(pixels)
		app.statusLabel.SetText(fmt.Sprintf("Loaded sample #%d (true label: %d)", idx, label))
		app.onDigitChanged(pixels)
	})
}

// loadTestData loads MNIST test data.
func (app *DualModeApp) loadTestData() {
	images, labels, err := mnist.LoadMNIST(app.dataDir, false) // false = test set
	if err != nil {
		mnistLog.Printf("Failed to load MNIST test data: %v, using synthetic data", err)
		app.testImages, app.testLabels = generateSyntheticData(200)
		// Notify user that we're using synthetic data
		fyne.Do(func() {
			app.statusLabel.SetText("Using synthetic test data (MNIST not found)")
		})
		return
	}

	if len(images) > 1000 {
		app.testImages = images[:1000]
		app.testLabels = labels[:1000]
	} else {
		app.testImages = images
		app.testLabels = labels
	}
}

// changeHiddenSize performs hidden size change with loading feedback
func (app *DualModeApp) changeHiddenSize(size int) {
	// Map size to weight file
	weightsFile := fmt.Sprintf("pretrained_30_h%d.json", size)
	weightsPath := filepath.Join(app.dataDir, weightsFile)

	// Show loading progress
	fyne.Do(func() {
		app.statusLabel.SetText(fmt.Sprintf("Loading weights for hidden size %d...", size))
	})

	// Check if file exists, fallback to default
	if _, err := os.Stat(weightsPath); os.IsNotExist(err) {
		// Try default weights file
		weightsPath = filepath.Join(app.dataDir, "pretrained_weights.json")
		fyne.Do(func() {
			app.statusLabel.SetText(fmt.Sprintf("Note: Using default weights (h%d weights not found)", size))
		})
	}

	err := app.network.LoadWeights(weightsPath)
	if err != nil {
		fyne.Do(func() {
			app.statusLabel.SetText(fmt.Sprintf("Error changing hidden size: %v", err))
			if app.window != nil {
				dialog.ShowError(fmt.Errorf("Failed to change hidden size: %w", err), app.window)
			}
		})
		return
	}

	app.updateWeightHeatmap()

	if len(app.lastPixels) > 0 {
		app.runInference(app.lastPixels)
	}

	fyne.Do(func() {
		app.statusLabel.SetText(fmt.Sprintf("Loaded network with hidden size %d", size))
	})
}

// updateWeightHeatmapWithProgress updates weight visualization with loading feedback.
// The done channel is used to signal completion; send nil for success, or an error.
func (app *DualModeApp) updateWeightHeatmapWithProgress(done chan error) {
	// Helper to report errors
	reportError := func(err error) {
		fyne.Do(func() {
			app.statusLabel.SetText(fmt.Sprintf("Error updating heatmap: %v", err))
			if app.window != nil {
				dialog.ShowError(fmt.Errorf("Failed to update weight heatmap: %w", err), app.window)
			}
		})
	}

	if !app.initialized {
		done <- nil
		return
	}

	// Show loading progress
	fyne.Do(func() {
		app.statusLabel.SetText("Updating weight heatmap...")
	})

	// Update heatmap if it exists
	var updateErr error
	if app.weightHeatmap != nil {
		fyne.Do(func() {
			app.weightHeatmap.Refresh()
		})
	}

	// Report any error that occurred
	if updateErr != nil {
		reportError(updateErr)
	}

	// Signal completion
	done <- updateErr
}

// tryLoadQATWeights attempts to load QAT weights optimized for the given level.
// Only reloads if the optimal weights are different from currently loaded.
func (app *DualModeApp) tryLoadQATWeights(targetLevel int) {
	// Check if we already have optimal weights loaded (thread-safe read)
	app.currentQATLevelMu.RLock()
	currentLevel := app.currentQATLevel
	app.currentQATLevelMu.RUnlock()
	if currentLevel == targetLevel {
		return
	}

	// Find the weights file for this level
	weightsPath := core.GetWeightsFilename(app.dataDir, targetLevel)

	// Check if the file exists
	if _, err := os.Stat(weightsPath); os.IsNotExist(err) {
		// No level-specific weights, notify user (but only once per level per session)
		app.warnedMissingLevelsMu.RLock()
		alreadyWarned := app.warnedMissingLevels[targetLevel]
		app.warnedMissingLevelsMu.RUnlock()

		if !alreadyWarned {
			app.warnedMissingLevelsMu.Lock()
			app.warnedMissingLevels[targetLevel] = true
			app.warnedMissingLevelsMu.Unlock()

			fyne.Do(func() {
				if app.statusLabel != nil {
					app.statusLabel.SetText(fmt.Sprintf("Note: No QAT weights for %d levels - using existing weights", targetLevel))
				}
				if app.window != nil {
					// Show a non-blocking notification instead of modal dialog
					// User can train weights using the "Train Weights" button
					dialog.ShowInformation("Weights Not Found",
						fmt.Sprintf("QAT weights for %d levels not found.\n\nUsing existing weights instead.\n\nTo train weights for this level, use Expert Mode → Train Weights.", targetLevel),
						app.window)
				}
			})
		} else {
			// Already warned about this level, just update status silently
			fyne.Do(func() {
				if app.statusLabel != nil {
					app.statusLabel.SetText(fmt.Sprintf("Using existing weights (no QAT weights for %d levels)", targetLevel))
				}
			})
		}
		return
	}

	// Load the new weights
	if err := app.network.LoadWeights(weightsPath); err != nil {
		// Failed to load, notify user
		fyne.Do(func() {
			if app.statusLabel != nil {
				app.statusLabel.SetText(fmt.Sprintf("Error: Failed to load QAT weights for %d levels: %v", targetLevel, err))
			}
			if app.window != nil {
				dialog.ShowError(fmt.Errorf("Failed to load QAT weights for %d levels: %w", targetLevel, err), app.window)
			}
		})
		return
	}

	// Update tracking (thread-safe write)
	app.currentQATLevelMu.Lock()
	app.currentQATLevel = targetLevel
	app.currentQATLevelMu.Unlock()

	// Update status
	fyne.Do(func() {
		if app.statusLabel != nil {
			app.statusLabel.SetText(fmt.Sprintf("Loaded QAT weights for %d levels", targetLevel))
		}
	})
}

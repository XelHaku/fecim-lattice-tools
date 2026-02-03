// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains voltage rules UI components for the unified view.
package gui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ====================================================================================
// VOLTAGE RULES UI - Program-Verify Sequence, ISPP Animation, V/2 Overlay
// ====================================================================================

// AnimationFrameDelayMs is the animation frame delay for smooth updates
const AnimationFrameDelayMs = 50

// UI Colors for voltage visualization
var (
	colorFullVoltage   = color.RGBA{255, 200, 0, 255}   // Bright Gold for target cell
	colorHalfSelect    = color.RGBA{255, 165, 0, 255}   // Amber for V/2 cells
	colorZeroVoltage   = color.RGBA{50, 50, 60, 255}    // Dim Gray for inactive
	colorAscending     = color.RGBA{100, 220, 120, 255} // Green for ascending
	colorDescending    = color.RGBA{220, 100, 100, 255} // Red for descending
	colorPhaseActive   = color.RGBA{100, 200, 255, 255} // Cyan for active phase
	colorPhaseInactive = color.RGBA{80, 80, 90, 255}    // Dim for inactive phase
)

// drawWriteSequenceTimingDiagram draws the program-verify timing diagram
// Shows: RESET -> HOLD -> WRITE -> HOLD -> VERIFY with phase highlighting
func (ca *CircuitsApp) drawWriteSequenceTimingDiagram() fyne.CanvasObject {
	phaseInfo := ca.deviceState.GetWritePhaseInfo()

	// Phase labels with durations
	phases := []struct {
		name     string
		duration int
		phase    WritePhase
	}{
		{"RESET", PhaseResetDurationNs, PhaseReset},
		{"HOLD", PhaseHold1DurationNs, PhaseHold1},
		{"WRITE", PhaseWriteDurationNs, PhaseWrite},
		{"HOLD", PhaseHold2DurationNs, PhaseHold2},
		{"VERIFY", PhaseVerifyDurationNs, PhaseVerify},
	}

	phaseBoxes := container.NewHBox()

	for _, p := range phases {
		// Choose color based on current phase
		bgColor := colorPhaseInactive
		if phaseInfo.Active && phaseInfo.Phase == p.phase {
			bgColor = colorPhaseActive
		}

		// Create phase box with label and duration
		label := canvas.NewText(fmt.Sprintf("%s\n%dns", p.name, p.duration), color.White)
		label.Alignment = fyne.TextAlignCenter
		label.TextSize = 10

		bg := canvas.NewRectangle(bgColor)
		bg.SetMinSize(fyne.NewSize(60, 40))

		phaseBox := container.NewStack(bg, container.NewCenter(label))
		phaseBoxes.Add(phaseBox)
	}

	// Progress bar
	progress := widget.NewProgressBar()
	progress.SetValue(phaseInfo.Progress)
	progress.Min = 0
	progress.Max = 1

	// Voltage label showing current applied phase voltage
	voltageLabel := widget.NewLabel(fmt.Sprintf("Phase Voltage: %.2fV", phaseInfo.PhaseVoltage))
	voltageLabel.TextStyle = fyne.TextStyle{Monospace: true}

	return container.NewVBox(
		widget.NewLabelWithStyle("Program-Verify Sequence", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		phaseBoxes,
		progress,
		voltageLabel,
	)
}

// animateWriteSequence runs the program-verify animation
// This is called from a goroutine
func (ca *CircuitsApp) animateWriteSequence() {
	for {
		// Check for stop signal
		if ca.shouldStop() {
			return
		}

		if ca.deviceState == nil {
			return
		}

		phaseInfo := ca.deviceState.GetWritePhaseInfo()
		if !phaseInfo.Active {
			return
		}

		// Get phase duration for timing
		duration := GetPhaseDuration(phaseInfo.Phase)

		// Update UI
		ca.updateWriteSequenceUI()

		// Delay proportional to phase duration (scaled for animation)
		animDelayMs := duration / 4
		if animDelayMs < AnimationFrameDelayMs {
			animDelayMs = AnimationFrameDelayMs
		}
		if ca.sleep(animDelayMs) {
			return // Interrupted
		}

		// Advance to next phase
		complete := ca.deviceState.AdvanceWritePhase()
		if complete {
			ca.updateWriteSequenceUI()
			return
		}
	}
}

// updateWriteSequenceUI refreshes the program-verify timing display
func (ca *CircuitsApp) updateWriteSequenceUI() {
	fyne.Do(func() {
		if ca.writeSequencePanel != nil {
			// Rebuild the timing diagram with current state
			newDiagram := ca.drawWriteSequenceTimingDiagram()
			ca.writeSequencePanel.Objects = []fyne.CanvasObject{newDiagram}
			ca.writeSequencePanel.Refresh()
		}
	})
}

// runISPPWithAnimation runs the ISPP loop with visual feedback
// This ENHANCES the existing writeReadVerifyLoop() by adding:
// - Program-verify sequence animation within each iteration
// - Calibrated per-level voltage lookup
// - Hysteresis direction tracking
// - V/2 visualization for 0T1R mode
func (ca *CircuitsApp) runISPPWithAnimation(row, col, targetLevel int) {
	const iterationDelay = 300 * time.Millisecond

	// Get current level
	ca.mu.Lock()
	currentLevel := 0
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		currentLevel = ca.arrayWeights[row][col]
	}
	ca.mu.Unlock()

	// Determine direction
	direction := ca.deviceState.GetWriteDirection(row, col, currentLevel, targetLevel)
	ascending := direction == DirectionAscending

	// Enable V/2 visualization if in passive (0T1R) mode
	if ca.deviceState.IsPassiveMode() {
		voltage := ca.deviceState.GetVoltageForLevel(targetLevel, ascending)
		ca.deviceState.EnableHalfSelectVisualization(row, col, voltage)
		ca.updateHalfSelectVisualization()
	}

	// Start ISPP
	ca.deviceState.StartISPP(row, col, targetLevel, currentLevel)

	for {
		// Check for stop signal at start of each iteration
		if ca.shouldStop() {
			ca.deviceState.CancelISPP()
			return
		}

		isppStatus := ca.deviceState.GetISPPStatus()
		if !isppStatus.Active {
			break
		}

		// Start program-verify sequence for this iteration
		ca.deviceState.StartWriteSequence(row, col, targetLevel, currentLevel)
		go ca.animateWriteSequence()

		// Update ISPP UI
		ca.updateISPPUI()

		// Wait for program-verify sequence to complete
		for {
			if ca.shouldStop() {
				ca.deviceState.CancelWriteSequence()
				ca.deviceState.CancelISPP()
				return
			}
			phaseInfo := ca.deviceState.GetWritePhaseInfo()
			if !phaseInfo.Active {
				break
			}
			if ca.sleep(AnimationFrameDelayMs) {
				ca.deviceState.CancelWriteSequence()
				ca.deviceState.CancelISPP()
				return
			}
		}

		// Simulate write: move at most one level per pulse toward the calibrated level
		ascending := isppStatus.Direction == DirectionAscending
		estimatedLevel := ca.deviceState.GetLevelForVoltage(isppStatus.Voltage, ascending)
		nextLevel := currentLevel
		if ascending {
			if estimatedLevel > currentLevel {
				nextLevel = currentLevel + 1
				if nextLevel > targetLevel {
					nextLevel = targetLevel
				}
			}
		} else if estimatedLevel < currentLevel {
			nextLevel = currentLevel - 1
			if nextLevel < targetLevel {
				nextLevel = targetLevel
			}
		}
		ca.mu.Lock()
		if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
			currentLevel = nextLevel
			ca.arrayWeights[row][col] = currentLevel
		}
		ca.mu.Unlock()

		// ISPP iteration with verification
		result := ca.deviceState.ISPPIterate(currentLevel)

		switch result {
		case ISPPResultVerified:
			fyne.Do(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("SUCCESS [%d,%d] = Level %d (%d iterations)",
					row, col, targetLevel, isppStatus.Iteration))
			})
			goto cleanup

		case ISPPResultOvershoot:
			fyne.Do(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("OVERSHOOT [%d,%d] - Resetting to saturation...", row, col))
			})
			ca.deviceState.HandleOvershoot(row, col)
			// Reset local currentLevel based on direction
			if direction == DirectionAscending {
				currentLevel = 0
			} else {
				currentLevel = ca.quantLevels - 1
			}
			ca.mu.Lock()
			if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
				ca.arrayWeights[row][col] = currentLevel
			}
			ca.mu.Unlock()
			ca.recomputeAndRefresh()
			time.Sleep(iterationDelay)
			continue

		case ISPPResultMaxIterations:
			fyne.Do(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("PARTIAL [%d,%d] = Level %d (target was %d)",
					row, col, currentLevel, targetLevel))
			})
			goto cleanup
		}

		// Continue - update UI and wait
		time.Sleep(iterationDelay)
	}

cleanup:
	// Disable V/2 visualization
	ca.deviceState.DisableHalfSelectVisualization()
	ca.updateHalfSelectVisualization()

	// Record the write in hysteresis state
	ca.deviceState.RecordWrite(row, col, currentLevel)

	ca.recomputeAndRefresh()
}

// updateISPPUI refreshes the ISPP status display
func (ca *CircuitsApp) updateISPPUI() {
	isppStatus := ca.deviceState.GetISPPStatus()

	// Direction indicator
	dirStr := "^ Ascending"
	dirColor := colorAscending
	if isppStatus.Direction == DirectionDescending {
		dirStr = "v Descending"
		dirColor = colorDescending
	}

	fyne.Do(func() {
		if ca.operationsStatusLabel != nil {
			ca.operationsStatusLabel.SetText(fmt.Sprintf("ISPP [%d/%d]: Level %d -> %d | V=%.2fV | %s",
				isppStatus.Iteration, isppStatus.MaxIter,
				isppStatus.CurrentLevel, isppStatus.TargetLevel,
				isppStatus.Voltage, dirStr))
		}

		// Update direction indicator if we have one
		if ca.hysteresisDirectionLabel != nil {
			ca.hysteresisDirectionLabel.SetText(dirStr)
			// Note: Fyne doesn't support dynamic text colors easily, so we just update text
		}
		_ = dirColor // Use in future canvas-based indicator
	})
}

// drawHalfSelectOverlay draws V/2 voltage indicators on the array
// Target cell: Bright Gold (full voltage)
// Half-selected cells: Amber (V/2)
func (ca *CircuitsApp) drawHalfSelectOverlay(arrayCanvas fyne.CanvasObject) {
	hsState := ca.deviceState.GetHalfSelectState()
	if !hsState.Enabled {
		return
	}

	// This would overlay colored rectangles on the array canvas
	// For now, we update the array refresh to include V/2 coloring
	// The actual drawing happens in refreshArrayHeatmap or similar
}

// updateHalfSelectVisualization enables/disables V/2 overlay
func (ca *CircuitsApp) updateHalfSelectVisualization() {
	hsState := ca.deviceState.GetHalfSelectState()

	fyne.Do(func() {
		if ca.halfSelectIndicator != nil {
			if hsState.Enabled {
				ca.halfSelectIndicator.SetText(fmt.Sprintf("V/2 Bias Active | Full: %.2fV | Half: %.2fV",
					hsState.FullVoltage, hsState.HalfVoltage))
				ca.halfSelectIndicator.Show()
			} else {
				ca.halfSelectIndicator.Hide()
			}
		}
	})

	// Trigger array refresh to update cell colors
	ca.recomputeAndRefresh()
}

// getHalfSelectCellColor returns the color for a cell based on half-select state
// Used by array visualization to color cells during V/2 mode
func (ca *CircuitsApp) getHalfSelectCellColor(row, col int) (color.Color, bool) {
	hsState := ca.deviceState.GetHalfSelectState()
	if !hsState.Enabled {
		return nil, false
	}

	// Target cell gets full voltage color
	if row == hsState.SelectedRow && col == hsState.SelectedCol {
		return colorFullVoltage, true
	}

	// Check if half-selected
	if ca.deviceState.IsHalfSelected(row, col) {
		return colorHalfSelect, true
	}

	return nil, false
}

// createPassiveVoltagePanel creates the V/2 panel for 0T1R (passive) mode
func (ca *CircuitsApp) createPassiveVoltagePanel() fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle("Passive Crossbar (0T1R)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	infoText := `In passive crossbar mode, V/2 biasing is required:
- Target cell receives full write voltage
- Same-row and same-column cells receive V/2
- This prevents unintended state changes

Watch for disturb effects on half-selected cells.`

	infoLabel := widget.NewLabel(infoText)
	infoLabel.Wrapping = fyne.TextWrapWord

	// V/2 indicator (updated during write operations)
	ca.halfSelectIndicator = widget.NewLabel("V/2 Bias: Inactive")
	ca.halfSelectIndicator.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		infoLabel,
		ca.halfSelectIndicator,
	)
}

// createActiveVoltagePanel creates the panel for 1T1R/2T1R (active) mode
func (ca *CircuitsApp) createActiveVoltagePanel() fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle("Active Cell Access (1T1R/2T1R)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	infoText := `In active transistor mode:
- Transistors isolate non-selected cells
- Only the target cell sees write voltage
- No V/2 disturb effects
- Higher area overhead but cleaner writes`

	infoLabel := widget.NewLabel(infoText)
	infoLabel.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		infoLabel,
	)
}

// createCompactPassivePanel creates a single-line info for passive mode
func (ca *CircuitsApp) createCompactPassivePanel() fyne.CanvasObject {
	// V/2 indicator (updated during write operations)
	ca.halfSelectIndicator = widget.NewLabel("0T1R: V/2 scheme - sneak currents add 5-20% error")
	ca.halfSelectIndicator.TextStyle = fyne.TextStyle{Italic: true}
	return ca.halfSelectIndicator
}

// createCompactActivePanel creates a single-line info for active mode
func (ca *CircuitsApp) createCompactActivePanel() fyne.CanvasObject {
	label := widget.NewLabel("1T1R/2T1R: Transistors isolate cells - no sneak currents")
	label.TextStyle = fyne.TextStyle{Italic: true}
	return label
}

// createCompactWritePanel creates a compact write panel without the program-verify sequence UI
func (ca *CircuitsApp) createCompactWritePanel() fyne.CanvasObject {
	maxLevel := ca.quantLevels - 1
	midLevel := ca.quantLevels / 2

	ca.mfuxWriteLevelSlider = widget.NewSlider(0, float64(maxLevel))
	ca.mfuxWriteLevelSlider.Step = 1
	ca.mfuxWriteLevelSlider.Value = float64(midLevel)
	ca.mfuxWriteLevelSlider.OnChanged = func(v float64) {
		ca.onWriteLevelChanged(int(v))
		ca.updateHysteresisDirectionUI(int(v))
	}

	ca.mfuxWriteLevelLabel = widget.NewLabel(fmt.Sprintf("Level: %d", midLevel))
	ca.mfuxWriteLevelLabel.TextStyle = fyne.TextStyle{Monospace: true}

	ca.mfuxWriteVoltageLabel = widget.NewLabel("V: 1.00")
	ca.mfuxWriteVoltageLabel.TextStyle = fyne.TextStyle{Monospace: true}

	ca.mfuxWriteTargetLabel = widget.NewLabel("Target: [0,0]")
	ca.mfuxWriteTargetLabel.TextStyle = fyne.TextStyle{Bold: true}

	ca.hysteresisDirectionLabel = widget.NewLabel("^")
	ca.hysteresisDirectionLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Single row: Target | Slider | Level | Voltage | Direction
	return container.NewBorder(nil, nil,
		container.NewHBox(ca.mfuxWriteTargetLabel, widget.NewLabel("Write:")),
		container.NewHBox(ca.mfuxWriteLevelLabel, ca.mfuxWriteVoltageLabel, ca.hysteresisDirectionLabel),
		ca.mfuxWriteLevelSlider,
	)
}

// updateArchitectureSpecificUI shows/hides panels based on architecture
func (ca *CircuitsApp) updateArchitectureSpecificUI() {
	isPassive := ca.deviceState.IsPassiveMode()

	fyne.Do(func() {
		if ca.passiveVoltagePanel != nil && ca.activeVoltagePanel != nil {
			if isPassive {
				ca.passiveVoltagePanel.Show()
				ca.activeVoltagePanel.Hide()
			} else {
				ca.passiveVoltagePanel.Hide()
				ca.activeVoltagePanel.Show()
			}
		}
	})
}

// updateHysteresisDirectionUI updates the direction indicator
func (ca *CircuitsApp) updateHysteresisDirectionUI(targetLevel int) {
	if ca.deviceState == nil {
		return
	}

	row := ca.deviceState.GetSelectedRow()
	col := ca.deviceState.GetSelectedCol()

	// Get current level
	ca.mu.RLock()
	currentLevel := 0
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		currentLevel = ca.arrayWeights[row][col]
	}
	ca.mu.RUnlock()

	direction := ca.deviceState.GetWriteDirection(row, col, currentLevel, targetLevel)

	dirStr := "- Unknown"
	if direction == DirectionAscending {
		dirStr = "^ Ascending"
	} else if direction == DirectionDescending {
		dirStr = "v Descending"
	}

	fyne.Do(func() {
		if ca.hysteresisDirectionLabel != nil {
			ca.hysteresisDirectionLabel.SetText(dirStr)
		}
	})
}

// createEnhancedWriteModePanel creates an enhanced write panel with voltage rules UI
// This adds hysteresis direction and program-verify sequence display to the existing write panel
func (ca *CircuitsApp) createEnhancedWriteModePanel() fyne.CanvasObject {
	// Existing write panel components
	maxLevel := ca.quantLevels - 1
	midLevel := ca.quantLevels / 2

	ca.mfuxWriteLevelSlider = widget.NewSlider(0, float64(maxLevel))
	ca.mfuxWriteLevelSlider.Step = 1
	ca.mfuxWriteLevelSlider.Value = float64(midLevel)
	ca.mfuxWriteLevelSlider.OnChanged = func(v float64) {
		ca.onWriteLevelChanged(int(v))
		ca.updateHysteresisDirectionUI(int(v))
	}

	// Level label with min/max indicators
	minLabel := widget.NewLabel("0")
	maxLabel := widget.NewLabel(fmt.Sprintf("%d", maxLevel))
	ca.mfuxWriteLevelLabel = widget.NewLabel(fmt.Sprintf("Level: %d", midLevel))
	ca.mfuxWriteLevelLabel.TextStyle = fyne.TextStyle{Monospace: true}

	ca.mfuxWriteVoltageLabel = widget.NewLabel("Voltage: 1.00V")
	ca.mfuxWriteVoltageLabel.TextStyle = fyne.TextStyle{Monospace: true}

	ca.mfuxWriteTargetLabel = widget.NewLabel("Target: Row 0, Col 0")
	ca.mfuxWriteTargetLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Hysteresis direction indicator
	ca.hysteresisDirectionLabel = widget.NewLabel("- Unknown")
	ca.hysteresisDirectionLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Layout
	titleLabel := widget.NewLabelWithStyle("Target Write Level:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	headerRow := container.NewHBox(
		titleLabel,
		layout.NewSpacer(),
		ca.mfuxWriteTargetLabel,
		widget.NewSeparator(),
		ca.hysteresisDirectionLabel,
	)

	sliderWithMinMax := container.NewBorder(nil, nil,
		minLabel,
		maxLabel,
		ca.mfuxWriteLevelSlider,
	)

	valueRow := container.NewHBox(
		ca.mfuxWriteLevelLabel,
		layout.NewSpacer(),
		ca.mfuxWriteVoltageLabel,
	)

	// Program-verify sequence container (populated during write)
	ca.writeSequencePanel = container.NewStack()

	return container.NewVBox(
		headerRow,
		sliderWithMinMax,
		valueRow,
		widget.NewSeparator(),
		ca.writeSequencePanel,
	)
}

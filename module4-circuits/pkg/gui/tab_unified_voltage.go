// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains voltage rules UI components for the unified view.
package gui

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	sharedphysics "fecim-lattice-tools/shared/physics"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// ====================================================================================
// VOLTAGE RULES UI - Program-Verify Sequence, ISPP Animation, V/2 Overlay
// ====================================================================================

// AnimationFrameDelayMs is the animation frame delay for smooth updates
const AnimationFrameDelayMs = 50

// Half-select disturb modeling (simple, deterministic for live UI updates)
const (
	halfSelectDisturbThresholdRatio = 0.3  // Vhalf/Vc threshold before any drift
	halfSelectDisturbRate           = 0.25 // Fractional level change per pulse at full strength
)

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
	if phaseInfo.Phase == PhaseWrite {
		isppStatus := ca.deviceState.GetISPPStatus()
		if isppStatus.Active {
			phaseInfo.PhaseVoltage = isppStatus.Voltage
		}
	}

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
		label.TextSize = 14

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

		ca.applyWritePhaseVoltages(phaseInfo)

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
			// Reset voltages after sequence completes
			finalPhase := ca.deviceState.GetWritePhaseInfo()
			ca.applyWritePhaseVoltages(finalPhase)
			ca.updateWriteSequenceUI()
			return
		}
	}
}

// updateWriteSequenceUI refreshes the program-verify timing display
func (ca *CircuitsApp) updateWriteSequenceUI() {
	sharedwidgets.SafeDo(func() {
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
	defer ca.setProgrammingActive(false)
	if ca.deviceState != nil && ca.deviceState.GetISPPEngine() == ISPPEngineLK {
		ca.runISPPWithLK(row, col, targetLevel)
		return
	}

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
		applied := ca.deviceState.AppliedWriteVoltageForLevel(targetLevel, ascending)
		ca.deviceState.EnableHalfSelectVisualization(row, col, applied)
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

		// Apply write voltage via DAC for this pulse and update half-select neighbors.
		appliedVoltage, _ := ca.applyWriteVoltages(row, col, isppStatus.Voltage)
		if ca.deviceState.IsPassiveMode() {
			ca.deviceState.EnableHalfSelectVisualization(row, col, appliedVoltage)
			ca.updateHalfSelectVisualization()
		}
		ca.applyHalfSelectDisturb(row, col)
		ca.recomputeAndRefresh()

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

		// Clear write voltages before verify/next iteration
		ca.deviceState.ResetWriteVoltages()

		// Simulate write using physics update from the *actual* coupled target-cell voltage
		// (post-DAC quantization + IR-drop/neighbor coupling), not the ideal pulse request.
		ascending := isppStatus.Direction == DirectionAscending
		effectiveV := ca.deviceState.GetEffectiveCellVoltage(row, col)
		if effectiveV == 0 {
			effectiveV = isppStatus.Voltage
		}
		nextLevel := ca.deviceState.programLevelFromCoupledVoltage(
			currentLevel,
			effectiveV,
			float64(PhaseWriteDurationNs)*1e-9,
			ca.quantLevels,
		)
		if ascending {
			if nextLevel < currentLevel {
				nextLevel = currentLevel
			}
			if nextLevel > targetLevel {
				nextLevel = targetLevel
			}
		} else {
			if nextLevel > currentLevel {
				nextLevel = currentLevel
			}
			if nextLevel < targetLevel {
				nextLevel = targetLevel
			}
		}
		changed := false
		ca.mu.Lock()
		if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
			currentLevel = nextLevel
			if ca.arrayWeights[row][col] != currentLevel {
				ca.arrayWeights[row][col] = currentLevel
				changed = true
			}
		}
		ca.mu.Unlock()

		if changed {
			logAction("ispp_step row=%d col=%d level=%d voltage=%.3fV",
				row, col, currentLevel, isppStatus.Voltage)
			ca.recomputeAndRefresh()
		}

		// ISPP iteration with verification
		result := ca.deviceState.ISPPIterate(currentLevel)

		switch result {
		case ISPPResultVerified:
			sharedwidgets.SafeDo(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("SUCCESS [%d,%d] = Level %d (%d iterations)",
					row, col, targetLevel, isppStatus.Iteration))
			})
			goto cleanup

		case ISPPResultOvershoot:
			sharedwidgets.SafeDo(func() {
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
			sharedwidgets.SafeDo(func() {
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
	ca.deviceState.ResetWriteVoltages()

	// Record the write in hysteresis state
	ca.deviceState.RecordWrite(row, col, currentLevel)

	ca.deviceState.ResetWriteVoltages()
	ca.recomputeAndRefresh()
}

func (ca *CircuitsApp) runISPPWithLK(row, col, targetLevel int) {
	const uiDelay = 150 * time.Millisecond

	ds := ca.deviceState
	if ds == nil {
		return
	}

	// Get current level
	ca.mu.RLock()
	currentLevel := 0
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		currentLevel = ca.arrayWeights[row][col]
	}
	ca.mu.RUnlock()

	// Determine direction
	direction := ds.GetWriteDirection(row, col, currentLevel, targetLevel)
	if direction == DirectionUnknown {
		if ca.operationsStatusLabel != nil {
			sharedwidgets.SafeDo(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("Already at target [%d,%d] = Level %d", row, col, targetLevel))
			})
		}
		return
	}

	// Enable V/2 visualization if in passive (0T1R) mode
	if ds.IsPassiveMode() {
		voltage := ds.GetVoltageForLevel(targetLevel, direction == DirectionAscending)
		ds.EnableHalfSelectVisualization(row, col, voltage)
		ca.updateHalfSelectVisualization()
	}

	mat := ds.GetMaterial()
	if mat == nil {
		mat = sharedphysics.FeCIMMaterial()
	}
	levels := ds.GetWriteRange().NumLevels
	if levels <= 0 {
		levels = ca.quantLevels
	}

	// Initialize solver state from current level.
	gmin, gmax := ds.conductanceBounds()
	currentG := ds.levelToConductance(currentLevel, levels)
	currentP := sharedphysics.ConductanceToPolarization(currentG, gmin, gmax, mat.Ps)

	solver := sharedphysics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.Temperature = 300
	solver.EnableNoise = false
	solver.UseNLS = false
	if !solver.UseMaterialAlpha {
		solver.UpdateParams()
	}
	solver.SetState(currentP)

	ctrl := sharedphysics.NewWriteController(solver, mat)
	ctrl.MaxIterations = ISPPMaxIterations
	ctrl.PulseWidth = mat.Tau
	if ctrl.PulseWidth <= 0 {
		ctrl.PulseWidth = 100e-9
	}
	ctrl.MaxVoltage = ds.GetWriteRange().Max
	if ctrl.MaxVoltage <= 0 {
		ctrl.MaxVoltage = 1.5
	}
	stepG := (gmax - gmin) / float64(levels-1)
	if stepG <= 0 {
		stepG = 1e-6
	}
	ctrl.Tolerance = stepG * 0.5

	ds.beginISPPTracking(row, col, targetLevel, currentLevel, direction, ctrl.MaxIterations)

	ctrl.EventHook = func(event sharedphysics.WriteEvent) {
		if ca.shouldStop() {
			return
		}
		level := ds.conductanceToLevel(event.CurrentG, levels)
		ds.updateISPPTracking(event.Attempt, math.Abs(event.VPulse), level)
		ca.updateISPPUI()

		// During write phases, drive the circuit path and visualize half-select behavior.
		if event.Phase == "Predict" || event.Phase == "BinarySearch" {
			appliedVoltage, _ := ca.applyWriteVoltages(row, col, event.VPulse)
			if ds.IsPassiveMode() {
				ds.EnableHalfSelectVisualization(row, col, appliedVoltage)
				ca.updateHalfSelectVisualization()
			}
			ca.applyHalfSelectDisturb(row, col)
			ca.recomputeAndRefresh()
		}

		// Verify phase uses a low read-like voltage for UI/circuit visualization.
		if event.Phase == "Verify" {
			verifyV := ds.GetReadRange().Max * 0.5
			ca.applyWritePhaseVoltages(WriteSequenceState{
				Phase:        PhaseVerify,
				TargetRow:    row,
				TargetCol:    col,
				PhaseVoltage: verifyV,
			})
		}

		// On verify/success, reflect the level update in the stored array weights.
		if event.Phase == "Verify" || event.Phase == "Success" {
			ca.mu.Lock()
			updated := false
			if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
				if ca.arrayWeights[row][col] != level {
					ca.arrayWeights[row][col] = level
					updated = true
				}
			}
			ca.mu.Unlock()

			if updated {
				logAction("lk_ispp_step row=%d col=%d level=%d voltage=%.3fV",
					row, col, level, math.Abs(event.VPulse))
				ca.recomputeAndRefresh()
			}
		}

		if ca.operationsStatusLabel != nil {
			msg := fmt.Sprintf("PROGRAMMING — controls locked | Landau-Khalatnikov ISPP [%d/%d]: Level %d -> %d | V=%.2fV | %s",
				event.Attempt, ctrl.MaxIterations, level, targetLevel, math.Abs(event.VPulse), event.Phase)
			sharedwidgets.SafeDo(func() {
				ca.operationsStatusLabel.SetText(msg)
			})
		}
		if event.Phase == "Verify" {
			time.Sleep(uiDelay)
		}
	}

	targetG := ds.levelToConductance(targetLevel, levels)
	attempts, success, overshoots := ctrl.WriteTargetWithReset(targetG, false)

	finalP := solver.GetState()
	finalG := sharedphysics.PolarizationToConductance(finalP, mat.Ps, gmin, gmax)
	finalLevel := ds.conductanceToLevel(finalG, levels)

	ds.endISPPTracking(success, finalLevel)

	ca.mu.Lock()
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		ca.arrayWeights[row][col] = finalLevel
	}
	ca.mu.Unlock()

	if ca.operationsStatusLabel != nil {
		status := "PARTIAL"
		if success {
			status = "SUCCESS"
		}
		sharedwidgets.SafeDo(func() {
			ca.operationsStatusLabel.SetText(fmt.Sprintf("%s [%d,%d] = Level %d | %d attempts | overshoots=%d",
				status, row, col, finalLevel, attempts, overshoots))
		})
	}

	// Disable V/2 visualization
	ds.DisableHalfSelectVisualization()
	ca.updateHalfSelectVisualization()
	ds.ResetWriteVoltages()

	// Record the write in hysteresis state
	ds.RecordWrite(row, col, finalLevel)

	ds.ResetWriteVoltages()
	ca.recomputeAndRefresh()
}

func (ca *CircuitsApp) applyWritePhaseVoltages(phaseInfo WriteSequenceState) {
	if ca.deviceState == nil {
		return
	}

	row := phaseInfo.TargetRow
	col := phaseInfo.TargetCol
	isPassive := ca.deviceState.IsPassiveMode()

	switch phaseInfo.Phase {
	case PhaseWrite:
		writeVoltage := phaseInfo.PhaseVoltage
		if isppStatus := ca.deviceState.GetISPPStatus(); isppStatus.Active {
			writeVoltage = isppStatus.Voltage
		}
		if isPassive {
			ca.deviceState.ApplyHalfSelectWrite(row, col, writeVoltage)
		} else {
			ca.deviceState.SetWLSingle(row)
			ca.deviceState.SetAllDACVoltages(0)
			ca.deviceState.SetDACVoltage(col, writeVoltage)
		}
		ca.deviceState.SetDACRangeMode(DACRangeWrite)
		neighborChanges := ca.applyHalfSelectDisturb(row, col)
		if neighborChanges > 0 {
			logAction("write_disturb rows=%d cols=%d changes=%d", ca.arrayRows, ca.arrayCols, neighborChanges)
		}

	case PhaseVerify:
		verifyVoltage := phaseInfo.PhaseVoltage
		if !isPassive {
			ca.deviceState.SetWLSingle(row)
		}
		ca.deviceState.SetAllDACVoltages(0)
		ca.deviceState.SetDACVoltage(col, verifyVoltage)
		ca.deviceState.SetDACRangeMode(DACRangeRead)

	default:
		ca.deviceState.ResetWriteVoltages()
	}

	ca.recomputeAndRefresh()
}

// applyHalfSelectDisturb updates disturb state during WRITE pulses.
//
// 0T1R architecture truth table (WRITE to target cell rt,ct):
//   Cell class                          | Expected Vcell | Disturb handling
//   ------------------------------------|----------------|---------------------------
//   Target (r=rt, c=ct)                 | full write V   | excluded (programmed path)
//   Same row only (r=rt, c!=ct)         | ~V/2           | half-select disturb
//   Same column only (r!=rt, c=ct)      | ~V/2           | half-select disturb
//   Neither same row nor same column    | ~0             | no half-select residue
//
// Note: 0T1R uses V/2 disturb on the full selected row + selected column.
func (ca *CircuitsApp) applyHalfSelectDisturb(targetRow, targetCol int) int {
	if ca.deviceState == nil {
		return 0
	}

	mat := ca.deviceState.GetMaterial()
	if mat == nil {
		mat = sharedphysics.FeCIMMaterial()
	}
	geom := ca.deviceState.GetCellGeometry().WithDefaults()
	if geom.Film.Thickness <= 0 || mat.Ps == 0 {
		return 0
	}

	pulseWidth := mat.Tau
	if pulseWidth <= 0 {
		pulseWidth = float64(PhaseWriteDurationNs) * 1e-9
	}

	gmin, gmax := ca.deviceState.conductanceBounds()
	levels := ca.quantLevels

	solver := sharedphysics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.Temperature = 300
	solver.EnableNoise = false
	solver.UseNLS = false
	if !solver.UseMaterialAlpha {
		solver.UpdateParams()
	}

	changes := 0

	ca.mu.Lock()
	defer ca.mu.Unlock()

	if len(ca.halfSelectResidue) != ca.arrayRows {
		ca.halfSelectResidue = make([][]float64, ca.arrayRows)
		for r := range ca.halfSelectResidue {
			ca.halfSelectResidue[r] = make([]float64, ca.arrayCols)
		}
	}
	targetVCell := ca.deviceState.GetWLVoltage(targetRow) - ca.deviceState.GetDACVoltage(targetCol)

	for r := 0; r < len(ca.arrayWeights); r++ {
		if !ca.deviceState.IsPassiveMode() && !ca.deviceState.IsRowActive(r) {
			continue
		}
		for c := 0; c < len(ca.arrayWeights[r]); c++ {
			if r == targetRow && c == targetCol {
				continue
			}

			var vCell float64
			if ca.deviceState.IsPassiveMode() {
				vCell = ca.deviceState.GetWLVoltage(r) - ca.deviceState.GetDACVoltage(c)
			} else {
				vCell = ca.deviceState.GetDACVoltage(c)
			}

			if vCell == 0 {
				continue
			}

			// Track accumulated half-select exposure during WRITE pulses so UI/test hooks
			// can observe disturb build-up even when physics solver changes are sub-level.
			// 0T1R: V/2 disturb on full row + column.
			if ca.deviceState.IsPassiveMode() && ((r == targetRow) != (c == targetCol)) {
				base := 0.01 * halfSelectDisturbRate
				delta := math.Copysign(base, targetVCell)
				ca.halfSelectResidue[r][c] += delta
			}

			level := ca.arrayWeights[r][c]
			conductance := ca.deviceState.levelToConductance(level, levels)
			polarization := sharedphysics.ConductanceToPolarization(conductance, gmin, gmax, mat.Ps)

			solver.SetState(polarization)
			eField := geom.Film.ElectricField(vCell)
			solver.Step(eField, pulseWidth)

			newP := solver.GetState()
			newG := sharedphysics.PolarizationToConductance(newP, mat.Ps, gmin, gmax)
			newLevel := ca.deviceState.conductanceToLevel(newG, levels)

			if newLevel < 0 {
				newLevel = 0
			}
			if newLevel >= levels {
				newLevel = levels - 1
			}
			if newLevel != level {
				ca.arrayWeights[r][c] = newLevel
				changes++
			}
		}
	}

	return changes
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

	sharedwidgets.SafeDo(func() {
		if ca.operationsStatusLabel != nil {
			appliedVoltage, dacCode := ca.deviceState.DACWriteVoltage(isppStatus.Voltage)
			voltageText := fmt.Sprintf("V=%.2fV", appliedVoltage)
			if dacCode >= 0 {
				voltageText = fmt.Sprintf("V=%.2fV (DAC %d)", appliedVoltage, dacCode)
			}

			trail := ""
			if n := len(isppStatus.History); n > 0 {
				start := 0
				if n > 6 {
					start = n - 6
				}
				trail = " | cycle "
				for i := start; i < n; i++ {
					if i > start {
						trail += "->"
					}
					trail += fmt.Sprintf("L%d", isppStatus.History[i].Level)
				}
			}

			ca.operationsStatusLabel.SetText(fmt.Sprintf("PROGRAMMING — controls locked | ISPP [%d/%d]: Level %d -> %d | %s | %s%s",
				isppStatus.Iteration, isppStatus.MaxIter,
				isppStatus.CurrentLevel, isppStatus.TargetLevel,
				voltageText, dirStr, trail))
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

	sharedwidgets.SafeDo(func() {
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

	// Target cell gets full-voltage highlight; neighbors get half-select highlight.
	if row == hsState.SelectedRow && col == hsState.SelectedCol {
		return colorFullVoltage, true
	}
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
	ca.halfSelectIndicator = widget.NewLabel("0T1R: V/2 on row+col")
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
	writeRow := container.NewBorder(nil, nil,
		container.NewHBox(ca.mfuxWriteTargetLabel, widget.NewLabel("Write:")),
		container.NewHBox(ca.mfuxWriteLevelLabel, ca.mfuxWriteVoltageLabel, ca.hysteresisDirectionLabel),
		ca.mfuxWriteLevelSlider,
	)

	engineSelect := widget.NewSelect([]string{"Preisach (Level-based)", "Landau-Khalatnikov (Physics ODE)"}, func(s string) {
		if s == "" || ca.deviceState == nil {
			return
		}
		engine := ISPPEngineLevel
		if s == "Landau-Khalatnikov (Physics ODE)" {
			engine = ISPPEngineLK
		}
		ca.deviceState.SetISPPEngine(engine)
		if ca.operationsStatusLabel != nil {
			ca.operationsStatusLabel.SetText(fmt.Sprintf("ISPP Engine: %s", engine.String()))
		}
	})
	selectedEngine := "Preisach (Level-based)"
	if ca.deviceState != nil {
		selectedEngine = ca.deviceState.GetISPPEngine().String()
	}
	engineSelect.SetSelected(selectedEngine)
	ca.isppEngineSelect = engineSelect

	engineRow := container.NewHBox(widget.NewLabel("ISPP Engine:"), engineSelect)

	return container.NewVBox(
		writeRow,
		engineRow,
	)
}

// updateArchitectureSpecificUI shows/hides panels based on architecture
func (ca *CircuitsApp) updateArchitectureSpecificUI() {
	isPassive := ca.deviceState.IsPassiveMode()

	sharedwidgets.SafeDo(func() {
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

// applyWriteVoltages converts the target voltage through the DAC and applies
// the resulting voltages to the array (including V/2 scheme in passive mode).
func (ca *CircuitsApp) applyWriteVoltages(row, col int, targetVoltage float64) (float64, int) {
	if ca.deviceState == nil {
		return targetVoltage, -1
	}
	applied, dacCode := ca.deviceState.DACWriteVoltage(targetVoltage)

	if ca.deviceState.IsPassiveMode() {
		ca.deviceState.ApplyHalfSelectWrite(row, col, applied)
	} else {
		// Non-passive: pass-through voltages should be 0 on unselected BLs.
		ca.deviceState.SetAllDACVoltages(0)
		ca.deviceState.SetDACVoltage(col, applied)
	}
	ca.deviceState.SetDACRangeMode(DACRangeWrite)

	return applied, dacCode
}

// applyHalfSelectDisturbLegacy updates neighbor polarization based on V/2 exposure.
// This is a lightweight, deterministic drift model for live UI feedback.
func (ca *CircuitsApp) applyHalfSelectDisturbLegacy(row, col int, direction HysteresisDirection, appliedVoltage float64) {
	if ca.deviceState == nil || !ca.deviceState.IsPassiveMode() {
		return
	}

	writeRange := ca.deviceState.GetWriteRange()
	if writeRange.Min <= 0 {
		return
	}

	halfVoltage := math.Abs(appliedVoltage) * HalfSelectVoltageRatio
	ratio := halfVoltage / writeRange.Min
	if ratio <= halfSelectDisturbThresholdRatio {
		return
	}

	strength := (ratio - halfSelectDisturbThresholdRatio) / (1.0 - halfSelectDisturbThresholdRatio)
	if strength <= 0 {
		return
	}
	step := strength * halfSelectDisturbRate

	deltaSign := 1.0
	if direction == DirectionDescending {
		deltaSign = -1.0
	}

	ca.mu.Lock()
	defer ca.mu.Unlock()

	if len(ca.halfSelectResidue) != ca.arrayRows {
		ca.halfSelectResidue = make([][]float64, ca.arrayRows)
		for r := range ca.halfSelectResidue {
			ca.halfSelectResidue[r] = make([]float64, ca.arrayCols)
		}
	}

	// Same row, different columns
	for c := 0; c < ca.arrayCols; c++ {
		if c == col {
			continue
		}
		ca.applyDisturbToCellLocked(row, c, step*deltaSign)
	}
	// Same column, different rows
	for r := 0; r < ca.arrayRows; r++ {
		if r == row {
			continue
		}
		ca.applyDisturbToCellLocked(r, col, step*deltaSign)
	}
}

func (ca *CircuitsApp) applyDisturbToCellLocked(row, col int, delta float64) {
	if row < 0 || row >= len(ca.arrayWeights) {
		return
	}
	if col < 0 || col >= len(ca.arrayWeights[row]) {
		return
	}

	ca.halfSelectResidue[row][col] += delta

	for ca.halfSelectResidue[row][col] >= 1.0 {
		if ca.arrayWeights[row][col] < ca.quantLevels-1 {
			ca.arrayWeights[row][col]++
		} else {
			ca.halfSelectResidue[row][col] = 0
			break
		}
		ca.halfSelectResidue[row][col] -= 1.0
	}

	for ca.halfSelectResidue[row][col] <= -1.0 {
		if ca.arrayWeights[row][col] > 0 {
			ca.arrayWeights[row][col]--
		} else {
			ca.halfSelectResidue[row][col] = 0
			break
		}
		ca.halfSelectResidue[row][col] += 1.0
	}
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

	sharedwidgets.SafeDo(func() {
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

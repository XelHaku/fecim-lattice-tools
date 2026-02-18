// Package gui - Enhanced app with all new widgets integrated
package gui

import (
	"fmt"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"

	"fecim-lattice-tools/shared/canvas"
	"fecim-lattice-tools/shared/crossbar"
)

// runEnhancedMVM performs MVM with full non-ideality analysis and updates all widgets.
func (ca *CrossbarApp) runEnhancedMVM() {
	// M5 UX fix: Prevent starting new MVM while one is running
	ca.stateMu.RLock()
	running := ca.isMVMRunning
	ca.stateMu.RUnlock()
	if running {
		return
	}

	getDebug().Println("runEnhancedMVM: Starting")

	// Create random input
	input := make([]float64, ca.config.Cols)
	for i := range input {
		input[i] = rand.Float64()
	}

	// Protected write to lastInput
	ca.stateMu.Lock()
	ca.lastInput = input
	ca.stateMu.Unlock()

	ca.mvmVis.SetInput(input)

	// Run animated MVM in goroutine with panic recovery
	utils.SafeGo("runEnhancedMVMAnimated", func() {
		ca.runEnhancedMVMAnimated(input)
	})
}

// runEnhancedMVMInstant performs instant MVM (no animation) for initial data population
func (ca *CrossbarApp) runEnhancedMVMInstant() {
	getDebug().Println("runEnhancedMVMInstant: Starting instant MVM for initial load")

	// Create random input
	input := make([]float64, ca.config.Cols)
	for i := range input {
		input[i] = rand.Float64()
	}

	ca.stateMu.Lock()
	ca.lastInput = input
	ca.stateMu.Unlock()

	ca.mvmVis.SetInput(input)

	// Perform enhanced MVM with all non-idealities
	opts := crossbar.DefaultMVMOptions()

	// Get architecture from GUI state
	ca.stateMu.RLock()
	arch := ca.architecture
	ca.stateMu.RUnlock()
	if arch != "" {
		opts.Architecture = arch
	}
	opts.Temperature = ca.currentTemperatureK()

	mvmResult, err := ca.array.MVMWithNonIdealities(input, opts)
	if err != nil {
		getDebug().Printf("runEnhancedMVMInstant: Error: %v", err)
		ca.updateStatus(fmt.Sprintf("Error: %v", err))
		return
	}

	ca.stateMu.Lock()
	ca.lastOutput = mvmResult.ActualOutput
	ca.lastMVMResult = mvmResult
	ca.stateMu.Unlock()

	ca.mvmVis.SetOutput(mvmResult.ActualOutput)

	// Update all visualizations (synchronously)
	ca.updateEnhancedWidgets(mvmResult)

	// Refresh the selected cell tooltip with new values
	ca.refreshSelectedCellTooltip()

	ca.updateStatus("Ready | Initial MVM complete. All analysis tabs populated.")
	getDebug().Println("runEnhancedMVMInstant: Complete")
}

// runEnhancedMVMAnimated performs the enhanced MVM with all analysis.
func (ca *CrossbarApp) runEnhancedMVMAnimated(input []float64) {
	// M5 UX fix: Set running flag and disable controls
	ca.stateMu.Lock()
	ca.isMVMRunning = true
	ca.stateMu.Unlock()
	fyne.Do(func() {
		ca.setControlsEnabled(false)
	})
	// Ensure we re-enable controls when done
	defer func() {
		ca.stateMu.Lock()
		ca.isMVMRunning = false
		ca.stateMu.Unlock()
		fyne.Do(func() {
			ca.setControlsEnabled(true)
		})
	}()

	// Phase 1: Input voltages applied (300ms)
	fyne.Do(func() {
		ca.modeIndicator.SetMode(int(DemoModeCompute))
		ca.updateStatus("COMPUTE | Phase 1/3: Applying input voltages...")

		cols := make([]int, ca.config.Cols)
		for i := range cols {
			cols[i] = i
		}
		ca.conductanceHeatmap.SetInputHighlight(cols)
		ca.conductanceHeatmap.SetAnimPhase(1, 0)
	})
	time.Sleep(300 * time.Millisecond)

	// Phase 2: Current flowing through cells (500ms)
	fyne.Do(func() {
		ca.updateStatus("COMPUTE | Phase 2/3: Current flowing (I = G × V)...")
	})

	steps := 10
	for i := 0; i <= steps; i++ {
		progress := float64(i) / float64(steps)
		fyne.Do(func() {
			ca.conductanceHeatmap.SetAnimPhase(2, progress)
		})
		time.Sleep(50 * time.Millisecond)
	}

	// Perform enhanced MVM with all non-idealities
	opts := crossbar.DefaultMVMOptions()

	// Get architecture from GUI state
	ca.stateMu.RLock()
	arch := ca.architecture
	ca.stateMu.RUnlock()
	if arch != "" {
		opts.Architecture = arch
	}
	opts.Temperature = ca.currentTemperatureK()

	mvmResult, err := ca.array.MVMWithNonIdealities(input, opts)
	if err != nil {
		fyne.Do(func() {
			ca.updateStatus(fmt.Sprintf("COMPUTE | Error: %v", err))
			ca.modeIndicator.SetMode(int(DemoModeIdle))
			ca.conductanceHeatmap.ClearAnimation()
		})
		return
	}

	ca.stateMu.Lock()
	ca.lastOutput = mvmResult.ActualOutput
	ca.lastMVMResult = mvmResult
	ca.stateMu.Unlock()

	// Phase 3: Output currents collected (300ms)
	fyne.Do(func() {
		ca.updateStatus("COMPUTE | Phase 3/3: Collecting outputs (ADC)...")
		ca.mvmVis.SetOutput(mvmResult.ActualOutput)

		rows := make([]int, ca.config.Rows)
		for i := range rows {
			rows[i] = i
		}
		ca.conductanceHeatmap.SetOutputHighlight(rows)
		ca.conductanceHeatmap.SetAnimPhase(3, 1)
	})
	time.Sleep(300 * time.Millisecond)

	// Update all visualizations
	fyne.Do(func() {
		ca.conductanceHeatmap.ClearAnimation()
		ca.updateEnhancedWidgets(mvmResult)
		ca.updateStatus(fmt.Sprintf("COMPUTE | Complete: %d MACs, %.2f pJ, %.0f× better than GPU",
			mvmResult.MACOperations, mvmResult.TotalEnergy, mvmResult.EnergyEfficiency))
		ca.modeIndicator.SetMode(int(DemoModeIdle))
	})

	getDebug().Println("runEnhancedMVM: Complete")
}

// runEnhancedMVMWithCurrentInput re-runs MVM using the existing input vector.
// Use this when changing architecture to compare the SAME computation with different physics.
func (ca *CrossbarApp) runEnhancedMVMWithCurrentInput() {
	// Get existing input (don't generate new random)
	ca.stateMu.RLock()
	input := ca.lastInput
	arch := ca.architecture
	ca.stateMu.RUnlock()

	// If no input yet OR input is wrong size (after array resize), generate one
	needNewInput := input == nil || len(input) == 0 || len(input) != ca.config.Cols
	if needNewInput {
		getDebug().Printf("[ARCH SWITCH] Creating new input: current len=%d, need=%d", len(input), ca.config.Cols)
		input = make([]float64, ca.config.Cols)
		for i := range input {
			input[i] = 0.5 // Use uniform input for fair comparison
		}
		ca.stateMu.Lock()
		ca.lastInput = input
		ca.stateMu.Unlock()
		ca.mvmVis.SetInput(input)
	} else {
		getDebug().Printf("[ARCH SWITCH] Reusing existing input: len=%d", len(input))
	}

	getDebug().Printf("[ARCH SWITCH] Architecture changed to: %s", arch)

	// Perform enhanced MVM with all non-idealities
	opts := crossbar.DefaultMVMOptions()
	opts.Architecture = arch // Always set architecture explicitly
	opts.Temperature = ca.currentTemperatureK()
	getDebug().Printf("[ARCH SWITCH] opts.Architecture=%s, Is1T1R=%v", opts.Architecture, opts.Is1T1R())

	mvmResult, err := ca.array.MVMWithNonIdealities(input, opts)
	if err != nil {
		getDebug().Printf("runEnhancedMVMWithCurrentInput: Error: %v", err)
		ca.updateStatus(fmt.Sprintf("Error: %v", err))
		return
	}

	ca.stateMu.Lock()
	ca.lastOutput = mvmResult.ActualOutput
	ca.lastMVMResult = mvmResult
	ca.stateMu.Unlock()

	ca.mvmVis.SetOutput(mvmResult.ActualOutput)

	// Update all visualizations
	ca.updateEnhancedWidgets(mvmResult)

	// Refresh the selected cell tooltip with new values
	ca.refreshSelectedCellTooltip()

	ca.updateStatus(fmt.Sprintf("Ready | Architecture: %s | Accuracy: %.1f%%", arch, 90.0-mvmResult.AccuracyLoss))
	getDebug().Println("runEnhancedMVMWithCurrentInput: Complete")
}

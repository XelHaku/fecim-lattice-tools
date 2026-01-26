// Package gui - Enhanced app with all new widgets integrated
package gui

import (
	"fmt"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"

	"multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar"
	"multilayer-ferroelectric-cim-visualizer/shared/utils"
)

// runEnhancedMVM performs MVM with full non-ideality analysis and updates all widgets.
func (ca *CrossbarApp) runEnhancedMVM() {
	debug.Println("runEnhancedMVM: Starting")

	ca.runMVMButton.Disable()

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
	debug.Println("runEnhancedMVMInstant: Starting instant MVM for initial load")

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
	mvmResult, err := ca.array.MVMWithNonIdealities(input, opts)
	if err != nil {
		debug.Printf("runEnhancedMVMInstant: Error: %v", err)
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

	ca.updateStatus("Ready | Initial MVM complete. All analysis tabs populated.")
	debug.Println("runEnhancedMVMInstant: Complete")
}

// runEnhancedMVMAnimated performs the enhanced MVM with all analysis.
func (ca *CrossbarApp) runEnhancedMVMAnimated(input []float64) {
	// Phase 1: Input voltages applied (300ms)
	fyne.Do(func() {
		ca.modeIndicator.SetMode(DemoModeCompute)
		ca.updateStatus("COMPUTE | Phase 1: Applying input voltages...")

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
		ca.updateStatus("COMPUTE | Phase 2: Current flowing through cells...")
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
	mvmResult, err := ca.array.MVMWithNonIdealities(input, opts)
	if err != nil {
		fyne.Do(func() {
			ca.updateStatus(fmt.Sprintf("COMPUTE | Error: %v", err))
			ca.modeIndicator.SetMode(DemoModeIdle)
			ca.conductanceHeatmap.ClearAnimation()
			ca.runMVMButton.Enable()
		})
		return
	}

	ca.lastOutput = mvmResult.ActualOutput
	ca.lastMVMResult = mvmResult

	// Phase 3: Output currents collected (300ms)
	fyne.Do(func() {
		ca.updateStatus("COMPUTE | Phase 3: Collecting output currents...")
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
		ca.modeIndicator.SetMode(DemoModeIdle)
		ca.runMVMButton.Enable()
	})

	debug.Println("runEnhancedMVM: Complete")
}

//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for crossbar visualization.
// animation.go contains MVM animation and auto-demo loop functions.
package gui

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// runMVM performs matrix-vector multiplication with animation.
func (ca *CrossbarApp) runMVM() {
	// M5 UX fix: Prevent starting new MVM while one is running
	ca.stateMu.RLock()
	running := ca.isMVMRunning
	ca.stateMu.RUnlock()
	if running {
		return
	}

	getDebug().Println("runMVM: Starting")

	// Create random input
	getDebug().Printf("runMVM: Creating input vector of size %d", ca.config.Cols)
	input := make([]float64, ca.config.Cols)
	for i := range input {
		input[i] = rand.Float64()
	}

	// Protected write to lastInput
	ca.stateMu.Lock()
	ca.lastInput = input
	ca.stateMu.Unlock()

	ca.mvmVis.SetInput(input)

	// Run animated MVM in goroutine
	go ca.runMVMAnimated(input)
}

// runMVMAnimated performs the MVM with visual animation.
func (ca *CrossbarApp) runMVMAnimated(input []float64) {
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
		ca.updateStatus("COMPUTE | Phase 1/3: Applying input voltages (DAC)")

		// Highlight all columns to show input voltages
		cols := make([]int, ca.config.Cols)
		for i := range cols {
			cols[i] = i
		}
		ca.conductanceHeatmap.SetInputHighlight(cols)
		ca.conductanceHeatmap.SetAnimPhase(1, 0)
	})
	time.Sleep(300 * time.Millisecond)

	// Phase 2: Current flowing through cells (500ms animation)
	fyne.Do(func() {
		ca.updateStatus("COMPUTE | Phase 2/3: Currents flowing (I = G × V)")
	})

	// Animate wave propagation
	steps := 10
	for i := 0; i <= steps; i++ {
		progress := float64(i) / float64(steps)
		fyne.Do(func() {
			ca.conductanceHeatmap.SetAnimPhase(2, progress)
		})
		time.Sleep(50 * time.Millisecond)
	}

	// Perform actual MVM computation
	output, err := ca.array.MVM(input)
	if err != nil {
		fyne.Do(func() {
			ca.updateStatus("COMPUTE | Error during MVM")
			ca.modeIndicator.SetMode(int(DemoModeIdle))
			ca.conductanceHeatmap.ClearAnimation()
			if ca.window != nil {
				dialog.ShowError(fmt.Errorf("MVM computation failed: %w", err), ca.window)
			}
		})
		return
	}

	// Protected write to lastOutput
	ca.stateMu.Lock()
	ca.lastOutput = output
	ca.stateMu.Unlock()

	// Phase 3: Output currents collected (300ms)
	fyne.Do(func() {
		ca.updateStatus("COMPUTE | Phase 3/3: Summing row currents (ADC)")
		ca.mvmVis.SetOutput(output)

		// Highlight all rows to show output currents
		rows := make([]int, ca.config.Rows)
		for i := range rows {
			rows[i] = i
		}
		ca.conductanceHeatmap.SetOutputHighlight(rows)
		ca.conductanceHeatmap.SetAnimPhase(3, 1)
	})
	time.Sleep(300 * time.Millisecond)

	// Finish and show results
	fyne.Do(func() {
		ca.conductanceHeatmap.ClearAnimation()

		// Calculate stats
		var sumInput, sumOutput float64
		for _, v := range input {
			sumInput += v
		}
		for _, v := range output {
			sumOutput += v
		}

		reads, writes := ca.array.GetStats()
		macOps := ca.config.Rows * ca.config.Cols

		ca.statsLabel.SetText(fmt.Sprintf(
			"MVM Complete!\n\n"+
				"Computation: I = W × V\n"+
				"Input Sum: %.4f\n"+
				"Output Sum: %.4f\n\n"+
				"Performance:\n"+
				"MAC Operations: %d\n"+
				"Parallelism: 100%%\n"+
				"Time: ~1ns (analog)\n\n"+
				"Statistics:\n"+
				"Total Reads: %d\n"+
				"Total Writes: %d",
			sumInput, sumOutput, macOps, reads, writes,
		))

		ca.updateStatus(fmt.Sprintf("COMPUTE | Complete: %d parallel MACs in ~1ns", macOps))
		ca.modeIndicator.SetMode(int(DemoModeIdle))

		// Auto-run IR Drop and Sneak Path analysis
		ca.runIRDropAnalysis()
		ca.runSneakPathAnalysis()

		ca.updateStatus("Ready | MVM complete. Check IR Drop and Sneak Paths tabs for analysis!")
	})
	getDebug().Println("runMVM: Complete")
}

// onDemoModeChanged handles demo mode selection changes.
func (ca *CrossbarApp) onDemoModeChanged(mode string) {
	// Stop any existing auto demo
	ca.stopAutoDemoLoop()

	switch mode {
	case "Auto Demo":
		ca.startAutoDemoLoop()
	case "Step-by-Step":
		ca.setEducationalContent("Step-by-Step Mode",
			"Click each button to see\nthe operation explained.\n\n"+
				"After 'Run MVM':\n"+
				"• IR Drop computed\n"+
				"• Sneak Paths computed\n"+
				"• Tabs auto-cycle")
	case "Manual":
		ca.setEducationalContent("What You're Seeing", "CROSSBAR MVM\n\nClick a button to start\na demonstration.")
	}
}

// startAutoDemoLoop starts the automatic demo loop.
func (ca *CrossbarApp) startAutoDemoLoop() {
	// Cancel any existing auto demo
	if ca.autoCancel != nil {
		ca.autoCancel()
	}

	ca.autoDemo = true
	ca.autoDemoStep = 0
	ca.autoCtx, ca.autoCancel = context.WithCancel(context.Background())
	ca.autoDemoTimer = time.NewTicker(3 * time.Second)

	ca.setEducationalContent("Auto Demo Mode",
		"Watch the demo cycle through\nall operations automatically.\n\n"+
			"Operations:\n"+
			"1. MVM Computation\n"+
			"2. IR Drop Analysis\n"+
			"3. Sneak Path Analysis\n"+
			"4. Reset & Repeat")

	go ca.autoDemoLoop(ca.autoCtx)
}

// stopAutoDemoLoop stops the automatic demo loop.
func (ca *CrossbarApp) stopAutoDemoLoop() {
	if ca.autoDemo {
		ca.autoDemo = false
		if ca.autoCancel != nil {
			ca.autoCancel()
			ca.autoCancel = nil
		}
		if ca.autoDemoTimer != nil {
			ca.autoDemoTimer.Stop()
			ca.autoDemoTimer = nil
		}
	}
}

// autoDemoLoop runs the automatic demonstration.
func (ca *CrossbarApp) autoDemoLoop(ctx context.Context) {
	// Run first operation immediately
	ca.runAutoDemoStep()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ca.autoDemoTimer.C:
			if !ca.autoDemo {
				return
			}
			ca.runAutoDemoStep()
		}
	}
}

// runAutoDemoStep executes one step of the auto demo.
func (ca *CrossbarApp) runAutoDemoStep() {
	switch ca.autoDemoStep {
	case 0:
		fyne.Do(func() {
			ca.runMVM()
		})
	case 1:
		fyne.Do(func() {
			ca.analyzeIRDrop()
		})
	case 2:
		fyne.Do(func() {
			ca.analyzeSneakPaths()
		})
	case 3:
		fyne.Do(func() {
			ca.resetArray()
		})
	}

	ca.autoDemoStep = (ca.autoDemoStep + 1) % 4
}

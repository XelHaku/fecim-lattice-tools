// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the COMPUTE mode panel and related functions.
package gui

import (
	"fmt"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ============================================================================
// COMPUTE MODE PANEL
// ============================================================================

// createComputeModePanel creates the compute mode configuration panel
func (ca *CircuitsApp) createComputeModePanel() {
	// Input vector entries
	ca.opsComputeInputs = make([]*widget.Entry, ca.arrayCols)
	ca.opsComputeVoltageLabels = make([]*widget.Label, ca.arrayCols)

	// Create horizontal input row with compact entries
	inputRow := container.NewHBox()
	maxDisplay := min(8, ca.arrayCols)
	for i := 0; i < maxDisplay; i++ {
		ca.opsComputeInputs[i] = widget.NewEntry()
		ca.opsComputeInputs[i].SetText(fmt.Sprintf("%d", ca.inputVector[i]))
		ca.opsComputeInputs[i].Resize(fyne.NewSize(45, 30)) // Compact width

		idx := i
		ca.opsComputeInputs[i].OnChanged = func(s string) {
			var v int
			fmt.Sscanf(s, "%d", &v)
			if v > 255 {
				v = 255
			}
			ca.mu.Lock()
			ca.inputVector[idx] = v
			ca.mu.Unlock()
			if ca.opsComputeVoltageLabels[idx] != nil {
				ca.opsComputeVoltageLabels[idx].SetText(fmt.Sprintf("%.2fV", float64(v)/255.0))
			}
			// Auto-compute on input change
			ca.computeAndUpdateAll()
		}

		// Compact column: label on top, entry below
		ca.opsComputeVoltageLabels[i] = widget.NewLabel(fmt.Sprintf("%.2fV", float64(ca.inputVector[i])/255.0))
		ca.opsComputeVoltageLabels[i].TextStyle = fyne.TextStyle{Monospace: true}

		col := container.NewVBox(
			widget.NewLabel(fmt.Sprintf("x%d", i)),
			ca.opsComputeInputs[i],
		)
		inputRow.Add(col)
	}

	// Output display
	ca.opsComputeOutputLabels = make([]*widget.Label, 8)
	outputGrid := container.NewGridWithColumns(2)
	for i := 0; i < 8; i++ {
		ca.opsComputeOutputLabels[i] = widget.NewLabel(fmt.Sprintf("y%d: --", i))
		outputGrid.Add(ca.opsComputeOutputLabels[i])
	}

	// Math breakdown
	ca.opsComputeMathLabel = widget.NewLabel(
		"I0 = G00*V0 + G01*V1 + ... (KCL sum)\n" +
			"All rows computed simultaneously!\n" +
			"Total latency: ~20ns",
	)
	ca.opsComputeMathLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Create Random Bits button
	randomBitsBtn := widget.NewButton("RANDOM BITS", func() {
		ca.mu.Lock()
		for i := range ca.inputVector {
			ca.inputVector[i] = rand.Intn(256)
		}
		ca.mu.Unlock()
		ca.updateOpsComputeInputs()
		ca.computeAndUpdateAll()
	})

	// Compact input section header (will be shown above array)
	inputHeader := container.NewHBox(
		widget.NewLabelWithStyle("INPUT VECTOR (0-255)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		randomBitsBtn,
	)

	physicsNote := widget.NewLabel("0.3-0.5V COMPUTE-safe (well below Vc)")
	physicsNote.TextStyle = fyne.TextStyle{Italic: true}

	// Populate the input row container that will appear above the array
	ca.computeInputRowContainer.Objects = []fyne.CanvasObject{
		widget.NewSeparator(),
		inputHeader,
		inputRow,
		physicsNote,
		widget.NewSeparator(),
	}

	// Output section
	// Ideal crossbar disclaimer
	idealDisclaimer := widget.NewLabel(
		"IDEAL CROSSBAR: No IR drop or sneak paths (see Module 2)")
	idealDisclaimer.TextStyle = fyne.TextStyle{Italic: true}

	outputSection := container.NewVBox(
		widget.NewLabelWithStyle("OUTPUT VECTOR", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("I_row -> TIA (10k) -> ADC (5-bit):"),
		outputGrid,
		idealDisclaimer,
	)

	// Math section
	mathSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("MATH (Row 0 Breakdown)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ca.opsComputeMathLabel,
	)

	// Performance info
	perfSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("PERFORMANCE", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("DAC: 5ns | Array settle: 5ns | ADC: 10ns"),
		widget.NewLabel("TOTAL: ~20ns for full MVM!"),
		widget.NewLabel("GPU equivalent: ~1000 cycles"),
	)

	ca.computeConfigPanel = container.NewVBox(
		outputSection,
		mathSection,
		perfSection,
	)
}

// updateOpsComputeInputs updates the compute input display
func (ca *CircuitsApp) updateOpsComputeInputs() {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	for i := 0; i < min(8, len(ca.opsComputeInputs)); i++ {
		if ca.opsComputeInputs[i] != nil {
			fyne.Do(func() {
				ca.opsComputeInputs[i].SetText(fmt.Sprintf("%d", ca.inputVector[i]))
			})
		}
		if ca.opsComputeVoltageLabels[i] != nil {
			voltage := float64(ca.inputVector[i]) / 255.0
			fyne.Do(func() {
				ca.opsComputeVoltageLabels[i].SetText(fmt.Sprintf("%.2fV", voltage))
			})
		}
	}
}

// computeAndUpdateAll performs MVM and updates all output displays
// Called by: input changes, RANDOM BITS, mode selector, COMPUTE button
// IMPORTANT: Does NOT call updateOpsComputeInputs() to prevent Entry->OnChanged recursion
func (ca *CircuitsApp) computeAndUpdateAll() {
	// 1. MVM computation
	ca.mu.Lock()
	rows := min(8, ca.arrayRows)
	cols := min(8, ca.arrayCols)

	for r := 0; r < rows && r < len(ca.arrayWeights); r++ {
		sum := 0.0
		for c := 0; c < cols && c < len(ca.arrayWeights[r]); c++ {
			conductance := 1.0 + float64(ca.arrayWeights[r][c])/29.0*99.0
			voltage := float64(ca.inputVector[c]) / 255.0
			sum += conductance * voltage
		}
		ca.outputVector[r] = sum
	}
	ca.mu.Unlock()

	// 2. Update output labels with TIA/ADC conversion
	ca.mu.RLock()
	for i := 0; i < 8 && i < len(ca.outputVector); i++ {
		if ca.opsComputeOutputLabels[i] != nil {
			rawCurrent := ca.outputVector[i]
			tiaVoltage := ca.tia.Convert(rawCurrent * 1e-6)
			adcLevel := ca.adc.Convert(tiaVoltage)
			isSaturated := rawCurrent > 100.0

			idx := i
			current := rawCurrent
			tiaV := tiaVoltage
			level := adcLevel
			sat := isSaturated
			fyne.Do(func() {
				// Show full pipeline: Current -> TIA Voltage -> ADC Level
				satSuffix := ""
				if sat {
					satSuffix = " SAT"
				}
				ca.opsComputeOutputLabels[idx].SetText(
					fmt.Sprintf("y%d: %.1fuA -> %.2fV -> L%d%s", idx, current, tiaV, level, satSuffix))
			})
		}
	}
	ca.mu.RUnlock()

	// 3. Update math breakdown
	ca.updateOpsComputeMath()

	// 4. Update data path displays
	ca.updateOpsComputeInputDataPath()
	ca.updateOpsComputeOutputDataPath()
}

// updateOpsComputeInputDataPath updates the input data path display
func (ca *CircuitsApp) updateOpsComputeInputDataPath() {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	if len(ca.inputVector) == 0 {
		return
	}

	// Show summary of input vector (first value as example)
	digitalVal := ca.inputVector[0]
	voltage := float64(digitalVal) / 255.0

	if ca.opsComputeInputDigitalLabel != nil {
		fyne.Do(func() {
			ca.opsComputeInputDigitalLabel.SetText(fmt.Sprintf("x0: %d\n0b%08b", digitalVal, digitalVal))
		})
	}
	if ca.opsComputeInputDACLabel != nil {
		fyne.Do(func() {
			ca.opsComputeInputDACLabel.SetText(fmt.Sprintf("%.2fV", voltage))
		})
	}
}

// updateOpsComputeOutputDataPath updates the output data path display
func (ca *CircuitsApp) updateOpsComputeOutputDataPath() {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	if len(ca.outputVector) == 0 {
		return
	}

	// Show y0 as the example
	rawCurrent := ca.outputVector[0] // uA

	// TIA conversion (saturates at 100 uA -> 1.0V output)
	tiaVoltage := ca.tia.Convert(rawCurrent * 1e-6) // uA to A

	// ADC conversion (5-bit: 0V->0, 1V->31)
	adcLevel := ca.adc.Convert(tiaVoltage)

	// Check for TIA saturation
	isSaturated := rawCurrent > 100.0

	satSuffix := ""
	if isSaturated {
		satSuffix = " (SAT)"
	}

	if ca.opsComputeOutputCurrentLabel != nil {
		fyne.Do(func() {
			ca.opsComputeOutputCurrentLabel.SetText(fmt.Sprintf("%.1f uA%s", rawCurrent, satSuffix))
		})
	}
	if ca.opsComputeOutputTIALabel != nil {
		fyne.Do(func() {
			ca.opsComputeOutputTIALabel.SetText(fmt.Sprintf("%.3f V%s", tiaVoltage, satSuffix))
		})
	}
	if ca.opsComputeOutputADCLabel != nil {
		fyne.Do(func() {
			ca.opsComputeOutputADCLabel.SetText(fmt.Sprintf("Level %d%s", adcLevel, satSuffix))
		})
	}
}

// updateOpsComputeMath updates the math breakdown display
func (ca *CircuitsApp) updateOpsComputeMath() {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	if len(ca.arrayWeights) == 0 || len(ca.arrayWeights[0]) == 0 {
		return
	}

	cols := min(4, len(ca.arrayWeights[0]))
	mathText := "I0 = "
	var terms []string
	totalCurrent := 0.0

	for c := 0; c < cols; c++ {
		conductance := 1.0 + float64(ca.arrayWeights[0][c])/29.0*99.0
		voltage := float64(ca.inputVector[c]) / 255.0
		current := conductance * voltage
		totalCurrent += current
		terms = append(terms, fmt.Sprintf("%.0f*%.2f", conductance, voltage))
	}

	mathText += terms[0]
	for i := 1; i < len(terms); i++ {
		mathText += " + " + terms[i]
	}
	mathText += " + ...\n"
	mathText += fmt.Sprintf("   = %.1f uA\n", ca.outputVector[0])
	mathText += "ALL ROWS IN PARALLEL!"

	if ca.opsComputeMathLabel != nil {
		fyne.Do(func() {
			ca.opsComputeMathLabel.SetText(mathText)
		})
	}
}

// ============================================================================
// COMPUTE MODE ACTION HANDLERS
// ============================================================================

// onOpsCompute performs matrix-vector multiplication
func (ca *CircuitsApp) onOpsCompute() {
	ca.computeAndUpdateAll()
	ca.operationsStatusLabel.SetText("Compute complete in ~20ns")
}

// onOpsAnimate animates the compute process step by step with visual feedback
func (ca *CircuitsApp) onOpsAnimate() {
	ca.mu.Lock()
	ca.animationActive = true
	ca.mu.Unlock()

	ca.operationsStatusLabel.SetText("Animating...")

	go func() {
		// Step 1: DAC highlight
		ca.mu.Lock()
		ca.animationStep = 1
		ca.mu.Unlock()
		ca.refreshSharedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Step 1: DAC conversion (5ns)")
		})
		time.Sleep(600 * time.Millisecond)

		// Step 2: Array highlight
		ca.mu.Lock()
		ca.animationStep = 2
		ca.mu.Unlock()
		ca.refreshSharedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Step 2: Array MVM (5ns)")
		})
		time.Sleep(600 * time.Millisecond)

		// Step 3: ADC highlight
		ca.mu.Lock()
		ca.animationStep = 3
		ca.mu.Unlock()
		ca.refreshSharedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Step 3: ADC conversion (10ns)")
		})
		time.Sleep(600 * time.Millisecond)

		// Complete
		ca.mu.Lock()
		ca.animationStep = 0
		ca.animationActive = false
		ca.mu.Unlock()
		ca.computeAndUpdateAll()
		ca.refreshSharedArray()
		fyne.Do(func() {
			ca.operationsStatusLabel.SetText("Compute complete in ~20ns")
		})
	}()
}

// onOpsReset resets the compute state
func (ca *CircuitsApp) onOpsReset() {
	ca.mu.Lock()
	for i := range ca.inputVector {
		ca.inputVector[i] = 0
	}
	for i := range ca.outputVector {
		ca.outputVector[i] = 0
	}
	ca.mu.Unlock()

	ca.updateOpsComputeInputs()
	for i := 0; i < 8 && i < len(ca.opsComputeOutputLabels); i++ {
		if ca.opsComputeOutputLabels[i] != nil {
			fyne.Do(func() {
				ca.opsComputeOutputLabels[i].SetText(fmt.Sprintf("y%d: --", i))
			})
		}
	}

	if ca.opsComputeMathLabel != nil {
		fyne.Do(func() {
			ca.opsComputeMathLabel.SetText(
				"I0 = G00*V0 + G01*V1 + ... (KCL sum)\n" +
					"All rows computed simultaneously!\n" +
					"Total latency: ~20ns",
			)
		})
	}

	ca.operationsStatusLabel.SetText("Reset complete")
}

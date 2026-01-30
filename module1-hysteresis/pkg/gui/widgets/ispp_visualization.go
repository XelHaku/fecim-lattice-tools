// Package widgets provides custom Fyne widgets for the hysteresis visualization.
// This file implements H14: ISPP (Incremental Step Pulse Programming) visualization.
package widgets

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/physics"
)

// ISPPVisualization displays ISPP write-verify statistics with convergence info.
// Shows:
// - Pulses histogram (how many pulses needed to reach target)
// - Success rate over time
// - Failure rate vs cycle count
// - Current ISPP state machine visualization
type ISPPVisualization struct {
	widget.BaseWidget

	stats *physics.WriteVerifyStats

	// Current ISPP animation state
	currentPulse   int     // Current pulse in sequence (0 = idle)
	targetLevel    int     // Target level being written
	currentLevel   int     // Current sensed level
	pulseVoltage   float64 // Current pulse voltage
	isConverged    bool    // Whether current write has converged
	isAnimating    bool    // Whether ISPP animation is running

	// Cached UI elements
	content fyne.CanvasObject
}

// NewISPPVisualization creates a new ISPP visualization widget.
func NewISPPVisualization() *ISPPVisualization {
	v := &ISPPVisualization{
		stats: physics.NewWriteVerifyStats(),
	}
	v.ExtendBaseWidget(v)
	return v
}

// GetStats returns the underlying stats tracker for recording operations.
func (v *ISPPVisualization) GetStats() *physics.WriteVerifyStats {
	return v.stats
}

// SetAnimationState updates the current ISPP animation state.
func (v *ISPPVisualization) SetAnimationState(pulse int, target int, current int, voltage float64, converged bool) {
	v.currentPulse = pulse
	v.targetLevel = target
	v.currentLevel = current
	v.pulseVoltage = voltage
	v.isConverged = converged
	v.isAnimating = pulse > 0
	v.Refresh()
}

// RecordWrite records a completed write and updates the display.
func (v *ISPPVisualization) RecordWrite(targetLevel int, pulsesUsed int, success bool, hadOvershoot bool) {
	v.stats.RecordWrite(targetLevel, pulsesUsed, success, hadOvershoot)
	v.Refresh()
}

// Reset clears all statistics.
func (v *ISPPVisualization) Reset() {
	v.stats.Reset()
	v.currentPulse = 0
	v.isAnimating = false
	v.isConverged = false
	v.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (v *ISPPVisualization) CreateRenderer() fyne.WidgetRenderer {
	v.content = v.createContent()
	return widget.NewSimpleRenderer(v.content)
}

func (v *ISPPVisualization) createContent() fyne.CanvasObject {
	// Header
	headerBg := canvas.NewRectangle(color.RGBA{30, 60, 90, 255})
	headerText := canvas.NewText("ISPP WRITE-VERIFY STATISTICS", color.RGBA{0, 212, 255, 255})
	headerText.TextSize = 14
	headerText.TextStyle = fyne.TextStyle{Bold: true}
	headerText.Alignment = fyne.TextAlignCenter

	header := container.NewStack(
		headerBg,
		container.NewCenter(headerText),
	)

	// Current state indicator
	stateWidget := v.createStateIndicator()

	// Statistics summary
	summaryWidget := v.createSummaryWidget()

	// Pulses histogram
	histogramWidget := v.createHistogramWidget()

	// Convergence explanation
	explanationWidget := v.createExplanationWidget()

	return container.NewVBox(
		header,
		widget.NewSeparator(),
		stateWidget,
		widget.NewSeparator(),
		summaryWidget,
		widget.NewSeparator(),
		histogramWidget,
		widget.NewSeparator(),
		explanationWidget,
	)
}

func (v *ISPPVisualization) createStateIndicator() fyne.CanvasObject {
	// Show current ISPP state machine position
	var stateText string
	var stateColor color.Color

	if !v.isAnimating {
		stateText = "IDLE - Ready for write"
		stateColor = color.RGBA{100, 100, 100, 255}
	} else if v.isConverged {
		stateText = fmt.Sprintf("CONVERGED at L%d in %d pulses", v.currentLevel, v.currentPulse)
		stateColor = color.RGBA{0, 200, 0, 255}
	} else {
		delta := v.targetLevel - v.currentLevel
		direction := "↑"
		if delta < 0 {
			direction = "↓"
			delta = -delta
		}
		stateText = fmt.Sprintf("PULSE %d: L%d → L%d (%s%d) @ %.2fV",
			v.currentPulse, v.currentLevel, v.targetLevel, direction, delta, v.pulseVoltage)
		stateColor = color.RGBA{255, 200, 0, 255}
	}

	stateLabel := canvas.NewText(stateText, stateColor)
	stateLabel.TextSize = 12
	stateLabel.TextStyle = fyne.TextStyle{Bold: true}
	stateLabel.Alignment = fyne.TextAlignCenter

	// State machine phases
	phases := []string{"APPLY", "WAIT", "VERIFY", "ADJUST"}
	phaseWidgets := make([]fyne.CanvasObject, len(phases))

	currentPhaseIdx := v.currentPulse % 4 // Simple round-robin for visual effect
	if !v.isAnimating {
		currentPhaseIdx = -1
	}

	for i, phase := range phases {
		bgColor := color.RGBA{40, 40, 50, 255}
		textColor := color.RGBA{150, 150, 150, 255}
		if v.isAnimating && i == currentPhaseIdx {
			bgColor = color.RGBA{0, 100, 180, 255}
			textColor = color.RGBA{255, 255, 255, 255}
		}

		bg := canvas.NewRectangle(bgColor)
		bg.SetMinSize(fyne.NewSize(50, 20))

		text := canvas.NewText(phase, textColor)
		text.TextSize = 10
		text.Alignment = fyne.TextAlignCenter

		phaseWidgets[i] = container.NewStack(bg, container.NewCenter(text))
	}

	phasesRow := container.NewHBox(phaseWidgets...)

	return container.NewVBox(
		container.NewCenter(stateLabel),
		container.NewCenter(phasesRow),
	)
}

func (v *ISPPVisualization) createSummaryWidget() fyne.CanvasObject {
	totalWrites := v.stats.TotalWrites
	successRate := v.stats.GetSuccessRate() * 100
	avgPulses := v.stats.GetAveragePulses()
	overshootRate := v.stats.GetOvershootRate() * 100

	// Color-code success rate
	successColor := color.RGBA{0, 200, 0, 255}
	if successRate < 99 {
		successColor = color.RGBA{255, 200, 0, 255}
	}
	if successRate < 95 {
		successColor = color.RGBA{255, 80, 80, 255}
	}

	row1 := container.NewHBox(
		widget.NewLabel("Total writes:"),
		widget.NewLabel(fmt.Sprintf("%d", totalWrites)),
		widget.NewLabel(" | "),
		widget.NewLabel("Success:"),
		canvas.NewText(fmt.Sprintf("%.1f%%", successRate), successColor),
	)

	row2 := container.NewHBox(
		widget.NewLabel("Avg pulses:"),
		widget.NewLabel(fmt.Sprintf("%.1f", avgPulses)),
		widget.NewLabel(" | "),
		widget.NewLabel("Overshoot:"),
		widget.NewLabel(fmt.Sprintf("%.1f%%", overshootRate)),
	)

	return container.NewVBox(row1, row2)
}

func (v *ISPPVisualization) createHistogramWidget() fyne.CanvasObject {
	histogram := v.stats.GetPulsesHistogram()

	// Find max for scaling
	maxCount := 1
	totalSuccessful := 0
	for _, count := range histogram {
		if count > maxCount {
			maxCount = count
		}
		totalSuccessful += count
	}

	// Create bar chart
	bars := make([]fyne.CanvasObject, 10)
	for i := 0; i < 10; i++ {
		count := histogram[i]
		height := float32(0)
		if maxCount > 0 {
			height = float32(count) / float32(maxCount) * 40 // Max 40px height
		}

		barColor := color.RGBA{0, 150, 255, 255}
		if i >= 5 {
			barColor = color.RGBA{255, 150, 0, 255} // Warn for 6+ pulses
		}

		bar := canvas.NewRectangle(barColor)
		bar.SetMinSize(fyne.NewSize(20, height))

		label := canvas.NewText(fmt.Sprintf("%d", i+1), color.RGBA{200, 200, 200, 255})
		label.TextSize = 9
		label.Alignment = fyne.TextAlignCenter

		countLabel := canvas.NewText(fmt.Sprintf("%d", count), color.RGBA{150, 150, 150, 255})
		countLabel.TextSize = 8
		countLabel.Alignment = fyne.TextAlignCenter

		bars[i] = container.NewVBox(
			countLabel,
			container.NewVBox(bar),
			label,
		)
	}

	histTitle := widget.NewLabel("Pulses to Converge:")
	histTitle.TextStyle = fyne.TextStyle{Bold: true}

	barsRow := container.NewHBox(bars...)

	return container.NewVBox(
		histTitle,
		container.NewCenter(barsRow),
	)
}

func (v *ISPPVisualization) createExplanationWidget() fyne.CanvasObject {
	// Educational content about ISPP
	explanation := widget.NewLabel(
		"ISPP: Incremental Step Pulse Programming\n" +
			"• Apply pulse → Verify → Adjust voltage → Repeat\n" +
			"• Typical: 1-3 pulses for well-calibrated devices\n" +
			"• >5 pulses indicates poor calibration or device variability",
	)
	explanation.TextStyle = fyne.TextStyle{Italic: true}
	explanation.Wrapping = fyne.TextWrapWord

	// Show projected failure rate vs cycles (if we have data)
	failureHistory := v.stats.GetFailureRateVsCycles()

	var projectionText string
	if len(failureHistory) >= 2 {
		// Simple linear projection
		lastRate := failureHistory[len(failureHistory)-1]
		prevRate := failureHistory[len(failureHistory)-2]
		trend := lastRate - prevRate

		if trend > 0 {
			projectionText = fmt.Sprintf("Failure trend: +%.2f%% per 100 cycles (fatigue detected)", trend*100)
		} else {
			projectionText = "Failure trend: Stable (no significant fatigue)"
		}
	} else {
		// Use simulated projection
		cycles := v.stats.CycleCount
		projectedRate := physics.SimulateFailureRateProgression(cycles, 1e9) * 100
		projectionText = fmt.Sprintf("Projected failure @ cycle %d: %.3f%% (model)", cycles, projectedRate)
	}

	projection := widget.NewLabel(projectionText)
	projection.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		explanation,
		projection,
	)
}

// Refresh updates the widget display.
func (v *ISPPVisualization) Refresh() {
	v.content = v.createContent()
	v.BaseWidget.Refresh()
}

// MinSize returns the minimum size for the widget.
func (v *ISPPVisualization) MinSize() fyne.Size {
	return fyne.NewSize(300, 300)
}

// SimulateISPPSequence runs a simulated ISPP sequence for demonstration.
// Returns the number of pulses used and whether it succeeded.
func SimulateISPPSequence(currentLevel, targetLevel int, deviceVariability float64) (pulses int, success bool, overshoots int) {
	if currentLevel == targetLevel {
		return 1, true, 0
	}

	// Simulate ISPP convergence
	// More variability = more pulses needed
	expectedPulses := 1 + int(math.Abs(float64(targetLevel-currentLevel))*deviceVariability*0.5)
	if expectedPulses > 10 {
		expectedPulses = 10
	}

	// Occasional overshoot
	if deviceVariability > 0.3 && math.Abs(float64(targetLevel-currentLevel)) > 5 {
		overshoots = 1
		expectedPulses += 2
	}

	// Success probability decreases with pulses needed
	successProb := 1.0 - float64(expectedPulses)*0.02
	success = successProb > 0.5 // Simplified

	return expectedPulses, success, overshoots
}

// Package visualization provides terminal-based visualization for crossbar arrays.
package visualization

import (
	"fmt"
	"math"
	"strings"

	"fecim-lattice-tools/shared/crossbar"
)

// TerminalVisualizer provides ASCII/Unicode visualization of crossbar operations.
type TerminalVisualizer struct {
	array    *crossbar.Array
	useColor bool
}

// NewTerminalVisualizer creates a new terminal visualizer.
func NewTerminalVisualizer(array *crossbar.Array, useColor bool) *TerminalVisualizer {
	return &TerminalVisualizer{
		array:    array,
		useColor: useColor,
	}
}

// ShowCrossbarState displays the current state of the crossbar array.
func (v *TerminalVisualizer) ShowCrossbarState() {
	fmt.Println("\n=== Crossbar Array State ===")
	fmt.Printf("Size: %d x %d\n", v.array.Rows(), v.array.Cols())

	matrix := v.array.GetConductanceMatrix()
	maxCols := min(v.array.Cols(), 32) // Limit display width
	maxRows := min(v.array.Rows(), 16) // Limit display height

	// Header
	fmt.Print("     ")
	for j := 0; j < maxCols; j++ {
		fmt.Printf("%2d ", j%10)
	}
	if v.array.Cols() > maxCols {
		fmt.Print("...")
	}
	fmt.Println()

	// Separator
	fmt.Print("   +" + strings.Repeat("---", maxCols))
	if v.array.Cols() > maxCols {
		fmt.Print("---")
	}
	fmt.Println("+")

	// Data rows
	for i := 0; i < maxRows; i++ {
		fmt.Printf("%2d |", i)
		for j := 0; j < maxCols; j++ {
			level := int(matrix[i][j] * 29)
			char := v.levelToChar(level)
			if v.useColor {
				color := v.levelToColor(level)
				fmt.Printf("%s%s%s ", color, char, "\033[0m")
			} else {
				fmt.Printf(" %s ", char)
			}
		}
		if v.array.Cols() > maxCols {
			fmt.Print("...")
		}
		fmt.Println("|")
	}

	if v.array.Rows() > maxRows {
		fmt.Println("   ... (more rows)")
	}

	// Footer
	fmt.Print("   +" + strings.Repeat("---", maxCols))
	if v.array.Cols() > maxCols {
		fmt.Print("---")
	}
	fmt.Println("+")

	// Legend
	fmt.Println("\nLevel Legend (30-level baseline):")
	fmt.Print("  Low ")
	for l := 0; l < 30; l += 3 {
		char := v.levelToChar(l)
		if v.useColor {
			color := v.levelToColor(l)
			fmt.Printf("%s%s%s", color, char, "\033[0m")
		} else {
			fmt.Print(char)
		}
	}
	fmt.Println(" High")
}

// ShowMVMOperation displays a matrix-vector multiplication visualization.
func (v *TerminalVisualizer) ShowMVMOperation(input []float64, output []float64) {
	fmt.Println("\n=== Matrix-Vector Multiplication ===")

	maxInput := min(len(input), 16)
	maxOutput := min(len(output), 16)

	// Input vector (voltages)
	fmt.Println("\nInput Voltages (V):")
	fmt.Print("  [")
	for i := 0; i < maxInput; i++ {
		bar := v.valueToBar(input[i])
		if v.useColor {
			fmt.Printf("\033[94m%s\033[0m", bar)
		} else {
			fmt.Print(bar)
		}
	}
	if len(input) > maxInput {
		fmt.Print("...")
	}
	fmt.Println("]")

	// Show MVM operation symbol
	fmt.Println("        |")
	fmt.Println("        v")
	fmt.Println("  [ Crossbar W ] × [ V ]")
	fmt.Println("        |")
	fmt.Println("        v")

	// Output vector (currents)
	fmt.Println("\nOutput Currents (I):")
	fmt.Print("  [")
	for i := 0; i < maxOutput; i++ {
		bar := v.valueToBar(output[i])
		if v.useColor {
			fmt.Printf("\033[93m%s\033[0m", bar)
		} else {
			fmt.Print(bar)
		}
	}
	if len(output) > maxOutput {
		fmt.Print("...")
	}
	fmt.Println("]")

	// Show numeric values
	fmt.Printf("\nInput (%d values):  ", len(input))
	for i := 0; i < min(8, len(input)); i++ {
		fmt.Printf("%.2f ", input[i])
	}
	if len(input) > 8 {
		fmt.Print("...")
	}
	fmt.Println()

	fmt.Printf("Output (%d values): ", len(output))
	for i := 0; i < min(8, len(output)); i++ {
		fmt.Printf("%.2f ", output[i])
	}
	if len(output) > 8 {
		fmt.Print("...")
	}
	fmt.Println()
}

// ShowNeuralNetworkInference displays neural network inference through crossbars.
func (v *TerminalVisualizer) ShowNeuralNetworkInference(layers int, input []float64, activations [][]float64, prediction int, confidence float64) {
	fmt.Println("\n=== Neural Network Inference ===")

	// Show input (e.g., MNIST digit)
	if len(input) == 784 {
		fmt.Println("\nInput Image (28x28):")
		for row := 0; row < 28; row++ {
			fmt.Print("  ")
			for col := 0; col < 28; col++ {
				val := input[row*28+col]
				if val > 0.75 {
					fmt.Print("##")
				} else if val > 0.5 {
					fmt.Print("++")
				} else if val > 0.25 {
					fmt.Print("..")
				} else {
					fmt.Print("  ")
				}
			}
			fmt.Println()
		}
	}

	// Show layer activations
	for i, act := range activations {
		fmt.Printf("\nLayer %d Activations (%d neurons):\n", i+1, len(act))
		maxShow := min(len(act), 32)
		fmt.Print("  ")
		for j := 0; j < maxShow; j++ {
			bar := v.valueToBar(act[j])
			if v.useColor {
				fmt.Printf("\033[92m%s\033[0m", bar)
			} else {
				fmt.Print(bar)
			}
		}
		if len(act) > maxShow {
			fmt.Print("...")
		}
		fmt.Println()
	}

	// Show prediction
	fmt.Println("\n=== Prediction ===")
	if len(activations) > 0 {
		lastLayer := activations[len(activations)-1]
		fmt.Println("Output probabilities:")
		for i, prob := range lastLayer {
			barLen := int(prob * 30)
			bar := strings.Repeat("█", barLen) + strings.Repeat("░", 30-barLen)
			marker := " "
			if i == prediction {
				marker = "→"
			}
			if v.useColor {
				if i == prediction {
					fmt.Printf("  %s %d: \033[92m%s\033[0m %.1f%%\n", marker, i, bar, prob*100)
				} else {
					fmt.Printf("  %s %d: %s %.1f%%\n", marker, i, bar, prob*100)
				}
			} else {
				fmt.Printf("  %s %d: %s %.1f%%\n", marker, i, bar, prob*100)
			}
		}
	}

	fmt.Printf("\nPredicted digit: %d (confidence: %.1f%%)\n", prediction, confidence*100)
}

func (v *TerminalVisualizer) levelToChar(level int) string {
	// Use block characters for different levels
	chars := []string{" ", "░", "▒", "▓", "█"}
	idx := level * len(chars) / 30
	if idx >= len(chars) {
		idx = len(chars) - 1
	}
	return chars[idx]
}

func (v *TerminalVisualizer) levelToColor(level int) string {
	// ANSI color codes: blue (low) -> white (mid) -> red (high)
	if level < 10 {
		return "\033[94m" // Blue
	} else if level < 20 {
		return "\033[97m" // White
	}
	return "\033[91m" // Red
}

func (v *TerminalVisualizer) valueToBar(value float64) string {
	// Clamp to [0, 1]
	value = math.Max(0, math.Min(1, value))
	bars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	idx := int(value * float64(len(bars)-1))
	return bars[idx]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ShowIRDropAnalysis displays IR drop heatmap and statistics.
func (v *TerminalVisualizer) ShowIRDropAnalysis(analysis *crossbar.IRDropAnalysis) {
	fmt.Println("\n=== IR Drop Analysis ===")
	fmt.Println("Wire resistance effects on voltage distribution")

	// Get normalized IR drop map
	irMap := analysis.GetIRDropMap()

	rows := len(irMap)
	cols := len(irMap[0])
	maxRows := min(rows, 20)
	maxCols := min(cols, 40)

	// Header
	fmt.Println("\nIR Drop Heatmap (darker = more drop):")
	fmt.Print("     ")
	for j := 0; j < maxCols; j += 2 {
		fmt.Printf("%2d", j%100)
	}
	fmt.Println()

	// Heatmap
	for i := 0; i < maxRows; i++ {
		fmt.Printf("%3d |", i)
		for j := 0; j < maxCols; j++ {
			char := v.irDropToChar(irMap[i][j])
			if v.useColor {
				color := v.irDropToColor(irMap[i][j])
				fmt.Printf("%s%s\033[0m", color, char)
			} else {
				fmt.Print(char)
			}
		}
		fmt.Println("|")
	}

	if rows > maxRows {
		fmt.Printf("    ... (%d more rows)\n", rows-maxRows)
	}

	// Statistics
	fmt.Println("\nIR Drop Statistics:")
	fmt.Printf("  Max IR Drop:      %.2f%% (at cell [%d,%d])\n",
		analysis.MaxIRDrop*100, analysis.WorstCaseCell[0], analysis.WorstCaseCell[1])
	fmt.Printf("  Avg IR Drop:      %.2f%%\n", analysis.AvgIRDrop*100)
	fmt.Printf("  Variance:         %.4f\n", analysis.IRDropVariance)

	// Legend
	fmt.Println("\n  Legend: ░ Low drop  ▒ Medium  ▓ High  █ Severe")

	// Impact assessment
	fmt.Println("\n  Impact Assessment:")
	if analysis.MaxIRDrop < 0.05 {
		fmt.Println("  ✓ IR drop is minimal - excellent voltage uniformity")
	} else if analysis.MaxIRDrop < 0.10 {
		fmt.Println("  ⚠ Moderate IR drop - may affect accuracy at edges")
	} else {
		fmt.Println("  ✗ Severe IR drop - consider wider metal lines")
	}
}

// ShowSneakPathAnalysis displays sneak path current analysis.
func (v *TerminalVisualizer) ShowSneakPathAnalysis(analysis *crossbar.SneakPathAnalysis, selectedRow, selectedCol int) {
	fmt.Println("\n=== Sneak Path Analysis ===")
	fmt.Printf("Selected cell: [%d, %d]\n", selectedRow, selectedCol)

	// Get normalized sneak map
	sneakMap := analysis.GetSneakMap()

	rows := len(sneakMap)
	cols := len(sneakMap[0])
	maxRows := min(rows, 20)
	maxCols := min(cols, 40)

	// Header
	fmt.Println("\nSneak Current Map (brighter = more sneak):")
	fmt.Print("     ")
	for j := 0; j < maxCols; j += 2 {
		fmt.Printf("%2d", j%100)
	}
	fmt.Println()

	// Heatmap
	for i := 0; i < maxRows; i++ {
		fmt.Printf("%3d |", i)
		for j := 0; j < maxCols; j++ {
			if i == selectedRow && j == selectedCol {
				// Mark selected cell
				if v.useColor {
					fmt.Print("\033[92m★\033[0m")
				} else {
					fmt.Print("★")
				}
			} else {
				char := v.sneakToChar(sneakMap[i][j])
				if v.useColor {
					color := v.sneakToColor(sneakMap[i][j])
					fmt.Printf("%s%s\033[0m", color, char)
				} else {
					fmt.Print(char)
				}
			}
		}
		fmt.Println("|")
	}

	if rows > maxRows {
		fmt.Printf("    ... (%d more rows)\n", rows-maxRows)
	}

	// Statistics
	fmt.Println("\nSneak Path Statistics:")
	fmt.Printf("  Signal Current:    %.4f (normalized)\n", analysis.TotalSignal)
	fmt.Printf("  Total Sneak:       %.4f (normalized)\n", analysis.TotalSneak)
	maxSneakDisplay := analysis.MaxSneakRatio * 100
	maxSneakNote := ""
	if maxSneakDisplay > 100.0 {
		maxSneakNote = fmt.Sprintf(" (actual: %.1f%%)", maxSneakDisplay)
		maxSneakDisplay = 100.0
	}
	fmt.Printf("  Max Sneak/Signal:  %.2f%%%s\n", maxSneakDisplay, maxSneakNote)
	fmt.Printf("  Avg Sneak/Signal:  %.2f%%\n", analysis.AvgSneakRatio*100)

	// Signal-to-noise assessment
	if analysis.TotalSignal > 0 {
		snr := analysis.TotalSignal / (analysis.TotalSneak + 1e-10)
		fmt.Printf("  Signal/Sneak:      %.1f:1\n", snr)
	}

	// Legend
	fmt.Println("\n  Legend: · None  ░ Low  ▒ Medium  ▓ High  ★ Selected")

	// Impact assessment
	fmt.Println("\n  Impact Assessment:")
	if analysis.MaxSneakRatio < 0.01 {
		fmt.Println("  ✓ Sneak paths negligible - excellent isolation")
	} else if analysis.MaxSneakRatio < 0.05 {
		fmt.Println("  ⚠ Moderate sneak paths - consider selector devices")
	} else {
		fmt.Println("  ✗ Significant sneak paths - 1T1R or selector required")
	}
}

func (v *TerminalVisualizer) ShowMVMSneakTrace(report *crossbar.MVMSneakTraceReport) {
	fmt.Println("\n=== MVM Sneak Path Current Trace ===")
	if report == nil {
		fmt.Println("No trace data available")
		return
	}
	fmt.Printf("Architecture: %s\n", report.Architecture)
	fmt.Printf("Total sneak current: %.6f\n", report.TotalSneak)
	fmt.Printf("Peak row: %d (%.6f)\n", report.PeakRow, report.PeakCurrent)
	for _, row := range report.Rows {
		fmt.Printf("  Row %2d: I_sneak=%.6f\n", row.Row, row.SneakCurrent)
		for i, path := range row.TopPaths {
			fmt.Printf("    #%d src(r=%d,c=%d) -> exit(c=%d), G=%.6f, I=%.6f\n", i+1, path.SourceRow, path.SourceCol, path.ExitCol, path.PathG, path.PathCurrent)
		}
	}
}

// ShowMVMWithNonidealities shows MVM operation with non-ideality effects.
func (v *TerminalVisualizer) ShowMVMWithNonidealities(input, idealOutput, actualOutput []float64, irAnalysis *crossbar.IRDropAnalysis) {
	fmt.Println("\n=== MVM with Non-Idealities ===")

	// Show comparison
	maxOutput := min(len(idealOutput), 16)

	fmt.Println("\nIdeal vs Actual Output Comparison:")
	fmt.Println("  Idx    Ideal    Actual   Error")
	fmt.Println("  " + strings.Repeat("─", 36))

	var totalError float64
	for i := 0; i < maxOutput; i++ {
		err := math.Abs(idealOutput[i] - actualOutput[i])
		totalError += err * err
		errBar := strings.Repeat("█", int(err*50))
		if v.useColor {
			if err > 0.1 {
				fmt.Printf("  %3d   %.3f    %.3f   \033[91m%.3f %s\033[0m\n",
					i, idealOutput[i], actualOutput[i], err, errBar)
			} else {
				fmt.Printf("  %3d   %.3f    %.3f   \033[92m%.3f %s\033[0m\n",
					i, idealOutput[i], actualOutput[i], err, errBar)
			}
		} else {
			fmt.Printf("  %3d   %.3f    %.3f   %.3f %s\n",
				i, idealOutput[i], actualOutput[i], err, errBar)
		}
	}

	if len(idealOutput) > maxOutput {
		fmt.Printf("  ... (%d more outputs)\n", len(idealOutput)-maxOutput)
	}

	// Summary
	rmse := math.Sqrt(totalError / float64(len(idealOutput)))
	fmt.Printf("\n  RMSE: %.4f\n", rmse)
	fmt.Printf("  Max IR Drop: %.2f%%\n", irAnalysis.MaxIRDrop*100)

	// Visual comparison bars
	fmt.Println("\nOutput Vector Comparison:")
	fmt.Print("  Ideal:  [")
	for i := 0; i < min(len(idealOutput), 20); i++ {
		bar := v.valueToBar(idealOutput[i])
		if v.useColor {
			fmt.Printf("\033[94m%s\033[0m", bar)
		} else {
			fmt.Print(bar)
		}
	}
	fmt.Println("]")

	fmt.Print("  Actual: [")
	for i := 0; i < min(len(actualOutput), 20); i++ {
		bar := v.valueToBar(actualOutput[i])
		if v.useColor {
			fmt.Printf("\033[93m%s\033[0m", bar)
		} else {
			fmt.Print(bar)
		}
	}
	fmt.Println("]")
}

func (v *TerminalVisualizer) irDropToChar(value float64) string {
	if value < 0.2 {
		return "░"
	} else if value < 0.4 {
		return "▒"
	} else if value < 0.7 {
		return "▓"
	}
	return "█"
}

func (v *TerminalVisualizer) irDropToColor(value float64) string {
	if value < 0.2 {
		return "\033[92m" // Green (low drop)
	} else if value < 0.4 {
		return "\033[93m" // Yellow
	} else if value < 0.7 {
		return "\033[91m" // Red
	}
	return "\033[95m" // Magenta (severe)
}

func (v *TerminalVisualizer) sneakToChar(value float64) string {
	if value < 0.1 {
		return "·"
	} else if value < 0.3 {
		return "░"
	} else if value < 0.6 {
		return "▒"
	}
	return "▓"
}

func (v *TerminalVisualizer) sneakToColor(value float64) string {
	if value < 0.1 {
		return "\033[90m" // Dark gray (none)
	} else if value < 0.3 {
		return "\033[94m" // Blue
	} else if value < 0.6 {
		return "\033[93m" // Yellow
	}
	return "\033[91m" // Red (high sneak)
}

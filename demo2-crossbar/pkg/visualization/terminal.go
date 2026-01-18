// Package visualization provides terminal-based visualization for crossbar arrays.
package visualization

import (
	"fmt"
	"math"
	"strings"

	"ironlattice-vis/demo2-crossbar/pkg/crossbar"
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
	fmt.Println("\nLevel Legend (30 discrete states):")
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

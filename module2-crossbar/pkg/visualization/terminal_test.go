package visualization

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/crossbar"
)

// Helper function to create a test array
func createTestArray(t *testing.T, rows, cols int) *crossbar.Array {
	t.Helper()
	cfg := &crossbar.Config{
		Rows:       rows,
		Cols:       cols,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}
	array, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}
	return array
}

// captureOutput captures stdout during function execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestNewTerminalVisualizer tests the constructor
func TestNewTerminalVisualizer(t *testing.T) {
	tests := []struct {
		name     string
		rows     int
		cols     int
		useColor bool
	}{
		{
			name:     "with color enabled",
			rows:     4,
			cols:     4,
			useColor: true,
		},
		{
			name:     "with color disabled",
			rows:     4,
			cols:     4,
			useColor: false,
		},
		{
			name:     "with 1x1 array",
			rows:     1,
			cols:     1,
			useColor: true,
		},
		{
			name:     "with large array",
			rows:     64,
			cols:     64,
			useColor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			array := createTestArray(t, tt.rows, tt.cols)
			defer array.Destroy()

			vis := NewTerminalVisualizer(array, tt.useColor)
			if vis == nil {
				t.Fatal("NewTerminalVisualizer returned nil")
			}
			if vis.array != array {
				t.Error("array not set correctly")
			}
			if vis.useColor != tt.useColor {
				t.Errorf("useColor = %v, want %v", vis.useColor, tt.useColor)
			}
		})
	}
}

// TestNewTerminalVisualizerNilArray tests handling of nil array
func TestNewTerminalVisualizerNilArray(t *testing.T) {
	vis := NewTerminalVisualizer(nil, false)
	if vis == nil {
		t.Fatal("NewTerminalVisualizer returned nil")
	}
	if vis.array != nil {
		t.Error("expected nil array to be stored")
	}
}

// TestShowCrossbarState tests the crossbar state display
func TestShowCrossbarState(t *testing.T) {
	tests := []struct {
		name          string
		rows          int
		cols          int
		useColor      bool
		expectedParts []string
	}{
		{
			name:     "small array without color",
			rows:     4,
			cols:     4,
			useColor: false,
			expectedParts: []string{
				"=== Crossbar Array State ===",
				"Size: 4 x 4",
				"Level Legend",
				"30-level baseline",
			},
		},
		{
			name:     "small array with color",
			rows:     4,
			cols:     4,
			useColor: true,
			expectedParts: []string{
				"=== Crossbar Array State ===",
				"Size: 4 x 4",
				"\033[", // ANSI color code
			},
		},
		{
			name:     "large array exceeding display limits",
			rows:     64,
			cols:     64,
			useColor: false,
			expectedParts: []string{
				"Size: 64 x 64",
				"...",
				"(more rows)",
			},
		},
		{
			name:     "1x1 array",
			rows:     1,
			cols:     1,
			useColor: false,
			expectedParts: []string{
				"Size: 1 x 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			array := createTestArray(t, tt.rows, tt.cols)
			defer array.Destroy()

			vis := NewTerminalVisualizer(array, tt.useColor)

			output := captureOutput(func() {
				vis.ShowCrossbarState()
			})

			for _, part := range tt.expectedParts {
				if !strings.Contains(output, part) {
					t.Errorf("output missing expected part: %q", part)
				}
			}

			// Verify header numbers are present
			if !strings.Contains(output, " 0 ") {
				t.Error("output missing column header")
			}
		})
	}
}

// TestShowMVMOperation tests the MVM operation display
func TestShowMVMOperation(t *testing.T) {
	tests := []struct {
		name          string
		input         []float64
		output        []float64
		useColor      bool
		expectedParts []string
	}{
		{
			name:     "standard MVM",
			input:    []float64{0.1, 0.5, 0.9, 0.2},
			output:   []float64{0.3, 0.7, 0.4, 0.8},
			useColor: false,
			expectedParts: []string{
				"=== Matrix-Vector Multiplication ===",
				"Input Voltages (V):",
				"Output Currents (I):",
				"[ Crossbar W ] × [ V ]",
			},
		},
		{
			name:     "with color enabled",
			input:    []float64{0.5, 0.5},
			output:   []float64{0.5, 0.5},
			useColor: true,
			expectedParts: []string{
				"\033[94m", // Blue color for input
				"\033[93m", // Yellow color for output
			},
		},
		{
			name:     "large vectors",
			input:    make([]float64, 32),
			output:   make([]float64, 32),
			useColor: false,
			expectedParts: []string{
				"Input (32 values):",
				"Output (32 values):",
				"...",
			},
		},
		{
			name:     "empty vectors",
			input:    []float64{},
			output:   []float64{},
			useColor: false,
			expectedParts: []string{
				"Input (0 values):",
				"Output (0 values):",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			array := createTestArray(t, 4, 4)
			defer array.Destroy()

			vis := NewTerminalVisualizer(array, tt.useColor)

			output := captureOutput(func() {
				vis.ShowMVMOperation(tt.input, tt.output)
			})

			for _, part := range tt.expectedParts {
				if !strings.Contains(output, part) {
					t.Errorf("output missing expected part: %q", part)
				}
			}
		})
	}
}

// TestShowNeuralNetworkInference tests neural network inference display
func TestShowNeuralNetworkInference(t *testing.T) {
	tests := []struct {
		name          string
		layers        int
		input         []float64
		activations   [][]float64
		prediction    int
		confidence    float64
		expectedParts []string
	}{
		{
			name:        "MNIST input",
			layers:      2,
			input:       make([]float64, 784),
			activations: [][]float64{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}},
			prediction:  5,
			confidence:  0.95,
			expectedParts: []string{
				"=== Neural Network Inference ===",
				"Input Image (28x28):",
				"Layer 1 Activations",
				"Layer 2 Activations",
				"=== Prediction ===",
				"Predicted digit: 5",
				"confidence: 95.0%",
			},
		},
		{
			name:        "non-MNIST input",
			layers:      1,
			input:       []float64{0.1, 0.2, 0.3},
			activations: [][]float64{{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}},
			prediction:  9,
			confidence:  0.87,
			expectedParts: []string{
				"=== Neural Network Inference ===",
				"Layer 1 Activations",
				"Predicted digit: 9",
			},
		},
		{
			name:        "no activations",
			layers:      0,
			input:       []float64{0.5},
			activations: [][]float64{},
			prediction:  0,
			confidence:  0.5,
			expectedParts: []string{
				"Predicted digit: 0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			array := createTestArray(t, 4, 4)
			defer array.Destroy()

			vis := NewTerminalVisualizer(array, false)

			output := captureOutput(func() {
				vis.ShowNeuralNetworkInference(tt.layers, tt.input, tt.activations, tt.prediction, tt.confidence)
			})

			for _, part := range tt.expectedParts {
				if !strings.Contains(output, part) {
					t.Errorf("output missing expected part: %q", part)
				}
			}
		})
	}
}

// TestShowIRDropAnalysis tests IR drop analysis display
func TestShowIRDropAnalysis(t *testing.T) {
	array := createTestArray(t, 8, 8)
	defer array.Destroy()

	input := make([]float64, 8)
	for i := range input {
		input[i] = 0.5
	}

	params := crossbar.DefaultWireParams()
	analysis := array.AnalyzeIRDrop(input, params)

	tests := []struct {
		name          string
		useColor      bool
		expectedParts []string
	}{
		{
			name:     "without color",
			useColor: false,
			expectedParts: []string{
				"=== IR Drop Analysis ===",
				"Wire resistance effects",
				"IR Drop Heatmap",
				"IR Drop Statistics:",
				"Max IR Drop:",
				"Avg IR Drop:",
				"Variance:",
				"Legend:",
				"Impact Assessment:",
			},
		},
		{
			name:     "with color",
			useColor: true,
			expectedParts: []string{
				"=== IR Drop Analysis ===",
				"\033[", // ANSI color codes present
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vis := NewTerminalVisualizer(array, tt.useColor)

			output := captureOutput(func() {
				vis.ShowIRDropAnalysis(analysis)
			})

			for _, part := range tt.expectedParts {
				if !strings.Contains(output, part) {
					t.Errorf("output missing expected part: %q", part)
				}
			}
		})
	}
}

// TestShowSneakPathAnalysis tests sneak path analysis display
func TestShowSneakPathAnalysis(t *testing.T) {
	array := createTestArray(t, 8, 8)
	defer array.Destroy()

	selectedRow, selectedCol := 3, 4
	analysis := array.AnalyzeSneakPaths(selectedRow, selectedCol)

	tests := []struct {
		name          string
		useColor      bool
		expectedParts []string
	}{
		{
			name:     "without color",
			useColor: false,
			expectedParts: []string{
				"=== Sneak Path Analysis ===",
				"Selected cell: [3, 4]",
				"Sneak Current Map",
				"Sneak Path Statistics:",
				"Signal Current:",
				"Total Sneak:",
				"Max Sneak/Signal:",
				"Avg Sneak/Signal:",
				"Legend:",
				"★ Selected",
				"Impact Assessment:",
			},
		},
		{
			name:     "with color",
			useColor: true,
			expectedParts: []string{
				"Selected cell:",
				"\033[92m★\033[0m", // Colored star marker
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vis := NewTerminalVisualizer(array, tt.useColor)

			output := captureOutput(func() {
				vis.ShowSneakPathAnalysis(analysis, selectedRow, selectedCol)
			})

			for _, part := range tt.expectedParts {
				if !strings.Contains(output, part) {
					t.Errorf("output missing expected part: %q", part)
				}
			}
		})
	}
}

// TestShowMVMWithNonidealities tests MVM with non-idealities display
func TestShowMVMWithNonidealities(t *testing.T) {
	array := createTestArray(t, 8, 8)
	defer array.Destroy()

	input := make([]float64, 8)
	for i := range input {
		input[i] = 0.5
	}

	idealOutput := make([]float64, 8)
	actualOutput := make([]float64, 8)
	for i := range idealOutput {
		idealOutput[i] = float64(i) * 0.1
		actualOutput[i] = idealOutput[i] * 0.95 // 5% error
	}

	params := crossbar.DefaultWireParams()
	irAnalysis := array.AnalyzeIRDrop(input, params)

	tests := []struct {
		name          string
		useColor      bool
		expectedParts []string
	}{
		{
			name:     "without color",
			useColor: false,
			expectedParts: []string{
				"=== MVM with Non-Idealities ===",
				"Ideal vs Actual Output Comparison:",
				"Idx    Ideal    Actual   Error",
				"RMSE:",
				"Max IR Drop:",
				"Output Vector Comparison:",
				"Ideal:",
				"Actual:",
			},
		},
		{
			name:     "with color",
			useColor: true,
			expectedParts: []string{
				"MVM with Non-Idealities",
				"\033[94m", // Blue for ideal
				"\033[93m", // Yellow for actual
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vis := NewTerminalVisualizer(array, tt.useColor)

			output := captureOutput(func() {
				vis.ShowMVMWithNonidealities(input, idealOutput, actualOutput, irAnalysis)
			})

			for _, part := range tt.expectedParts {
				if !strings.Contains(output, part) {
					t.Errorf("output missing expected part: %q", part)
				}
			}
		})
	}
}

// TestLevelToChar tests the level to character conversion
func TestLevelToChar(t *testing.T) {
	array := createTestArray(t, 2, 2)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, false)

	tests := []struct {
		level        int
		expectedChar string
	}{
		{0, " "},   // Lowest level
		{7, "░"},   // Low level
		{15, "▒"},  // Medium level
		{22, "▓"},  // High level
		{29, "█"},  // Highest level
		{30, "█"},  // Beyond max (clamped)
		{100, "█"}, // Far beyond max (clamped)
	}

	for _, tt := range tests {
		t.Run(tt.expectedChar, func(t *testing.T) {
			result := vis.levelToChar(tt.level)
			if result != tt.expectedChar {
				t.Errorf("levelToChar(%d) = %q, want %q", tt.level, result, tt.expectedChar)
			}
		})
	}
}

// TestLevelToColor tests the level to color conversion
func TestLevelToColor(t *testing.T) {
	array := createTestArray(t, 2, 2)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, false)

	tests := []struct {
		level         int
		expectedColor string
	}{
		{0, "\033[94m"},  // Blue (low)
		{9, "\033[94m"},  // Blue
		{10, "\033[97m"}, // White (mid)
		{19, "\033[97m"}, // White
		{20, "\033[91m"}, // Red (high)
		{29, "\033[91m"}, // Red
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.level)), func(t *testing.T) {
			result := vis.levelToColor(tt.level)
			if result != tt.expectedColor {
				t.Errorf("levelToColor(%d) = %q, want %q", tt.level, result, tt.expectedColor)
			}
		})
	}
}

// TestValueToBar tests the value to bar character conversion
func TestValueToBar(t *testing.T) {
	array := createTestArray(t, 2, 2)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, false)

	tests := []struct {
		value       float64
		expectedBar string
	}{
		{0.0, "▁"},  // 0.0 * 7 = 0 -> bars[0]
		{0.14, "▁"}, // 0.14 * 7 = 0.98 -> int(0.98) = 0 -> bars[0]
		{0.28, "▂"}, // 0.28 * 7 = 1.96 -> int(1.96) = 1 -> bars[1]
		{0.42, "▃"}, // 0.42 * 7 = 2.94 -> int(2.94) = 2 -> bars[2]
		{0.57, "▄"}, // 0.57 * 7 = 3.99 -> int(3.99) = 3 -> bars[3]
		{0.71, "▅"}, // 0.71 * 7 = 4.97 -> int(4.97) = 4 -> bars[4]
		{0.85, "▆"}, // 0.85 * 7 = 5.95 -> int(5.95) = 5 -> bars[5]
		{1.0, "█"},  // 1.0 * 7 = 7 -> int(7) = 7 -> bars[7]
		{-0.5, "▁"}, // Clamped to 0
		{1.5, "█"},  // Clamped to 1
	}

	for _, tt := range tests {
		t.Run(tt.expectedBar, func(t *testing.T) {
			result := vis.valueToBar(tt.value)
			if result != tt.expectedBar {
				t.Errorf("valueToBar(%f) = %q, want %q", tt.value, result, tt.expectedBar)
			}
		})
	}
}

// TestIRDropToChar tests IR drop to character conversion
func TestIRDropToChar(t *testing.T) {
	array := createTestArray(t, 2, 2)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, false)

	tests := []struct {
		value        float64
		expectedChar string
	}{
		{0.0, "░"},
		{0.1, "░"},
		{0.2, "▒"},
		{0.3, "▒"},
		{0.4, "▓"},
		{0.6, "▓"},
		{0.7, "█"},
		{1.0, "█"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedChar, func(t *testing.T) {
			result := vis.irDropToChar(tt.value)
			if result != tt.expectedChar {
				t.Errorf("irDropToChar(%f) = %q, want %q", tt.value, result, tt.expectedChar)
			}
		})
	}
}

// TestIRDropToColor tests IR drop to color conversion
func TestIRDropToColor(t *testing.T) {
	array := createTestArray(t, 2, 2)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, false)

	tests := []struct {
		value         float64
		expectedColor string
	}{
		{0.0, "\033[92m"}, // Green
		{0.1, "\033[92m"},
		{0.2, "\033[93m"}, // Yellow
		{0.3, "\033[93m"},
		{0.4, "\033[91m"}, // Red
		{0.6, "\033[91m"},
		{0.7, "\033[95m"}, // Magenta
		{1.0, "\033[95m"},
	}

	for _, tt := range tests {
		t.Run(string(rune(int(tt.value*100))), func(t *testing.T) {
			result := vis.irDropToColor(tt.value)
			if result != tt.expectedColor {
				t.Errorf("irDropToColor(%f) = %q, want %q", tt.value, result, tt.expectedColor)
			}
		})
	}
}

// TestSneakToChar tests sneak path to character conversion
func TestSneakToChar(t *testing.T) {
	array := createTestArray(t, 2, 2)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, false)

	tests := []struct {
		value        float64
		expectedChar string
	}{
		{0.0, "·"},
		{0.05, "·"},
		{0.1, "░"},
		{0.2, "░"},
		{0.3, "▒"},
		{0.5, "▒"},
		{0.6, "▓"},
		{1.0, "▓"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedChar, func(t *testing.T) {
			result := vis.sneakToChar(tt.value)
			if result != tt.expectedChar {
				t.Errorf("sneakToChar(%f) = %q, want %q", tt.value, result, tt.expectedChar)
			}
		})
	}
}

// TestSneakToColor tests sneak path to color conversion
func TestSneakToColor(t *testing.T) {
	array := createTestArray(t, 2, 2)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, false)

	tests := []struct {
		value         float64
		expectedColor string
	}{
		{0.0, "\033[90m"}, // Dark gray
		{0.05, "\033[90m"},
		{0.1, "\033[94m"}, // Blue
		{0.2, "\033[94m"},
		{0.3, "\033[93m"}, // Yellow
		{0.5, "\033[93m"},
		{0.6, "\033[91m"}, // Red
		{1.0, "\033[91m"},
	}

	for _, tt := range tests {
		t.Run(string(rune(int(tt.value*100))), func(t *testing.T) {
			result := vis.sneakToColor(tt.value)
			if result != tt.expectedColor {
				t.Errorf("sneakToColor(%f) = %q, want %q", tt.value, result, tt.expectedColor)
			}
		})
	}
}

// TestMinFunction tests the min helper function
func TestMinFunction(t *testing.T) {
	tests := []struct {
		a        int
		b        int
		expected int
	}{
		{5, 10, 5},
		{10, 5, 5},
		{7, 7, 7},
		{0, 5, 0},
		{-5, 3, -5},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// TestShowCrossbarStateWithDifferentValues tests display with various conductance values
func TestShowCrossbarStateWithDifferentValues(t *testing.T) {
	array := createTestArray(t, 4, 4)
	defer array.Destroy()

	// Set some specific conductance values using ProgramWeight
	array.ProgramWeight(0, 0, 0.0) // Min
	array.ProgramWeight(0, 1, 0.5) // Mid
	array.ProgramWeight(0, 2, 1.0) // Max

	vis := NewTerminalVisualizer(array, false)

	output := captureOutput(func() {
		vis.ShowCrossbarState()
	})

	// Verify that different levels produce different characters
	if !strings.Contains(output, " ") || !strings.Contains(output, "█") {
		t.Error("expected to see both low and high level characters")
	}
}

// TestShowMVMOperationEdgeCases tests edge cases for MVM operation
func TestShowMVMOperationEdgeCases(t *testing.T) {
	array := createTestArray(t, 4, 4)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, false)

	t.Run("nil inputs", func(t *testing.T) {
		// Should not panic
		output := captureOutput(func() {
			vis.ShowMVMOperation(nil, nil)
		})
		if !strings.Contains(output, "Matrix-Vector Multiplication") {
			t.Error("expected MVM header even with nil inputs")
		}
	})

	t.Run("mismatched input/output sizes", func(t *testing.T) {
		input := []float64{1.0, 2.0}
		output := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

		output_text := captureOutput(func() {
			vis.ShowMVMOperation(input, output)
		})

		if !strings.Contains(output_text, "Input (2 values)") {
			t.Error("expected input size indication")
		}
		if !strings.Contains(output_text, "Output (5 values)") {
			t.Error("expected output size indication")
		}
	})
}

// TestNeuralNetworkInferenceWithColoredOutput tests colored neural network display
func TestNeuralNetworkInferenceWithColoredOutput(t *testing.T) {
	array := createTestArray(t, 4, 4)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, true)

	activations := [][]float64{
		{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
	}

	output := captureOutput(func() {
		vis.ShowNeuralNetworkInference(1, []float64{0.5}, activations, 5, 0.92)
	})

	// Should contain ANSI color codes
	if !strings.Contains(output, "\033[") {
		t.Error("expected ANSI color codes in colored output")
	}

	// Should highlight the predicted digit
	if !strings.Contains(output, "\033[92m") {
		t.Error("expected green highlighting for prediction")
	}
}

// TestLargeArrayDisplayLimits tests that large arrays are properly truncated
func TestLargeArrayDisplayLimits(t *testing.T) {
	array := createTestArray(t, 100, 100)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, false)

	output := captureOutput(func() {
		vis.ShowCrossbarState()
	})

	// Should show truncation indicators
	if !strings.Contains(output, "...") {
		t.Error("expected truncation indicator for large array")
	}

	// Should mention more rows
	if !strings.Contains(output, "more rows") {
		t.Error("expected 'more rows' indicator")
	}
}

// TestIRDropAnalysisImpactAssessment tests different impact assessment messages
func TestIRDropAnalysisImpactAssessment(t *testing.T) {
	// This is a white-box test verifying the impact assessment logic
	array := createTestArray(t, 8, 8)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, false)

	input := make([]float64, 8)
	for i := range input {
		input[i] = 0.5
	}

	params := crossbar.DefaultWireParams()
	analysis := array.AnalyzeIRDrop(input, params)

	output := captureOutput(func() {
		vis.ShowIRDropAnalysis(analysis)
	})

	// Should contain one of the assessment messages
	hasAssessment := strings.Contains(output, "minimal") ||
		strings.Contains(output, "Moderate") ||
		strings.Contains(output, "Severe")

	if !hasAssessment {
		t.Error("expected impact assessment message")
	}
}

// TestSneakPathAnalysisImpactAssessment tests sneak path impact messages
func TestSneakPathAnalysisImpactAssessment(t *testing.T) {
	array := createTestArray(t, 8, 8)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, false)

	analysis := array.AnalyzeSneakPaths(3, 3)

	output := captureOutput(func() {
		vis.ShowSneakPathAnalysis(analysis, 3, 3)
	})

	// Should contain one of the assessment messages
	hasAssessment := strings.Contains(output, "negligible") ||
		strings.Contains(output, "Moderate") ||
		strings.Contains(output, "Significant")

	if !hasAssessment {
		t.Error("expected impact assessment message")
	}
}

// TestMVMWithNonidealitiesRMSECalculation tests RMSE calculation display
func TestMVMWithNonidealitiesRMSECalculation(t *testing.T) {
	array := createTestArray(t, 4, 4)
	defer array.Destroy()

	vis := NewTerminalVisualizer(array, false)

	input := []float64{0.5, 0.5, 0.5, 0.5}
	idealOutput := []float64{1.0, 2.0, 3.0, 4.0}
	actualOutput := []float64{1.0, 2.0, 3.0, 4.0} // Perfect match

	params := crossbar.DefaultWireParams()
	irAnalysis := array.AnalyzeIRDrop(input, params)

	output := captureOutput(func() {
		vis.ShowMVMWithNonidealities(input, idealOutput, actualOutput, irAnalysis)
	})

	// Should show RMSE
	if !strings.Contains(output, "RMSE:") {
		t.Error("expected RMSE in output")
	}

	// Should show zero error for perfect match
	if !strings.Contains(output, "0.000") {
		t.Error("expected zero RMSE for identical outputs")
	}
}

//go:build ignore

// mnist_inference.go - Educational demonstration of MNIST inference on FeCIM
//
// This program demonstrates how FeCIM performs neural network inference
// for handwritten digit classification (MNIST dataset).
//
// Key Concepts:
//   - Neural network weights mapped to ferroelectric crossbar arrays
//   - Analog compute-in-memory performs matrix-vector multiplications
//   - Quantization-Aware Training (QAT) optimizes for discrete levels
//   - 30-level analog weights achieve ~93% accuracy on MNIST
//
// Architecture:
//   - Input: 28×28 = 784 pixels (grayscale image)
//   - Hidden: 128 neurons with ReLU activation
//   - Output: 10 neurons (digits 0-9) with softmax
//
// Run: go run examples/mnist_inference.go
//
// For full training and evaluation, see: module3-mnist/
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SimpleNetwork represents a 2-layer neural network for MNIST
type SimpleNetwork struct {
	// Layer 1: 784 → 128
	Weights1 [][]float64
	Bias1    []float64

	// Layer 2: 128 → 10
	Weights2 [][]float64
	Bias2    []float64

	// Quantization
	NumLevels int
}

// NewSimpleNetwork creates a new network with Xavier initialization
func NewSimpleNetwork(numLevels int) *SimpleNetwork {
	net := &SimpleNetwork{
		NumLevels: numLevels,
	}

	// Initialize Layer 1: 784 → 128
	net.Weights1 = make([][]float64, 128)
	net.Bias1 = make([]float64, 128)
	for i := range net.Weights1 {
		net.Weights1[i] = make([]float64, 784)
		// Xavier initialization: w ~ N(0, sqrt(2/(fan_in + fan_out)))
		scale := math.Sqrt(2.0 / float64(784+128))
		for j := range net.Weights1[i] {
			net.Weights1[i][j] = rand.NormFloat64() * scale
		}
	}

	// Initialize Layer 2: 128 → 10
	net.Weights2 = make([][]float64, 10)
	net.Bias2 = make([]float64, 10)
	for i := range net.Weights2 {
		net.Weights2[i] = make([]float64, 128)
		scale := math.Sqrt(2.0 / float64(128+10))
		for j := range net.Weights2[i] {
			net.Weights2[i][j] = rand.NormFloat64() * scale
		}
	}

	return net
}

// LoadWeights loads pretrained weights if available
func (net *SimpleNetwork) LoadWeights(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var weights struct {
		Layer1Weights [][]float64 `json:"layer1_weights"`
		Layer2Weights [][]float64 `json:"layer2_weights"`
		Biases1       []float64   `json:"biases1"`
		Biases2       []float64   `json:"biases2"`
	}

	if err := json.Unmarshal(data, &weights); err != nil {
		return err
	}

	if len(weights.Layer1Weights) > 0 {
		net.Weights1 = weights.Layer1Weights
	}
	if len(weights.Layer2Weights) > 0 {
		net.Weights2 = weights.Layer2Weights
	}
	if len(weights.Biases1) > 0 {
		net.Bias1 = weights.Biases1
	}
	if len(weights.Biases2) > 0 {
		net.Bias2 = weights.Biases2
	}

	return nil
}

// Quantize applies uniform quantization to a value
func (net *SimpleNetwork) Quantize(x float64) float64 {
	if net.NumLevels <= 1 {
		return x
	}

	// Assume weights are normalized to [-1, 1]
	// Map to discrete levels
	level := int(math.Round((x + 1) / 2 * float64(net.NumLevels-1)))
	if level < 0 {
		level = 0
	}
	if level >= net.NumLevels {
		level = net.NumLevels - 1
	}

	// Map back to [-1, 1]
	return float64(level)/float64(net.NumLevels-1)*2 - 1
}

// Forward performs inference with optional quantization
func (net *SimpleNetwork) Forward(input []float64, quantize bool) ([]float64, []float64) {
	// Layer 1: hidden = ReLU(W1 × input + b1)
	hidden := make([]float64, 128)
	for i := 0; i < 128; i++ {
		sum := net.Bias1[i]
		for j := 0; j < 784; j++ {
			w := net.Weights1[i][j]
			if quantize {
				w = net.Quantize(w)
			}
			sum += w * input[j]
		}
		// ReLU activation
		if sum < 0 {
			sum = 0
		}
		hidden[i] = sum
	}

	// Layer 2: output = W2 × hidden + b2
	output := make([]float64, 10)
	for i := 0; i < 10; i++ {
		sum := net.Bias2[i]
		for j := 0; j < 128; j++ {
			w := net.Weights2[i][j]
			if quantize {
				w = net.Quantize(w)
			}
			sum += w * hidden[j]
		}
		output[i] = sum
	}

	return hidden, output
}

// Softmax converts logits to probabilities
func Softmax(x []float64) []float64 {
	// Find max for numerical stability
	max := x[0]
	for _, v := range x[1:] {
		if v > max {
			max = v
		}
	}

	// Compute exp and sum
	probs := make([]float64, len(x))
	sum := 0.0
	for i, v := range x {
		probs[i] = math.Exp(v - max)
		sum += probs[i]
	}

	// Normalize
	for i := range probs {
		probs[i] /= sum
	}

	return probs
}

// Argmax returns the index of the maximum value
func Argmax(x []float64) int {
	maxIdx := 0
	maxVal := x[0]
	for i, v := range x[1:] {
		if v > maxVal {
			maxVal = v
			maxIdx = i + 1
		}
	}
	return maxIdx
}

// GenerateSampleDigit creates a simple test pattern for a digit
func GenerateSampleDigit(digit int) []float64 {
	img := make([]float64, 784)

	// Create simple patterns for each digit
	switch digit {
	case 0:
		// Circle-like pattern
		for y := 8; y < 20; y++ {
			for x := 8; x < 20; x++ {
				dx := float64(x-14) / 6
				dy := float64(y-14) / 6
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist > 0.7 && dist < 1.3 {
					img[y*28+x] = 1.0
				}
			}
		}
	case 1:
		// Vertical line
		for y := 6; y < 22; y++ {
			img[y*28+14] = 1.0
			img[y*28+13] = 0.5
		}
	case 2:
		// S-shaped curve (simplified)
		for x := 10; x < 18; x++ {
			img[7*28+x] = 1.0
		}
		for y := 7; y < 14; y++ {
			img[y*28+17] = 1.0
		}
		for x := 10; x < 18; x++ {
			img[14*28+x] = 1.0
		}
		for y := 14; y < 21; y++ {
			img[y*28+10] = 1.0
		}
		for x := 10; x < 18; x++ {
			img[20*28+x] = 1.0
		}
	case 7:
		// Top bar and diagonal
		for x := 8; x < 20; x++ {
			img[7*28+x] = 1.0
		}
		for i := 0; i < 14; i++ {
			x := 19 - i/2
			y := 7 + i
			if x >= 0 && x < 28 && y >= 0 && y < 28 {
				img[y*28+x] = 1.0
			}
		}
	default:
		// Random noise for other digits (not trained patterns)
		for i := range img {
			if rand.Float64() < 0.1 {
				img[i] = rand.Float64()
			}
		}
	}

	return img
}

// PrintDigit renders a 28×28 image as ASCII art
func PrintDigit(img []float64, label string) {
	fmt.Printf("   %s:\n", label)
	fmt.Println("   ┌" + strings.Repeat("─", 28) + "┐")
	for y := 0; y < 28; y++ {
		fmt.Print("   │")
		for x := 0; x < 28; x++ {
			v := img[y*28+x]
			if v > 0.75 {
				fmt.Print("█")
			} else if v > 0.5 {
				fmt.Print("▓")
			} else if v > 0.25 {
				fmt.Print("▒")
			} else if v > 0.1 {
				fmt.Print("░")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println("│")
	}
	fmt.Println("   └" + strings.Repeat("─", 28) + "┘")
}

// PrintProbabilityBar renders a horizontal bar chart of probabilities
func PrintProbabilityBar(probs []float64, prediction int) {
	fmt.Println("\n   Prediction Probabilities:")
	fmt.Println("   ┌────────────────────────────────────────────────────┐")

	for i := 0; i < 10; i++ {
		marker := " "
		if i == prediction {
			marker = "►"
		}

		// Bar chart
		barLen := int(probs[i] * 40)
		bar := strings.Repeat("█", barLen) + strings.Repeat("░", 40-barLen)

		fmt.Printf("   │%s %d │%s│ %5.1f%% │\n", marker, i, bar, probs[i]*100)
	}

	fmt.Println("   └────────────────────────────────────────────────────┘")
}

func main() {
	rand.Seed(42) // For reproducibility

	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║     FeCIM MNIST Neural Network Inference Demo                    ║")
	fmt.Println("║     Handwritten Digit Classification on Analog Hardware          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// =========================================================================
	// 1. NETWORK ARCHITECTURE
	// =========================================================================
	fmt.Println("1. NEURAL NETWORK ARCHITECTURE")
	fmt.Println()
	printArchitectureDiagram()

	// =========================================================================
	// 2. CREATE NETWORK
	// =========================================================================
	fmt.Println("\n2. CREATING NETWORK")
	fmt.Println()

	numLevels := 30 // FeCIM's 30 analog levels
	net := NewSimpleNetwork(numLevels)

	// Try to load pretrained weights
	weightsPath := filepath.Join("module3-mnist", "data", "pretrained_weights.json")
	if err := net.LoadWeights(weightsPath); err != nil {
		fmt.Printf("   Note: No pretrained weights found at %s\n", weightsPath)
		fmt.Println("   Using random initialization (demo mode)")
		fmt.Println("   Run: go run module3-mnist/train_mnist_proper.go to train")
	} else {
		fmt.Println("   ✓ Loaded pretrained weights")
	}

	fmt.Println()
	fmt.Printf("   Quantization: %d levels (%.1f bits per weight)\n",
		numLevels, math.Log2(float64(numLevels)))
	fmt.Printf("   Layer 1: 784 → 128 neurons (%d weights)\n", 784*128)
	fmt.Printf("   Layer 2: 128 → 10 neurons (%d weights)\n", 128*10)
	fmt.Printf("   Total parameters: %d\n", 784*128+128+128*10+10)

	// =========================================================================
	// 3. INFERENCE DEMONSTRATION
	// =========================================================================
	fmt.Println("\n\n3. INFERENCE DEMONSTRATION")
	fmt.Println("   └─ Running inference on sample digits")
	fmt.Println()

	testDigits := []int{0, 1, 2, 7}

	for _, digit := range testDigits {
		img := GenerateSampleDigit(digit)

		PrintDigit(img, fmt.Sprintf("Sample Digit %d", digit))

		// Full precision inference
		_, outputFP := net.Forward(img, false)
		probsFP := Softmax(outputFP)
		predFP := Argmax(probsFP)

		// Quantized inference (simulating FeCIM)
		_, outputQ := net.Forward(img, true)
		probsQ := Softmax(outputQ)
		predQ := Argmax(probsQ)

		fmt.Printf("\n   Full Precision:  Prediction = %d (confidence: %.1f%%)\n",
			predFP, probsFP[predFP]*100)
		fmt.Printf("   Quantized (30L): Prediction = %d (confidence: %.1f%%)\n",
			predQ, probsQ[predQ]*100)

		if predFP == predQ {
			fmt.Println("   ✓ FP and Quantized agree!")
		} else {
			fmt.Println("   △ Different predictions (quantization effect)")
		}

		PrintProbabilityBar(probsQ, predQ)
		fmt.Println()
	}

	// =========================================================================
	// 4. QUANTIZATION ANALYSIS
	// =========================================================================
	fmt.Println("\n4. QUANTIZATION LEVEL ANALYSIS")
	fmt.Println("   └─ Impact of analog precision on inference")
	fmt.Println()

	analyzeQuantizationImpact(net)

	// =========================================================================
	// 5. ENERGY ESTIMATION
	// =========================================================================
	fmt.Println("\n\n5. ENERGY ANALYSIS")
	fmt.Println()

	printEnergyAnalysis(numLevels)

	// =========================================================================
	// 6. HARDWARE MAPPING
	// =========================================================================
	fmt.Println("\n\n6. HARDWARE MAPPING TO CROSSBAR")
	fmt.Println()

	printHardwareMapping()

	fmt.Println("\n═══════════════════════════════════════════════════════════════════")
	fmt.Println("Demo complete! For full MNIST training and evaluation:")
	fmt.Println("  cd module3-mnist && go run train_mnist_proper.go")
	fmt.Println("For interactive inference GUI:")
	fmt.Println("  fecim-lattice-tools mnist")
}

func printArchitectureDiagram() {
	diagram := `
     Input Layer          Hidden Layer         Output Layer
     (28×28 pixels)       (128 neurons)        (10 classes)

     ┌───────────┐
     │ ███  ███  │
     │  ██  ███  │         ┌─────────┐
     │   █  ██   │   →→→   │ ReLU(.) │   →→→   [0] P(digit=0)
     │    ██     │         │ neurons │         [1] P(digit=1)
     │   ████    │         │  ×128   │         [2] P(digit=2)
     │  ██  ███  │         └─────────┘         ...
     │ ███  ███  │              │              [9] P(digit=9)
     └───────────┘              ↓
                                                  ↓
         784                   128             Softmax → Prediction

    ═══════════════════════════════════════════════════════════════
    Layer 1: Y₁ = ReLU(W₁ × X + b₁)     [784×128 matrix multiply]
    Layer 2: Y₂ = W₂ × Y₁ + b₂          [128×10 matrix multiply]
    Output:  P = Softmax(Y₂)            [probability distribution]
    ═══════════════════════════════════════════════════════════════
`
	fmt.Print(diagram)
}

func analyzeQuantizationImpact(net *SimpleNetwork) {
	// Test with different quantization levels
	levels := []int{2, 4, 8, 16, 30, 64, 128}

	img := GenerateSampleDigit(7) // Use digit 7 for testing

	// Get FP reference
	_, outputFP := net.Forward(img, false)
	probsFP := Softmax(outputFP)

	fmt.Println("   Levels │ Bits │ Confidence │ KL Divergence │ Energy/MAC")
	fmt.Println("   ───────┼──────┼────────────┼───────────────┼───────────")

	for _, L := range levels {
		net.NumLevels = L

		_, outputQ := net.Forward(img, true)
		probsQ := Softmax(outputQ)

		// Compute KL divergence
		kl := 0.0
		for i := range probsFP {
			if probsFP[i] > 1e-10 && probsQ[i] > 1e-10 {
				kl += probsFP[i] * math.Log(probsFP[i]/probsQ[i])
			}
		}

		predQ := Argmax(probsQ)
		bits := math.Log2(float64(L))
		energy := 10.0 * bits // fJ per MAC

		marker := ""
		if L == 30 {
			marker = " ← FeCIM"
		}

		fmt.Printf("   %4d   │ %4.1f │  %6.1f%%   │    %.4f      │  %5.1f fJ%s\n",
			L, bits, probsQ[predQ]*100, kl, energy, marker)
	}

	net.NumLevels = 30 // Reset
}

func printEnergyAnalysis(numLevels int) {
	macs1 := 784 * 128
	macs2 := 128 * 10
	totalMACs := macs1 + macs2

	bitsPerWeight := math.Log2(float64(numLevels))
	energyPerMAC := 10.0 * bitsPerWeight // fJ

	totalEnergy := float64(totalMACs) * energyPerMAC / 1e6 // µJ

	fmt.Println("   ┌─────────────────────────────────────────────────────────┐")
	fmt.Println("   │              ENERGY BREAKDOWN                           │")
	fmt.Println("   ├─────────────────────────────────────────────────────────┤")
	fmt.Printf("   │  Layer 1 MACs:     %6d × %3d = %,8d            │\n", 784, 128, macs1)
	fmt.Printf("   │  Layer 2 MACs:     %6d × %3d = %,8d              │\n", 128, 10, macs2)
	fmt.Printf("   │  Total MACs:                    %,8d            │\n", totalMACs)
	fmt.Println("   ├─────────────────────────────────────────────────────────┤")
	fmt.Printf("   │  Bits per weight:  %.1f (30 levels)                      │\n", bitsPerWeight)
	fmt.Printf("   │  Energy per MAC:   %.1f fJ                               │\n", energyPerMAC)
	fmt.Printf("   │  Total Energy:     %.3f µJ per inference                │\n", totalEnergy)
	fmt.Println("   ├─────────────────────────────────────────────────────────┤")
	fmt.Println("   │  Comparison to Digital (32-bit):                        │")
	digitalEnergy := float64(totalMACs) * 10.0 / 1e6 // ~10 pJ per MAC
	speedup := digitalEnergy / totalEnergy
	fmt.Printf("   │    Digital: ~%.2f µJ  →  FeCIM: ~%.3f µJ               │\n",
		digitalEnergy*1000, totalEnergy)
	fmt.Printf("   │    Energy reduction: ~%.0f×                              │\n", speedup)
	fmt.Println("   └─────────────────────────────────────────────────────────┘")
}

func printHardwareMapping() {
	fmt.Println("   CROSSBAR ARRAY MAPPING:")
	fmt.Println()
	fmt.Println("   Layer 1: 784×128 = 100,352 FeCIM cells")
	fmt.Println("   ┌────────────────────────────────────────────────┐")
	fmt.Println("   │   Crossbar 1a   │   Crossbar 1b   │    ...     │")
	fmt.Println("   │   (392×128)     │   (392×128)     │            │")
	fmt.Println("   │   Word lines:   │   Word lines:   │            │")
	fmt.Println("   │   pixels 0-391  │   pixels 392+   │            │")
	fmt.Println("   └────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("   Layer 2: 128×10 = 1,280 FeCIM cells")
	fmt.Println("   ┌─────────────────────────────────────────┐")
	fmt.Println("   │          Crossbar 2                     │")
	fmt.Println("   │          (128×10)                       │")
	fmt.Println("   │   Word lines: hidden activations        │")
	fmt.Println("   │   Bit lines: output classes             │")
	fmt.Println("   └─────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("   Key Insight: Each layer is a single crossbar MVM operation!")
	fmt.Println("   All 100K+ weights computed in parallel (O(1) time complexity)")
}

// Helper functions
func init() {
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			fmt.Println("FeCIM MNIST Demo - Neural network inference on analog hardware")
			fmt.Println()
			fmt.Println("Usage: go run examples/mnist_inference.go")
			fmt.Println()
			fmt.Println("This program demonstrates:")
			fmt.Println("  • MNIST neural network architecture")
			fmt.Println("  • Quantized inference simulation")
			fmt.Println("  • Energy analysis for FeCIM")
			fmt.Println("  • Hardware mapping to crossbar arrays")
			fmt.Println()
			fmt.Println("For training: go run module3-mnist/train_mnist_proper.go")
			fmt.Println("For GUI:      fecim-lattice-tools mnist")
			os.Exit(0)
		}
	}
}

func sortedKeys(m map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

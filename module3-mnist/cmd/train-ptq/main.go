// Command train-ptq trains MNIST network and applies Post-Training Quantization (PTQ)
// with different quantization levels per layer for optimal accuracy.
package mnisttrainptq

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/shared/canvas"
)

// Network is a simple 2-layer network with per-layer quantization support
type Network struct {
	W1     [][]float64 // [hidden][784]
	W2     [][]float64 // [10][hidden]
	B1     []float64   // [hidden]
	B2     []float64   // [10]
	Hidden int
}

// PTQConfig holds per-layer quantization configuration
type PTQConfig struct {
	Layer1Levels int // Quantization levels for layer 1 (hidden layer)
	Layer2Levels int // Quantization levels for layer 2 (output layer)
}

func NewNetwork(hidden int) *Network {
	n := &Network{
		Hidden: hidden,
		W1:     make([][]float64, hidden),
		W2:     make([][]float64, 10),
		B1:     make([]float64, hidden),
		B2:     make([]float64, 10),
	}

	// Xavier initialization
	scale1 := math.Sqrt(2.0 / float64(784+hidden))
	for i := 0; i < hidden; i++ {
		n.W1[i] = make([]float64, 784)
		for j := 0; j < 784; j++ {
			n.W1[i][j] = rand.NormFloat64() * scale1
		}
	}

	scale2 := math.Sqrt(2.0 / float64(hidden+10))
	for i := 0; i < 10; i++ {
		n.W2[i] = make([]float64, hidden)
		for j := 0; j < hidden; j++ {
			n.W2[i][j] = rand.NormFloat64() * scale2
		}
	}

	return n
}

// quantize to N levels in range [min, max]
func quantize(val, min, max float64, levels int) float64 {
	if levels <= 1 {
		return (min + max) / 2 // Return midpoint for 1 level
	}
	norm := (val - min) / (max - min)
	norm = math.Max(0, math.Min(1, norm))
	step := 1.0 / float64(levels-1)
	quantized := math.Round(norm/step) * step
	return quantized*(max-min) + min
}

// getWeightRange returns min/max for a weight matrix
func getWeightRange(w [][]float64) (float64, float64) {
	wMin, wMax := w[0][0], w[0][0]
	for i := range w {
		for j := range w[i] {
			if w[i][j] < wMin {
				wMin = w[i][j]
			}
			if w[i][j] > wMax {
				wMax = w[i][j]
			}
		}
	}
	return wMin, wMax
}

// ForwardPTQ performs forward pass with per-layer PTQ
func (n *Network) ForwardPTQ(input []float64, config PTQConfig) []float64 {
	w1Min, w1Max := getWeightRange(n.W1)
	w2Min, w2Max := getWeightRange(n.W2)

	// Layer 1 with PTQ
	hidden := make([]float64, n.Hidden)
	for i := 0; i < n.Hidden; i++ {
		sum := n.B1[i]
		for j := 0; j < len(input); j++ {
			w := n.W1[i][j]
			if config.Layer1Levels > 0 {
				w = quantize(w, w1Min, w1Max, config.Layer1Levels)
			}
			sum += input[j] * w
		}
		if sum > 0 {
			hidden[i] = sum
		}
	}

	// Layer 2 with PTQ
	output := make([]float64, 10)
	for i := 0; i < 10; i++ {
		sum := n.B2[i]
		for j := 0; j < n.Hidden; j++ {
			w := n.W2[i][j]
			if config.Layer2Levels > 0 {
				w = quantize(w, w2Min, w2Max, config.Layer2Levels)
			}
			sum += hidden[j] * w
		}
		output[i] = sum
	}

	return softmax(output)
}

// Forward performs FP32 forward pass (no quantization)
func (n *Network) Forward(input []float64) []float64 {
	return n.ForwardPTQ(input, PTQConfig{Layer1Levels: 0, Layer2Levels: 0})
}

func softmax(x []float64) []float64 {
	max := x[0]
	for _, v := range x {
		if v > max {
			max = v
		}
	}
	sum := 0.0
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = math.Exp(v - max)
		sum += result[i]
	}
	for i := range result {
		result[i] /= sum
	}
	return result
}

func argmax(x []float64) int {
	maxIdx := 0
	for i, v := range x {
		if v > x[maxIdx] {
			maxIdx = i
		}
	}
	return maxIdx
}

func clipGrad(g, maxNorm float64) float64 {
	if g > maxNorm {
		return maxNorm
	}
	if g < -maxNorm {
		return -maxNorm
	}
	return g
}

// Train trains the network with FP32 weights
func (n *Network) Train(trainImages, testImages [][]float64, trainLabels, testLabels []int,
	epochs int, lr float64) {

	fmt.Println("Training with FP32 weights...")

	const batchSize = 32
	const gradClip = 1.0

	for epoch := 0; epoch < epochs; epoch++ {
		indices := rand.Perm(len(trainImages))
		totalLoss := 0.0

		for batchStart := 0; batchStart < len(indices); batchStart += batchSize {
			batchEnd := batchStart + batchSize
			if batchEnd > len(indices) {
				batchEnd = len(indices)
			}
			batch := indices[batchStart:batchEnd]

			gradW2 := make([][]float64, 10)
			gradB2 := make([]float64, 10)
			for i := 0; i < 10; i++ {
				gradW2[i] = make([]float64, n.Hidden)
			}

			gradW1 := make([][]float64, n.Hidden)
			gradB1 := make([]float64, n.Hidden)
			for i := 0; i < n.Hidden; i++ {
				gradW1[i] = make([]float64, 784)
			}

			for _, idx := range batch {
				input := trainImages[idx]
				target := trainLabels[idx]

				probs := n.Forward(input)
				prob := probs[target]
				if prob < 1e-10 {
					prob = 1e-10
				}
				loss := -math.Log(prob)
				if !math.IsNaN(loss) && !math.IsInf(loss, 0) {
					totalLoss += loss
				}

				grad2 := make([]float64, 10)
				for i := range grad2 {
					grad2[i] = probs[i]
					if i == target {
						grad2[i] -= 1
					}
				}

				hidden := make([]float64, n.Hidden)
				for i := 0; i < n.Hidden; i++ {
					sum := n.B1[i]
					for j := 0; j < 784; j++ {
						sum += input[j] * n.W1[i][j]
					}
					if sum > 0 {
						hidden[i] = sum
					}
				}

				for i := 0; i < 10; i++ {
					for j := 0; j < n.Hidden; j++ {
						gradW2[i][j] += grad2[i] * hidden[j]
					}
					gradB2[i] += grad2[i]
				}

				grad1 := make([]float64, n.Hidden)
				for j := 0; j < n.Hidden; j++ {
					for i := 0; i < 10; i++ {
						grad1[j] += grad2[i] * n.W2[i][j]
					}
					if hidden[j] <= 0 {
						grad1[j] = 0
					}
				}

				for i := 0; i < n.Hidden; i++ {
					for j := 0; j < 784; j++ {
						gradW1[i][j] += grad1[i] * input[j]
					}
					gradB1[i] += grad1[i]
				}
			}

			batchLen := float64(len(batch))
			for i := 0; i < 10; i++ {
				for j := 0; j < n.Hidden; j++ {
					g := clipGrad(gradW2[i][j]/batchLen, gradClip)
					n.W2[i][j] -= lr * g
				}
				g := clipGrad(gradB2[i]/batchLen, gradClip)
				n.B2[i] -= lr * g
			}

			for i := 0; i < n.Hidden; i++ {
				for j := 0; j < 784; j++ {
					g := clipGrad(gradW1[i][j]/batchLen, gradClip)
					n.W1[i][j] -= lr * g
				}
				g := clipGrad(gradB1[i]/batchLen, gradClip)
				n.B1[i] -= lr * g
			}
		}

		fpAcc := n.Evaluate(testImages, testLabels, PTQConfig{0, 0})
		cim30Acc := n.Evaluate(testImages, testLabels, PTQConfig{30, 30})

		fmt.Printf("Epoch %2d: loss=%.4f, FP32=%.2f%%, PTQ(30,30)=%.2f%%\n",
			epoch+1, totalLoss/float64(len(trainImages)), fpAcc*100, cim30Acc*100)

		if epoch > 0 && epoch%5 == 0 {
			lr *= 0.9
		}
	}
}

func (n *Network) Evaluate(images [][]float64, labels []int, config PTQConfig) float64 {
	correct := 0
	for i, img := range images {
		probs := n.ForwardPTQ(img, config)
		if argmax(probs) == labels[i] {
			correct++
		}
	}
	return float64(correct) / float64(len(images))
}

// PTQWeightsFile extends the weight file format with per-layer quantization
type PTQWeightsFile struct {
	Layer1Weights     [][]float64 `json:"layer1_weights"`
	Layer2Weights     [][]float64 `json:"layer2_weights"`
	Biases1           []float64   `json:"biases1"`
	Biases2           []float64   `json:"biases2"`
	L1Scale           float64     `json:"l1_scale"`
	L1Offset          float64     `json:"l1_offset"`
	L2Scale           float64     `json:"l2_scale"`
	L2Offset          float64     `json:"l2_offset"`
	Layer1QuantLevels int         `json:"layer1_quant_levels"`
	Layer2QuantLevels int         `json:"layer2_quant_levels"`
	QuantLevels       int         `json:"quant_levels"` // Legacy: uniform levels
}

func (n *Network) SavePTQ(filename string, config PTQConfig) error {
	w1Min, w1Max := getWeightRange(n.W1)
	w2Min, w2Max := getWeightRange(n.W2)

	// Quantize layer 1 weights
	qW1 := make([][]float64, len(n.W1))
	for i := range n.W1 {
		qW1[i] = make([]float64, len(n.W1[i]))
		for j := range n.W1[i] {
			norm := (n.W1[i][j] - w1Min) / (w1Max - w1Min)
			if config.Layer1Levels > 1 {
				qW1[i][j] = math.Round(norm*float64(config.Layer1Levels-1)) / float64(config.Layer1Levels-1)
			} else {
				qW1[i][j] = norm
			}
		}
	}

	// Quantize layer 2 weights
	qW2 := make([][]float64, len(n.W2))
	for i := range n.W2 {
		qW2[i] = make([]float64, len(n.W2[i]))
		for j := range n.W2[i] {
			norm := (n.W2[i][j] - w2Min) / (w2Max - w2Min)
			if config.Layer2Levels > 1 {
				qW2[i][j] = math.Round(norm*float64(config.Layer2Levels-1)) / float64(config.Layer2Levels-1)
			} else {
				qW2[i][j] = norm
			}
		}
	}

	data := PTQWeightsFile{
		Layer1Weights:     qW1,
		Layer2Weights:     qW2,
		Biases1:           n.B1,
		Biases2:           n.B2,
		L1Scale:           w1Max - w1Min,
		L1Offset:          w1Min,
		L2Scale:           w2Max - w2Min,
		L2Offset:          w2Min,
		Layer1QuantLevels: config.Layer1Levels,
		Layer2QuantLevels: config.Layer2Levels,
		QuantLevels:       config.Layer1Levels, // Legacy compatibility
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

func Run(args []string) error {
	fmt.Println("=== MNIST PTQ Training (Per-Layer Post-Training Quantization) ===")
	fmt.Println("Training FP32 weights, then evaluating with different quantization per layer")
	fmt.Println("")

	dataDir := utils.FindModuleDataDir("module3-mnist", "train-images-idx3-ubyte.gz")
	if dataDir == "" {
		fmt.Println("Error: Could not find MNIST data directory")
		os.Exit(1)
	}

	fmt.Println("Loading MNIST data...")
	trainImages, trainLabels, err := mnist.LoadMNIST(dataDir, true)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	testImages, testLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d train, %d test images\n\n", len(trainImages), len(testImages))

	// Train with FP32
	fmt.Println("=== Training 2-Layer Network (784 -> 128 -> 10) ===")
	net := NewNetwork(128)
	net.Train(trainImages, testImages, trainLabels, testLabels, 20, 0.001)

	// PTQ Evaluation with different per-layer configurations
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("=== PTQ EVALUATION (Per-Layer Post-Training Quantization) ===")
	fmt.Println(strings.Repeat("=", 70))

	// Test uniform quantization first
	fmt.Println("\n--- Uniform Quantization (same levels for both layers) ---")
	fmt.Printf("%-12s %10s %10s\n", "Levels", "Accuracy", "Bits/Cell")
	fmt.Println(strings.Repeat("-", 35))

	fpAcc := net.Evaluate(testImages, testLabels, PTQConfig{0, 0})
	fmt.Printf("%-12s %9.2f%% %10s\n", "FP32", fpAcc*100, "32.00")

	uniformLevels := []int{30, 20, 16, 10, 8, 4, 2}
	for _, levels := range uniformLevels {
		acc := net.Evaluate(testImages, testLabels, PTQConfig{levels, levels})
		bits := math.Log2(float64(levels))
		fmt.Printf("%-12d %9.2f%% %10.2f\n", levels, acc*100, bits)
	}

	// Test per-layer quantization
	fmt.Println("\n--- Per-Layer Quantization (different levels per layer) ---")
	fmt.Printf("%-8s %-8s %10s %12s\n", "Layer1", "Layer2", "Accuracy", "Avg Bits")
	fmt.Println(strings.Repeat("-", 45))

	// Test various combinations
	perLayerConfigs := []PTQConfig{
		{30, 30}, // Baseline: FeCIM 30 levels
		{30, 20}, // More precision on hidden layer
		{20, 30}, // More precision on output layer
		{30, 10}, // Reduced output precision
		{20, 20}, // Reduced both
		{30, 8},  // Very reduced output
		{16, 30}, // Reduced hidden, full output
		{16, 16}, // Both reduced
		{10, 30}, // Low hidden, high output
		{30, 4},  // Extreme: 30 hidden, 4 output
		{8, 30},  // Extreme: 8 hidden, 30 output
		{8, 8},   // Both low
	}

	var bestConfig PTQConfig
	bestAcc := 0.0

	for _, config := range perLayerConfigs {
		acc := net.Evaluate(testImages, testLabels, config)
		avgBits := (math.Log2(float64(config.Layer1Levels)) + math.Log2(float64(config.Layer2Levels))) / 2
		fmt.Printf("%-8d %-8d %9.2f%% %11.2f\n", config.Layer1Levels, config.Layer2Levels, acc*100, avgBits)

		if acc > bestAcc {
			bestAcc = acc
			bestConfig = config
		}
	}

	fmt.Println("\n--- Analysis Summary ---")
	fmt.Printf("Best PTQ config: Layer1=%d, Layer2=%d (%.2f%% accuracy)\n",
		bestConfig.Layer1Levels, bestConfig.Layer2Levels, bestAcc*100)

	// Compare hidden vs output layer sensitivity
	fmt.Println("\n--- Layer Sensitivity Analysis ---")
	fmt.Println("Testing which layer is more sensitive to quantization...")

	// Fix output at 30, vary hidden
	fmt.Println("\nHidden layer sensitivity (Output fixed at 30 levels):")
	for _, l1 := range []int{30, 20, 16, 10, 8, 4} {
		acc := net.Evaluate(testImages, testLabels, PTQConfig{l1, 30})
		fmt.Printf("  Layer1=%2d: %.2f%%\n", l1, acc*100)
	}

	// Fix hidden at 30, vary output
	fmt.Println("\nOutput layer sensitivity (Hidden fixed at 30 levels):")
	for _, l2 := range []int{30, 20, 16, 10, 8, 4} {
		acc := net.Evaluate(testImages, testLabels, PTQConfig{30, l2})
		fmt.Printf("  Layer2=%2d: %.2f%%\n", l2, acc*100)
	}

	// Save weights with PTQ configuration
	weightsPath := filepath.Join(dataDir, "pretrained_weights_ptq.json")
	fmt.Printf("\nSaving weights with PTQ config (30,30) to %s...\n", weightsPath)
	if err := net.SavePTQ(weightsPath, PTQConfig{30, 30}); err != nil {
		fmt.Printf("Error saving: %v\n", err)
		os.Exit(1)
	}

	// Also save optimized configuration
	if bestConfig.Layer1Levels != 30 || bestConfig.Layer2Levels != 30 {
		optPath := filepath.Join(dataDir, fmt.Sprintf("pretrained_weights_ptq_%d_%d.json",
			bestConfig.Layer1Levels, bestConfig.Layer2Levels))
		fmt.Printf("Saving optimized PTQ weights to %s...\n", optPath)
		if err := net.SavePTQ(optPath, bestConfig); err != nil {
			fmt.Printf("Error saving: %v\n", err)
		}
	}

	fmt.Println("\nDone!")
	return nil
}

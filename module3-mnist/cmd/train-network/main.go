// Command train-network trains both 2-layer and single-layer MNIST networks
// with quantization-aware training for proper CIM accuracy.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/shared/utils"
)

// Simple 2-layer network with quantization-aware training
type Network struct {
	// Weights stored as FP for training, quantized for inference
	W1     [][]float64 // [hidden][784]
	W2     [][]float64 // [10][hidden]
	B1     []float64   // [hidden]
	B2     []float64   // [10]
	Hidden int
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

// Quantize to N levels in range [min, max]
func quantize(val, min, max float64, levels int) float64 {
	// Normalize to [0, 1]
	norm := (val - min) / (max - min)
	norm = math.Max(0, math.Min(1, norm))

	// Quantize
	step := 1.0 / float64(levels-1)
	quantized := math.Round(norm/step) * step

	// Map back
	return quantized*(max-min) + min
}

// Forward pass with optional quantization
func (n *Network) Forward(input []float64, quantLevels int) []float64 {
	// Find weight ranges for quantization
	w1Min, w1Max := n.W1[0][0], n.W1[0][0]
	for i := range n.W1 {
		for j := range n.W1[i] {
			if n.W1[i][j] < w1Min {
				w1Min = n.W1[i][j]
			}
			if n.W1[i][j] > w1Max {
				w1Max = n.W1[i][j]
			}
		}
	}

	w2Min, w2Max := n.W2[0][0], n.W2[0][0]
	for i := range n.W2 {
		for j := range n.W2[i] {
			if n.W2[i][j] < w2Min {
				w2Min = n.W2[i][j]
			}
			if n.W2[i][j] > w2Max {
				w2Max = n.W2[i][j]
			}
		}
	}

	// Layer 1
	hidden := make([]float64, n.Hidden)
	for i := 0; i < n.Hidden; i++ {
		sum := n.B1[i]
		for j := 0; j < len(input); j++ {
			w := n.W1[i][j]
			if quantLevels > 0 {
				w = quantize(w, w1Min, w1Max, quantLevels)
			}
			sum += input[j] * w
		}
		// ReLU
		if sum > 0 {
			hidden[i] = sum
		}
	}

	// Layer 2
	output := make([]float64, 10)
	for i := 0; i < 10; i++ {
		sum := n.B2[i]
		for j := 0; j < n.Hidden; j++ {
			w := n.W2[i][j]
			if quantLevels > 0 {
				w = quantize(w, w2Min, w2Max, quantLevels)
			}
			sum += hidden[j] * w
		}
		output[i] = sum
	}

	return softmax(output)
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

// Clip gradient to prevent explosion
func clipGrad(g, maxNorm float64) float64 {
	if g > maxNorm {
		return maxNorm
	}
	if g < -maxNorm {
		return -maxNorm
	}
	return g
}

// Train with quantization-aware forward pass
func (n *Network) Train(trainImages, testImages [][]float64, trainLabels, testLabels []int,
	epochs int, lr float64, quantLevels int) {

	fmt.Printf("Training with %d quantization levels...\n", quantLevels)

	const batchSize = 32
	const gradClip = 1.0

	for epoch := 0; epoch < epochs; epoch++ {
		// Shuffle
		indices := rand.Perm(len(trainImages))

		totalLoss := 0.0

		// Mini-batch training
		for batchStart := 0; batchStart < len(indices); batchStart += batchSize {
			batchEnd := batchStart + batchSize
			if batchEnd > len(indices) {
				batchEnd = len(indices)
			}
			batch := indices[batchStart:batchEnd]

			// Accumulate gradients over batch
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

				// Forward with quantization
				probs := n.Forward(input, quantLevels)
				prob := probs[target]
				if prob < 1e-10 {
					prob = 1e-10
				}
				loss := -math.Log(prob)
				if !math.IsNaN(loss) && !math.IsInf(loss, 0) {
					totalLoss += loss
				}

				// Output gradient
				grad2 := make([]float64, 10)
				for i := range grad2 {
					grad2[i] = probs[i]
					if i == target {
						grad2[i] -= 1
					}
				}

				// Compute hidden activations (for gradient)
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

				// Accumulate W2 gradients
				for i := 0; i < 10; i++ {
					for j := 0; j < n.Hidden; j++ {
						gradW2[i][j] += grad2[i] * hidden[j]
					}
					gradB2[i] += grad2[i]
				}

				// Backprop to hidden
				grad1 := make([]float64, n.Hidden)
				for j := 0; j < n.Hidden; j++ {
					for i := 0; i < 10; i++ {
						grad1[j] += grad2[i] * n.W2[i][j]
					}
					if hidden[j] <= 0 {
						grad1[j] = 0 // ReLU derivative
					}
				}

				// Accumulate W1 gradients
				for i := 0; i < n.Hidden; i++ {
					for j := 0; j < 784; j++ {
						gradW1[i][j] += grad1[i] * input[j]
					}
					gradB1[i] += grad1[i]
				}
			}

			// Apply gradients with clipping
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

		// Evaluate
		fpAcc := n.Evaluate(testImages, testLabels, 0)
		cimAcc := n.Evaluate(testImages, testLabels, quantLevels)

		fmt.Printf("Epoch %2d: loss=%.4f, FP=%.2f%%, CIM(%d)=%.2f%%\n",
			epoch+1, totalLoss/float64(len(trainImages)), fpAcc*100, quantLevels, cimAcc*100)

		// LR decay
		if epoch > 0 && epoch%5 == 0 {
			lr *= 0.9
		}
	}
}

func (n *Network) Evaluate(images [][]float64, labels []int, quantLevels int) float64 {
	correct := 0
	for i, img := range images {
		probs := n.Forward(img, quantLevels)
		if argmax(probs) == labels[i] {
			correct++
		}
	}
	return float64(correct) / float64(len(images))
}

func (n *Network) Save(filename string, levels int) error {
	// Find weight ranges for proper scaling
	w1Min, w1Max := n.W1[0][0], n.W1[0][0]
	for i := range n.W1 {
		for j := range n.W1[i] {
			if n.W1[i][j] < w1Min {
				w1Min = n.W1[i][j]
			}
			if n.W1[i][j] > w1Max {
				w1Max = n.W1[i][j]
			}
		}
	}

	w2Min, w2Max := n.W2[0][0], n.W2[0][0]
	for i := range n.W2 {
		for j := range n.W2[i] {
			if n.W2[i][j] < w2Min {
				w2Min = n.W2[i][j]
			}
			if n.W2[i][j] > w2Max {
				w2Max = n.W2[i][j]
			}
		}
	}

	// Quantize weights to specified levels for storage
	qW1 := make([][]float64, len(n.W1))
	for i := range n.W1 {
		qW1[i] = make([]float64, len(n.W1[i]))
		for j := range n.W1[i] {
			// Normalize to [0, 1]
			norm := (n.W1[i][j] - w1Min) / (w1Max - w1Min)
			// Quantize to N levels
			qW1[i][j] = math.Round(norm*float64(levels-1)) / float64(levels-1)
		}
	}

	qW2 := make([][]float64, len(n.W2))
	for i := range n.W2 {
		qW2[i] = make([]float64, len(n.W2[i]))
		for j := range n.W2[i] {
			norm := (n.W2[i][j] - w2Min) / (w2Max - w2Min)
			qW2[i][j] = math.Round(norm*float64(levels-1)) / float64(levels-1)
		}
	}

	data := map[string]interface{}{
		"layer1_weights": qW1,
		"layer2_weights": qW2,
		"biases1":        n.B1,
		"biases2":        n.B2,
		"l1_scale":       w1Max - w1Min,
		"l1_offset":      w1Min,
		"l2_scale":       w2Max - w2Min,
		"l2_offset":      w2Min,
		"quant_levels":   levels,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

func main() {
	fmt.Println("=== MNIST Network Training (PTQ - Post-Training Quantization) ===")
	fmt.Println("Training with FP32, quantization applied at inference time")
	fmt.Println("")

	dataDir := utils.FindDirectoryWithMarker("module3-mnist/data", "train-images-idx3-ubyte.gz")
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

	// Train with FP32 (no quantization during training)
	fmt.Println("=== Training 2-Layer Network (784 -> 128 -> 10) ===")
	net := NewNetwork(128)
	net.Train(trainImages, testImages, trainLabels, testLabels, 20, 0.001, 0) // 0 = FP training

	// Final evaluation at different quantization levels (PTQ)
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== PTQ EVALUATION (Post-Training Quantization) ===")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("%-10s %10s %10s\n", "Levels", "Accuracy", "Bits/Cell")
	fmt.Println(strings.Repeat("-", 35))

	fpAcc := net.Evaluate(testImages, testLabels, 0)
	fmt.Printf("%-10s %9.2f%% %10s\n", "FP32", fpAcc*100, "32.00")

	for _, levels := range []int{31, 30, 29, 20, 10, 8, 4, 2} {
		cimAcc := net.Evaluate(testImages, testLabels, levels)
		bits := math.Log2(float64(levels))
		fmt.Printf("%-10d %9.2f%% %10.2f\n", levels, cimAcc*100, bits)
	}

	// Save FP weights (quantization happens at load time based on slider)
	weightsPath := filepath.Join(dataDir, "pretrained_weights.json")
	fmt.Printf("\nSaving to %s...\n", weightsPath)
	if err := net.Save(weightsPath, 30); err != nil {
		fmt.Printf("Error saving: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nDone! Weights saved. Quantization applied dynamically based on UI slider.")
}


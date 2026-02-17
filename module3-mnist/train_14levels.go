//go:build ignore
// +build ignore

// Training script for 14-level quantization (quatorze)
// Evaluates at epochs 10 and 20
// Run with: go run train_14levels.go
package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/shared/canvas"
)

const (
	QuantLevels = 14 // quatorze levels
	Hidden      = 128
)

type Network struct {
	W1 [][]float64 // [hidden][784]
	W2 [][]float64 // [10][hidden]
	B1 []float64
	B2 []float64
}

func NewNetwork() *Network {
	n := &Network{
		W1: make([][]float64, Hidden),
		W2: make([][]float64, 10),
		B1: make([]float64, Hidden),
		B2: make([]float64, 10),
	}

	// Xavier initialization
	scale1 := math.Sqrt(2.0 / float64(784+Hidden))
	for i := 0; i < Hidden; i++ {
		n.W1[i] = make([]float64, 784)
		for j := 0; j < 784; j++ {
			n.W1[i][j] = rand.NormFloat64() * scale1
		}
	}

	scale2 := math.Sqrt(2.0 / float64(Hidden+10))
	for i := 0; i < 10; i++ {
		n.W2[i] = make([]float64, Hidden)
		for j := 0; j < Hidden; j++ {
			n.W2[i][j] = rand.NormFloat64() * scale2
		}
	}

	return n
}

func quantize(val, min, max float64, levels int) float64 {
	if levels <= 1 {
		return (min + max) / 2
	}
	norm := (val - min) / (max - min)
	norm = math.Max(0, math.Min(1, norm))
	step := 1.0 / float64(levels-1)
	quantized := math.Round(norm/step) * step
	return quantized*(max-min) + min
}

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

func (n *Network) ForwardQuantized(input []float64, levels int) []float64 {
	w1Min, w1Max := getWeightRange(n.W1)
	w2Min, w2Max := getWeightRange(n.W2)

	// Layer 1 with quantization
	hidden := make([]float64, Hidden)
	for i := 0; i < Hidden; i++ {
		sum := n.B1[i]
		for j := 0; j < len(input); j++ {
			w := n.W1[i][j]
			if levels > 0 {
				w = quantize(w, w1Min, w1Max, levels)
			}
			sum += input[j] * w
		}
		if sum > 0 {
			hidden[i] = sum
		}
	}

	// Layer 2 with quantization
	output := make([]float64, 10)
	for i := 0; i < 10; i++ {
		sum := n.B2[i]
		for j := 0; j < Hidden; j++ {
			w := n.W2[i][j]
			if levels > 0 {
				w = quantize(w, w2Min, w2Max, levels)
			}
			sum += hidden[j] * w
		}
		output[i] = sum
	}

	return softmax(output)
}

func (n *Network) Forward(input []float64) []float64 {
	return n.ForwardQuantized(input, 0)
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

func (n *Network) Evaluate(images [][]float64, labels []int, levels int) float64 {
	correct := 0
	for i, img := range images {
		probs := n.ForwardQuantized(img, levels)
		if argmax(probs) == labels[i] {
			correct++
		}
	}
	return float64(correct) / float64(len(images))
}

func (n *Network) Train(trainImages, testImages [][]float64, trainLabels, testLabels []int,
	maxEpochs int, lr float64) {

	const batchSize = 32
	const gradClip = 1.0

	fmt.Println("Training with FP32 weights, evaluating with 14-level PTQ...")
	fmt.Printf("%-8s %12s %12s %12s\n", "Epoch", "Loss", "FP32 Acc", "14-Level Acc")
	fmt.Println("-----------------------------------------------")

	for epoch := 1; epoch <= maxEpochs; epoch++ {
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
				gradW2[i] = make([]float64, Hidden)
			}

			gradW1 := make([][]float64, Hidden)
			gradB1 := make([]float64, Hidden)
			for i := 0; i < Hidden; i++ {
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

				hidden := make([]float64, Hidden)
				for i := 0; i < Hidden; i++ {
					sum := n.B1[i]
					for j := 0; j < 784; j++ {
						sum += input[j] * n.W1[i][j]
					}
					if sum > 0 {
						hidden[i] = sum
					}
				}

				for i := 0; i < 10; i++ {
					for j := 0; j < Hidden; j++ {
						gradW2[i][j] += grad2[i] * hidden[j]
					}
					gradB2[i] += grad2[i]
				}

				grad1 := make([]float64, Hidden)
				for j := 0; j < Hidden; j++ {
					for i := 0; i < 10; i++ {
						grad1[j] += grad2[i] * n.W2[i][j]
					}
					if hidden[j] <= 0 {
						grad1[j] = 0
					}
				}

				for i := 0; i < Hidden; i++ {
					for j := 0; j < 784; j++ {
						gradW1[i][j] += grad1[i] * input[j]
					}
					gradB1[i] += grad1[i]
				}
			}

			batchLen := float64(len(batch))
			for i := 0; i < 10; i++ {
				for j := 0; j < Hidden; j++ {
					g := clipGrad(gradW2[i][j]/batchLen, gradClip)
					n.W2[i][j] -= lr * g
				}
				g := clipGrad(gradB2[i]/batchLen, gradClip)
				n.B2[i] -= lr * g
			}

			for i := 0; i < Hidden; i++ {
				for j := 0; j < 784; j++ {
					g := clipGrad(gradW1[i][j]/batchLen, gradClip)
					n.W1[i][j] -= lr * g
				}
				g := clipGrad(gradB1[i]/batchLen, gradClip)
				n.B1[i] -= lr * g
			}
		}

		// Evaluate at every epoch, but highlight epochs 10 and 20
		fpAcc := n.Evaluate(testImages, testLabels, 0)
		q14Acc := n.Evaluate(testImages, testLabels, QuantLevels)

		marker := ""
		if epoch == 10 || epoch == 20 {
			marker = " <-- TARGET"
		}

		fmt.Printf("%-8d %12.4f %11.2f%% %11.2f%%%s\n",
			epoch, totalLoss/float64(len(trainImages)), fpAcc*100, q14Acc*100, marker)

		if epoch > 0 && epoch%5 == 0 {
			lr *= 0.9
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("=== MNIST Training with 14 Levels (Quatorze) ===")
	fmt.Printf("Quantization: %d discrete levels (%.2f bits/cell)\n", QuantLevels, math.Log2(float64(QuantLevels)))
	fmt.Println("Target epochs: 10 and 20")
	fmt.Println("")

	dataDir := utils.FindModuleDataDir("module3-mnist", "train-images-idx3-ubyte.gz")
	if dataDir == "" {
		fmt.Println("\n╔════════════════════════════════════════════════════════════════╗")
		fmt.Println("║ ERROR: Could not find MNIST data directory                      ║")
		fmt.Println("╠════════════════════════════════════════════════════════════════╣")
		fmt.Println("║ The training script could not locate the MNIST dataset.         ║")
		fmt.Println("╠════════════════════════════════════════════════════════════════╣")
		fmt.Println("║ To fix this:                                                    ║")
		fmt.Println("║ 1. Download MNIST from: http://yann.lecun.com/exdb/mnist/       ║")
		fmt.Println("║ 2. Place files in: module3-mnist/data/                          ║")
		fmt.Println("║    Required files:                                              ║")
		fmt.Println("║    - train-images-idx3-ubyte.gz                                 ║")
		fmt.Println("║    - train-labels-idx1-ubyte.gz                                 ║")
		fmt.Println("║    - t10k-images-idx3-ubyte.gz                                  ║")
		fmt.Println("║    - t10k-labels-idx1-ubyte.gz                                  ║")
		fmt.Println("║ 3. Run from repository root: cd fecim-lattice-tools             ║")
		fmt.Println("╚════════════════════════════════════════════════════════════════╝")
		os.Exit(1)
	}

	fmt.Println("Loading MNIST data...")
	trainImages, trainLabels, err := mnist.LoadMNIST(dataDir, true)
	if err != nil {
		fmt.Printf("\nError loading training data: %v\n", err)
		fmt.Println("Check that train-images-idx3-ubyte.gz and train-labels-idx1-ubyte.gz exist and are valid.")
		os.Exit(1)
	}

	testImages, testLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		fmt.Printf("\nError loading test data: %v\n", err)
		fmt.Println("Check that t10k-images-idx3-ubyte.gz and t10k-labels-idx1-ubyte.gz exist and are valid.")
		os.Exit(1)
	}

	fmt.Printf("Loaded %d train, %d test images\n\n", len(trainImages), len(testImages))

	net := NewNetwork()
	net.Train(trainImages, testImages, trainLabels, testLabels, 20, 0.001)

	fmt.Println("\n=== Final Results ===")
	fmt.Printf("14-level accuracy at epoch 10: (see above)\n")
	fmt.Printf("14-level accuracy at epoch 20: (see above)\n")
	fmt.Printf("Bits per cell: %.2f\n", math.Log2(float64(QuantLevels)))
}

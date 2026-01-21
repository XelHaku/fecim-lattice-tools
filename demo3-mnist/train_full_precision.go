//go:build ignore
// +build ignore

// Full precision training then quantize to 30 levels.
// Run with: go run train_full_precision.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"multilayer-ferroelectric-cim-visualizer/demo3-mnist/pkg/mnist"
)

const (
	inputSize  = 784
	hiddenSize = 128
	outputSize = 10
	feCIMLevels = 30
)

// SimpleNetwork is a 2-layer network trained with full precision
type SimpleNetwork struct {
	w1      [][]float64 // hiddenSize x inputSize
	w2      [][]float64 // outputSize x hiddenSize
	b1      []float64   // hiddenSize
	b2      []float64   // outputSize
}

func newSimpleNetwork() *SimpleNetwork {
	net := &SimpleNetwork{
		w1: make([][]float64, hiddenSize),
		w2: make([][]float64, outputSize),
		b1: make([]float64, hiddenSize),
		b2: make([]float64, outputSize),
	}

	// Xavier initialization
	scale1 := math.Sqrt(2.0 / float64(inputSize+hiddenSize))
	for i := range net.w1 {
		net.w1[i] = make([]float64, inputSize)
		for j := range net.w1[i] {
			net.w1[i][j] = rand.NormFloat64() * scale1
		}
	}

	scale2 := math.Sqrt(2.0 / float64(hiddenSize+outputSize))
	for i := range net.w2 {
		net.w2[i] = make([]float64, hiddenSize)
		for j := range net.w2[i] {
			net.w2[i][j] = rand.NormFloat64() * scale2
		}
	}

	return net
}

func relu(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0
}

func reluDeriv(x float64) float64 {
	if x > 0 {
		return 1
	}
	return 0
}

func softmax(x []float64) []float64 {
	result := make([]float64, len(x))
	max := x[0]
	for _, v := range x {
		if v > max {
			max = v
		}
	}
	var sum float64
	for i, v := range x {
		result[i] = math.Exp(v - max)
		sum += result[i]
	}
	for i := range result {
		result[i] /= sum
	}
	return result
}

func (n *SimpleNetwork) forward(input []float64) ([]float64, []float64, []float64) {
	// Hidden layer
	hidden := make([]float64, hiddenSize)
	for i := 0; i < hiddenSize; i++ {
		sum := n.b1[i]
		for j := 0; j < len(input); j++ {
			sum += n.w1[i][j] * input[j]
		}
		hidden[i] = relu(sum)
	}

	// Output layer
	output := make([]float64, outputSize)
	for i := 0; i < outputSize; i++ {
		sum := n.b2[i]
		for j := 0; j < hiddenSize; j++ {
			sum += n.w2[i][j] * hidden[j]
		}
		output[i] = sum
	}

	probs := softmax(output)
	return hidden, output, probs
}

func (n *SimpleNetwork) predict(input []float64) int {
	_, _, probs := n.forward(input)
	maxIdx := 0
	maxVal := probs[0]
	for i, v := range probs {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}

func (n *SimpleNetwork) trainBatch(images [][]float64, labels []int, lr float64) float64 {
	totalLoss := 0.0

	for idx := range images {
		input := images[idx]
		target := labels[idx]

		// Forward
		hidden, _, probs := n.forward(input)

		// Loss
		loss := -math.Log(probs[target] + 1e-10)
		totalLoss += loss

		// Output gradients
		outputGrad := make([]float64, outputSize)
		for i := range outputGrad {
			outputGrad[i] = probs[i]
			if i == target {
				outputGrad[i] -= 1.0
			}
		}

		// Hidden gradients
		hiddenGrad := make([]float64, hiddenSize)
		for j := 0; j < hiddenSize; j++ {
			for i := 0; i < outputSize; i++ {
				hiddenGrad[j] += outputGrad[i] * n.w2[i][j]
			}
			// Pre-ReLU value for derivative
			preRelu := n.b1[j]
			for k := 0; k < len(input); k++ {
				preRelu += n.w1[j][k] * input[k]
			}
			hiddenGrad[j] *= reluDeriv(preRelu)
		}

		// Update layer 2
		for i := 0; i < outputSize; i++ {
			for j := 0; j < hiddenSize; j++ {
				n.w2[i][j] -= lr * outputGrad[i] * hidden[j]
			}
			n.b2[i] -= lr * outputGrad[i]
		}

		// Update layer 1
		for i := 0; i < hiddenSize; i++ {
			for j := 0; j < len(input); j++ {
				n.w1[i][j] -= lr * hiddenGrad[i] * input[j]
			}
			n.b1[i] -= lr * hiddenGrad[i]
		}
	}

	return totalLoss / float64(len(images))
}

func (n *SimpleNetwork) evaluate(images [][]float64, labels []int) float64 {
	correct := 0
	for i, img := range images {
		if n.predict(img) == labels[i] {
			correct++
		}
	}
	return float64(correct) / float64(len(images))
}

// Quantize weights to 30 levels and save
func (n *SimpleNetwork) saveQuantized(filename string) error {
	// Find weight ranges
	min1, max1 := n.w1[0][0], n.w1[0][0]
	for _, row := range n.w1 {
		for _, w := range row {
			if w < min1 {
				min1 = w
			}
			if w > max1 {
				max1 = w
			}
		}
	}

	min2, max2 := n.w2[0][0], n.w2[0][0]
	for _, row := range n.w2 {
		for _, w := range row {
			if w < min2 {
				min2 = w
			}
			if w > max2 {
				max2 = w
			}
		}
	}

	fmt.Printf("Weight ranges: L1 [%.4f, %.4f], L2 [%.4f, %.4f]\n", min1, max1, min2, max2)

	// Normalize and quantize to 30 levels [0, 1]
	qw1 := make([][]float64, hiddenSize)
	for i := range n.w1 {
		qw1[i] = make([]float64, inputSize)
		for j := range n.w1[i] {
			// Normalize to [0, 1]
			normalized := (n.w1[i][j] - min1) / (max1 - min1)
			// Quantize to 30 levels
			level := math.Round(normalized * float64(feCIMLevels-1))
			qw1[i][j] = level / float64(feCIMLevels-1)
		}
	}

	qw2 := make([][]float64, outputSize)
	for i := range n.w2 {
		qw2[i] = make([]float64, hiddenSize)
		for j := range n.w2[i] {
			normalized := (n.w2[i][j] - min2) / (max2 - min2)
			level := math.Round(normalized * float64(feCIMLevels-1))
			qw2[i][j] = level / float64(feCIMLevels-1)
		}
	}

	// Save with scale info for inference
	data := struct {
		Layer1Weights [][]float64 `json:"layer1_weights"`
		Layer2Weights [][]float64 `json:"layer2_weights"`
		Biases1       []float64   `json:"biases1"`
		Biases2       []float64   `json:"biases2"`
		L1Scale       float64     `json:"l1_scale"`
		L1Offset      float64     `json:"l1_offset"`
		L2Scale       float64     `json:"l2_scale"`
		L2Offset      float64     `json:"l2_offset"`
	}{
		Layer1Weights: qw1,
		Layer2Weights: qw2,
		Biases1:       n.b1,
		Biases2:       n.b2,
		L1Scale:       max1 - min1,
		L1Offset:      min1,
		L2Scale:       max2 - min2,
		L2Offset:      min2,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("============================================")
	fmt.Println("Full Precision MNIST Training")
	fmt.Println("Then quantize to 30 FeCIM levels")
	fmt.Println("============================================")
	fmt.Println()

	// Load MNIST
	fmt.Println("Loading MNIST dataset...")
	trainImages, trainLabels, err := mnist.LoadMNIST("demo3-mnist/data", true)
	if err != nil {
		log.Fatalf("Failed to load training data: %v", err)
	}
	fmt.Printf("Loaded %d training images\n", len(trainImages))

	testImages, testLabels, err := mnist.LoadMNIST("demo3-mnist/data", false)
	if err != nil {
		log.Fatalf("Failed to load test data: %v", err)
	}
	fmt.Printf("Loaded %d test images\n", len(testImages))

	// Create network
	net := newSimpleNetwork()

	// Training params
	epochs := 10
	lr := 0.1
	batchSize := 32

	fmt.Printf("\nTraining with:\n")
	fmt.Printf("  - Epochs: %d\n", epochs)
	fmt.Printf("  - Learning rate: %.2f\n", lr)
	fmt.Printf("  - Batch size: %d\n", batchSize)
	fmt.Println()

	for epoch := 1; epoch <= epochs; epoch++ {
		startTime := time.Now()

		// Shuffle
		perm := rand.Perm(len(trainImages))

		// Train in batches
		totalLoss := 0.0
		for i := 0; i < len(trainImages); i += batchSize {
			end := i + batchSize
			if end > len(trainImages) {
				end = len(trainImages)
			}

			batchImages := make([][]float64, end-i)
			batchLabels := make([]int, end-i)
			for j := i; j < end; j++ {
				batchImages[j-i] = trainImages[perm[j]]
				batchLabels[j-i] = trainLabels[perm[j]]
			}

			totalLoss += net.trainBatch(batchImages, batchLabels, lr) * float64(len(batchImages))
		}

		avgLoss := totalLoss / float64(len(trainImages))
		acc := net.evaluate(testImages[:2000], testLabels[:2000])
		elapsed := time.Since(startTime)

		fmt.Printf("Epoch %2d: Loss=%.4f, Accuracy=%.1f%% [%v]\n",
			epoch, avgLoss, acc*100, elapsed.Round(time.Second))

		// LR decay
		if epoch%5 == 0 {
			lr *= 0.5
			fmt.Printf("  -> Learning rate reduced to %.3f\n", lr)
		}
	}

	// Final evaluation
	fmt.Println("\nFinal evaluation on full test set...")
	finalAcc := net.evaluate(testImages, testLabels)
	fmt.Printf("Final Test Accuracy (full precision): %.1f%%\n", finalAcc*100)

	// Quantize and save
	weightsFile := "demo3-mnist/data/pretrained_weights.json"
	fmt.Printf("\nQuantizing to 30 levels and saving to %s...\n", weightsFile)
	if err := net.saveQuantized(weightsFile); err != nil {
		log.Printf("Failed to save: %v", err)
	} else {
		fmt.Println("Weights saved successfully.")
	}

	fmt.Println("\n============================================")
	fmt.Printf("Training complete! Accuracy: %.1f%%\n", finalAcc*100)
	fmt.Println("============================================")
}

//go:build ignore
// +build ignore

// Proper training script for MNISTNetwork.
// This trains using the MNISTNetwork.TrainEpoch method to ensure
// weights are compatible with MNISTNetwork.Forward.
// Run with: go run train_mnist_proper.go
package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/module3-mnist/pkg/training"
	"fecim-lattice-tools/shared/crossbar"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("============================================")
	fmt.Println("FeCIM MNIST Training (MNISTNetwork)")
	fmt.Println("Target: Physics-limited (typically >85%)")
	fmt.Println("30 discrete analog levels (demo baseline; conference claim)")
	fmt.Println("============================================")
	fmt.Println()

	// Load MNIST data
	fmt.Println("Loading MNIST dataset...")
	trainImages, trainLabels, err := mnist.LoadMNIST("module3-mnist/data", true)
	if err != nil {
		fmt.Println("\n╔════════════════════════════════════════════════════════════════╗")
		fmt.Println("║ ERROR: Could not load MNIST training data                       ║")
		fmt.Println("╠════════════════════════════════════════════════════════════════╣")
		fmt.Printf("║ Details: %v\n", err)
		fmt.Println("╠════════════════════════════════════════════════════════════════╣")
		fmt.Println("║ To fix this:                                                    ║")
		fmt.Println("║ 1. Run from the repository root:                                ║")
		fmt.Println("║    cd fecim-lattice-tools                                       ║")
		fmt.Println("║ 2. Download MNIST data:                                         ║")
		fmt.Println("║    wget http://yann.lecun.com/exdb/mnist/train-images-idx3-ubyte.gz ║")
		fmt.Println("║    wget http://yann.lecun.com/exdb/mnist/train-labels-idx1-ubyte.gz ║")
		fmt.Println("║    wget http://yann.lecun.com/exdb/mnist/t10k-images-idx3-ubyte.gz  ║")
		fmt.Println("║    wget http://yann.lecun.com/exdb/mnist/t10k-labels-idx1-ubyte.gz  ║")
		fmt.Println("║ 3. Move files to: module3-mnist/data/                           ║")
		fmt.Println("╚════════════════════════════════════════════════════════════════╝")
		log.Fatalf("Failed to load training data: %v", err)
	}
	fmt.Printf("Loaded %d training images\n", len(trainImages))

	testImages, testLabels, err := mnist.LoadMNIST("module3-mnist/data", false)
	if err != nil {
		fmt.Println("\n╔════════════════════════════════════════════════════════════════╗")
		fmt.Println("║ ERROR: Could not load MNIST test data                           ║")
		fmt.Println("╠════════════════════════════════════════════════════════════════╣")
		fmt.Printf("║ Details: %v\n", err)
		fmt.Println("╠════════════════════════════════════════════════════════════════╣")
		fmt.Println("║ Training data loaded OK. Test data files may be missing.        ║")
		fmt.Println("║ Expected files in module3-mnist/data/:                          ║")
		fmt.Println("║   - t10k-images-idx3-ubyte.gz                                   ║")
		fmt.Println("║   - t10k-labels-idx1-ubyte.gz                                   ║")
		fmt.Println("╚════════════════════════════════════════════════════════════════╝")
		log.Fatalf("Failed to load test data: %v", err)
	}
	fmt.Printf("Loaded %d test images\n", len(testImages))

	// Use subset for training (full 60k is slow but more accurate)
	trainSubset := 30000
	if len(trainImages) > trainSubset {
		trainImages = trainImages[:trainSubset]
		trainLabels = trainLabels[:trainSubset]
	}

	// Create crossbar arrays with no noise for training
	hidden := 128
	layer1, err := crossbar.NewArray(&crossbar.Config{
		Rows: hidden, Cols: 784, NoiseLevel: 0, ADCBits: 16, DACBits: 16, // High resolution for training
	})
	if err != nil {
		fmt.Printf("\nError creating layer 1 crossbar (%dx%d): %v\n", hidden, 784, err)
		fmt.Println("This may indicate insufficient memory. Try reducing hidden layer size.")
		log.Fatalf("Failed to create layer1: %v", err)
	}

	layer2, err := crossbar.NewArray(&crossbar.Config{
		Rows: 10, Cols: hidden, NoiseLevel: 0, ADCBits: 16, DACBits: 16, // High resolution for training
	})
	if err != nil {
		fmt.Printf("\nError creating layer 2 crossbar (%dx%d): %v\n", 10, hidden, err)
		fmt.Println("This may indicate insufficient memory. Try reducing hidden layer size.")
		log.Fatalf("Failed to create layer2: %v", err)
	}

	// Create network using MNISTNetwork
	net := training.NewMNISTNetwork(layer1, layer2)

	// Training parameters
	// Higher learning rate needed to overcome 30-level quantization threshold
	epochs := 20
	learningRate := 1.0

	fmt.Printf("\nTraining with:\n")
	fmt.Printf("  - Epochs: %d\n", epochs)
	fmt.Printf("  - Learning rate: %.2f\n", learningRate)
	fmt.Printf("  - Training samples: %d\n", len(trainImages))
	fmt.Printf("  - Hidden units: %d\n", hidden)
	fmt.Println()

	bestAcc := 0.0
	for epoch := 1; epoch <= epochs; epoch++ {
		startTime := time.Now()

		// Shuffle training data
		perm := rand.Perm(len(trainImages))
		shuffledImages := make([][]float64, len(trainImages))
		shuffledLabels := make([]int, len(trainImages))
		for i, p := range perm {
			shuffledImages[i] = trainImages[p]
			shuffledLabels[i] = trainLabels[p]
		}

		// Train one epoch
		loss := net.TrainEpoch(shuffledImages, shuffledLabels, learningRate)

		// Evaluate on test set (smaller subset for speed)
		testSubset := testImages
		testSubsetLabels := testLabels
		if len(testImages) > 2000 {
			testSubset = testImages[:2000]
			testSubsetLabels = testLabels[:2000]
		}
		acc := net.Evaluate(testSubset, testSubsetLabels)

		if acc > bestAcc {
			bestAcc = acc
		}

		elapsed := time.Since(startTime)
		fmt.Printf("Epoch %2d: Loss=%.4f, Accuracy=%.1f%% (best: %.1f%%) [%v]\n",
			epoch, loss, acc*100, bestAcc*100, elapsed.Round(time.Second))

		// Learning rate decay
		if epoch%5 == 0 {
			learningRate *= 0.5
			fmt.Printf("  -> Learning rate reduced to %.3f\n", learningRate)
		}
	}

	// Final evaluation on full test set
	fmt.Println("\nFinal evaluation on full test set...")
	finalAcc := net.Evaluate(testImages, testLabels)
	fmt.Printf("Final Test Accuracy: %.1f%% (Target: >85%%)\n", finalAcc*100)

	if finalAcc >= 0.85 {
		fmt.Println("✓ FeCIM target ACHIEVED!")
	}

	// Save weights
	weightsFile := "module3-mnist/data/pretrained_weights.json"
	fmt.Printf("\nSaving weights to %s...\n", weightsFile)
	if err := net.SaveWeights(weightsFile); err != nil {
		log.Printf("Warning: Failed to save weights: %v", err)
	} else {
		fmt.Println("Weights saved successfully.")
	}

	fmt.Println("\n============================================")
	fmt.Printf("Training complete! Best accuracy: %.1f%%\n", bestAcc*100)
	fmt.Println("============================================")
}

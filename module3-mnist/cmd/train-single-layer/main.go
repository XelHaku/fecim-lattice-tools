// Command train-single-layer trains a single-layer MNIST network matching Dr. Tour's demo.
// Expected accuracy: ~85-92% (single-layer networks are limited by linear separability of MNIST)
package mnisttrainsingle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/module3-mnist/pkg/training"
	"fecim-lattice-tools/shared/utils"
)

func Run(args []string) error {
	fmt.Println("=== Single-Layer MNIST Training (Tour Mode) ===")
	fmt.Println("Architecture: 784 -> 10 (no hidden layer)")
	fmt.Println("Single-layer limit: ~85-92% (MNIST not fully linearly separable)")
	fmt.Println("")

	// Find data directory
	dataDir := utils.FindDirectoryWithMarker("module3-mnist/data", "train-images-idx3-ubyte.gz")
	if dataDir == "" {
		fmt.Println("Error: Could not find MNIST data directory")
		os.Exit(1)
	}

	// Load MNIST data
	fmt.Println("Loading MNIST data...")
	trainImages, trainLabels, err := mnist.LoadMNIST(dataDir, true)
	if err != nil {
		fmt.Printf("Error loading training data: %v\n", err)
		os.Exit(1)
	}

	testImages, testLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		fmt.Printf("Error loading test data: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d training images, %d test images\n", len(trainImages), len(testImages))

	// Create and train network
	fmt.Println("\nTraining single-layer network...")
	net := training.NewSingleLayerNetwork()

	// Train for 20 epochs with learning rate 0.1
	bestAcc := net.Train(trainImages, trainLabels, testImages, testLabels, 20, 0.1)

	fmt.Printf("\n=== Training Complete ===\n")
	fmt.Printf("Best test accuracy: %.2f%%\n", bestAcc*100)

	// Save weights
	weightsPath := filepath.Join(dataDir, "single_layer_weights.json")
	fmt.Printf("Saving weights to %s...\n", weightsPath)
	if err := net.SaveWeights(weightsPath); err != nil {
		fmt.Printf("Error saving weights: %v\n", err)
		os.Exit(1)
	}

	// Also update the main pretrained_weights.json with single-layer weights
	mainWeightsPath := filepath.Join(dataDir, "pretrained_weights.json")
	if err := appendSingleLayerToMainWeights(mainWeightsPath, net); err != nil {
		fmt.Printf("Warning: Could not append to main weights file: %v\n", err)
	} else {
		fmt.Printf("Updated %s with single-layer weights\n", mainWeightsPath)
	}

	fmt.Println("\nDone!")
	return nil
}

func appendSingleLayerToMainWeights(mainPath string, net *training.SingleLayerNetwork) error {
	// Read existing weights
	existing := make(map[string]interface{})
	if data, err := os.ReadFile(mainPath); err == nil {
		json.Unmarshal(data, &existing)
	}

	// Add single-layer weights
	existing["single_layer_weights"] = net.GetWeights()
	existing["single_layer_bias"] = net.GetBiases()

	// Write back
	jsonData, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(mainPath, jsonData, 0644)
}

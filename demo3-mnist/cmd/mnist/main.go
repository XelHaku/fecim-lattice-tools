// Demo 3: MNIST Digit Recognition on Ferroelectric Crossbar Arrays
//
// This demo allows users to draw digits and see them classified
// through a neural network implemented on ferroelectric crossbar arrays.
// Target: 87% accuracy (matching Dr. Tour's research results)
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"ironlattice-vis/demo2-crossbar/pkg/crossbar"
	"ironlattice-vis/demo3-mnist/pkg/mnist"
	"ironlattice-vis/demo3-mnist/pkg/training"
)

func main() {
	// Command-line flags
	train := flag.Bool("train", false, "Train the network on MNIST")
	evaluate := flag.Bool("evaluate", false, "Evaluate trained network on test set")
	interactive := flag.Bool("interactive", false, "Interactive digit drawing mode")
	epochs := flag.Int("epochs", 5, "Number of training epochs")
	hiddenSize := flag.Int("hidden", 128, "Hidden layer size")
	noiseLevel := flag.Float64("noise", 0.02, "Device noise level (0-1)")
	loadWeights := flag.String("load", "", "Load weights from file")
	saveWeights := flag.String("save", "", "Save weights to file")
	flag.Parse()

	fmt.Println("================================================")
	fmt.Println("  IronLattice Demo 3: MNIST Digit Recognition")
	fmt.Println("  Ferroelectric Compute-in-Memory Neural Network")
	fmt.Println("================================================")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Input layer: 784 (28x28 pixels)\n")
	fmt.Printf("  Hidden layer: %d neurons\n", *hiddenSize)
	fmt.Printf("  Output layer: 10 classes (digits 0-9)\n")
	fmt.Printf("  Device noise: %.2f%%\n", *noiseLevel*100)
	fmt.Printf("  Discrete levels: 30 (IronLattice advantage)\n")
	fmt.Printf("  Target accuracy: 87%%\n")

	// Create crossbar arrays for each layer
	// Layer 1: 784 inputs -> hidden neurons
	layer1Cfg := &crossbar.Config{
		Rows:       *hiddenSize,
		Cols:       784,
		NoiseLevel: *noiseLevel,
		ADCBits:    6,
		DACBits:    8,
	}
	layer1, err := crossbar.NewArray(layer1Cfg)
	if err != nil {
		log.Fatalf("Failed to create layer 1 crossbar: %v", err)
	}

	// Layer 2: hidden neurons -> 10 outputs
	layer2Cfg := &crossbar.Config{
		Rows:       10,
		Cols:       *hiddenSize,
		NoiseLevel: *noiseLevel,
		ADCBits:    6,
		DACBits:    8,
	}
	layer2, err := crossbar.NewArray(layer2Cfg)
	if err != nil {
		log.Fatalf("Failed to create layer 2 crossbar: %v", err)
	}

	// Create the network
	net := training.NewMNISTNetwork(layer1, layer2)

	// Load weights if specified
	if *loadWeights != "" {
		fmt.Printf("\nLoading weights from: %s\n", *loadWeights)
		if err := net.LoadWeights(*loadWeights); err != nil {
			log.Printf("Warning: Failed to load weights: %v", err)
			fmt.Println("Using random initialization instead.")
		} else {
			fmt.Println("Weights loaded successfully.")
		}
	}

	// Training mode
	if *train {
		runTraining(net, *epochs, *saveWeights)
		return
	}

	// Evaluation mode
	if *evaluate {
		runEvaluation(net)
		return
	}

	// Interactive mode (default)
	if *interactive || (!*train && !*evaluate) {
		runInteractive(net)
		return
	}
}

func runTraining(net *training.MNISTNetwork, epochs int, saveFile string) {
	fmt.Println("\n=== Training Mode ===")
	fmt.Println("Note: MNIST dataset required in ./data/ directory")
	fmt.Println("Download from: http://yann.lecun.com/exdb/mnist/")

	// Try to load MNIST data
	trainImages, trainLabels, err := mnist.LoadMNIST("demo3-mnist/data", true)
	if err != nil {
		fmt.Printf("\nCould not load MNIST training data: %v\n", err)
		fmt.Println("Running with synthetic training data for demonstration...")

		// Use synthetic data for demo
		trainImages, trainLabels = generateSyntheticData(1000)
	}

	fmt.Printf("\nTraining on %d samples for %d epochs...\n", len(trainImages), epochs)

	// Train
	for epoch := 0; epoch < epochs; epoch++ {
		loss := net.TrainEpoch(trainImages, trainLabels, 0.01)
		acc := net.Evaluate(trainImages[:1000], trainLabels[:1000])
		fmt.Printf("Epoch %d/%d - Loss: %.4f, Train Accuracy: %.1f%%\n",
			epoch+1, epochs, loss, acc*100)
	}

	// Final evaluation
	fmt.Println("\n=== Final Results ===")
	acc := net.Evaluate(trainImages, trainLabels)
	fmt.Printf("Training Accuracy: %.1f%%\n", acc*100)

	// Save weights if requested
	if saveFile != "" {
		fmt.Printf("Saving weights to: %s\n", saveFile)
		if err := net.SaveWeights(saveFile); err != nil {
			log.Printf("Failed to save weights: %v", err)
		}
	}
}

func runEvaluation(net *training.MNISTNetwork) {
	fmt.Println("\n=== Evaluation Mode ===")

	// Try to load MNIST test data
	testImages, testLabels, err := mnist.LoadMNIST("demo3-mnist/data", false)
	if err != nil {
		fmt.Printf("Could not load MNIST test data: %v\n", err)
		fmt.Println("Running with synthetic test data...")
		testImages, testLabels = generateSyntheticData(100)
	}

	fmt.Printf("Evaluating on %d test samples...\n", len(testImages))

	accuracy := net.Evaluate(testImages, testLabels)
	fmt.Printf("\n=== Test Accuracy: %.1f%% ===\n", accuracy*100)

	if accuracy >= 0.87 {
		fmt.Println("Target accuracy (87%) ACHIEVED!")
	} else {
		fmt.Printf("Below target (87%%). Train with more data/epochs.\n")
	}

	// Show some sample predictions
	fmt.Println("\nSample predictions:")
	for i := 0; i < min(10, len(testImages)); i++ {
		pred, conf := net.Predict(testImages[i])
		fmt.Printf("  Sample %d: Predicted=%d (%.1f%%), Actual=%d %s\n",
			i, pred, conf*100, testLabels[i],
			checkMark(pred == testLabels[i]))
	}
}

func runInteractive(net *training.MNISTNetwork) {
	fmt.Println("\n=== Interactive Mode ===")
	fmt.Println("Draw digits using ASCII art or enter coordinates.")
	fmt.Println("Commands:")
	fmt.Println("  draw    - Enter drawing mode (28x28 grid)")
	fmt.Println("  sample N - Classify sample digit N (0-9)")
	fmt.Println("  test    - Run on random test samples")
	fmt.Println("  quit    - Exit")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\nmnist> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return

		case "draw":
			fmt.Println("\nEnter 28 lines of 28 characters each.")
			fmt.Println("Use '#' or '*' for filled pixels, space or '.' for empty.")
			fmt.Println("Enter 'done' when finished, 'cancel' to abort.")

			img := make([]float64, 784)
			row := 0

			for row < 28 {
				fmt.Printf("Row %2d: ", row)
				if !scanner.Scan() {
					break
				}
				line := scanner.Text()

				if line == "done" {
					break
				}
				if line == "cancel" {
					fmt.Println("Cancelled.")
					continue
				}

				// Parse line
				for col := 0; col < 28 && col < len(line); col++ {
					ch := line[col]
					if ch == '#' || ch == '*' || ch == 'X' || ch == 'x' {
						img[row*28+col] = 1.0
					}
				}
				row++
			}

			// Show the image and classify
			showImage(img)
			pred, conf := net.Predict(img)
			showPrediction(net, img, pred, conf)

		case "sample":
			digit := 0
			if len(parts) > 1 {
				digit, _ = strconv.Atoi(parts[1])
				digit = digit % 10
			}

			img := createSampleDigit(digit)
			showImage(img)
			pred, conf := net.Predict(img)
			showPrediction(net, img, pred, conf)

		case "test":
			fmt.Println("\nRunning on 5 random test samples...")
			for i := 0; i < 5; i++ {
				digit := rand.Intn(10)
				img := createSampleDigit(digit)

				fmt.Printf("\n--- Sample %d (Expected: %d) ---\n", i+1, digit)
				showImage(img)
				pred, conf := net.Predict(img)
				showPrediction(net, img, pred, conf)
			}

		default:
			fmt.Println("Unknown command. Type 'help' for commands.")
		}
	}
}

func showImage(img []float64) {
	fmt.Println("\nInput Image:")
	for row := 0; row < 28; row++ {
		fmt.Print("  ")
		for col := 0; col < 28; col++ {
			val := img[row*28+col]
			if val > 0.75 {
				fmt.Print("██")
			} else if val > 0.5 {
				fmt.Print("▓▓")
			} else if val > 0.25 {
				fmt.Print("░░")
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
}

func showPrediction(net *training.MNISTNetwork, img []float64, pred int, conf float64) {
	fmt.Println("\n=== Crossbar Inference Result ===")

	// Show output probabilities
	probs := net.GetOutputProbabilities(img)
	fmt.Println("\nOutput probabilities (softmax):")
	for i := 0; i < 10; i++ {
		barLen := int(probs[i] * 40)
		bar := strings.Repeat("█", barLen) + strings.Repeat("░", 40-barLen)
		marker := " "
		if i == pred {
			marker = "→"
		}
		fmt.Printf("  %s %d: %s %.1f%%\n", marker, i, bar, probs[i]*100)
	}

	fmt.Printf("\nPredicted digit: %d (confidence: %.1f%%)\n", pred, conf*100)
}

func createSampleDigit(digit int) []float64 {
	img := make([]float64, 784)

	switch digit {
	case 0:
		// Draw a 0
		for row := 6; row < 24; row++ {
			for col := 8; col < 20; col++ {
				// Circle outline
				dr := float64(row - 14)
				dc := float64(col - 14)
				dist := math.Sqrt(dr*dr + dc*dc)
				if dist > 5 && dist < 9 {
					img[row*28+col] = 1.0
				}
			}
		}
	case 1:
		// Draw a 1
		for row := 6; row < 24; row++ {
			img[row*28+14] = 1.0
			img[row*28+15] = 1.0
		}
		// Top serif
		img[6*28+12] = 1.0
		img[6*28+13] = 1.0
		// Bottom line
		for col := 11; col < 19; col++ {
			img[23*28+col] = 1.0
		}
	case 2:
		// Draw a 2
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0  // Top
			img[14*28+col] = 1.0 // Middle
			img[23*28+col] = 1.0 // Bottom
		}
		for row := 6; row < 14; row++ {
			img[row*28+19] = 1.0 // Right top
		}
		for row := 14; row < 24; row++ {
			img[row*28+8] = 1.0 // Left bottom
		}
	case 3:
		// Draw a 3
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0  // Top
			img[14*28+col] = 1.0 // Middle
			img[23*28+col] = 1.0 // Bottom
		}
		for row := 6; row < 24; row++ {
			img[row*28+19] = 1.0 // Right side
		}
	case 4:
		// Draw a 4
		for row := 6; row < 15; row++ {
			img[row*28+8] = 1.0 // Left top
		}
		for col := 8; col < 20; col++ {
			img[14*28+col] = 1.0 // Middle
		}
		for row := 6; row < 24; row++ {
			img[row*28+16] = 1.0 // Right
		}
	case 5:
		// Draw a 5
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0  // Top
			img[14*28+col] = 1.0 // Middle
			img[23*28+col] = 1.0 // Bottom
		}
		for row := 6; row < 14; row++ {
			img[row*28+8] = 1.0 // Left top
		}
		for row := 14; row < 24; row++ {
			img[row*28+19] = 1.0 // Right bottom
		}
	case 6:
		// Draw a 6
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0  // Top
			img[14*28+col] = 1.0 // Middle
			img[23*28+col] = 1.0 // Bottom
		}
		for row := 6; row < 24; row++ {
			img[row*28+8] = 1.0 // Left
		}
		for row := 14; row < 24; row++ {
			img[row*28+19] = 1.0 // Right bottom
		}
	case 7:
		// Draw a 7
		for col := 8; col < 22; col++ {
			for row := 4; row < 7; row++ {
				img[row*28+col] = 1.0
			}
		}
		for i := 0; i < 20; i++ {
			row := 6 + i
			col := 20 - i/2
			if row < 28 && col >= 0 && col < 28 {
				img[row*28+col] = 1.0
				if col > 0 {
					img[row*28+col-1] = 1.0
				}
			}
		}
	case 8:
		// Draw an 8
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0  // Top
			img[14*28+col] = 1.0 // Middle
			img[23*28+col] = 1.0 // Bottom
		}
		for row := 6; row < 24; row++ {
			img[row*28+8] = 1.0  // Left
			img[row*28+19] = 1.0 // Right
		}
	case 9:
		// Draw a 9
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0  // Top
			img[14*28+col] = 1.0 // Middle
			img[23*28+col] = 1.0 // Bottom
		}
		for row := 6; row < 14; row++ {
			img[row*28+8] = 1.0 // Left top
		}
		for row := 6; row < 24; row++ {
			img[row*28+19] = 1.0 // Right
		}
	}

	return img
}

func generateSyntheticData(n int) ([][]float64, []int) {
	images := make([][]float64, n)
	labels := make([]int, n)

	for i := 0; i < n; i++ {
		digit := rand.Intn(10)
		labels[i] = digit
		images[i] = createSampleDigit(digit)

		// Add some noise for variation
		for j := range images[i] {
			if rand.Float64() < 0.05 {
				images[i][j] = 1.0 - images[i][j]
			}
		}
	}

	return images, labels
}

func checkMark(correct bool) string {
	if correct {
		return "✓"
	}
	return "✗"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//go:build ignore
// +build ignore

// Quick script to train and save weights for demo
package main

import (
	"fmt"
	"log"
	"math/rand"

	"ironlattice-vis/demo2-crossbar/pkg/crossbar"
	"ironlattice-vis/demo3-mnist/pkg/training"
)

func main() {
	rand.Seed(42) // Reproducible results

	// Create crossbar arrays
	layer1, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 128, Cols: 784, NoiseLevel: 0.01, ADCBits: 8, DACBits: 8,
	})
	layer2, _ := crossbar.NewArray(&crossbar.Config{
		Rows: 10, Cols: 128, NoiseLevel: 0.01, ADCBits: 8, DACBits: 8,
	})

	net := training.NewMNISTNetwork(layer1, layer2)

	// Generate training data
	fmt.Println("Generating training data...")
	trainImages, trainLabels := generateData(5000)
	testImages, testLabels := generateData(500)

	// Train
	fmt.Println("Training...")
	for epoch := 0; epoch < 10; epoch++ {
		loss := net.TrainEpoch(trainImages, trainLabels, 0.1)
		acc := net.Evaluate(testImages, testLabels)
		fmt.Printf("Epoch %d: Loss=%.4f, Accuracy=%.1f%%\n", epoch+1, loss, acc*100)
	}

	// Quantize to 30 levels
	net.QuantizeWeightsTo30Levels()

	// Final evaluation
	acc := net.Evaluate(testImages, testLabels)
	fmt.Printf("\nFinal Accuracy: %.1f%%\n", acc*100)

	// Save weights
	if err := net.SaveWeights("demo3-mnist/data/pretrained_weights.json"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Weights saved to demo3-mnist/data/pretrained_weights.json")
}

func generateData(n int) ([][]float64, []int) {
	images := make([][]float64, n)
	labels := make([]int, n)

	for i := 0; i < n; i++ {
		digit := rand.Intn(10)
		labels[i] = digit
		images[i] = createDigit(digit)
		// Add noise
		for j := range images[i] {
			if rand.Float64() < 0.05 {
				images[i][j] = 1.0 - images[i][j]
			}
		}
	}
	return images, labels
}

func createDigit(d int) []float64 {
	img := make([]float64, 784)
	// Simplified digit patterns
	patterns := [][]int{
		{6,8,20,6,24,8,6,24,20}, // 0: top, left-top, right-top, left-bottom, right-bottom, bottom
		{14,6,14,24,10,18},      // 1
		{6,8,20,14,8,20,23,8,20},// 2
		{6,8,20,14,8,20,23,8,20},// 3
		{6,8,14,8,14,8,20,6,24}, // 4
		{6,8,20,6,14,8,14,20,23},// 5
		{6,8,20,6,24,8,14,20,23},// 6
		{6,8,22,10,20},          // 7
		{6,8,20,6,24,14,8,20,23},// 8
		{6,8,20,6,14,14,20,23},  // 9
	}

	_ = patterns // Keep the complex pattern logic simple for this demo
	
	switch d {
	case 0:
		for row := 6; row < 24; row++ {
			img[row*28+8] = 1.0
			img[row*28+19] = 1.0
		}
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0
			img[23*28+col] = 1.0
		}
	case 1:
		for row := 6; row < 24; row++ {
			img[row*28+14] = 1.0
		}
	case 7:
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0
		}
		for i := 0; i < 17; i++ {
			img[(7+i)*28+(19-i/2)] = 1.0
		}
	default:
		// Generic digit shape
		for col := 8; col < 20; col++ {
			img[6*28+col] = 1.0
			if d != 1 && d != 7 {
				img[23*28+col] = 1.0
			}
		}
		if d != 1 && d != 7 {
			for row := 6; row < 24; row++ {
				if d%2 == 0 {
					img[row*28+8] = 1.0
				}
				if d != 4 && d != 5 && d != 6 {
					img[row*28+19] = 1.0
				}
			}
		}
	}
	return img
}

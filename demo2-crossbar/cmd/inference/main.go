// Demo 2: Neural Network Inference on Ferroelectric Crossbar Array
//
// This demo visualizes matrix-vector multiplication (MVM) operations
// on a simulated ferroelectric crossbar array for neural network inference.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"

	"ironlattice-vis/demo2-crossbar/pkg/crossbar"
	"ironlattice-vis/demo2-crossbar/pkg/network"
	"ironlattice-vis/demo2-crossbar/pkg/visualization"
)

func main() {
	// Command-line flags
	arraySize := flag.Int("size", 64, "Crossbar array size (NxN)")
	numLayers := flag.Int("layers", 3, "Number of neural network layers")
	batchSize := flag.Int("batch", 1, "Inference batch size")
	noiseLevel := flag.Float64("noise", 0.02, "Device noise level (0-1)")
	adcBits := flag.Int("adc", 6, "ADC resolution in bits")
	noColor := flag.Bool("no-color", false, "Disable colored output")
	benchmark := flag.Bool("benchmark", false, "Run inference benchmark")
	showArray := flag.Bool("show-array", false, "Show crossbar array state")
	showMVM := flag.Bool("show-mvm", false, "Show MVM operation")
	flag.Parse()

	fmt.Println("============================================")
	fmt.Println("  IronLattice Demo 2: Crossbar Array MVM")
	fmt.Println("  Ferroelectric Compute-in-Memory")
	fmt.Println("============================================")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Crossbar size: %d x %d\n", *arraySize, *arraySize)
	fmt.Printf("  Layers: %d\n", *numLayers)
	fmt.Printf("  Noise level: %.2f%%\n", *noiseLevel*100)
	fmt.Printf("  ADC bits: %d (DAC bits: 8)\n", *adcBits)
	fmt.Printf("  Discrete levels: 30\n")

	// Create crossbar array configuration
	arrayCfg := &crossbar.Config{
		Rows:       *arraySize,
		Cols:       *arraySize,
		NoiseLevel: *noiseLevel,
		ADCBits:    *adcBits,
		DACBits:    8,
	}

	// Create the crossbar array
	array, err := crossbar.NewArray(arrayCfg)
	if err != nil {
		log.Fatalf("Failed to create crossbar array: %v", err)
	}

	// Create visualizer
	vis := visualization.NewTerminalVisualizer(array, !*noColor)

	// Show array state if requested
	if *showArray {
		// Program some random weights for demonstration
		programRandomWeights(array)
		vis.ShowCrossbarState()
	}

	// Show MVM operation if requested
	if *showMVM {
		// Program some weights
		programRandomWeights(array)

		// Create random input
		input := make([]float64, *arraySize)
		for i := range input {
			input[i] = rand.Float64()
		}

		// Perform MVM
		output, err := array.MVM(input)
		if err != nil {
			log.Fatalf("MVM failed: %v", err)
		}

		vis.ShowCrossbarState()
		vis.ShowMVMOperation(input, output)
		return
	}

	// Create neural network for inference demo
	netCfg := &network.Config{
		InputSize:  784, // MNIST input (28x28)
		HiddenSize: *arraySize,
		OutputSize: 10, // MNIST classes
		NumLayers:  *numLayers,
	}

	net, err := network.NewNetwork(netCfg, array)
	if err != nil {
		log.Fatalf("Failed to create network: %v", err)
	}

	if *benchmark {
		runBenchmark(net, *batchSize)
		return
	}

	// Run inference demo with a sample digit
	fmt.Println("\n--- Running Neural Network Inference Demo ---")

	// Create a sample "digit" pattern (a simple 7)
	input := createSampleDigit7()

	// Run inference
	output := net.Forward(input)

	// Find prediction
	maxIdx := 0
	maxVal := output[0]
	for i, v := range output {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}

	// Show inference visualization
	activations := [][]float64{output}
	vis.ShowNeuralNetworkInference(*numLayers, input, activations, maxIdx, maxVal)

	fmt.Println("\n--- Demo Complete ---")
	fmt.Println("Note: Weights are randomly initialized. Train with MNIST for accurate predictions.")
}

func runBenchmark(net *network.Network, batchSize int) {
	fmt.Printf("\nRunning benchmark (batch=%d)...\n", batchSize)

	// Generate random inputs
	inputs := make([][]float64, batchSize)
	for i := range inputs {
		inputs[i] = make([]float64, 784)
		for j := range inputs[i] {
			inputs[i][j] = float64(i*j%256) / 255.0
		}
	}

	// Run inference
	var totalOps int64
	for _, input := range inputs {
		net.Forward(input)
		totalOps += net.GetOpsCount()
	}

	fmt.Printf("Total MAC operations: %d\n", totalOps)
	fmt.Println("Benchmark complete.")
}

func programRandomWeights(array *crossbar.Array) {
	rows, cols := array.Rows(), array.Cols()
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Quantize to 30 levels
			level := rand.Intn(30)
			weight := float64(level) / 29.0
			array.ProgramWeight(i, j, weight)
		}
	}
}

func createSampleDigit7() []float64 {
	// Create a 28x28 image of digit "7"
	img := make([]float64, 784)

	// Draw horizontal line at top (row 4-6)
	for col := 8; col < 22; col++ {
		for row := 4; row < 7; row++ {
			img[row*28+col] = 1.0
		}
	}

	// Draw diagonal line from top-right to bottom-left
	for i := 0; i < 20; i++ {
		row := 6 + i
		col := 20 - i/2
		if row < 28 && col >= 0 && col < 28 {
			// Make the stroke thicker
			for dr := -1; dr <= 1; dr++ {
				for dc := -1; dc <= 1; dc++ {
					r := row + dr
					c := col + dc
					if r >= 0 && r < 28 && c >= 0 && c < 28 {
						img[r*28+c] = 1.0
					}
				}
			}
		}
	}

	return img
}

// Demo 2: Neural Network Inference on Ferroelectric Crossbar Array
//
// This demo visualizes matrix-vector multiplication (MVM) operations
// on a simulated ferroelectric crossbar array for neural network inference.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ironlattice/vis/demo2-inference/pkg/crossbar"
	"github.com/ironlattice/vis/demo2-inference/pkg/network"
)

func main() {
	// Command-line flags
	arraySize := flag.Int("size", 64, "Crossbar array size (NxN)")
	numLayers := flag.Int("layers", 3, "Number of neural network layers")
	batchSize := flag.Int("batch", 1, "Inference batch size")
	noiseLevel := flag.Float64("noise", 0.02, "Device noise level (0-1)")
	adcBits := flag.Int("adc", 6, "ADC resolution in bits")
	headless := flag.Bool("headless", false, "Run without visualization")
	benchmark := flag.Bool("benchmark", false, "Run inference benchmark")
	flag.Parse()

	fmt.Println("IronLattice Demo 2: Neural Network Inference")
	fmt.Println("=============================================")
	fmt.Printf("Crossbar size: %dx%d\n", *arraySize, *arraySize)
	fmt.Printf("Layers: %d\n", *numLayers)
	fmt.Printf("Noise level: %.2f%%\n", *noiseLevel*100)
	fmt.Printf("ADC bits: %d\n", *adcBits)

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

	// Create neural network
	netCfg := &network.Config{
		InputSize:  784, // MNIST input (28x28)
		HiddenSize: *arraySize,
		OutputSize: 10,  // MNIST classes
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

	if *headless {
		// Run single inference
		input := make([]float64, 784)
		for i := range input {
			input[i] = 0.5 // Dummy input
		}
		output := net.Forward(input)
		fmt.Printf("\nOutput: %v\n", output)
		return
	}

	// TODO: Launch visualization
	fmt.Println("\nVisualization mode not yet implemented.")
	fmt.Println("Use --headless for command-line inference.")
	os.Exit(0)
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

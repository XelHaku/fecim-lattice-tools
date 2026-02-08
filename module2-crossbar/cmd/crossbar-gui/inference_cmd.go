package crossbarcmd

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
	"fecim-lattice-tools/module2-crossbar/pkg/network"
	"fecim-lattice-tools/module2-crossbar/pkg/visualization"
)

func RunInference(args []string) error {
	fs := flag.NewFlagSet("inference", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	arraySize := fs.Int("size", 64, "Crossbar array size (NxN)")
	numLayers := fs.Int("layers", 3, "Number of neural network layers")
	batchSize := fs.Int("batch", 1, "Inference batch size")
	noiseLevel := fs.Float64("noise", 0.02, "Device noise level (0-1)")
	adcBits := fs.Int("adc", 6, "ADC resolution in bits")
	seed := fs.Int64("seed", 1, "Random seed for deterministic runs (0 = time-based)")
	noColor := fs.Bool("no-color", false, "Disable colored output")
	benchmark := fs.Bool("benchmark", false, "Run inference benchmark")
	showArray := fs.Bool("show-array", false, "Show crossbar array state")
	showMVM := fs.Bool("show-mvm", false, "Show MVM operation")
	showIRDrop := fs.Bool("show-irdrop", false, "Show IR drop analysis")
	showSneak := fs.Bool("show-sneak", false, "Show sneak path analysis")
	showNonidealities := fs.Bool("show-nonidealities", false, "Show all non-ideality effects")
	help := fs.Bool("help", false, "Show help")
	helpShort := fs.Bool("h", false, "Show help (shorthand)")

	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintln(out, "FeCIM Crossbar Inference (Terminal)")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out, "  go run ./cmd/fecim-lattice-tools crossbar inference [options]")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Options:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(fs.Output(), "Error:", err)
		fs.Usage()
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if *help || *helpShort {
		fs.Usage()
		return nil
	}

	fmt.Println("============================================")
	fmt.Println("  FeCIM Demo 2: Crossbar Array MVM")
	fmt.Println("  Ferroelectric Compute-in-Memory")
	fmt.Println("============================================")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Crossbar size: %d x %d\n", *arraySize, *arraySize)
	fmt.Printf("  Layers: %d\n", *numLayers)
	fmt.Printf("  Noise level: %.2f%%\n", *noiseLevel*100)
	fmt.Printf("  ADC bits: %d (DAC bits: 8)\n", *adcBits)
	fmt.Printf("  Discrete levels: 30 (demo baseline; conference claim)\n")

	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(*seed))

	arrayCfg := &crossbar.Config{
		Rows:       *arraySize,
		Cols:       *arraySize,
		NoiseLevel: *noiseLevel,
		ADCBits:    *adcBits,
		DACBits:    8,
	}

	array, err := crossbar.NewArray(arrayCfg)
	if err != nil {
		return fmt.Errorf("failed to create crossbar array: %w", err)
	}

	vis := visualization.NewTerminalVisualizer(array, !*noColor)

	if *showArray {
		programRandomWeights(array, rng)
		vis.ShowCrossbarState()
	}

	if *showMVM {
		programRandomWeights(array, rng)

		input := make([]float64, *arraySize)
		for i := range input {
			input[i] = rng.Float64()
		}

		output, err := array.MVM(input)
		if err != nil {
			return fmt.Errorf("MVM failed: %w", err)
		}

		vis.ShowCrossbarState()
		vis.ShowMVMOperation(input, output)
		return nil
	}

	if *showIRDrop || *showNonidealities {
		programRandomWeights(array, rng)

		input := make([]float64, *arraySize)
		for i := range input {
			input[i] = rng.Float64()
		}

		wireParams := crossbar.DefaultWireParams()
		irAnalysis := array.AnalyzeIRDrop(input, wireParams)

		vis.ShowCrossbarState()
		vis.ShowIRDropAnalysis(irAnalysis)

		if !*showNonidealities {
			return nil
		}
	}

	if *showSneak || *showNonidealities {
		programRandomWeights(array, rng)

		selectedRow := *arraySize / 2
		selectedCol := *arraySize / 2

		sneakAnalysis := array.AnalyzeSneakPaths(selectedRow, selectedCol)

		if !*showIRDrop && !*showNonidealities {
			vis.ShowCrossbarState()
		}
		vis.ShowSneakPathAnalysis(sneakAnalysis, selectedRow, selectedCol)

		if !*showNonidealities {
			return nil
		}
	}

	if *showNonidealities {
		fmt.Println("\n=== MVM Comparison: Ideal vs With Non-Idealities ===")

		input := make([]float64, *arraySize)
		for i := range input {
			input[i] = rng.Float64()
		}

		idealOutput, err := array.MVM(input)
		if err != nil {
			return fmt.Errorf("MVM failed: %w", err)
		}

		wireParams := crossbar.DefaultWireParams()
		actualOutput, irAnalysis, err := array.MVMWithIRDrop(input, wireParams)
		if err != nil {
			return fmt.Errorf("MVM with IR drop failed: %w", err)
		}

		vis.ShowMVMWithNonidealities(input, idealOutput, actualOutput, irAnalysis)
		return nil
	}

	netCfg := &network.Config{
		InputSize:  784,
		HiddenSize: *arraySize,
		OutputSize: 10,
		NumLayers:  *numLayers,
	}

	net, err := network.NewNetwork(netCfg, array)
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	if *benchmark {
		runBenchmark(net, *batchSize)
		return nil
	}

	fmt.Println("\n--- Running Neural Network Inference Demo ---")

	input := createSampleDigit7()

	output, err := net.Forward(input)
	if err != nil {
		return fmt.Errorf("inference failed: %w", err)
	}

	maxIdx := 0
	maxVal := output[0]
	for i, v := range output {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}

	activations := [][]float64{output}
	vis.ShowNeuralNetworkInference(*numLayers, input, activations, maxIdx, maxVal)

	fmt.Println("\n--- Demo Complete ---")
	fmt.Println("Note: Weights are randomly initialized. Train with MNIST for accurate predictions.")
	return nil
}

func runBenchmark(net *network.Network, batchSize int) {
	fmt.Printf("\nRunning benchmark (batch=%d)...\n", batchSize)

	inputs := make([][]float64, batchSize)
	for i := range inputs {
		inputs[i] = make([]float64, 784)
		for j := range inputs[i] {
			inputs[i][j] = float64(i*j%256) / 255.0
		}
	}

	var totalOps int64
	for _, input := range inputs {
		_, err := net.Forward(input)
		if err != nil {
			log.Printf("Inference error: %v", err)
			continue
		}
		totalOps += net.GetOpsCount()
	}

	fmt.Printf("Total MAC operations: %d\n", totalOps)
	fmt.Println("Benchmark complete.")
}

func programRandomWeights(array *crossbar.Array, rng *rand.Rand) {
	rows, cols := array.Rows(), array.Cols()
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			level := rng.Intn(30)
			weight := float64(level) / 29.0
			array.ProgramWeight(i, j, weight)
		}
	}
}

func createSampleDigit7() []float64 {
	img := make([]float64, 784)

	for col := 8; col < 22; col++ {
		for row := 4; row < 7; row++ {
			img[row*28+col] = 1.0
		}
	}

	for i := 0; i < 20; i++ {
		row := 6 + i
		col := 20 - i/2
		if row < 28 && col >= 0 && col < 28 {
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

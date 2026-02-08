// Demo 3: MNIST Digit Recognition on Ferroelectric Crossbar Arrays
//
// This demo allows users to draw digits and see them classified
// through a neural network implemented on ferroelectric crossbar arrays.
// Target: Physics-limited accuracy (typically 85-90% with 30 levels)
//
// Common flags:
//   - --json: Output results as JSON
//   - --quiet: Suppress informational output
//   - --config: Load configuration from YAML/JSON file
//   - --batch: Process multiple images from file
package mnistcli

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
	"fecim-lattice-tools/module3-mnist/pkg/core"
	"fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/module3-mnist/pkg/training"
	"fecim-lattice-tools/shared/cli"
	"fecim-lattice-tools/shared/logging"
)

// MNISTConfig holds configuration for the MNIST CLI.
type MNISTConfig struct {
	HiddenSize int     `json:"hidden_size" yaml:"hidden_size"`
	NoiseLevel float64 `json:"noise_level" yaml:"noise_level"`
	Epochs     int     `json:"epochs" yaml:"epochs"`
	Levels     []int   `json:"levels" yaml:"levels"`
}

// EvaluationResult represents evaluation results for JSON output.
type EvaluationResult struct {
	Samples      int     `json:"samples"`
	Accuracy     float64 `json:"accuracy"`
	FPAccuracy   float64 `json:"fp_accuracy,omitempty"`
	CIMAccuracy  float64 `json:"cim_accuracy,omitempty"`
	AgreeRate    float64 `json:"agree_rate,omitempty"`
	AvgKL        float64 `json:"avg_kl,omitempty"`
	AvgEnergy    float64 `json:"avg_energy_uj,omitempty"`
	Levels       int     `json:"levels,omitempty"`
}

func Run(args []string) error {
	logging.EnableFileLogging()
	_ = logging.NewLogger("mnist-cli")

	fs := flag.NewFlagSet("mnist", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	// Common CLI flags
	commonFlags := cli.NewCommonFlags()
	commonFlags.Register(fs)

	// Command-line flags
	train := fs.Bool("train", false, "Train the network on MNIST")
	evaluate := fs.Bool("evaluate", false, "Evaluate trained network on test set")
	interactive := fs.Bool("interactive", false, "Interactive digit drawing mode")
	epochs := fs.Int("epochs", 5, "Number of training epochs")
	hiddenSize := fs.Int("hidden", 128, "Hidden layer size")
	noiseLevel := fs.Float64("noise", 0.02, "Device noise level (0-1)")
	loadWeights := fs.String("load", "", "Load weights from file")
	saveWeights := fs.String("save", "", "Save weights to file")
	coreEvaluate := fs.Bool("core-eval", false, "Evaluate using dual-mode core network (FP vs CIM)")
	coreSamples := fs.Int("core-samples", 1000, "Samples for core-eval (0=all)")
	coreLevels := fs.String("core-levels", "", "Comma-separated levels for core-eval sweep (e.g., 8,16,24,31)")
	exportLevels := fs.String("export-levels", "", "Comma-separated levels to export (pretrained_weights_{N}.json)")
	exportDir := fs.String("export-dir", "", "Comma-separated output directories for export-levels (default: data/pretrained-weigths)")

	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintln(out, "FeCIM MNIST CLI")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out, "  fecim-lattice-tools mnist cli [options]")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Options:")
		fs.PrintDefaults()
		fmt.Fprintln(out, cli.CommonUsage())
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(fs.Output(), "Error:", err)
		fs.Usage()
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if commonFlags.WantsHelp() {
		fs.Usage()
		return nil
	}

	// Load config file if specified
	var cfg MNISTConfig
	if commonFlags.Config != "" {
		loader := cli.NewConfigLoader(commonFlags.Config)
		if err := loader.Load(&cfg); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		// Apply config values (flags take precedence)
		if cfg.HiddenSize > 0 && *hiddenSize == 128 {
			*hiddenSize = cfg.HiddenSize
		}
		if cfg.NoiseLevel > 0 && *noiseLevel == 0.02 {
			*noiseLevel = cfg.NoiseLevel
		}
		if cfg.Epochs > 0 && *epochs == 5 {
			*epochs = cfg.Epochs
		}
	}

	// Create output writer
	out, err := cli.NewOutputWriter(commonFlags)
	if err != nil {
		return err
	}
	defer out.Close()

	out.Println("================================================")
	out.Println("  FeCIM Demo 3: MNIST Digit Recognition")
	out.Println("  Ferroelectric Compute-in-Memory Neural Network")
	out.Println("================================================")
	out.Print("\nConfiguration:\n")
	out.Print("  Input layer: 784 (28x28 pixels)\n")
	out.Print("  Hidden layer: %d neurons\n", *hiddenSize)
	out.Print("  Output layer: 10 classes (digits 0-9)\n")
	out.Print("  Device noise: %.2f%%\n", *noiseLevel*100)
	out.Print("  Discrete levels: 30 (demo baseline; conference claim)\n")
	out.Print("  Target accuracy: Physics-limited\n")

	if *exportLevels != "" {
		levels, err := parseLevelList(*exportLevels)
		if err != nil {
			return fmt.Errorf("invalid export levels: %w\n\nHint: Specify levels as comma-separated integers, e.g., --export-levels 8,16,24,31", err)
		}
		outDirs, err := parseDirList(*exportDir)
		if err != nil {
			return fmt.Errorf("invalid export directories: %w", err)
		}
		if len(outDirs) == 0 {
			outDirs = []string{filepath.Join("data", "pretrained-weigths")}
		}
		if err := runExportQuantizedWeights(levels, *loadWeights, outDirs, *hiddenSize); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
		return nil
	}

	if *coreEvaluate {
		levels, err := parseLevelList(*coreLevels)
		if err != nil {
			return fmt.Errorf("invalid core levels: %w\n\nHint: Specify levels as comma-separated integers, e.g., --core-levels 8,16,24,31", err)
		}
		runCoreEvaluation(*hiddenSize, *noiseLevel, *loadWeights, *coreSamples, levels)
		return nil
	}

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
		return fmt.Errorf("failed to create layer 1 crossbar: %w\n\nThis may indicate insufficient memory or invalid configuration", err)
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
		return fmt.Errorf("failed to create layer 2 crossbar: %w\n\nThis may indicate insufficient memory or invalid configuration", err)
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

	// If evaluating or running interactively without explicit weights, try defaults
	if *loadWeights == "" && !*train {
		defaultWeights := "module3-mnist/data/pretrained_weights.json"
		if _, err := os.Stat(defaultWeights); err == nil {
			fmt.Printf("\nLoading default weights from: %s\n", defaultWeights)
			if err := net.LoadWeights(defaultWeights); err != nil {
				log.Printf("Warning: Failed to load default weights: %v", err)
				fmt.Println("Using random initialization instead.")
			} else {
				fmt.Println("Default weights loaded successfully.")
			}
		}
	}

	// Training mode
	if *train {
		runTraining(net, *epochs, *saveWeights)
		return nil
	}

	// Evaluation mode
	if *evaluate {
		runEvaluation(net)
		return nil
	}

	// Interactive mode (default)
	if *interactive || (!*train && !*evaluate) {
		runInteractive(net)
		return nil
	}

	return nil
}

func runTraining(net *training.MNISTNetwork, epochs int, saveFile string) {
	fmt.Println("\n=== Training Mode ===")
	fmt.Println("Note: MNIST dataset required in ./data/ directory")
	fmt.Println("Download from: http://yann.lecun.com/exdb/mnist/")

	// Try to load MNIST data
	trainImages, trainLabels, err := mnist.LoadMNIST("module3-mnist/data", true)
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
	testImages, testLabels, err := mnist.LoadMNIST("module3-mnist/data", false)
	if err != nil {
		fmt.Printf("Could not load MNIST test data: %v\n", err)
		fmt.Println("Running with synthetic test data...")
		testImages, testLabels = generateSyntheticData(100)
	}

	fmt.Printf("Evaluating on %d test samples...\n", len(testImages))

	accuracy := net.Evaluate(testImages, testLabels)
	fmt.Printf("\n=== Test Accuracy: %.1f%% ===\n", accuracy*100)
	log.Printf("Evaluation complete: accuracy=%.1f%% on %d samples", accuracy*100, len(testImages))

	if accuracy >= 0.85 {
		fmt.Println("Target accuracy (>85%) ACHIEVED!")
	} else {
		fmt.Printf("Below target (>85%%). Train with more data/epochs.\n")
	}

	// Compute and display confusion matrix
	confMatrix := net.ComputeConfusionMatrix(testImages, testLabels)
	showConfusionMatrix(confMatrix)

	// Show per-class metrics
	precision, recall, f1 := net.GetPerClassMetrics(confMatrix)
	showPerClassMetrics(precision, recall, f1)

	// Show some sample predictions
	fmt.Println("\nSample predictions:")
	for i := 0; i < min(10, len(testImages)); i++ {
		pred, conf := net.Predict(testImages[i])
		fmt.Printf("  Sample %d: Predicted=%d (%.1f%%), Actual=%d %s\n",
			i, pred, conf*100, testLabels[i],
			checkMark(pred == testLabels[i]))
	}
}

type coreEvalMetrics struct {
	fpAcc      float64
	cimAcc     float64
	agreeRate  float64
	avgKL      float64
	avgEnergy  float64
	sample0    *core.InferenceResult
	sample0Hit bool
	samples    int
}

func runCoreEvaluation(hiddenSize int, noiseLevel float64, loadFile string, maxSamples int, levels []int) {
	fmt.Println("\n=== Core Dual-Path Evaluation (FP vs CIM) ===")

	testImages, testLabels, err := mnist.LoadMNIST("module3-mnist/data", false)
	if err != nil {
		fmt.Printf("Could not load MNIST test data: %v\n", err)
		fmt.Println("Running with synthetic test data...")
		testImages, testLabels = generateSyntheticData(200)
	}

	totalSamples := len(testImages)
	if maxSamples > 0 && maxSamples < totalSamples {
		totalSamples = maxSamples
		testImages = testImages[:totalSamples]
		testLabels = testLabels[:totalSamples]
	}
	fmt.Printf("Evaluating on %d test samples...\n", totalSamples)

	net := core.NewDualModeNetwork(784, hiddenSize, 10)
	net.Config.NoiseLevel = noiseLevel
	net.Config.ADCBits = 8
	net.Config.DACBits = 8

	weightsPath, err := resolveWeightsPath(loadFile)
	if err != nil {
		log.Printf("Warning: %v", err)
	} else {
		fmt.Printf("\nLoading weights from: %s\n", weightsPath)
		if err := net.LoadWeights(weightsPath); err != nil {
			log.Printf("Warning: Failed to load weights: %v", err)
			fmt.Println("Continuing with random initialization.")
		} else {
			fmt.Println("Weights loaded successfully.")
		}
	}

	// Ensure noise level is aligned with CLI after loading weights.
	net.Config.NoiseLevel = noiseLevel

	if len(levels) == 0 {
		levels = []int{net.Config.NumLevels}
	}

	perLayerEnabled, l1Levels, l2Levels := net.GetPerLayerQuantInfo()
	log.Printf("Core eval config: levels=%d l1Levels=%d l2Levels=%d perLayer=%v noise=%.4f adcBits=%d dacBits=%d singleLayer=%v samples=%d",
		net.Config.NumLevels, l1Levels, l2Levels, perLayerEnabled,
		net.Config.NoiseLevel, net.Config.ADCBits, net.Config.DACBits,
		net.Config.SingleLayer, totalSamples)

	if len(levels) > 1 {
		fmt.Println("\nLevels | FP Acc | CIM Acc | Agree | Avg KL | Avg Energy (uJ)")
		fmt.Println("------------------------------------------------------------")
	}

	for _, level := range levels {
		net.SetNumLevels(level)
		metrics := evaluateCoreMetrics(net, testImages, testLabels)
		if metrics.samples == 0 {
			fmt.Printf("Levels %d: no samples evaluated\n", level)
			continue
		}
		if metrics.sample0Hit && metrics.sample0 != nil {
			log.Printf("Core eval sample0 (levels=%d): fpPred=%d (conf=%.4f) cimPred=%d (conf=%.4f) agree=%v kl=%.6f energy_uJ=%.6f",
				level,
				metrics.sample0.FPPrediction, metrics.sample0.FPConfidence,
				metrics.sample0.CIMPrediction, metrics.sample0.CIMConfidence,
				metrics.sample0.Agree, metrics.sample0.Disagreement, metrics.sample0.EnergyUsed)
		}

		if len(levels) > 1 {
			fmt.Printf("%6d | %6.1f%% | %7.1f%% | %5.1f%% | %6.4f | %13.6f\n",
				level, metrics.fpAcc*100, metrics.cimAcc*100,
				metrics.agreeRate*100, metrics.avgKL, metrics.avgEnergy)
		} else {
			fmt.Printf("\nFP Accuracy: %.1f%%\n", metrics.fpAcc*100)
			fmt.Printf("CIM Accuracy: %.1f%%\n", metrics.cimAcc*100)
			fmt.Printf("Agreement Rate: %.1f%%\n", metrics.agreeRate*100)
			fmt.Printf("Average KL Divergence: %.6f\n", metrics.avgKL)
			fmt.Printf("Average Energy: %.6f μJ\n", metrics.avgEnergy)
		}

		log.Printf("Core evaluation complete (levels=%d): fpAcc=%.1f%% cimAcc=%.1f%% agree=%.1f%% avgKL=%.6f avgEnergy_uJ=%.6f samples=%d",
			level, metrics.fpAcc*100, metrics.cimAcc*100, metrics.agreeRate*100,
			metrics.avgKL, metrics.avgEnergy, metrics.samples)
	}
}

func evaluateCoreMetrics(net *core.DualModeNetwork, images [][]float64, labels []int) coreEvalMetrics {
	var fpCorrect, cimCorrect, agreeCount int
	var totalKL, totalEnergy float64
	var evaluated int
	var sample0 *core.InferenceResult

	for i := 0; i < len(images); i++ {
		result := net.Infer(images[i])
		if result == nil {
			continue
		}
		evaluated++
		if result.FPPrediction == labels[i] {
			fpCorrect++
		}
		if result.CIMPrediction == labels[i] {
			cimCorrect++
		}
		if result.Agree {
			agreeCount++
		}
		totalKL += result.Disagreement
		totalEnergy += result.EnergyUsed

		if i == 0 {
			sample0 = result
		}
	}

	if evaluated == 0 {
		return coreEvalMetrics{}
	}

	return coreEvalMetrics{
		fpAcc:      float64(fpCorrect) / float64(evaluated),
		cimAcc:     float64(cimCorrect) / float64(evaluated),
		agreeRate:  float64(agreeCount) / float64(evaluated),
		avgKL:      totalKL / float64(evaluated),
		avgEnergy:  totalEnergy / float64(evaluated),
		sample0:    sample0,
		sample0Hit: sample0 != nil,
		samples:    evaluated,
	}
}

func runExportQuantizedWeights(levels []int, loadFile string, outDirs []string, hiddenSize int) error {
	fmt.Println("\n=== Export Quantized Weights ===")

	if len(levels) == 0 {
		return fmt.Errorf("no levels specified")
	}

	if len(outDirs) == 0 {
		outDirs = []string{"module3-mnist/data"}
	}
	for _, outDir := range outDirs {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("create output dir %s: %w", outDir, err)
		}
	}

	net := core.NewDualModeNetwork(784, hiddenSize, 10)

	weightsPath, err := resolveWeightsPath(loadFile)
	if err != nil {
		return err
	}
	baseBytes, err := os.ReadFile(weightsPath)
	if err != nil {
		return fmt.Errorf("read base weights: %w", err)
	}
	if err := net.LoadWeights(weightsPath); err != nil {
		return fmt.Errorf("load weights: %w", err)
	}

	for _, outDir := range outDirs {
		baseOut := filepath.Join(outDir, "pretrained_weights.json")
		if err := os.WriteFile(baseOut, baseBytes, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", baseOut, err)
		}
		fmt.Printf("Wrote %s\n", baseOut)
	}

	w1Min, w1Max := weightRange(net.FPWeights1)
	w2Min, w2Max := weightRange(net.FPWeights2)

	for _, level := range levels {
		if level < 2 {
			return fmt.Errorf("levels must be >= 2, got %d", level)
		}

		qW1, l1Scale, l1Offset := quantizeNormalized(net.FPWeights1, w1Min, w1Max, level)
		qW2, l2Scale, l2Offset := quantizeNormalized(net.FPWeights2, w2Min, w2Max, level)

		data := core.WeightsFile{
			Layer1Weights:     qW1,
			Layer2Weights:     qW2,
			Biases1:           net.FPBias1,
			Biases2:           net.FPBias2,
			L1Scale:           l1Scale,
			L1Offset:          l1Offset,
			L2Scale:           l2Scale,
			L2Offset:          l2Offset,
			QuantLevels:       level,
			Layer1QuantLevels: level,
			Layer2QuantLevels: level,
		}

		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal weights for level %d: %w", level, err)
		}

		for _, outDir := range outDirs {
			outPath := filepath.Join(outDir, fmt.Sprintf("pretrained_weights_%d.json", level))
			if err := os.WriteFile(outPath, jsonData, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", outPath, err)
			}
			fmt.Printf("Wrote %s\n", outPath)
		}
	}

	fmt.Println("Export complete.")
	return nil
}

func weightRange(weights [][]float64) (float64, float64) {
	if len(weights) == 0 || len(weights[0]) == 0 {
		return 0, 0
	}
	minVal := weights[0][0]
	maxVal := weights[0][0]
	for i := range weights {
		for j := range weights[i] {
			if weights[i][j] < minVal {
				minVal = weights[i][j]
			}
			if weights[i][j] > maxVal {
				maxVal = weights[i][j]
			}
		}
	}
	return minVal, maxVal
}

func quantizeNormalized(weights [][]float64, minVal, maxVal float64, levels int) ([][]float64, float64, float64) {
	scale := maxVal - minVal
	offset := minVal
	rows := len(weights)
	if rows == 0 {
		return weights, 0, 0
	}
	cols := len(weights[0])
	q := make([][]float64, rows)

	if scale == 0 {
		scale = 1
		for i := range weights {
			q[i] = make([]float64, cols)
		}
		return q, scale, offset
	}

	for i := range weights {
		q[i] = make([]float64, cols)
		for j := range weights[i] {
			norm := (weights[i][j] - minVal) / scale
			if norm < 0 {
				norm = 0
			}
			if norm > 1 {
				norm = 1
			}
			if levels > 1 {
				q[i][j] = math.Round(norm*float64(levels-1)) / float64(levels-1)
			} else {
				q[i][j] = norm
			}
		}
	}
	return q, scale, offset
}

func parseLevelList(levelsStr string) ([]int, error) {
	trimmed := strings.TrimSpace(levelsStr)
	if trimmed == "" {
		return nil, nil
	}
	parts := strings.Split(trimmed, ",")
	levelSet := make(map[int]struct{})
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		level, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid level %q", part)
		}
		levelSet[level] = struct{}{}
	}
	if len(levelSet) == 0 {
		return nil, fmt.Errorf("no valid levels found")
	}
	result := make([]int, 0, len(levelSet))
	for level := range levelSet {
		result = append(result, level)
	}
	sort.Ints(result)
	return result, nil
}

func parseDirList(dirsStr string) ([]string, error) {
	trimmed := strings.TrimSpace(dirsStr)
	if trimmed == "" {
		return nil, nil
	}
	parts := strings.Split(trimmed, ",")
	dirSet := make(map[string]struct{})
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		dirSet[filepath.Clean(part)] = struct{}{}
	}
	if len(dirSet) == 0 {
		return nil, fmt.Errorf("no valid directories found")
	}
	result := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		result = append(result, dir)
	}
	sort.Strings(result)
	return result, nil
}

func resolveWeightsPath(loadFile string) (string, error) {
	if loadFile != "" {
		return loadFile, nil
	}

	candidates := []string{
		filepath.Join("data", "pretrained-weigths", "pretrained_weights.json"),
		filepath.Join("data", "pretrained-weights", "pretrained_weights.json"),
		filepath.Join("module3-mnist", "data", "pretrained_weights.json"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("default weights not found (checked %s)", strings.Join(candidates, ", "))
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

	// Show layer-by-layer activations
	showLayerActivations(net, img)

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

func showLayerActivations(net *training.MNISTNetwork, img []float64) {
	input, hidden, output := net.GetLayerActivations(img)

	fmt.Println("\n─── Layer-by-Layer Activations ───")

	// Input layer summary
	activePixels := 0
	for _, v := range input {
		if v > 0.5 {
			activePixels++
		}
	}
	fmt.Printf("\nInput Layer (784 pixels):\n")
	fmt.Printf("  Active pixels: %d / 784 (%.1f%%)\n", activePixels, float64(activePixels)/784*100)

	// Hidden layer visualization (show first 64 neurons)
	fmt.Printf("\nHidden Layer Activations (%d neurons):\n", len(hidden))
	fmt.Print("  ")
	maxShow := min(64, len(hidden))
	for i := 0; i < maxShow; i++ {
		char := activationToChar(hidden[i])
		fmt.Print(char)
		if (i+1)%32 == 0 {
			fmt.Println()
			if i < maxShow-1 {
				fmt.Print("  ")
			}
		}
	}
	if len(hidden) > maxShow {
		fmt.Printf("... (%d more)\n", len(hidden)-maxShow)
	}

	// Stats
	activeHidden := 0
	maxAct := 0.0
	sumAct := 0.0
	for _, h := range hidden {
		if h > 0.1 {
			activeHidden++
		}
		if h > maxAct {
			maxAct = h
		}
		sumAct += h
	}
	fmt.Printf("  Active neurons: %d / %d (%.1f%%)\n", activeHidden, len(hidden), float64(activeHidden)/float64(len(hidden))*100)
	fmt.Printf("  Max activation: %.3f, Mean: %.3f\n", maxAct, sumAct/float64(len(hidden)))

	// Output layer raw values (before softmax, for insight)
	fmt.Println("\nOutput Layer (10 classes):")
	fmt.Print("  Pre-softmax: ")
	for i := 0; i < 10; i++ {
		// Compute pre-softmax value
		rawVal := 0.0
		for j := range hidden {
			rawVal += hidden[j] * 0.5 // Simplified
		}
		fmt.Printf("%d:%.2f ", i, output[i])
	}
	fmt.Println()
}

func activationToChar(value float64) string {
	if value < 0.1 {
		return "·"
	} else if value < 0.3 {
		return "░"
	} else if value < 0.5 {
		return "▒"
	} else if value < 0.7 {
		return "▓"
	}
	return "█"
}

func showConfusionMatrix(matrix [][]int) {
	fmt.Println("\n═══════════════════════════════════════════════════")
	fmt.Println("              Confusion Matrix")
	fmt.Println("═══════════════════════════════════════════════════")

	// Header
	fmt.Print("       Predicted\n")
	fmt.Print("       ")
	for i := 0; i < 10; i++ {
		fmt.Printf("%4d", i)
	}
	fmt.Println("  │ Total")
	fmt.Println("      +" + strings.Repeat("────", 10) + "──┼──────")

	// Matrix rows
	totalCorrect := 0
	totalSamples := 0

	for i := 0; i < 10; i++ {
		rowTotal := 0
		for j := 0; j < 10; j++ {
			rowTotal += matrix[i][j]
		}
		totalSamples += rowTotal

		if i == 5 {
			fmt.Printf("A %d │ ", i)
		} else {
			fmt.Printf("  %d │ ", i)
		}

		for j := 0; j < 10; j++ {
			val := matrix[i][j]
			if i == j {
				// Diagonal (correct predictions)
				fmt.Printf("\033[92m%4d\033[0m", val)
				totalCorrect += val
			} else if val > rowTotal/10 {
				// High error
				fmt.Printf("\033[91m%4d\033[0m", val)
			} else if val > 0 {
				// Some error
				fmt.Printf("\033[93m%4d\033[0m", val)
			} else {
				fmt.Printf("%4d", val)
			}
		}
		fmt.Printf("  │ %4d\n", rowTotal)
	}

	fmt.Println("      +" + strings.Repeat("────", 10) + "──┼──────")

	// Column totals
	fmt.Print("Total ")
	for j := 0; j < 10; j++ {
		colTotal := 0
		for i := 0; i < 10; i++ {
			colTotal += matrix[i][j]
		}
		fmt.Printf("%4d", colTotal)
	}
	fmt.Printf("  │ %4d\n", totalSamples)

	// Summary
	accuracy := float64(totalCorrect) / float64(totalSamples) * 100
	fmt.Printf("\nOverall Accuracy: %.1f%% (%d/%d correct)\n", accuracy, totalCorrect, totalSamples)

	// Legend
	fmt.Println("\nLegend: \033[92m■ Correct\033[0m  \033[93m■ Minor error\033[0m  \033[91m■ Major error\033[0m")
}

func showPerClassMetrics(precision, recall, f1 []float64) {
	fmt.Println("\n─── Per-Class Performance Metrics ───")
	fmt.Println("Class  Precision  Recall    F1-Score")
	fmt.Println("─────  ─────────  ────────  ────────")

	for i := 0; i < 10; i++ {
		pBar := strings.Repeat("█", int(precision[i]*10)) + strings.Repeat("░", 10-int(precision[i]*10))
		rBar := strings.Repeat("█", int(recall[i]*10)) + strings.Repeat("░", 10-int(recall[i]*10))

		fmt.Printf("  %d    %s %.1f%%  %s %.1f%%  %.3f\n",
			i, pBar, precision[i]*100, rBar, recall[i]*100, f1[i])
	}

	// Macro averages
	var avgP, avgR, avgF1 float64
	for i := 0; i < 10; i++ {
		avgP += precision[i]
		avgR += recall[i]
		avgF1 += f1[i]
	}
	avgP /= 10
	avgR /= 10
	avgF1 /= 10

	fmt.Println("─────────────────────────────────────")
	fmt.Printf("Macro  %-23s %.1f%%  %-11s %.1f%%  %.3f\n", "", avgP*100, "", avgR*100, avgF1)
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

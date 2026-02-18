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
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fecim-lattice-tools/module3-mnist/pkg/training"
	"fecim-lattice-tools/shared/cli"
	"fecim-lattice-tools/shared/crossbar"
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
	Samples     int     `json:"samples"`
	Accuracy    float64 `json:"accuracy"`
	FPAccuracy  float64 `json:"fp_accuracy,omitempty"`
	CIMAccuracy float64 `json:"cim_accuracy,omitempty"`
	AgreeRate   float64 `json:"agree_rate,omitempty"`
	AvgKL       float64 `json:"avg_kl,omitempty"`
	AvgEnergy   float64 `json:"avg_energy_uj,omitempty"`
	Levels      int     `json:"levels,omitempty"`
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

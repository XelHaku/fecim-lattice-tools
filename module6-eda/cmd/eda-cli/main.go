// cmd/eda-cli/main.go
// CLI tool for FeCIM array design generation and export
//
// Supports three operation modes:
//   - storage: High-density non-volatile storage (NAND replacement)
//   - memory:  High-speed zero-refresh memory (DRAM replacement)
//   - compute: Analog compute-in-memory for AI inference
//
// For storage and memory modes, no input file is required.
// For compute mode, weights are optional - omit -input for unprogrammed arrays.
//
// Common flags:
//   - --json: Output results as JSON
//   - --quiet: Suppress informational output
//   - --config: Load configuration from YAML/JSON file
package edacli

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/module6-eda/pkg/export"
	"fecim-lattice-tools/shared/cli"
	"fecim-lattice-tools/shared/logging"
)

// WeightsFile represents a JSON file containing neural network weights
// Only used in compute mode when pre-programming with trained weights
type WeightsFile struct {
	Name    string      `json:"name"`
	Rows    int         `json:"rows"`
	Cols    int         `json:"cols"`
	Weights [][]float64 `json:"weights"`
}

// EDAConfig holds configuration for the EDA CLI.
type EDAConfig struct {
	Mode       string  `json:"mode" yaml:"mode"`
	Rows       int     `json:"rows" yaml:"rows"`
	Cols       int     `json:"cols" yaml:"cols"`
	Levels     int     `json:"levels" yaml:"levels"`
	Technology string  `json:"technology" yaml:"technology"`
	VDD        float64 `json:"vdd" yaml:"vdd"`
}

// EDAResult represents design generation results for JSON output.
type EDAResult struct {
	DesignName     string  `json:"design_name"`
	Mode           string  `json:"mode"`
	Rows           int     `json:"rows"`
	Cols           int     `json:"cols"`
	TotalCells     int     `json:"total_cells"`
	ActiveCells    int     `json:"active_cells"`
	AreaMM2        float64 `json:"area_mm2"`
	PowerMW        float64 `json:"power_mw"`
	ThroughputGOPS float64 `json:"throughput_gops,omitempty"`
	Technology     string  `json:"technology"`
	OutputFiles    []string `json:"output_files"`
}

func Run(args []string) error {
	homeDir, _ := os.UserHomeDir()
	logPath := filepath.Join(homeDir, ".fecim", "logs", "module6-eda-cli.log")
	if err := logging.Init("module6-eda-cli", logPath); err != nil {
		logging.GlobalError("Failed to initialize logging: %v\n", err)
		return err
	}
	defer logging.CloseGlobal()

	// Enable logging by default
	logging.SetVerbosity(logging.VerbosityInfo)

	fs := flag.NewFlagSet("eda-cli", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	// Common CLI flags (use explicit names to avoid conflict with export-json)
	jsonOutput := fs.Bool("json-output", false, "Output results as JSON to stdout")
	quiet := fs.Bool("quiet", false, "Suppress informational output")
	configFile := fs.String("config", "", "Load configuration from YAML/JSON file")

	// Operation mode
	mode := fs.String("mode", "compute", "Operation mode: storage, memory, or compute")

	// Input (optional for compute mode, ignored for others)
	inputFile := fs.String("input", "", "Input weights JSON file (optional, compute mode only)")

	// Output
	outputDir := fs.String("output", "data", "Output directory")
	designName := fs.String("name", "fecim_array", "Design name for output files")

	// Array parameters
	rows := fs.Int("rows", 128, "Array rows")
	cols := fs.Int("cols", 128, "Array cols")
	levels := fs.Int("levels", 30, "Conductance levels (2-30)")

	// Technology selection
	tech := fs.String("tech", "SKY130", "Technology: SKY130, GF180MCU, IHP_SG13G2")
	arch := fs.String("arch", "passive", "Architecture: passive or 1T1R")

	// Electrical parameters
	vdd := fs.Float64("vdd", 1.8, "Supply voltage (V)")
	gmin := fs.Float64("gmin", 10.0, "Min conductance (μS)")
	gmax := fs.Float64("gmax", 100.0, "Max conductance (μS)")

	// Export options
	exportJSONFile := fs.Bool("export-json", true, "Export JSON mapping file")
	exportCSV := fs.Bool("csv", true, "Export CSV cell assignments")
	exportSPICE := fs.Bool("spice", true, "Export SPICE netlist")
	exportVerilog := fs.Bool("verilog", true, "Export Verilog netlist")
	exportDEF := fs.Bool("def", true, "Export DEF placement")
	help := fs.Bool("help", false, "Show help")

	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintln(out, "FeCIM EDA CLI")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out, "  fecim-lattice-tools eda cli [options]")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Options:")
		fs.PrintDefaults()
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Common Options:")
		fmt.Fprintln(out, "  --json-output     Output results as JSON to stdout")
		fmt.Fprintln(out, "  --quiet           Suppress informational output")
		fmt.Fprintln(out, "  --config FILE     Load configuration from YAML/JSON file")
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(fs.Output(), "Error:", err)
		fs.Usage()
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if *help {
		fs.Usage()
		return nil
	}

	// Load config file if specified
	var cfg EDAConfig
	if *configFile != "" {
		loader := cli.NewConfigLoader(*configFile)
		if err := loader.Load(&cfg); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if cfg.Mode != "" && *mode == "compute" {
			*mode = cfg.Mode
		}
		if cfg.Rows > 0 && *rows == 128 {
			*rows = cfg.Rows
		}
		if cfg.Cols > 0 && *cols == 128 {
			*cols = cfg.Cols
		}
		if cfg.Levels > 0 && *levels == 30 {
			*levels = cfg.Levels
		}
		if cfg.Technology != "" && *tech == "SKY130" {
			*tech = cfg.Technology
		}
		if cfg.VDD > 0 && *vdd == 1.8 {
			*vdd = cfg.VDD
		}
	}

	// Parse operation mode
	var opMode compiler.OperationMode
	switch strings.ToLower(*mode) {
	case "storage":
		opMode = compiler.ModeStorage
	case "memory":
		opMode = compiler.ModeMemory
	case "compute":
		opMode = compiler.ModeCompute
	default:
		logging.Printf("Error: unknown mode '%s'. Use: storage, memory, or compute\n", *mode)
		return fmt.Errorf("unknown mode %q", *mode)
	}

	logging.Printf("FeCIM Array Generator - %s Mode\n", strings.Title(*mode))
	logging.Printf("========================================\n\n")

	// Create configuration
	config := compiler.NewArrayConfig(opMode, *rows, *cols)
	config.Name = *designName
	config.Technology = *tech
	config.Levels = *levels
	config.GMin = *gmin
	config.GMax = *gmax
	config.Peripherals.VDD = *vdd

	// Handle architecture
	if strings.ToLower(*arch) == "1t1r" {
		config.With1T1R()
	}

	// Load weights for compute mode (if provided)
	if opMode == compiler.ModeCompute && *inputFile != "" {
		data, err := os.ReadFile(*inputFile)
		if err != nil {
			logging.Printf("Error reading weights file: %v\n", err)
			return err
		}

		var wf WeightsFile
		if err := json.Unmarshal(data, &wf); err != nil {
			logging.Printf("Error parsing weights JSON: %v\n", err)
			return err
		}

		logging.Printf("Loaded weights: %s (%dx%d = %d weights)\n",
			wf.Name, len(wf.Weights), len(wf.Weights[0]),
			len(wf.Weights)*len(wf.Weights[0]))

		config.ComputeConfig.InitialWeights = wf.Weights
	}

	// Print configuration
	logging.Printf("Configuration:\n")
	logging.Printf("  Mode:         %s\n", config.Mode)
	logging.Printf("  Array Size:   %d × %d (%d cells)\n", config.ArrayRows, config.ArrayCols, config.ArrayRows*config.ArrayCols)
	logging.Printf("  Technology:   %s\n", config.Technology)
	logging.Printf("  Architecture: %s\n", config.Architecture)
	logging.Printf("  Levels:       %d (%.2f bits/cell)\n", config.Levels, math.Log2(float64(config.Levels)))
	logging.Printf("  Conductance:  %.1f - %.1f μS\n", config.GMin, config.GMax)
	if opMode == compiler.ModeCompute && config.ComputeConfig.InitialWeights != nil {
		logging.Printf("  Weights:      %dx%d loaded\n",
			len(config.ComputeConfig.InitialWeights),
			len(config.ComputeConfig.InitialWeights[0]))
	} else if opMode == compiler.ModeCompute {
		logging.Printf("  Weights:      None (unprogrammed array)\n")
	}
	logging.Println()

	// Generate design
	design, err := compiler.GenerateDesign(config)
	if err != nil {
		logging.Printf("Design generation error: %v\n", err)
		return err
	}

	// Print results
	logging.Printf("Design Statistics:\n")
	logging.Printf("  Total Cells:  %d\n", design.Stats.TotalCells)
	logging.Printf("  Active Cells: %d\n", design.Stats.ActiveCells)
	logging.Printf("  Area:         %.4f mm²\n", design.Stats.AreaMM2)
	logging.Printf("  Est. Power:   %.2f mW\n", design.Stats.PowerMW)

	if opMode == compiler.ModeCompute {
		logging.Printf("  Throughput:   %.2f GOPS\n", design.Stats.ThroughputGOPS)
		if config.ComputeConfig.InitialWeights != nil {
			logging.Printf("  Weight Range: [%.4f, %.4f]\n", design.Stats.WeightMin, design.Stats.WeightMax)
			logging.Printf("  Quant PSNR:   %.2f dB\n", design.Stats.QuantPSNR)
		}
	}
	logging.Println()

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		logging.Printf("Error creating output directory: %v\n", err)
		return err
	}

	// Track output files for JSON result
	outputFiles := make([]string, 0)

	// Export files
	if !*quiet {
		logging.Printf("Exporting files to %s/\n", *outputDir)
	}

	if *exportJSONFile {
		path := filepath.Join(*outputDir, *designName+"_design.json")
		if err := export.ExportJSON(design, path); err != nil {
			logging.Printf("  JSON export error: %v\n", err)
		} else {
			if !*quiet {
				logging.Printf("  ✓ %s\n", path)
			}
			outputFiles = append(outputFiles, path)
		}
	}

	if *exportCSV {
		path := filepath.Join(*outputDir, *designName+"_cells.csv")
		if err := export.ExportCSV(design, path); err != nil {
			logging.Printf("  CSV export error: %v\n", err)
		} else {
			if !*quiet {
				logging.Printf("  ✓ %s\n", path)
			}
			outputFiles = append(outputFiles, path)
		}
	}

	if *exportSPICE {
		path := filepath.Join(*outputDir, *designName+".sp")
		if err := export.ExportSPICE(design, path, *vdd); err != nil {
			logging.Printf("  SPICE export error: %v\n", err)
		} else {
			if !*quiet {
				logging.Printf("  ✓ %s\n", path)
			}
			outputFiles = append(outputFiles, path)
		}
	}

	if *exportVerilog {
		path := filepath.Join(*outputDir, *designName+".v")
		if err := export.ExportVerilog(design, path); err != nil {
			logging.Printf("  Verilog export error: %v\n", err)
		} else {
			if !*quiet {
				logging.Printf("  ✓ %s\n", path)
			}
			outputFiles = append(outputFiles, path)
		}
	}

	if *exportDEF {
		path := filepath.Join(*outputDir, *designName+".def")
		if err := export.ExportDEF(design, path); err != nil {
			logging.Printf("  DEF export error: %v\n", err)
		} else {
			if !*quiet {
				logging.Printf("  ✓ %s\n", path)
			}
			outputFiles = append(outputFiles, path)
		}
	}

	// JSON output mode
	if *jsonOutput {
		result := EDAResult{
			DesignName:     *designName,
			Mode:           *mode,
			Rows:           config.ArrayRows,
			Cols:           config.ArrayCols,
			TotalCells:     design.Stats.TotalCells,
			ActiveCells:    design.Stats.ActiveCells,
			AreaMM2:        design.Stats.AreaMM2,
			PowerMW:        design.Stats.PowerMW,
			ThroughputGOPS: design.Stats.ThroughputGOPS,
			Technology:     *tech,
			OutputFiles:    outputFiles,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	if !*quiet {
		logging.Println("\nDone!")
	}
	return nil
}

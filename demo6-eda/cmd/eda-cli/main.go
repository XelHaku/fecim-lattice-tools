// cmd/eda-cli/main.go
// CLI tool for headless compilation and export
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"demo6-eda/pkg/compiler"
	"demo6-eda/pkg/export"
)

type WeightsFile struct {
	Name    string      `json:"name"`
	Rows    int         `json:"rows"`
	Cols    int         `json:"cols"`
	Weights [][]float64 `json:"weights"`
}

func main() {
	// Flags
	inputFile := flag.String("input", "", "Input weights JSON file (required)")
	outputDir := flag.String("output", ".", "Output directory")
	levels := flag.Int("levels", 30, "Quantization levels (2-30)")
	rows := flag.Int("rows", 128, "Array rows")
	cols := flag.Int("cols", 128, "Array cols")
	vdd := flag.Float64("vdd", 1.8, "VDD for SPICE export")
	exportJSON := flag.Bool("json", true, "Export JSON mapping")
	exportCSV := flag.Bool("csv", true, "Export CSV")
	exportSPICE := flag.Bool("spice", true, "Export SPICE netlist")

	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: eda-cli -input <weights.json> [options]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load weights
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var wf WeightsFile
	if err := json.Unmarshal(data, &wf); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded: %s (%dx%d = %d weights)\n", wf.Name, len(wf.Weights), len(wf.Weights[0]), len(wf.Weights)*len(wf.Weights[0]))

	// Configure
	config := compiler.DefaultConfig()
	config.ArrayRows = *rows
	config.ArrayCols = *cols
	config.Levels = *levels

	// Compile
	mapping, err := compiler.Compile(wf.Weights, config)
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nCompilation Results:\n")
	fmt.Printf("  Cells: %d / %d (%.1f%%)\n", mapping.Stats.UsedCells, mapping.Stats.TotalCells, mapping.Stats.Utilization*100)
	fmt.Printf("  Unique levels: %d / %d\n", mapping.Stats.UniqueLevels, config.Levels)
	fmt.Printf("  Weight range: [%.4f, %.4f]\n", mapping.Stats.WeightMin, mapping.Stats.WeightMax)
	fmt.Printf("  Quant PSNR: %.2f dB\n", mapping.Stats.QuantPSNR)

	// Export
	os.MkdirAll(*outputDir, 0755)

	if *exportJSON {
		path := filepath.Join(*outputDir, "crossbar_mapping.json")
		if err := export.ExportJSON(mapping, path); err != nil {
			fmt.Printf("JSON export error: %v\n", err)
		} else {
			fmt.Printf("\nExported: %s\n", path)
		}
	}

	if *exportCSV {
		path := filepath.Join(*outputDir, "cell_assignments.csv")
		if err := export.ExportCSV(mapping, path); err != nil {
			fmt.Printf("CSV export error: %v\n", err)
		} else {
			fmt.Printf("Exported: %s\n", path)
		}
	}

	if *exportSPICE {
		path := filepath.Join(*outputDir, "crossbar.sp")
		if err := export.ExportSPICE(mapping, path, *vdd); err != nil {
			fmt.Printf("SPICE export error: %v\n", err)
		} else {
			fmt.Printf("Exported: %s\n", path)
		}
	}

	fmt.Println("\nDone!")
}

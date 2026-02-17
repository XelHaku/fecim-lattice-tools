//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func main() {
	fmt.Println("=== Module 2 Crossbar Logging Demo ===\n")

	// Create array
	cfg := &crossbar.Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0.05,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create crossbar array: %v\n", err)
		fmt.Fprintln(os.Stderr, "check that the crossbar configuration is valid")
		os.Exit(1)
	}

	// Program some weights
	fmt.Println("Programming weights...")
	arr.ProgramWeight(0, 0, 0.8)
	arr.ProgramWeight(1, 1, 0.6)
	arr.ProgramWeight(2, 2, 0.4)

	// Perform MVM
	fmt.Println("Performing MVM...")
	input := []float64{1.0, 0.8, 0.6, 0.4}
	output, err := arr.MVM(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MVM failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Output: %v\n\n", output)

	// Analyze IR drop
	fmt.Println("Analyzing IR drop...")
	irAnalysis := arr.AnalyzeIRDrop(input, nil)
	fmt.Printf("Max IR drop: %.6f V\n", irAnalysis.MaxIRDrop)
	fmt.Printf("Avg IR drop: %.6f V\n\n", irAnalysis.AvgIRDrop)

	// Analyze sneak paths
	fmt.Println("Analyzing sneak paths...")
	sneakAnalysis := arr.AnalyzeSneakPaths(0, 0)
	fmt.Printf("Max sneak ratio: %.4f\n", sneakAnalysis.MaxSneakRatio)
	fmt.Printf("Avg sneak ratio: %.4f\n\n", sneakAnalysis.AvgSneakRatio)

	fmt.Println("=== Demo Complete ===")
	fmt.Println("Check logs/ directory for detailed TRACE logs")
}

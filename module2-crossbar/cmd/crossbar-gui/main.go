// Demo 2 GUI: Crossbar Array Visualization with Fyne
//
// This provides an interactive GUI for visualizing matrix-vector multiplication
// operations on a simulated ferroelectric crossbar array.
//
// Features:
// - Interactive heatmap visualization of conductance states
// - IR drop analysis with heatmap overlay
// - Sneak path current analysis
// - Real-time MVM operations with full physics
// - 30 discrete FeCIM levels (4.9 bits/cell)
//
// Standard Mode:
//   go run ./cmd/crossbar-gui
//
// Enhanced Mode (all features):
//   go run ./cmd/crossbar-gui -enhanced
//
// Enhanced features include:
// - Color legends for all heatmaps
// - Live metrics panel (accuracy, energy, performance)
// - Before/after comparison view
// - Accuracy waterfall chart
// - Energy comparison badges
// - Enhanced MVM with integrated non-idealities
// - Data export (CSV, JSON)
package main

import (
	"flag"
	"fmt"

	"multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/gui"
)

func main() {
	enhanced := flag.Bool("enhanced", false, "Enable enhanced UI with all features")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help {
		fmt.Println("FeCIM Crossbar Array Visualization")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  go run ./cmd/crossbar-gui [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -enhanced    Enable enhanced UI with all features")
		fmt.Println("  -help        Show this help message")
		fmt.Println()
		fmt.Println("Features:")
		fmt.Println("  • 64×64 crossbar array (configurable 8-128)")
		fmt.Println("  • 30 discrete FeCIM levels (4.9 bits/cell)")
		fmt.Println("  • Matrix-vector multiplication in O(1)")
		fmt.Println("  • IR drop analysis")
		fmt.Println("  • Sneak path analysis")
		fmt.Println("  • Device variation simulation")
		fmt.Println()
		fmt.Println("Enhanced Features:")
		fmt.Println("  • Color legends with level indicators")
		fmt.Println("  • Live metrics (accuracy, energy, performance)")
		fmt.Println("  • Before/after comparison view")
		fmt.Println("  • Accuracy degradation waterfall")
		fmt.Println("  • Energy comparison with GPU")
		fmt.Println("  • Differential array for signed weights")
		fmt.Println("  • Write-verify programming simulation")
		fmt.Println("  • Data export (CSV/JSON)")
		return
	}

	app := gui.NewCrossbarApp()

	if *enhanced {
		fmt.Println("Starting FeCIM Crossbar Visualizer (Enhanced Mode)")
		fmt.Println("→ All features enabled")
		app.RunEnhanced()
	} else {
		fmt.Println("Starting FeCIM Crossbar Visualizer (Standard Mode)")
		fmt.Println("→ Run with -enhanced for all features")
		app.Run()
	}
}

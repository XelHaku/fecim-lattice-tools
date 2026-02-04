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
// - 30 discrete FeCIM levels (4.9 bits/cell, conference claim baseline)
//
// Standard Mode:
//
//	go run ./cmd/fecim-lattice-tools crossbar
//
// Enhanced Mode (all features):
//
//	go run ./cmd/fecim-lattice-tools crossbar -enhanced
//
// Terminal Inference (CLI):
//
//	go run ./cmd/fecim-lattice-tools crossbar inference [options]
//
// Enhanced features include:
// - Color legends for all heatmaps
// - Live metrics panel (accuracy, energy, performance)
// - Before/after comparison view
// - Accuracy waterfall chart
// - Energy comparison badges
// - Enhanced MVM with integrated non-idealities
// - Data export (CSV, JSON)
package crossbarcmd

import (
	"flag"
	"fmt"
	"os"

	"fecim-lattice-tools/module2-crossbar/pkg/gui"
)

func RunGUI(args []string) error {
	fs := flag.NewFlagSet("crossbar-gui", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	enhanced := fs.Bool("enhanced", false, "Enable enhanced UI with all features")
	help := fs.Bool("help", false, "Show help")
	helpShort := fs.Bool("h", false, "Show help (shorthand)")

	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintln(out, "FeCIM Crossbar Array Visualization")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out, "  go run ./cmd/fecim-lattice-tools crossbar [options]")
		fmt.Fprintln(out, "  go run ./cmd/fecim-lattice-tools crossbar inference [options]")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Options:")
		fmt.Fprintln(out, "  -enhanced    Enable enhanced UI with all features")
		fmt.Fprintln(out, "  -help        Show this help message")
		fmt.Fprintln(out, "  -h           Show this help message (shorthand)")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Inference:")
		fmt.Fprintln(out, "  Use: go run ./cmd/fecim-lattice-tools crossbar inference -help")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Features:")
		fmt.Fprintln(out, "  • 64×64 crossbar array (configurable 8-128)")
		fmt.Fprintln(out, "  • 30 discrete FeCIM levels (4.9 bits/cell, conference claim)")
		fmt.Fprintln(out, "  • Matrix-vector multiplication in O(1)")
		fmt.Fprintln(out, "  • IR drop analysis")
		fmt.Fprintln(out, "  • Sneak path analysis")
		fmt.Fprintln(out, "  • Device variation simulation")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Enhanced Features:")
		fmt.Fprintln(out, "  • Color legends with level indicators")
		fmt.Fprintln(out, "  • Live metrics (accuracy, energy, performance)")
		fmt.Fprintln(out, "  • Before/after comparison view")
		fmt.Fprintln(out, "  • Accuracy degradation waterfall")
		fmt.Fprintln(out, "  • Energy comparison with GPU")
		fmt.Fprintln(out, "  • Differential array for signed weights")
		fmt.Fprintln(out, "  • Write-verify programming simulation")
		fmt.Fprintln(out, "  • Data export (CSV/JSON)")
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

	app, err := gui.NewCrossbarApp()
	if err != nil {
		return fmt.Errorf("failed to initialize crossbar app: %w", err)
	}

	if *enhanced {
		fmt.Println("Starting FeCIM Crossbar Visualizer (Enhanced Mode)")
		fmt.Println("→ All features enabled")
		app.RunEnhanced()
	} else {
		fmt.Println("Starting FeCIM Crossbar Visualizer (Standard Mode)")
		fmt.Println("→ Run with -enhanced for all features")
		app.Run()
	}

	return nil
}

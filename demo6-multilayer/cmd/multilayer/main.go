package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"ironlattice-vis/demo6-multilayer/pkg/multilayer"
)

func main() {
	// Flags
	showAll := flag.Bool("all", false, "Show all visualizations")
	show3D := flag.Bool("3d", false, "Show 3D stack view")
	showExploded := flag.Bool("exploded", false, "Show exploded view")
	showDataFlow := flag.Bool("dataflow", false, "Show data flow visualization")
	showMetrics := flag.Bool("metrics", false, "Show stack metrics")
	showVias := flag.Bool("vias", false, "Show via network details")
	showEnergy := flag.Bool("energy", false, "Show energy comparison")
	useMNIST := flag.Bool("mnist", false, "Use MNIST stack (default: small demo stack)")
	noColor := flag.Bool("no-color", false, "Disable color output")

	flag.Parse()

	// Default to all if nothing specified
	if !*show3D && !*showExploded && !*showDataFlow && !*showMetrics && !*showVias && !*showEnergy {
		*showAll = true
	}

	// Create stack
	var stack *multilayer.Stack
	if *useMNIST {
		stack = multilayer.MNISTStack()
	} else {
		stack = multilayer.SmallStack()
	}

	// Create renderer
	renderer := multilayer.DefaultRenderer()
	renderer.UseColor = !*noColor

	// Header
	printHeader(stack)

	// Show visualizations
	if *showAll || *show3D {
		fmt.Println(renderer.Render3DView(stack))
	}

	if *showAll || *showExploded {
		fmt.Println(renderer.RenderExplodedView(stack))
	}

	if *showAll || *showDataFlow {
		// Create sample input
		inputSize := stack.Layers[0].Rows
		input := make([]float64, inputSize)
		for i := range input {
			input[i] = float64(i%10) / 10.0
		}
		fmt.Println(renderer.RenderDataFlow(stack, input))
	}

	if *showAll || *showMetrics {
		fmt.Println(renderer.RenderMetrics(stack))
	}

	if *showAll || *showVias {
		printViaDetails(stack)
	}

	if *showAll || *showEnergy {
		printEnergyComparison(stack)
	}

	// Summary
	printSummary(stack)
}

func printHeader(stack *multilayer.Stack) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           IRONLATTICE 3D MULTI-LAYER ARCHITECTURE            ║")
	fmt.Println("║              Ferroelectric Compute-in-Memory                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Stack: %s\n", stack.Name)
	fmt.Printf("Technology: %s\n", stack.Technology)
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()
}

func printViaDetails(stack *multilayer.Stack) {
	viaNet := multilayer.NewViaNetwork(stack)
	stats := viaNet.GetStats(stack.FootprintArea())

	fmt.Println("Via Network Details:")
	fmt.Println("══════════════════════════════════════════")
	fmt.Println()

	// Via array details
	for i, array := range viaNet.Arrays {
		fmt.Printf("Via Array %d (Layer %d → %d):\n", i+1, array.FromLayer+1, array.ToLayer+1)
		fmt.Printf("  Count:        %d vias\n", array.Count)
		fmt.Printf("  Pitch:        %.0f nm\n", array.Pitch)
		fmt.Printf("  Diameter:     %.0f nm\n", array.Diameter)
		fmt.Printf("  Aspect Ratio: %.2f\n", array.AspectRatio)
		fmt.Println()
	}

	// Network summary
	fmt.Println("Network Summary:")
	fmt.Printf("  Total Vias:       %d\n", stats.TotalVias)
	fmt.Printf("  Total Length:     %.2f µm\n", stats.TotalLength)
	fmt.Printf("  Avg Resistance:   %.1f Ω per via\n", stats.AvgResistance)
	fmt.Printf("  Total Capacitance: %.2f fF\n", stats.TotalCapacitance)
	fmt.Printf("  Propagation Delay: %.3f ps\n", stats.PropagationDelay)
	fmt.Printf("  Via Density:      %.2f vias/µm²\n", stats.ViaDensity)
	fmt.Println()

	// Yield estimation
	fmt.Println("Manufacturing Yield Estimates:")
	defectDensities := []float64{0.01, 0.1, 1.0, 10.0, 100.0}
	for _, density := range defectDensities {
		yield := viaNet.EstimateViaYield(density)
		fmt.Printf("  Defect density %.2f/cm²: %.4f (%.2f%%) yield\n",
			density, yield, yield*100)
	}
	fmt.Println()
}

func printEnergyComparison(stack *multilayer.Stack) {
	fmt.Println("Energy Comparison (per inference):")
	fmt.Println("══════════════════════════════════════════")
	fmt.Println()

	estimates := stack.EstimateEnergy()

	// Table header
	fmt.Printf("%-12s %12s %12s %12s %12s\n",
		"Layer", "MAC Energy", "Data Move", "Total", "vs Traditional")
	fmt.Println(strings.Repeat("-", 64))

	totalCIM := 0.0
	totalTraditional := 0.0

	for _, est := range estimates {
		traditional := est.TotalEnergy * est.TraditionalComp
		fmt.Printf("%-12s %10.3f pJ %10.3f pJ %10.3f pJ %10.0fx better\n",
			est.LayerName, est.MACEnergy, est.DataMoveEnergy, est.TotalEnergy, est.TraditionalComp)
		totalCIM += est.TotalEnergy
		totalTraditional += traditional
	}

	fmt.Println(strings.Repeat("-", 64))
	fmt.Printf("%-12s %10s %13s %10.3f pJ %10.0fx better\n",
		"TOTAL", "", "", totalCIM, totalTraditional/totalCIM)
	fmt.Println()

	// Visual comparison
	fmt.Println("Energy Comparison Visualization:")
	fmt.Println()

	maxWidth := 50
	cimWidth := 1 // Minimum bar
	tradWidth := maxWidth

	fmt.Printf("IronLattice:  [%s] %.3f pJ\n",
		strings.Repeat("█", cimWidth)+strings.Repeat(" ", maxWidth-cimWidth),
		totalCIM)
	fmt.Printf("Traditional:  [%s] %.1f pJ\n",
		strings.Repeat("█", tradWidth),
		totalTraditional)
	fmt.Println()
	fmt.Printf("IronLattice achieves %.0fx energy reduction!\n", totalTraditional/totalCIM)
	fmt.Println()

	// Data flow advantage
	fmt.Println("Data Movement Analysis:")
	fmt.Println()
	dataStats := stack.AnalyzeDataFlow()
	for _, stat := range dataStats {
		fmt.Printf("  %s: %.1fx less data movement\n", stat.LayerName, stat.CIMAdvantage)
	}
	fmt.Println()
}

func printSummary(stack *multilayer.Stack) {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                         SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Key metrics
	fmt.Printf("  Layers:           %d\n", len(stack.Layers))
	fmt.Printf("  Total Cells:      %d\n", stack.TotalCells())
	fmt.Printf("  Total Parameters: %d\n", stack.TotalParameters())
	fmt.Printf("  Bits per Cell:    %.2f (30 levels)\n", stack.BitsPerCell())
	fmt.Printf("  Total Storage:    %.0f bits (%.2f KB)\n",
		stack.TotalBits(), stack.TotalBits()/8/1024)
	fmt.Println()

	// Physical
	fmt.Printf("  Stack Height:     %.0f nm (%.3f µm)\n",
		stack.StackHeight(), stack.StackHeight()/1000)
	fmt.Printf("  Footprint:        %.2f µm²\n", stack.FootprintArea())
	fmt.Printf("  Areal Density:    %.2f bits/µm²\n", stack.ArealDensity())
	fmt.Printf("  Volume Density:   %.2f bits/µm³\n", stack.VolumetricDensity())
	fmt.Println()

	// Via network
	viaNet := multilayer.NewViaNetwork(stack)
	fmt.Printf("  Total Vias:       %d\n", viaNet.TotalVias)
	fmt.Printf("  Via Length:       %.2f µm\n", viaNet.TotalLength)
	fmt.Println()

	// Utilization
	fmt.Println("Layer Utilization:")
	util := stack.LayerUtilization()
	for i, u := range util {
		barWidth := int(u * 30)
		bar := strings.Repeat("█", barWidth) + strings.Repeat("░", 30-barWidth)
		fmt.Printf("  Layer %d: [%s] %.1f%%\n", i+1, bar, u*100)
	}
	fmt.Println()

	// Quote from Dr. Tour
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println("  \"Compute in memory where the same device does the memory")
	fmt.Println("   and the computation.\" - Dr. external research group")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println()

	os.Exit(0)
}

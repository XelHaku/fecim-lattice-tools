package main

import (
	"flag"
	"fmt"
	"strings"

	"ironlattice-vis/demo7-nonidealities/pkg/nonidealities"
)

func main() {
	// Flags
	showAll := flag.Bool("all", false, "Show all analyses")
	showIRDrop := flag.Bool("ir", false, "Show IR drop analysis")
	showSneak := flag.Bool("sneak", false, "Show sneak path analysis")
	showDrift := flag.Bool("drift", false, "Show drift analysis")
	showCompare := flag.Bool("compare", false, "Show technology comparison")
	arraySize := flag.Int("size", 16, "Array size (NxN)")
	noColor := flag.Bool("no-color", false, "Disable color output")

	flag.Parse()

	// Default to all if nothing specified
	if !*showIRDrop && !*showSneak && !*showDrift && !*showCompare {
		*showAll = true
	}

	renderer := nonidealities.NewRenderer()
	renderer.UseColor = !*noColor

	// Header
	printHeader()

	if *showAll || *showIRDrop {
		analyzeIRDrop(renderer, *arraySize)
	}

	if *showAll || *showSneak {
		analyzeSneakPaths(renderer, *arraySize)
	}

	if *showAll || *showDrift {
		analyzeDrift(renderer, *arraySize)
	}

	if *showAll || *showCompare {
		compareTechnologies(renderer, *arraySize)
	}

	// Summary
	printSummary(*arraySize)
}

func printHeader() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           IRONLATTICE NON-IDEALITIES ANALYSIS                ║")
	fmt.Println("║        IR Drop • Sneak Paths • Conductance Drift             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func analyzeIRDrop(renderer *nonidealities.Renderer, size int) {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                    IR DROP ANALYSIS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Create simulator
	ir := nonidealities.NewIRDropSimulator(size, size)

	// Set realistic input pattern (varying voltages)
	for i := 0; i < size; i++ {
		voltage := 0.3 + 0.2*float64(i%5)/4.0 // 0.3V to 0.5V
		ir.SetInputVoltage(i, voltage)
	}

	// Set varying conductances (simulating weight distribution)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			// Weight pattern: higher near center
			distFromCenter := float64((i-size/2)*(i-size/2) + (j-size/2)*(j-size/2))
			g := 50e-6 + 30e-6*distFromCenter/float64(size*size/2)
			ir.SetConductance(i, j, g)
		}
	}

	// Run simulation
	ir.Simulate(100)

	// Get statistics
	statsBefore := ir.GetStats()

	// Render IR drop map
	fmt.Println(renderer.RenderIRDropMap(ir))
	fmt.Println(renderer.RenderIRDropStats(statsBefore))

	// Show mitigation effect
	fmt.Println("\nApplying Mitigation: 2x Wider Metal Lines")
	fmt.Println(strings.Repeat("─", 60))

	mitigation := nonidealities.IRDropMitigation{
		UseWidenedLines:   true,
		LineWidthIncrease: 2.0,
	}
	ir.ApplyMitigation(mitigation)

	statsAfter := ir.GetStats()
	fmt.Println(renderer.RenderMitigationComparison(statsBefore, statsAfter, "2x Wider Lines"))
}

func analyzeSneakPaths(renderer *nonidealities.Renderer, size int) {
	fmt.Println("\n═══════════════════════════════════════════════════════════════")
	fmt.Println("                  SNEAK PATH ANALYSIS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Create analyzer
	sp := nonidealities.NewSneakPathAnalyzer(size, size)

	// Set conductances with some variation
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			// Random-ish conductance between 10-90 µS
			g := (10 + float64((i*7+j*11)%80)) * 1e-6
			sp.SetConductance(i, j, g)
		}
	}

	// Analyze center cell
	targetRow := size / 2
	targetCol := size / 2
	voltage := 0.5

	sp.AnalyzeTarget(targetRow, targetCol, voltage)
	statsBefore := sp.GetStats(voltage)

	fmt.Println(renderer.RenderSneakPathMap(sp))
	fmt.Println(renderer.RenderSneakStats(statsBefore))

	// Show top sneak paths
	fmt.Println("Top 5 Sneak Paths:")
	fmt.Println(strings.Repeat("─", 60))
	topPaths := sp.GetTopSneakPaths(5)
	for i, path := range topPaths {
		fmt.Printf("  %d. Path through cells: ", i+1)
		for _, cell := range path.PathCells {
			fmt.Printf("(%d,%d) ", cell[0], cell[1])
		}
		fmt.Printf("| Current: %.3f nA\n", path.PathCurrent*1e9)
	}

	// Show mitigation effect
	fmt.Println("\nApplying Mitigation: Selector Device (1000:1 on/off)")
	fmt.Println(strings.Repeat("─", 60))

	mitigation := nonidealities.SneakMitigation{
		UseSelector:   true,
		SelectorOnOff: 1000,
	}
	statsAfter := sp.AnalyzeWithMitigation(targetRow, targetCol, voltage, mitigation)

	fmt.Printf("Before: SNR = %.1f dB, Sneak Ratio = %.2f%%\n",
		statsBefore.SignalToNoiseRatio, statsBefore.SneakRatio*100)
	fmt.Printf("After:  SNR = %.1f dB, Sneak Ratio = %.4f%%\n",
		statsAfter.SignalToNoiseRatio, statsAfter.SneakRatio*100)
	fmt.Printf("Improvement: SNR increased by %.1f dB!\n",
		statsAfter.SignalToNoiseRatio-statsBefore.SignalToNoiseRatio)
	fmt.Println()
}

func analyzeDrift(renderer *nonidealities.Renderer, size int) {
	fmt.Println("\n═══════════════════════════════════════════════════════════════")
	fmt.Println("                CONDUCTANCE DRIFT ANALYSIS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Create drift simulator for IronLattice (FeFET)
	d := nonidealities.NewDriftSimulator(size, size, 30)

	// Set some initial weight pattern
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			level := (i*3 + j*5) % 30
			d.SetConductanceLevel(i, j, level)
		}
	}

	// Simulate 10000 seconds (about 2.8 hours) with snapshots
	fmt.Println("Simulating conductance drift over time...")
	fmt.Printf("Technology: IronLattice (FeFET), Drift Coefficient: %.4f\n\n", d.DriftCoeff)

	numSteps := 50
	dt := 200.0 // 200 seconds per step = 10000 seconds total
	for step := 0; step < numSteps; step++ {
		d.SimulateTimeStep(dt)
		d.RecordSnapshot()
	}

	// Render history
	fmt.Println(renderer.RenderDriftHistory(d))

	// Get final stats
	stats := d.GetStats()
	fmt.Println(renderer.RenderDriftStats(stats))
}

func compareTechnologies(renderer *nonidealities.Renderer, size int) {
	fmt.Println("\n═══════════════════════════════════════════════════════════════")
	fmt.Println("                TECHNOLOGY COMPARISON")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Run comparison
	results := nonidealities.CompareTechnologies(size, size, 86400) // 1 day

	fmt.Println(renderer.RenderTechComparison(results))

	// Detailed comparison table
	fmt.Println("Detailed Metrics (after 24 hours):")
	fmt.Println(strings.Repeat("─", 70))
	fmt.Printf("%-20s %12s %12s %12s %12s\n",
		"Technology", "Drift Coeff", "Avg Drift", "Level Errors", "Retention")
	fmt.Println(strings.Repeat("─", 70))

	order := []string{"IronLattice (FeFET)", "Flash", "RRAM", "PCM"}
	for _, name := range order {
		stats := results[name]
		fmt.Printf("%-20s %12.4f %11.4f%% %12d %11.2f%%\n",
			name, stats.TechnologyComparison.FeFETDrift,
			stats.AvgDriftPercent, stats.NumLevelErrors, stats.RetentionPrediction)
	}
	fmt.Println()
}

func printSummary(size int) {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                         SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Printf("Array Size Analyzed: %dx%d = %d cells\n", size, size, size*size)
	fmt.Println()

	fmt.Println("Key Findings:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println()

	fmt.Println("1. IR DROP")
	fmt.Println("   • Worst case in far corner from drivers")
	fmt.Println("   • Mitigated by wider metal lines (2x width → ~50% reduction)")
	fmt.Println("   • IronLattice low voltages minimize IR drop effects")
	fmt.Println()

	fmt.Println("2. SNEAK PATHS")
	fmt.Println("   • 3-cell sneak paths through parallel resistive network")
	fmt.Println("   • Selector devices provide 1000x improvement")
	fmt.Println("   • IronLattice non-volatile state enables selector-free operation")
	fmt.Println()

	fmt.Println("3. CONDUCTANCE DRIFT")
	fmt.Println("   • FeFET: 0.001 drift coefficient (50x better than RRAM)")
	fmt.Println("   • 10-year retention: >99.9% for IronLattice")
	fmt.Println("   • 30 discrete levels provide margin for small drift")
	fmt.Println()

	fmt.Println("IronLattice Advantages:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println("  ✓ Ferroelectric polarization → excellent retention")
	fmt.Println("  ✓ Low operating voltage → reduced IR drop")
	fmt.Println("  ✓ Non-volatile operation → minimal read disturb")
	fmt.Println("  ✓ 30 discrete levels → noise margin for non-idealities")
	fmt.Println("  ✓ Cool operation → no thermal drift acceleration")
	fmt.Println()

	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println("  \"This is what can go wrong (and how we fix it)\"")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println()
}

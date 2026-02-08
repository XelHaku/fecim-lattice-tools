//go:build ignore

// hysteresis_demo.go - Educational demonstration of ferroelectric hysteresis
//
// This program demonstrates the fundamental physics of ferroelectric hysteresis
// in HfO2-ZrO2 (HZO) materials, which is the core enabling technology behind
// Ferroelectric Compute-in-Memory (FeCIM).
//
// Key Concepts:
//   - Ferroelectric materials exhibit bistable polarization states (+P and -P)
//   - Switching between states requires an electric field exceeding Ec (coercive field)
//   - The P-E hysteresis loop shows path-dependent behavior (memory!)
//   - Partial switching enables multi-level analog storage (30+ discrete states)
//
// Run: go run examples/hysteresis_demo.go
//
// For detailed physics, see: docs/physics-model/
package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║     FeCIM Ferroelectric Hysteresis Demo                          ║")
	fmt.Println("║     Demonstrating P-E Loop Physics for Analog Memory             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// =========================================================================
	// 1. MATERIAL SELECTION
	// =========================================================================
	// FeCIM uses HfO2-ZrO2 superlattice for CMOS compatibility
	fmt.Println("1. MATERIAL: HfO2-ZrO2 Superlattice (Cheema 2020)")
	fmt.Println("   └─ CMOS-compatible ferroelectric at 10nm scale")
	fmt.Println()

	material := ferroelectric.LiteratureSuperlattice()
	printMaterialProperties(material)

	// =========================================================================
	// 2. CREATE PREISACH MODEL
	// =========================================================================
	// The Preisach model captures hysteresis through distributions of hysterons
	// (microscopic dipole units with switching thresholds)
	fmt.Println("2. PHYSICS MODEL: Preisach Hysteresis")
	fmt.Println("   └─ Models ferroelectric as ensemble of bistable hysterons")
	fmt.Println("   └─ Each hysteron switches at threshold (α, β) on Preisach plane")
	fmt.Println()

	model := ferroelectric.NewPreisachModel(material)

	// =========================================================================
	// 3. GENERATE HYSTERESIS LOOP
	// =========================================================================
	// Drive with sinusoidal E-field to trace complete P-E loop
	fmt.Println("3. GENERATING P-E HYSTERESIS LOOP")
	fmt.Println("   └─ Sweeping E-field: -2.5×Ec → +2.5×Ec → -2.5×Ec")
	fmt.Println()

	Emax := 2.5 * material.Ec // Drive amplitude in V/m
	points := 100             // Points per half-cycle

	E, P := model.GetHysteresisLoop(Emax, points)
	printHysteresisLoop(E, P, material)

	// =========================================================================
	// 4. DEMONSTRATE ANALOG LEVELS
	// =========================================================================
	// FeCIM's key innovation: 30 discrete analog states within the hysteresis loop
	fmt.Println("\n4. ANALOG MEMORY: 30 Discrete Polarization States")
	fmt.Println("   └─ Each level stores ~5 bits of information per cell")
	fmt.Println("   └─ Enables compute-in-memory with analog weights")
	fmt.Println()

	demonstrateAnalogLevels(model, material)

	// =========================================================================
	// 5. TEMPERATURE DEPENDENCE
	// =========================================================================
	fmt.Println("\n5. TEMPERATURE EFFECTS")
	fmt.Println("   └─ Ferroelectric properties change with temperature")
	fmt.Println()

	demonstrateTemperature(material)

	// =========================================================================
	// 6. REMANENT POLARIZATION (MEMORY!)
	// =========================================================================
	fmt.Println("\n6. NON-VOLATILE MEMORY: Remanent Polarization")
	fmt.Println("   └─ Polarization persists when E-field returns to zero")
	fmt.Println("   └─ This is the essence of ferroelectric memory!")
	fmt.Println()

	demonstrateMemory(model, material)

	fmt.Println("\n═══════════════════════════════════════════════════════════════════")
	fmt.Println("Demo complete! Run with --gui for interactive visualization.")
	fmt.Println("See: module1-hysteresis/cmd/hysteresis")
}

func printMaterialProperties(m *ferroelectric.HZOMaterial) {
	fmt.Printf("   ┌─────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("   │ Material: %-48s │\n", m.Name)
	fmt.Printf("   ├─────────────────────────────────────────────────────────────┤\n")
	fmt.Printf("   │ Ps (Saturation):    %8.1f µC/cm²  (max polarization)     │\n", m.Ps*100)
	fmt.Printf("   │ Pr (Remanent):      %8.1f µC/cm²  (at E=0, memory!)      │\n", m.Pr*100)
	fmt.Printf("   │ Ec (Coercive):      %8.2f MV/cm   (switching field)      │\n", m.Ec/1e8)
	fmt.Printf("   │ Endurance:          %8.0e cycles (write/erase)          │\n", m.EnduranceCycles)
	fmt.Printf("   │ Film Thickness:     %8.1f nm                             │\n", m.Thickness*1e9)
	fmt.Printf("   └─────────────────────────────────────────────────────────────┘\n")
	fmt.Println()
}

func printHysteresisLoop(E, P []float64, m *ferroelectric.HZOMaterial) {
	// ASCII visualization of the hysteresis loop
	const width = 60
	const height = 20

	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Find bounds
	Emax, Pmax := 0.0, 0.0
	for i := range E {
		if math.Abs(E[i]) > Emax {
			Emax = math.Abs(E[i])
		}
		if math.Abs(P[i]) > Pmax {
			Pmax = math.Abs(P[i])
		}
	}

	// Draw axes
	midY := height / 2
	midX := width / 2
	for x := 0; x < width; x++ {
		grid[midY][x] = '─'
	}
	for y := 0; y < height; y++ {
		grid[y][midX] = '│'
	}
	grid[midY][midX] = '┼'

	// Plot loop
	for i := range E {
		x := int((E[i]/Emax + 1) / 2 * float64(width-1))
		y := int((1 - (P[i]/Pmax+1)/2) * float64(height-1))
		if x >= 0 && x < width && y >= 0 && y < height {
			if i < len(E)/2 {
				grid[y][x] = '●' // Ascending branch
			} else {
				grid[y][x] = '○' // Descending branch
			}
		}
	}

	// Mark key points
	// +Pr (positive remanent)
	prY := int((1 - (m.Pr/Pmax+1)/2) * float64(height-1))
	if prY >= 0 && prY < height {
		grid[prY][midX] = '◆'
	}
	// -Pr (negative remanent)
	prY = int((1 - (-m.Pr/Pmax+1)/2) * float64(height-1))
	if prY >= 0 && prY < height {
		grid[prY][midX] = '◆'
	}
	// +Ec
	ecX := int((m.Ec/Emax + 1) / 2 * float64(width-1))
	if ecX >= 0 && ecX < width {
		grid[midY][ecX] = '◇'
	}
	// -Ec
	ecX = int((-m.Ec/Emax + 1) / 2 * float64(width-1))
	if ecX >= 0 && ecX < width {
		grid[midY][ecX] = '◇'
	}

	// Print
	fmt.Printf("   +P ↑  P-E Hysteresis Loop (● ascending, ○ descending)\n")
	fmt.Printf("      │\n")
	for i, row := range grid {
		if i == 0 {
			fmt.Printf("   +Ps")
		} else if i == midY {
			fmt.Printf("     0")
		} else if i == height-1 {
			fmt.Printf("   -Ps")
		} else {
			fmt.Printf("      ")
		}
		fmt.Printf(" %s\n", string(row))
	}
	fmt.Printf("      └────────────────────────────●────────────────────────────→ +E\n")
	fmt.Printf("      -Ec                          0                           +Ec\n")
	fmt.Println()
	fmt.Println("   Legend: ◆ = Remanent polarization (±Pr), ◇ = Coercive field (±Ec)")
}

func demonstrateAnalogLevels(model *ferroelectric.PreisachModel, m *ferroelectric.HZOMaterial) {
	// Show how partial switching creates discrete analog states
	numLevels := 30
	model.Reset()

	// First, saturate to -Ps
	model.Update(-2.5 * m.Ec)

	fmt.Printf("   Starting from negative saturation (P ≈ -Ps)\n")
	fmt.Printf("   Applying incremental positive fields to reach discrete levels:\n\n")

	// Show a few representative levels
	targetLevels := []int{1, 5, 10, 15, 20, 25, 30}

	fmt.Printf("   Level  │  Target P (µC/cm²)  │  Field Applied  │  Achieved P\n")
	fmt.Printf("   ───────┼────────────────────┼─────────────────┼────────────────\n")

	for _, level := range targetLevels {
		// Calculate target polarization for this level
		targetP := -m.Ps + 2*m.Ps*float64(level-1)/float64(numLevels-1)

		// Apply field to reach target (simplified - real ISPP uses iterative write-verify)
		// Map level to approximate field using hysteresis curve
		fieldFrac := float64(level-1) / float64(numLevels-1)
		field := -m.Ec + 2.5*m.Ec*fieldFrac
		model.Update(field)

		achievedP := model.Polarization()

		fmt.Printf("   %3d    │     %+7.2f         │    %+5.2f×Ec      │   %+7.2f\n",
			level, targetP*100, field/m.Ec, achievedP*100)
	}

	fmt.Println()
	fmt.Printf("   → 30 levels = %.1f bits of information per cell\n", math.Log2(30))
	fmt.Printf("   → Enables efficient matrix-vector multiplication in memory\n")
}

func demonstrateTemperature(m *ferroelectric.HZOMaterial) {
	temperatures := []float64{4, 77, 200, 300, 400}
	tempNames := []string{"Cryo (4K)", "LN2 (77K)", "Cold (200K)", "Room (300K)", "Hot (400K)"}

	fmt.Printf("   Temp      │  Ec (MV/cm)  │  Pr (µC/cm²)  │  Notes\n")
	fmt.Printf("   ──────────┼──────────────┼───────────────┼─────────────────────\n")

	for i, temp := range temperatures {
		// Temperature coefficients from material
		deltaT := temp - 300 // Relative to room temperature
		ec := m.Ec + m.TempCoeffEc*deltaT
		pr := m.Pr + m.TempCoeffPr*deltaT

		// Clamp to physical limits
		if ec < 0.1*m.Ec {
			ec = 0.1 * m.Ec
		}
		if pr < 0.1*m.Pr {
			pr = 0.1 * m.Pr
		}

		notes := ""
		if temp == 4 {
			notes = "Quantum computing"
		} else if temp == 300 {
			notes = "Standard operation"
		} else if temp == 400 {
			notes = "Automotive spec"
		}

		fmt.Printf("   %-10s│    %5.2f     │     %5.1f      │ %s\n",
			tempNames[i], ec/1e8, pr*100, notes)
	}
}

func demonstrateMemory(model *ferroelectric.PreisachModel, m *ferroelectric.HZOMaterial) {
	model.Reset()

	// Write a state
	fmt.Println("   Step 1: Apply positive field (+2×Ec) to write positive state")
	model.Update(2.0 * m.Ec)
	pAfterWrite := model.Polarization()
	fmt.Printf("           P = %+.2f µC/cm²\n", pAfterWrite*100)

	// Remove field
	fmt.Println("\n   Step 2: Remove field (E = 0) - test memory retention")
	model.Update(0)
	pAfterRemove := model.Polarization()
	fmt.Printf("           P = %+.2f µC/cm² (remanent state!)\n", pAfterRemove*100)

	// Read the state
	fmt.Println("\n   Step 3: Read operation (small sensing field)")
	smallField := 0.1 * m.Ec
	model.Update(smallField)
	pRead := model.Polarization()
	fmt.Printf("           P = %+.2f µC/cm² (non-destructive read)\n", pRead*100)

	// Check retention
	retention := math.Abs(pAfterRemove / m.Pr)
	fmt.Printf("\n   → Retention: %.0f%% of Pr maintained at E=0\n", retention*100)
	fmt.Printf("   → This is the essence of non-volatile ferroelectric memory!\n")

	// Show the two stable states
	fmt.Println("\n   Bistable States (Digital Memory Mode):")
	fmt.Println("   ┌─────────┬───────────────────┬─────────────────────┐")
	fmt.Println("   │  State  │   Polarization    │     Represents      │")
	fmt.Println("   ├─────────┼───────────────────┼─────────────────────┤")
	fmt.Printf("   │   '1'   │   P ≈ +%.0f µC/cm²  │   Logic HIGH        │\n", m.Pr*100)
	fmt.Printf("   │   '0'   │   P ≈ -%.0f µC/cm²  │   Logic LOW         │\n", m.Pr*100)
	fmt.Println("   └─────────┴───────────────────┴─────────────────────┘")
}

// Check if help requested
func init() {
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			fmt.Println("FeCIM Hysteresis Demo - Educational ferroelectric physics demonstration")
			fmt.Println()
			fmt.Println("Usage: go run examples/hysteresis_demo.go")
			fmt.Println()
			fmt.Println("This program demonstrates:")
			fmt.Println("  • HZO material properties (Ps, Pr, Ec)")
			fmt.Println("  • P-E hysteresis loop generation")
			fmt.Println("  • 30-level analog state encoding")
			fmt.Println("  • Temperature dependence")
			fmt.Println("  • Non-volatile memory (remanent polarization)")
			fmt.Println()
			fmt.Println("For interactive GUI, run: fecim-lattice-tools hysteresis")
			os.Exit(0)
		}
	}
}

// Helper to create a horizontal bar
func bar(value, max float64, width int) string {
	filled := int(math.Abs(value/max) * float64(width))
	if filled > width {
		filled = width
	}
	if value >= 0 {
		return "[" + strings.Repeat("█", filled) + strings.Repeat(" ", width-filled) + "]"
	}
	return "[" + strings.Repeat(" ", width-filled) + strings.Repeat("█", filled) + "]"
}

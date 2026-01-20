// Demo 5: Thermal Simulation for Ferroelectric CIM
//
// This demo visualizes thermal behavior in a ferroelectric compute-in-memory
// system. It demonstrates:
// - 2D heat map visualization
// - Real-time heat diffusion simulation
// - Multi-layer thermal coupling
// - Hotspot identification
// - Thermal throttling warning system
// - FeCIM's low-power thermal advantage
package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"multilayer-ferroelectric-cim-visualizer/demo5-thermal/pkg/thermal"
)

func main() {
	// Command-line flags
	showSingle := flag.Bool("single", false, "Show single layer simulation")
	showMulti := flag.Bool("multi", false, "Show multi-layer simulation")
	showCompare := flag.Bool("compare", false, "Compare FeCIM vs traditional thermal")
	showAll := flag.Bool("all", false, "Show all thermal demonstrations")
	animate := flag.Bool("animate", false, "Animate heat diffusion (terminal)")
	steps := flag.Int("steps", 100, "Number of simulation steps")
	flag.Parse()

	fmt.Println("================================================")
	fmt.Println("  FeCIM Demo 5: Thermal Simulation")
	fmt.Println("  Heat Management for Ferroelectric CIM")
	fmt.Println("================================================")
	fmt.Println()

	// Show thermal overview
	showThermalOverview()

	// Run specific demos
	if *showAll || *showSingle {
		showSingleLayerDemo(*steps, *animate)
	}
	if *showAll || *showMulti {
		showMultiLayerDemo(*steps)
	}
	if *showAll || *showCompare {
		showComparisonDemo()
	}

	// If no specific flag, show brief overview
	if !*showSingle && !*showMulti && !*showCompare && !*showAll {
		showBriefOverview(*steps)
	}

	fmt.Println()
	fmt.Println("================================================")
	fmt.Println("  FeCIM: 1000x lower power = cool operation")
	fmt.Println("  \"This could lower data center requirements")
	fmt.Println("   by 80 to 90%\" — Dr. external research group")
	fmt.Println("================================================")
}

func showThermalOverview() {
	fmt.Println("Thermal Architecture:")
	fmt.Println()
	fmt.Println("Top View (Heat Map)        Side View (Layers)")
	fmt.Println("─────────────────────      ─────────────────────")
	fmt.Println()
	fmt.Println("░░░▒▒▓▓████▓▓▒▒░░░         ███ Layer 3 (BEOL)")
	fmt.Println("░░▒▒▓██████████▓▒░░           ↕ heat flow")
	fmt.Println("░▒▓████████████████▓▒░      ███ Layer 2 (Crossbar)")
	fmt.Println("░░▒▒▓██████████▓▒░░           ↕ heat flow")
	fmt.Println("░░░▒▒▓▓████▓▓▒▒░░░         ███ Layer 1 (Substrate)")
	fmt.Println("                              ↕ heat flow")
	fmt.Println("25°C ░▒▓█ 85°C            ░░░░ Heat Sink (25°C)")
	fmt.Println()
}

func showSingleLayerDemo(steps int, animate bool) {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│         Single Layer Thermal Analysis       │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	sim := thermal.DefaultThermalSim()
	sim.Reset()
	renderer := thermal.DefaultRenderer()

	// Create a realistic workload pattern
	// Concentrated activity in center (simulating MVM operation)
	fmt.Println("Applying workload pattern (crossbar MVM operation)...")
	for y := 10; y < 22; y++ {
		for x := 10; x < 22; x++ {
			// Higher activity in center
			dx := float64(x - 16)
			dy := float64(y - 16)
			dist := dx*dx + dy*dy
			power := 5e5 * (1 - dist/150) // Gaussian-like power distribution
			if power > 0 {
				sim.SetPower(x, y, power)
			}
		}
	}
	fmt.Println()

	if animate {
		// Animated display
		fmt.Println("Starting animated simulation (press Ctrl+C to stop)...")
		fmt.Println()

		for i := 0; i < steps; i++ {
			// Clear screen (ANSI escape)
			fmt.Print("\033[H\033[2J")

			fmt.Printf("Step %d/%d\n", i+1, steps)
			fmt.Println()

			sim.Step(1e-8)
			fmt.Println(renderer.Render(sim))
			fmt.Println()
			fmt.Println(renderer.RenderStats(sim))

			time.Sleep(100 * time.Millisecond)
		}
	} else {
		// Static display with before/after
		fmt.Println("Initial State:")
		fmt.Println(renderer.Render(sim))
		fmt.Println()

		// Run simulation
		fmt.Printf("Running %d simulation steps...\n", steps)
		sim.StepMultiple(steps, 1e-8)
		fmt.Println()

		fmt.Println("After Heat Diffusion:")
		fmt.Println(renderer.RenderWithOverlay(sim))
		fmt.Println()

		fmt.Println(renderer.RenderStats(sim))
		fmt.Println()

		// Show gradient
		fmt.Println(renderer.TemperatureGradient(sim))
	}
}

func showMultiLayerDemo(steps int) {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│        Multi-Layer Thermal Coupling         │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	mlSim := thermal.DefaultMultiLayerSim()
	mlSim.Reset()
	renderer := thermal.DefaultRenderer()

	fmt.Println("3-Layer FeCIM Stack:")
	fmt.Println("  Layer 3: BEOL Interconnects (Cu, 100 W/m·K)")
	fmt.Println("  Layer 2: Crossbar Array (HZO/Metal, 50 W/m·K)")
	fmt.Println("  Layer 1: Silicon Substrate (Si, 150 W/m·K)")
	fmt.Println("  Heat Sink: 25°C constant temperature")
	fmt.Println()

	// Apply heat to crossbar layer (Layer 1, 0-indexed)
	fmt.Println("Applying heat to crossbar layer (Layer 2)...")
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			mlSim.SetCellPower(1, x, y, 3e5)
		}
	}
	fmt.Println()

	fmt.Println("Initial Layer Temperatures:")
	for i, layer := range mlSim.Layers {
		fmt.Printf("  Layer %d: avg=%.2f°C, max=%.2f°C\n",
			i+1, layer.GetAverageTemperature(), layer.GetMaxTemperature())
	}
	fmt.Println()

	// Run simulation
	fmt.Printf("Running %d simulation steps with inter-layer coupling...\n", steps)
	mlSim.StepMultiple(steps, 1e-8)
	fmt.Println()

	fmt.Println("Final Layer Temperatures:")
	for i, layer := range mlSim.Layers {
		fmt.Printf("  Layer %d: avg=%.2f°C, max=%.2f°C\n",
			i+1, layer.GetAverageTemperature(), layer.GetMaxTemperature())
	}
	fmt.Println()

	// Render side view
	fmt.Println(renderer.RenderSideView(mlSim.Layers))

	// Stack-wide warning check
	warning := mlSim.CheckStackWarning()
	if warning != nil {
		fmt.Printf("\n[THERMAL WARNING Level %d]: %s\n", warning.Level, warning.Message)
		fmt.Printf("  Max Temperature: %.2f°C\n", warning.MaxTemp)
		fmt.Printf("  Hotspots: %d\n", warning.Hotspots)
	} else {
		fmt.Println("Stack Status: Normal operating temperature")
	}
	fmt.Println()

	// Vertical profile at center
	profile := mlSim.VerticalTemperatureProfile(16, 16)
	fmt.Println("Vertical Temperature Profile (center):")
	for i, temp := range profile {
		bar := strings.Repeat("█", int((temp-25)/5))
		fmt.Printf("  Layer %d: %.2f°C %s\n", i+1, temp, bar)
	}
	fmt.Println("  Heat Sink: 25.00°C")
	fmt.Println()
}

func showComparisonDemo() {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│    FeCIM vs Traditional Thermal       │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	// Create two simulations
	feCIM := thermal.DefaultThermalSim()
	traditional := thermal.DefaultThermalSim()
	feCIM.Reset()
	traditional.Reset()

	// Power densities based on Dr. Tour's claims
	// FeCIM: ~0.001 pJ/MAC → 1000x lower
	// Traditional: ~1-10 pJ/MAC
	fmt.Println("Power Comparison (from Dr. Tour's presentation):")
	fmt.Println("  FeCIM: ~0.001 pJ per MAC operation")
	fmt.Println("  Traditional: ~1-10 pJ per MAC operation")
	fmt.Println("  Ratio: 1000-10,000× lower power!")
	fmt.Println()

	feCIMPower := 1e4   // W/m² - very low power density (FeCIM advantage)
	traditionalPower := 1e7  // W/m² - typical high-performance CMOS

	// Apply uniform workload
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			feCIM.SetPower(x, y, feCIMPower)
			traditional.SetPower(x, y, traditionalPower)
		}
	}

	// Run both simulations for 10 microseconds
	simSteps := 10000
	fmt.Printf("Running %d simulation steps (10 µs simulated time)...\n", simSteps)
	fmt.Println()

	feCIM.StepMultiple(simSteps, 1e-9)
	traditional.StepMultiple(simSteps, 1e-9)

	// Display results
	renderer := thermal.DefaultRenderer()

	fmt.Println("FeCIM Heat Map:")
	renderer.MinTemp = 25
	renderer.MaxTemp = 85
	fmt.Println(renderer.Render(feCIM))
	fmt.Printf("  Max Temp: %.2f°C\n", feCIM.GetMaxTemperature())
	fmt.Printf("  Avg Temp: %.2f°C\n", feCIM.GetAverageTemperature())
	warning1 := feCIM.CheckThermalWarning()
	if warning1 != nil {
		fmt.Printf("  Status: %s\n", warning1.Message)
	} else {
		fmt.Println("  Status: Normal operation - no cooling required!")
	}
	fmt.Println()

	fmt.Println("Traditional CIM Heat Map:")
	fmt.Println(renderer.Render(traditional))
	fmt.Printf("  Max Temp: %.2f°C\n", traditional.GetMaxTemperature())
	fmt.Printf("  Avg Temp: %.2f°C\n", traditional.GetAverageTemperature())
	warning2 := traditional.CheckThermalWarning()
	if warning2 != nil {
		fmt.Printf("  Status: %s\n", warning2.Message)
	} else {
		fmt.Println("  Status: Normal operation")
	}
	fmt.Println()

	// Temperature comparison
	ilTempRise := feCIM.GetMaxTemperature() - feCIM.AmbientTemp
	tradTempRise := traditional.GetMaxTemperature() - traditional.AmbientTemp

	fmt.Println("Temperature Rise Comparison:")
	fmt.Println("───────────────────────────────────────")
	fmt.Printf("  FeCIM:  %.2f°C above ambient\n", ilTempRise)
	fmt.Printf("  Traditional: %.2f°C above ambient\n", tradTempRise)

	if ilTempRise > 0.01 {
		fmt.Printf("  Reduction: %.0fx cooler operation!\n", tradTempRise/ilTempRise)
	} else {
		fmt.Println("  FeCIM: Essentially at ambient temperature!")
	}
	fmt.Println()

	// Data center impact
	fmt.Println("Data Center Impact (from Dr. Tour):")
	fmt.Println("───────────────────────────────────────")
	fmt.Println("  \"This could lower the requirements in a data center")
	fmt.Println("   by 80 to 90% of the energy requirements.\"")
	fmt.Println()
	fmt.Println("  Why? Lower power means:")
	fmt.Println("   • No active cooling required")
	fmt.Println("   • Higher integration density possible")
	fmt.Println("   • Longer component lifetime")
	fmt.Println("   • Reduced TCO (Total Cost of Ownership)")
	fmt.Println()
}

func showBriefOverview(steps int) {
	fmt.Println("Brief Thermal Overview:")
	fmt.Println("───────────────────────")
	fmt.Println()

	sim := thermal.DefaultThermalSim()
	sim.Reset()
	renderer := thermal.DefaultRenderer()

	// Apply moderate workload
	for y := 12; y < 20; y++ {
		for x := 12; x < 20; x++ {
			sim.SetPower(x, y, 2e5)
		}
	}

	sim.StepMultiple(steps, 1e-8)

	fmt.Println("Heat Map (32×32 crossbar array):")
	fmt.Println(renderer.Render(sim))
	fmt.Println()
	fmt.Printf("Max Temperature: %.2f°C (limit: %.0f°C)\n",
		sim.GetMaxTemperature(), sim.MaxTemp)
	fmt.Printf("Average Temperature: %.2f°C\n", sim.GetAverageTemperature())

	warning := sim.CheckThermalWarning()
	if warning != nil {
		fmt.Printf("Status: %s\n", warning.Message)
	} else {
		fmt.Println("Status: Normal operating temperature")
	}
	fmt.Println()

	fmt.Println("FeCIM Thermal Advantage:")
	fmt.Println("  • 1000× lower power → negligible heating")
	fmt.Println("  • No thermal throttling required")
	fmt.Println("  • Higher density integration possible")
	fmt.Println("  • 80-90% data center energy reduction")
	fmt.Println()

	fmt.Println("Run with --all for detailed thermal analysis")
	fmt.Println("Run with --compare for FeCIM vs traditional comparison")
	fmt.Println("Run with --animate for real-time heat diffusion visualization")
}

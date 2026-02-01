// Command hysteresis provides an interactive visualization of ferroelectric
// hysteresis in HfO2-ZrO2 superlattice materials.
//
// This is Demo 1 of the FeCIM Visualizer project.
//
// Run modes:
//   - Default: Fyne GUI with real-time P-E curve animation (recommended)
//   - --tui: Terminal user interface (for SSH/remote)
//   - --headless: ASCII terminal output (static, no interactivity)
//   - --vulkan: Vulkan-based graphical interface (advanced)
//
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module1-hysteresis/pkg/gui"
	"fecim-lattice-tools/module1-hysteresis/pkg/render"
	"fecim-lattice-tools/module1-hysteresis/pkg/simulation"
	"fecim-lattice-tools/module1-hysteresis/pkg/tui"
)

// Available materials for --material flag
var materialNames = map[string]*ferroelectric.HZOMaterial{
	"default":      nil, // Will use DefaultHZO()
	"fecim":        nil, // Will use FeCIMMaterial()
	"superlattice": nil, // Will use LiteratureSuperlattice()
	"cryogenic":    nil, // Will use CryogenicHZO()
	"hzo32":        nil, // Will use HZOStandard32()
	"ftj140":       nil, // Will use HZOFJT140()
	"alscn":        nil, // Will use AlScN()
}

func getMaterial(name string) *ferroelectric.HZOMaterial {
	switch name {
	case "fecim":
		return ferroelectric.FeCIMMaterial()
	case "superlattice":
		return ferroelectric.LiteratureSuperlattice()
	case "cryogenic":
		return ferroelectric.CryogenicHZO()
	case "hzo32":
		return ferroelectric.HZOStandard32()
	case "ftj140":
		return ferroelectric.HZOFJT140()
	case "alscn":
		return ferroelectric.AlScN()
	default:
		return ferroelectric.DefaultHZO()
	}
}

func listMaterials() {
	fmt.Println("Available materials (--material <name>):")
	fmt.Println()
	for _, m := range ferroelectric.AllMaterials() {
		fmt.Printf("  %-12s - %s\n", getMaterialKey(m), m.Name)
	}
	fmt.Println()
}

func getMaterialKey(m *ferroelectric.HZOMaterial) string {
	switch m.Name {
	case "HZO (Si-doped)":
		return "default"
	case "FeCIM HZO":
		return "fecim"
	case "FeCIM HZO (TARGET - NOT DEMONSTRATED)":
		return "fecim-target"
	case "Literature Superlattice (Cheema 2020)":
		return "superlattice"
	case "Cryogenic HZO (4K)":
		return "cryogenic"
	case "HZO Standard (32 states)":
		return "hzo32"
	case "HZO FTJ (140 states)":
		return "ftj140"
	case "AlScN (8-16 states)":
		return "alscn"
	default:
		return "default"
	}
}

func printMaterialInfo(m *ferroelectric.HZOMaterial) {
	fmt.Println("\nMaterial Parameters:")
	fmt.Printf("  Remanent Polarization (Pr): %.1f μC/cm²\n", m.Pr*100)
	fmt.Printf("  Saturation Polarization (Ps): %.1f μC/cm²\n", m.Ps*100)
	fmt.Printf("  Coercive Field (Ec): %.2f MV/cm\n", m.Ec/1e8)
	fmt.Printf("  Coercive Voltage (Vc): %.2f V\n", m.CoerciveVoltage())
	fmt.Printf("  Film Thickness: %.0f nm\n", m.Thickness*1e9)
	fmt.Printf("  Relative Permittivity: %.0f\n", m.Epsilon)
	fmt.Println()
}

func runHeadless(engine *simulation.Engine, material *ferroelectric.HZOMaterial) {
	fmt.Println("Running enhanced terminal visualization...")
	fmt.Println()

	// Create advanced Preisach model
	model := ferroelectric.NewPreisachModel(material)

	// Create renderer
	renderer := ferroelectric.NewPERenderer()

	// Generate and render P-E loop
	Emax := material.Ec * 2
	E, P := model.GetHysteresisLoop(Emax, 100)
	fmt.Println(renderer.RenderPELoop(E, P, material))

	// Render domain states (Deprecated in simplified model)
	// alphas, betas, states := model.GetPreisachPlane()
	// fmt.Println(renderer.RenderDomainStates(alphas, betas, states))

	// Render discrete states
	discreteStates := model.DiscreteStates(30)
	fmt.Println(renderer.RenderDiscreteStates(discreteStates))

	// Render switching dynamics (Deprecated in simplified model)
	// times, pols, switched := model.SimulateDomainSwitching(Emax, 10*material.Tau, 50)
	// fmt.Println(renderer.RenderSwitchingDynamics(times, pols, switched, material))

	// Render temperature dependence
	fmt.Println(renderer.RenderTemperatureDependence(material))

	// Render material comparison
	fmt.Println(renderer.RenderMaterialComparison())

	// Summary
	fmt.Println("═════════════════════════════════════════════════════════════════")
	fmt.Println("                     SIMULATION SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════")
	fmt.Println("  Material: %s", material.Name)
	fmt.Printf("  Remanent Polarization: %.1f μC/cm²\n", material.Pr*100)
	fmt.Printf("  Saturation Polarization ( Ps): %.1f μC/cm²\n", material.Ps*100)
	fmt.Printf("  Coercive Field ( Ec): %.2f MV/cm\n", material.Ec/1e8)
	fmt.Printf("  Coercive Voltage (Vc): %.2f V\n", m.CoerciveVoltage())
	fmt.Printf("  Film Thickness: %.0f nm\n", m.Thickness*1e9)
	fmt.Printf("  Relative Permittivity: %.0f\n", m.Epsilon)
	fmt.Println()
}

func runGraphical(engine *simulation.Engine, material *ferroelectric.HZOMaterial) {
	fmt.Println("Starting Vulkan-based graphical interface...")
	fmt.Println("Press ESC or close window to exit.")
	fmt.Println()

	// Create Vulkan renderer
	config := render.DefaultConfig()
	renderer := render.NewVulkanRenderer(config)

	// Create hysteresis plot
	Emax := material.Ec * 1.5
	Pmax := material.Ps * 1.2
	plot := render.NewHysteresisPlot(Emax, Pmax)
	renderer.SetHysteresisPlot(plot)

	// Set up update callback
	frameCount := 0
	engine.Start()
	renderer.SetUpdateCallback(func() {
		// Step simulation
		engine.Step()
		state := engine.State()

		// Update renderer with new polarization
		renderer.UpdatePolarization(state.NormPol)

		// Add point to plot
		plot.AddPoint(state.ElectricField, state.Polarization)

		frameCount++
	})

	// Initialize Vulkan
	if err := renderer.Initialize(); err != nil {
		log.Printf("Failed to initialize Vulkan renderer: %v", err)
		fmt.Println()
		fmt.Println("Vulkan initialization failed. Running in headless mode instead.")
		fmt.Println()
		runHeadless(engine, material)
		os.Exit(0)
	}
	defer renderer.Cleanup()

	// Run render loop
	if err := renderer.Run(); err != nil {
		log.Fatalf("Renderer error: %v", err)
		}
		fmt.Printf("\nSimulation completed. Rendered %d frames.\n", frameCount)
	}
}

// Default: Fyne GUI mode (recommended)
// Pass the selected material name to GUI
	if err := gui.RunWithMaterial(material.Name); err != nil {
		log.Printf("GUI error: %v\n", err)
		fmt.Println("\nFalling back to TUI mode...")
		
		if err := tui.RunWithMaterial(material.Name); err != nil {
			engine := simulation.NewEngine(material)
			engine.SetFrequency(*freq)
			runHeadless(engine, material)
		} else {
			log.Printf("TUI error: %v\n", err)
			fmt.Println("\nFalling back to headless mode...")
		}
	} else {
		log.Printf("GUI error: %v\n", err)
		fmt.Println("\nFalling back to TUI mode...")
	}
}

	// Initialize logging system if requested
	if *logger {
		// Enable file logging
		logging.EnableFileLogging()

		// Set verbosity level
		switch *verbosity {
		case "off":
			// No change - already off
		case "info":
			logging.SetVerbosity(logging.VerbosityInfo)
		case "debug":
			logging.SetVerbosity(logging.VerbosityDebug)
		case "trace":
			logging.SetVerbosity(logging.VerbosityTrace)
		default:
			fmt.Printf("Unknown verbosity level: %s. Using 'info' instead.\n", *verbosity)
			logging.SetVerbosity(logging.VerbosityInfo)
		}
	}

	// Get selected material
	material := getMaterial(*materialName)

	// Determine run mode based on flags
	if *headless {
		// Headless mode - static ASCII output
		fmt.Println("===========================================")
		fmt.Println("  FeCIM Hysteresis Visualizer")
		fmt.Println("  Demo 1: Ferroelectric P-E Curve")
		fmt.Println("===========================================")
		fmt.Println()

		fmt.Printf("Using: %s\n", material.Name)

		printMaterialInfo(material)

		engine := simulation.NewEngine(material)
		engine.SetFrequency(*freq)
		runHeadless(engine, material)
		return
	}

// Command hysteresis provides an interactive visualization of ferroelectric
// hysteresis in HfO2-ZrO2 superlattice materials.
//
// This is Demo 1 of the IronLattice Visualizer project.
//
// Lab Bench Controls:
//   E/D   - Increase/Decrease Electric Field
//   T/G   - Increase/Decrease Temperature
//   F/V   - Increase/Decrease Frequency
//   W     - Cycle Waveform (Sine/Triangle/Square)
//   Space - Pause/Resume
//   R     - Reset
//   Q     - Quit
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/go-gl/glfw/v3.3/glfw"

	"ironlattice-vis/demo1-hysteresis/pkg/ferroelectric"
	"ironlattice-vis/demo1-hysteresis/pkg/gui"
	"ironlattice-vis/demo1-hysteresis/pkg/render"
	"ironlattice-vis/demo1-hysteresis/pkg/simulation"
)

func main() {
	// Command line flags
	optimized := flag.Bool("optimized", false, "Use optimized superlattice parameters")
	freq := flag.Float64("freq", 1e6, "Waveform frequency in Hz")
	headless := flag.Bool("headless", false, "Run in headless mode (no graphics)")
	flag.Parse()

	fmt.Println("===========================================")
	fmt.Println("  IronLattice Hysteresis Visualizer")
	fmt.Println("  Demo 1: Ferroelectric P-E Curve")
	fmt.Println("===========================================")
	fmt.Println()

	// Select material parameters
	var material *ferroelectric.HZOMaterial
	if *optimized {
		material = ferroelectric.OptimizedHZO()
		fmt.Println("Using: Optimized HfO2/ZrO2 Superlattice")
	} else {
		material = ferroelectric.DefaultHZO()
		fmt.Println("Using: Default HZO Parameters")
	}

	// Print material parameters
	printMaterialInfo(material)

	// Create simulation engine
	engine := simulation.NewEngine(material)
	engine.SetFrequency(*freq)

	if *headless {
		// Run headless simulation and print results
		runHeadless(engine)
	} else {
		// Run with graphics
		runGraphical(engine)
	}
}

func printMaterialInfo(m *ferroelectric.HZOMaterial) {
	fmt.Println("\nMaterial Parameters:")
	fmt.Printf("  Remanent Polarization (Pr): %.1f μC/cm²\n", m.Pr*1e4)
	fmt.Printf("  Saturation Polarization (Ps): %.1f μC/cm²\n", m.Ps*1e4)
	fmt.Printf("  Coercive Field (Ec): %.2f MV/cm\n", m.Ec/1e8)
	fmt.Printf("  Coercive Voltage (Vc): %.2f V\n", m.CoerciveVoltage())
	fmt.Printf("  Film Thickness: %.0f nm\n", m.Thickness*1e9)
	fmt.Printf("  Relative Permittivity: %.0f\n", m.Epsilon)
	fmt.Println()
}

func runHeadless(engine *simulation.Engine) {
	fmt.Println("Running headless simulation...")
	fmt.Println()

	// Generate hysteresis loop data
	E, P := engine.GetHysteresisData()

	fmt.Println("Hysteresis Loop Data (E, P):")
	fmt.Println("-----------------------------")

	// Print a subset of points
	step := len(E) / 20
	for i := 0; i < len(E); i += step {
		fmt.Printf("  E: %+8.2e V/m, P: %+8.4f (normalized)\n", E[i], P[i])
	}

	fmt.Println()
	fmt.Println("Discrete States (30-level analog):")
	fmt.Println("-----------------------------------")

	model := ferroelectric.NewPreisachModel(ferroelectric.DefaultHZO())
	states := model.DiscreteStates(30)
	for i, s := range states {
		fmt.Printf("  State %2d: P = %+.4f C/m²\n", i, s)
	}

	fmt.Println()
	fmt.Println("Simulation complete.")
}

func runGraphical(engine *simulation.Engine) {
	fmt.Println("Starting Vulkan-based graphical interface...")
	fmt.Println()

	// Create Lab Bench for interactive control
	labBench := gui.NewLabBench()

	// Print controls
	fmt.Println(labBench.ControlsHelp())
	fmt.Println()

	// Create Vulkan renderer
	config := render.DefaultConfig()
	renderer := render.NewVulkanRenderer(config)

	// Create hysteresis plot
	material := ferroelectric.DefaultHZO()
	Emax := material.Ec * 1.5
	Pmax := material.Ps * 1.2
	plot := render.NewHysteresisPlot(Emax, Pmax)
	renderer.SetHysteresisPlot(plot)

	// Connect Lab Bench callbacks to simulation engine
	labBench.OnEFieldChange = func(E float64) {
		// Convert E-field to voltage for the waveform amplitude
		voltage := E * material.Thickness
		engine.SetAmplitude(voltage)
		fmt.Printf("  Electric Field: %.2f MV/cm (V = %.2f V)\n", E/1e8, voltage)
	}

	labBench.OnTemperatureChange = func(T float64) {
		// Temperature affects Landau coefficients
		// For now, print the effect (full implementation would modify material)
		effects := gui.ComputeTemperatureEffects(T)
		phase := "Ferroelectric"
		if !effects.IsFerro {
			phase = "Paraelectric"
		}
		fmt.Printf("  Temperature: %.0f K → α(T) = %+.2e (%s)\n", T, effects.AlphaT, phase)
	}

	labBench.OnFrequencyChange = func(f float64) {
		engine.SetFrequency(f)
		fmt.Printf("  Frequency: %s\n", formatFreq(f))
	}

	labBench.OnWaveformChange = func(idx int) {
		engine.SetWaveform(simulation.WaveformType(idx))
		fmt.Printf("  Waveform: %s\n", labBench.WaveformNames[idx])
	}

	labBench.OnPauseToggle = func(paused bool) {
		engine.Pause()
		if paused {
			fmt.Println("  [PAUSED]")
		} else {
			fmt.Println("  [RESUMED]")
		}
	}

	labBench.OnReset = func() {
		engine.Reset()
		plot.Clear()
		fmt.Println("  [RESET]")
	}

	labBench.OnQuit = func() {
		renderer.Stop()
	}

	// Set up keyboard callback
	renderer.SetKeyCallback(func(key glfw.Key, action glfw.Action) {
		// Handle H for help toggle
		if key == glfw.KeyH && action == glfw.Press {
			fmt.Println(labBench.ControlsHelp())
			return
		}
		labBench.HandleKeyPress(key, action)
	})

	// Set up update callback
	frameCount := 0
	engine.Start()
	renderer.SetUpdateCallback(func() {
		// Skip update if Lab Bench quit
		if !labBench.Running {
			return
		}

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
		runHeadless(engine)
		os.Exit(0)
	}
	defer renderer.Cleanup()

	fmt.Println("\nLab Bench ready. Use keyboard controls to manipulate the simulation.")
	fmt.Printf("Initial: E = %.2f MV/cm, T = %.0f K, f = %s\n",
		labBench.ElectricFieldAmplitude/1e8,
		labBench.Temperature,
		formatFreq(labBench.Frequency),
	)
	fmt.Println()

	// Run render loop
	if err := renderer.Run(); err != nil {
		log.Fatalf("Renderer error: %v", err)
	}

	fmt.Printf("\nSimulation completed. Rendered %d frames.\n", frameCount)
}

// formatFreq converts Hz to human-readable format.
func formatFreq(f float64) string {
	switch {
	case f >= 1e9:
		return fmt.Sprintf("%.1f GHz", f/1e9)
	case f >= 1e6:
		return fmt.Sprintf("%.1f MHz", f/1e6)
	case f >= 1e3:
		return fmt.Sprintf("%.1f kHz", f/1e3)
	default:
		return fmt.Sprintf("%.0f Hz", f)
	}
}

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
// Common flags:
//   - --json: Output results as JSON
//   - --quiet: Suppress informational output
//   - --config: Load configuration from YAML/JSON file
//   - --batch: Process multiple materials from file
package hysteresiscli

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
	"fecim-lattice-tools/shared/cli"
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

// HysteresisConfig holds configuration for the hysteresis command.
// Can be loaded from a YAML/JSON config file via --config.
type HysteresisConfig struct {
	Material    string  `json:"material" yaml:"material"`
	Frequency   float64 `json:"frequency" yaml:"frequency"`
	Temperature float64 `json:"temperature" yaml:"temperature"`
}

// HysteresisResult represents the output for JSON mode.
type HysteresisResult struct {
	Material           string  `json:"material"`
	RemanentPol        float64 `json:"remanent_polarization_uC_cm2"`
	SaturationPol      float64 `json:"saturation_polarization_uC_cm2"`
	CoerciveField      float64 `json:"coercive_field_MV_cm"`
	CoerciveVoltage    float64 `json:"coercive_voltage_V"`
	Thickness          float64 `json:"thickness_nm"`
	Permittivity       float64 `json:"permittivity"`
	SwitchingTime      float64 `json:"switching_time_ns"`
	EnduranceCycles    float64 `json:"endurance_cycles"`
	DiscreteLevels     int     `json:"discrete_levels"`
	BitsPerCell        float64 `json:"bits_per_cell"`
}

func Run(args []string) error {
	fs := flag.NewFlagSet("hysteresis", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	// Common CLI flags
	commonFlags := cli.NewCommonFlags()
	commonFlags.Register(fs)

	// Command-specific flags
	materialName := fs.String("material", "superlattice", "Material: default, fecim, superlattice, cryogenic, hzo32, ftj140, alscn")
	freq := fs.Float64("freq", 1e6, "Waveform frequency in Hz")
	headless := fs.Bool("headless", false, "Run in headless mode (static ASCII output)")
	tuiMode := fs.Bool("tui", false, "Run terminal UI mode (for SSH/remote)")
	vulkan := fs.Bool("vulkan", false, "Run with Vulkan graphics (GPU accelerated)")
	listMats := fs.Bool("list-materials", false, "List available materials and exit")

	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintln(out, "FeCIM Hysteresis Visualizer")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out, "  fecim-lattice-tools hysteresis [options]")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Options:")
		fs.PrintDefaults()
		fmt.Fprintln(out, cli.CommonUsage())
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(fs.Output(), "Error:", err)
		fs.Usage()
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if commonFlags.WantsHelp() {
		fs.Usage()
		return nil
	}

	// Load config file if specified
	var cfg HysteresisConfig
	if commonFlags.Config != "" {
		loader := cli.NewConfigLoader(commonFlags.Config)
		if err := loader.Load(&cfg); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		// Apply config values (flags take precedence)
		if cfg.Material != "" && *materialName == "superlattice" {
			*materialName = cfg.Material
		}
		if cfg.Frequency > 0 && *freq == 1e6 {
			*freq = cfg.Frequency
		}
	}

	// Create output writer
	out, err := cli.NewOutputWriter(commonFlags)
	if err != nil {
		return err
	}
	defer out.Close()

	// Handle batch processing
	batch, err := cli.NewBatchProcessor(commonFlags.Batch)
	if err != nil {
		return fmt.Errorf("failed to load batch file: %w", err)
	}

	if batch.HasItems() {
		return runBatchHysteresis(batch, commonFlags, out)
	}

	// List materials and exit
	if *listMats {
		if commonFlags.JSON {
			materials := make([]map[string]interface{}, 0)
			for _, m := range ferroelectric.AllMaterials() {
				materials = append(materials, map[string]interface{}{
					"key":  getMaterialKey(m),
					"name": m.Name,
				})
			}
			return out.Result(materials)
		}
		listMaterials()
		return nil
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
		return nil
	}

	if *tuiMode {
		// Terminal UI mode with selected material
		if err := tui.RunWithMaterial(material.Name); err != nil {
			log.Printf("TUI error: %v\n", err)
			fmt.Println("\nFalling back to headless mode...")

			engine := simulation.NewEngine(material)
			engine.SetFrequency(*freq)
			runHeadless(engine, material)
		}
		return nil
	}

	if *vulkan {
		// Vulkan graphical mode
		fmt.Println("===========================================")
		fmt.Println("  FeCIM Hysteresis Visualizer")
		fmt.Println("  Demo 1: Ferroelectric P-E Curve (Vulkan)")
		fmt.Println("===========================================")
		fmt.Println()
		fmt.Printf("Using: %s\n", material.Name)

		printMaterialInfo(material)
		engine := simulation.NewEngine(material)
		engine.SetFrequency(*freq)
		runGraphical(engine, material)
		return nil
	}

	// Default: Fyne GUI mode (recommended)
	// Pass the selected material name to the GUI
	if err := gui.RunWithMaterial(material.Name); err != nil {
		log.Printf("GUI error: %v\n", err)
		fmt.Println("\nFalling back to TUI mode...")

		if err := tui.RunWithMaterial(material.Name); err != nil {
			log.Printf("TUI error: %v\n", err)
			fmt.Println("\nFalling back to headless mode...")

			engine := simulation.NewEngine(material)
			engine.SetFrequency(*freq)
			runHeadless(engine, material)
		}
	}

	return nil
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

	// Create Preisach model
	model := ferroelectric.NewPreisachModel(material)

	// Create renderer
	renderer := ferroelectric.NewPERenderer()

	// Generate and render P-E loop
	Emax := material.Ec * 2
	E, P := model.GetHysteresisLoop(Emax, 100)
	fmt.Println(renderer.RenderPELoop(E, P, material))

	// Render discrete states
	discreteStates := model.DiscreteStates(30)
	fmt.Println(renderer.RenderDiscreteStates(discreteStates))

	// Render temperature dependence
	fmt.Println(renderer.RenderTemperatureDependence(material))

	// Render material comparison
	fmt.Println(renderer.RenderMaterialComparison())

	// Summary
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                     SIMULATION SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Material: %s\n", material.Name)
	fmt.Printf("  Remanent Polarization: %.1f µC/cm²\n", material.Pr*100)
	fmt.Printf("  Coercive Field: %.2f MV/cm\n", material.Ec/1e8)
	fmt.Printf("  Switching Time: %.2f ns\n", material.Tau*1e9)
	fmt.Printf("  Endurance: %.0e cycles\n", material.EnduranceCycles)
	fmt.Printf("  30 Discrete States (conference-claim baseline): %.1f bits/cell\n", 4.91)
	fmt.Println()
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println("  \"It's got 30 discrete states. So it's not 0-1-0-1.\"")
	fmt.Println("  - Dr. external research group (COSM 2025; conference claim)")
	fmt.Println("─────────────────────────────────────────────────────────────")
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
		fmt.Printf("\nRenderer error: %v\n", err)
		fmt.Println("The visualization window encountered an error.")
		fmt.Println("This may be due to:")
		fmt.Println("  - Window was closed unexpectedly")
		fmt.Println("  - GPU driver issues")
		fmt.Println("  - Display server (X11/Wayland) disconnection")
		fmt.Println("\nTry running with --headless flag for non-graphical output.")
		os.Exit(1)
	}

	fmt.Printf("\nSimulation completed. Rendered %d frames.\n", frameCount)
}

// buildMaterialResult creates a JSON-friendly result from a material.
func buildMaterialResult(m *ferroelectric.HZOMaterial) HysteresisResult {
	return HysteresisResult{
		Material:        m.Name,
		RemanentPol:     m.Pr * 100, // Convert to μC/cm²
		SaturationPol:   m.Ps * 100,
		CoerciveField:   m.Ec / 1e8, // Convert to MV/cm
		CoerciveVoltage: m.CoerciveVoltage(),
		Thickness:       m.Thickness * 1e9, // Convert to nm
		Permittivity:    m.Epsilon,
		SwitchingTime:   m.Tau * 1e9, // Convert to ns
		EnduranceCycles: m.EnduranceCycles,
		DiscreteLevels:  30,
		BitsPerCell:     4.91,
	}
}

// runBatchHysteresis processes multiple materials from a batch file.
func runBatchHysteresis(batch *cli.BatchProcessor, flags *cli.CommonFlags, out *cli.OutputWriter) error {
	batchResult := cli.NewBatchResult()

	for _, matName := range batch.Items() {
		material := getMaterial(matName)
		result := buildMaterialResult(material)

		if flags.JSON {
			batchResult.Add(cli.NewResult(result))
		} else {
			out.Print("\n=== Material: %s ===\n", matName)
			printMaterialInfo(material)
		}
	}

	if flags.JSON {
		return out.Result(batchResult)
	}

	return nil
}

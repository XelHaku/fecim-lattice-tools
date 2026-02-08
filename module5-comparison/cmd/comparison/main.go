package comparisoncli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/shared/cli"
)

// ComparisonConfig holds configuration for the comparison CLI.
type ComparisonConfig struct {
	Workload   string  `json:"workload" yaml:"workload"`
	Throughput float64 `json:"throughput" yaml:"throughput"`
}

// ComparisonJSONResult represents the comparison results for JSON output.
type ComparisonJSONResult struct {
	Workload     string                   `json:"workload"`
	Throughput   float64                  `json:"target_throughput"`
	Architectures []ArchitectureResult    `json:"architectures"`
	Advantages   AdvantagesResult         `json:"advantages"`
}

type ArchitectureResult struct {
	Name       string  `json:"name"`
	TDP        float64 `json:"tdp_watts"`
	TOPSPerW   float64 `json:"tops_per_watt"`
	Latency    float64 `json:"latency_ms"`
	Energy     float64 `json:"energy_mj"`
	Throughput float64 `json:"throughput_infs"`
}

type AdvantagesResult struct {
	VsCPU_EnergyReduction  float64 `json:"vs_cpu_energy_reduction"`
	VsCPU_CostReduction    float64 `json:"vs_cpu_cost_reduction"`
	VsGPU_PowerReduction   float64 `json:"vs_gpu_power_reduction"`
	VsGPU_AreaReduction    float64 `json:"vs_gpu_area_reduction"`
}

func Run(args []string) error {
	fs := flag.NewFlagSet("comparison", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	// Common CLI flags
	commonFlags := cli.NewCommonFlags()
	commonFlags.Register(fs)

	// Flags
	showAll := fs.Bool("all", false, "Show all comparisons")
	showSpecs := fs.Bool("specs", false, "Show architecture specifications")
	showInference := fs.Bool("inference", false, "Show inference comparison")
	showDataCenter := fs.Bool("datacenter", false, "Show data center comparison")
	showAdvantages := fs.Bool("advantages", false, "Show FeCIM advantages")
	workloadName := fs.String("workload", "mnist", "Workload: mnist, resnet, bert, gpt2, llm")
	targetTP := fs.Float64("throughput", 10000, "Target throughput (inferences/sec)")
	noColor := fs.Bool("no-color", false, "Disable color output")

	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintln(out, "FeCIM Architecture Comparison CLI")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out, "  fecim-lattice-tools comparison cli [options]")
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
	var cfg ComparisonConfig
	if commonFlags.Config != "" {
		loader := cli.NewConfigLoader(commonFlags.Config)
		if err := loader.Load(&cfg); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if cfg.Workload != "" && *workloadName == "mnist" {
			*workloadName = cfg.Workload
		}
		if cfg.Throughput > 0 && *targetTP == 10000 {
			*targetTP = cfg.Throughput
		}
	}

	// Create output writer
	out, err := cli.NewOutputWriter(commonFlags)
	if err != nil {
		return err
	}
	defer out.Close()

	// Default to all if nothing specified
	if !*showSpecs && !*showInference && !*showDataCenter && !*showAdvantages {
		*showAll = true
	}

	renderer := comparison.NewRenderer()
	renderer.UseColor = !*noColor

	// Select workload
	var workload comparison.Workload
	switch *workloadName {
	case "mnist":
		workload = comparison.MNISTWorkload()
	case "resnet":
		workload = comparison.ResNet50Workload()
	case "bert":
		workload = comparison.BERTBaseWorkload()
	case "gpt2":
		workload = comparison.GPT2Workload()
	case "llm":
		workload = comparison.LLMWorkload()
	default:
		workload = comparison.MNISTWorkload()
	}

	// Run comparison
	comp := comparison.CompareArchitectures(workload, 1, *targetTP)
	advantages := comparison.CalculateAdvantages(comp)

	// JSON mode: output structured data
	if commonFlags.JSON {
		result := buildComparisonResult(comp, advantages, *workloadName, *targetTP)
		return out.Result(result)
	}

	// Header
	printHeader()

	if *showAll || *showSpecs {
		showArchitectureSpecs(renderer, comp.Architectures)
	}

	if *showAll || *showInference {
		showInferenceComparison(renderer, comp, workload)
	}

	if *showAll || *showDataCenter {
		showDataCenterComparison(renderer, comp, *targetTP)
	}

	if *showAll || *showAdvantages {
		showFeCIMAdvantages(renderer, advantages)
	}

	// Summary
	printSummary(comp, advantages)

	return nil
}

func printHeader() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║         FECIM TECHNOLOGY COMPARISON (MODEL INPUTS)     ║")
	fmt.Println("║     CPU+DRAM vs GPU vs FeCIM Compute-in-Memory         ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// IMPORTANT: Print warning banner about estimated specifications
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║  ⚠️  MODEL INPUTS ONLY - NOT VALIDATED ⚠️               ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Println("║  FeCIM is at TRL 4 (lab validation only).             ║")
	fmt.Println("║  Dr. Tour did NOT disclose chip-level specs (TDP, TOPS).    ║")
	fmt.Println("║                                                              ║")
	fmt.Println("║  All CPU/GPU/FeCIM values below are model inputs for  ║")
	fmt.Println("║  visualization only. Do NOT use for investment decisions.   ║")
	fmt.Println("║                                                              ║")
	fmt.Println("║  Reference inputs (not validated):                           ║")
	fmt.Println("║  - 30 analog levels (conference claim)                       ║")
	fmt.Println("║  - Literature ranges for analog states/MNIST                 ║")
	fmt.Println("║  - Public CPU/GPU datasheets                                 ║")
	fmt.Println("║                                                              ║")
	fmt.Println("║  See: docs/comparison/HONESTY_AUDIT.md                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func showArchitectureSpecs(renderer *comparison.Renderer, archs []*comparison.Architecture) {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                  ARCHITECTURE SPECIFICATIONS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println(renderer.RenderArchitectureSpecs(archs))

	// Visual comparison
	fmt.Println("Power Consumption (TDP):")
	fmt.Println(strings.Repeat("─", 60))
	labels := make([]string, len(archs))
	powers := make([]float64, len(archs))
	for i, a := range archs {
		labels[i] = a.Name[:min(18, len(a.Name))]
		powers[i] = a.TDP
	}
	fmt.Println(renderer.RenderBarChart("", labels, powers, "W", true))

	fmt.Println("Chip Area:")
	fmt.Println(strings.Repeat("─", 60))
	areas := make([]float64, len(archs))
	for i, a := range archs {
		areas[i] = a.ChipArea
	}
	fmt.Println(renderer.RenderBarChart("", labels, areas, "mm²", true))

	fmt.Println("Energy Efficiency (TOPS/W):")
	fmt.Println(strings.Repeat("─", 60))
	efficiency := make([]float64, len(archs))
	for i, a := range archs {
		efficiency[i] = a.TOPSPerWatt
	}
	fmt.Println(renderer.RenderBarChart("", labels, efficiency, "TOPS/W", false))
}

func showInferenceComparison(renderer *comparison.Renderer, comp comparison.ComparisonResult, workload comparison.Workload) {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                   INFERENCE COMPARISON")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println(renderer.RenderInferenceComparison(comp.Results, workload))

	// Visual comparisons
	labels := make([]string, len(comp.Results))
	latencies := make([]float64, len(comp.Results))
	energies := make([]float64, len(comp.Results))
	throughputs := make([]float64, len(comp.Results))

	for i, r := range comp.Results {
		labels[i] = r.Architecture[:min(18, len(r.Architecture))]
		latencies[i] = r.Latency
		energies[i] = r.Energy
		throughputs[i] = r.Throughput
	}

	fmt.Println("\nLatency per Inference:")
	fmt.Println(renderer.RenderBarChart("", labels, latencies, "ms", true))

	fmt.Println("Energy per Inference:")
	fmt.Println(renderer.RenderBarChart("", labels, energies, "mJ", true))

	fmt.Println("Throughput:")
	fmt.Println(renderer.RenderBarChart("", labels, throughputs, "inf/s", false))
}

func showDataCenterComparison(renderer *comparison.Renderer, comp comparison.ComparisonResult, targetTP float64) {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                  DATA CENTER COMPARISON")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println(renderer.RenderDataCenterComparison(comp.DataCenter, targetTP))

	// Visual comparisons
	labels := make([]string, len(comp.DataCenter))
	powers := make([]float64, len(comp.DataCenter))
	tcos := make([]float64, len(comp.DataCenter))
	co2s := make([]float64, len(comp.DataCenter))

	for i, m := range comp.DataCenter {
		labels[i] = m.Architecture[:min(18, len(m.Architecture))]
		powers[i] = m.TotalPower
		tcos[i] = m.TCO
		co2s[i] = m.CO2Emissions
	}

	fmt.Println("\nTotal Power Consumption:")
	fmt.Println(renderer.RenderBarChart("", labels, powers, "kW", true))

	fmt.Println("Total Cost of Ownership (Annual):")
	fmt.Println(renderer.RenderBarChart("", labels, tcos, "$/yr", true))

	fmt.Println("CO2 Emissions (Daily):")
	fmt.Println(renderer.RenderBarChart("", labels, co2s, "kg/day", true))
}

func showFeCIMAdvantages(renderer *comparison.Renderer, advantages comparison.FeCIMAdvantage) {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                  FECIM ADVANTAGES")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println(renderer.RenderAdvantages(advantages))

	// Key advantage summary
	fmt.Println("\nKey Advantage Summary:")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("  ⚡ %.0fx more energy efficient than GPU (model output)\n", advantages.VsGPU.EnergyReduction)
	fmt.Printf("  📐 %.0fx smaller chip area than GPU (model output)\n", advantages.VsGPU.AreaReduction)
	fmt.Printf("  🔋 %.0fx lower power than GPU (model output)\n", advantages.VsGPU.PowerReduction)
	fmt.Printf("  💰 %.0fx lower TCO than CPU (model output)\n", advantages.VsCPU.CostReduction)
	fmt.Println()
}

func printSummary(comp comparison.ComparisonResult, adv comparison.FeCIMAdvantage) {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                         SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Printf("Workload: %s\n", comp.Workload.Name)
	fmt.Printf("Operations: %s MACs\n", formatNumber(float64(comp.Workload.TotalOps)))
	fmt.Println()

	// Find FeCIM metrics
	var fecimDC comparison.DataCenterMetrics
	for _, dc := range comp.DataCenter {
		if dc.Architecture == "FeCIM CIM" {
			fecimDC = dc
		}
	}

	fmt.Println("Model output indicates:")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("  • %.0fx energy reduction vs traditional CPU (model output)\n", adv.VsCPU.EnergyReduction)
	fmt.Printf("  • %.0fx throughput improvement vs traditional CPU (model output)\n", adv.VsCPU.ThroughputIncrease)
	fmt.Printf("  • %.0fx power reduction vs GPU accelerator (model output)\n", adv.VsGPU.PowerReduction)
	fmt.Printf("  • %.0fx area reduction vs GPU accelerator (model output)\n", adv.VsGPU.AreaReduction)
	fmt.Printf("  • %.0fx lower annual TCO (model output)\n", adv.VsCPU.CostReduction)
	fmt.Printf("  • %.1f kg CO2 reduction per day\n",
		comp.DataCenter[0].CO2Emissions-fecimDC.CO2Emissions)
	fmt.Println()

	// Calculate total savings
	cpuTCO := comp.DataCenter[0].TCO
	annualSavings := cpuTCO - fecimDC.TCO
	percentSavings := annualSavings / cpuTCO * 100

	fmt.Println("Data Center Impact:")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("  Total Power Reduction:    %.1f kW → %.1f kW\n",
		comp.DataCenter[0].TotalPower, fecimDC.TotalPower)
	fmt.Printf("  Annual TCO Savings:       $%.0f (%.0f%% reduction)\n",
		annualSavings, percentSavings)
	fmt.Printf("  Carbon Footprint:         %.0f%% reduction\n",
		(1-fecimDC.CO2Emissions/comp.DataCenter[0].CO2Emissions)*100)
	fmt.Println()

	// Why FeCIM wins
	fmt.Println("Why FeCIM Wins (model assumptions):")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("  ✓ Compute-in-Memory: No von Neumann bottleneck (assumption)")
	fmt.Println("  ✓ 30 Analog Levels (conference claim; model input)")
	fmt.Println("  ✓ Ferroelectric: Non-volatile, low power switching (assumption)")
	fmt.Println("  ✓ Parallel MVM: Massively parallel matrix operations (assumption)")
	fmt.Println("  ✓ Low Voltage: Minimal energy per operation (assumption)")
	fmt.Println()

	// Dr. Tour quote
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println("  \"This could lower the requirements in a data center")
	fmt.Println("   by 80 to 90%.\" - Dr. external research group (claim, unverified)")
	fmt.Println()
	fmt.Printf("  Model output: %.0f%% reduction in this comparison\n", percentSavings)
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println()
}

func formatNumber(n float64) string {
	if n >= 1e12 {
		return fmt.Sprintf("%.1fT", n/1e12)
	} else if n >= 1e9 {
		return fmt.Sprintf("%.1fB", n/1e9)
	} else if n >= 1e6 {
		return fmt.Sprintf("%.1fM", n/1e6)
	} else if n >= 1e3 {
		return fmt.Sprintf("%.1fK", n/1e3)
	}
	return fmt.Sprintf("%.0f", n)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// buildComparisonResult creates a JSON-friendly comparison result.
func buildComparisonResult(comp comparison.ComparisonResult, adv comparison.FeCIMAdvantage, workload string, throughput float64) ComparisonJSONResult {
	result := ComparisonJSONResult{
		Workload:      workload,
		Throughput:    throughput,
		Architectures: make([]ArchitectureResult, 0),
		Advantages: AdvantagesResult{
			VsCPU_EnergyReduction: adv.VsCPU.EnergyReduction,
			VsCPU_CostReduction:   adv.VsCPU.CostReduction,
			VsGPU_PowerReduction:  adv.VsGPU.PowerReduction,
			VsGPU_AreaReduction:   adv.VsGPU.AreaReduction,
		},
	}

	for i, arch := range comp.Architectures {
		if i < len(comp.Results) {
			r := comp.Results[i]
			result.Architectures = append(result.Architectures, ArchitectureResult{
				Name:       arch.Name,
				TDP:        arch.TDP,
				TOPSPerW:   arch.TOPSPerWatt,
				Latency:    r.Latency,
				Energy:     r.Energy,
				Throughput: r.Throughput,
			})
		}
	}

	return result
}

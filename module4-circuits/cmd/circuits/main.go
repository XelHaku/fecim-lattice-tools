// Demo 4: Peripheral Circuits for Ferroelectric CIM
//
// This demo visualizes the peripheral circuits required for a complete
// ferroelectric compute-in-memory system: DAC, ADC, TIA, and Charge Pump.
// Shows how digital values are converted to/from analog for crossbar operations.
//
// Common flags:
//   - --json: Output results as JSON
//   - --quiet: Suppress informational output
//   - --config: Load configuration from YAML/JSON file
package circuitscli

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"fecim-lattice-tools/shared/cli"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/peripherals"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// CircuitsConfig holds configuration for the circuits CLI.
type CircuitsConfig struct {
	Level     int  `json:"level" yaml:"level"`
	ShowAll   bool `json:"show_all" yaml:"show_all"`
	Verbosity int  `json:"verbosity" yaml:"verbosity"`
}

// CircuitsResult represents peripheral circuit specs for JSON output.
type CircuitsResult struct {
	DAC  *DACResult  `json:"dac,omitempty"`
	ADC  *ADCResult  `json:"adc,omitempty"`
	TIA  *TIAResult  `json:"tia,omitempty"`
	Pump *PumpResult `json:"pump,omitempty"`
}

type DACResult struct {
	Bits          int     `json:"bits"`
	Levels        int     `json:"levels"`
	VrefLow       float64 `json:"vref_low_v"`
	VrefHigh      float64 `json:"vref_high_v"`
	Resolution    float64 `json:"resolution_v"`
	EnergyPerConv float64 `json:"energy_fj"`
}

type ADCResult struct {
	Bits           int     `json:"bits"`
	Levels         int     `json:"levels"`
	ConversionTime float64 `json:"conversion_time_ns"`
	ENOB           float64 `json:"enob"`
	EnergyPerConv  float64 `json:"energy_fj"`
}

type TIAResult struct {
	Gain         float64 `json:"gain_kohm"`
	Bandwidth    float64 `json:"bandwidth_mhz"`
	DynamicRange float64 `json:"dynamic_range_db"`
	Power        float64 `json:"power_uw"`
}

type PumpResult struct {
	InputVoltage  float64 `json:"input_voltage_v"`
	OutputVoltage float64 `json:"output_voltage_v"`
	Efficiency    float64 `json:"efficiency_percent"`
	Stages        int     `json:"stages"`
}

func printCircuitsUsage(fs *flag.FlagSet, out io.Writer) {
	fmt.Fprintln(out, "FeCIM Peripheral Circuits CLI")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  fecim-lattice-tools circuits cli [options]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Options:")
	previous := fs.Output()
	fs.SetOutput(out)
	fs.PrintDefaults()
	fs.SetOutput(previous)
	fmt.Fprintln(out, cli.CommonUsage())
}

func Run(args []string) error {
	return runCircuits(args, os.Stdout, os.Stderr)
}

func runCircuits(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("circuits", flag.ContinueOnError)
	fs.SetOutput(stderr)

	// Common CLI flags
	commonFlags := cli.NewCommonFlags()
	commonFlags.Register(fs)

	// Command-line flags
	showDAC := fs.Bool("dac", false, "Show DAC (Digital-to-Analog) details")
	showADC := fs.Bool("adc", false, "Show ADC (Analog-to-Digital) details")
	showTIA := fs.Bool("tia", false, "Show TIA (Transimpedance Amplifier) details")
	showPump := fs.Bool("pump", false, "Show Charge Pump details")
	showAll := fs.Bool("all", false, "Show all peripheral circuits")
	showLinearity := fs.Bool("linearity", false, "Show INL/DNL linearity analysis")
	showTiming := fs.Bool("timing", false, "Show timing diagrams")
	showPower := fs.Bool("power", false, "Show power breakdown")
	showISPP := fs.Bool("ispp", false, "Run ISPP write/verify demo (shared hysteresis physics)")
	demoLevel := fs.Int("level", 15, "Demo level for conversion (0-29)")
	enableLogger := fs.Bool("logger", false, "Enable file logging (logs/)")
	verbosity := fs.Int("verbosity", 2, "Logging verbosity: 0=off, 1=info, 2=debug, 3=trace")

	fs.Usage = func() { printCircuitsUsage(fs, fs.Output()) }

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, "Error:", err)
		printCircuitsUsage(fs, stderr)
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if commonFlags.WantsHelp() {
		printCircuitsUsage(fs, stdout)
		return nil
	}

	// Load config file if specified
	var cfg CircuitsConfig
	if commonFlags.Config != "" {
		loader := cli.NewConfigLoader(commonFlags.Config)
		if err := loader.Load(&cfg); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if cfg.Level > 0 && *demoLevel == 15 {
			*demoLevel = cfg.Level
		}
		if cfg.ShowAll {
			*showAll = true
		}
	}

	// Create output writer
	out, err := cli.NewOutputWriter(commonFlags)
	if err != nil {
		return err
	}
	defer out.Close()

	if *enableLogger {
		logging.EnableFileLogging()
		logging.SetVerbosity(logging.VerbosityLevel(*verbosity))
		peripherals.EnableLogging()
	}

	// JSON mode: output structured data
	if commonFlags.JSON {
		result := buildCircuitsResult(*showDAC, *showADC, *showTIA, *showPump, *showAll)
		return out.Result(result)
	}

	fmt.Println("================================================")
	fmt.Println("  FeCIM Demo 4: Peripheral Circuits")
	fmt.Println("  Full System Integration for CIM")
	fmt.Println("================================================")
	fmt.Println()

	// Show system overview
	showSystemOverview()

	// Show specific circuits or all
	if *showAll || *showDAC {
		showDACDemo(*demoLevel)
	}
	if *showAll || *showADC {
		showADCDemo(*demoLevel)
	}
	if *showAll || *showTIA {
		showTIADemo()
	}
	if *showAll || *showPump {
		showChargePumpDemo()
	}

	// Show linearity analysis
	if *showLinearity || *showAll {
		showLinearityAnalysis()
	}

	// Show timing diagrams
	if *showTiming || *showAll {
		showTimingDiagram()
	}

	// Show power breakdown
	if *showPower || *showAll {
		showPowerBreakdown()
	}
	if *showISPP {
		showISPPDemo(*demoLevel)
	}

	// If no specific flag, show brief overview of all
	if !*showDAC && !*showADC && !*showTIA && !*showPump && !*showAll && !*showLinearity && !*showTiming && !*showPower {
		showBriefOverview(*demoLevel)
	}

	fmt.Println()
	fmt.Println("================================================")
	fmt.Println("  Peripheral circuits enable CMOS-compatible")
	fmt.Println("  ferroelectric compute-in-memory systems")
	fmt.Println("================================================")

	return nil
}

func showSystemOverview() {
	fmt.Println("System Architecture:")
	fmt.Println()
	fmt.Println("     WRITE PATH                    READ PATH")
	fmt.Println("     ──────────                    ─────────")
	fmt.Println()
	fmt.Println("  Digital Level ──┐            ┌── Digital Level")
	fmt.Println("      (0-29)      │            │      (0-29)")
	fmt.Println("                  ▼            ▲")
	fmt.Println("            ┌─────────┐  ┌─────────┐")
	fmt.Println("            │   DAC   │  │   ADC   │")
	fmt.Println("            │  5-bit  │  │  5-bit  │")
	fmt.Println("            └────┬────┘  └────┬────┘")
	fmt.Println("                 │            ▲")
	fmt.Println("                 ▼            │")
	fmt.Println("            ┌─────────┐  ┌─────────┐")
	fmt.Println("            │ Charge  │  │   TIA   │")
	fmt.Println("            │  Pump   │  │ Current │")
	fmt.Println("            │ 1V→1.5V │  │→Voltage │")
	fmt.Println("            └────┬────┘  └────┬────┘")
	fmt.Println("                 │            ▲")
	fmt.Println("                 ▼            │")
	fmt.Println("            ┌────────────────────┐")
	fmt.Println("            │                    │")
	fmt.Println("            │     CROSSBAR       │")
	fmt.Println("            │ 30-Level FeFET     │")
	fmt.Println("            │                    │")
	fmt.Println("            └────────────────────┘")
	fmt.Println()
}

func showDACDemo(level int) {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│           DAC (Write Path)                  │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	dac := peripherals.DefaultDAC()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Resolution: %d bits (%d levels)\n", dac.Bits, dac.Levels())
	fmt.Printf("  Vref Range: %.1fV to %.1fV\n", dac.VrefLow, dac.VrefHigh)
	fmt.Printf("  LSB Size: %.3f V\n", dac.Resolution())
	fmt.Printf("  INL: %.2f LSB, DNL: %.2f LSB\n", dac.INL, dac.DNL)
	fmt.Printf("  Settling Time: %.0f ns\n", dac.SettleTime)
	fmt.Printf("  Energy/Conv: %.2f fJ\n", dac.EnergyPerConversion()*1e15)
	fmt.Println()

	// Show conversion for specified level
	if level < 0 || level > 29 {
		level = 15
	}
	voltage := dac.Convert(level)
	voltageNL := dac.ConvertWithNonlinearity(level)

	fmt.Printf("Level %d Conversion:\n", level)
	fmt.Printf("  Ideal Voltage: %+.3f V\n", voltage)
	fmt.Printf("  With NL Error: %+.3f V (Δ = %.3f mV)\n", voltageNL, (voltageNL-voltage)*1000)
	fmt.Println()

	// Show voltage ladder mapped from 30 FeCIM levels onto DAC code space.
	fmt.Println("Voltage Ladder (30 FeCIM levels mapped to DAC codes):")
	fmt.Println()
	for i := 0; i < 30; i++ {
		dacCode := int(math.Round(float64(i) * float64(dac.Levels()-1) / 29.0))
		v := dac.Convert(dacCode)
		bar := int((v - dac.VrefLow) / (dac.VrefHigh - dac.VrefLow) * 40)
		if bar < 0 {
			bar = 0
		}
		if bar > 40 {
			bar = 40
		}
		marker := " "
		if i == level {
			marker = "→"
		}
		fmt.Printf("  %s %2d (DAC %2d): %+.2fV │%s│\n", marker, i, dacCode, v, strings.Repeat("█", bar)+strings.Repeat("░", 40-bar))
	}
	fmt.Println()
}

func showADCDemo(level int) {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│           ADC (Read Path)                   │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	adc := peripherals.DefaultADC()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Resolution: %d bits (%d levels)\n", adc.Bits, adc.Levels())
	fmt.Printf("  Vref Range: %.1fV to %.1fV\n", adc.VrefLow, adc.VrefHigh)
	fmt.Printf("  LSB Size: %.3f V\n", adc.Resolution())
	fmt.Printf("  Type: SAR (Successive Approximation)\n")
	fmt.Printf("  Conversion Time: %.0f ns\n", adc.ConversionTime)
	fmt.Printf("  ENOB: %.2f bits\n", adc.ENOB())
	fmt.Printf("  Theoretical SNR: %.1f dB\n", adc.TheoreticalSNR())
	fmt.Printf("  Effective SNR: %.1f dB\n", adc.EffectiveSNR())
	fmt.Printf("  Energy/Conv: %.2f fJ\n", adc.EnergyPerConversion()*1e15)
	fmt.Println()

	// Demo: Convert voltage back to level
	voltage := float64(level)/29.0*(adc.VrefHigh-adc.VrefLow) + adc.VrefLow
	convertedLevel := adc.Convert(voltage)
	convertedLevelNL := adc.ConvertWithNonlinearity(voltage)

	fmt.Printf("ADC Conversion (Input: %.3fV for level %d):\n", voltage, level)
	fmt.Printf("  Ideal Output: Level %d\n", convertedLevel)
	fmt.Printf("  With NL: Level %d\n", convertedLevelNL)
	fmt.Println()

	// Show quantization
	fmt.Println("Quantization Thresholds:")
	maxShow := adc.Levels() - 1
	if maxShow > 8 {
		maxShow = 8
	}
	for i := 0; i < maxShow; i++ {
		threshold := adc.VrefLow + float64(i+1)*adc.Resolution()
		fmt.Printf("  Level %d-%d boundary: %.3fV\n", i, i+1, threshold)
	}
	fmt.Printf("  ... (%d total thresholds for %d-bit ADC)\n", adc.Levels()-1, adc.Bits)
	fmt.Println()
}

func showTIADemo() {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│     TIA (Transimpedance Amplifier)          │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	tia := peripherals.DefaultTIA()

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Transimpedance Gain: %.0f kΩ\n", tia.Gain/1e3)
	fmt.Printf("  Bandwidth: %.0f MHz\n", tia.Bandwidth/1e6)
	fmt.Printf("  Input Noise: %.1f pA/√Hz\n", tia.InputNoiseRMS*1e12)
	fmt.Printf("  Output Offset: %.1f mV\n", tia.OutputOffset*1000)
	fmt.Printf("  Max Input Current: %.0f µA\n", tia.MaxInputCurrent*1e6)
	fmt.Printf("  Max Output Voltage: %.1f V\n", tia.MaxOutputVoltage)
	fmt.Println()

	fmt.Printf("Performance:\n")
	fmt.Printf("  Min Detectable Current: %.2f nA\n", tia.MinDetectableCurrent()*1e9)
	fmt.Printf("  Dynamic Range: %.1f dB\n", tia.DynamicRange())
	fmt.Printf("  Settling Time: %.1f ns\n", tia.SettlingTime()*1e9)
	fmt.Printf("  Power: %.1f µW\n", tia.PowerConsumption()*1e6)
	fmt.Println()

	// Show current-to-voltage conversion examples
	fmt.Println("Current-to-Voltage Conversion:")
	testCurrents := []float64{1e-6, 10e-6, 50e-6, 100e-6}
	for _, current := range testCurrents {
		voltage := tia.Convert(current)
		snr := tia.SNR(current)
		fmt.Printf("  %5.0f µA → %.3f V (SNR: %.1f dB)\n", current*1e6, voltage, snr)
	}
	fmt.Println()
}

func showISPPDemo(level int) {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│  ISPP Write/Verify (Shared Hysteresis)      │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	mat := sharedphysics.FeCIMMaterial()
	numLevels := mat.GetNumLevels()
	if numLevels <= 0 {
		numLevels = sharedphysics.DefaultLevels
	}

	if level < 0 || level >= numLevels {
		level = numLevels / 2
	}

	gmin := mat.Gmin
	gmax := mat.Gmax
	if gmin == 0 && gmax == 0 {
		gmin = 1e-6
		gmax = 100e-6
	}

	// Use the same physics path as the hysteresis module (L-K + write controller).
	solver := sharedphysics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.Temperature = 300
	solver.EnableNoise = false
	solver.UseNLS = false
	if !solver.UseMaterialAlpha {
		solver.UpdateParams()
	}

	controller := sharedphysics.NewWriteController(solver, mat)
	controller.MaxIterations = 50
	controller.Tolerance = 0.03 * (gmax - gmin)
	controller.PulseWidth = 20e-9
	if mat.Ec > 0 && mat.Thickness > 0 {
		controller.MaxVoltage = mat.Ec * mat.Thickness * 3.5
	}

	targetG := mat.DiscreteLevel(level, numLevels)
	attempts, success, overshoots := controller.WriteTargetWithReset(targetG, false)

	finalP := solver.GetState()
	finalG := sharedphysics.PolarizationToConductanceWithParams(finalP, mat.Ps, gmin, gmax, sharedphysics.ParseConductanceModel(mat.ConductanceModel), mat.KvT, mat.VGSReadV, mat.VT0V)
	finalLevel := int(math.Round((finalG - gmin) / (gmax - gmin) * float64(numLevels-1)))
	if finalLevel < 0 {
		finalLevel = 0
	}
	if finalLevel >= numLevels {
		finalLevel = numLevels - 1
	}

	fmt.Printf("Target Level: %d of %d\n", level, numLevels-1)
	fmt.Printf("Target Conductance: %.2f µS\n", targetG*1e6)
	fmt.Printf("Attempts: %d | Overshoots: %d | Success: %v\n", attempts, overshoots, success)
	fmt.Printf("Final Conductance: %.2f µS (Level ~%d)\n", finalG*1e6, finalLevel)
	fmt.Println()
}

func showChargePumpDemo() {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│          Charge Pump (Voltage Boost)        │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	// Positive pump
	pumpPos := peripherals.DefaultChargePump()
	fmt.Println("Positive Charge Pump (+1.5V):")
	fmt.Printf("  Input: %.1f V (CMOS supply)\n", pumpPos.InputVoltage)
	fmt.Printf("  Target Output: %.1f V\n", pumpPos.OutputVoltage)
	fmt.Printf("  Stages: %d (Dickson topology)\n", pumpPos.Stages)
	fmt.Printf("  Ideal Output: %.2f V\n", pumpPos.IdealOutputVoltage())
	fmt.Printf("  Actual Output: %.2f V (with losses)\n", pumpPos.ActualOutputVoltage())
	fmt.Printf("  Boost Factor: %.2fx\n", pumpPos.BoostFactor())
	fmt.Printf("  Efficiency: %.0f%%\n", pumpPos.Efficiency*100)
	fmt.Printf("  Ripple: %.1f mV\n", pumpPos.OutputRipple()*1000)
	fmt.Printf("  Rise Time: %.1f µs\n", pumpPos.RiseTime()*1e6)
	fmt.Println()

	// Negative pump
	pumpNeg := peripherals.NegativePump()
	fmt.Println("Negative Charge Pump (-1.5V):")
	fmt.Printf("  Input: %.1f V\n", pumpNeg.InputVoltage)
	fmt.Printf("  Target Output: %.1f V\n", pumpNeg.OutputVoltage)
	fmt.Printf("  Stages: %d\n", pumpNeg.Stages)
	fmt.Println()

	// Energy analysis
	fmt.Println("Energy Analysis:")
	pulseDuration := 100e-9 // 100 ns write pulse
	energy := pumpPos.EnergyPerOperation(pulseDuration)
	fmt.Printf("  Write Pulse Duration: %.0f ns\n", pulseDuration*1e9)
	fmt.Printf("  Energy per Write: %.2f pJ\n", energy*1e12)
	fmt.Printf("  Power Input: %.1f µW\n", pumpPos.PowerInput()*1e6)
	fmt.Printf("  Power Loss: %.1f µW\n", pumpPos.PowerLoss()*1e6)
	fmt.Println()
}

func showBriefOverview(level int) {
	dac := peripherals.DefaultDAC()
	adc := peripherals.DefaultADC()
	tia := peripherals.DefaultTIA()
	pump := peripherals.DefaultChargePump()

	fmt.Println("Peripheral Circuit Summary:")
	fmt.Println()

	// DAC summary
	voltage := dac.Convert(level)
	fmt.Printf("  DAC: Level %d → %+.3f V (5-bit, %.0f fJ/conv)\n",
		level, voltage, dac.EnergyPerConversion()*1e15)

	// Charge pump summary
	fmt.Printf("  Charge Pump: %.1f V → %.2f V (%.0f%% efficient)\n",
		pump.InputVoltage, pump.ActualOutputVoltage(), pump.Efficiency*100)

	// TIA summary
	current := float64(level) / 29.0 * tia.MaxInputCurrent
	tiaVoltage := tia.Convert(current)
	fmt.Printf("  TIA: %.1f µA → %.3f V (%.0f kΩ gain)\n",
		current*1e6, tiaVoltage, tia.Gain/1e3)

	// ADC summary
	adcLevel := adc.Convert(tiaVoltage)
	fmt.Printf("  ADC: %.3f V → Level %d (5-bit, %.0f fJ/conv)\n",
		tiaVoltage, adcLevel, adc.EnergyPerConversion()*1e15)
	fmt.Println()

	// Total energy estimate
	totalEnergy := dac.EnergyPerConversion() + adc.EnergyPerConversion() + pump.EnergyPerOperation(100e-9)
	fmt.Printf("Estimated Energy per Operation: %.1f fJ\n", totalEnergy*1e15)
	fmt.Println()

	fmt.Println("Run with --all for detailed view of all circuits")
	fmt.Println("Or use --dac, --adc, --tia, --pump for specific circuits")
	fmt.Println("   --linearity: INL/DNL analysis")
	fmt.Println("   --timing: Timing diagrams")
	fmt.Println("   --power: Power breakdown")
}

func showLinearityAnalysis() {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│        INL/DNL Linearity Analysis           │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	dac := peripherals.DefaultDAC()
	adc := peripherals.DefaultADC()

	// DAC INL/DNL
	dacAnalysis := dac.AnalyzeINLDNL()
	fmt.Println("DAC Linearity (5-bit):")
	fmt.Println()
	fmt.Println("INL Plot (Integral Nonlinearity in LSB):")
	showINLPlot(dacAnalysis.INLValues, len(dacAnalysis.INLValues))
	fmt.Printf("  Max INL: %.3f LSB at code %d\n", dacAnalysis.MaxINL, dacAnalysis.WorstCode)
	fmt.Println()

	fmt.Println("DNL Plot (Differential Nonlinearity in LSB):")
	showDNLPlot(dacAnalysis.DNLValues, len(dacAnalysis.DNLValues))
	fmt.Printf("  Max DNL: +%.3f / %.3f LSB\n", dacAnalysis.MaxDNL, dacAnalysis.MinDNL)
	fmt.Println()

	// ADC INL/DNL
	adcAnalysis := adc.AnalyzeINLDNL()
	fmt.Println("ADC Linearity (5-bit SAR):")
	fmt.Println()
	fmt.Println("INL Plot:")
	showINLPlot(adcAnalysis.INLValues, len(adcAnalysis.INLValues))
	fmt.Printf("  Max INL: %.3f LSB at code %d\n", adcAnalysis.MaxINL, adcAnalysis.WorstCode)
	fmt.Println()

	fmt.Println("DNL Plot:")
	showDNLPlot(adcAnalysis.DNLValues, len(adcAnalysis.DNLValues))
	fmt.Printf("  Max DNL: +%.3f / %.3f LSB\n", adcAnalysis.MaxDNL, adcAnalysis.MinDNL)
	fmt.Println()

	// Monotonicity check
	fmt.Println("Monotonicity Check:")
	dacMonotonic := checkMonotonicity(dacAnalysis.DNLValues)
	adcMonotonic := checkMonotonicity(adcAnalysis.DNLValues)
	fmt.Printf("  DAC: %s (DNL > -1 LSB everywhere)\n", passFailMark(dacMonotonic))
	fmt.Printf("  ADC: %s (DNL > -1 LSB everywhere)\n", passFailMark(adcMonotonic))
	fmt.Println()
}

func showINLPlot(inl []float64, levels int) {
	// Scale: -1 to +1 LSB
	width := 50
	center := width / 2

	fmt.Printf("  +1.0 LSB %s┐\n", strings.Repeat(" ", center-5))
	for i := 0; i < levels; i++ {
		// Map INL to position
		pos := center + int(inl[i]*float64(center))
		if pos < 0 {
			pos = 0
		}
		if pos >= width {
			pos = width - 1
		}

		line := make([]rune, width)
		for j := range line {
			line[j] = '·'
		}
		line[center] = '│'
		line[pos] = '●'

		fmt.Printf("  %2d: %s\n", i, string(line))
	}
	fmt.Printf("  -1.0 LSB %s┘\n", strings.Repeat(" ", center-5))
}

func showDNLPlot(dnl []float64, levels int) {
	// Scale: -1 to +1 LSB
	width := 50
	center := width / 2

	fmt.Printf("  +1.0 LSB %s┐\n", strings.Repeat(" ", center-5))
	for i := 1; i < levels; i++ {
		// Map DNL to position
		pos := center + int(dnl[i]*float64(center))
		if pos < 0 {
			pos = 0
		}
		if pos >= width {
			pos = width - 1
		}

		line := make([]rune, width)
		for j := range line {
			line[j] = '·'
		}
		line[center] = '│'
		line[pos] = '■'

		fmt.Printf("  %2d: %s\n", i, string(line))
	}
	fmt.Printf("  -1.0 LSB %s┘\n", strings.Repeat(" ", center-5))
}

func checkMonotonicity(dnl []float64) bool {
	for _, d := range dnl {
		if d < -1.0 {
			return false
		}
	}
	return true
}

func passFailMark(pass bool) string {
	if pass {
		return "✓ PASS"
	}
	return "✗ FAIL"
}

func showTimingDiagram() {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│           Timing Diagram                    │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	dac := peripherals.DefaultDAC()
	adc := peripherals.DefaultADC()
	tia := peripherals.DefaultTIA()
	pump := peripherals.DefaultChargePump()

	timing := peripherals.AnalyzeTiming(dac, adc, tia, pump)

	// ASCII timing diagram
	fmt.Println("Write Cycle:")
	fmt.Println("  ┌────────────────────────────────────────────────────────────┐")
	fmt.Println("  │  DAC   │   Pump   │   FeFET Write   │  Array  │            │")
	fmt.Println("  │ Settle │   Rise   │     Pulse       │ Settle  │            │")
	fmt.Println("  └────────────────────────────────────────────────────────────┘")
	fmt.Printf("  │%.0fns │ %.0fns   │     %.0fns       │   %.0fns  │            │\n",
		timing.DACSettle*1e9, timing.PumpRise*1e9, timing.WritePulse*1e9, timing.ArraySettle*1e9)
	fmt.Println()

	fmt.Println("Read Cycle:")
	fmt.Println("  ┌───────────────────────────────────────────────────────┐")
	fmt.Println("  │   DAC   │  Array  │   TIA    │    ADC     │")
	fmt.Println("  │ Settle  │ Settle  │  Settle  │   Convert  │")
	fmt.Println("  └───────────────────────────────────────────────────────┘")
	fmt.Printf("  │  %.0fns  │  %.0fns  │  %.0fns   │   %.0fns    │\n",
		timing.DACSettle*1e9, timing.ArraySettle*1e9, timing.TIASettle*1e9, timing.ADCConvert*1e9)
	fmt.Println()

	// Waveform visualization
	fmt.Println("Signal Waveforms:")
	fmt.Println()
	fmt.Println("  CLK     ┌─┐ ┌─┐ ┌─┐ ┌─┐ ┌─┐ ┌─┐ ┌─┐ ┌─┐")
	fmt.Println("          ─┘ └─┘ └─┘ └─┘ └─┘ └─┘ └─┘ └─┘ └─")
	fmt.Println()
	fmt.Println("  WREN    ────┐                   ┌────────")
	fmt.Println("              └───────────────────┘")
	fmt.Println()
	fmt.Println("  VWL     ────────┐           ┌───────────")
	fmt.Println("                  └───────────┘")
	fmt.Println()
	fmt.Println("  VDAC    ────────────┬───────┬───────────")
	fmt.Println("          ────────────┘       └───────────")
	fmt.Println()
	fmt.Println("  VPUMP            ╱‾‾‾‾‾‾‾‾‾‾╲")
	fmt.Println("          ────────╱            ╲──────────")
	fmt.Println()

	// Timing summary
	fmt.Println("Timing Summary:")
	fmt.Printf("  Write Time: %.1f ns\n", timing.WriteTime*1e9)
	fmt.Printf("  Read Time:  %.1f ns\n", timing.ReadTime*1e9)
	fmt.Printf("  Cycle Time: %.1f ns\n", timing.CycleTime*1e9)
	fmt.Printf("  Max Throughput: %.2f GOPS\n", timing.MaxThroughput/1e9)
	fmt.Println()
}

func showPowerBreakdown() {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│           Power Breakdown                   │")
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	dac := peripherals.DefaultDAC()
	adc := peripherals.DefaultADC()
	tia := peripherals.DefaultTIA()
	pump := peripherals.DefaultChargePump()

	timing := peripherals.AnalyzeTiming(dac, adc, tia, pump)
	power := peripherals.AnalyzePower(dac, adc, tia, pump, timing)

	// Energy breakdown
	fmt.Println("Energy per Operation:")
	fmt.Printf("  DAC:   %6.2f fJ  %s %.0f%%\n", power.DACEnergy*1e15,
		makeBarChart(power.DACFraction, 30), power.DACFraction*100)
	fmt.Printf("  ADC:   %6.2f fJ  %s %.0f%%\n", power.ADCEnergy*1e15,
		makeBarChart(power.ADCFraction, 30), power.ADCFraction*100)
	fmt.Printf("  TIA:   %6.2f fJ  %s %.0f%%\n", power.TIAEnergy*1e15,
		makeBarChart(power.TIAFraction, 30), power.TIAFraction*100)
	fmt.Printf("  Pump:  %6.2f fJ  %s %.0f%%\n", power.PumpEnergy*1e15,
		makeBarChart(power.PumpFraction, 30), power.PumpFraction*100)
	fmt.Println("  " + strings.Repeat("─", 45))
	fmt.Printf("  Total: %6.2f fJ\n", power.TotalEnergy*1e15)
	fmt.Println()

	// Power consumption at max throughput
	fmt.Println("Power at Max Throughput:")
	fmt.Printf("  DAC:   %6.2f µW\n", power.DACPower*1e6)
	fmt.Printf("  ADC:   %6.2f µW\n", power.ADCPower*1e6)
	fmt.Printf("  TIA:   %6.2f µW\n", power.TIAPower*1e6)
	fmt.Printf("  Pump:  %6.2f µW\n", power.PumpPower*1e6)
	fmt.Println("  " + strings.Repeat("─", 20))
	fmt.Printf("  Total: %6.2f µW\n", power.TotalPower*1e6)
	fmt.Println()

	// Pie chart visualization (ASCII)
	fmt.Println("Energy Distribution:")
	fmt.Println()
	showAsciiPieChart(map[string]float64{
		"DAC":  power.DACFraction,
		"ADC":  power.ADCFraction,
		"TIA":  power.TIAFraction,
		"Pump": power.PumpFraction,
	})
	fmt.Println()

	// Comparison with digital
	fmt.Println("Efficiency Comparison:")
	digitalEnergy := 1e-12 // 1 pJ typical for digital multiply-accumulate
	cimEnergy := power.TotalEnergy
	improvement := digitalEnergy / cimEnergy
	fmt.Printf("  Digital MAC:     ~1000 fJ\n")
	fmt.Printf("  CIM Operation:   %.0f fJ\n", cimEnergy*1e15)
	fmt.Printf("  Improvement:     %.0fx more efficient\n", improvement)
	fmt.Println()
}

func makeBarChart(fraction float64, width int) string {
	filled := int(fraction * float64(width))
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

func showAsciiPieChart(data map[string]float64) {
	// Simple horizontal bar representation instead of actual pie
	fmt.Println("  ┌" + strings.Repeat("─", 50) + "┐")
	totalWidth := 50
	for name, frac := range data {
		segWidth := int(frac * float64(totalWidth))
		if segWidth < 1 && frac > 0 {
			segWidth = 1
		}
		if segWidth > totalWidth {
			segWidth = totalWidth
		}
		char := "█"
		switch name {
		case "DAC":
			char = "█"
		case "ADC":
			char = "▓"
		case "TIA":
			char = "▒"
		case "Pump":
			char = "░"
		}
		bar := strings.Repeat(char, segWidth) + strings.Repeat(" ", totalWidth-segWidth)
		fmt.Printf("  │%s│ %s: %.0f%%\n", bar, name, frac*100)
	}
	fmt.Println("  └" + strings.Repeat("─", 50) + "┘")
}

// buildCircuitsResult creates a JSON-friendly result for peripheral circuits.
func buildCircuitsResult(showDAC, showADC, showTIA, showPump, showAll bool) CircuitsResult {
	result := CircuitsResult{}

	if showAll || showDAC {
		dac := peripherals.DefaultDAC()
		result.DAC = &DACResult{
			Bits:          dac.Bits,
			Levels:        dac.Levels(),
			VrefLow:       dac.VrefLow,
			VrefHigh:      dac.VrefHigh,
			Resolution:    dac.Resolution(),
			EnergyPerConv: dac.EnergyPerConversion() * 1e15,
		}
	}

	if showAll || showADC {
		adc := peripherals.DefaultADC()
		result.ADC = &ADCResult{
			Bits:           adc.Bits,
			Levels:         adc.Levels(),
			ConversionTime: adc.ConversionTime,
			ENOB:           adc.ENOB(),
			EnergyPerConv:  adc.EnergyPerConversion() * 1e15,
		}
	}

	if showAll || showTIA {
		tia := peripherals.DefaultTIA()
		result.TIA = &TIAResult{
			Gain:         tia.Gain / 1e3,
			Bandwidth:    tia.Bandwidth / 1e6,
			DynamicRange: tia.DynamicRange(),
			Power:        tia.PowerConsumption() * 1e6,
		}
	}

	if showAll || showPump {
		pump := peripherals.DefaultChargePump()
		result.Pump = &PumpResult{
			InputVoltage:  pump.InputVoltage,
			OutputVoltage: pump.ActualOutputVoltage(),
			Efficiency:    pump.Efficiency * 100,
			Stages:        pump.Stages,
		}
	}

	return result
}

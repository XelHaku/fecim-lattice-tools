// pkg/export/summary.go
// Design summary generator for FeCIM crossbar arrays.
//
// GenerateDesignSummary returns a structured human-readable text report
// covering physical, electrical, compute, and timing parameters. This
// is the canonical "first look" report for an array configuration,
// analogous to OpenROAD's report_design_area + report_power output
// or OpenLane's metrics.json presented as human-readable text.
//
// The report is intentionally conservative: it flags placeholder values
// and avoids making unsupported claims about silicon-level performance.
package export

import (
	"fmt"
	"math"
	"strings"
	"time"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// GenerateDesignSummary returns a plain-text design summary for the given array config.
// The summary covers: identification, physical, electrical, compute, peripherals,
// timing anchors, and a note on confidence level.
func GenerateDesignSummary(cfg config.ArrayConfig) string {
	var sb strings.Builder

	tech := cfg.Technology
	if tech == "" {
		tech = "sky130"
	}

	// Physical dimensions
	totalCells := cfg.Rows * cfg.Cols
	cellArea := cfg.CellWidth * cfg.CellHeight          // µm²
	arrayArea := float64(totalCells) * cellArea          // µm²
	wlLength := float64(cfg.Cols) * cfg.CellWidth        // µm
	blLength := float64(cfg.Rows) * cfg.CellHeight       // µm
	dieMarginUm := 20.0
	dieW := float64(cfg.Cols)*cfg.CellWidth + 2*dieMarginUm
	dieH := float64(cfg.Rows)*cfg.CellHeight + 2*dieMarginUm
	dieArea := dieW * dieH

	// Conductance range (architecture-dependent, from crosssim.go convention)
	var gMaxUS, gMinUS float64
	switch strings.ToLower(cfg.Architecture) {
	case "1t1r", "2t1r":
		gMaxUS = 100.0
		gMinUS = 0.01
	default: // passive
		gMaxUS = 10.0
		gMinUS = 0.001
	}
	nLevels := 30

	// Wire resistance (1 Ω/cell estimate)
	// WL is horizontal, spanning cfg.Cols cells; BL is vertical, spanning cfg.Rows cells.
	wlResOhm := float64(cfg.Cols) * 1.0

	// IR drop estimate: V_read × R_wire × G_mean
	// For passive arrays, worst-case is far cell in large array
	vRead := 0.1 // V
	gMeanUS := (gMaxUS + gMinUS) / 2.0
	irDropMV := vRead * wlResOhm * (gMeanUS * 1e-6) * 1000.0 // mV estimate

	// Compute throughput estimate (MAC/s at 100 MHz clock, 1 MAC per clock for demo)
	clockMHz := 100.0
	throughputGOPS := float64(totalCells) * clockMHz / 1000.0 // GOPs

	// Cell defaults (tech-specific for timing anchors)
	var cellCfg config.CellConfig
	switch strings.ToLower(tech) {
	case "gf180mcu", "gf180":
		cellCfg = config.DefaultGF180CellConfig()
	case "ihp_sg13g2", "ihp", "sg13g2":
		cellCfg = config.DefaultIHPCellConfig()
	default:
		cellCfg = config.DefaultCellConfig()
	}

	// Architecture note
	archNote := "Passive 0T1R — sneak paths present, recommend ≤16×16"
	switch strings.ToLower(cfg.Architecture) {
	case "1t1r":
		archNote = "1T1R — selector transistor eliminates sneak paths, scalable to 128×128+"
	case "2t1r":
		archNote = "2T1R — dual selectors for enhanced isolation"
	}

	title := fmt.Sprintf("FeCIM Crossbar Design Summary — %dx%d (%s / %s)",
		cfg.Rows, cfg.Cols, cfg.Architecture, tech)
	sep := strings.Repeat("─", 60)

	sb.WriteString(title + "\n")
	sb.WriteString(sep + "\n")
	sb.WriteString(fmt.Sprintf("Generated:  %s\n", time.Now().Format("2006-01-02 15:04")))
	sb.WriteString(fmt.Sprintf("Mode:       %s\n", cfg.Mode))
	sb.WriteString("\n")

	// Physical
	sb.WriteString("── Physical ───────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("  Array size:         %d × %d = %d cells\n", cfg.Rows, cfg.Cols, totalCells))
	sb.WriteString(fmt.Sprintf("  Architecture:       %s\n", archNote))
	sb.WriteString(fmt.Sprintf("  Cell dimensions:    %.3f × %.3f µm\n", cfg.CellWidth, cfg.CellHeight))
	sb.WriteString(fmt.Sprintf("  Cell area:          %.4f µm²\n", cellArea))
	sb.WriteString(fmt.Sprintf("  Array footprint:    %.4f µm²  (%.3f × %.3f µm)\n",
		arrayArea, float64(cfg.Cols)*cfg.CellWidth, float64(cfg.Rows)*cfg.CellHeight))
	sb.WriteString(fmt.Sprintf("  Die area (est.):    %.2f µm²  (%.1f × %.1f µm, +%.0fµm margin)\n",
		dieArea, dieW, dieH, dieMarginUm))
	sb.WriteString(fmt.Sprintf("  Array utilization:  %.1f%%  (array / die)\n", arrayArea/dieArea*100))
	sb.WriteString(fmt.Sprintf("  WL wire length:     %.3f µm  (%d cols × %.3f µm/cell)\n",
		wlLength, cfg.Cols, cfg.CellWidth))
	sb.WriteString(fmt.Sprintf("  BL wire length:     %.3f µm  (%d rows × %.3f µm/cell)\n",
		blLength, cfg.Rows, cfg.CellHeight))
	sb.WriteString("\n")

	// Electrical
	sb.WriteString("── Electrical ─────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("  VDD:                %.1f V\n", cellCfg.Voltage))
	sb.WriteString(fmt.Sprintf("  V_read (MVM):       %.2f V\n", vRead))
	sb.WriteString(fmt.Sprintf("  G_max (LRS):        %.3f µS  (R = %.0f kΩ)\n",
		gMaxUS, 1.0/(gMaxUS*1e-3)))
	sb.WriteString(fmt.Sprintf("  G_min (HRS):        %.4f µS  (R = %.0f MΩ)\n",
		gMinUS, 1.0/gMinUS))
	sb.WriteString(fmt.Sprintf("  Conductance levels: %d (30-level simulation baseline)\n", nLevels))
	sb.WriteString(fmt.Sprintf("  WL line resistance: %.1f Ω  (%d cols × 1 Ω/cell estimate)\n",
		wlResOhm, cfg.Cols))
	if strings.ToLower(cfg.Architecture) == "passive" {
		sb.WriteString(fmt.Sprintf("  IR drop (est.):     %.2f mV  (worst case, far cell)\n", irDropMV))
		sb.WriteString("  Sneak paths:        Present — use 1T1R for arrays > 16×16\n")
	} else {
		sb.WriteString("  IR drop:            Reduced by selector transistor\n")
		sb.WriteString("  Sneak paths:        Suppressed by selector\n")
	}
	sb.WriteString("\n")

	// Compute mode metrics
	if strings.ToLower(cfg.Mode) == "compute" {
		sb.WriteString("── Compute Mode ────────────────────────────────────────\n")
		sb.WriteString(fmt.Sprintf("  MVM dimensions:     %d × %d  (rows=inputs, cols=outputs)\n",
			cfg.Rows, cfg.Cols))
		sb.WriteString(fmt.Sprintf("  Parallelism:        %d MACs/cycle\n", totalCells))
		sb.WriteString(fmt.Sprintf("  Throughput (est.):  %.2f GOPS  @ %.0f MHz, 1 MAC/cycle\n",
			throughputGOPS, clockMHz))
		qsnr := 6.02*math.Log2(float64(nLevels)) + 1.76 // SQNR estimate for uniform quantization
		sb.WriteString(fmt.Sprintf("  Quantization SQNR:  %.1f dB  (%d levels, uniform)\n", qsnr, nLevels))
		sb.WriteString("  Weight precision:   30 levels → ~5-bit effective (simulation only)\n")
		sb.WriteString("\n")
	}

	// Peripheral circuit model
	sb.WriteString("── Peripheral Circuits ────────────────────────────────\n")
	sb.WriteString("  DAC resolution:     4 bits  (literature-optimal, Xu et al. 2021)\n")
	sb.WriteString("  ADC resolution:     4 bits  (ADC dominates 40-60% system energy)\n")
	sb.WriteString("  TIA:                Column sense amplifier (virtual GND)\n")
	sb.WriteString(fmt.Sprintf("  Clock:              %.0f MHz  (peripheral control logic)\n", clockMHz))
	sb.WriteString("\n")

	// Timing anchors (from published FeFET references)
	sb.WriteString("── Timing (Published Anchors) ─────────────────────────\n")
	sb.WriteString(fmt.Sprintf("  FE write time:      %.0f ns  (Trentzsch et al., IEDM 2016)\n",
		cellCfg.RiseTime))
	sb.WriteString(fmt.Sprintf("  FE read time:       %.0f ns  (fast read; FE state sensed)\n",
		cellCfg.FallTime))
	sb.WriteString(fmt.Sprintf("  Input capacitance:  %.4f pF  (mid-range FeFET estimate)\n",
		cellCfg.InputCap))
	sb.WriteString(fmt.Sprintf("  Leakage power:      %.4f nW/cell  (NC-FinFET low-leak envelope)\n",
		cellCfg.LeakagePower))
	sb.WriteString(fmt.Sprintf("  Array leakage:      %.4f nW  (%d cells)\n",
		cellCfg.LeakagePower*float64(totalCells), totalCells))
	sb.WriteString("\n")

	// Technology note
	sb.WriteString("── Technology ─────────────────────────────────────────\n")
	switch strings.ToLower(tech) {
	case "ihp_sg13g2", "ihp", "sg13g2":
		sb.WriteString("  PDK:                IHP SG13G2 (130nm BiCMOS, open-source)\n")
		sb.WriteString("  Cell site:          0.48 × 3.78 µm  (from IHP-Open-PDK LEF)\n")
		sb.WriteString("  Supply voltage:     1.5 V (LV core)\n")
		sb.WriteString("  Metal layers:       7 (met1–met5 thin, TopMet1/TopMet2 thick)\n")
	case "gf180mcu", "gf180":
		sb.WriteString("  PDK:                GF180MCU (180nm, open-source via Volare)\n")
		sb.WriteString("  Cell site:          0.46 × 3.75 µm  (9-track standard cell)\n")
		sb.WriteString("  Supply voltage:     1.8 V\n")
	default: // sky130
		sb.WriteString("  PDK:                SKY130 (130nm, open-source via Volare)\n")
		sb.WriteString("  Cell site:          0.46 × 2.72 µm  (unithd standard cell site)\n")
		sb.WriteString("  Supply voltage:     1.8 V\n")
		sb.WriteString("  Metal layers:       5 (met1–met5)\n")
	}
	sb.WriteString("\n")

	// Confidence banner
	sb.WriteString("── Confidence ─────────────────────────────────────────\n")
	sb.WriteString("  [ESTIMATED] This summary uses simulation defaults, not SPICE-\n")
	sb.WriteString("  characterized device models. Timing values are published FeFET\n")
	sb.WriteString("  anchors, not measured on this array. IR drop and throughput\n")
	sb.WriteString("  estimates assume idealized peripheral circuits.\n")
	sb.WriteString("  For fabrication: replace with PDK-validated SPICE models.\n")

	return sb.String()
}

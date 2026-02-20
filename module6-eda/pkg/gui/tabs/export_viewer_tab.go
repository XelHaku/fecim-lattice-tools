package tabs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

var exportFormats = []string{"LEF", "Liberty", "Liberty (Multi-Corner)", "Verilog", "DEF", "Config (JSON)", "SDC", "Design Summary", "SPICE", "SVG Layout", "CSV Table", "Array Statistics", "Export Manifest"}

// MakeExportViewerTab creates a read-only export preview tab for LEF/Liberty/Verilog/DEF/SPICE.
func MakeExportViewerTab(cfg *config.ArrayConfig, window fyne.Window) fyne.CanvasObject {
	if cfg == nil {
		cfg = &config.ArrayConfig{Rows: 4, Cols: 4, Mode: "storage", Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72}
	}

	formatSelect := widget.NewSelect(exportFormats, nil)
	formatSelect.SetSelected("LEF")

	status := widget.NewLabel("Ready")
	preview := widget.NewMultiLineEntry()
	preview.Wrapping = fyne.TextWrapOff
	preview.TextStyle.Monospace = true
	preview.Disable()

	refresh := func() {
		content, source := loadExportPreviewContent(formatSelect.Selected, cfg)
		preview.SetText(content)
		status.SetText("Source: " + source)
	}

	formatSelect.OnChanged = func(string) { refresh() }

	refreshBtn := widget.NewButton("Refresh", refresh)

	saveBtn := widget.NewButton("Save to File…", func() {
		if window == nil {
			return
		}
		ext := formatExtension(formatSelect.Selected)
		design := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
		defaultName := design + ext

		dlg := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
			if err != nil || w == nil {
				return
			}
			defer w.Close()
			if _, werr := w.Write([]byte(preview.Text)); werr != nil {
				dialog.ShowError(werr, window)
			}
		}, window)
		dlg.SetFileName(defaultName)
		dlg.Show()
	})

	copyBtn := widget.NewButton("Copy", func() {
		if window != nil {
			window.Clipboard().SetContent(preview.Text)
			status.SetText(fmt.Sprintf("Copied %d bytes to clipboard", len(preview.Text)))
		}
	})

	header := container.NewHBox(
		widget.NewLabel("Format:"),
		formatSelect,
		refreshBtn,
		saveBtn,
		copyBtn,
		widget.NewSeparator(),
		status,
	)

	refresh()

	return container.NewBorder(header, nil, nil, nil, container.NewScroll(preview))
}

// formatExtension returns the canonical file extension for the given format name.
func formatExtension(format string) string {
	switch format {
	case "LEF":
		return ".lef"
	case "Liberty":
		return ".lib"
	case "Liberty (Multi-Corner)":
		return ".lib"
	case "Verilog":
		return ".v"
	case "DEF":
		return ".def"
	case "Config (JSON)":
		return ".json"
	case "SDC":
		return ".sdc"
	case "Design Summary":
		return ".txt"
	case "SPICE":
		return ".sp"
	case "SVG Layout":
		return ".svg"
	case "CSV Table":
		return ".csv"
	case "Array Statistics":
		return ".txt"
	case "Export Manifest":
		return ".txt"
	default:
		return ".txt"
	}
}

func loadExportPreviewContent(format string, cfg *config.ArrayConfig) (content string, source string) {
	tech := cfg.Technology
	if tech == "" {
		tech = "sky130"
	}

	// Derive cell name and directory from architecture so LEF/Liberty reflect the correct cell type.
	cellName := "fecim_bitcell"
	cellDir := "fecim_bitcell"
	switch cfg.Architecture {
	case "1t1r":
		cellName = "fecim_1t1r_bitcell"
		cellDir = "fecim_1t1r_bitcell"
	case "2t1r":
		cellName = "fecim_2t1r_bitcell"
		cellDir = "fecim_2t1r_bitcell"
	}

	cellCfg := config.CellConfig{
		Name:         cellName,
		Width:        cfg.CellWidth,
		Height:       cfg.CellHeight,
		CellType:     cfg.Architecture,
		Technology:   tech,
		RiseTime:     10.0,
		FallTime:     10.0,
		InputCap:     0.015,
		LeakagePower: 0.0003,
	}

	design := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	dataDir := "data"

	tryRead := func(path string) (string, bool) {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", false
		}
		return string(b), true
	}

	switch format {
	case "LEF":
		// Try the architecture-specific cell file first.
		archLEF := filepath.Join("cells", cellDir, cellName+".lef")
		if s, ok := tryRead(archLEF); ok {
			return s, archLEF
		}
		// Fall back to the appropriate in-memory generator for the architecture.
		switch cfg.Architecture {
		case "1t1r":
			return export.Generate1T1RLEF(cellCfg), "generated (in-memory)"
		case "2t1r":
			return export.Generate2T1RLEF(cellCfg), "generated (in-memory)"
		default:
			return export.GenerateLEF(cellCfg), "generated (in-memory)"
		}

	case "Liberty":
		archLib := filepath.Join("cells", cellDir, cellName+".lib")
		if s, ok := tryRead(archLib); ok {
			return s, archLib
		}
		return export.GenerateLiberty(cellCfg), "generated (in-memory)"

	case "Liberty (Multi-Corner)":
		return export.GenerateMultiCornerLiberty(cellCfg), "generated (in-memory, TT/SS/FF corners)"

	case "Verilog":
		// Check flat path (Generate All) then bundled path (Export Package).
		for _, p := range []string{
			filepath.Join(dataDir, design+".v"),
			filepath.Join(dataDir, design, design+".v"),
		} {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		return export.GenerateArrayVerilog(*cfg), "generated (in-memory)"

	case "DEF":
		// Check flat path (Generate All), legacy output/, then bundled path (Export Package).
		for _, p := range []string{
			filepath.Join(dataDir, design+".def"),
			filepath.Join("output", design+".def"),
			filepath.Join(dataDir, design, design+".def"),
		} {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		// Generate architecture-aware DEF in-memory (passive/1t1r/2t1r).
		return generateBuilderDEF(*cfg), "generated (in-memory)"

	case "Config (JSON)":
		// Flat path written by Generate All; subdirectory path written by Export Package.
		for _, p := range []string{
			filepath.Join(dataDir, "config.json"),
			filepath.Join(dataDir, design, "config.json"),
		} {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		return export.GenerateLibreLaneConfig(*cfg), "generated (in-memory)"

	case "SDC":
		// Flat path written by Generate All; subdirectory path written by Export Package.
		for _, p := range []string{
			filepath.Join(dataDir, "constraints.sdc"),
			filepath.Join(dataDir, design, "constraints.sdc"),
		} {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		return export.GenerateSDC(*cfg), "generated (in-memory)"

	case "Design Summary":
		// Only Export Package writes this file (to the bundled subdirectory).
		// Generate All does not write a design_summary.txt.
		for _, p := range []string{
			filepath.Join(dataDir, design, "design_summary.txt"),
			filepath.Join(dataDir, "design_summary.txt"),
		} {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		return export.GenerateDesignSummary(*cfg), "generated (in-memory)"

	case "SPICE":
		paths := []string{
			filepath.Join(dataDir, design+".sp"),
			filepath.Join(dataDir, "fecim_array.sp"),
		}
		for _, p := range paths {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		// Generate a subcircuit preview: FeFET definition + architecture-specific bitcell.
		mat := export.DefaultHzoFeFETMaterial()
		preview := fmt.Sprintf(
			"* FeCIM Array SPICE Subcircuit Preview\n"+
				"* Array: %dx%d  Architecture: %s  Technology: %s\n"+
				"* NOTE: This shows cell subcircuit definitions only.\n"+
				"*       Full array netlist: use CLI with --spice flag.\n\n",
			cfg.Rows, cfg.Cols, cfg.Architecture, tech)
		preview += export.GenerateFeFETSubcircuit(mat)
		switch cfg.Architecture {
		case "1t1r":
			preview += export.Generate1T1RSubcircuit()
		case "2t1r":
			preview += export.Generate2T1RSubcircuit()
		}
		return preview, "generated (subcircuit preview)"

	case "SVG Layout":
		p := filepath.Join("data", fmt.Sprintf("fecim_crossbar_%dx%d.svg", cfg.Rows, cfg.Cols))
		if s, ok := tryRead(p); ok {
			return s, p
		}
		return export.GenerateLayoutSVGWithDefaults(*cfg), "generated (in-memory)"

	case "CSV Table":
		return generateCSVPreview(cfg), "generated (synthetic sample)"

	case "Array Statistics":
		return generateArrayStatistics(cfg), "generated (in-memory)"

	case "Export Manifest":
		return generateExportManifest(cfg), "generated (in-memory)"

	default:
		return "", "unknown format"
	}
}

// generateCSVPreview generates a synthetic CSV sample showing what the exported
// conductance table looks like. Values are derived from a linear level gradient
// across cells to illustrate the full G_min..G_max range.
//
// The real CSV (via CLI --csv flag) encodes actual programmed conductance states;
// this preview shows representative values only.
//
// Conductance range is architecture-specific (matching CrossSim config defaults):
//   - passive (0T1R): G_min=0.001 µS (HRS ~1 GΩ), G_max=10 µS (LRS ~100 kΩ)
//   - 1T1R / 2T1R:    G_min=0.01 µS  (HRS ~100 MΩ), G_max=100 µS (LRS ~10 kΩ)
//
// Header: row,col,level,conductance_uS,resistance_ohm,program_V
func generateCSVPreview(cfg *config.ArrayConfig) string {
	const quantLevels = 30
	const maxPreview = 16 // rows to show before truncating

	// Architecture-specific conductance range (matches crosssim.go convention)
	var gMin, gMax float64
	switch strings.ToLower(cfg.Architecture) {
	case "1t1r", "2t1r":
		gMin = 0.01  // µS — HRS with selector (~100 MΩ)
		gMax = 100.0 // µS — LRS with selector (~10 kΩ)
	default: // passive 0T1R
		gMin = 0.001 // µS — HRS passive (~1 GΩ)
		gMax = 10.0  // µS — LRS passive (~100 kΩ)
	}

	// Programming voltage range (scaled to match conductance range)
	const vMin = 2.0 // V (VProgMin default)
	const vMax = 5.0 // V (VProgMax default)

	rows, cols := cfg.Rows, cfg.Cols
	if rows <= 0 {
		rows = 4
	}
	if cols <= 0 {
		cols = 4
	}
	total := rows * cols

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# FeCIM Crossbar CSV Export — %dx%d array, mode=%s, arch=%s\n", rows, cols, cfg.Mode, cfg.Architecture))
	sb.WriteString(fmt.Sprintf("# GMin=%.4fµS (HRS)  GMax=%.4fµS (LRS)  Levels=%d\n", gMin, gMax, quantLevels))
	sb.WriteString("# row,col,level,conductance_uS,resistance_ohm,program_V\n")
	sb.WriteString("# NOTE: Synthetic gradient sample; actual data requires compilation.\n")
	sb.WriteString("#       Use CLI: fecim-lattice-tools --csv output.csv\n")
	sb.WriteString("row,col,level,conductance_uS,resistance_ohm,program_V\n")

	shown := 0
	for r := 0; r < rows && shown < maxPreview; r++ {
		for c := 0; c < cols && shown < maxPreview; c++ {
			// Linear level gradient across cells (row-major).
			cellIdx := r*cols + c
			denom := max(total-1, 1)
			level := int(float64(cellIdx) / float64(denom) * float64(quantLevels-1))
			gNorm := float64(level) / float64(quantLevels-1) // [0,1]
			gUS := gMin + gNorm*(gMax-gMin)                  // µS
			rOhm := 1e6 / gUS                                // Ω (1/G, G in S = G_µS * 1e-6)
			progV := vMin + gNorm*(vMax-vMin)
			fmt.Fprintf(&sb, "%d,%d,%d,%.4f,%.2f,%.4f\n", r, c, level, gUS, rOhm, progV)
			shown++
		}
	}
	if total > maxPreview {
		fmt.Fprintf(&sb, "# ... (%d more rows not shown — %d total cells)\n", total-maxPreview, total)
	}
	return sb.String()
}

// generateArrayStatistics produces a concise complexity and feasibility report
// for the configured crossbar array. Values are computed from config only (no simulation).
func generateArrayStatistics(cfg *config.ArrayConfig) string {
	const quantLevels = 30
	const bitsPerLevel = 4 // ≈ log2(30) ≈ 4.9, rounded for display

	rows, cols := cfg.Rows, cfg.Cols
	if rows <= 0 {
		rows = 1
	}
	if cols <= 0 {
		cols = 1
	}
	totalCells := rows * cols
	dieW := float64(cols) * cfg.CellWidth
	dieH := float64(rows) * cfg.CellHeight
	dieArea := dieW * dieH

	// Memory capacity in bits and bytes.
	capacityBits := int64(totalCells) * bitsPerLevel
	capacityKB := float64(capacityBits) / 8 / 1024

	// Sneak path risk: in a passive crossbar, writing one cell half-selects
	// (rows-1) same-column cells and (cols-1) same-row cells = (rows+cols-2) half-select paths.
	// The (rows-1)*(cols-1) unselected cells can also form parasitic paths.
	// 1T1R eliminates column sneak, 2T1R eliminates both.
	sneakRisk := ""
	switch cfg.Architecture {
	case "passive":
		halfSelect := (rows - 1) + (cols - 1) // directly half-selected per write
		parasitic := (rows - 1) * (cols - 1)  // unselected cells forming parasitic paths
		if rows > 32 || cols > 32 { // > recommended passive limit (32×32)
			sneakRisk = fmt.Sprintf("HIGH (%d half-select + %d parasitic paths per write — use 1T1R or 2T1R)", halfSelect, parasitic)
		} else if rows > 16 || cols > 16 {
			sneakRisk = fmt.Sprintf("MODERATE (%d half-select paths per write — consider 1T1R)", halfSelect)
		} else {
			sneakRisk = fmt.Sprintf("LOW (%d half-select paths per write — acceptable for passive)", halfSelect)
		}
	case "1t1r":
		// WL gates each row's transistor. Writing row r: only row-r transistors ON.
		// Column-direction sneak (different rows) eliminated; same-row half-select remain.
		sneakRisk = fmt.Sprintf("ROW-ONLY (%d same-row half-select paths remain; column sneak eliminated by transistors)", cols-1)
	case "2t1r":
		sneakRisk = "NONE (both row and column transistors isolate cells)"
	default:
		sneakRisk = "Unknown"
	}

	// Word line and bit line wire resistance (rough estimate: sky130 M1 ~0.04 Ω/µm).
	const metalRes = 0.04 // Ω/µm
	wlRes := float64(cols) * cfg.CellWidth * metalRes
	blRes := float64(rows) * cfg.CellHeight * metalRes

	// Recommended maximum array size per architecture.
	recMax := ""
	switch cfg.Architecture {
	case "passive":
		recMax = "≤32×32 (sneak path limited)"
		if rows > 32 || cols > 32 {
			recMax = "⚠ EXCEEDS recommended 32×32 for passive — sneak paths likely dominant"
		}
	case "1t1r":
		recMax = "≤128×128 (row transistor isolates sneak paths)"
		if rows > 128 || cols > 128 {
			recMax = "⚠ EXCEEDS recommended 128×128 for 1T1R"
		}
	case "2t1r":
		recMax = "≤512×512 (dual transistors provide full isolation)"
		if rows > 512 || cols > 512 {
			recMax = "⚠ EXCEEDS recommended 512×512 for 2T1R"
		}
	}

	out := fmt.Sprintf(`Array Statistics Report
═══════════════════════════════════════════════════════

CONFIGURATION
  Architecture:   %s
  Mode:           %s
  Technology:     %s
  Array size:     %d rows × %d columns
  Cell size:      %.3f µm × %.3f µm

PHYSICAL DIMENSIONS
  Die width:      %.2f µm
  Die height:     %.2f µm
  Die area:       %.4f µm²  (%.6f mm²)
  WL length:      %.2f µm  (R_sheet ≈ %.2f Ω)
  BL length:      %.2f µm  (R_sheet ≈ %.2f Ω)

CAPACITY
  Total cells:    %d
  States/cell:    %d (%d bits/cell)
  Capacity:       %d bits  (%.2f KB)

SNEAK PATH RISK
  Assessment:     %s

SCALABILITY
  Recommended:    %s

NOTES
  • Cell resistance, WL/BL RC delay, and IR-drop require SPICE simulation.
  • Liberty timing values in this tool are placeholders — not from SPICE char.
  • Capacity estimate uses log2(%d) ≈ %d bits/cell; actual encoding may vary.
  • Sneak path risk counts: (rows+cols-2) directly half-selected cells +
    (rows-1)×(cols-1) parasitic paths per write. Actual impact depends on
    the on/off ratio (ION/IOFF) of the FeFET device.
`,
		cfg.Architecture, cfg.Mode, func() string {
			if cfg.Technology == "" {
				return "sky130"
			}
			return cfg.Technology
		}(),
		rows, cols,
		cfg.CellWidth, cfg.CellHeight,
		dieW, dieH,
		dieArea, dieArea/1e6,
		float64(cols)*cfg.CellWidth, wlRes,
		float64(rows)*cfg.CellHeight, blRes,
		totalCells, quantLevels, bitsPerLevel,
		capacityBits, capacityKB,
		sneakRisk,
		recMax,
		quantLevels, bitsPerLevel,
	)
	return out
}

// generateExportManifest produces a human-readable list of all files that would
// be created by "Export Package", grouped by category, with brief descriptions.
// It uses the current cfg to derive design name, architecture, and technology.
func generateExportManifest(cfg *config.ArrayConfig) string {
	design := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	dir := fmt.Sprintf("data/%s/", design)

	tech := cfg.Technology
	if tech == "" {
		tech = "sky130"
	}
	arch := cfg.Architecture
	if arch == "" {
		arch = "passive"
	}

	cellName := "fecim_bitcell"
	switch arch {
	case "1t1r":
		cellName = "fecim_1t1r_bitcell"
	case "2t1r":
		cellName = "fecim_2t1r_bitcell"
	}

	out := fmt.Sprintf(`Export Package Manifest
═══════════════════════════════════════════════════════
Design:      %s
Technology:  %s
Architecture: %s
Output dir:  %s

CELL LIBRARY  (cells/)
  %-42s  Abstract view (MACRO, PIN, LAYER)
  %-42s  Timing/power model (placeholder — not signoff)
  %-42s  Behavioral Verilog (Yosys hierarchy blackbox)

DESIGN FILES
  %-42s  Structural array Verilog netlist
  %-42s  Physical placement (FIXED — no routing needed)
  %-42s  Physical/electrical/timing report

FLOW SCRIPTS (physical design)
  %-42s  LibreLane / OpenLane v2 configuration
  %-42s  SDC timing constraints (BASE_SDC_FILE)
  %-42s  Yosys hierarchy check
  %-42s  KLayout DEF+LEF → GDSII stream-out
  %-42s  OpenROAD placement validation
  %-42s  OpenSTA standalone timing analysis
  %-42s  OpenLane v1 config (legacy TCL format)
  %-42s  OpenLane v1 macro placement constraints
  %-42s  Full flow (Yosys→KLayout→OpenROAD→LibreLane)

VERIFICATION SCRIPTS
  %-42s  Magic DRC (Design Rule Check)
  %-42s  Netgen LVS (Layout vs. Schematic)

SIMULATION SCRIPTS
  %-42s  CrossSim hardware-accurate MVM config
  %-42s  CrossSim Python runner
  %-42s  PySpice/Ngspice crossbar simulation
  %-42s  OpenVAF Verilog-A L-K compact model

METADATA
  %-42s  Machine-readable design parameters (JSON)
  %-42s  Setup instructions and quick start

TOTAL FILES: 23  (3 cell library + 3 design + 9 flow + 2 verify + 4 sim + 2 meta)

NOTES
  • Liberty timing values are structural placeholders.
    Real signoff requires SPICE characterization with a validated FeFET model.
  • DEF placement is FIXED — no floorplan/placement run required.
    OpenROAD validates the pre-placed structure only.
  • Run:  cd %s && ./run_flow.sh
`,
		design, tech, arch, dir,
		"cells/"+cellName+".lef",
		"cells/"+cellName+".lib",
		"cells/"+cellName+".v",
		design+".v",
		design+".def",
		"design_summary.txt",
		"config.json",
		"constraints.sdc",
		"synthesis.tcl",
		"gen_gds.py",
		"openroad_flow.tcl",
		"opensta_check.tcl",
		"config.tcl",
		"macros.cfg",
		"run_flow.sh",
		"run_drc.sh",
		"run_lvs.sh",
		"crosssim.yaml",
		"run_crosssim.py",
		"run_pyspice.py",
		"fecim_lk.va",
		design+".json",
		"README.md",
		dir,
	)
	return out
}

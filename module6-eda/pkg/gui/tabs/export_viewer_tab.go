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
// Header: row,col,level,conductance_uS,resistance_ohm,program_V
// Constants: GMin=10 µS, GMax=100 µS, VProgMin=2.0 V, VProgMax=5.0 V (defaults).
func generateCSVPreview(cfg *config.ArrayConfig) string {
	const quantLevels = 30
	const gMin = 10.0  // µS (GMin = 10e-6 S)
	const gMax = 100.0 // µS (GMax = 100e-6 S)
	const vMin = 2.0   // V (VProgMin default)
	const vMax = 5.0   // V (VProgMax default)
	const maxPreview = 16 // rows to show before truncating

	rows, cols := cfg.Rows, cfg.Cols
	if rows <= 0 {
		rows = 4
	}
	if cols <= 0 {
		cols = 4
	}
	total := rows * cols

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# FeCIM Crossbar CSV Export — %dx%d array, mode=%s\n", rows, cols, cfg.Mode))
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

	// Sneak path risk: passive crossbars scale as N² paths per write.
	// 1T1R eliminates row sneak, 2T1R eliminates both.
	sneakRisk := ""
	switch cfg.Architecture {
	case "passive":
		pathsPerWrite := rows * cols
		if pathsPerWrite > 32*32 { // > recommended passive limit (32×32)
			sneakRisk = fmt.Sprintf("HIGH (%d×%d = %d sneak paths per write — use 1T1R or 2T1R)", rows, cols, pathsPerWrite)
		} else if pathsPerWrite > 16*16 {
			sneakRisk = fmt.Sprintf("MODERATE (%d paths per write — consider 1T1R)", pathsPerWrite)
		} else {
			sneakRisk = fmt.Sprintf("LOW (%d paths per write — acceptable for passive)", pathsPerWrite)
		}
	case "1t1r":
		sneakRisk = fmt.Sprintf("ROW-ONLY (%d paths suppressed by row transistors)", rows)
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
  • Sneak path risk is a rough structural estimate; actual impact depends on
    on/off ratio (ION/IOFF) of the FeFET device.
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
  %-38s  Abstract view (MACRO, PIN, LAYER)
  %-38s  Timing/power model (placeholder — not signoff)
  %-38s  Behavioral Verilog (Yosys hierarchy blackbox)

DESIGN FILES
  %-38s  Structural array Verilog netlist
  %-38s  Physical placement (FIXED — no routing needed)
  %-38s  Physical/electrical/timing report

FLOW SCRIPTS
  %-38s  LibreLane / OpenLane v1 configuration
  %-38s  SDC timing constraints (BASE_SDC_FILE)
  %-38s  Yosys hierarchy check
  %-38s  KLayout DEF+LEF → GDSII stream-out
  %-38s  OpenROAD placement validation
  %-38s  Full flow orchestration (Yosys→KLayout→OpenROAD→LibreLane)

METADATA
  %-38s  Machine-readable design parameters (JSON)
  %-38s  Setup instructions and quick start

TOTAL FILES: 14

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
		"run_flow.sh",
		design+".json",
		"README.md",
		dir,
	)
	return out
}

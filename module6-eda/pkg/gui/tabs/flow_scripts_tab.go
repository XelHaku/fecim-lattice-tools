// pkg/gui/tabs/flow_scripts_tab.go
// Flow Scripts tab: shows generated EDA tool scripts for the open-source RTL-to-GDS flow.
//
// Scripts generated cover:
//   Design Summary    — formatted design report (area, electrical, timing, validation)
//   Yosys TCL         — hierarchy check / synthesis
//   OpenROAD TCL      — placement check + timing
//   KLayout Python    — DEF+LEF → GDS II
//   OpenSTA TCL       — timing analysis
//   SDC               — timing constraints (required by OpenLane/LibreLane)
//   LibreLane JSON    — extended config.json for LibreLane/OpenLane v2
//   OpenLane v1 TCL   — config.tcl for OpenLane v1 (set ::env(...) format)
//   Shell runner      — orchestrates all steps
//   Netgen LVS        — Layout vs. Schematic verification script
//   Magic DRC         — Design Rule Check script
//   CrossSim YAML     — CrossSim crossbar simulation config
//   CrossSim Python   — CrossSim runner script
//   PySpice Python    — Ngspice crossbar simulation via PySpice
//   OpenVAF Verilog-A — FeCIM L-K compact model (compile with OpenVAF)
package tabs

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

var flowScriptFormats = []string{
	// Summary report
	"Design Summary (design_summary.txt)",
	// Physical design flow
	"Yosys TCL (synthesis.tcl)",
	"OpenROAD TCL (openroad_flow.tcl)",
	"KLayout Python (gen_gds.py)",
	"OpenSTA TCL (opensta_check.tcl)",
	"SDC Constraints (constraints.sdc)",
	"LibreLane JSON (config.json)",
	"OpenLane v1 TCL (config.tcl)",
	"OpenLane v1 Macro Placement (macros.cfg)",
	"Shell Runner (run_flow.sh)",
	// Verification
	"Netgen LVS Script (run_lvs.sh)",
	"Magic DRC Script (run_drc.sh)",
	// Simulation
	"CrossSim YAML (crosssim.yaml)",
	"CrossSim Python (run_crosssim.py)",
	"PySpice Simulation (run_pyspice.py)",
	"OpenVAF Verilog-A (fecim_lk.va)",
}

// MakeFlowScriptsTab creates a tab that shows generated flow scripts.
// Users can select a script format, view its content, and copy it to clipboard.
func MakeFlowScriptsTab(cfg *config.ArrayConfig, window fyne.Window) fyne.CanvasObject {
	if cfg == nil {
		cfg = &config.ArrayConfig{
			Rows: 4, Cols: 4,
			Mode: "storage", Architecture: "passive",
			CellWidth: 0.46, CellHeight: 2.72,
		}
	}

	formatSelect := widget.NewSelect(flowScriptFormats, nil)
	formatSelect.SetSelected(flowScriptFormats[0])

	status := widget.NewLabel("Select a format to preview the script")
	preview := widget.NewMultiLineEntry()
	preview.Wrapping = fyne.TextWrapOff
	preview.TextStyle.Monospace = true
	preview.Disable()

	refresh := func() {
		content, desc := loadFlowScriptContent(formatSelect.Selected, cfg)
		preview.SetText(content)
		status.SetText(desc)
	}

	formatSelect.OnChanged = func(string) { refresh() }

	refreshBtn := widget.NewButton("Refresh", refresh)

	copyBtn := widget.NewButton("Copy to Clipboard", func() {
		if window == nil {
			return
		}
		window.Clipboard().SetContent(preview.Text)
		dialog.ShowInformation("Copied", "Script copied to clipboard.", window)
	})

	saveBtn := widget.NewButton("Save to File…", func() {
		if window == nil {
			return
		}
		ext := flowScriptExtension(formatSelect.Selected)
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

	toolNote := widget.NewLabel("Tools: yosys · klayout · openroad · librelane · magic · netgen · ngspice · PySpice · CrossSim · openvaf")
	toolNote.Wrapping = fyne.TextWrapWord

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Script:"),
			formatSelect,
			refreshBtn,
			copyBtn,
			saveBtn,
		),
		status,
		toolNote,
		widget.NewSeparator(),
	)

	refresh()

	return container.NewBorder(header, nil, nil, nil, container.NewScroll(preview))
}

// flowScriptExtension returns a suitable file extension for the given script format.
func flowScriptExtension(format string) string {
	switch format {
	case "Design Summary (design_summary.txt)":
		return ".txt"
	case "Yosys TCL (synthesis.tcl)", "OpenROAD TCL (openroad_flow.tcl)", "OpenSTA TCL (opensta_check.tcl)":
		return ".tcl"
	case "KLayout Python (gen_gds.py)", "CrossSim Python (run_crosssim.py)", "PySpice Simulation (run_pyspice.py)":
		return ".py"
	case "SDC Constraints (constraints.sdc)":
		return ".sdc"
	case "LibreLane JSON (config.json)":
		return ".json"
	case "OpenLane v1 TCL (config.tcl)":
		return ".tcl"
	case "OpenLane v1 Macro Placement (macros.cfg)":
		return ".cfg"
	case "Shell Runner (run_flow.sh)", "Netgen LVS Script (run_lvs.sh)", "Magic DRC Script (run_drc.sh)":
		return ".sh"
	case "CrossSim YAML (crosssim.yaml)":
		return ".yaml"
	case "OpenVAF Verilog-A (fecim_lk.va)":
		return ".va"
	default:
		return ".txt"
	}
}

// loadFlowScriptContent generates or loads the script for the given format.
func loadFlowScriptContent(format string, cfg *config.ArrayConfig) (content, desc string) {
	// Default cell config used for single-cell generators (netgen, magic, OpenVAF)
	cellCfg := config.DefaultCellConfig()
	cellCfg.CellType = cfg.Architecture
	cellCfg.Technology = cfg.Technology

	switch format {
	// ── Summary ────────────────────────────────────────────────────────────
	case "Design Summary (design_summary.txt)":
		return export.GenerateDesignSummary(*cfg),
			"Array design summary — physical, electrical, and timing parameters for the configured crossbar"

	// ── Physical design flow ───────────────────────────────────────────────
	case "Yosys TCL (synthesis.tcl)":
		return export.GenerateSynthesisScript(*cfg),
			"Yosys hierarchy check — validates structural Verilog (run: yosys synth.tcl)"

	case "OpenROAD TCL (openroad_flow.tcl)":
		return export.GenerateOpenROADFlowScript(*cfg),
			"OpenROAD placement check + timing — verify pre-placed DEF (run: openroad -exit openroad_flow.tcl)"

	case "KLayout Python (gen_gds.py)":
		return export.GenerateKLayoutGDSScript(*cfg),
			"KLayout DEF+LEF → GDS II — generates EXTRA_GDS_FILES for OpenLane (run: klayout -z -r gen_gds.py)"

	case "OpenSTA TCL (opensta_check.tcl)":
		return export.GenerateOpenSTAScript(*cfg),
			"OpenSTA timing analysis — clockless design, no violations expected (run: opensta < opensta_check.tcl)"

	case "SDC Constraints (constraints.sdc)":
		return export.GenerateSDC(*cfg),
			"SDC timing constraints — required by OpenLane/LibreLane (BASE_SDC_FILE)"

	case "LibreLane JSON (config.json)":
		return export.GenerateLibreLaneConfig(*cfg),
			"LibreLane config.json — recommended for new designs (pip install librelane)"

	case "OpenLane v1 TCL (config.tcl)":
		return export.GenerateOpenLaneTCLConfig(*cfg),
			"OpenLane v1 config.tcl — set ::env() variable format for TCL-based flow.tcl (OpenLane v1)"

	case "OpenLane v1 Macro Placement (macros.cfg)":
		return export.GenerateOpenLaneTCLMacroPlacement(*cfg),
			"OpenLane v1 macro placement constraints — pins FeCIM array at die center (MACRO_PLACEMENT_CFG)"

	case "Shell Runner (run_flow.sh)":
		return export.GenerateFlowRunner(*cfg),
			"Shell runner — orchestrates Yosys → KLayout → OpenROAD → LibreLane (chmod +x run_flow.sh && ./run_flow.sh)"

	// ── Verification ──────────────────────────────────────────────────────
	case "Netgen LVS Script (run_lvs.sh)":
		return export.GenerateNetgenLVSScript(cellCfg),
			"Netgen LVS — compares schematic SPICE vs extracted layout SPICE (run: bash run_lvs.sh)"

	case "Magic DRC Script (run_drc.sh)":
		return export.GenerateMagicDRCScript(*cfg),
			"Magic DRC — runs Design Rule Check on layout in batch mode (run: bash run_drc.sh)"

	// ── Simulation ────────────────────────────────────────────────────────
	case "CrossSim YAML (crosssim.yaml)":
		return export.GenerateCrossSIMConfig(*cfg),
			"CrossSim YAML config — hardware-accurate MVM simulation (Sandia CrossSim v3.1)"

	case "CrossSim Python (run_crosssim.py)":
		return export.GenerateCrossSIMRunScript(*cfg),
			"CrossSim runner — loads YAML config and runs AnalogCore MVM (run: python3 run_crosssim.py)"

	case "PySpice Simulation (run_pyspice.py)":
		return export.GeneratePySpiceScript(*cfg),
			"PySpice/Ngspice crossbar simulation — builds resistive MVM netlist and runs DC analysis"

	case "OpenVAF Verilog-A (fecim_lk.va)":
		return export.GenerateOpenVAFVerilogA(cellCfg),
			"Verilog-A L-K compact model — compile with OpenVAF for Ngspice OSDI simulation"

	default:
		return "", "unknown script format"
	}
}

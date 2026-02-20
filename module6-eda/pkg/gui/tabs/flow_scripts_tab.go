// pkg/gui/tabs/flow_scripts_tab.go
// Flow Scripts tab: shows generated Yosys/OpenROAD/KLayout/SDC/LibreLane scripts
//
// This tab provides ready-to-run scripts for the complete open-source
// RTL-to-GDS flow. Users can copy scripts directly from the viewer.
//
// Scripts generated (all from pkg/export/scripts.go + sdc.go):
//   Yosys TCL      — hierarchy check / synthesis (run: yosys synth.tcl)
//   OpenROAD TCL   — placement check + timing (run: openroad -exit openroad_flow.tcl)
//   KLayout Python — DEF+LEF → GDS II (run: klayout -z -r gen_gds.py)
//   OpenSTA TCL    — timing analysis (run: opensta < opensta_check.tcl)
//   SDC            — timing constraints (required by OpenLane/LibreLane)
//   LibreLane JSON — extended config.json for LibreLane/OpenLane v1
//   Shell runner   — orchestrates all steps (run: bash run_flow.sh)
package tabs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

var flowScriptFormats = []string{
	"Yosys TCL (synthesis.tcl)",
	"OpenROAD TCL (openroad_flow.tcl)",
	"KLayout Python (gen_gds.py)",
	"OpenSTA TCL (opensta_check.tcl)",
	"SDC Constraints (constraints.sdc)",
	"LibreLane JSON (config.json)",
	"Shell Runner (run_flow.sh)",
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
		window.Clipboard().SetContent(preview.Text)
		dialog.ShowInformation("Copied", "Script copied to clipboard.", window)
	})

	toolNote := widget.NewLabel("Scripts require: yosys · klayout · openroad · librelane (pip install librelane)")
	toolNote.Wrapping = fyne.TextWrapWord

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Script:"),
			formatSelect,
			refreshBtn,
			copyBtn,
		),
		status,
		toolNote,
		widget.NewSeparator(),
	)

	refresh()

	return container.NewBorder(header, nil, nil, nil, container.NewScroll(preview))
}

// loadFlowScriptContent generates or loads the script for the given format.
func loadFlowScriptContent(format string, cfg *config.ArrayConfig) (content, desc string) {
	switch format {
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

	case "Shell Runner (run_flow.sh)":
		return export.GenerateFlowRunner(*cfg),
			"Shell runner — orchestrates Yosys → KLayout → OpenROAD → LibreLane (chmod +x run_flow.sh && ./run_flow.sh)"

	default:
		return "", "unknown script format"
	}
}

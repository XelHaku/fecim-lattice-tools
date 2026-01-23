// pkg/gui/tabs/hdl_tab.go
// HDL Generation tab for FeCIM Design Suite
// Generates Verilog netlist and DEF placement files from compiled CrossbarMapping

package tabs

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/compiler"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/export"
)

// MakeHDLTab creates the HDL generation tab with split view
func MakeHDLTab(state interface{}, w fyne.Window) fyne.CanvasObject {
	appState := state.(*AppState)

	// Code display labels - use monospace styling
	verilogLabel := widget.NewLabel("// Verilog code will appear here\n// Compile weights first, then click 'Generate HDL'")
	verilogLabel.Wrapping = fyne.TextWrapOff
	verilogLabel.TextStyle = fyne.TextStyle{Monospace: true}

	defLabel := widget.NewLabel("# DEF placement will appear here\n# Compile weights first, then click 'Generate HDL'")
	defLabel.Wrapping = fyne.TextWrapOff
	defLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Status label
	statusLabel := widget.NewLabel("Status: Ready - compile weights and generate HDL")
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Architecture selection
	archSelect := widget.NewSelect([]string{"Passive", "1T1R"}, nil)
	archSelect.SetSelected("Passive")

	// Generate button
	generateButton := widget.NewButtonWithIcon("Generate HDL", theme.MediaPlayIcon(), func() {
		if appState.CurrentMapping == nil {
			dialog.ShowError(fmt.Errorf("no compiled mapping available - compile weights first"), w)
			return
		}

		// Update architecture based on selection
		mapping := appState.CurrentMapping
		if archSelect.Selected == "1T1R" {
			mapping.Config.Architecture = compiler.Arch1T1R
			mapping.Config.CellPitch = 0.92 // Larger cell for 1T1R
		} else {
			mapping.Config.Architecture = compiler.ArchPassive
			mapping.Config.CellPitch = 0.46 // Default passive cell
		}

		// Generate Verilog
		verilogContent := export.GenerateVerilogWithDefaults(mapping)
		verilogLabel.SetText(verilogContent)

		// Generate DEF
		defContent := export.GenerateDEFWithDefaults(mapping)
		defLabel.SetText(defContent)

		// Update status
		arch := mapping.Config.Architecture
		if arch == "" {
			arch = "passive"
		}
		statusLabel.SetText(fmt.Sprintf("Status: Generated %s architecture - %d cells, %d rows × %d cols",
			arch, len(mapping.Cells), getMaxRow(mapping)+1, getMaxCol(mapping)+1))
	})

	// Export to files button
	exportButton := widget.NewButtonWithIcon("Export to Files", theme.DocumentSaveIcon(), func() {
		if appState.CurrentMapping == nil {
			dialog.ShowError(fmt.Errorf("no compiled mapping available - compile weights first"), w)
			return
		}

		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return // User cancelled
			}

			dir := uri.Path()
			mapping := appState.CurrentMapping

			// Update architecture based on selection
			if archSelect.Selected == "1T1R" {
				mapping.Config.Architecture = compiler.Arch1T1R
				mapping.Config.CellPitch = 0.92
			} else {
				mapping.Config.Architecture = compiler.ArchPassive
				mapping.Config.CellPitch = 0.46
			}

			var exported []string
			var errors []string

			// Export Verilog
			verilogPath := filepath.Join(dir, "lattice.v")
			if err := export.ExportVerilog(mapping, verilogPath); err != nil {
				errors = append(errors, fmt.Sprintf("Verilog: %v", err))
			} else {
				exported = append(exported, "lattice.v")
			}

			// Export DEF
			defPath := filepath.Join(dir, "placement.def")
			if err := export.ExportDEF(mapping, defPath); err != nil {
				errors = append(errors, fmt.Sprintf("DEF: %v", err))
			} else {
				exported = append(exported, "placement.def")
			}

			// Export OpenLane config
			configPath := filepath.Join(dir, "config.tcl")
			if err := writeOpenLaneConfig(mapping, configPath); err != nil {
				errors = append(errors, fmt.Sprintf("Config: %v", err))
			} else {
				exported = append(exported, "config.tcl")
			}

			// Show results
			if len(errors) > 0 {
				dialog.ShowError(fmt.Errorf("export errors: %v", errors), w)
			} else {
				dialog.ShowInformation("Export Complete",
					fmt.Sprintf("Exported to %s:\n• %s", dir, joinStrings(exported, "\n• ")), w)
			}

			statusLabel.SetText(fmt.Sprintf("Status: Exported %d files to %s", len(exported), dir))
		}, w)
	})

	// Cell grid visualization button
	showGridButton := widget.NewButtonWithIcon("Show Cell Grid", theme.GridIcon(), func() {
		if appState.CurrentMapping == nil {
			dialog.ShowError(fmt.Errorf("no compiled mapping available"), w)
			return
		}
		showCellGridDialog(appState.CurrentMapping, w)
	})

	// Control section
	archLabel := widget.NewLabel("Architecture:")
	controls := container.NewHBox(
		generateButton,
		widget.NewSeparator(),
		archLabel,
		archSelect,
		widget.NewSeparator(),
		exportButton,
		showGridButton,
	)

	// Headers for split view
	verilogHeader := widget.NewLabelWithStyle("Verilog Netlist", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	defHeader := widget.NewLabelWithStyle("DEF Placement", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Split view with scrollable code areas
	verilogScroll := container.NewScroll(verilogLabel)
	verilogScroll.SetMinSize(fyne.NewSize(400, 400))

	defScroll := container.NewScroll(defLabel)
	defScroll.SetMinSize(fyne.NewSize(400, 400))

	// Left panel (Verilog)
	leftPanel := container.NewBorder(
		verilogHeader, nil, nil, nil,
		verilogScroll,
	)

	// Right panel (DEF)
	rightPanel := container.NewBorder(
		defHeader, nil, nil, nil,
		defScroll,
	)

	// Split view using grid
	splitView := container.NewGridWithColumns(2,
		leftPanel,
		rightPanel,
	)

	// Main layout
	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("HDL Generation", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			controls,
			widget.NewSeparator(),
			statusLabel,
		),
		nil, nil, nil,
		splitView,
	)

	return content
}

// getMaxRow returns the maximum row index from mapping
func getMaxRow(mapping *compiler.CrossbarMapping) int {
	maxRow := 0
	for _, cell := range mapping.Cells {
		if cell.Row > maxRow {
			maxRow = cell.Row
		}
	}
	return maxRow
}

// getMaxCol returns the maximum column index from mapping
func getMaxCol(mapping *compiler.CrossbarMapping) int {
	maxCol := 0
	for _, cell := range mapping.Cells {
		if cell.Col > maxCol {
			maxCol = cell.Col
		}
	}
	return maxCol
}

// joinStrings joins strings with separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// writeOpenLaneConfig writes an OpenLane configuration file
func writeOpenLaneConfig(mapping *compiler.CrossbarMapping, path string) error {
	maxRow := getMaxRow(mapping)
	maxCol := getMaxCol(mapping)
	numRows := maxRow + 1
	numCols := maxCol + 1

	// Calculate die area
	cellPitch := mapping.Config.CellPitch
	if cellPitch == 0 {
		cellPitch = 0.46
	}
	rowHeight := mapping.Config.RowHeight
	if rowHeight == 0 {
		rowHeight = 2.72
	}

	arrayWidth := float64(numCols)*cellPitch + 30.0  // 10um origin + 20um margin
	arrayHeight := float64(numRows)*rowHeight + 30.0 // 10um origin + 20um margin

	content := fmt.Sprintf(`# OpenLane Configuration for FeCIM Crossbar
# Generated by Demo 6: FeCIM Design Suite
# Architecture: %s

# Design name
set ::env(DESIGN_NAME) "fecim_crossbar"

# Input files
set ::env(VERILOG_FILES) [glob $::env(DESIGN_DIR)/lattice.v]
set ::env(CURRENT_DEF) $::env(DESIGN_DIR)/placement.def

# Technology
set ::env(PDK) "sky130A"
set ::env(STD_CELL_LIBRARY) "sky130_fd_sc_hd"

# Die area (microns)
set ::env(FP_SIZING) "absolute"
set ::env(DIE_AREA) "0 0 %.2f %.2f"

# Placement
set ::env(PL_TARGET_DENSITY) 0.5
set ::env(PL_RANDOM_GLB_PLACEMENT) 0

# Routing
set ::env(ROUTING_CORES) 4
set ::env(GLB_RT_ADJUSTMENT) 0.1

# Power
set ::env(VDD_NETS) "VDD"
set ::env(GND_NETS) "VSS"

# Array info
# Rows: %d, Cols: %d, Cells: %d
# Cell pitch: %.2f um, Row height: %.2f um
`, mapping.Config.Architecture, arrayWidth, arrayHeight, numRows, numCols, len(mapping.Cells), cellPitch, rowHeight)

	return os.WriteFile(path, []byte(content), 0644)
}

// showCellGridDialog displays a dialog with cell placement coordinates
func showCellGridDialog(mapping *compiler.CrossbarMapping, w fyne.Window) {
	maxRow := getMaxRow(mapping)
	maxCol := getMaxCol(mapping)
	numRows := maxRow + 1
	numCols := maxCol + 1

	// Build grid content
	var gridContent string
	gridContent = fmt.Sprintf("Cell Grid: %d×%d (%d cells)\n", numRows, numCols, len(mapping.Cells))
	gridContent += fmt.Sprintf("Architecture: %s\n\n", mapping.Config.Architecture)
	gridContent += "Cell Placement (row, col) → coordinates:\n"
	gridContent += "─────────────────────────────────────────\n"

	cellPitch := mapping.Config.CellPitch
	if cellPitch == 0 {
		cellPitch = 0.46
	}
	rowHeight := mapping.Config.RowHeight
	if rowHeight == 0 {
		rowHeight = 2.72
	}

	for _, cell := range mapping.Cells {
		x := 10.0 + float64(cell.Col)*cellPitch
		y := 10.0 + float64(cell.Row)*rowHeight
		gridContent += fmt.Sprintf("R_%d_%d → (%.2f, %.2f) μm, Level=%d, G=%.2f μS\n",
			cell.Row, cell.Col, x, y, cell.QuantLevel, cell.Conductance)
	}

	// Create scrollable label
	gridLabel := widget.NewLabel(gridContent)
	gridLabel.Wrapping = fyne.TextWrapOff
	gridLabel.TextStyle = fyne.TextStyle{Monospace: true}

	scrollContent := container.NewScroll(gridLabel)
	scrollContent.SetMinSize(fyne.NewSize(500, 400))

	dialog.ShowCustom("Cell Grid", "Close", scrollContent, w)
}

// pkg/gui/tabs/export_tab.go
package tabs

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/export"
	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/validate"
	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// MakeExportTab creates the export tab UI
func MakeExportTab(state interface{}, w fyne.Window) fyne.CanvasObject {
	appState := state.(*AppState)

	// Export format checkboxes
	jsonCheck := widget.NewCheck("JSON (mapping)", nil)
	jsonCheck.SetChecked(true)

	csvCheck := widget.NewCheck("CSV (cell data)", nil)
	csvCheck.SetChecked(true)

	spiceCheck := widget.NewCheck("SPICE (circuit)", nil)
	spiceCheck.SetChecked(true)

	defCheck := widget.NewCheck("DEF (placement)", nil)
	defCheck.SetChecked(true)

	verilogCheck := widget.NewCheck("Verilog (netlist)", nil)
	verilogCheck.SetChecked(true)

	optionsSection := container.NewVBox(
		widget.NewLabel("EXPORT FORMATS"),
		jsonCheck,
		csvCheck,
		spiceCheck,
		defCheck,
		verilogCheck,
	)

	// SPICE VDD setting
	vddEntry := widget.NewEntry()
	vddEntry.SetText("1.8")
	vddEntry.SetPlaceHolder("VDD voltage")

	spiceSettings := container.NewHBox(
		widget.NewLabel("SPICE VDD (V):"),
		vddEntry,
	)

	// Results
	resultsLabel := widget.NewLabel("No exports yet...")
	resultsLabel.Wrapping = fyne.TextWrapWord

	// Validation Output
	validateOutput := widget.NewMultiLineEntry()
	validateOutput.Disable() // Read-only
	validateOutput.SetText("Validation ready...")
	validateOutput.SetMinRowsVisible(8)

	var lastVerilogPath string

	// Validate button logic
	validateButton := widget.NewButton("VALIDATE (Yosys)", func() {
		sharedwidgets.DebugInteraction("Export VALIDATE button pressed")
		if lastVerilogPath == "" {
			dialog.ShowError(fmt.Errorf("please export Verilog first"), w)
			return
		}

		validateOutput.SetText("Running Yosys check on " + lastVerilogPath + "...")

		// Run validation
		output, err := validate.RunYosysCheck(lastVerilogPath)
		if err != nil {
			validateOutput.SetText(fmt.Sprintf("VALIDATION FAILED:\n%v\n\nOutput:\n%s", err, output))
		} else {
			validateOutput.SetText(fmt.Sprintf("VALIDATION SUCCESS:\n%s", output))
		}
	})
	validateButton.Disable() // Enable only after export

	// Export button
	exportButton := widget.NewButton("EXPORT FILES", func() {
		sharedwidgets.DebugInteraction("Export EXPORT FILES button pressed")
		if appState.CurrentMapping == nil {
			dialog.ShowError(fmt.Errorf("compile weights first (Tab 1)"), w)
			return
		}

		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}

			dir := uri.Path()
			var exported []string
			var errors []string

			// Export JSON
			if jsonCheck.Checked {
				path := filepath.Join(dir, "crossbar_mapping.json")
				if err := export.ExportJSON(appState.CurrentMapping, path); err != nil {
					errors = append(errors, fmt.Sprintf("JSON: %v", err))
				} else {
					exported = append(exported, path)
				}
			}

			// Export CSV
			if csvCheck.Checked {
				path := filepath.Join(dir, "cell_assignments.csv")
				if err := export.ExportCSV(appState.CurrentMapping, path); err != nil {
					errors = append(errors, fmt.Sprintf("CSV: %v", err))
				} else {
					exported = append(exported, path)
				}
			}

			// Export SPICE
			if spiceCheck.Checked {
				vdd := 1.8
				fmt.Sscanf(vddEntry.Text, "%f", &vdd)
				path := filepath.Join(dir, "crossbar.sp")
				if err := export.ExportSPICE(appState.CurrentMapping, path, vdd); err != nil {
					errors = append(errors, fmt.Sprintf("SPICE: %v", err))
				} else {
					exported = append(exported, path)
				}
			}

			// Export DEF
			if defCheck.Checked {
				path := filepath.Join(dir, "lattice.def")
				if err := export.ExportDEF(appState.CurrentMapping, path); err != nil {
					errors = append(errors, fmt.Sprintf("DEF: %v", err))
				} else {
					exported = append(exported, path)
				}
			}

			// Export Verilog
			if verilogCheck.Checked {
				path := filepath.Join(dir, "lattice.v")
				if err := export.ExportVerilog(appState.CurrentMapping, path); err != nil {
					errors = append(errors, fmt.Sprintf("Verilog: %v", err))
				} else {
					exported = append(exported, path)
					lastVerilogPath = path
					validateButton.Enable()
				}
			}

			// Show results
			result := fmt.Sprintf("EXPORT COMPLETE\n\nExported %d files:\n", len(exported))
			for _, f := range exported {
				result += fmt.Sprintf("  - %s\n", filepath.Base(f))
			}
			if len(errors) > 0 {
				result += "\nErrors:\n"
				for _, e := range errors {
					result += fmt.Sprintf("  - %s\n", e)
				}
			}
			resultsLabel.SetText(result)
		}, w)
	})

	// Status section
	statusSection := container.NewVBox(
		widget.NewLabel("STATUS"),
		resultsLabel,
		widget.NewSeparator(),
		widget.NewLabel("VALIDATION"),
		validateButton,
		container.NewScroll(validateOutput),
	)

	// Main layout
	content := container.NewVBox(
		optionsSection,
		widget.NewSeparator(),
		spiceSettings,
		widget.NewSeparator(),
		exportButton,
		widget.NewSeparator(),
		statusSection,
	)

	return container.NewScroll(content)
}

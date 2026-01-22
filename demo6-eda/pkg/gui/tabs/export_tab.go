// pkg/gui/tabs/export_tab.go
package tabs

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"demo6-eda/pkg/export"
)

// MakeExportTab creates the export tab UI
func MakeExportTab(state interface{}, w fyne.Window) fyne.CanvasObject {
	appState := state.(*AppState)

	// Export options
	jsonCheck := widget.NewCheck("JSON (mapping + stats)", nil)
	jsonCheck.SetChecked(true)

	csvCheck := widget.NewCheck("CSV (cell assignments)", nil)
	csvCheck.SetChecked(true)

	spiceCheck := widget.NewCheck("SPICE netlist (ngspice)", nil)
	spiceCheck.SetChecked(true)

	optionsSection := container.NewVBox(
		widget.NewLabel("EXPORT FORMATS"),
		jsonCheck,
		csvCheck,
		spiceCheck,
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

	// Export button
	exportButton := widget.NewButton("EXPORT FILES", func() {
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

			// Show results
			result := fmt.Sprintf("EXPORT COMPLETE\n\nExported %d files:\n", len(exported))
			for _, f := range exported {
				result += fmt.Sprintf("  - %s\n", f)
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

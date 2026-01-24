// pkg/gui/tabs/compiler_tab.go
package tabs

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/compiler"
	sharedwidgets "multilayer-ferroelectric-cim-visualizer/shared/widgets"
)

// MakeCompilerTab creates the array builder tab UI
func MakeCompilerTab(state interface{}, w fyne.Window) fyne.CanvasObject {
	// Get state
	appState := state.(*AppState)

	// Status section
	statusLabel := widget.NewLabel("Configure array geometry to begin.")
	statusLabel.Wrapping = fyne.TextWrapWord

	// Config section - Geometry
	rowsEntry := widget.NewEntry()
	rowsEntry.SetText("128")
	rowsEntry.SetPlaceHolder("Rows")

	colsEntry := widget.NewEntry()
	colsEntry.SetText("128")
	colsEntry.SetPlaceHolder("Cols")

	// Config section - Device Params
	levelsEntry := widget.NewEntry()
	levelsEntry.SetText("30") // Standard from transcript
	levelsEntry.SetPlaceHolder("Levels (2-30)")

	gMinEntry := widget.NewEntry()
	gMinEntry.SetText("1.0")
	gMinEntry.SetPlaceHolder("G_min (μS)")

	gMaxEntry := widget.NewEntry()
	gMaxEntry.SetText("100.0")
	gMaxEntry.SetPlaceHolder("G_max (μS)")

	vMinEntry := widget.NewEntry()
	vMinEntry.SetText("2.0")
	vMinEntry.SetPlaceHolder("V_prog_min (V)")

	vMaxEntry := widget.NewEntry()
	vMaxEntry.SetText("5.0")
	vMaxEntry.SetPlaceHolder("V_prog_max (V)")

	configForm := container.NewVBox(
		widget.NewLabel("ARRAY CONFIGURATION"),
		container.NewGridWithColumns(2,
			widget.NewLabel("Array Rows:"), rowsEntry,
			widget.NewLabel("Array Cols:"), colsEntry,
			widget.NewLabel("Analog Levels:"), levelsEntry,
			widget.NewLabel("G_min (μS):"), gMinEntry,
			widget.NewLabel("G_max (μS):"), gMaxEntry,
			widget.NewLabel("V_prog_min (V):"), vMinEntry,
			widget.NewLabel("V_prog_max (V):"), vMaxEntry,
		),
	)

	// Results section
	resultsLabel := widget.NewLabel("Generate design to see stats...")
	resultsLabel.Wrapping = fyne.TextWrapWord

	generateButton := widget.NewButton("GENERATE DESIGN", func() {
		sharedwidgets.DebugInteraction("Compiler GENERATE DESIGN button pressed")
		// Parse config
		rows, errR := strconv.Atoi(rowsEntry.Text)
		cols, errC := strconv.Atoi(colsEntry.Text)
		levels, errL := strconv.Atoi(levelsEntry.Text)
		gMin, _ := strconv.ParseFloat(gMinEntry.Text, 64)
		gMax, _ := strconv.ParseFloat(gMaxEntry.Text, 64)
		vMin, _ := strconv.ParseFloat(vMinEntry.Text, 64)
		vMax, _ := strconv.ParseFloat(vMaxEntry.Text, 64)

		if errR != nil || errC != nil || errL != nil {
			dialog.ShowError(fmt.Errorf("invalid numeric inputs"), w)
			return
		}

		config := compiler.CompileConfig{
			ArrayRows: rows,
			ArrayCols: cols,
			Levels:    levels,
			GMin:      gMin,
			GMax:      gMax,
			VProgMin:  vMin,
			VProgMax:  vMax,
			TPulse:    50.0,
		}

		// Compile (Generate Blank Array)
		// We pass nil for weights to trigger blank generation
		mapping, err := compiler.Compile(nil, config) 
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		appState.CurrentMapping = mapping
		appState.Compiled = true // "Compiled" means "Generated" in this context

		statusLabel.SetText(fmt.Sprintf("Design Generated: %dx%d (%d levels)", 
			rows, cols, levels))

		// Display results
		// Estimate area: cell pitch is roughly 0.5um x 0.5um = 0.25um^2 per cell
		// Total area = cells * 0.25e-6 mm^2
		areaMM2 := float64(mapping.Stats.TotalCells) * 0.25 * 1e-6 

		resultsLabel.SetText(fmt.Sprintf(
			"DESIGN STATISTICS\n\n"+
				"  Geometry:      %d x %d\n"+
				"  Total Cells:   %d\n"+
				"  Analog States: %d (30-layer Superlattice)\n"+
				"  Est. Area:     %.6f mm² (Core)\n"+
				"  Est. Power:    < 10 μW (Standby)\n",
			rows, cols, mapping.Stats.TotalCells,
			config.Levels, areaMM2))
	})

	resultsSection := container.NewVBox(
		generateButton,
		widget.NewSeparator(),
		resultsLabel,
	)

	// Main layout
	content := container.NewVBox(
		statusLabel,
		widget.NewSeparator(),
		configForm,
		widget.NewSeparator(),
		resultsSection,
	)

	return container.NewScroll(content)
}

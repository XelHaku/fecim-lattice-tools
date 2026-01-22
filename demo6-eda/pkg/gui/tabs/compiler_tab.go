// pkg/gui/tabs/compiler_tab.go
package tabs

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"demo6-eda/pkg/compiler"
)

// WeightsFile represents the JSON structure for loading weights
type WeightsFile struct {
	Name    string      `json:"name"`
	Rows    int         `json:"rows"`
	Cols    int         `json:"cols"`
	Weights [][]float64 `json:"weights"`
}

// MakeCompilerTab creates the compiler tab UI
func MakeCompilerTab(state interface{}, w fyne.Window) fyne.CanvasObject {
	// Get state
	appState := state.(*AppState)

	// Source section
	sourceLabel := widget.NewLabel("No weights loaded")
	sourceLabel.Wrapping = fyne.TextWrapWord

	var loadedWeights [][]float64

	loadButton := widget.NewButton("Load Weights (JSON)", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			data, err := os.ReadFile(reader.URI().Path())
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			var wf WeightsFile
			if err := json.Unmarshal(data, &wf); err != nil {
				dialog.ShowError(err, w)
				return
			}

			loadedWeights = wf.Weights
			appState.WeightsLoaded = true

			// Find weight range
			wMin, wMax := loadedWeights[0][0], loadedWeights[0][0]
			for _, row := range loadedWeights {
				for _, v := range row {
					if v < wMin {
						wMin = v
					}
					if v > wMax {
						wMax = v
					}
				}
			}

			sourceLabel.SetText(fmt.Sprintf("Loaded: %s\nSize: %dx%d\nWeights: %d values\nRange: [%.4f, %.4f]",
				wf.Name, len(loadedWeights), len(loadedWeights[0]),
				len(loadedWeights)*len(loadedWeights[0]), wMin, wMax))
		}, w)
	})

	sourceSection := container.NewVBox(
		widget.NewLabel("SOURCE"),
		loadButton,
		sourceLabel,
	)

	// Config section
	rowsEntry := widget.NewEntry()
	rowsEntry.SetText("128")
	rowsEntry.SetPlaceHolder("Array Rows")

	colsEntry := widget.NewEntry()
	colsEntry.SetText("128")
	colsEntry.SetPlaceHolder("Array Cols")

	levelsEntry := widget.NewEntry()
	levelsEntry.SetText("30")
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
		widget.NewLabel("TARGET CONFIGURATION"),
		container.NewGridWithColumns(2,
			widget.NewLabel("Array Rows:"), rowsEntry,
			widget.NewLabel("Array Cols:"), colsEntry,
			widget.NewLabel("Levels:"), levelsEntry,
			widget.NewLabel("G_min (μS):"), gMinEntry,
			widget.NewLabel("G_max (μS):"), gMaxEntry,
			widget.NewLabel("V_prog_min (V):"), vMinEntry,
			widget.NewLabel("V_prog_max (V):"), vMaxEntry,
		),
	)

	// Results section
	resultsLabel := widget.NewLabel("Compile to see results...")
	resultsLabel.Wrapping = fyne.TextWrapWord

	compileButton := widget.NewButton("COMPILE", func() {
		if !appState.WeightsLoaded || loadedWeights == nil {
			dialog.ShowError(fmt.Errorf("please load weights first"), w)
			return
		}

		// Parse config
		rows, _ := strconv.Atoi(rowsEntry.Text)
		cols, _ := strconv.Atoi(colsEntry.Text)
		levels, _ := strconv.Atoi(levelsEntry.Text)
		gMin, _ := strconv.ParseFloat(gMinEntry.Text, 64)
		gMax, _ := strconv.ParseFloat(gMaxEntry.Text, 64)
		vMin, _ := strconv.ParseFloat(vMinEntry.Text, 64)
		vMax, _ := strconv.ParseFloat(vMaxEntry.Text, 64)

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

		// Compile
		mapping, err := compiler.Compile(loadedWeights, config)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		appState.CurrentMapping = mapping
		appState.Compiled = true

		// Display results
		resultsLabel.SetText(fmt.Sprintf(
			"COMPILED SUCCESSFULLY\n\n"+
				"Statistics:\n"+
				"  Cells used: %d / %d (%.1f%%)\n"+
				"  Unique levels: %d of %d\n"+
				"  Weight range: [%.4f, %.4f]\n"+
				"  Quant MSE: %.6f\n"+
				"  Quant PSNR: %.2f dB\n",
			mapping.Stats.UsedCells, mapping.Stats.TotalCells, mapping.Stats.Utilization*100,
			mapping.Stats.UniqueLevels, config.Levels,
			mapping.Stats.WeightMin, mapping.Stats.WeightMax,
			mapping.Stats.QuantMSE, mapping.Stats.QuantPSNR))
	})

	resultsSection := container.NewVBox(
		compileButton,
		widget.NewSeparator(),
		widget.NewLabel("COMPILATION RESULTS"),
		resultsLabel,
	)

	// Main layout
	content := container.NewVBox(
		sourceSection,
		widget.NewSeparator(),
		configForm,
		widget.NewSeparator(),
		resultsSection,
	)

	return container.NewScroll(content)
}

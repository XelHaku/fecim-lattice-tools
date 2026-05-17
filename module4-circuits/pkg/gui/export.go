//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// export.go provides data export functionality for circuit simulation results.
package gui

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/shared/export"
)

// CircuitsExportConfig contains the configuration for circuits exports
type CircuitsExportConfig struct {
	Metadata         export.ExportMetadata `json:"metadata"`
	ArrayConfig      ArrayConfig           `json:"array_config"`
	PeripheralConfig PeripheralConfig      `json:"peripheral_config"`
	SimulationState  *SimulationState      `json:"simulation_state,omitempty"`
	SneakMetrics     *SneakPathMetrics     `json:"sneak_metrics,omitempty"`
	TopSneakCells    []SneakCellImpact     `json:"top_sneak_cells,omitempty"`
	ExportedFiles    map[string]string     `json:"exported_files,omitempty"`
}

// ArrayConfig represents the crossbar array configuration
type ArrayConfig struct {
	Rows         int    `json:"rows"`
	Cols         int    `json:"cols"`
	QuantLevels  int    `json:"quant_levels"`
	Architecture string `json:"architecture"`
}

// PeripheralConfig represents the peripheral circuit configuration
type PeripheralConfig struct {
	DACBits     int     `json:"dac_bits"`
	ADCBits     int     `json:"adc_bits"`
	VMin        float64 `json:"v_min"`
	VMax        float64 `json:"v_max"`
	PulseWidth  float64 `json:"pulse_width_ns"`
	ReadVoltage float64 `json:"read_voltage"`
	TIAGain     float64 `json:"tia_gain"`
}

// SimulationState represents the current simulation state
type SimulationState struct {
	SelectedRow  int       `json:"selected_row"`
	SelectedCol  int       `json:"selected_col"`
	TargetLevel  int       `json:"target_level"`
	CurrentLevel int       `json:"current_level"`
	ArrayWeights [][]int   `json:"array_weights"`
	InputVector  []int     `json:"input_vector"`
	OutputVector []float64 `json:"output_vector"`
	Timestamp    time.Time `json:"timestamp"`
}

type peripheralSnapshotRow struct {
	Row            int
	Col            int
	WeightLevel    int
	CellVoltageV   float64
	CellCurrentA   float64
	IsTargetCell   bool
	IsHalfSelected bool
}

// exportSimulationData exports the current simulation data, peripheral snapshot, and diagram.
func (ca *CircuitsApp) exportSimulationData() {
	// Guard against nil state / panics during export
	defer func() {
		if r := recover(); r != nil {
			ca.showExportError(fmt.Sprintf("Export failed unexpectedly: %v", r))
		}
	}()

	timestamp := time.Now().Format("2006-01-02_15-04-05")

	dataDir := filepath.Join("exports", "circuits")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		ca.showExportError(fmt.Sprintf("Cannot create exports folder: %v", err))
		return
	}

	ca.mu.RLock()
	rows := ca.arrayRows
	cols := ca.arrayCols
	weights := make([][]int, rows)
	for i := 0; i < rows; i++ {
		weights[i] = make([]int, cols)
		if i < len(ca.arrayWeights) {
			copy(weights[i], ca.arrayWeights[i])
		}
	}
	inputVector := append([]int(nil), ca.inputVector...)
	outputVector := append([]float64(nil), ca.outputVector...)
	selectedRow := ca.selectedRow
	selectedCol := ca.selectedCol
	targetLevel := ca.targetLevel
	currentLevel := 0
	if selectedRow >= 0 && selectedRow < len(weights) && selectedCol >= 0 && selectedCol < len(weights[selectedRow]) {
		currentLevel = weights[selectedRow][selectedCol]
	}
	quantLevels := ca.quantLevels
	arch := ca.architecture
	ca.mu.RUnlock()

	var coupledVoltages, coupledCurrents [][]float64
	if ca.deviceState != nil {
		coupledVoltages, coupledCurrents = ca.deviceState.GetCoupledCellSnapshot()
	}
	if coupledVoltages == nil {
		coupledVoltages = make([][]float64, rows)
		for r := 0; r < rows; r++ {
			coupledVoltages[r] = make([]float64, cols)
			for c := 0; c < cols; c++ {
				if ca.deviceState != nil {
					coupledVoltages[r][c] = ca.deviceState.GetEffectiveCellVoltage(r, c)
				}
			}
		}
	}
	if coupledCurrents == nil {
		coupledCurrents = make([][]float64, rows)
		for r := 0; r < rows; r++ {
			coupledCurrents[r] = make([]float64, cols)
		}
	}

	sneakMetrics := computeSneakPathMetrics(coupledCurrents, selectedRow, selectedCol)
	peripheralRows := buildPeripheralSnapshotRows(weights, coupledVoltages, coupledCurrents, selectedRow, selectedCol)

	config := CircuitsExportConfig{
		Metadata: *export.NewExportMetadata("module4-circuits"),
		ArrayConfig: ArrayConfig{
			Rows:         rows,
			Cols:         cols,
			QuantLevels:  quantLevels,
			Architecture: string(arch),
		},
		PeripheralConfig: PeripheralConfig{
			DACBits:     ca.dacBits,
			ADCBits:     ca.adcBits,
			VMin:        ca.vMin,
			VMax:        ca.vMax,
			PulseWidth:  ca.pulseWidth,
			ReadVoltage: ca.readVoltage,
			TIAGain:     ca.tiaGain,
		},
		SimulationState: &SimulationState{
			SelectedRow:  selectedRow,
			SelectedCol:  selectedCol,
			TargetLevel:  targetLevel,
			CurrentLevel: currentLevel,
			ArrayWeights: weights,
			InputVector:  inputVector,
			OutputVector: outputVector,
			Timestamp:    time.Now(),
		},
		SneakMetrics:  &sneakMetrics,
		TopSneakCells: sneakMetrics.TopAffectedCells,
		ExportedFiles: map[string]string{},
	}

	jsonExporter := export.NewExporter(dataDir, fmt.Sprintf("circuits-config_%s", timestamp))
	jsonResult := jsonExporter.ExportJSON(config)
	if jsonResult.Error != nil {
		ca.showExportError(fmt.Sprintf("JSON export failed: %v", jsonResult.Error))
		return
	}
	config.ExportedFiles["config_json"] = jsonResult.FilePath

	headers := make([]string, cols+1)
	headers[0] = "Row"
	for j := 0; j < cols; j++ {
		headers[j+1] = fmt.Sprintf("Col%d", j)
	}
	weightsCSV := export.NewCSVData(headers...)
	for i := 0; i < rows; i++ {
		row := make([]string, cols+1)
		row[0] = fmt.Sprintf("%d", i)
		for j := 0; j < cols; j++ {
			row[j+1] = fmt.Sprintf("%d", weights[i][j])
		}
		weightsCSV.AddRow(row...)
	}
	weightsExporter := export.NewExporter(dataDir, fmt.Sprintf("circuits-weights_%s", timestamp))
	weightsResult := weightsExporter.ExportCSV(weightsCSV.Headers, weightsCSV.Rows)
	if weightsResult.Error != nil {
		ca.showExportError(fmt.Sprintf("Weights CSV export failed: %v", weightsResult.Error))
		return
	}
	config.ExportedFiles["weights_csv"] = weightsResult.FilePath

	peripheralCSV := peripheralSnapshotCSV(peripheralRows)
	periphExporter := export.NewExporter(dataDir, fmt.Sprintf("circuits-peripheral_%s", timestamp))
	periphResult := periphExporter.ExportCSV(peripheralCSV.Headers, peripheralCSV.Rows)
	if periphResult.Error != nil {
		ca.showExportError(fmt.Sprintf("Peripheral CSV export failed: %v", periphResult.Error))
		return
	}
	config.ExportedFiles["peripheral_csv"] = periphResult.FilePath

	if ca.deviceState != nil {
		spicePath := filepath.Join(dataDir, fmt.Sprintf("circuits-crossbar_%s.sp", timestamp))
		if spiceNetlist, err := ca.buildSpiceNetlist(weights); err == nil {
			if writeErr := os.WriteFile(spicePath, []byte(spiceNetlist), 0644); writeErr == nil {
				config.ExportedFiles["crossbar_spice"] = spicePath
			}
		}
	}

	if ca.window != nil && ca.window.Canvas() != nil {
		vizExporter := export.NewExporter(dataDir, fmt.Sprintf("circuits-viz_%s", timestamp))
		vizResult := vizExporter.ExportPNG(ca.window.Canvas().Capture())
		if vizResult.Error == nil {
			config.ExportedFiles["diagram_png"] = vizResult.FilePath
		}
	}

	ca.showExportSuccess(fmt.Sprintf("Exported:\n• %s\n• %s\n• %s", jsonResult.FilePath, weightsResult.FilePath, periphResult.FilePath))
}

// exportVisualization exports the current visualization as a PNG
func (ca *CircuitsApp) exportVisualization() {
	export.ExportVisualization(ca.window, "circuits", nil)
}

// createExportButtons creates the export button panel for circuits
func (ca *CircuitsApp) createExportButtons() fyne.CanvasObject {
	config := export.DefaultExportConfig("circuits")
	config.OutputDir = filepath.Join("exports", "circuits")

	return export.NewExportButtons(config, &circuitsExportProvider{app: ca}, ca.window)
}

// circuitsExportProvider implements export.ExportDataProvider for circuits
type circuitsExportProvider struct {
	app *CircuitsApp
}

// GetCSVData returns circuits array data as CSV
func (p *circuitsExportProvider) GetCSVData() (*export.CSVData, error) {
	if p == nil || p.app == nil {
		return nil, fmt.Errorf("app state not initialized")
	}

	p.app.mu.RLock()
	defer p.app.mu.RUnlock()

	// Build headers (columns)
	headers := make([]string, p.app.arrayCols+1)
	headers[0] = "Row"
	for j := 0; j < p.app.arrayCols; j++ {
		headers[j+1] = fmt.Sprintf("Col%d", j)
	}

	// Build rows
	data := export.NewCSVData(headers...)
	for i := 0; i < p.app.arrayRows; i++ {
		row := make([]string, p.app.arrayCols+1)
		row[0] = fmt.Sprintf("%d", i)
		for j := 0; j < p.app.arrayCols; j++ {
			w := 0
			if i < len(p.app.arrayWeights) && j < len(p.app.arrayWeights[i]) {
				w = p.app.arrayWeights[i][j]
			}
			row[j+1] = fmt.Sprintf("%d", w)
		}
		data.AddRow(row...)
	}

	return data, nil
}

// GetJSONConfig returns circuits configuration as JSON-serializable data
func (p *circuitsExportProvider) GetJSONConfig() (interface{}, error) {
	if p == nil || p.app == nil {
		return nil, fmt.Errorf("app state not initialized")
	}

	config := CircuitsExportConfig{
		Metadata: *export.NewExportMetadata("module4-circuits"),
		ArrayConfig: ArrayConfig{
			Rows:         p.app.arrayRows,
			Cols:         p.app.arrayCols,
			QuantLevels:  p.app.quantLevels,
			Architecture: string(p.app.architecture),
		},
		PeripheralConfig: PeripheralConfig{
			DACBits:     p.app.dacBits,
			ADCBits:     p.app.adcBits,
			VMin:        p.app.vMin,
			VMax:        p.app.vMax,
			PulseWidth:  p.app.pulseWidth,
			ReadVoltage: p.app.readVoltage,
			TIAGain:     p.app.tiaGain,
		},
	}
	if p.app.deviceState != nil {
		_, currents := p.app.deviceState.GetCoupledCellSnapshot()
		metrics := computeSneakPathMetrics(currents, p.app.deviceState.GetSelectedRow(), p.app.deviceState.GetSelectedCol())
		config.SneakMetrics = &metrics
		config.TopSneakCells = metrics.TopAffectedCells
	}
	return config, nil
}

// GetVisualization returns the current visualization as an image
func (p *circuitsExportProvider) GetVisualization() (image.Image, error) {
	if p == nil || p.app == nil || p.app.window == nil || p.app.window.Canvas() == nil {
		return nil, fmt.Errorf("window not available")
	}
	return p.app.window.Canvas().Capture(), nil
}

func buildPeripheralSnapshotRows(weights [][]int, voltages [][]float64, currents [][]float64, selectedRow, selectedCol int) []peripheralSnapshotRow {
	rows := make([]peripheralSnapshotRow, 0)
	for r := range weights {
		for c := range weights[r] {
			cell := peripheralSnapshotRow{
				Row:            r,
				Col:            c,
				WeightLevel:    weights[r][c],
				IsTargetCell:   r == selectedRow && c == selectedCol,
				IsHalfSelected: (r == selectedRow || c == selectedCol) && !(r == selectedRow && c == selectedCol),
			}
			if r < len(voltages) && c < len(voltages[r]) {
				cell.CellVoltageV = voltages[r][c]
			}
			if r < len(currents) && c < len(currents[r]) {
				cell.CellCurrentA = currents[r][c]
			}
			rows = append(rows, cell)
		}
	}
	return rows
}

func (ca *CircuitsApp) buildSpiceNetlist(weights [][]int) (string, error) {
	if ca == nil || ca.deviceState == nil {
		return "", fmt.Errorf("device state unavailable")
	}
	ca.deviceState.mu.RLock()
	defer ca.deviceState.mu.RUnlock()

	rows := ca.deviceState.rows
	cols := ca.deviceState.cols
	conductance := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		conductance[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}
			conductance[r][c] = ca.deviceState.levelToConductance(level, ca.quantLevels)
		}
	}

	params := arraysim.SolveParams{
		WLVoltages:  append([]float64(nil), ca.deviceState.wlVoltages...),
		BLVoltages:  append([]float64(nil), ca.deviceState.dacVoltages...),
		Conductance: conductance,
		ActiveRows:  append([]bool(nil), ca.deviceState.activeRows...),
		Geometry:    ca.deviceState.cellGeometry,
		Wire:        ca.deviceState.wireParams,
	}
	return arraysim.ExportCrossbarSPICE(params, arraysim.SpiceExportConfig{Title: "Module4 Circuits crossbar behavioral deck"})
}

func peripheralSnapshotCSV(rows []peripheralSnapshotRow) *export.CSVData {
	data := export.NewCSVData("Row", "Col", "WeightLevel", "CellVoltage_V", "CellCurrent_A", "IsTargetCell", "IsHalfSelected")
	for _, row := range rows {
		data.AddRow(
			fmt.Sprintf("%d", row.Row),
			fmt.Sprintf("%d", row.Col),
			fmt.Sprintf("%d", row.WeightLevel),
			fmt.Sprintf("%.9g", row.CellVoltageV),
			fmt.Sprintf("%.9g", row.CellCurrentA),
			fmt.Sprintf("%t", row.IsTargetCell),
			fmt.Sprintf("%t", row.IsHalfSelected),
		)
	}
	return data
}

// showExportError displays an export error dialog
func (ca *CircuitsApp) showExportError(msg string) {
	export.ShowExportError(ca.window, nil, msg)
}

// showExportSuccess displays an export success dialog
func (ca *CircuitsApp) showExportSuccess(msg string) {
	export.ShowExportSuccess(ca.window, nil, msg)
}

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
	"fyne.io/fyne/v2/dialog"

	"fecim-lattice-tools/shared/export"
)

// CircuitsExportConfig contains the configuration for circuits exports
type CircuitsExportConfig struct {
	Metadata      export.ExportMetadata `json:"metadata"`
	ArrayConfig   ArrayConfig           `json:"array_config"`
	PeripheralConfig PeripheralConfig   `json:"peripheral_config"`
	SimulationState  *SimulationState   `json:"simulation_state,omitempty"`
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
	SelectedRow   int         `json:"selected_row"`
	SelectedCol   int         `json:"selected_col"`
	TargetLevel   int         `json:"target_level"`
	CurrentLevel  int         `json:"current_level"`
	ArrayWeights  [][]int     `json:"array_weights"`
	InputVector   []int       `json:"input_vector"`
	OutputVector  []float64   `json:"output_vector"`
	Timestamp     time.Time   `json:"timestamp"`
}

// exportSimulationData exports the current simulation data to CSV and JSON files
func (ca *CircuitsApp) exportSimulationData() {
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Ensure data/circuits folder exists
	dataDir := filepath.Join("exports", "circuits")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		ca.showExportError(fmt.Sprintf("Cannot create exports folder: %v", err))
		return
	}

	// Build export configuration
	config := CircuitsExportConfig{
		Metadata: *export.NewExportMetadata("module4-circuits"),
		ArrayConfig: ArrayConfig{
			Rows:         ca.arrayRows,
			Cols:         ca.arrayCols,
			QuantLevels:  ca.quantLevels,
			Architecture: string(ca.architecture),
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
	}

	// Add simulation state
	ca.mu.RLock()
	config.SimulationState = &SimulationState{
		SelectedRow:  ca.selectedRow,
		SelectedCol:  ca.selectedCol,
		TargetLevel:  ca.targetLevel,
		CurrentLevel: ca.arrayWeights[ca.selectedRow][ca.selectedCol],
		ArrayWeights: ca.arrayWeights,
		InputVector:  ca.inputVector,
		OutputVector: ca.outputVector,
		Timestamp:    time.Now(),
	}
	ca.mu.RUnlock()

	// Export JSON configuration
	exporter := export.NewExporter(dataDir, fmt.Sprintf("circuits-config_%s", timestamp))
	jsonResult := exporter.ExportJSON(config)
	if jsonResult.Error != nil {
		ca.showExportError(fmt.Sprintf("JSON export failed: %v", jsonResult.Error))
		return
	}

	// Export array weights as CSV
	csvExporter := export.NewExporter(dataDir, fmt.Sprintf("circuits-weights_%s", timestamp))
	
	// Build headers (columns)
	headers := make([]string, ca.arrayCols+1)
	headers[0] = "Row"
	for j := 0; j < ca.arrayCols; j++ {
		headers[j+1] = fmt.Sprintf("Col%d", j)
	}
	
	// Build rows
	csvData := export.NewCSVData(headers...)
	ca.mu.RLock()
	for i := 0; i < ca.arrayRows; i++ {
		row := make([]string, ca.arrayCols+1)
		row[0] = fmt.Sprintf("%d", i)
		for j := 0; j < ca.arrayCols; j++ {
			row[j+1] = fmt.Sprintf("%d", ca.arrayWeights[i][j])
		}
		csvData.AddRow(row...)
	}
	ca.mu.RUnlock()
	
	csvResult := csvExporter.ExportCSV(csvData.Headers, csvData.Rows)
	if csvResult.Error != nil {
		ca.showExportError(fmt.Sprintf("CSV export failed: %v", csvResult.Error))
		return
	}

	ca.showExportSuccess(fmt.Sprintf("Exported:\n• %s\n• %s", jsonResult.FilePath, csvResult.FilePath))
}

// exportVisualization exports the current visualization as a PNG
func (ca *CircuitsApp) exportVisualization() {
	if ca.window == nil {
		return
	}

	dataDir := filepath.Join("exports", "circuits")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		ca.showExportError(fmt.Sprintf("Cannot create exports folder: %v", err))
		return
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	exporter := export.NewExporter(dataDir, fmt.Sprintf("circuits-viz_%s", timestamp))
	img := ca.window.Canvas().Capture()
	result := exporter.ExportPNG(img)
	
	if result.Error != nil {
		ca.showExportError(fmt.Sprintf("Image export failed: %v", result.Error))
		return
	}
	
	ca.showExportSuccess(fmt.Sprintf("Image saved:\n• %s", result.FilePath))
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
			row[j+1] = fmt.Sprintf("%d", p.app.arrayWeights[i][j])
		}
		data.AddRow(row...)
	}

	return data, nil
}

// GetJSONConfig returns circuits configuration as JSON-serializable data
func (p *circuitsExportProvider) GetJSONConfig() (interface{}, error) {
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
	return config, nil
}

// GetVisualization returns the current visualization as an image
func (p *circuitsExportProvider) GetVisualization() (image.Image, error) {
	if p.app.window == nil {
		return nil, fmt.Errorf("window not available")
	}
	return p.app.window.Canvas().Capture(), nil
}

// showExportError displays an export error dialog
func (ca *CircuitsApp) showExportError(msg string) {
	if ca.window != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("%s", msg), ca.window)
		})
	}
}

// showExportSuccess displays an export success dialog
func (ca *CircuitsApp) showExportSuccess(msg string) {
	if ca.window != nil {
		fyne.Do(func() {
			dialog.ShowInformation("Export Complete", msg, ca.window)
		})
	}
}

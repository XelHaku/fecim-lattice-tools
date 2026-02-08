// Package gui provides Fyne-based GUI components for MNIST visualization.
// export.go provides data export functionality for MNIST inference results.
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

// MNISTExportConfig contains the configuration for MNIST exports
type MNISTExportConfig struct {
	Metadata         export.ExportMetadata `json:"metadata"`
	NetworkConfig    NetworkConfig         `json:"network_config"`
	InferenceResults *InferenceResults     `json:"inference_results,omitempty"`
}

// NetworkConfig represents the MNIST network configuration
type NetworkConfig struct {
	InputSize  int     `json:"input_size"`
	HiddenSize int     `json:"hidden_size"`
	OutputSize int     `json:"output_size"`
	Levels     int     `json:"quantization_levels"`
	NoiseLevel float64 `json:"noise_level"`
	ADCBits    int     `json:"adc_bits"`
	DACBits    int     `json:"dac_bits"`
}

// InferenceResults contains the results of MNIST inference
type InferenceResults struct {
	DrawnDigit       []float64   `json:"drawn_digit,omitempty"`
	FPPrediction     int         `json:"fp_prediction"`
	FPConfidence     float64     `json:"fp_confidence"`
	FPProbabilities  []float64   `json:"fp_probabilities"`
	CIMPrediction    int         `json:"cim_prediction"`
	CIMConfidence    float64     `json:"cim_confidence"`
	CIMProbabilities []float64   `json:"cim_probabilities"`
	Agreement        bool        `json:"agreement"`
	EnergyRatio      float64     `json:"energy_ratio"`
	Timestamp        time.Time   `json:"timestamp"`
}

// exportInferenceData exports the current inference data to CSV and JSON files
func (app *DualModeApp) exportInferenceData() {
	// Ensure data/mnist folder exists
	dataDir := filepath.Join("exports", "mnist")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		app.showExportError(fmt.Sprintf("Cannot create exports folder: %v", err))
		return
	}

	// Build export configuration
	config := MNISTExportConfig{
		Metadata: *export.NewExportMetadata("module3-mnist"),
		NetworkConfig: NetworkConfig{
			InputSize:  MNISTInputSize,
			HiddenSize: MNISTHiddenSize,
			OutputSize: MNISTOutputSize,
			Levels:     app.networkCtrl.GetNumLevels(),
			NoiseLevel: app.networkCtrl.Network().Config.NoiseLevel,
			ADCBits:    FeCIMDefaultADC,
			DACBits:    FeCIMDefaultDAC,
		},
	}

	// Add inference results if available
	if app.lastPixels != nil && len(app.lastPixels) > 0 {
		// Run inference to get probabilities
		result := app.networkCtrl.Infer(app.lastPixels)
		
		config.InferenceResults = &InferenceResults{
			DrawnDigit:       app.lastPixels,
			FPPrediction:     result.FPPrediction,
			FPConfidence:     result.FPConfidence,
			FPProbabilities:  result.FPProbabilities,
			CIMPrediction:    result.CIMPrediction,
			CIMConfidence:    result.CIMConfidence,
			CIMProbabilities: result.CIMProbabilities,
			Agreement:        result.Agree,
			EnergyRatio:      EnergyRatioGPU,
			Timestamp:        time.Now(),
		}
	}

	// Export JSON configuration
	exporter := export.NewExporter(dataDir, "mnist-config")
	jsonResult := exporter.ExportJSON(config)
	if jsonResult.Error != nil {
		app.showExportError(fmt.Sprintf("JSON export failed: %v", jsonResult.Error))
		return
	}

	// Export probabilities as CSV
	if config.InferenceResults != nil {
		csvExporter := export.NewExporter(dataDir, "mnist-probs")
		headers := []string{"Digit", "FP_Probability", "CIM_Probability", "Difference"}
		
		csvData := export.NewCSVData(headers...)
		for i := 0; i < 10; i++ {
			fpProb := 0.0
			cimProb := 0.0
			if i < len(config.InferenceResults.FPProbabilities) {
				fpProb = config.InferenceResults.FPProbabilities[i]
			}
			if i < len(config.InferenceResults.CIMProbabilities) {
				cimProb = config.InferenceResults.CIMProbabilities[i]
			}
			csvData.AddRow(
				fmt.Sprintf("%d", i),
				fmt.Sprintf("%.6f", fpProb),
				fmt.Sprintf("%.6f", cimProb),
				fmt.Sprintf("%.6f", fpProb-cimProb),
			)
		}
		
		csvResult := csvExporter.ExportCSV(csvData.Headers, csvData.Rows)
		if csvResult.Error != nil {
			app.showExportError(fmt.Sprintf("CSV export failed: %v", csvResult.Error))
			return
		}
		
		app.showExportSuccess(fmt.Sprintf("Exported:\n• %s\n• %s", jsonResult.FilePath, csvResult.FilePath))
	} else {
		app.showExportSuccess(fmt.Sprintf("Exported configuration:\n• %s", jsonResult.FilePath))
	}
}

// exportVisualization exports the current visualization as a PNG
func (app *DualModeApp) exportVisualization() {
	if app.window == nil {
		return
	}

	dataDir := filepath.Join("exports", "mnist")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		app.showExportError(fmt.Sprintf("Cannot create exports folder: %v", err))
		return
	}

	exporter := export.NewExporter(dataDir, "mnist-viz")
	img := app.window.Canvas().Capture()
	result := exporter.ExportPNG(img)
	
	if result.Error != nil {
		app.showExportError(fmt.Sprintf("Image export failed: %v", result.Error))
		return
	}
	
	app.showExportSuccess(fmt.Sprintf("Image saved:\n• %s", result.FilePath))
}

// createExportButtons creates the export button panel for MNIST
func (app *DualModeApp) createExportButtons() fyne.CanvasObject {
	config := export.DefaultExportConfig("mnist")
	config.OutputDir = filepath.Join("exports", "mnist")
	
	return export.NewExportButtons(config, &mnistExportProvider{app: app}, app.window)
}

// mnistExportProvider implements export.ExportDataProvider for MNIST
type mnistExportProvider struct {
	app *DualModeApp
}

// GetCSVData returns MNIST inference data as CSV
func (p *mnistExportProvider) GetCSVData() (*export.CSVData, error) {
	if p.app.lastPixels == nil || len(p.app.lastPixels) == 0 {
		return nil, fmt.Errorf("no inference data available - draw a digit first")
	}

	// Run inference to get probabilities
	result := p.app.networkCtrl.Infer(p.app.lastPixels)
	fpProbs := result.FPProbabilities
	cimProbs := result.CIMProbabilities

	data := export.NewCSVData("Digit", "FP_Probability", "CIM_Probability", "Difference")
	for i := 0; i < 10; i++ {
		fpProb := 0.0
		cimProb := 0.0
		if i < len(fpProbs) {
			fpProb = fpProbs[i]
		}
		if i < len(cimProbs) {
			cimProb = cimProbs[i]
		}
		data.AddRowFromFloats(float64(i), fpProb, cimProb, fpProb-cimProb)
	}

	return data, nil
}

// GetJSONConfig returns MNIST configuration as JSON-serializable data
func (p *mnistExportProvider) GetJSONConfig() (interface{}, error) {
	config := MNISTExportConfig{
		Metadata: *export.NewExportMetadata("module3-mnist"),
		NetworkConfig: NetworkConfig{
			InputSize:  MNISTInputSize,
			HiddenSize: MNISTHiddenSize,
			OutputSize: MNISTOutputSize,
			Levels:     p.app.networkCtrl.GetNumLevels(),
			NoiseLevel: p.app.networkCtrl.Network().Config.NoiseLevel,
			ADCBits:    FeCIMDefaultADC,
			DACBits:    FeCIMDefaultDAC,
		},
	}
	return config, nil
}

// GetVisualization returns the current visualization as an image
func (p *mnistExportProvider) GetVisualization() (image.Image, error) {
	if p.app.window == nil {
		return nil, fmt.Errorf("window not available")
	}
	return p.app.window.Canvas().Capture(), nil
}

// showExportError displays an export error dialog
func (app *DualModeApp) showExportError(msg string) {
	if app.window != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("%s", msg), app.window)
		})
	}
	mnistLog.Printf("Export error: %s", msg)
}

// showExportSuccess displays an export success dialog
func (app *DualModeApp) showExportSuccess(msg string) {
	if app.window != nil {
		fyne.Do(func() {
			dialog.ShowInformation("Export Complete", msg, app.window)
		})
	}
	mnistLog.Info("Export complete: %s", msg)
}

// maxWithConfidence returns the index and value of the maximum element
func maxWithConfidence(probs []float64) (int, float64) {
	maxIdx := 0
	maxVal := 0.0
	for i, p := range probs {
		if p > maxVal {
			maxVal = p
			maxIdx = i
		}
	}
	return maxIdx, maxVal
}

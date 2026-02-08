// Package gui provides Fyne-based GUI components for architecture comparison.
// export.go provides data export functionality for comparison results.
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

// ComparisonExportConfig contains the configuration for comparison exports
type ComparisonExportConfig struct {
	Metadata          export.ExportMetadata  `json:"metadata"`
	WorkloadConfig    WorkloadConfig         `json:"workload_config"`
	EnergySpecs       EnergySpecsExport      `json:"energy_specs"`
	CalculatedResults *CalculatedResults     `json:"calculated_results,omitempty"`
}

// WorkloadConfig represents the workload configuration
type WorkloadConfig struct {
	WorkloadName     string  `json:"workload_name"`
	InferencesPerDay int     `json:"inferences_per_day"`
	MACsPerInference int64   `json:"macs_per_inference"`
	Description      string  `json:"description,omitempty"`
}

// EnergySpecsExport represents energy specifications for export
type EnergySpecsExport struct {
	CPUSpec   EnergySpecExport `json:"cpu_spec"`
	GPUSpec   EnergySpecExport `json:"gpu_spec"`
	FeCIMSpec EnergySpecExport `json:"fecim_spec"`
}

// EnergySpecExport represents a single energy spec for export
type EnergySpecExport struct {
	Name          string  `json:"name"`
	EnergyFJ      float64 `json:"energy_fj_per_mac"`
	Source        string  `json:"source"`
	Verified      bool    `json:"verified"`
	SourceDetails string  `json:"source_details,omitempty"`
}

// CalculatedResults contains the calculated comparison results
type CalculatedResults struct {
	TotalMACsPerDay     float64 `json:"total_macs_per_day"`
	CPUEnergyKWh        float64 `json:"cpu_energy_kwh_per_day"`
	GPUEnergyKWh        float64 `json:"gpu_energy_kwh_per_day"`
	FeCIMEnergyKWh      float64 `json:"fecim_energy_kwh_per_day"`
	CPUSavingsKWh       float64 `json:"cpu_savings_kwh_per_day"`
	GPUSavingsKWh       float64 `json:"gpu_savings_kwh_per_day"`
	CPUSavingsPercent   float64 `json:"cpu_savings_percent"`
	GPUSavingsPercent   float64 `json:"gpu_savings_percent"`
	AnnualCostSavingsM  float64 `json:"annual_cost_savings_millions"`
	Timestamp           time.Time `json:"timestamp"`
}

// exportComparisonData exports the current comparison data to CSV and JSON files
func (ca *ComparisonApp) exportComparisonData() {
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Ensure data/comparison folder exists
	dataDir := filepath.Join("exports", "comparison")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		ca.showExportError(fmt.Sprintf("Cannot create exports folder: %v", err))
		return
	}

	// Build export configuration
	config := ComparisonExportConfig{
		Metadata: *export.NewExportMetadata("module5-comparison"),
		WorkloadConfig: WorkloadConfig{
			WorkloadName:     ca.currentWorkload,
			InferencesPerDay: int(ca.currentInferences),
			MACsPerInference: getWorkloadMACs(ca.currentWorkload),
		},
		EnergySpecs: EnergySpecsExport{
			CPUSpec: EnergySpecExport{
				Name:          ca.cpuSpec.Name,
				EnergyFJ:      ca.cpuSpec.EnergyFJ,
				Source:        ca.cpuSpec.Source,
				Verified:      ca.cpuSpec.Verified,
				SourceDetails: ca.cpuSpec.SourceDetails,
			},
			GPUSpec: EnergySpecExport{
				Name:          ca.gpuSpec.Name,
				EnergyFJ:      ca.gpuSpec.EnergyFJ,
				Source:        ca.gpuSpec.Source,
				Verified:      ca.gpuSpec.Verified,
				SourceDetails: ca.gpuSpec.SourceDetails,
			},
			FeCIMSpec: EnergySpecExport{
				Name:          ca.fecimSpec.Name,
				EnergyFJ:      ca.fecimSpec.EnergyFJ,
				Source:        ca.fecimSpec.Source,
				Verified:      ca.fecimSpec.Verified,
				SourceDetails: ca.fecimSpec.SourceDetails,
			},
		},
	}

	// Calculate and add results
	totalMACs := float64(ca.currentInferences) * float64(getWorkloadMACs(ca.currentWorkload))
	cpuEnergy := totalMACs * ca.cpuSpec.EnergyFJ / 1e15 / 3600 // fJ to kWh
	gpuEnergy := totalMACs * ca.gpuSpec.EnergyFJ / 1e15 / 3600
	fecimEnergy := totalMACs * ca.fecimSpec.EnergyFJ / 1e15 / 3600
	
	config.CalculatedResults = &CalculatedResults{
		TotalMACsPerDay:    totalMACs,
		CPUEnergyKWh:       cpuEnergy,
		GPUEnergyKWh:       gpuEnergy,
		FeCIMEnergyKWh:     fecimEnergy,
		CPUSavingsKWh:      cpuEnergy - fecimEnergy,
		GPUSavingsKWh:      gpuEnergy - fecimEnergy,
		CPUSavingsPercent:  (1 - fecimEnergy/cpuEnergy) * 100,
		GPUSavingsPercent:  (1 - fecimEnergy/gpuEnergy) * 100,
		AnnualCostSavingsM: (gpuEnergy - fecimEnergy) * 365 * 0.10 / 1000, // $0.10/kWh
		Timestamp:          time.Now(),
	}

	// Export JSON configuration
	exporter := export.NewExporter(dataDir, fmt.Sprintf("comparison-config_%s", timestamp))
	jsonResult := exporter.ExportJSON(config)
	if jsonResult.Error != nil {
		ca.showExportError(fmt.Sprintf("JSON export failed: %v", jsonResult.Error))
		return
	}

	// Export comparison as CSV
	csvExporter := export.NewExporter(dataDir, fmt.Sprintf("comparison-data_%s", timestamp))
	
	csvData := export.NewCSVData("Architecture", "Energy_FJ_per_MAC", "Energy_kWh_per_day", "Savings_vs_FeCIM_percent")
	csvData.AddRow("CPU + DRAM", 
		fmt.Sprintf("%.0f", ca.cpuSpec.EnergyFJ),
		fmt.Sprintf("%.6f", cpuEnergy),
		fmt.Sprintf("%.2f", config.CalculatedResults.CPUSavingsPercent))
	csvData.AddRow("GPU + HBM",
		fmt.Sprintf("%.0f", ca.gpuSpec.EnergyFJ),
		fmt.Sprintf("%.6f", gpuEnergy),
		fmt.Sprintf("%.2f", config.CalculatedResults.GPUSavingsPercent))
	csvData.AddRow("FeCIM",
		fmt.Sprintf("%.0f", ca.fecimSpec.EnergyFJ),
		fmt.Sprintf("%.6f", fecimEnergy),
		"0.00")
	
	csvResult := csvExporter.ExportCSV(csvData.Headers, csvData.Rows)
	if csvResult.Error != nil {
		ca.showExportError(fmt.Sprintf("CSV export failed: %v", csvResult.Error))
		return
	}

	ca.showExportSuccess(fmt.Sprintf("Exported:\n• %s\n• %s", jsonResult.FilePath, csvResult.FilePath))
}

// exportVisualization exports the current visualization as a PNG
func (ca *ComparisonApp) exportVisualization() {
	if ca.window == nil {
		return
	}

	dataDir := filepath.Join("exports", "comparison")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		ca.showExportError(fmt.Sprintf("Cannot create exports folder: %v", err))
		return
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	exporter := export.NewExporter(dataDir, fmt.Sprintf("comparison-viz_%s", timestamp))
	img := ca.window.Canvas().Capture()
	result := exporter.ExportPNG(img)
	
	if result.Error != nil {
		ca.showExportError(fmt.Sprintf("Image export failed: %v", result.Error))
		return
	}
	
	ca.showExportSuccess(fmt.Sprintf("Image saved:\n• %s", result.FilePath))
}

// createExportButtons creates the export button panel for comparison
func (ca *ComparisonApp) createExportButtons() fyne.CanvasObject {
	config := export.DefaultExportConfig("comparison")
	config.OutputDir = filepath.Join("exports", "comparison")
	
	return export.NewExportButtons(config, &comparisonExportProvider{app: ca}, ca.window)
}

// comparisonExportProvider implements export.ExportDataProvider for comparison
type comparisonExportProvider struct {
	app *ComparisonApp
}

// GetCSVData returns comparison data as CSV
func (p *comparisonExportProvider) GetCSVData() (*export.CSVData, error) {
	totalMACs := float64(p.app.currentInferences) * float64(getWorkloadMACs(p.app.currentWorkload))
	cpuEnergy := totalMACs * p.app.cpuSpec.EnergyFJ / 1e15 / 3600
	gpuEnergy := totalMACs * p.app.gpuSpec.EnergyFJ / 1e15 / 3600
	fecimEnergy := totalMACs * p.app.fecimSpec.EnergyFJ / 1e15 / 3600

	data := export.NewCSVData("Architecture", "Energy_FJ_per_MAC", "Energy_kWh_per_day", "Savings_percent")
	data.AddRow("CPU + DRAM",
		fmt.Sprintf("%.0f", p.app.cpuSpec.EnergyFJ),
		fmt.Sprintf("%.6f", cpuEnergy),
		fmt.Sprintf("%.2f", (1-fecimEnergy/cpuEnergy)*100))
	data.AddRow("GPU + HBM",
		fmt.Sprintf("%.0f", p.app.gpuSpec.EnergyFJ),
		fmt.Sprintf("%.6f", gpuEnergy),
		fmt.Sprintf("%.2f", (1-fecimEnergy/gpuEnergy)*100))
	data.AddRow("FeCIM",
		fmt.Sprintf("%.0f", p.app.fecimSpec.EnergyFJ),
		fmt.Sprintf("%.6f", fecimEnergy),
		"0.00")

	return data, nil
}

// GetJSONConfig returns comparison configuration as JSON-serializable data
func (p *comparisonExportProvider) GetJSONConfig() (interface{}, error) {
	config := ComparisonExportConfig{
		Metadata: *export.NewExportMetadata("module5-comparison"),
		WorkloadConfig: WorkloadConfig{
			WorkloadName:     p.app.currentWorkload,
			InferencesPerDay: int(p.app.currentInferences),
			MACsPerInference: getWorkloadMACs(p.app.currentWorkload),
		},
		EnergySpecs: EnergySpecsExport{
			CPUSpec: EnergySpecExport{
				Name:     p.app.cpuSpec.Name,
				EnergyFJ: p.app.cpuSpec.EnergyFJ,
				Source:   p.app.cpuSpec.Source,
				Verified: p.app.cpuSpec.Verified,
			},
			GPUSpec: EnergySpecExport{
				Name:     p.app.gpuSpec.Name,
				EnergyFJ: p.app.gpuSpec.EnergyFJ,
				Source:   p.app.gpuSpec.Source,
				Verified: p.app.gpuSpec.Verified,
			},
			FeCIMSpec: EnergySpecExport{
				Name:     p.app.fecimSpec.Name,
				EnergyFJ: p.app.fecimSpec.EnergyFJ,
				Source:   p.app.fecimSpec.Source,
				Verified: p.app.fecimSpec.Verified,
			},
		},
	}
	return config, nil
}

// GetVisualization returns the current visualization as an image
func (p *comparisonExportProvider) GetVisualization() (image.Image, error) {
	if p.app.window == nil {
		return nil, fmt.Errorf("window not available")
	}
	return p.app.window.Canvas().Capture(), nil
}

// showExportError displays an export error dialog
func (ca *ComparisonApp) showExportError(msg string) {
	if ca.window != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("%s", msg), ca.window)
		})
	}
	debug.Printf("Export error: %s", msg)
}

// showExportSuccess displays an export success dialog
func (ca *ComparisonApp) showExportSuccess(msg string) {
	if ca.window != nil {
		fyne.Do(func() {
			dialog.ShowInformation("Export Complete", msg, ca.window)
		})
	}
	debug.Printf("Export complete: %s", msg)
}

// getWorkloadMACs returns the MACs per inference for a given workload
func getWorkloadMACs(workload string) int64 {
	// These are model-based estimates, not verified values
	switch workload {
	case "GPT-2":
		return 774_000_000 // 774M parameters
	case "GPT-3":
		return 175_000_000_000 // 175B parameters
	case "LLaMA-7B":
		return 7_000_000_000 // 7B parameters
	case "BERT":
		return 340_000_000 // 340M parameters
	case "ResNet-50":
		return 4_100_000_000 // 4.1G FLOPs
	default:
		return 1_000_000_000 // Default 1B MACs
	}
}

package gui

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"

	sharedio "fecim-lattice-tools/shared/io"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// PEDataExport represents the complete P-E hysteresis data export structure
type PEDataExport struct {
	Metadata ExportMetadata `json:"metadata"`
	Data     ExportData     `json:"data"`
}

// ExportMetadata contains material and simulation parameters
type ExportMetadata struct {
	Material     string  `json:"material"`
	EcMVcm       float64 `json:"ec_mv_cm"`
	PsUcCm2      float64 `json:"ps_uc_cm2"`
	PrUcCm2      float64 `json:"pr_uc_cm2"`
	TemperatureK float64 `json:"temperature_k"`
	NumLevels    int     `json:"num_levels"`
	ExportedAt   string  `json:"exported_at"`
	DataPoints   int     `json:"data_points"`
	Waveform     string  `json:"waveform"`
	FrequencyHz  float64 `json:"frequency_hz,omitempty"`
}

// ExportData contains the actual measurement data
type ExportData struct {
	EFieldMVcm        []float64 `json:"e_field_mv_cm"`
	PolarizationUcCm2 []float64 `json:"polarization_uc_cm2"`
}

// exportPEDataToJSON exports P-E hysteresis data to a JSON file
// THREAD-SAFE: Copies data under lock, then writes without lock.
// Safe to call from goroutines - no UI updates needed.
func (a *App) exportPEDataToJSON(filename string) error {
	// Copy data under lock
	a.mu.RLock()
	eData, pData := a.historySnapshotLocked()

	// Copy metadata
	materialName := "Unknown"
	ec := 0.0
	ps := 0.0
	pr := 0.0
	temp := 300.0
	numLevels := a.numLevels
	waveform := a.waveform.String()
	frequency := a.frequency

	if a.material != nil {
		materialName = a.material.Name
		ec = a.material.Ec
		ps = a.material.Ps
		pr = a.material.Pr
	}
	temp = a.currentTemperature()
	a.mu.RUnlock()

	// Convert to convenient units (MV/cm and μC/cm²)
	eFieldMVcm := make([]float64, len(eData))
	polarizationUcCm2 := make([]float64, len(pData))
	for i := range eData {
		eFieldMVcm[i] = sharedphysics.VPerMToMVPerCm(eData[i]) // V/m → MV/cm
		polarizationUcCm2[i] = pData[i] * 1e2 // C/m² to μC/cm²
	}

	// Build export structure
	export := PEDataExport{
		Metadata: ExportMetadata{
			Material:     materialName,
			EcMVcm:       sharedphysics.VPerMToMVPerCm(ec),
			PsUcCm2:      ps * 1e2,
			PrUcCm2:      pr * 1e2,
			TemperatureK: temp,
			NumLevels:    numLevels,
			ExportedAt:   time.Now().Format(time.RFC3339),
			DataPoints:   len(eData),
			Waveform:     waveform,
			FrequencyHz:  frequency,
		},
		Data: ExportData{
			EFieldMVcm:        eFieldMVcm,
			PolarizationUcCm2: polarizationUcCm2,
		},
	}

	// Write to file
	if err := sharedio.SaveJSON(filename, export); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	log.Info("Exported P-E data to JSON: %s (%d points)", filename, len(eData))
	return nil
}

// exportPEDataToCSV exports P-E hysteresis data to a CSV file
// THREAD-SAFE: Copies data under lock, then writes without lock.
// Safe to call from goroutines - no UI updates needed.
func (a *App) exportPEDataToCSV(filename string) error {
	// Copy data under lock
	a.mu.RLock()
	eData, pData := a.historySnapshotLocked()
	a.mu.RUnlock()

	// Create CSV file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"E_field_MV_cm", "Polarization_uC_cm2"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for i := range eData {
		eMVcm := sharedphysics.VPerMToMVPerCm(eData[i]) // V/m → MV/cm
		pUcCm2 := pData[i] * 1e2 // C/m² to μC/cm²

		row := []string{
			fmt.Sprintf("%.6f", eMVcm),
			fmt.Sprintf("%.6f", pUcCm2),
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row %d: %w", i, err)
		}
	}

	log.Info("Exported P-E data to CSV: %s (%d points)", filename, len(eData))
	return nil
}

// exportPEData exports P-E hysteresis data to both JSON and CSV formats
// THREAD-SAFE: Safe to call from goroutines. Uses fyne.Do() for UI updates.
func (a *App) exportPEData() {
	// Create data directory if it doesn't exist
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("Error creating data directory: %v", err)
		fyne.Do(func() {
			a.setStatus(fmt.Sprintf("Export failed: %v", err))
		})
		return
	}

	// Generate timestamped filenames
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	jsonFile := filepath.Join(dataDir, fmt.Sprintf("pe-data-%s.json", timestamp))
	csvFile := filepath.Join(dataDir, fmt.Sprintf("pe-data-%s.csv", timestamp))

	// Check if there's data to export
	a.mu.RLock()
	dataPoints := a.historyLengthLocked()
	a.mu.RUnlock()

	if dataPoints == 0 {
		log.Info("No P-E data to export")
		fyne.Do(func() {
			a.setStatus("No data to export")
		})
		return
	}

	// Export JSON
	if err := a.exportPEDataToJSON(jsonFile); err != nil {
		log.Printf("Error exporting JSON: %v", err)
		fyne.Do(func() {
			a.setStatus(fmt.Sprintf("JSON export failed: %v", err))
		})
		return
	}

	// Export CSV
	if err := a.exportPEDataToCSV(csvFile); err != nil {
		log.Printf("Error exporting CSV: %v", err)
		fyne.Do(func() {
			a.setStatus(fmt.Sprintf("CSV export failed: %v", err))
		})
		return
	}

	// Success - update status
	fyne.Do(func() {
		a.setStatus(fmt.Sprintf("Exported %d points to data/", dataPoints))
	})
}

// setStatus updates the status label (must be called from main UI thread via fyne.Do)
func (a *App) setStatus(message string) {
	if a.statusLabel != nil {
		a.statusLabel.SetText(message)
	}
}

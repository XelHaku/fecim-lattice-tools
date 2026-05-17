//go:build legacy_fyne

package gui

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"

	sharedexport "fecim-lattice-tools/shared/export"
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

func defaultCSVColumns() []string {
	return []string{"e_field_mv_cm", "polarization_uc_cm2"}
}

func resolveCSVColumnsFromEnv() []string {
	raw := strings.TrimSpace(os.Getenv("FECIM_EXPORT_COLUMNS"))
	if raw == "" {
		return defaultCSVColumns()
	}

	valid := map[string]bool{
		"index":               true,
		"e_field_mv_cm":       true,
		"polarization_uc_cm2": true,
		"e_field_v_m":         true,
		"polarization_c_m2":   true,
	}

	out := make([]string, 0, 5)
	seen := make(map[string]bool)
	for _, tok := range strings.Split(raw, ",") {
		key := strings.ToLower(strings.TrimSpace(tok))
		if key == "" || seen[key] || !valid[key] {
			continue
		}
		seen[key] = true
		out = append(out, key)
	}

	if len(out) == 0 {
		return defaultCSVColumns()
	}
	return out
}

func csvHeaderForColumns(cols []string) []string {
	headers := make([]string, 0, len(cols))
	for _, c := range cols {
		switch c {
		case "index":
			headers = append(headers, "Index")
		case "e_field_mv_cm":
			headers = append(headers, "E_field_MV_cm")
		case "polarization_uc_cm2":
			headers = append(headers, "Polarization_uC_cm2")
		case "e_field_v_m":
			headers = append(headers, "E_field_V_m")
		case "polarization_c_m2":
			headers = append(headers, "Polarization_C_m2")
		}
	}
	return headers
}

func csvRowForColumns(cols []string, i int, e, p float64) []string {
	row := make([]string, 0, len(cols))
	for _, c := range cols {
		switch c {
		case "index":
			row = append(row, fmt.Sprintf("%d", i))
		case "e_field_mv_cm":
			row = append(row, fmt.Sprintf("%.6f", sharedphysics.VPerMToMVPerCm(e)))
		case "polarization_uc_cm2":
			row = append(row, fmt.Sprintf("%.6f", p*1e2))
		case "e_field_v_m":
			row = append(row, fmt.Sprintf("%.6e", e))
		case "polarization_c_m2":
			row = append(row, fmt.Sprintf("%.6e", p))
		}
	}
	return row
}

func buildCSVContent(eData, pData []float64, cols []string) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if err := writer.Write(csvHeaderForColumns(cols)); err != nil {
		return nil, err
	}
	for i := range eData {
		if err := writer.Write(csvRowForColumns(cols, i, eData[i], pData[i])); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// exportPEDataToJSON exports P-E hysteresis data to a JSON file
func (a *App) exportPEDataToJSON(filename string) error {
	a.mu.RLock()
	eData, pData := a.historySnapshotLocked()

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

	eFieldMVcm := make([]float64, len(eData))
	polarizationUcCm2 := make([]float64, len(pData))
	for i := range eData {
		eFieldMVcm[i] = sharedphysics.VPerMToMVPerCm(eData[i])
		polarizationUcCm2[i] = pData[i] * 1e2
	}

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

	if err := sharedio.SaveJSON(filename, export); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	log.Info("Exported P-E data to JSON: %s (%d points)", filename, len(eData))
	return nil
}

// exportPEDataToCSV exports P-E hysteresis data to CSV (configurable via FECIM_EXPORT_COLUMNS)
func (a *App) exportPEDataToCSV(filename string) error {
	a.mu.RLock()
	eData, pData := a.historySnapshotLocked()
	a.mu.RUnlock()

	cols := resolveCSVColumnsFromEnv()
	content, err := buildCSVContent(eData, pData, cols)
	if err != nil {
		return fmt.Errorf("failed to build CSV data: %w", err)
	}
	if err := os.WriteFile(filename, content, 0644); err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}

	log.Info("Exported P-E data to CSV: %s (%d points, columns=%s)", filename, len(eData), strings.Join(cols, ","))
	return nil
}

// exportPEDataToClipboard copies CSV-formatted P-E data to the system clipboard.
func (a *App) exportPEDataToClipboard() error {
	a.mu.RLock()
	eData, pData := a.historySnapshotLocked()
	a.mu.RUnlock()

	if len(eData) == 0 {
		return fmt.Errorf("no data to export")
	}
	if a.mainWindow == nil || a.mainWindow.Clipboard() == nil {
		return fmt.Errorf("clipboard unavailable")
	}

	content, err := buildCSVContent(eData, pData, resolveCSVColumnsFromEnv())
	if err != nil {
		return fmt.Errorf("failed to build clipboard CSV: %w", err)
	}

	fyne.Do(func() {
		a.mainWindow.Clipboard().SetContent(string(content))
		a.setStatus(fmt.Sprintf("Copied %d points to clipboard", len(eData)))
	})
	return nil
}

// exportPEData exports P-E hysteresis data to both JSON and CSV formats
func (a *App) exportPEData() {
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("Error creating data directory: %v", err)
		fyne.Do(func() {
			a.setStatus(fmt.Sprintf("Export failed: %v", err))
		})
		return
	}

	timestamp := time.Now().Format("2006-01-02T15-04-05")
	jsonFile := filepath.Join(dataDir, fmt.Sprintf("pe-data-%s.json", timestamp))
	csvFile := filepath.Join(dataDir, fmt.Sprintf("pe-data-%s.csv", timestamp))

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

	if err := a.exportPEDataToJSON(jsonFile); err != nil {
		log.Printf("Error exporting JSON: %v", err)
		fyne.Do(func() {
			a.setStatus(fmt.Sprintf("JSON export failed: %v", err))
		})
		return
	}

	if err := a.exportPEDataToCSV(csvFile); err != nil {
		log.Printf("Error exporting CSV: %v", err)
		fyne.Do(func() {
			a.setStatus(fmt.Sprintf("CSV export failed: %v", err))
		})
		return
	}

	// SVG vector export for publication-quality figures
	svgFile := filepath.Join(dataDir, fmt.Sprintf("pe-loop-%s.svg", timestamp))
	if err := a.exportPEDataToSVG(svgFile); err != nil {
		log.Printf("Warning: SVG export failed: %v", err)
		// Non-fatal — CSV/JSON already exported
	}

	fyne.Do(func() {
		a.setStatus(fmt.Sprintf("Exported %d points to data/ (JSON+CSV+SVG)", dataPoints))
	})
}

// exportPEDataToSVG exports a publication-quality SVG of the P-E hysteresis loop.
func (a *App) exportPEDataToSVG(filename string) error {
	a.mu.RLock()
	eHist := make([]float64, len(a.eHistory))
	pHist := make([]float64, len(a.pHistory))
	copy(eHist, a.eHistory)
	copy(pHist, a.pHistory)
	matName := ""
	if a.material != nil {
		matName = a.material.Name
	}
	a.mu.RUnlock()

	if len(eHist) < 2 {
		return fmt.Errorf("not enough data points for SVG export")
	}

	cfg := sharedexport.DefaultSVGPlotConfig()
	cfg.Title = fmt.Sprintf("P-E Hysteresis Loop — %s", matName)
	cfg.Citation = fmt.Sprintf("FeCIM Lattice Tools | %s | %s", matName, time.Now().Format("2006-01-02"))

	svgContent, err := sharedexport.GeneratePELoopSVG(eHist, pHist, cfg)
	if err != nil {
		return fmt.Errorf("SVG generation failed: %w", err)
	}

	return os.WriteFile(filename, []byte(svgContent), 0644)
}

// setStatus updates the status label (must be called from main UI thread via fyne.Do)
func (a *App) setStatus(message string) {
	if a.statusLabel != nil {
		a.statusLabel.SetText(message)
	}
}

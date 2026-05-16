// Package export provides unified data export utilities for FeCIM tools.
// Supports CSV, JSON, and PNG export for simulation results, configurations,
// and visualizations.
package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExportFormat represents the type of export
type ExportFormat string

const (
	FormatCSV  ExportFormat = "csv"
	FormatHTML ExportFormat = "html"
	FormatJSON ExportFormat = "json"
	FormatPNG  ExportFormat = "png"
	FormatSVG  ExportFormat = "svg"
)

// ExportMetadata contains common metadata for all exports
type ExportMetadata struct {
	ModuleName   string            `json:"module_name"`
	ExportedAt   time.Time         `json:"exported_at"`
	Version      string            `json:"version"`
	Description  string            `json:"description,omitempty"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
}

// NewExportMetadata creates a new metadata instance with defaults
func NewExportMetadata(moduleName string) *ExportMetadata {
	return &ExportMetadata{
		ModuleName:   moduleName,
		ExportedAt:   time.Now(),
		Version:      "1.0.0",
		CustomFields: make(map[string]string),
	}
}

// ExportResult represents the result of an export operation
type ExportResult struct {
	FilePath     string
	Format       ExportFormat
	BytesWritten int64
	Error        error
}

// Exporter provides a unified interface for exporting data
type Exporter struct {
	OutputDir string
	Prefix    string
}

// NewExporter creates a new Exporter with the given output directory and file prefix
func NewExporter(outputDir, prefix string) *Exporter {
	return &Exporter{
		OutputDir: outputDir,
		Prefix:    prefix,
	}
}

// ensureOutputDir creates the output directory if it doesn't exist
func (e *Exporter) ensureOutputDir() error {
	return os.MkdirAll(e.OutputDir, 0755)
}

// generateFilename creates a timestamped filename with the given extension
func (e *Exporter) generateFilename(extension string) string {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return filepath.Join(e.OutputDir, fmt.Sprintf("%s_%s.%s", e.Prefix, timestamp, extension))
}

// ExportCSV exports tabular data to a CSV file
func (e *Exporter) ExportCSV(headers []string, rows [][]string) *ExportResult {
	result := &ExportResult{Format: FormatCSV}

	if err := e.ensureOutputDir(); err != nil {
		result.Error = fmt.Errorf("failed to create output directory: %w", err)
		return result
	}

	result.FilePath = e.generateFilename("csv")
	file, err := os.Create(result.FilePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to create CSV file: %w", err)
		return result
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(headers); err != nil {
		result.Error = fmt.Errorf("failed to write CSV header: %w", err)
		return result
	}

	// Write rows
	for i, row := range rows {
		if err := writer.Write(row); err != nil {
			result.Error = fmt.Errorf("failed to write CSV row %d: %w", i, err)
			return result
		}
	}

	// Get file size
	if info, err := file.Stat(); err == nil {
		result.BytesWritten = info.Size()
	}

	return result
}

// ExportCSVFromFloats exports float64 data columns to CSV
// ExportHTMLTable exports tabular data as an accessible HTML table.
// The output includes <caption>, <thead>, and semantic table markup for screen readers.
func (e *Exporter) ExportHTMLTable(title string, headers []string, rows [][]string) *ExportResult {
	result := &ExportResult{Format: FormatHTML}

	if err := e.ensureOutputDir(); err != nil {
		result.Error = fmt.Errorf("failed to create output directory: %w", err)
		return result
	}

	result.FilePath = e.generateFilename("html")

	var b strings.Builder
	b.WriteString("<!doctype html><html lang=\"en\"><head><meta charset=\"utf-8\">")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">")
	b.WriteString("<title>")
	b.WriteString(html.EscapeString(title))
	b.WriteString("</title>")
	b.WriteString("<style>body{font-family:sans-serif;padding:1rem;}table{border-collapse:collapse;width:100%;}th,td{border:1px solid #ccc;padding:.4rem;text-align:left;}caption{font-weight:700;text-align:left;margin-bottom:.5rem;}</style>")
	b.WriteString("</head><body><table>")
	b.WriteString("<caption>")
	b.WriteString(html.EscapeString(title))
	b.WriteString("</caption><thead><tr>")
	for _, header := range headers {
		b.WriteString("<th scope=\"col\">")
		b.WriteString(html.EscapeString(header))
		b.WriteString("</th>")
	}
	b.WriteString("</tr></thead><tbody>")
	for _, row := range rows {
		b.WriteString("<tr>")
		for _, col := range row {
			b.WriteString("<td>")
			b.WriteString(html.EscapeString(col))
			b.WriteString("</td>")
		}
		b.WriteString("</tr>")
	}
	b.WriteString("</tbody></table></body></html>")

	payload := []byte(b.String())
	if err := os.WriteFile(result.FilePath, payload, 0644); err != nil {
		result.Error = fmt.Errorf("failed to write HTML file: %w", err)
		return result
	}
	result.BytesWritten = int64(len(payload))
	return result
}

func (e *Exporter) ExportCSVFromFloats(headers []string, columns ...[]float64) *ExportResult {
	if len(columns) == 0 {
		return &ExportResult{Error: fmt.Errorf("no data columns provided")}
	}

	// Find the maximum length
	maxLen := 0
	for _, col := range columns {
		if len(col) > maxLen {
			maxLen = len(col)
		}
	}

	// Convert to string rows
	rows := make([][]string, maxLen)
	for i := 0; i < maxLen; i++ {
		row := make([]string, len(columns))
		for j, col := range columns {
			if i < len(col) {
				row[j] = fmt.Sprintf("%.9g", col[i])
			} else {
				row[j] = ""
			}
		}
		rows[i] = row
	}

	return e.ExportCSV(headers, rows)
}

// ExportJSON exports data to a JSON file with pretty formatting
func (e *Exporter) ExportJSON(data interface{}) *ExportResult {
	result := &ExportResult{Format: FormatJSON}

	if err := e.ensureOutputDir(); err != nil {
		result.Error = fmt.Errorf("failed to create output directory: %w", err)
		return result
	}

	result.FilePath = e.generateFilename("json")

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		result.Error = fmt.Errorf("failed to marshal JSON: %w", err)
		return result
	}

	if err := os.WriteFile(result.FilePath, jsonBytes, 0644); err != nil {
		result.Error = fmt.Errorf("failed to write JSON file: %w", err)
		return result
	}

	result.BytesWritten = int64(len(jsonBytes))
	return result
}

// ExportPNG exports an image to a PNG file
func (e *Exporter) ExportPNG(img image.Image) *ExportResult {
	result := &ExportResult{Format: FormatPNG}

	if err := e.ensureOutputDir(); err != nil {
		result.Error = fmt.Errorf("failed to create output directory: %w", err)
		return result
	}

	result.FilePath = e.generateFilename("png")
	file, err := os.Create(result.FilePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to create PNG file: %w", err)
		return result
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		result.Error = fmt.Errorf("failed to encode PNG: %w", err)
		return result
	}

	// Get file size
	if info, err := file.Stat(); err == nil {
		result.BytesWritten = info.Size()
	}

	return result
}

// QuickExport provides one-click export with automatic filename generation
func QuickExport(outputDir, prefix string, format ExportFormat, data interface{}) *ExportResult {
	exporter := NewExporter(outputDir, prefix)

	switch format {
	case FormatJSON:
		return exporter.ExportJSON(data)
	case FormatCSV:
		// For CSV, data should be CSVData
		if csvData, ok := data.(*CSVData); ok {
			return exporter.ExportCSV(csvData.Headers, csvData.Rows)
		}
		return &ExportResult{Error: fmt.Errorf("CSV export requires *CSVData type")}
	case FormatHTML:
		if csvData, ok := data.(*CSVData); ok {
			return exporter.ExportHTMLTable("FeCIM Export", csvData.Headers, csvData.Rows)
		}
		return &ExportResult{Error: fmt.Errorf("HTML export requires *CSVData type")}
	case FormatPNG:
		if img, ok := data.(image.Image); ok {
			return exporter.ExportPNG(img)
		}
		return &ExportResult{Error: fmt.Errorf("PNG export requires image.Image type")}
	default:
		return &ExportResult{Error: fmt.Errorf("unsupported format: %s", format)}
	}
}

// CSVData represents tabular data for CSV export
type CSVData struct {
	Headers []string
	Rows    [][]string
}

// NewCSVData creates a new CSVData with the given headers
func NewCSVData(headers ...string) *CSVData {
	return &CSVData{
		Headers: headers,
		Rows:    make([][]string, 0),
	}
}

// AddRow adds a row of string values
func (c *CSVData) AddRow(values ...string) {
	c.Rows = append(c.Rows, values)
}

// AddRowFromFloats adds a row of float64 values formatted as strings
func (c *CSVData) AddRowFromFloats(values ...float64) {
	row := make([]string, len(values))
	for i, v := range values {
		row[i] = fmt.Sprintf("%.9g", v)
	}
	c.Rows = append(c.Rows, row)
}

// AddRowFromInts adds a row of int values formatted as strings
func (c *CSVData) AddRowFromInts(values ...int) {
	row := make([]string, len(values))
	for i, v := range values {
		row[i] = fmt.Sprintf("%d", v)
	}
	c.Rows = append(c.Rows, row)
}

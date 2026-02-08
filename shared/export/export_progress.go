// Package export provides progress-aware export utilities for FeCIM tools.
// This file adds progress tracking, ETA, and cancellation support to exports.
package export

import (
	"context"
	"encoding/csv"
	"fmt"
	"image"
	"image/png"
	"os"

	"fecim-lattice-tools/shared/progress"
)

// ProgressExporter provides export functions with progress tracking
type ProgressExporter struct {
	*Exporter
	Progress *progress.Progress
}

// NewProgressExporter creates an exporter with progress tracking
func NewProgressExporter(outputDir, prefix, operation string, totalRows int64) *ProgressExporter {
	return &ProgressExporter{
		Exporter: NewExporter(outputDir, prefix),
		Progress: progress.NewProgress(operation, totalRows),
	}
}

// NewProgressExporterWithContext creates an exporter with external cancellation context
func NewProgressExporterWithContext(ctx context.Context, outputDir, prefix, operation string, totalRows int64) *ProgressExporter {
	return &ProgressExporter{
		Exporter: NewExporter(outputDir, prefix),
		Progress: progress.NewProgressWithContext(ctx, operation, totalRows),
	}
}

// ExportCSVWithProgress exports CSV data with progress tracking and cancellation support
func (pe *ProgressExporter) ExportCSVWithProgress(headers []string, rows [][]string) *ExportResult {
	result := &ExportResult{Format: FormatCSV}
	p := pe.Progress

	// Update total if needed
	p.SetTotal(int64(len(rows)))
	p.Start()
	p.SetPhase("Preparing export")

	if err := pe.ensureOutputDir(); err != nil {
		p.Fail(err)
		result.Error = fmt.Errorf("failed to create output directory: %w", err)
		return result
	}

	result.FilePath = pe.generateFilename("csv")
	p.SetDetail(fmt.Sprintf("Creating %s", result.FilePath))

	file, err := os.Create(result.FilePath)
	if err != nil {
		p.Fail(err)
		result.Error = fmt.Errorf("failed to create CSV file: %w", err)
		return result
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	p.SetPhase("Writing header")
	if err := writer.Write(headers); err != nil {
		p.Fail(err)
		result.Error = fmt.Errorf("failed to write CSV header: %w", err)
		return result
	}

	// Write rows with progress
	p.SetPhase("Writing data")
	for i, row := range rows {
		// Check for cancellation
		select {
		case <-p.Context().Done():
			result.Error = context.Canceled
			return result
		default:
		}

		if err := writer.Write(row); err != nil {
			p.Fail(err)
			result.Error = fmt.Errorf("failed to write CSV row %d: %w", i, err)
			return result
		}

		// Update progress every 100 rows or at end
		if i%100 == 0 || i == len(rows)-1 {
			p.UpdateWithStatus(int64(i+1), "Writing data",
				fmt.Sprintf("Row %d of %d", i+1, len(rows)))
		}
	}

	// Finalize
	p.SetPhase("Finalizing")
	writer.Flush()

	// Get file size
	if info, err := file.Stat(); err == nil {
		result.BytesWritten = info.Size()
		p.SetDetail(fmt.Sprintf("Written %s (%d bytes)", result.FilePath, result.BytesWritten))
	}

	p.Complete()
	return result
}

// ExportCSVFromFloatsWithProgress exports float64 columns with progress tracking
func (pe *ProgressExporter) ExportCSVFromFloatsWithProgress(headers []string, columns ...[]float64) *ExportResult {
	if len(columns) == 0 {
		pe.Progress.Fail(fmt.Errorf("no data columns provided"))
		return &ExportResult{Error: fmt.Errorf("no data columns provided")}
	}

	pe.Progress.Start()
	pe.Progress.SetPhase("Converting data")

	// Find the maximum length
	maxLen := 0
	for _, col := range columns {
		if len(col) > maxLen {
			maxLen = len(col)
		}
	}

	pe.Progress.SetTotal(int64(maxLen))

	// Convert to string rows with progress
	rows := make([][]string, maxLen)
	for i := 0; i < maxLen; i++ {
		// Check cancellation
		select {
		case <-pe.Progress.Context().Done():
			return &ExportResult{Error: context.Canceled}
		default:
		}

		row := make([]string, len(columns))
		for j, col := range columns {
			if i < len(col) {
				row[j] = fmt.Sprintf("%.9g", col[i])
			} else {
				row[j] = ""
			}
		}
		rows[i] = row

		if i%1000 == 0 {
			pe.Progress.UpdateWithStatus(int64(i), "Converting data",
				fmt.Sprintf("Converting row %d of %d", i, maxLen))
		}
	}

	// Reset progress for writing phase
	pe.Progress.SetTotal(int64(len(rows)))
	pe.Progress.Update(0)

	return pe.ExportCSVWithProgress(headers, rows)
}

// ExportPNGWithProgress exports an image with progress tracking
func (pe *ProgressExporter) ExportPNGWithProgress(img image.Image) *ExportResult {
	result := &ExportResult{Format: FormatPNG}
	p := pe.Progress

	// PNG export is typically fast, use indeterminate progress
	p.SetTotal(0) // Indeterminate
	p.Start()
	p.SetPhase("Preparing image export")

	if err := pe.ensureOutputDir(); err != nil {
		p.Fail(err)
		result.Error = fmt.Errorf("failed to create output directory: %w", err)
		return result
	}

	result.FilePath = pe.generateFilename("png")
	p.SetDetail(fmt.Sprintf("Creating %s", result.FilePath))

	file, err := os.Create(result.FilePath)
	if err != nil {
		p.Fail(err)
		result.Error = fmt.Errorf("failed to create PNG file: %w", err)
		return result
	}
	defer file.Close()

	p.SetPhase("Encoding PNG")
	p.SetDetail(fmt.Sprintf("Image size: %dx%d", img.Bounds().Dx(), img.Bounds().Dy()))

	if err := png.Encode(file, img); err != nil {
		p.Fail(err)
		result.Error = fmt.Errorf("failed to encode PNG: %w", err)
		return result
	}

	// Get file size
	if info, err := file.Stat(); err == nil {
		result.BytesWritten = info.Size()
		p.SetDetail(fmt.Sprintf("Written %s (%d bytes)", result.FilePath, result.BytesWritten))
	}

	p.Complete()
	return result
}

// Cancel cancels the ongoing export
func (pe *ProgressExporter) Cancel() {
	pe.Progress.Cancel()
}

// IsCancelled returns true if the export was cancelled
func (pe *ProgressExporter) IsCancelled() bool {
	return pe.Progress.IsCancelled()
}

// BatchExporter exports multiple items with overall progress tracking
type BatchExporter struct {
	*Exporter
	Progress   *progress.Progress
	Items      int
	Completed  int
	Results    []*ExportResult
	OnProgress func(item int, total int, result *ExportResult)
}

// NewBatchExporter creates a batch exporter for multiple items
func NewBatchExporter(outputDir, prefix string, itemCount int) *BatchExporter {
	return &BatchExporter{
		Exporter: NewExporter(outputDir, prefix),
		Progress: progress.NewProgress("Batch Export", int64(itemCount)),
		Items:    itemCount,
		Results:  make([]*ExportResult, 0, itemCount),
	}
}

// Start begins the batch export
func (be *BatchExporter) Start() {
	be.Progress.Start()
	be.Progress.SetPhase("Starting batch export")
}

// ExportItem exports a single item in the batch
func (be *BatchExporter) ExportItem(name string, format ExportFormat, data interface{}) *ExportResult {
	be.Completed++
	be.Progress.UpdateWithStatus(
		int64(be.Completed),
		fmt.Sprintf("Exporting %s", name),
		fmt.Sprintf("Item %d of %d", be.Completed, be.Items),
	)

	// Check cancellation
	if be.Progress.IsCancelled() {
		result := &ExportResult{Error: context.Canceled}
		be.Results = append(be.Results, result)
		return result
	}

	// Perform export
	result := QuickExport(be.OutputDir, be.Prefix+"-"+name, format, data)
	be.Results = append(be.Results, result)

	if be.OnProgress != nil {
		be.OnProgress(be.Completed, be.Items, result)
	}

	return result
}

// Complete finishes the batch export
func (be *BatchExporter) Complete() {
	be.Progress.Complete()
}

// Cancel cancels the batch export
func (be *BatchExporter) Cancel() {
	be.Progress.Cancel()
}

// Summary returns a summary of the batch export
func (be *BatchExporter) Summary() (succeeded, failed int) {
	for _, r := range be.Results {
		if r.Error == nil {
			succeeded++
		} else {
			failed++
		}
	}
	return
}

// SimulationExporter is specialized for exporting simulation results
type SimulationExporter struct {
	*ProgressExporter
	SimulationName string
	Parameters     map[string]interface{}
	DataPoints     int
}

// NewSimulationExporter creates an exporter optimized for simulation results
func NewSimulationExporter(outputDir, simName string, dataPoints int) *SimulationExporter {
	return &SimulationExporter{
		ProgressExporter: NewProgressExporter(outputDir, simName, "Exporting "+simName, int64(dataPoints)),
		SimulationName:   simName,
		Parameters:       make(map[string]interface{}),
		DataPoints:       dataPoints,
	}
}

// SetParameter records a simulation parameter for metadata
func (se *SimulationExporter) SetParameter(key string, value interface{}) {
	se.Parameters[key] = value
}

// ExportWithMetadata exports data along with simulation metadata
func (se *SimulationExporter) ExportWithMetadata(headers []string, rows [][]string) (*ExportResult, *ExportResult) {
	// Export data
	se.Progress.SetPhase("Exporting simulation data")
	dataResult := se.ExportCSVWithProgress(headers, rows)

	if dataResult.Error != nil {
		return dataResult, nil
	}

	// Export metadata
	se.Progress.SetPhase("Exporting metadata")
	metadata := ExportMetadata{
		ModuleName:  se.SimulationName,
		Version:     "1.0.0",
		Description: fmt.Sprintf("%s simulation results", se.SimulationName),
		CustomFields: map[string]string{
			"data_points": fmt.Sprintf("%d", se.DataPoints),
			"data_file":   dataResult.FilePath,
		},
	}

	// Add parameters to metadata
	for k, v := range se.Parameters {
		metadata.CustomFields[k] = fmt.Sprintf("%v", v)
	}

	metaResult := se.Exporter.ExportJSON(metadata)
	return dataResult, metaResult
}

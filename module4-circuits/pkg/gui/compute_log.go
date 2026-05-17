//go:build legacy_fyne

// Package gui provides compute logging for debugging MVM operations.
package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	sharedio "fecim-lattice-tools/shared/io"
	"fecim-lattice-tools/shared/logging"
)

// ComputeLogEntry represents a single MVM compute operation
type ComputeLogEntry struct {
	Timestamp    string             `json:"timestamp"`
	ArraySize    string             `json:"array_size"`
	Material     string             `json:"material"`
	QuantLevels  int                `json:"quant_levels"`
	InputVector  []float64          `json:"input_vector_volts"`
	Weights      [][]int            `json:"weight_matrix"`
	Conductances [][]float64        `json:"conductance_matrix_uS"`
	RowResults   []ComputeRowResult `json:"row_results"`
}

// ComputeRowResult holds the compute result for a single row
type ComputeRowResult struct {
	Row        int       `json:"row"`
	Active     bool      `json:"active"`
	CurrentUA  float64   `json:"current_uA"`
	TIAVoltage float64   `json:"tia_voltage_V"`
	ADCLevel   int       `json:"adc_level"`
	Saturated  bool      `json:"saturated"`
	CellDetail []CellMVM `json:"cell_details"`
}

// CellMVM holds per-cell MVM calculation details
type CellMVM struct {
	Col           int     `json:"col"`
	Weight        int     `json:"weight"`
	ConductanceUS float64 `json:"conductance_uS"`
	VoltageV      float64 `json:"voltage_V"`
	CurrentUA     float64 `json:"current_uA"`
}

// ComputeLog manages the compute log file
type ComputeLog struct {
	mu       sync.Mutex
	entries  []ComputeLogEntry
	filePath string
	enabled  bool
}

var globalComputeLog = &ComputeLog{
	entries:  make([]ComputeLogEntry, 0),
	filePath: filepath.Join(logging.LogsDir(), "compute_log.json"),
	enabled:  os.Getenv("FECIM_CIRCUITS_COMPUTE_LOG") != "",
}

// EnableComputeLog enables or disables compute logging
func EnableComputeLog(enabled bool) {
	globalComputeLog.mu.Lock()
	defer globalComputeLog.mu.Unlock()
	globalComputeLog.enabled = enabled
}

// ComputeLogEnabled reports whether compute logging is currently enabled.
func ComputeLogEnabled() bool {
	globalComputeLog.mu.Lock()
	defer globalComputeLog.mu.Unlock()
	return globalComputeLog.enabled
}

// SetComputeLogPath sets the path for the compute log file.
// Returns an error if the path is empty or contains traversal sequences.
func SetComputeLogPath(path string) error {
	cleanPath, err := sharedio.ValidatePath(path)
	if err != nil {
		return fmt.Errorf("invalid compute log path: %w", err)
	}
	globalComputeLog.mu.Lock()
	defer globalComputeLog.mu.Unlock()
	globalComputeLog.filePath = cleanPath
	return nil
}

// ClearComputeLog clears all logged entries
func ClearComputeLog() {
	globalComputeLog.mu.Lock()
	defer globalComputeLog.mu.Unlock()
	globalComputeLog.entries = make([]ComputeLogEntry, 0)
}

// GetComputeLogEntries returns a copy of all logged entries
func GetComputeLogEntries() []ComputeLogEntry {
	globalComputeLog.mu.Lock()
	defer globalComputeLog.mu.Unlock()
	result := make([]ComputeLogEntry, len(globalComputeLog.entries))
	copy(result, globalComputeLog.entries)
	return result
}

// LogCompute logs a compute operation (called from DeviceState.Compute)
func (ds *DeviceState) LogCompute(weights [][]int, quantLevels int) {
	if !globalComputeLog.enabled {
		return
	}

	globalComputeLog.mu.Lock()
	defer globalComputeLog.mu.Unlock()

	entry := ComputeLogEntry{
		Timestamp:   time.Now().Format("2006-01-02 15:04:05.000"),
		ArraySize:   fmt.Sprintf("%dx%d", ds.rows, ds.cols),
		QuantLevels: quantLevels,
	}

	// Material name
	if ds.material != nil {
		entry.Material = ds.material.Name
	} else {
		entry.Material = "default"
	}

	// Input vector (DAC voltages)
	entry.InputVector = make([]float64, ds.cols)
	copy(entry.InputVector, ds.dacVoltages)

	// Weight matrix
	entry.Weights = make([][]int, len(weights))
	for r := range weights {
		entry.Weights[r] = make([]int, len(weights[r]))
		copy(entry.Weights[r], weights[r])
	}

	// Conductance matrix
	entry.Conductances = make([][]float64, ds.rows)
	for r := 0; r < ds.rows; r++ {
		entry.Conductances[r] = make([]float64, ds.cols)
		for c := 0; c < ds.cols; c++ {
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}
			var conductanceS float64
			if ds.material != nil {
				conductanceS = ds.material.DiscreteLevel(level, quantLevels)
			} else {
				conductanceS = (1.0 + float64(level)/float64(quantLevels-1)*99.0) * 1e-6
			}
			entry.Conductances[r][c] = conductanceS * 1e6 // Convert to µS
		}
	}

	// Row results with cell details
	entry.RowResults = make([]ComputeRowResult, ds.rows)
	for r := 0; r < ds.rows; r++ {
		result := ComputeRowResult{
			Row:        r,
			Active:     ds.activeRows[r],
			CurrentUA:  ds.rowCurrents[r],
			TIAVoltage: ds.rowVoltages[r],
			ADCLevel:   ds.rowLevels[r],
			Saturated:  ds.saturated[r],
			CellDetail: make([]CellMVM, ds.cols),
		}

		// Per-cell breakdown
		for c := 0; c < ds.cols; c++ {
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}
			conductanceUS := entry.Conductances[r][c]
			voltage := ds.dacVoltages[c]
			currentUA := conductanceUS * voltage
			if ds.couplingMode == arraysim.CouplingTierA && ds.coupledCellVoltages != nil {
				if r < len(ds.coupledCellVoltages) && c < len(ds.coupledCellVoltages[r]) {
					voltage = ds.coupledCellVoltages[r][c]
				}
			}
			if ds.couplingMode == arraysim.CouplingTierA && ds.coupledCellCurrents != nil {
				if r < len(ds.coupledCellCurrents) && c < len(ds.coupledCellCurrents[r]) {
					currentUA = ds.coupledCellCurrents[r][c] * 1e6
				}
			}

			result.CellDetail[c] = CellMVM{
				Col:           c,
				Weight:        level,
				ConductanceUS: conductanceUS,
				VoltageV:      voltage,
				CurrentUA:     currentUA,
			}
		}
		entry.RowResults[r] = result
	}

	globalComputeLog.entries = append(globalComputeLog.entries, entry)

	// Keep only last 100 entries to prevent memory bloat
	if len(globalComputeLog.entries) > 100 {
		globalComputeLog.entries = globalComputeLog.entries[len(globalComputeLog.entries)-100:]
	}
}

// SaveComputeLog saves all logged entries to the JSON file
func SaveComputeLog() error {
	globalComputeLog.mu.Lock()
	defer globalComputeLog.mu.Unlock()

	if len(globalComputeLog.entries) == 0 {
		return nil
	}

	if err := sharedio.SaveJSON(globalComputeLog.filePath, globalComputeLog.entries); err != nil {
		return fmt.Errorf("failed to write compute log: %w", err)
	}

	return nil
}

// SaveComputeLogTo saves all logged entries to a specific file
func SaveComputeLogTo(path string) error {
	globalComputeLog.mu.Lock()
	entries := make([]ComputeLogEntry, len(globalComputeLog.entries))
	copy(entries, globalComputeLog.entries)
	globalComputeLog.mu.Unlock()

	if len(entries) == 0 {
		return nil
	}

	if err := sharedio.SaveJSON(path, entries); err != nil {
		return fmt.Errorf("failed to write compute log: %w", err)
	}

	return nil
}

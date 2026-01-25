// pkg/export/csv.go
package export

import (
	"encoding/csv"
	"fmt"
	"os"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/compiler"
)

// ExportCSV writes the array design to a CSV file.
// Works with all operation modes (Storage, Memory, Compute).
// For compute mode with weights, includes initial_weight column.
// For storage/memory modes, initial_weight is empty.
func ExportCSV(design *compiler.ArrayDesign, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Filter cells to export
	cellsToExport := filterActiveCells(design)

	// Determine if this is compute mode with weights
	hasWeights := design.Config.Mode == compiler.ModeCompute &&
		design.Config.ComputeConfig != nil &&
		design.Config.ComputeConfig.InitialWeights != nil

	// Header - use "weight" and "level" columns for backward compatibility with existing tools
	var header []string
	if hasWeights {
		header = []string{"row", "col", "weight", "level", "conductance_uS", "resistance_ohm", "program_V"}
	} else {
		header = []string{"row", "col", "level", "conductance_uS", "resistance_ohm", "program_V"}
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Data rows
	for _, cell := range cellsToExport {
		var record []string
		if hasWeights {
			record = []string{
				fmt.Sprintf("%d", cell.Row),
				fmt.Sprintf("%d", cell.Col),
				fmt.Sprintf("%.6f", cell.InitialWeight),
				fmt.Sprintf("%d", cell.Level),
				fmt.Sprintf("%.4f", cell.Conductance),
				fmt.Sprintf("%.2f", cell.Resistance),
				fmt.Sprintf("%.4f", cell.ProgramV),
			}
		} else {
			record = []string{
				fmt.Sprintf("%d", cell.Row),
				fmt.Sprintf("%d", cell.Col),
				fmt.Sprintf("%d", cell.Level),
				fmt.Sprintf("%.4f", cell.Conductance),
				fmt.Sprintf("%.2f", cell.Resistance),
				fmt.Sprintf("%.4f", cell.ProgramV),
			}
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

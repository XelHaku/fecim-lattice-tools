// pkg/export/csv.go
package export

import (
	"encoding/csv"
	"fmt"
	"os"

	"demo6-eda/pkg/compiler"
)

// ExportCSV writes the crossbar mapping to a CSV file
func ExportCSV(mapping *compiler.CrossbarMapping, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	if err := writer.Write([]string{"row", "col", "weight", "level", "conductance_uS", "program_V"}); err != nil {
		return err
	}

	// Data rows
	for _, cell := range mapping.Cells {
		record := []string{
			fmt.Sprintf("%d", cell.Row),
			fmt.Sprintf("%d", cell.Col),
			fmt.Sprintf("%.6f", cell.WeightValue),
			fmt.Sprintf("%d", cell.QuantLevel),
			fmt.Sprintf("%.4f", cell.Conductance),
			fmt.Sprintf("%.4f", cell.ProgramV),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

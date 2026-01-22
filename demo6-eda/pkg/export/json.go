// pkg/export/json.go
package export

import (
	"encoding/json"
	"os"

	"demo6-eda/pkg/compiler"
)

// ExportJSON writes the crossbar mapping to a JSON file
func ExportJSON(mapping *compiler.CrossbarMapping, path string) error {
	data, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

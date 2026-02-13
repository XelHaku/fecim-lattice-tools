package export

import (
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

func TestGeneratedEDAFilesContainSignoffWarning(t *testing.T) {
	cellCfg := config.CellConfig{Name: "fecim_bitcell", Width: 0.46, Height: 2.72, CellType: "passive"}
	mapping := getTestMapping()

	outputs := map[string]string{
		"liberty": GenerateLiberty(cellCfg),
		"lef":     GenerateLEF(cellCfg),
		"verilog": GenerateVerilogWithDefaults(mapping),
		"spice":   GenerateSPICE(mapping, 1.8),
	}

	for name, content := range outputs {
		if !strings.Contains(content, provenanceWarningText) {
			t.Fatalf("%s missing provenance warning: %q", name, provenanceWarningText)
		}
	}
}

package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModule6ExportFixture_DeterministicMetadataContract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	readme := filepath.Join(repoRoot, "module6-eda", "pkg", "gui", "tabs", "data", "fecim_crossbar_2x2", "README.md")
	verilog := filepath.Join(repoRoot, "module6-eda", "pkg", "gui", "tabs", "data", "fecim_crossbar_2x2", "fecim_crossbar_2x2.v")

	readmeBytes, err := os.ReadFile(readme)
	if err != nil {
		t.Fatalf("read %s: %v", readme, err)
	}
	readmeText := string(readmeBytes)
	if !strings.Contains(readmeText, "Date: reproducible-build") {
		t.Fatalf("%s missing deterministic date marker", readme)
	}

	verilogBytes, err := os.ReadFile(verilog)
	if err != nil {
		t.Fatalf("read %s: %v", verilog, err)
	}
	verilogText := string(verilogBytes)
	if !strings.Contains(verilogText, "// Date: reproducible-build") {
		t.Fatalf("%s missing deterministic date marker", verilog)
	}
}

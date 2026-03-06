package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSenseChainRegressionArtifact_MetadataContract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	p := filepath.Join(repoRoot, "validation", "output", "regression", "module4", "sense_chain_4x4.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	var rec senseChainGolden
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("decode %s: %v", p, err)
	}
	if rec.Version != "v1" {
		t.Fatalf("%s version=%q want v1", p, rec.Version)
	}
	if rec.Scenario != "sense_chain_4x4" {
		t.Fatalf("%s scenario=%q", p, rec.Scenario)
	}
	if rec.Generated != "1970-01-01T00:00:00Z" {
		t.Fatalf("%s generated_utc=%q want deterministic", p, rec.Generated)
	}
	if rec.Parameters["rows"] != float64(4) || rec.Parameters["cols"] != float64(4) {
		t.Fatalf("%s rows/cols mismatch: rows=%v cols=%v", p, rec.Parameters["rows"], rec.Parameters["cols"])
	}
	if len(rec.Results.RowCurrents) != 4 || len(rec.Results.ADCCodes) != 4 {
		t.Fatalf("%s results length mismatch: currents=%d adc=%d", p, len(rec.Results.RowCurrents), len(rec.Results.ADCCodes))
	}
}

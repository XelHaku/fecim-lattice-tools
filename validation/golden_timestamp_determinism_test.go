package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSenseGoldenUpdate_UsesDeterministicGeneratedUTC(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(updateSenseGoldenEnv, "1")
	t.Setenv("FECIM_ARTIFACT_TIMESTAMP_UTC", "")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(cwd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}

	TestSenseChainRegression_4x4(t)

	p := filepath.Join(senseRegressionDir, senseGoldenFile)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	var got senseChainGolden
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("decode golden file: %v", err)
	}
	if got.Generated != "1970-01-01T00:00:00Z" {
		t.Fatalf("expected deterministic generated_utc, got %q", got.Generated)
	}
}

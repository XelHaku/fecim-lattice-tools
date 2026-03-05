package arraysim

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteWriteBoundaryArtifact_ForcesDeterministicGeneratedUnix(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("FECIM_M4_REGRESSION_JSON_DIR", tmp)

	art := writeBoundaryArtifact{
		Version:       "v1",
		GeneratedUnix: 987654321,
		TestName:      "determinism-test",
		Pass:          true,
	}
	writeWriteBoundaryArtifact(t, art)

	path := filepath.Join(tmp, "write_boundary_integrity.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	var payload map[string]writeBoundaryArtifact
	if err := json.Unmarshal(b, &payload); err != nil {
		t.Fatalf("decode artifact: %v", err)
	}
	got, ok := payload["determinism-test"]
	if !ok {
		t.Fatalf("missing record determinism-test in payload")
	}
	if got.GeneratedUnix != 0 {
		t.Fatalf("expected deterministic generated_unix=0, got %d", got.GeneratedUnix)
	}
}

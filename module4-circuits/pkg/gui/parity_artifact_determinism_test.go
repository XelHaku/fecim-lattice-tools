//go:build legacy_fyne

package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteParityArtifact_ForcesDeterministicGeneratedUnix(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "parity.json")
	t.Setenv("FECIM_PARITY_JSON_PATH", out)

	art := &parityArtifact{
		Version:       "v1",
		Profile:       "pr",
		GeneratedUnix: 123456789,
		Records:       []parityStepArtifact{},
	}
	writeParityArtifact(t, art)

	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read parity artifact: %v", err)
	}
	var got parityArtifact
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("decode parity artifact: %v", err)
	}
	if got.GeneratedUnix != 0 {
		t.Fatalf("expected deterministic generated_unix=0, got %d", got.GeneratedUnix)
	}
}

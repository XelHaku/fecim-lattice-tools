package export

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateReproducibilityPackContainsExpectedFiles(t *testing.T) {
	tmp := t.TempDir()
	artifact := filepath.Join(tmp, "artifact.txt")
	if err := os.WriteFile(artifact, []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(tmp, "pack")
	_, err := CreateReproducibilityPack(outDir, ReproducibilityPackInput{
		ConfigYAML:      []byte("seed: 42\n"),
		RandomSeeds:     map[string]int64{"global": 42},
		GitCommitHash:   "abc123",
		TestResults:     "PASS",
		GeneratedAssets: []string{artifact},
	})
	if err != nil {
		t.Fatalf("CreateReproducibilityPack failed: %v", err)
	}

	if err := ValidateReproducibilityPack(outDir); err != nil {
		t.Fatalf("ValidateReproducibilityPack failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "artifacts", "artifact.txt")); err != nil {
		t.Fatalf("missing artifact copy: %v", err)
	}
}

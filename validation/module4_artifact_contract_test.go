package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestModule4RegressionArtifacts_DeterministicMetadataContract(t *testing.T) {
	type parityArtifact struct {
		Version       string `json:"version"`
		Profile       string `json:"profile"`
		GeneratedUnix int64  `json:"generated_unix"`
	}
	type writeBoundaryRecord struct {
		Version       string `json:"version"`
		GeneratedUnix int64  `json:"generated_unix"`
		TestName      string `json:"test_name"`
		Pass          bool   `json:"pass"`
	}

	repoRoot := filepath.Clean(filepath.Join(".."))

	t.Run("gui_vs_headless_parity", func(t *testing.T) {
		p := filepath.Join(repoRoot, "module4-circuits", "pkg", "gui", "output", "regression", "module4", "gui_vs_headless_parity.json")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var art parityArtifact
		if err := json.Unmarshal(b, &art); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if art.Version != "v1" {
			t.Fatalf("%s version=%q want v1", p, art.Version)
		}
		if art.Profile == "" {
			t.Fatalf("%s profile is empty", p)
		}
		if art.GeneratedUnix != 0 {
			t.Fatalf("%s generated_unix=%d want 0", p, art.GeneratedUnix)
		}
	})

	t.Run("write_boundary_integrity", func(t *testing.T) {
		p := filepath.Join(repoRoot, "module4-circuits", "pkg", "arraysim", "output", "regression", "module4", "write_boundary_integrity.json")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		payload := map[string]writeBoundaryRecord{}
		if err := json.Unmarshal(b, &payload); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if len(payload) == 0 {
			t.Fatalf("%s has no records", p)
		}
		for name, rec := range payload {
			if rec.Version != "v1" {
				t.Fatalf("%s[%s] version=%q want v1", p, name, rec.Version)
			}
			if rec.GeneratedUnix != 0 {
				t.Fatalf("%s[%s] generated_unix=%d want 0", p, name, rec.GeneratedUnix)
			}
			if rec.TestName == "" {
				t.Fatalf("%s[%s] test_name empty", p, name)
			}
		}
	})
}

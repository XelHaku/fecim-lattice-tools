package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type literaturePELoopArtifact struct {
	SchemaVersion string `json:"schema_version"`
	TimestampUTC  string `json:"timestamp_utc"`
	Commit        string `json:"commit"`
	Gate          string `json:"gate"`
	TestID        string `json:"test_id"`
	Verdict       string `json:"verdict"`
	GeneratedAt   string `json:"generated_at"`
	Dataset       string `json:"dataset"`
	DOI           string `json:"doi"`
	Pass          bool   `json:"pass"`
}

func TestLiteraturePELoopArtifactFamily_Contract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	glob := filepath.Join(repoRoot, "output", "validation", "literature", "module1_pe_loop_*.json")
	paths, err := filepath.Glob(glob)
	if err != nil {
		t.Fatalf("glob %s: %v", glob, err)
	}
	if len(paths) != 9 {
		t.Fatalf("expected exactly 9 module1_pe_loop artifacts, got %d", len(paths))
	}

	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec literaturePELoopArtifact
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec.SchemaVersion != "v1" || rec.TimestampUTC != "1970-01-01T00:00:00Z" || rec.Commit != "reproducible-build" {
			t.Fatalf("%s envelope mismatch schema=%q ts=%q commit=%q", p, rec.SchemaVersion, rec.TimestampUTC, rec.Commit)
		}
		if rec.GeneratedAt != "1970-01-01T00:00:00Z" {
			t.Fatalf("%s generated_at=%q want deterministic", p, rec.GeneratedAt)
		}
		if rec.Gate == "" || rec.TestID == "" || rec.Verdict == "" || rec.Dataset == "" {
			t.Fatalf("%s missing required fields gate/test_id/verdict/dataset", p)
		}
		if rec.DOI == "" || !strings.HasPrefix(strings.ToLower(rec.DOI), "10.") {
			t.Fatalf("%s invalid doi=%q", p, rec.DOI)
		}
		if !rec.Pass {
			t.Fatalf("%s pass=false", p)
		}
	}
}

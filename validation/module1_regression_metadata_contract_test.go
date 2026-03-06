package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type module1RegressionContract struct {
	Suite     string `json:"suite"`
	Material  string `json:"material"`
	Model     string `json:"model"`
	Timestamp string `json:"timestamp"`
	AllPass   bool   `json:"all_pass"`
}

func TestModule1RegressionArtifacts_MetadataContract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	paths := []string{
		filepath.Join(repoRoot, "output", "regression", "module1", "lk_wrd_ispp_regression_default_hzo.json"),
		filepath.Join(repoRoot, "output", "regression", "module1", "preisach_wrd_ispp_regression_default_hzo.json"),
		filepath.Join(repoRoot, "module1-hysteresis", "pkg", "controller", "output", "regression", "module1", "lk_wrd_ispp_regression_default_hzo.json"),
		filepath.Join(repoRoot, "module1-hysteresis", "pkg", "controller", "output", "regression", "module1", "preisach_wrd_ispp_regression_default_hzo.json"),
	}

	for _, p := range paths {
		p := p
		t.Run(filepath.Base(p), func(t *testing.T) {
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("read %s: %v", p, err)
			}
			var rec module1RegressionContract
			if err := json.Unmarshal(b, &rec); err != nil {
				t.Fatalf("decode %s: %v", p, err)
			}
			if rec.Suite != "headless-wrd-ispp-regression" {
				t.Fatalf("%s suite=%q", p, rec.Suite)
			}
			if rec.Model == "" || rec.Material == "" {
				t.Fatalf("%s missing model/material", p)
			}
			if rec.Timestamp != "1970-01-01T00:00:00Z" {
				t.Fatalf("%s timestamp=%q want deterministic", p, rec.Timestamp)
			}
		})
	}
}

package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDeterministicArtifactEnvelope_TrackedOutputs(t *testing.T) {
	type envelope struct {
		SchemaVersion string `json:"schema_version"`
		TimestampUTC  string `json:"timestamp_utc"`
		Commit        string `json:"commit"`
		Gate          string `json:"gate"`
		TestID        string `json:"test_id"`
		Verdict       string `json:"verdict"`
	}

	paths := []string{
		"output/write_stats/write_verify_stats_default_hzo.json",
		"output/write_stats/write_verify_stats_fecim_hzo.json",
		"output/write_stats/write_verify_stats_literature_superlattice.json",
		"output/montecarlo/mc_pe_uncertainty_default_hzo.json",
	}

	for _, rel := range paths {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			p := filepath.Join(rel)
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("read %s: %v", p, err)
			}
			var env envelope
			if err := json.Unmarshal(b, &env); err != nil {
				t.Fatalf("decode %s: %v", p, err)
			}
			if env.SchemaVersion != "v1" {
				t.Fatalf("%s schema_version=%q want v1", p, env.SchemaVersion)
			}
			if env.TimestampUTC != "1970-01-01T00:00:00Z" {
				t.Fatalf("%s timestamp_utc=%q want deterministic default", p, env.TimestampUTC)
			}
			if env.Commit != "reproducible-build" {
				t.Fatalf("%s commit=%q want reproducible-build", p, env.Commit)
			}
			if env.Gate == "" || env.TestID == "" || env.Verdict == "" {
				t.Fatalf("%s missing required envelope fields: gate=%q test_id=%q verdict=%q", p, env.Gate, env.TestID, env.Verdict)
			}
		})
	}
}

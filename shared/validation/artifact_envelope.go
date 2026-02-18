package validation

import (
	"os"
	"os/exec"
	"strings"
	"time"
)

// ArtifactEnvelope is the standard metadata header required on every Module 1
// test artifact. Field names match MODULE1_REQUIRED_KEYS in
// scripts/ci/validate_regression_artifacts.py.
type ArtifactEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	TimestampUTC  string `json:"timestamp_utc"`
	Commit        string `json:"commit"`
	Gate          string `json:"gate"`
	TestID        string `json:"test_id"`
	Verdict       string `json:"verdict"` // "pass" | "fail"
}

// ArtifactUncertainty is the standard uncertainty sub-object required on every
// artifact. The validator checks method, confidence, and sample_size.
type ArtifactUncertainty struct {
	Method     string  `json:"method"`      // e.g. "monte_carlo", "analytical", "none"
	Confidence float64 `json:"confidence"`  // 0.95 for 95% CI
	SampleSize int     `json:"sample_size"` // number of trials / data points
}

// NewEnvelope builds a populated ArtifactEnvelope.
//
// testID is the plan ID, e.g. "RG-PHY-OBS-01".
// gate is the CI lane ("pr", "nightly", "release"); if empty, read from
// FECIM_CI_GATE env var; defaults to "pr".
// pass determines the verdict: true → "pass", false → "fail".
func NewEnvelope(testID, gate string, pass bool) ArtifactEnvelope {
	if gate == "" {
		gate = os.Getenv("FECIM_CI_GATE")
		if gate == "" {
			gate = "pr"
		}
	}
	verdict := "fail"
	if pass {
		verdict = "pass"
	}
	return ArtifactEnvelope{
		SchemaVersion: "v1",
		TimestampUTC:  time.Now().UTC().Format(time.RFC3339),
		Commit:        resolveCommit(),
		Gate:          gate,
		TestID:        testID,
		Verdict:       verdict,
	}
}

// resolveCommit returns the short git commit hash.
// Priority: FECIM_GIT_COMMIT env var → git rev-parse --short HEAD → "unknown".
func resolveCommit() string {
	if c := strings.TrimSpace(os.Getenv("FECIM_GIT_COMMIT")); c != "" {
		return c
	}
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

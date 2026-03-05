package validation

import (
	"os"
	"strings"
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
		TimestampUTC:  artifactTimestampUTC(),
		Commit:        resolveCommit(),
		Gate:          gate,
		TestID:        testID,
		Verdict:       verdict,
	}
}

func artifactTimestampUTC() string {
	if ts := strings.TrimSpace(os.Getenv("FECIM_ARTIFACT_TIMESTAMP_UTC")); ts != "" {
		return ts
	}
	return "1970-01-01T00:00:00Z"
}

// resolveCommit returns a deterministic artifact commit marker unless explicitly overridden.
// Priority: FECIM_GIT_COMMIT env var → "reproducible-build".
func resolveCommit() string {
	if c := strings.TrimSpace(os.Getenv("FECIM_GIT_COMMIT")); c != "" {
		return c
	}
	return "reproducible-build"
}

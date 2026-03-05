package validation

import "testing"

func TestNewEnvelope_DefaultsAreDeterministic(t *testing.T) {
	t.Setenv("FECIM_CI_GATE", "")
	t.Setenv("FECIM_GIT_COMMIT", "")
	t.Setenv("FECIM_ARTIFACT_TIMESTAMP_UTC", "")

	env := NewEnvelope("RG-TEST-DET", "", true)
	if env.SchemaVersion != "v1" {
		t.Fatalf("schema_version mismatch: got %q", env.SchemaVersion)
	}
	if env.TimestampUTC != "1970-01-01T00:00:00Z" {
		t.Fatalf("timestamp_utc not deterministic default: got %q", env.TimestampUTC)
	}
	if env.Commit != "reproducible-build" {
		t.Fatalf("commit not deterministic default: got %q", env.Commit)
	}
	if env.Gate != "pr" {
		t.Fatalf("default gate mismatch: got %q want pr", env.Gate)
	}
	if env.Verdict != "pass" {
		t.Fatalf("verdict mismatch: got %q", env.Verdict)
	}
}

func TestNewEnvelope_EnvOverridesStillWork(t *testing.T) {
	t.Setenv("FECIM_CI_GATE", "nightly")
	t.Setenv("FECIM_GIT_COMMIT", "abc1234")
	t.Setenv("FECIM_ARTIFACT_TIMESTAMP_UTC", "2026-01-01T00:00:00Z")

	env := NewEnvelope("RG-TEST-DET", "", false)
	if env.TimestampUTC != "2026-01-01T00:00:00Z" {
		t.Fatalf("timestamp override mismatch: got %q", env.TimestampUTC)
	}
	if env.Commit != "abc1234" {
		t.Fatalf("commit override mismatch: got %q", env.Commit)
	}
	if env.Gate != "nightly" {
		t.Fatalf("gate override mismatch: got %q", env.Gate)
	}
	if env.Verdict != "fail" {
		t.Fatalf("verdict mismatch: got %q", env.Verdict)
	}
}

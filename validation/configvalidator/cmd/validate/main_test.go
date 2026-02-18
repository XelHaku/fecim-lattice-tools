package main

import (
	"os"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/validation/configvalidator"
)

// TestValidateJSON_InvalidJSON verifies that raw invalid JSON returns a failed
// validation result rather than a Go error (the validator wraps parse failures).
func TestValidateJSON_InvalidJSON(t *testing.T) {
	result := configvalidator.ValidateJSON([]byte(`not json`))
	if result == nil {
		t.Fatal("ValidateJSON returned nil for invalid JSON")
	}
	if result.Valid {
		t.Error("ValidateJSON: expected Valid=false for malformed JSON, got true")
	}
}

// TestValidateJSON_EmptyObject verifies that an empty JSON object returns a
// non-nil ValidationResult. The validator treats unknown types as a warning
// (not an error), so Valid may be true — but at least one warning is expected
// because the type discriminator is absent.
func TestValidateJSON_EmptyObject(t *testing.T) {
	result := configvalidator.ValidateJSON([]byte(`{}`))
	if result == nil {
		t.Fatal("ValidateJSON returned nil for empty object")
	}
	// Unknown type should produce at least one warning about the unknown schema.
	if len(result.Warnings) == 0 && len(result.Errors) == 0 {
		t.Error("ValidateJSON: expected at least one warning or error for empty object, got none")
	}
	t.Logf("ValidateJSON({}): Valid=%v Errors=%d Warnings=%d",
		result.Valid, len(result.Errors), len(result.Warnings))
}

// TestValidateFile_NonExistent verifies that ValidateFile returns an error for
// a path that does not exist (not a silent success).
func TestValidateFile_NonExistent(t *testing.T) {
	_, err := configvalidator.ValidateFile("/nonexistent/path/config.json")
	if err == nil {
		t.Error("ValidateFile: expected error for non-existent path, got nil")
	}
}

// TestValidateFile_ValidCalibration exercises ValidateFile on a real
// calibration JSON fixture from the testdata directory.
func TestValidateFile_ValidCalibration(t *testing.T) {
	// Walk up to repo root to find a usable calibration fixture.
	candidates := []string{
		filepath.Join("..", "..", "..", "testdata", "fecim_hzo.json"),
		filepath.Join("..", "..", "..", "testdata", "calibrations", "fecim_hzo.json"),
	}
	// Try any JSON in the data directory of the validation package.
	repoRoot := filepath.Join("..", "..", "..", "..")
	dataDir := filepath.Join(repoRoot, "data", "calibrations")
	if entries, err := os.ReadDir(dataDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
				candidates = append(candidates, filepath.Join(dataDir, e.Name()))
				break
			}
		}
	}

	var usablePath string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			usablePath = c
			break
		}
	}
	if usablePath == "" {
		t.Skip("no calibration fixture found — skipping integration file test")
	}

	result, err := configvalidator.ValidateFile(usablePath)
	if err != nil {
		t.Fatalf("ValidateFile(%q): unexpected Go error: %v", usablePath, err)
	}
	if result == nil {
		t.Fatalf("ValidateFile(%q): returned nil result", usablePath)
	}
	t.Logf("ValidateFile(%q): Valid=%v Errors=%d Warnings=%d",
		filepath.Base(usablePath), result.Valid, len(result.Errors), len(result.Warnings))
}

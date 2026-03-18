package physics

import (
	"testing"
)

func TestConfigVersion_Constant(t *testing.T) {
	if ConfigVersion != "1.0.0" {
		t.Errorf("ConfigVersion = %q, want %q", ConfigVersion, "1.0.0")
	}
}

func TestConfigVersionMethod(t *testing.T) {
	cfg := &Config{}
	if got := cfg.Version(); got != ConfigVersion {
		t.Errorf("Version() = %q, want %q", got, ConfigVersion)
	}
}

func TestValidateVersion_SameVersion(t *testing.T) {
	cfg := &Config{LoadedVersion: ConfigVersion}
	if err := cfg.ValidateVersion(); err != nil {
		t.Errorf("ValidateVersion with same version should return nil, got: %v", err)
	}
}

func TestValidateVersion_EmptyIsLegacy(t *testing.T) {
	cfg := &Config{LoadedVersion: ""}
	if err := cfg.ValidateVersion(); err != nil {
		t.Errorf("ValidateVersion with empty LoadedVersion (legacy) should return nil, got: %v", err)
	}
}

func TestValidateVersion_MinorDifference(t *testing.T) {
	cfg := &Config{LoadedVersion: "1.1.0"}
	if err := cfg.ValidateVersion(); err != nil {
		t.Errorf("ValidateVersion with minor version difference should return nil, got: %v", err)
	}
}

func TestValidateVersion_PatchDifference(t *testing.T) {
	cfg := &Config{LoadedVersion: "1.0.5"}
	if err := cfg.ValidateVersion(); err != nil {
		t.Errorf("ValidateVersion with patch version difference should return nil, got: %v", err)
	}
}

func TestValidateVersion_MajorDifference(t *testing.T) {
	cfg := &Config{LoadedVersion: "2.0.0"}
	if err := cfg.ValidateVersion(); err == nil {
		t.Error("ValidateVersion with major version mismatch should return error, got nil")
	}
}

func TestValidateVersion_MajorDifference_ErrorMessage(t *testing.T) {
	cfg := &Config{LoadedVersion: "3.5.1"}
	err := cfg.ValidateVersion()
	if err == nil {
		t.Fatal("ValidateVersion with major version mismatch should return error")
	}
	// Verify the error message mentions both versions
	msg := err.Error()
	if !containsSubstring(msg, "3.5.1") {
		t.Errorf("error message should mention loaded version '3.5.1', got: %s", msg)
	}
	if !containsSubstring(msg, ConfigVersion) {
		t.Errorf("error message should mention current version %q, got: %s", ConfigVersion, msg)
	}
}

// containsSubstring is a small helper to avoid importing strings in this file.
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

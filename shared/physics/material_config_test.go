package physics

import (
	"os"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/shared/testutil"
)

func TestLoadMaterialConfig_HfO2(t *testing.T) {
	// Walk up to find the repo root (configs/ sits next to shared/)
	repoRoot := findRepoRoot(t)
	path := filepath.Join(repoRoot, "configs", "materials", "hfo2.yaml")

	cfg, err := LoadMaterialConfig(path)
	if err != nil {
		t.Fatalf("LoadMaterialConfig(%q): %v", path, err)
	}

	if cfg.Name != "HfO2" {
		t.Errorf("Name = %q, want HfO2", cfg.Name)
	}
	if cfg.RemnantPolarization <= 0 {
		t.Errorf("RemnantPolarization = %v, want > 0", cfg.RemnantPolarization)
	}
	if cfg.CoerciveField <= 0 {
		t.Errorf("CoerciveField = %v, want > 0", cfg.CoerciveField)
	}
}

func TestLoadMaterialConfig_BTO(t *testing.T) {
	repoRoot := findRepoRoot(t)
	cfg, err := LoadMaterialConfig(filepath.Join(repoRoot, "configs", "materials", "bto.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "BaTiO3" {
		t.Errorf("Name = %q, want BaTiO3", cfg.Name)
	}
}

func TestLoadMaterialConfig_Missing(t *testing.T) {
	_, err := LoadMaterialConfig("/nonexistent/path/material.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadMaterialConfig_InvalidYAML(t *testing.T) {
	f, err := os.CreateTemp("", "badmaterial*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("name: [unclosed bracket")
	f.Close()

	_, err = LoadMaterialConfig(f.Name())
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestMaterialConfig_ToHZOMaterial(t *testing.T) {
	repoRoot := findRepoRoot(t)
	cfg, err := LoadMaterialConfig(filepath.Join(repoRoot, "configs", "materials", "hfo2.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	mat := cfg.ToHZOMaterial()

	if mat.Name != cfg.Name {
		t.Errorf("Name mismatch: %q vs %q", mat.Name, cfg.Name)
	}
	// Pr converted from µC/cm² to C/m²: 15 µC/cm² = 0.15 C/m²
	wantPr := cfg.RemnantPolarization * 1e-2
	if testutil.AbsFloat64(mat.Pr-wantPr) > 1e-12 {
		t.Errorf("Pr = %g, want %g", mat.Pr, wantPr)
	}
	// Ec converted from kV/cm to V/m: 1000 kV/cm = 1e8 V/m
	wantEc := cfg.CoerciveField * 1e5
	if testutil.AbsFloat64(mat.Ec-wantEc) > 1 {
		t.Errorf("Ec = %g, want %g", mat.Ec, wantEc)
	}
	if mat.NumLevels != 30 {
		t.Errorf("NumLevels = %d, want 30", mat.NumLevels)
	}
	if mat.Gmin <= 0 || mat.Gmax <= mat.Gmin {
		t.Errorf("conductance window invalid: Gmin=%g Gmax=%g", mat.Gmin, mat.Gmax)
	}
}

func TestMaterialConfig_DefaultsForZeroFields(t *testing.T) {
	// A minimal config with only required fields
	f, err := os.CreateTemp("", "minmat*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("name: TestMaterial\ncoercive_field_kv_cm: 100\nremnant_polarization_uc_cm2: 10\n")
	f.Close()

	cfg, err := LoadMaterialConfig(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mat := cfg.ToHZOMaterial()

	// Zero num_levels → default 30
	if mat.NumLevels != 30 {
		t.Errorf("NumLevels = %d, want 30 (default)", mat.NumLevels)
	}
	// Zero Gmin/Gmax → sensible defaults applied
	if mat.Gmin <= 0 {
		t.Errorf("Gmin should have a positive default, got %g", mat.Gmin)
	}
}

// findRepoRoot walks upward from the test file's directory until it finds go.mod.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (go.mod not found)")
		}
		dir = parent
	}
}

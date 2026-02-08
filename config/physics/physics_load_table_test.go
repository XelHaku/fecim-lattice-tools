package physics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_TableDriven_ConfigOverrides(t *testing.T) {
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(origWD)
		// Best-effort: restore cached config for other tests in this package.
		_, _ = Reload()
	}()

	tests := []struct {
		name         string
		constantsYML string
		wantBits     float64
	}{
		{
			name:     "embedded-defaults",
			wantBits: 4.91,
		},
		{
			name: "override-constants",
			constantsYML: "constants:\n  bits_per_cell: 3.14\n  boltzmann_ev: 8.617e-5\n  epsilon_0: 8.854e-12\n  room_temperature: 300\n",
			wantBits: 3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wd := t.TempDir()
			if tt.constantsYML != "" {
				cfgDir := filepath.Join(wd, "config")
				if err := os.MkdirAll(cfgDir, 0o755); err != nil {
					t.Fatalf("mkdir config: %v", err)
				}
				if err := os.WriteFile(filepath.Join(cfgDir, "constants.yaml"), []byte(tt.constantsYML), 0o644); err != nil {
					t.Fatalf("write constants.yaml: %v", err)
				}
			}

			if err := os.Chdir(wd); err != nil {
				t.Fatalf("chdir: %v", err)
			}

			cfg, err := Reload()
			if err != nil {
				t.Fatalf("Reload: %v", err)
			}
			// Allow small float formatting differences.
			if cfg.Constants.BitsPerCell < tt.wantBits-1e-9 || cfg.Constants.BitsPerCell > tt.wantBits+1e-9 {
				t.Fatalf("BitsPerCell=%v want %v", cfg.Constants.BitsPerCell, tt.wantBits)
			}
			if cfg.DefaultMaterial() == nil {
				t.Fatalf("expected embedded materials to load (default material missing)")
			}
		})
	}
}

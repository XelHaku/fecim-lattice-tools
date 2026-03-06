package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExternalValidationArtifacts_Contract(t *testing.T) {
	repoRoot := filepath.Clean("..")

	t.Run("badcrossbar_crossval", func(t *testing.T) {
		p := filepath.Join(repoRoot, "output", "validation", "external", "badcrossbar_crossval.json")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec map[string]any
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec["test"] != "badcrossbar_crossval" || rec["tool"] != "badcrossbar" {
			t.Fatalf("%s identity mismatch: test=%v tool=%v", p, rec["test"], rec["tool"])
		}
		if rec["pass"] != true || rec["tier1_pass"] != true || rec["tier2_pass"] != true {
			t.Fatalf("%s expected pass/tier passes true", p)
		}
	})

	t.Run("crosssim_mvm_accuracy", func(t *testing.T) {
		p := filepath.Join(repoRoot, "output", "validation", "external", "crosssim_mvm_accuracy.json")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec struct {
			Test      string  `json:"test"`
			Tool      string  `json:"tool"`
			Pass      bool    `json:"pass"`
			Rows      int     `json:"rows"`
			Cols      int     `json:"cols"`
			MaxRelErr float64 `json:"max_rel_err"`
			Threshold float64 `json:"threshold"`
		}
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec.Test != "crosssim_mvm_accuracy" || rec.Tool != "CrossSim" {
			t.Fatalf("%s identity mismatch: %+v", p, rec)
		}
		if !rec.Pass || rec.Rows <= 0 || rec.Cols <= 0 || rec.MaxRelErr > rec.Threshold {
			t.Fatalf("%s invalid pass/shape/error-threshold relation: %+v", p, rec)
		}
	})

	t.Run("lk_small_signal_analytical", func(t *testing.T) {
		p := filepath.Join(repoRoot, "output", "validation", "external", "lk_small_signal_analytical.json")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec map[string]any
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec["test"] != "lk_small_signal_analytical" || rec["pass"] != true {
			t.Fatalf("%s identity/pass mismatch", p)
		}
		for _, k := range []string{"alpha", "beta", "gamma", "tau_lin_s", "max_err_uC_cm2"} {
			if _, ok := rec[k]; !ok {
				t.Fatalf("%s missing key %s", p, k)
			}
		}
	})
}

package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLiteratureMiscArtifactFamily_Contract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	base := filepath.Join(repoRoot, "output", "validation", "literature")

	// 1) Merz law artifact
	{
		p := filepath.Join(base, "lk_merz_law.json")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec struct {
			Test    string `json:"test"`
			DOI     string `json:"doi"`
			NPoints int    `json:"n_points"`
			PassR2  bool   `json:"pass_r2"`
		}
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec.Test != "lk_merz_law_field_dependence" || rec.DOI == "" || rec.NPoints < 8 || !rec.PassR2 {
			t.Fatalf("%s invalid merz contract: %+v", p, rec)
		}
	}

	// 2) SciPy comparison artifacts
	for _, name := range []string{"lk_scipy_ode_comparison.json", "lk_scipy_subcoercive.json"} {
		p := filepath.Join(base, name)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec struct {
			DOI      string    `json:"doi"`
			Dt       float64   `json:"dt_s"`
			ErrPctPs float64   `json:"err_pct_Ps"`
			GoP      []float64 `json:"go_P"`
		}
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec.DOI == "" || rec.Dt <= 0 || rec.ErrPctPs < 0 || len(rec.GoP) < 10 {
			t.Fatalf("%s invalid scipy contract: dt=%g err=%g len(go_P)=%d", p, rec.Dt, rec.ErrPctPs, len(rec.GoP))
		}
	}

	// 3) Rayleigh law + loop area normalization + temperature + endurance
	for _, name := range []string{"rayleigh_law_validation.json", "loop_area_normalization.json", "temperature_dependent_switching.json", "endurance_weibull.json"} {
		p := filepath.Join(base, name)
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec map[string]any
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec["test"] == nil {
			t.Fatalf("%s missing test field", p)
		}
		if results, ok := rec["results"].([]any); !ok || len(results) == 0 {
			t.Fatalf("%s missing/empty results array", p)
		}
	}

	// 4) Chi-squared fit artifact
	{
		p := filepath.Join(base, "pe_chisquared_fit.json")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec struct {
			Description string `json:"description"`
			SigmaRule   string `json:"sigma_rule"`
			Results     []struct {
				Material string  `json:"material"`
				DOI      string  `json:"doi"`
				Chi2r    float64 `json:"chi2_r"`
				Pass     bool    `json:"pass"`
			} `json:"results"`
		}
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec.Description == "" || rec.SigmaRule == "" || len(rec.Results) < 6 {
			t.Fatalf("%s invalid chi2 contract", p)
		}
		for i, r := range rec.Results {
			if r.Material == "" || r.DOI == "" || r.Chi2r <= 0 {
				t.Fatalf("%s result[%d] invalid: %+v", p, i, r)
			}
		}
	}
}

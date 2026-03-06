package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type tempSwitchPoint struct {
	TK     float64 `json:"T_K"`
	TauSim float64 `json:"tau_simulated_s"`
}

type tempSwitchArtifact struct {
	Test    string            `json:"test"`
	DOI     string            `json:"doi"`
	ArrOK   bool              `json:"arrhenius_valid"`
	Results []tempSwitchPoint `json:"results"`
}

type endurancePoint struct {
	NCycles   float64 `json:"N_cycles"`
	PrRemain  float64 `json:"Pr_remaining_frac"`
	DegradePc float64 `json:"degradation_pct"`
}

type enduranceArtifact struct {
	Test    string           `json:"test"`
	DOI     string           `json:"doi"`
	Results []endurancePoint `json:"results"`
}

func TestLiteratureTemperatureAndEndurance_Contract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	base := filepath.Join(repoRoot, "output", "validation", "literature")

	// Temperature-dependent switching contract
	{
		p := filepath.Join(base, "temperature_dependent_switching.json")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec tempSwitchArtifact
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec.Test != "temperature_dependent_switching" || rec.DOI == "" || !rec.ArrOK {
			t.Fatalf("%s invalid header contract: %+v", p, rec)
		}
		if len(rec.Results) < 5 {
			t.Fatalf("%s expected >=5 temperature points, got %d", p, len(rec.Results))
		}
		prevT := 0.0
		for i, r := range rec.Results {
			if r.TK <= 0 || r.TauSim <= 0 {
				t.Fatalf("%s[%d] invalid T/tau: T=%g tau=%g", p, i, r.TK, r.TauSim)
			}
			if i > 0 && r.TK <= prevT {
				t.Fatalf("%s[%d] non-increasing temperature: prev=%g cur=%g", p, i, prevT, r.TK)
			}
			prevT = r.TK
		}
	}

	// Endurance Weibull contract
	{
		p := filepath.Join(base, "endurance_weibull.json")
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec enduranceArtifact
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec.Test != "endurance_weibull" || rec.DOI == "" {
			t.Fatalf("%s invalid header contract: %+v", p, rec)
		}
		if len(rec.Results) < 6 {
			t.Fatalf("%s expected >=6 endurance points, got %d", p, len(rec.Results))
		}
		prevN := 0.0
		prevPr := 2.0
		for i, r := range rec.Results {
			if r.NCycles <= 0 || r.PrRemain <= 0 || r.PrRemain > 1 || r.DegradePc < 0 {
				t.Fatalf("%s[%d] invalid endurance metrics: %+v", p, i, r)
			}
			if i > 0 {
				if r.NCycles <= prevN {
					t.Fatalf("%s[%d] non-increasing cycles: prev=%g cur=%g", p, i, prevN, r.NCycles)
				}
				if r.PrRemain > prevPr {
					t.Fatalf("%s[%d] Pr_remaining increased unexpectedly: prev=%g cur=%g", p, i, prevPr, r.PrRemain)
				}
			}
			prevN, prevPr = r.NCycles, r.PrRemain
		}
	}
}

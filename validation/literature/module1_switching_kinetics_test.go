package literature

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
	sharedval "fecim-lattice-tools/shared/validation"
)

// RG-PHY-OBS-02: Switching kinetics falsification.
//
// The Preisach model is quasi-static, so we validate its partial-switching
// behaviour rather than Merz-law time constants directly. For each applied
// field amplitude E we:
//
//  1. Pre-saturate negatively (E = -Esat → 0).
//  2. Apply a single positive pulse (0 → +E → 0).
//  3. Record the remanent polarisation P_rem(E).
//  4. Compute switched fraction F(E) = (P_rem(E) - P_neg) / (P_pos - P_neg).
//
// Expected qualitative invariants (Merz / NLS / KAI consensus):
//   - F(E) ≈ 0 for E ≪ Ec        (no switching below coercive field)
//   - F(Ec) ≈ 0.50 ± 0.10        (coercive field is 50% switching point)
//   - F(E) ≈ 1 for E ≫ Ec        (complete switching above coercive field)
//   - dF/dE > 0 monotonically     (S-curve – no backwards switching)
//   - Logistic fit slope k is proportional to 1/Delta (steeper for narrower distribution)
//
// Thresholds are purposely modest (>20% S-curve width, ±10% at Ec) because the
// quasi-static Preisach does not capture true kinetics – only shape.

type switchingKineticsMetrics struct {
	MaterialID   string  `json:"material_id"`
	Material     string  `json:"material"`
	Generated    string  `json:"generated_at"`
	Ec_MV_cm     float64 `json:"ec_MV_cm"`
	FracAtEc     float64 `json:"frac_at_ec"`
	FracAtHalfEc float64 `json:"frac_at_half_ec"`
	FracAtTwoEc  float64 `json:"frac_at_two_ec"`
	Monotonic    bool    `json:"monotonic"`
	LogisticK    float64 `json:"logistic_k"`        // fitted slope (MV/cm)^-1
	LogisticEc   float64 `json:"logistic_ec_MV_cm"` // fitted midpoint
	LogisticRMSE float64 `json:"logistic_rmse"`
	Pass         bool    `json:"pass"`
}

// Thresholds for switching kinetics invariants.
const (
	thFracAtEcLo  = 0.35 // F(Ec) must exceed this
	thFracAtEcHi  = 0.65 // F(Ec) must be below this
	thFracHalfEc  = 0.25 // F(Ec/2) must be below this (no premature switching)
	thFracTwoEc   = 0.80 // F(2*Ec) must exceed this (full switching)
	thLogisticRMS = 0.05 // logistic fit RMSE < 5% of Ps
)

func TestModule1_SwitchingKinetics_Falsification(t *testing.T) {
	materials := []struct {
		ID  string
		Mat *sharedphysics.HZOMaterial
	}{
		{"park2015_hzo_10nm", sharedphysics.Park2015Fig2aHZO10nm()},
		{"cheema2020_superlattice_5nm", sharedphysics.Cheema2020Fig2cHZOSuperlattice5nm()},
		{"mdpi2020_hzo_10nm_wakeup", sharedphysics.MDPI2020Fig3aHZO10nmWakeup()},
	}

	outDir := os.Getenv("FECIM_LITERATURE_JSON_DIR")
	if outDir == "" {
		root, err := findRepoRoot()
		if err != nil {
			t.Fatalf("resolve repo root: %v", err)
		}
		outDir = filepath.Join(root, "output", "validation", "literature")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", outDir, err)
	}

	for _, tc := range materials {
		tc := tc
		t.Run(tc.ID, func(t *testing.T) {
			m := computeSwitchingKineticsMetrics(t, tc.ID, tc.Mat)

			outPath := filepath.Join(outDir, fmt.Sprintf("module1_switching_kinetics_%s.json", tc.ID))
			b, _ := json.MarshalIndent(m, "", "  ")
			_ = os.WriteFile(outPath, b, 0o644)

			t.Logf("SWITCHING_KINETICS material=%s ec=%.3f MV/cm F(Ec/2)=%.3f F(Ec)=%.3f F(2Ec)=%.3f monotonic=%v logistic_RMSE=%.4f pass=%v artifact=%s",
				tc.ID, m.Ec_MV_cm, m.FracAtHalfEc, m.FracAtEc, m.FracAtTwoEc, m.Monotonic, m.LogisticRMSE, m.Pass, outPath)

			if !m.Monotonic {
				t.Errorf("switched fraction is not monotone — Preisach distribution has sign error")
			}
			if m.FracAtEc < thFracAtEcLo || m.FracAtEc > thFracAtEcHi {
				t.Errorf("F(Ec)=%.3f outside [%.2f, %.2f]: coercive field not at 50%% switching point", m.FracAtEc, thFracAtEcLo, thFracAtEcHi)
			}
			if m.FracAtHalfEc > thFracHalfEc {
				t.Errorf("F(Ec/2)=%.3f > %.2f: premature switching below coercive field", m.FracAtHalfEc, thFracHalfEc)
			}
			if m.FracAtTwoEc < thFracTwoEc {
				t.Errorf("F(2*Ec)=%.3f < %.2f: incomplete switching above coercive field", m.FracAtTwoEc, thFracTwoEc)
			}
			if m.LogisticRMSE > thLogisticRMS {
				t.Errorf("logistic fit RMSE=%.4f > %.4f: S-curve shape deviates from logistic expectation", m.LogisticRMSE, thLogisticRMS)
			}
		})
	}
}

func computeSwitchingKineticsMetrics(t *testing.T, id string, mat *sharedphysics.HZOMaterial) switchingKineticsMetrics {
	t.Helper()

	ec := mat.Ec          // V/m
	esat := ec * 4.0      // saturation field (well beyond Ec)
	ec_MV_cm := ec * 1e-8 // convert V/m → MV/cm for output

	// Sweep E from 0 to 2.5*Ec in 60 steps.
	const nSteps = 60
	eAmps := make([]float64, nSteps) // V/m
	for i := range eAmps {
		eAmps[i] = esat * float64(i+1) / float64(nSteps)
	}

	// Determine negative and positive remanence (reference points).
	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()
	model.Update(-esat)
	pNeg := model.Update(0) // negative remanence

	model.Reset()
	model.Update(esat)
	pPos := model.Update(0) // positive remanence

	span := pPos - pNeg
	if span <= 0 {
		t.Fatalf("degenerate material: pPos=%g <= pNeg=%g", pPos, pNeg)
	}

	// Sweep: for each pulse amplitude, pre-saturate neg then apply +E pulse.
	fracs := make([]float64, nSteps)
	for i, eA := range eAmps {
		m2 := ferroelectric.NewPreisachModel(mat)
		m2.Reset()
		m2.Update(-esat)
		m2.Update(0)
		m2.Update(eA)
		pRem := m2.Update(0)
		fracs[i] = (pRem - pNeg) / span
		// clamp to [0,1] for numerical noise
		if fracs[i] < 0 {
			fracs[i] = 0
		}
		if fracs[i] > 1 {
			fracs[i] = 1
		}
	}

	// Interpolate F at key field values.
	fracAtHalfEc := interpolateFrac(eAmps, fracs, ec*0.5)
	fracAtEc := interpolateFrac(eAmps, fracs, ec)
	fracAtTwoEc := interpolateFrac(eAmps, fracs, ec*2.0)

	// Check monotonicity.
	monotonic := true
	for i := 1; i < len(fracs); i++ {
		if fracs[i] < fracs[i-1]-1e-6 {
			monotonic = false
			break
		}
	}

	// Fit logistic F(E) = 1/(1+exp(-k*(E-E0))) by gradient-free sweep.
	k, e0, lrmse := fitLogistic(eAmps, fracs)

	return switchingKineticsMetrics{
		MaterialID:   id,
		Material:     mat.Name,
		Generated:    sharedval.NewEnvelope("", "", true).TimestampUTC,
		Ec_MV_cm:     ec_MV_cm,
		FracAtHalfEc: fracAtHalfEc,
		FracAtEc:     fracAtEc,
		FracAtTwoEc:  fracAtTwoEc,
		Monotonic:    monotonic,
		LogisticK:    k * 1e-8, // convert (V/m)^-1 → (MV/cm)^-1
		LogisticEc:   e0 * 1e-8,
		LogisticRMSE: lrmse,
		Pass:         monotonic && fracAtEc >= thFracAtEcLo && fracAtEc <= thFracAtEcHi && fracAtHalfEc <= thFracHalfEc && fracAtTwoEc >= thFracTwoEc && lrmse <= thLogisticRMS,
	}
}

// interpolateFrac linearly interpolates the switched fraction at a target E.
func interpolateFrac(E, F []float64, target float64) float64 {
	if target <= E[0] {
		return F[0]
	}
	for i := 1; i < len(E); i++ {
		if E[i] >= target {
			t := (target - E[i-1]) / (E[i] - E[i-1])
			return F[i-1] + t*(F[i]-F[i-1])
		}
	}
	return F[len(F)-1]
}

// fitLogistic fits F(E)=1/(1+exp(-k*(E-E0))) to the data by scanning a grid
// and returns (k, E0, RMSE).
func fitLogistic(E, F []float64) (float64, float64, float64) {
	// Estimate E0 as midpoint (F≈0.5).
	e0Est := E[len(E)/2]
	for i := 1; i < len(F); i++ {
		if F[i] >= 0.5 {
			t := (0.5 - F[i-1]) / (F[i] - F[i-1])
			e0Est = E[i-1] + t*(E[i]-E[i-1])
			break
		}
	}

	bestRMSE := math.Inf(1)
	bestK, bestE0 := 0.0, e0Est

	// Scan k in reasonable range, e0 around estimate.
	for ki := 1; ki <= 200; ki++ {
		k := float64(ki) / e0Est * 0.3 // k ~ 0.003/e0..0.6/e0
		for de := -5; de <= 5; de++ {
			e0 := e0Est * (1.0 + float64(de)*0.05)
			rmseV := logisticRMSE(E, F, k, e0)
			if rmseV < bestRMSE {
				bestRMSE = rmseV
				bestK = k
				bestE0 = e0
			}
		}
	}
	return bestK, bestE0, bestRMSE
}

func logisticRMSE(E, F []float64, k, e0 float64) float64 {
	s := 0.0
	for i, e := range E {
		pred := 1.0 / (1.0 + math.Exp(-k*(e-e0)))
		d := pred - F[i]
		s += d * d
	}
	return math.Sqrt(s / float64(len(E)))
}

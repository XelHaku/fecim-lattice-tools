package literature

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// RG-PHY-OBS-03: Minor loops / FORC falsification.
//
// FORC (First-Order Reversal Curve) measures the mixed second derivative of
// polarisation with respect to reversal and measurement field:
//
//	ρ(Ha, Hb) = -½ ∂²P(Ha,Hb) / (∂Ha ∂Hb)
//
// where Ha is the reversal field (where the sweep turns around) and
// Hb is the measurement field (Hb ≥ Ha).
//
// In rotated "coercivity–interaction" coordinates:
//
//	Hc = (Hb - Ha) / 2   (coercivity axis, Hc ≥ 0)
//	Hu = (Ha + Hb) / 2   (interaction / bias axis)
//
// Qualitative invariants from the Preisach / NLS consensus:
//   - ρ(Ha,Hb) ≥ 0 everywhere (by construction for positive Preisach density)
//   - ρ integrates to approximately Ps (total irreversible polarisation)
//   - Peak of ρ occurs near Hc ≈ Ec (coercive field), Hu ≈ 0 (no bias)
//   - FORC is symmetric in Hu (no pinning bias for an ideal HZO material)
//
// We use central-difference finite differences on a 25×25 grid of FORC curves.

type forcMetrics struct {
	MaterialID      string  `json:"material_id"`
	Material        string  `json:"material"`
	Generated       string  `json:"generated_at"`
	Ec_MV_cm        float64 `json:"ec_MV_cm"`
	Ps_uC_cm2       float64 `json:"ps_uC_cm2"`
	NegativeFrac    float64 `json:"negative_fraction"` // fraction of ρ < 0 (must be ~0)
	PeakHc_MV_cm    float64 `json:"peak_hc_MV_cm"`     // peak of ρ in Hc (expect ≈ Ec)
	PeakHu_MV_cm    float64 `json:"peak_hu_MV_cm"`     // peak of ρ in Hu (expect ≈ 0)
	Integral_uC_cm2 float64 `json:"integral_uC_cm2"`   // ∫∫ρ dHc dHu (expect ≈ Ps)
	IntegralErrPct  float64 `json:"integral_err_pct"`
	HcErrPct        float64 `json:"hc_err_pct"`
	HuAbsMV_cm      float64 `json:"hu_abs_MV_cm"` // |peak Hu| (expect ≈ 0)
	Pass            bool    `json:"pass"`
}

// Thresholds for FORC invariants.
const (
	thFORCNegFrac   = 0.06 // <6% of FORC cells may be negative (numerical noise, 25×25 grid)
	thFORCHcErrPct  = 20.0 // peak Hc within 20% of Ec
	thFORCHuMV_cm   = 0.20 // |peak Hu| < 0.2 MV/cm (unbiased material)
	thFORCIntErrPct = 30.0 // integral within 30% of Ps
)

func TestModule1_FORC_Falsification(t *testing.T) {
	materials := []struct {
		ID  string
		Mat *sharedphysics.HZOMaterial
	}{
		{"park2015_hzo_10nm", sharedphysics.Park2015Fig2aHZO10nm()},
		{"cheema2020_superlattice_5nm", sharedphysics.Cheema2020Fig2cHZOSuperlattice5nm()},
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
			m := computeFORCMetrics(t, tc.ID, tc.Mat)

			outPath := filepath.Join(outDir, fmt.Sprintf("module1_forc_%s.json", tc.ID))
			b, _ := json.MarshalIndent(m, "", "  ")
			_ = os.WriteFile(outPath, b, 0o644)

			t.Logf("FORC material=%s Ec=%.3f MV/cm Ps=%.1f μC/cm² peak_Hc=%.3f(err=%.1f%%) peak_Hu=%.3f neg_frac=%.4f integral=%.1f μC/cm²(err=%.1f%%) pass=%v artifact=%s",
				tc.ID, m.Ec_MV_cm, m.Ps_uC_cm2, m.PeakHc_MV_cm, m.HcErrPct, m.PeakHu_MV_cm, m.NegativeFrac, m.Integral_uC_cm2, m.IntegralErrPct, m.Pass, outPath)

			if m.NegativeFrac > thFORCNegFrac {
				t.Errorf("FORC negative fraction %.4f > %.4f: Preisach density has sign issue", m.NegativeFrac, thFORCNegFrac)
			}
			if m.HcErrPct > thFORCHcErrPct {
				t.Errorf("FORC peak Hc=%.3f MV/cm, Ec=%.3f MV/cm, err=%.1f%% > %.1f%%", m.PeakHc_MV_cm, m.Ec_MV_cm, m.HcErrPct, thFORCHcErrPct)
			}
			if m.HuAbsMV_cm > thFORCHuMV_cm {
				t.Errorf("FORC peak |Hu|=%.3f MV/cm > %.3f MV/cm: unexpected interaction field bias", m.HuAbsMV_cm, thFORCHuMV_cm)
			}
			if m.IntegralErrPct > thFORCIntErrPct {
				t.Errorf("FORC integral=%.2f μC/cm², Ps=%.2f μC/cm², err=%.1f%% > %.1f%%", m.Integral_uC_cm2, m.Ps_uC_cm2, m.IntegralErrPct, thFORCIntErrPct)
			}
		})
	}
}

func computeFORCMetrics(t *testing.T, id string, mat *sharedphysics.HZOMaterial) forcMetrics {
	t.Helper()

	ec := mat.Ec       // V/m
	esat := ec * 3.5   // saturation field
	ps := mat.Ps * 1e2 // C/m² → μC/cm²

	// FORC grid: NHa reversal fields from -esat to +esat, NHb measurement steps.
	const NHa = 25
	const NHb = 25

	ha := make([]float64, NHa)
	for i := range ha {
		ha[i] = -esat + float64(i)*2*esat/float64(NHa-1)
	}

	hb := make([]float64, NHb)
	for i := range hb {
		hb[i] = -esat + float64(i)*2*esat/float64(NHb-1)
	}

	// P[i][j] = P measured at (Ha=ha[i], Hb=hb[j]).
	// Only defined when hb[j] >= ha[i].
	P := make([][]float64, NHa)
	for i := range P {
		P[i] = make([]float64, NHb)
	}

	// FORC measurement protocol: for each reversal field Ha, maintain a single
	// Preisach model instance and sweep Hb continuously from Ha to +esat.
	// Path-dependent models require continuous sweeps, not fresh instances per point.
	for i, haVal := range ha {
		// Mark all points below Ha as undefined.
		for j, hbVal := range hb {
			if hbVal < haVal {
				P[i][j] = math.NaN()
			}
		}

		m := ferroelectric.NewPreisachModel(mat)
		m.Reset()
		// 1. Saturate positively (initial saturated state).
		m.Update(esat)
		m.Update(0)
		// 2. Drive continuously to reversal field Ha.
		// Use intermediate steps for path accuracy.
		steps := 10
		for s := 1; s <= steps; s++ {
			m.Update(esat + float64(s)*(haVal-esat)/float64(steps))
		}
		// 3. Sweep Hb continuously from Ha upward, recording P at each grid point.
		prevHb := haVal
		for j, hbVal := range hb {
			if hbVal < haVal {
				continue
			}
			// Step from prevHb to hbVal in small sub-steps.
			subSteps := 5
			for s := 1; s <= subSteps; s++ {
				m.Update(prevHb + float64(s)*(hbVal-prevHb)/float64(subSteps))
			}
			P[i][j] = m.Update(hbVal) * 1e2 // C/m² → μC/cm²
			prevHb = hbVal
		}
	}

	// Compute FORC density using central differences where possible.
	// ρ(i,j) = -0.5 * d²P / (dHa * dHb)
	dHa := ha[1] - ha[0] // V/m
	dHb := hb[1] - hb[0]

	type forcPoint struct {
		hc, hu, rho float64 // MV/cm
	}
	var points []forcPoint

	for i := 1; i < NHa-1; i++ {
		for j := 1; j < NHb-1; j++ {
			// Need P[i-1][j], P[i+1][j], P[i][j-1], P[i][j+1], P[i-1][j-1], P[i+1][j+1]
			// Use mixed derivative: d²P/(dHa dHb) ≈ (P[i+1,j+1]-P[i+1,j-1]-P[i-1,j+1]+P[i-1,j-1])/(4 dHa dHb)
			p11 := P[i+1][j+1]
			p1m := P[i+1][j-1]
			pm1 := P[i-1][j+1]
			pmm := P[i-1][j-1]
			if math.IsNaN(p11) || math.IsNaN(p1m) || math.IsNaN(pm1) || math.IsNaN(pmm) {
				continue
			}
			d2P := (p11 - p1m - pm1 + pmm) / (4 * dHa * dHb * 1e-16) // 1e-16 converts (V/m)² → (MV/cm)²
			rho := -0.5 * d2P                                        // μC/cm² / (MV/cm)²

			hc := (hb[j] - ha[i]) * 0.5e-8 // MV/cm
			hu := (ha[i] + hb[j]) * 0.5e-8

			if hc < 0 {
				continue
			}
			points = append(points, forcPoint{hc, hu, rho})
		}
	}

	if len(points) == 0 {
		t.Fatal("no valid FORC grid points computed")
	}

	// Find peak, count negatives, integrate.
	negCount := 0
	totalCount := 0
	peakRho := -math.Inf(1)
	peakHc, peakHu := 0.0, 0.0
	integral := 0.0
	// Cell area in (Ha,Hb) space converted to (MV/cm)².
	// dHa and dHb are in V/m; 1 MV/cm = 1e8 V/m, so 1 (V/m)² = 1e-16 (MV/cm)².
	// Integrating ρ [μC/cm²/(MV/cm)²] over dHa·dHb [(MV/cm)²] gives [μC/cm²].
	// Note: dHa·dHb = 4·dHc·dHu, so using the (Ha,Hb) cell area avoids Jacobian error.
	cellArea := dHa * dHb * 1e-16 // (MV/cm)²

	for _, p := range points {
		totalCount++
		if p.rho < 0 {
			negCount++
		}
		if p.rho > peakRho {
			peakRho = p.rho
			peakHc = p.hc
			peakHu = p.hu
		}
		if p.rho > 0 {
			integral += p.rho * cellArea
		}
	}

	negFrac := float64(negCount) / float64(totalCount)
	hcErrPct := pctErr(peakHc, ec*1e-8)
	intErrPct := pctErr(integral, ps)

	ec_MV_cm := ec * 1e-8

	return forcMetrics{
		MaterialID:      id,
		Material:        mat.Name,
		Generated:       time.Now().UTC().Format(time.RFC3339),
		Ec_MV_cm:        ec_MV_cm,
		Ps_uC_cm2:       ps,
		NegativeFrac:    negFrac,
		PeakHc_MV_cm:    peakHc,
		PeakHu_MV_cm:    peakHu,
		Integral_uC_cm2: integral,
		IntegralErrPct:  intErrPct,
		HcErrPct:        hcErrPct,
		HuAbsMV_cm:      math.Abs(peakHu),
		Pass: negFrac <= thFORCNegFrac &&
			hcErrPct <= thFORCHcErrPct &&
			math.Abs(peakHu) <= thFORCHuMV_cm &&
			intErrPct <= thFORCIntErrPct,
	}
}

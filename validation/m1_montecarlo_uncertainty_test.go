package validation

// RG-VAL-M1-04: Monte Carlo uncertainty on P–E + ISPP.
//
// 1. P-E Monte Carlo: perturb Ps/Ec/Pr by ±σ (σ=5%) for N=200 trials,
//    compute 5th/95th-percentile CI bands for simulated Pr, Ec, and loop area,
//    assert CI width < 60% of mean and all values physically valid.
// 2. Seed determinism: run twice with the same seed; assert all trial
//    results are bit-for-bit identical.
// 3. ISPP convergence rate: across N=50 perturbed materials, assert ≥90%
//    converge to the target level.

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
	sharedval "fecim-lattice-tools/shared/validation"
)

const (
	mcNTrialsPE   = 200
	mcNTrialsISPP = 50
	mcSigmaFrac   = 0.05 // 5% RSD on Ps, Ec, Pr
	mcSeed        = 42
	mcNumLevels   = 30
	mcISPPTarget  = 15
)

// mcLoopMetrics holds extracted P-E loop metrics from a single trial.
type mcLoopMetrics struct {
	Pr   float64 // Remanent polarization (C/m²) at E=0 on descending branch
	Ec   float64 // Coercive field (V/m) at P=0 on ascending branch
	Area float64 // Hysteresis loop area (V·C/m³ = J/m³)
}

// mcPercentileStats holds distribution summary for a single metric.
type mcPercentileStats struct {
	Mean    float64 `json:"mean"`
	Std     float64 `json:"std"`
	P5      float64 `json:"p5"`
	P95     float64 `json:"p95"`
	CIWidth float64 `json:"ci_width_frac"` // (P95-P5)/mean
}

// mcMetricsBlock holds the key scalar metrics for machine-readable validation.
type mcMetricsBlock struct {
	CIWidthPr           float64 `json:"ci_width_pr"`
	CIWidthEc           float64 `json:"ci_width_ec"`
	CIWidthArea         float64 `json:"ci_width_area"`
	SeedDeterminismOK   bool    `json:"seed_determinism_ok"`
	ISPPConvergenceFrac float64 `json:"ispp_convergence_frac"`
}

// mcThresholds carries the hard-gate values for this artifact.
type mcThresholds struct {
	CIWidthMax         float64 `json:"ci_width_max"`
	ISPPConvergenceMin float64 `json:"ispp_convergence_min"`
}

// mcReport is the JSON artifact for this test.
type mcReport struct {
	sharedval.ArtifactEnvelope // schema_version, timestamp_utc, commit, gate, test_id, verdict

	MaterialID  string            `json:"material_id"`
	Material    string            `json:"material"`
	Dataset     string            `json:"dataset"`
	NTrialsPE   int               `json:"n_trials_pe"`
	NTrialsISPP int               `json:"n_trials_ispp"`
	Seed        int64             `json:"seed"`
	SigmaFrac   float64           `json:"sigma_frac"`
	Generated   string            `json:"generated_at"` // kept for backwards compat; see timestamp_utc
	PrStats     mcPercentileStats `json:"pr_stats"`
	EcStats     mcPercentileStats `json:"ec_stats"`
	AreaStats   mcPercentileStats `json:"area_stats"`
	DetermOK    bool              `json:"seed_determinism_ok"`

	Metrics     mcMetricsBlock                `json:"metrics"`
	Uncertainty sharedval.ArtifactUncertainty `json:"uncertainty"`
	Thresholds  mcThresholds                  `json:"thresholds"`
}

func TestM1_MonteCarlo_Uncertainty(t *testing.T) {
	mat := ferroelectric.DefaultHZO()

	outDir := os.Getenv("FECIM_LITERATURE_JSON_DIR")
	if outDir == "" {
		outDir = filepath.Join("output", "montecarlo")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", outDir, err)
	}

	t.Run("PE_CIBands", func(t *testing.T) {
		metrics1 := runMCPETrials(mat, mcNTrialsPE, mcSeed)
		metrics2 := runMCPETrials(mat, mcNTrialsPE, mcSeed)

		// Seed determinism: both runs must be identical.
		determOK := true
		for i := range metrics1 {
			if metrics1[i] != metrics2[i] {
				determOK = false
				t.Errorf("seed determinism failed at trial %d: %+v vs %+v", i, metrics1[i], metrics2[i])
			}
		}

		prVals := make([]float64, len(metrics1))
		ecVals := make([]float64, len(metrics1))
		areaVals := make([]float64, len(metrics1))
		for i, m := range metrics1 {
			if m.Pr <= 0 {
				t.Errorf("trial %d: Pr=%e ≤ 0 (non-physical)", i, m.Pr)
			}
			if m.Ec <= 0 {
				t.Errorf("trial %d: Ec=%e ≤ 0 (non-physical)", i, m.Ec)
			}
			if m.Area <= 0 {
				t.Errorf("trial %d: Area=%e ≤ 0 (non-physical)", i, m.Area)
			}
			prVals[i] = m.Pr
			ecVals[i] = m.Ec
			areaVals[i] = m.Area
		}

		prStats := computeMCStats(prVals)
		ecStats := computeMCStats(ecVals)
		areaStats := computeMCStats(areaVals)

		const maxCIWidth = 0.60 // 60% of mean — loose bound, just checking non-explosion

		if prStats.CIWidth > maxCIWidth {
			t.Errorf("Pr CI width=%.3f > %.2f (sigma=%.0f%%)", prStats.CIWidth, maxCIWidth, mcSigmaFrac*100)
		}
		if ecStats.CIWidth > maxCIWidth {
			t.Errorf("Ec CI width=%.3f > %.2f (sigma=%.0f%%)", ecStats.CIWidth, maxCIWidth, mcSigmaFrac*100)
		}
		if areaStats.CIWidth > maxCIWidth {
			t.Errorf("Area CI width=%.3f > %.2f (sigma=%.0f%%)", areaStats.CIWidth, maxCIWidth, mcSigmaFrac*100)
		}

		t.Logf("MC_PE n=%d sigma=%.0f%% Pr: mean=%.3f p5=%.3f p95=%.3f ci=%.2f%%",
			mcNTrialsPE, mcSigmaFrac*100, prStats.Mean*1e6, prStats.P5*1e6, prStats.P95*1e6, prStats.CIWidth*100)
		t.Logf("MC_PE n=%d sigma=%.0f%% Ec: mean=%.3f p5=%.3f p95=%.3f ci=%.2f%%",
			mcNTrialsPE, mcSigmaFrac*100, ecStats.Mean/1e8, ecStats.P5/1e8, ecStats.P95/1e8, ecStats.CIWidth*100)
		t.Logf("MC_PE n=%d sigma=%.0f%% Area: mean=%.3e p5=%.3e p95=%.3e ci=%.2f%%",
			mcNTrialsPE, mcSigmaFrac*100, areaStats.Mean, areaStats.P5, areaStats.P95, areaStats.CIWidth*100)

		pePass := determOK &&
			prStats.CIWidth <= 0.60 &&
			ecStats.CIWidth <= 0.60 &&
			areaStats.CIWidth <= 0.60

		// Write artifact.
		report := mcReport{
			ArtifactEnvelope: sharedval.NewEnvelope("RG-VAL-M1-04", "", pePass),
			MaterialID:       "default_hzo",
			Material:         mat.Name,
			Dataset:          "default_hzo_pe_mc",
			NTrialsPE:        mcNTrialsPE,
			NTrialsISPP:      mcNTrialsISPP,
			Seed:             mcSeed,
			SigmaFrac:        mcSigmaFrac,
			Generated:        sharedval.NewEnvelope("", "", true).TimestampUTC,
			PrStats:          prStats,
			EcStats:          ecStats,
			AreaStats:        areaStats,
			DetermOK:         determOK,
			Metrics: mcMetricsBlock{
				CIWidthPr:         prStats.CIWidth,
				CIWidthEc:         ecStats.CIWidth,
				CIWidthArea:       areaStats.CIWidth,
				SeedDeterminismOK: determOK,
			},
			Uncertainty: sharedval.ArtifactUncertainty{
				Method:     "monte_carlo",
				Confidence: 0.90, // 5th–95th percentile
				SampleSize: mcNTrialsPE,
			},
			Thresholds: mcThresholds{
				CIWidthMax:         0.60,
				ISPPConvergenceMin: 0.90,
			},
		}

		outPath := filepath.Join(outDir, fmt.Sprintf("mc_pe_uncertainty_%s.json", "default_hzo"))
		b, _ := json.MarshalIndent(report, "", "  ")
		if err := os.WriteFile(outPath, b, 0o644); err != nil {
			t.Logf("warn: could not write artifact %s: %v", outPath, err)
		} else {
			t.Logf("artifact: %s", outPath)
		}
	})

	t.Run("ISPP_ConvergenceRate", func(t *testing.T) {
		rng := rand.New(rand.NewSource(mcSeed + 1))
		converged := 0
		for i := 0; i < mcNTrialsISPP; i++ {
			pertMat := perturbMaterial(mat, rng, mcSigmaFrac)
			model := ferroelectric.NewPreisachModel(pertMat)
			model.Reset()

			wc := controller.NewWriteController(mcNumLevels, pertMat.Ec, pertMat.Ec*2.5, nil)
			wc.PulseDuration = 5e-4
			wc.Start(mcISPPTarget, true)

			p := -pertMat.Ps
			currentField := 0.0
			const (
				maxIters = 50000
				dt       = 5e-6
			)
			for j := 0; j < maxIters; j++ {
				curLevel := preisachLevelFromP(p, pertMat.Ps, mcNumLevels)
				targetField, done := wc.Update(dt, currentField, curLevel, 0)
				currentField = targetField
				p = model.Update(currentField)
				if done {
					break
				}
			}
			if wc.State == controller.StateSuccess {
				converged++
			}
		}

		convFrac := float64(converged) / float64(mcNTrialsISPP)
		const minConvFrac = 0.90
		if convFrac < minConvFrac {
			t.Errorf("ISPP convergence rate=%.2f < %.2f (%d/%d converged)",
				convFrac, minConvFrac, converged, mcNTrialsISPP)
		}
		t.Logf("MC_ISPP n=%d sigma=%.0f%% converged=%d/%d (%.1f%%)",
			mcNTrialsISPP, mcSigmaFrac*100, converged, mcNTrialsISPP, convFrac*100)
	})
}

// runMCPETrials runs N perturbed P-E loop trials with a seeded RNG.
func runMCPETrials(baseMat *sharedphysics.HZOMaterial, n int, seed int64) []mcLoopMetrics {
	rng := rand.New(rand.NewSource(seed))
	results := make([]mcLoopMetrics, n)
	for i := 0; i < n; i++ {
		pertMat := perturbMaterial(baseMat, rng, mcSigmaFrac)
		Emax := 2.0 * baseMat.Ec
		model := ferroelectric.NewPreisachModel(pertMat)
		xE, yP := model.GetHysteresisLoop(Emax, 201)
		results[i] = extractLoopMetrics(xE, yP)
	}
	return results
}

// perturbMaterial creates a copy of mat with Ps, Ec, Pr perturbed by ±σ.
func perturbMaterial(mat *sharedphysics.HZOMaterial, rng *rand.Rand, sigmaFrac float64) *sharedphysics.HZOMaterial {
	copy := *mat
	copy.Ps = mat.Ps * (1 + rng.NormFloat64()*sigmaFrac)
	copy.Ec = mat.Ec * (1 + rng.NormFloat64()*sigmaFrac)
	copy.Pr = mat.Pr * (1 + rng.NormFloat64()*sigmaFrac)
	// Keep physical constraints.
	if copy.Ps < mat.Ps*0.5 {
		copy.Ps = mat.Ps * 0.5
	}
	if copy.Ec < mat.Ec*0.5 {
		copy.Ec = mat.Ec * 0.5
	}
	if copy.Pr > copy.Ps*0.99 {
		copy.Pr = copy.Ps * 0.99
	}
	if copy.Pr < mat.Pr*0.1 {
		copy.Pr = mat.Pr * 0.1
	}
	return &copy
}

// extractLoopMetrics extracts Pr, Ec, and loop area from a full P-E loop.
//
// The loop from GetHysteresisLoop(Emax, 201) has:
//   - Ascending (0..402): E from -Emax → +Emax
//   - Descending (403..804): E from +Emax → -Emax
func extractLoopMetrics(xE, yP []float64) mcLoopMetrics {
	const nHalf = 402 // index where descending branch starts

	// Extract Pr: P at E≈0 on descending branch (interpolated).
	pr := 0.0
	for i := nHalf; i < len(xE)-1; i++ {
		if xE[i] >= 0 && xE[i+1] < 0 {
			t := xE[i] / (xE[i] - xE[i+1])
			pr = yP[i] + t*(yP[i+1]-yP[i])
			break
		}
	}

	// Extract Ec: E at P≈0 on ascending branch (positive-going zero crossing, interpolated).
	ec := 0.0
	for i := 0; i < nHalf-1; i++ {
		if yP[i] <= 0 && yP[i+1] > 0 {
			t := -yP[i] / (yP[i+1] - yP[i])
			ec = xE[i] + t*(xE[i+1]-xE[i])
			break
		}
	}

	// Loop area: |∮ P dE| via trapezoidal rule.
	area := 0.0
	for i := 0; i < len(xE)-1; i++ {
		area += (yP[i] + yP[i+1]) * 0.5 * (xE[i+1] - xE[i])
	}
	area = math.Abs(area)

	return mcLoopMetrics{Pr: pr, Ec: ec, Area: area}
}

// computeMCStats computes mean, std, and 5th/95th percentile CI for a slice.
func computeMCStats(vals []float64) mcPercentileStats {
	n := float64(len(vals))
	if n == 0 {
		return mcPercentileStats{}
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)

	mean := 0.0
	for _, v := range vals {
		mean += v
	}
	mean /= n

	variance := 0.0
	for _, v := range vals {
		d := v - mean
		variance += d * d
	}
	variance /= n

	p5 := sorted[int(math.Floor(n*0.05))]
	p95 := sorted[int(math.Min(math.Floor(n*0.95), n-1))]

	ciWidth := 0.0
	if mean != 0 {
		ciWidth = (p95 - p5) / math.Abs(mean)
	}

	return mcPercentileStats{
		Mean:    mean,
		Std:     math.Sqrt(variance),
		P5:      p5,
		P95:     p95,
		CIWidth: ciWidth,
	}
}

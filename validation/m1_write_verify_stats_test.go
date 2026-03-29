package validation

// RG-VAL-M1-03: Export WriteVerifyStats into regression JSON and assert bounds.
//
// Runs the Preisach ISPP controller for selected materials and targets,
// records per-target pulse counts, overshoot counts, stuck counts, and guard
// pulse counts, emits a JSON artifact, and asserts reasonable physics bounds.

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
	sharedval "fecim-lattice-tools/shared/validation"
)

type writeTargetStats struct {
	Target         int     `json:"target"`
	FinalLevel     int     `json:"final_level"`
	Converged      bool    `json:"converged"`
	TotalPulses    int     `json:"total_pulses"`
	OvershootCount int     `json:"overshoot_count"`
	StuckCount     int     `json:"stuck_count"`
	GuardPulses    int     `json:"guard_pulses"`
	RetryCount     int     `json:"retry_count"`
	Fluence        float64 `json:"cumulative_fluence"`
}

// wvsMetricsBlock mirrors key scalar metrics for machine-readable validation.
type wvsMetricsBlock struct {
	AllConverge   bool `json:"all_converge"`
	MaxPulses     int  `json:"max_pulses_observed"`
	MaxOvershoots int  `json:"max_overshoots_observed"`
}

// wvsThresholds carries the hard-gate values for this artifact.
type wvsThresholds struct {
	MaxPulsesPerTarget int `json:"max_pulses_per_target"`
	MaxOvershoots      int `json:"max_overshoots"`
	MaxStuckPerTarget  int `json:"max_stuck_per_target"`
	MaxGuardPerTarget  int `json:"max_guard_per_target"`
}

type writeVerifyStatsReport struct {
	sharedval.ArtifactEnvelope // schema_version, timestamp_utc, commit, gate, test_id, verdict

	MaterialID  string             `json:"material_id"`
	Material    string             `json:"material"`
	Dataset     string             `json:"dataset"`
	NumLevels   int                `json:"num_levels"`
	Generated   string             `json:"generated_at"` // kept for backwards compat; see timestamp_utc
	Targets     []writeTargetStats `json:"targets"`
	AllConverge bool               `json:"all_converge"`

	Metrics     wvsMetricsBlock               `json:"metrics"`
	Uncertainty sharedval.ArtifactUncertainty `json:"uncertainty"`
	Thresholds  wvsThresholds                 `json:"thresholds"`
}

// Bounds for write verify stats.
const (
	maxPulsesPerTarget = 200 // Upper bound on ISPP pulses per target
	maxOvershoots      = 30  // Must not exceed OvershootLimit (which terminates as success)
	maxStuckPerTarget  = 5   // Stuck detection should fire rarely
	maxGuardPerTarget  = 10  // Guard pulses should be rare
)

func TestM1_WriteVerifyStats_Regression(t *testing.T) {
	type matCase struct {
		id  string
		mat *sharedphysics.HZOMaterial
	}

	cases := []matCase{
		{id: "fecim_hzo", mat: ferroelectric.FeCIMMaterial()},
		{id: "literature_superlattice", mat: ferroelectric.LiteratureSuperlattice()},
		{id: "default_hzo", mat: ferroelectric.DefaultHZO()},
	}

	const numLevels = sharedphysics.DefaultLevels
	// Targets: lo=5, mid=15, hi=25 (well-separated; avoids near-saturation edges
	// where the Preisach model requires fields beyond 2.5×Ec to converge).
	targets := []int{5, 15, 25}

	outDir := os.Getenv("FECIM_LITERATURE_JSON_DIR")
	if outDir == "" {
		outDir = filepath.Join("output", "write_stats")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", outDir, err)
	}

	for _, mc := range cases {
		mc := mc
		t.Run(mc.id, func(t *testing.T) {
			allConverge := true
			maxPulsesObs := 0
			maxOvershootsObs := 0
			var targetStats []writeTargetStats

			for _, tgt := range targets {
				stats := runISPPTarget(mc.mat, numLevels, tgt)
				targetStats = append(targetStats, stats)

				if !stats.Converged {
					allConverge = false
					t.Errorf("material=%s target=%d did not converge (final=%d pulses=%d)",
						mc.id, tgt, stats.FinalLevel, stats.TotalPulses)
				}
				if stats.TotalPulses > maxPulsesPerTarget {
					t.Errorf("material=%s target=%d pulses=%d > limit=%d",
						mc.id, tgt, stats.TotalPulses, maxPulsesPerTarget)
				}
				if stats.OvershootCount > maxOvershoots {
					t.Errorf("material=%s target=%d overshoots=%d > limit=%d",
						mc.id, tgt, stats.OvershootCount, maxOvershoots)
				}
				if stats.StuckCount > maxStuckPerTarget {
					t.Errorf("material=%s target=%d stuck=%d > limit=%d",
						mc.id, tgt, stats.StuckCount, maxStuckPerTarget)
				}
				if stats.GuardPulses > maxGuardPerTarget {
					t.Errorf("material=%s target=%d guard_pulses=%d > limit=%d",
						mc.id, tgt, stats.GuardPulses, maxGuardPerTarget)
				}

				t.Logf("WRITE_STATS material=%s target=%d final=%d converged=%v pulses=%d overshoot=%d stuck=%d guard=%d",
					mc.id, tgt, stats.FinalLevel, stats.Converged, stats.TotalPulses,
					stats.OvershootCount, stats.StuckCount, stats.GuardPulses)

				if stats.TotalPulses > maxPulsesObs {
					maxPulsesObs = stats.TotalPulses
				}
				if stats.OvershootCount > maxOvershootsObs {
					maxOvershootsObs = stats.OvershootCount
				}
			}

			report := writeVerifyStatsReport{
				ArtifactEnvelope: sharedval.NewEnvelope("RG-VAL-M1-03", "", allConverge),
				MaterialID:       mc.id,
				Material:         mc.mat.Name,
				Dataset:          fmt.Sprintf("%s_ispp_write_verify", mc.id),
				NumLevels:        numLevels,
				Generated:        sharedval.NewEnvelope("", "", true).TimestampUTC,
				Targets:          targetStats,
				AllConverge:      allConverge,
				Metrics: wvsMetricsBlock{
					AllConverge:   allConverge,
					MaxPulses:     maxPulsesObs,
					MaxOvershoots: maxOvershootsObs,
				},
				Uncertainty: sharedval.ArtifactUncertainty{
					Method:     "none",
					Confidence: 1.0,
					SampleSize: len(targets),
				},
				Thresholds: wvsThresholds{
					MaxPulsesPerTarget: maxPulsesPerTarget,
					MaxOvershoots:      maxOvershoots,
					MaxStuckPerTarget:  maxStuckPerTarget,
					MaxGuardPerTarget:  maxGuardPerTarget,
				},
			}

			outPath := filepath.Join(outDir, fmt.Sprintf("write_verify_stats_%s.json", mc.id))
			b, _ := json.MarshalIndent(report, "", "  ")
			if err := os.WriteFile(outPath, b, 0o644); err != nil {
				t.Logf("warn: could not write artifact %s: %v", outPath, err)
			} else {
				t.Logf("artifact: %s", outPath)
			}
		})
	}
}

// runISPPTarget drives the Preisach ISPP controller to write a single target
// level, starting from negative saturation, and returns per-target stats.
func runISPPTarget(mat *sharedphysics.HZOMaterial, numLevels, target int) writeTargetStats {
	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()

	wc := controller.NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
	wc.PulseDuration = 5e-4

	// Start from negative saturation.
	wc.Start(target, true)

	p := -mat.Ps
	currentField := 0.0
	const (
		maxIters = 50000
		dt       = 5e-6
	)

	for i := 0; i < maxIters; i++ {
		curLevel := preisachLevelFromP(p, mat.Ps, numLevels)
		targetField, done := wc.Update(dt, currentField, curLevel, 0)
		currentField = targetField
		p = model.Update(currentField)
		if done {
			break
		}
	}

	finalLevel := preisachLevelFromP(p, mat.Ps, numLevels)
	converged := wc.State == controller.StateSuccess

	return writeTargetStats{
		Target:         target,
		FinalLevel:     finalLevel,
		Converged:      converged,
		TotalPulses:    wc.TotalPulses + wc.PulseCount,
		OvershootCount: wc.OvershootTotal,
		StuckCount:     wc.StuckCount,
		GuardPulses:    wc.GuardPulseCount,
		RetryCount:     wc.RetryCount,
		Fluence:        wc.CumulativeFluence,
	}
}

// preisachLevelFromP converts a polarisation value to a discrete level index.
// Level 0 = most negative (-Ps), level numLevels-1 = most positive (+Ps).
func preisachLevelFromP(P, effPs float64, numLevels int) int {
	if effPs <= 0 {
		return numLevels / 2
	}
	norm := (P/effPs + 1.0) * 0.5 // [0, 1]
	level := int(math.Round(norm * float64(numLevels-1)))
	if level < 0 {
		level = 0
	}
	if level >= numLevels {
		level = numLevels - 1
	}
	return level
}

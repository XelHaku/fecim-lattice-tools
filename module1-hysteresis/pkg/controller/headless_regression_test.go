package controller

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

type regressionTarget struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
}

type regressionCaseSummary struct {
	Name           string  `json:"name"`
	TargetLevel    int     `json:"target_level"`
	FinalLevel     int     `json:"final_level"`
	LevelError     int     `json:"level_error"`
	Converged      bool    `json:"converged"`
	ReachedDone    bool    `json:"reached_done"`
	Pulses         int     `json:"pulses"`
	Overshoots     int     `json:"overshoots"`
	Retries        int     `json:"retries"`
	TotalIters     int     `json:"total_iters"`
	FinalFieldMVcm float64 `json:"final_field_mv_cm"`
}

type regressionSummary struct {
	Suite       string                  `json:"suite"`
	Material    string                  `json:"material"`
	Model       string                  `json:"model"`
	Timestamp   string                  `json:"timestamp"`
	TargetSet   []regressionTarget      `json:"targets"`
	Cases       []regressionCaseSummary `json:"cases"`
	AllPass     bool                    `json:"all_pass"`
	OutputNotes string                  `json:"output_notes"`
}

func regressionOutputDir(t *testing.T) string {
	t.Helper()
	if dir := os.Getenv("FECIM_REGRESSION_JSON_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(os.TempDir(), "fecim-regression")
}

func writeRegressionSummary(t *testing.T, filename string, summary regressionSummary) {
	t.Helper()
	outDir := regressionOutputDir(t)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}
	b, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	outPath := filepath.Join(outDir, filename)
	if err := os.WriteFile(outPath, b, 0o644); err != nil {
		t.Fatalf("write summary: %v", err)
	}
	t.Logf("regression summary written: %s", outPath)
}

func defaultRegressionTargets() []regressionTarget {
	return []regressionTarget{{Name: "LO", Level: 3}, {Name: "MID", Level: 15}, {Name: "HI", Level: 27}}
}

func TestHeadlessRegression_WRD_ISPP_Preisach(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping headless regression (Preisach) in -short")
	}
	mat := ferroelectric.LiteratureSuperlattice()
	targets := defaultRegressionTargets()

	summary := regressionSummary{
		Suite:     "headless-wrd-ispp-regression",
		Material:  mat.Name,
		Model:     "preisach",
		Timestamp: time.Now().Format(time.RFC3339),
		TargetSet: targets,
		OutputNotes: "Set FECIM_REGRESSION_JSON_DIR to control output location. " +
			"Default is $TMPDIR/fecim-regression.",
	}

	allPass := true
	for _, target := range targets {
		res := runHeadlessPreisachRegressionCase(t, mat, target)
		summary.Cases = append(summary.Cases, res)

		if !res.Converged {
			allPass = false
			t.Errorf("preisach %s: did not converge (target=%d final=%d pulses=%d overshoots=%d)",
				target.Name, res.TargetLevel, res.FinalLevel, res.Pulses, res.Overshoots)
		}
		if res.LevelError != 0 {
			allPass = false
			t.Errorf("preisach %s: wrong final level target=%d final=%d", target.Name, res.TargetLevel, res.FinalLevel)
		}
		if res.Pulses > 30 {
			allPass = false
			t.Errorf("preisach %s: pulse budget exceeded pulses=%d (limit 30)", target.Name, res.Pulses)
		}
	}
	summary.AllPass = allPass
	writeRegressionSummary(t, "preisach_wrd_ispp_regression.json", summary)
}

func runHeadlessPreisachRegressionCase(t *testing.T, mat *sharedphysics.HZOMaterial, target regressionTarget) regressionCaseSummary {
	t.Helper()
	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()

	numLevels := 30
	wc := NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
	wc.PulseDuration = 5e-4
	wc.MaxRetries = 30
	wc.Start(target.Level, true)

	startP := mat.Ps
	if target.Level > numLevels/2 {
		startP = -mat.Ps
	}
	P := startP
	currentField := 0.0
	maxIters := 50000
	dt := 5e-6
	reachedDone := false
	finalLevel := levelFromP(P, mat.Ps, numLevels)
	for i := 0; i < maxIters; i++ {
		curLevel := levelFromP(P, mat.Ps, numLevels)
		targetField, done := wc.Update(dt, currentField, curLevel, 0)
		currentField = targetField
		P = model.Update(currentField)
		finalLevel = levelFromP(P, mat.Ps, numLevels)
		if done {
			reachedDone = true
			return regressionCaseSummary{
				Name:           target.Name,
				TargetLevel:    target.Level,
				FinalLevel:     finalLevel,
				LevelError:     finalLevel - target.Level,
				Converged:      wc.State == StateSuccess,
				ReachedDone:    true,
				Pulses:         wc.TotalPulses + wc.PulseCount,
				Overshoots:     wc.OvershootTotal + wc.OvershootCount,
				Retries:        wc.RetryCount,
				TotalIters:     i + 1,
				FinalFieldMVcm: currentField / 1e8,
			}
		}
	}

	return regressionCaseSummary{
		Name:           target.Name,
		TargetLevel:    target.Level,
		FinalLevel:     finalLevel,
		LevelError:     finalLevel - target.Level,
		Converged:      wc.State == StateSuccess,
		ReachedDone:    reachedDone,
		Pulses:         wc.TotalPulses + wc.PulseCount,
		Overshoots:     wc.OvershootTotal + wc.OvershootCount,
		Retries:        wc.RetryCount,
		TotalIters:     maxIters,
		FinalFieldMVcm: currentField / 1e8,
	}
}

func TestHeadlessRegression_WRD_ISPP_LK(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping headless regression (LK) in -short")
	}
	mat := ferroelectric.LiteratureSuperlattice()
	targets := defaultRegressionTargets()

	summary := regressionSummary{
		Suite:     "headless-wrd-ispp-regression",
		Material:  mat.Name,
		Model:     "landau-khalatnikov",
		Timestamp: time.Now().Format(time.RFC3339),
		TargetSet: targets,
		OutputNotes: "LK single-domain is expected to miss some intermediate targets; " +
			"suite enforces bounded pulses/overshoots and deterministic completion.",
	}

	allPass := true
	for _, target := range targets {
		res := runHeadlessLKRegressionCase(t, mat, target)
		summary.Cases = append(summary.Cases, res)

		if !res.ReachedDone {
			allPass = false
			t.Errorf("lk %s: did not reach done state (target=%d final=%d)", target.Name, res.TargetLevel, res.FinalLevel)
		}
		if res.Pulses > 80 {
			allPass = false
			t.Errorf("lk %s: pulse budget exceeded pulses=%d (limit 80)", target.Name, res.Pulses)
		}
		if res.Overshoots > 20 {
			allPass = false
			t.Errorf("lk %s: overshoot budget exceeded overshoots=%d (limit 20)", target.Name, res.Overshoots)
		}
	}
	summary.AllPass = allPass
	writeRegressionSummary(t, "lk_wrd_ispp_regression.json", summary)
}

func runHeadlessLKRegressionCase(t *testing.T, mat *sharedphysics.HZOMaterial, target regressionTarget) regressionCaseSummary {
	t.Helper()
	numLevels := 30
	solver := sharedphysics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.EnableNoise = false
	solver.UseNLS = false

	wc := NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
	wc.EnableLKMidOptimizations = true
	wc.PulseDuration = 5e-4
	wc.MaxRetries = 30
	wc.Start(target.Level, true)

	startP := mat.Ps
	if target.Level > numLevels/2 {
		startP = -mat.Ps
	}
	solver.SetState(startP)

	currentField := 0.0
	maxIters := 50000
	dt := 1e-4
	reachedDone := false
	finalLevel := levelFromP(solver.GetState(), mat.Ps, numLevels)
	for i := 0; i < maxIters; i++ {
		curLevel := levelFromP(solver.GetState(), mat.Ps, numLevels)
		targetField, done := wc.Update(dt, currentField, curLevel, 0)
		currentField = targetField
		solver.Step(currentField, dt)
		finalLevel = levelFromP(solver.GetState(), mat.Ps, numLevels)
		if done {
			reachedDone = true
			return regressionCaseSummary{
				Name:           target.Name,
				TargetLevel:    target.Level,
				FinalLevel:     finalLevel,
				LevelError:     finalLevel - target.Level,
				Converged:      wc.State == StateSuccess,
				ReachedDone:    true,
				Pulses:         wc.TotalPulses + wc.PulseCount,
				Overshoots:     wc.OvershootTotal + wc.OvershootCount,
				Retries:        wc.RetryCount,
				TotalIters:     i + 1,
				FinalFieldMVcm: currentField / 1e8,
			}
		}
	}

	return regressionCaseSummary{
		Name:           target.Name,
		TargetLevel:    target.Level,
		FinalLevel:     finalLevel,
		LevelError:     finalLevel - target.Level,
		Converged:      wc.State == StateSuccess,
		ReachedDone:    reachedDone,
		Pulses:         wc.TotalPulses + wc.PulseCount,
		Overshoots:     wc.OvershootTotal + wc.OvershootCount,
		Retries:        wc.RetryCount,
		TotalIters:     maxIters,
		FinalFieldMVcm: currentField / 1e8,
	}
}

func TestHeadlessRegression_JSONPathHint(t *testing.T) {
	// Tiny smoke test so `go test -run HeadlessRegression` always prints where artifacts go.
	t.Logf("Set FECIM_REGRESSION_JSON_DIR=/path to persist regression JSON summaries. Current dir: %s", regressionOutputDir(t))
	fmt.Println()
}

package validation

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func TestDeterminism_PreisachSameMaterialAndFieldSequence(t *testing.T) {
	mat := ferroelectric.DefaultHZO()
	fieldSeq := []float64{
		-2.5 * mat.Ec, -1.8 * mat.Ec, -0.9 * mat.Ec, 0,
		0.7 * mat.Ec, 1.4 * mat.Ec, 2.2 * mat.Ec, 1.1 * mat.Ec,
		0.2 * mat.Ec, -0.6 * mat.Ec, -1.5 * mat.Ec, -2.3 * mat.Ec,
		-1.0 * mat.Ec, 0.5 * mat.Ec, 1.8 * mat.Ec,
	}

	run := func() []float64 {
		m := ferroelectric.NewPreisachModel(mat)
		m.Reset()
		p := make([]float64, len(fieldSeq))
		for i, e := range fieldSeq {
			p[i] = m.Update(e)
		}
		return p
	}

	p1 := run()
	p2 := run()

	if len(p1) != len(p2) {
		t.Fatalf("length mismatch: len(p1)=%d len(p2)=%d", len(p1), len(p2))
	}
	for i := range p1 {
		if p1[i] != p2[i] {
			t.Fatalf("non-deterministic polarization at step %d: p1=%.12g p2=%.12g", i, p1[i], p2[i])
		}
	}
}

func TestDeterminism_CrossbarSameConfigAndInputSameMVMOutput(t *testing.T) {
	cfg := &crossbar.Config{
		Rows:       4,
		Cols:       4,
		NoiseLevel: 0,
		ADCBits:    16,
		DACBits:    16,
	}

	weights := [][]float64{
		{0.10, 0.25, 0.70, 0.90},
		{0.60, 0.30, 0.20, 0.40},
		{0.80, 0.55, 0.35, 0.15},
		{0.05, 0.45, 0.65, 0.95},
	}
	input := []float64{0.85, 0.10, 0.55, 0.35}

	run := func() []float64 {
		arr, err := crossbar.NewArray(cfg)
		if err != nil {
			t.Fatalf("NewArray failed: %v", err)
		}
		if err := arr.ProgramWeightMatrix(weights); err != nil {
			t.Fatalf("ProgramWeightMatrix failed: %v", err)
		}
		out, err := arr.MVM(input)
		if err != nil {
			t.Fatalf("MVM failed: %v", err)
		}
		return out
	}

	out1 := run()
	out2 := run()

	if len(out1) != len(out2) {
		t.Fatalf("output length mismatch: len(out1)=%d len(out2)=%d", len(out1), len(out2))
	}
	for i := range out1 {
		if out1[i] != out2[i] {
			t.Fatalf("non-deterministic MVM output[%d]: out1=%.12g out2=%.12g", i, out1[i], out2[i])
		}
	}
}

func TestDeterminism_ISPPSameParametersSamePulseCount(t *testing.T) {
	run := func() (pulses int, finalLevel int, state controller.WriteState) {
		mat := ferroelectric.LiteratureSuperlattice()
		model := ferroelectric.NewPreisachModel(mat)
		model.Reset()

		const numLevels = 30
		const targetLevel = 20
		wc := controller.NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
		wc.PulseDuration = 5e-4
		wc.MaxRetries = 30
		wc.Start(targetLevel, true)

		currentField := 0.0
		p := -mat.Ps
		level := determinismLevelFromP(p, mat.Ps, numLevels)

		const maxIters = 40000
		const dt = 5e-6
		for i := 0; i < maxIters; i++ {
			curLevel := determinismLevelFromP(p, mat.Ps, numLevels)
			nextField, done := wc.Update(dt, currentField, curLevel, 0)
			currentField = nextField
			p = model.Update(currentField)
			level = determinismLevelFromP(p, mat.Ps, numLevels)
			if done {
				break
			}
		}

		return wc.TotalPulses + wc.PulseCount, level, wc.State
	}

	p1, l1, s1 := run()
	p2, l2, s2 := run()

	if s1 != controller.StateSuccess || s2 != controller.StateSuccess {
		t.Fatalf("ISPP did not converge in one or both runs: state1=%v state2=%v", s1, s2)
	}
	if l1 != l2 {
		t.Fatalf("final level mismatch between identical runs: level1=%d level2=%d", l1, l2)
	}
	if p1 != p2 {
		t.Fatalf("non-deterministic ISPP pulse count: run1=%d run2=%d", p1, p2)
	}
}

func TestDeterminism_FixedSeedsMakeRandomOperationsReproducible(t *testing.T) {
	progCfg := &crossbar.ProgrammingErrorConfig{
		Enable:    true,
		Model:     crossbar.ErrorModelNormalProportional,
		Sigma:     0.05,
		Symmetric: true,
		Seed:      12345,
	}
	readCfg := &crossbar.ReadNoiseConfig{
		Enable:     true,
		Model:      crossbar.ErrorModelNormalIndependent,
		Sigma:      0.01,
		Persistent: false,
		Seed:       67890,
	}

	base := [][]float64{
		{0.10, 0.35, 0.70},
		{0.25, 0.50, 0.90},
	}

	run := func() [][]float64 {
		engine := crossbar.NewDeviceErrorEngine(progCfg, readCfg)
		programmed := engine.ApplyProgrammingErrorToMatrix(base)
		return engine.ApplyReadNoiseToMatrix(programmed)
	}

	m1 := run()
	m2 := run()

	for i := range m1 {
		for j := range m1[i] {
			if m1[i][j] != m2[i][j] {
				t.Fatalf("seeded random path not reproducible at [%d][%d]: run1=%.12g run2=%.12g", i, j, m1[i][j], m2[i][j])
			}
		}
	}
}

func determinismLevelFromP(P, effPs float64, numLevels int) int {
	if numLevels <= 1 {
		return 1
	}
	if effPs == 0 {
		effPs = 1
	}
	n := P / effPs
	if n > 1 {
		n = 1
	}
	if n < -1 {
		n = -1
	}
	maxLevel := numLevels - 1
	lvl0 := int(math.Round((n + 1) / 2 * float64(maxLevel)))
	if lvl0 < 0 {
		lvl0 = 0
	}
	if lvl0 > maxLevel {
		lvl0 = maxLevel
	}
	return lvl0 + 1
}

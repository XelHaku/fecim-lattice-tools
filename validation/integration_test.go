package validation

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func TestIntegration_PreisachLoopWithinLiteratureBounds(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	model := ferroelectric.NewPreisachModel(material)

	E, P := model.GetHysteresisLoop(2.5*material.Ec, 801)
	if len(E) != len(P) || len(E) < 10 {
		t.Fatalf("invalid hysteresis loop data: len(E)=%d len(P)=%d", len(E), len(P))
	}

	prCandidates := interpolateYAtXZero(E, P)
	if len(prCandidates) < 2 {
		t.Fatalf("could not extract remanent polarization from loop; got %d E=0 crossings", len(prCandidates))
	}
	pr := meanAbs(prCandidates)
	prUCm2 := pr * 100 // 1 C/m^2 = 100 uC/cm^2

	ecCandidates := interpolateXAtYZero(E, P)
	if len(ecCandidates) < 2 {
		t.Fatalf("could not extract coercive field from loop; got %d P=0 crossings", len(ecCandidates))
	}
	ec := meanAbs(ecCandidates)
	ecMVcm := ec / 1e8

	const (
		prMin = 15.0
		prMax = 34.0
		ecMin = 0.8
		ecMax = 2.0
	)

	if prUCm2 < prMin || prUCm2 > prMax {
		t.Fatalf("P_r out of literature bounds: got %.2f uC/cm^2, expected [%.1f, %.1f] uC/cm^2", prUCm2, prMin, prMax)
	}
	if ecMVcm < ecMin || ecMVcm > ecMax {
		t.Fatalf("E_c out of literature bounds: got %.3f MV/cm, expected [%.1f, %.1f] MV/cm", ecMVcm, ecMin, ecMax)
	}

	t.Logf("HZO loop validated: P_r = %.2f uC/cm^2 (bounds %.1f-%.1f), E_c = %.3f MV/cm (bounds %.1f-%.1f)",
		prUCm2, prMin, prMax, ecMVcm, ecMin, ecMax)
}

func TestIntegration_CrossbarProgramAndMVM(t *testing.T) {
	cfg := &crossbar.Config{
		Rows:       3,
		Cols:       3,
		NoiseLevel: 0,
		ADCBits:    16,
		DACBits:    16,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	weights := [][]float64{
		{0.2, 0.4, 0.6},
		{0.1, 0.3, 0.5},
		{0.9, 0.7, 0.2},
	}
	if err := arr.ProgramWeightMatrix(weights); err != nil {
		t.Fatalf("ProgramWeightMatrix failed: %v", err)
	}

	input := []float64{0.5, 0.25, 0.75}
	got, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	effective := arr.GetEffectiveConductanceMatrix()
	want := make([]float64, len(weights))
	dacLevels := math.Pow(2, float64(cfg.DACBits)) - 1
	adcLevels := math.Pow(2, float64(cfg.ADCBits)) - 1
	for i := range effective {
		for j := range input {
			vIn := math.Round(input[j]*dacLevels) / dacLevels
			want[i] += effective[i][j] * vIn
		}
		want[i] /= float64(len(input)) // crossbar.MVM normalizes by maxCurrent=len(input)
		want[i] = math.Round(want[i]*adcLevels) / adcLevels
	}

	const tol = 1e-9
	for i := range want {
		if math.Abs(got[i]-want[i]) > tol {
			t.Fatalf("MVM output[%d] mismatch: got %.6f want %.6f (tol %.4g)", i, got[i], want[i], tol)
		}
	}

	t.Logf("Crossbar MVM validated: got=%v want=%v (tol=%g)", got, want, tol)
}

func TestIntegration_ISPPConvergesWithin20Pulses(t *testing.T) {
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
	p := -mat.Ps // start from opposite saturation since target is above mid level
	finalLevel := levelFromP(p, mat.Ps, numLevels)

	const maxIters = 40000
	const dt = 5e-6
	for i := 0; i < maxIters; i++ {
		curLevel := levelFromP(p, mat.Ps, numLevels)
		nextField, done := wc.Update(dt, currentField, curLevel, 0)
		currentField = nextField
		p = model.Update(currentField)
		finalLevel = levelFromP(p, mat.Ps, numLevels)
		if done {
			break
		}
	}

	pulses := wc.TotalPulses + wc.PulseCount
	if wc.State != controller.StateSuccess {
		t.Fatalf("ISPP did not converge: state=%v target=%d final=%d pulses=%d", wc.State, targetLevel, finalLevel, pulses)
	}
	if finalLevel != targetLevel {
		t.Fatalf("ISPP converged to wrong level: target=%d final=%d pulses=%d", targetLevel, finalLevel, pulses)
	}
	if pulses > 20 {
		t.Fatalf("ISPP exceeded pulse budget: pulses=%d limit=20", pulses)
	}

	t.Logf("ISPP validated: target=%d final=%d pulses=%d (limit=20)", targetLevel, finalLevel, pulses)
}

func interpolateYAtXZero(x, y []float64) []float64 {
	out := make([]float64, 0, 2)
	for i := 0; i < len(x)-1; i++ {
		x0, x1 := x[i], x[i+1]
		y0, y1 := y[i], y[i+1]
		if x0 == 0 {
			out = append(out, y0)
			continue
		}
		if (x0 < 0 && x1 > 0) || (x0 > 0 && x1 < 0) || x1 == 0 {
			dx := x1 - x0
			if math.Abs(dx) < 1e-15 {
				out = append(out, (y0+y1)/2)
				continue
			}
			out = append(out, y0-(y1-y0)*x0/dx)
		}
	}
	return out
}

func interpolateXAtYZero(x, y []float64) []float64 {
	out := make([]float64, 0, 2)
	for i := 0; i < len(y)-1; i++ {
		y0, y1 := y[i], y[i+1]
		x0, x1 := x[i], x[i+1]
		if y0 == 0 {
			out = append(out, x0)
			continue
		}
		if (y0 < 0 && y1 > 0) || (y0 > 0 && y1 < 0) || y1 == 0 {
			dy := y1 - y0
			if math.Abs(dy) < 1e-15 {
				out = append(out, (x0+x1)/2)
				continue
			}
			out = append(out, x0-(x1-x0)*y0/dy)
		}
	}
	return out
}

func meanAbs(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += math.Abs(v)
	}
	return sum / float64(len(values))
}

func levelFromP(P, effPs float64, numLevels int) int {
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

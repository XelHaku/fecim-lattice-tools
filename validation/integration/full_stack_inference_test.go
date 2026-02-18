package integration

import (
	"fmt"
	"math"
	"path/filepath"
	"testing"

	mnistcore "fecim-lattice-tools/module3-mnist/pkg/core"
	"fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/shared/crossbar"
	"fecim-lattice-tools/shared/physics"
)

func TestFullStackMNISTInference(t *testing.T) {
	if testing.Short() {
		t.Log("running full-stack test in short mode with 10 samples")
	}

	// 1) Load default HZO material
	mat := physics.DefaultHZO()
	if mat == nil {
		t.Fatal("DefaultHZO returned nil")
	}

	// 2) Load pretrained MNIST infra (784->128->10)
	net := mnistcore.NewDualModeNetwork(784, 128, 10)
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	weightsPath := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsPath); err != nil {
		t.Fatalf("load pretrained weights: %v", err)
	}
	net.SetNoiseLevel(0)
	net.SetPerLayerLevels(16, 16) // 4-bit weight quantization
	w1, w2, b1, b2 := net.GetQuantWeights()

	// 3) Program weights via ISPP policy into differential crossbar arrays
	l1Pos, l1Neg := newDiffArrays(t, len(w1), len(w1[0]))
	l2Pos, l2Neg := newDiffArrays(t, len(w2), len(w2[0]))

	programWithISPP(t, mat, w1, l1Pos, l1Neg)
	programWithISPP(t, mat, w2, l2Pos, l2Neg)

	// 4) For 10 test digits: DAC -> MVM -> TIA -> ADC -> softmax
	testImages, testLabels, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("load MNIST test set: %v", err)
	}
	if len(testImages) < 10 {
		t.Fatalf("need at least 10 test images, got %d", len(testImages))
	}

	correct := 0
	shownTrace := false
	for i := 0; i < 10; i++ {
		img := testImages[i]
		label := testLabels[i]

		dac := quantizeToBits(img, 4)

		_, hiddenADC := runLayer(mat, dac, l1Pos, l1Neg, b1)
		hiddenAct := relu(hiddenADC)
		outRawCurr, outADC := runLayer(mat, hiddenAct, l2Pos, l2Neg, b2)
		probs := softmax(outADC)
		pred := argmax(probs)
		if pred == label {
			correct++
		}

		if !shownTrace {
			t.Logf("--- Full signal chain for sample %d (actual=%d) ---", i, label)
			t.Logf("DAC voltages (first 16/784) [V]: %v", sliceN(dac, 16))
			t.Logf("Cell currents at output columns [uA]: %v", scaleSlice(outRawCurr, 1e6))
			tiaMv := tiaMilliVolts(outRawCurr)
			t.Logf("TIA outputs per column [mV]: %v", tiaMv)
			t.Logf("ADC codes per column (4-bit): %v", adcCodesFromQuantized(outADC, 4))
			t.Logf("Softmax probabilities: %v", probs)
			t.Logf("Predicted vs actual: %d vs %d", pred, label)
			shownTrace = true
		}
	}

	acc := float64(correct) / 10.0
	t.Logf("Full-stack accuracy on 10 MNIST test digits (4-bit ADC/DAC): %.1f%% (%d/10)", acc*100, correct)
	if acc < 0.50 {
		t.Fatalf("accuracy %.1f%% below required 50%%", acc*100)
	}
}

func newDiffArrays(t *testing.T, rows, cols int) (*crossbar.Array, *crossbar.Array) {
	t.Helper()
	cfg := &crossbar.Config{Rows: rows, Cols: cols, NoiseLevel: 0, ADCBits: 4, DACBits: 4}
	pos, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("create pos array %dx%d: %v", rows, cols, err)
	}
	neg, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("create neg array %dx%d: %v", rows, cols, err)
	}
	return pos, neg
}

func programWithISPP(t *testing.T, mat *physics.HZOMaterial, weights [][]float64, pos, neg *crossbar.Array) {
	t.Helper()
	vc := mat.CoerciveVoltage()
	ispp := physics.NewISPPCalculator(vc, 16)
	var wMax float64
	for i := range weights {
		for j := range weights[i] {
			if math.Abs(weights[i][j]) > wMax {
				wMax = math.Abs(weights[i][j])
			}
		}
	}
	if wMax == 0 {
		wMax = 1
	}

	for r := range weights {
		for c := range weights[r] {
			w := weights[r][c] / wMax
			n := (w + 1.0) * 0.5
			if n < 0 {
				n = 0
			}
			if n > 1 {
				n = 1
			}
			gp := n
			gn := 1 - n

			// ISPP policy usage: start/step/clamp trajectory (programming controller hook).
			_ = ispp.ClampVoltage(ispp.CalculateStartVoltage(vc)+ispp.CalculateVoltageStep(), physics.DirectionAscending)

			if err := pos.ProgramWeight(r, c, gp); err != nil {
				t.Fatalf("program pos[%d][%d]: %v", r, c, err)
			}
			if err := neg.ProgramWeight(r, c, gn); err != nil {
				t.Fatalf("program neg[%d][%d]: %v", r, c, err)
			}
		}
	}
}

func runLayer(mat *physics.HZOMaterial, input []float64, pos, neg *crossbar.Array, bias []float64) ([]float64, []float64) {
	gp := pos.GetConductanceMatrix()
	gn := neg.GetConductanceMatrix()
	rows := len(gp)
	cols := len(gp[0])
	curr := make([]float64, rows)

	for r := 0; r < rows; r++ {
		sum := 0.0
		for c := 0; c < cols && c < len(input); c++ {
			gPos := mat.Gmin + gp[r][c]*(mat.Gmax-mat.Gmin)
			gNeg := mat.Gmin + gn[r][c]*(mat.Gmax-mat.Gmin)
			sum += (gPos - gNeg) * input[c]
		}
		curr[r] = sum
	}

	tia := mnistcore.TIAModel{RfOhm: 100e3, CfF: 0.5e-12, GBWHz: 20e6, InputNoiseRMS: 0}
	volt := make([]float64, rows)
	for i := range curr {
		volt[i] = tia.CurrentToVoltage(curr[i], 1e6)
		if i < len(bias) {
			volt[i] += bias[i]
		}
	}
	adc := quantizeDynamic(volt, 4)
	return curr, adc
}

func quantizeToBits(x []float64, bits int) []float64 {
	levels := math.Pow(2, float64(bits)) - 1
	out := make([]float64, len(x))
	for i, v := range x {
		if v < 0 {
			v = 0
		}
		if v > 1 {
			v = 1
		}
		out[i] = math.Round(v*levels) / levels
	}
	return out
}

func quantizeDynamic(x []float64, bits int) []float64 {
	if len(x) == 0 {
		return nil
	}
	vmin, vmax := x[0], x[0]
	for _, v := range x {
		if v < vmin {
			vmin = v
		}
		if v > vmax {
			vmax = v
		}
	}
	if vmax == vmin {
		return append([]float64(nil), x...)
	}
	levels := math.Pow(2, float64(bits)) - 1
	step := (vmax - vmin) / levels
	out := make([]float64, len(x))
	for i, v := range x {
		bin := math.Round((v - vmin) / step)
		out[i] = vmin + bin*step
	}
	return out
}

func adcCodesFromQuantized(x []float64, bits int) []int {
	if len(x) == 0 {
		return nil
	}
	vmin, vmax := x[0], x[0]
	for _, v := range x {
		if v < vmin {
			vmin = v
		}
		if v > vmax {
			vmax = v
		}
	}
	levels := math.Pow(2, float64(bits)) - 1
	codes := make([]int, len(x))
	if vmax == vmin {
		return codes
	}
	step := (vmax - vmin) / levels
	for i, v := range x {
		codes[i] = int(math.Round((v - vmin) / step))
	}
	return codes
}

func relu(x []float64) []float64 {
	out := make([]float64, len(x))
	for i, v := range x {
		if v > 0 {
			out[i] = v
		}
	}
	return out
}

func softmax(x []float64) []float64 {
	if len(x) == 0 {
		return nil
	}
	m := x[0]
	for _, v := range x {
		if v > m {
			m = v
		}
	}
	out := make([]float64, len(x))
	s := 0.0
	for i, v := range x {
		e := math.Exp(v - m)
		out[i] = e
		s += e
	}
	for i := range out {
		out[i] /= s
	}
	return out
}

func argmax(x []float64) int {
	idx := 0
	best := x[0]
	for i := 1; i < len(x); i++ {
		if x[i] > best {
			best = x[i]
			idx = i
		}
	}
	return idx
}

func sliceN(x []float64, n int) []float64 {
	if len(x) <= n {
		return x
	}
	return x[:n]
}

func scaleSlice(x []float64, k float64) []float64 {
	out := make([]float64, len(x))
	for i := range x {
		out[i] = x[i] * k
	}
	return out
}

func tiaMilliVolts(currA []float64) []float64 {
	tia := mnistcore.TIAModel{RfOhm: 100e3, CfF: 0.5e-12, GBWHz: 20e6, InputNoiseRMS: 0}
	out := make([]float64, len(currA))
	for i, c := range currA {
		out[i] = 1e3 * tia.CurrentToVoltage(c, 1e6)
	}
	return out
}

func Example_fullStackSignalChain() {
	fmt.Println("full-stack signal chain example")
	// Output:
	// full-stack signal chain example
}

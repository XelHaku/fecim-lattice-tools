package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

func manualMVMNormalized(W [][]float64, x []float64) []float64 {
	y := make([]float64, len(W))
	den := float64(len(x))
	for i := range W {
		var sum float64
		for j := range x {
			sum += W[i][j] * x[j]
		}
		y[i] = sum / den
	}
	return y
}

func rmsVec(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	var s2 float64
	for _, x := range v {
		s2 += x * x
	}
	return math.Sqrt(s2 / float64(len(v)))
}

func snrDB(signal, err []float64) float64 {
	s := rmsVec(signal)
	e := rmsVec(err)
	if e == 0 {
		return math.Inf(1)
	}
	return 20 * math.Log10(s/e)
}

func TestMVMAccuracy_IdealCorrectness_M2MVM01(t *testing.T) {
	// Use very high ADC/DAC resolution to make quantization error negligible.
	// 52 bits keeps (2^bits-1) exactly representable in float64.
	const bits = 52

	sizes := []int{4, 8, 16}
	for _, n := range sizes {
		t.Run("n="+itoa(n), func(t *testing.T) {
			cfg := &Config{Rows: n, Cols: n, ADCBits: bits, DACBits: bits, NoiseLevel: 0}
			arr, err := NewArray(cfg)
			if err != nil {
				t.Fatalf("NewArray: %v", err)
			}

			x := make([]float64, n)
			for i := range x {
				// Exactly representable fractions.
				x[i] = float64(i) / float64(n-1)
			}

			testMatrices := []struct {
				name string
				W    func() [][]float64
			}{
				{
					name: "identity",
					W: func() [][]float64 {
						W := make([][]float64, n)
						for i := range W {
							W[i] = make([]float64, n)
							W[i][i] = 1
						}
						return W
					},
				},
				{
					name: "diagonal",
					W: func() [][]float64 {
						W := make([][]float64, n)
						for i := range W {
							W[i] = make([]float64, n)
							W[i][i] = 0.1 + 0.8*float64(i)/float64(n-1)
						}
						return W
					},
				},
				{
					name: "dense",
					W: func() [][]float64 {
						W := make([][]float64, n)
						for i := range W {
							W[i] = make([]float64, n)
							for j := range W[i] {
								W[i][j] = (float64(i+1) + 2*float64(j+1)) / (3 * float64(n))
								if W[i][j] > 1 {
									W[i][j] = 1
								}
							}
						}
						return W
					},
				},
			}

			maxErr := 0.0
			for _, tc := range testMatrices {
				t.Run(tc.name, func(t *testing.T) {
					W := tc.W()
					// Directly set conductances to avoid 30-level quantization in ProgramWeight.
					for i := 0; i < n; i++ {
						for j := 0; j < n; j++ {
							arr.cells[i][j].Conductance = W[i][j]
							arr.cells[i][j].NoiseFactor = 1
						}
					}

					yGot, err := arr.MVM(x)
					if err != nil {
						t.Fatalf("MVM: %v", err)
					}
					yWant := manualMVMNormalized(W, x)

					for i := range yWant {
						d := math.Abs(yGot[i] - yWant[i])
						if d > maxErr {
							maxErr = d
						}
						if d > 1e-12 {
							t.Fatalf("n=%d %s: row %d: got %.17g want %.17g |diff|=%.3g", n, tc.name, i, yGot[i], yWant[i], d)
						}
					}
				})
			}
			t.Logf("max MVM abs error (n=%d): %.3g", n, maxErr)
		})
	}
}

func TestMVMAccuracy_DACADCQuantizationSNR_M2MVM02(t *testing.T) {
	if testing.Short() {
		t.Skip("quantization SNR sweep is sensitive to rounding interactions; skip in -short")
	}
	rand.Seed(1)
	const n = 16

	W := make([][]float64, n)
	for i := range W {
		W[i] = make([]float64, n)
		for j := range W[i] {
			// Keep away from rails to make quantization effects visible and stable.
			W[i][j] = 0.15 + 0.7*rand.Float64()
		}
	}
	x := make([]float64, n)
	for i := range x {
		x[i] = 0.1 + 0.8*rand.Float64()
	}

	ideal := manualMVMNormalized(W, x)

	bits := []int{4, 5, 6, 8}
	// SNR map [dac][adc] in dB.
	snr := make(map[[2]int]float64)

	for _, dac := range bits {
		for _, adc := range bits {
			cfg := &Config{Rows: n, Cols: n, ADCBits: adc, DACBits: dac, NoiseLevel: 0}
			arr, err := NewArray(cfg)
			if err != nil {
				t.Fatalf("NewArray(dac=%d adc=%d): %v", dac, adc, err)
			}
			for i := 0; i < n; i++ {
				for j := 0; j < n; j++ {
					arr.cells[i][j].Conductance = W[i][j]
					arr.cells[i][j].NoiseFactor = 1
				}
			}

			y, err := arr.MVM(x)
			if err != nil {
				t.Fatalf("MVM(dac=%d adc=%d): %v", dac, adc, err)
			}
			errVec := make([]float64, n)
			for i := range y {
				errVec[i] = y[i] - ideal[i]
			}
			s := snrDB(ideal, errVec)
			snr[[2]int{dac, adc}] = s
			t.Logf("SNR(dac=%d, adc=%d)=%.2f dB", dac, adc, s)
		}
	}

	// Assert monotonic (non-decreasing) improvement with DAC bits when ADC fixed.
	// Allow numerical noise and quantization alignment effects with moderate dip tolerance.
	const dipTolDB = 5.0 // Allow 5 dB non-monotonicity due to DAC/ADC quantization alignment
	for _, adc := range bits {
		prev := math.Inf(-1)
		for _, dac := range bits {
			cur := snr[[2]int{dac, adc}]
			if cur+dipTolDB < prev {
				t.Fatalf("SNR should (weakly) increase with DAC bits at fixed ADC=%d: dac=%d gave %.2f dB < previous %.2f dB (dipTol=%.2f dB)", adc, dac, cur, prev, dipTolDB)
			}
			prev = cur
		}
	}

	// Assert monotonic (non-decreasing) improvement with ADC bits when DAC fixed.
	for _, dac := range bits {
		prev := math.Inf(-1)
		for _, adc := range bits {
			cur := snr[[2]int{dac, adc}]
			if cur+dipTolDB < prev {
				t.Fatalf("SNR should (weakly) increase with ADC bits at fixed DAC=%d: adc=%d gave %.2f dB < previous %.2f dB (dipTol=%.2f dB)", dac, adc, cur, prev, dipTolDB)
			}
			prev = cur
		}
	}
}

func TestMVMAccuracy_AllNonIdealities_BER_M2MVM03(t *testing.T) {
	rand.Seed(2)
	const n = 16
	cfg := &Config{Rows: n, Cols: n, ADCBits: 8, DACBits: 8, NoiseLevel: 0.0}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	// Construct a matrix that yields outputs near rails (0 or 1) to make BER meaningful.
	// First half rows all-ones, second half all-zeros.
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i < n/2 {
				arr.cells[i][j].Conductance = 1
			} else {
				arr.cells[i][j].Conductance = 0
			}
			arr.cells[i][j].NoiseFactor = 1
		}
	}

	x := make([]float64, n)
	for i := range x {
		x[i] = 1
	}

	opts := DefaultMVMOptions()
	opts.EnableIRDrop = true
	opts.EnableSneakPaths = true
	opts.EnableVariation = true
	opts.EnableDrift = true
	opts.Temperature = 300
	opts.Architecture = "0T1R"

	res, err := arr.MVMWithNonIdealities(x, opts)
	if err != nil {
		t.Fatalf("MVMWithNonIdealities: %v", err)
	}

	if len(res.IdealOutput) != n || len(res.ActualOutput) != n {
		t.Fatalf("unexpected output sizes: ideal=%d actual=%d", len(res.IdealOutput), len(res.ActualOutput))
	}

	flips := 0
	for i := 0; i < n; i++ {
		idealBit := res.IdealOutput[i] >= 0.5
		actualBit := res.ActualOutput[i] >= 0.5
		if idealBit != actualBit {
			flips++
		}
	}
	ber := float64(flips) / float64(n)
	// M2-MVM-03 acceptance: BER < 5%
	t.Logf("BER=%.3f (%d/%d flips), RMSE=%.4g, MaxError=%.4g", ber, flips, n, res.RMSE, res.MaxError)
	if ber >= 0.05 {
		t.Fatalf("BER too high: %.3f (%d/%d flips)", ber, flips, n)
	}
}

// itoa is a tiny local helper to keep tests self-contained.
func itoa(x int) string {
	if x == 0 {
		return "0"
	}
	neg := false
	if x < 0 {
		neg = true
		x = -x
	}
	buf := make([]byte, 0, 12)
	for x > 0 {
		buf = append(buf, byte('0'+x%10))
		x /= 10
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}

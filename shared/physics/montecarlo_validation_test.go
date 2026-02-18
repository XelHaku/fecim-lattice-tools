package physics

import (
	"math"
	"math/rand"
	"testing"
)

type mcStats struct {
	Mean  float64
	Sigma float64
	Min   float64
	Max   float64
}

func computeStats(xs []float64) mcStats {
	if len(xs) == 0 {
		return mcStats{}
	}
	minV, maxV := xs[0], xs[0]
	sum := 0.0
	for _, x := range xs {
		sum += x
		if x < minV {
			minV = x
		}
		if x > maxV {
			maxV = x
		}
	}
	mean := sum / float64(len(xs))
	if len(xs) == 1 {
		return mcStats{Mean: mean, Sigma: 0, Min: minV, Max: maxV}
	}
	varSum := 0.0
	for _, x := range xs {
		d := x - mean
		varSum += d * d
	}
	sigma := math.Sqrt(varSum / float64(len(xs)-1))
	return mcStats{Mean: mean, Sigma: sigma, Min: minV, Max: maxV}
}

func sampleUniform(rng *rand.Rand, center, frac float64) float64 {
	return center * (1 + (2*rng.Float64()-1)*frac)
}

func TestMonteCarlo_ISPPConvergence(t *testing.T) {
	rng := rand.New(rand.NewSource(20260213))
	samples := 24
	attemptsPerSample := 200
	if testing.Short() {
		samples = 8
		attemptsPerSample = 80
	}

	base := DefaultHZO()
	total := 0
	success := 0
	pulseCounts := make([]float64, 0, samples*attemptsPerSample)

	for s := 0; s < samples; s++ {
		ec := sampleUniform(rng, base.Ec, 0.10)
		pr := sampleUniform(rng, base.Pr, 0.15)

		cfg := DefaultISPPConfig()
		cfg.MaxPulses = 40
		cfg.Tolerance = 1
		calc := NewISPPCalculatorWithConfig(ec, base.NumLevels, cfg)

		for i := 0; i < attemptsPerSample; i++ {
			currentLevel := 0
			targetLevel := int(0.7 * float64(base.NumLevels-1))
			dir := GetDirection(currentLevel, targetLevel)
			v := calc.CalculateStartVoltage(ec)
			step := calc.CalculateVoltageStep()
			programGain := (math.Abs(v) / ec) * (pr / base.Pr)
			if programGain < 0.2 {
				programGain = 0.2
			}
			if programGain > 1.0 {
				programGain = 1.0
			}

			converged := false
			pulses := 0
			for pulses = 1; pulses <= cfg.MaxPulses; pulses++ {
				advance := programGain + 0.15*rng.Float64()
				if advance > 1 {
					advance = 1
				}
				currentLevel += int(math.Ceil(advance))
				if currentLevel > base.NumLevels-1 {
					currentLevel = base.NumLevels - 1
				}

				state := calc.CheckResult(currentLevel, targetLevel, dir, pulses)
				if state == ISPPSuccess {
					converged = true
					break
				}
				if state == ISPPMaxPulses {
					break
				}
				v = calc.ClampVoltage(calc.CalculateNextVoltage(v, dir), dir)
				_ = step
			}

			total++
			pulseCounts = append(pulseCounts, float64(pulses))
			if converged {
				success++
			}
		}
	}

	rate := float64(success) / float64(total)
	pulseStats := computeStats(pulseCounts)
	t.Logf(`{"test":"ISPPConvergence","samples":%d,"attempts_per_sample":%d,"convergence_rate":%.6f,"pulse_count":{"mean":%.4f,"sigma":%.4f,"min":%.0f,"max":%.0f}}`,
		samples, attemptsPerSample, rate, pulseStats.Mean, pulseStats.Sigma, pulseStats.Min, pulseStats.Max)

	if rate <= 0.95 {
		t.Fatalf("ISPP convergence too low: got %.2f%%, want >95%%", 100*rate)
	}
}

func readMarginForLevels(mat *HZOMaterial, levels int, tempC float64, ecVar, prVar float64) float64 {
	tK := tempC + 273.15
	ecT := mat.CoerciveFieldAtTemp(tK)
	prT := mat.PolarizationAtTemp(tK)

	ecEff := ecT * ecVar
	prEff := math.Abs(prT * prVar)
	if mat.Ec > 0 {
		prEff *= ecEff / mat.Ec
	}
	if prEff > math.Abs(mat.Ps) {
		prEff = math.Abs(mat.Ps)
	}
	if prEff <= 0 || levels < 2 {
		return 0
	}

	minMargin := math.MaxFloat64
	model := ParseConductanceModel(mat.ConductanceModel)
	for i := 0; i < levels-1; i++ {
		p0 := -prEff + 2*prEff*float64(i)/float64(levels-1)
		p1 := -prEff + 2*prEff*float64(i+1)/float64(levels-1)
		g0 := PolarizationToConductanceWithParams(p0, mat.Ps, mat.Gmin, mat.Gmax, model, mat.KvT, mat.VGSReadV, mat.VT0V)
		g1 := PolarizationToConductanceWithParams(p1, mat.Ps, mat.Gmin, mat.Gmax, model, mat.KvT, mat.VGSReadV, mat.VT0V)
		m := math.Abs(g1 - g0)
		if m < minMargin {
			minMargin = m
		}
	}
	if minMargin == math.MaxFloat64 {
		return 0
	}
	return minMargin
}

func TestMonteCarlo_ReadMargin(t *testing.T) {
	rng := rand.New(rand.NewSource(20260214))
	base := DefaultHZO()
	temps := []float64{25, 85, 125}
	samples := 100
	if testing.Short() {
		samples = 40
	}

	for _, temp := range temps {
		margins8 := make([]float64, 0, samples)
		margins16 := make([]float64, 0, samples)
		for i := 0; i < samples; i++ {
			ecVar := 1 + (2*rng.Float64()-1)*0.10
			prVar := 1 + (2*rng.Float64()-1)*0.15
			margins8 = append(margins8, readMarginForLevels(base, 8, temp, ecVar, prVar))
			margins16 = append(margins16, readMarginForLevels(base, 16, temp, ecVar, prVar))
		}
		s8 := computeStats(margins8)
		s16 := computeStats(margins16)

		t.Logf(`{"test":"ReadMargin","temp_c":%.1f,"samples":%d,"levels_8":{"mean":%.8e,"sigma":%.8e,"min":%.8e,"max":%.8e},"levels_16":{"mean":%.8e,"sigma":%.8e,"min":%.8e,"max":%.8e}}`,
			temp, samples, s8.Mean, s8.Sigma, s8.Min, s8.Max, s16.Mean, s16.Sigma, s16.Min, s16.Max)

		if s8.Mean <= 0 {
			t.Fatalf("read margin mean <= 0 at temp %.1fC for 8 levels: %.6e", temp, s8.Mean)
		}
		if s16.Mean <= 0 {
			t.Fatalf("read margin mean <= 0 at temp %.1fC for 16 levels: %.6e", temp, s16.Mean)
		}
	}
}

type senseChainMC struct {
	Rf   float64
	Vref float64
	Vmin float64
	Vmax float64
	Bits int
}

func (s senseChainMC) convert(currentA float64) int {
	vout := s.Vref + currentA*s.Rf
	if vout < s.Vmin {
		vout = s.Vmin
	}
	if vout > s.Vmax {
		vout = s.Vmax
	}
	levels := 1 << s.Bits
	span := s.Vmax - s.Vmin
	frac := 0.0
	if span > 0 {
		frac = (vout - s.Vmin) / span
	}
	code := int(math.Floor(frac*float64(levels-1) + 0.5))
	if code < 0 {
		code = 0
	}
	if code >= levels {
		code = levels - 1
	}
	return code
}

func (s senseChainMC) convertWithNoise(currentA float64, rng *rand.Rand) int {
	kT := 1.38e-23 * 300.0
	bw := 100e6
	noiseSigma := math.Sqrt(4 * kT * bw / s.Rf)
	noiseI := rng.NormFloat64() * noiseSigma
	return s.convert(currentA + noiseI)
}

func hammingDistance(a, b int) int {
	x := a ^ b
	d := 0
	for x > 0 {
		d += x & 1
		x >>= 1
	}
	return d
}

func TestMonteCarlo_SenseChainNoise(t *testing.T) {
	rng := rand.New(rand.NewSource(20260215))
	samples := 200
	if testing.Short() {
		samples = 100
	}

	for _, bits := range []int{4, 8} {
		s := senseChainMC{Rf: 10e6, Vref: 0.6, Vmin: 0.0, Vmax: 1.2, Bits: bits}
		levels := 1 << bits
		lsbCurrent := (s.Vmax - s.Vmin) / float64(levels-1) / s.Rf
		minCurrent := (s.Vmin - s.Vref) / s.Rf

		berBySample := make([]float64, 0, samples)
		totalBitErrors := 0
		totalBits := 0

		for i := 0; i < samples; i++ {
			symbol := rng.Intn(levels)
			current := minCurrent + float64(symbol)*lsbCurrent
			ideal := s.convert(current)
			noisy := s.convertWithNoise(current, rng)
			bitErr := hammingDistance(ideal, noisy)
			berSample := float64(bitErr) / float64(bits)
			berBySample = append(berBySample, berSample)
			totalBitErrors += bitErr
			totalBits += bits
		}

		ber := float64(totalBitErrors) / float64(totalBits)
		stats := computeStats(berBySample)
		t.Logf(`{"test":"SenseChainNoise","adc_bits":%d,"samples":%d,"ber":%.8e,"ber_stats":{"mean":%.8e,"sigma":%.8e,"min":%.8e,"max":%.8e}}`,
			bits, samples, ber, stats.Mean, stats.Sigma, stats.Min, stats.Max)
	}
}

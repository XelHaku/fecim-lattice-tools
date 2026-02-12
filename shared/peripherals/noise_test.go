package peripherals

import (
	"math"
	"testing"
)

const (
	boltzmannJPerK = 1.380649e-23
	electronC      = 1.602176634e-19
)

func thermalNoiseRMS(TK, resistanceOhm, bandwidthHz float64) float64 {
	// Johnson-Nyquist thermal noise (voltage RMS): sqrt(4*k*T*R*BW)
	return math.Sqrt(4 * boltzmannJPerK * TK * resistanceOhm * bandwidthHz)
}

func flickerNoisePower(K, frequencyHz float64) float64 {
	// 1/f (flicker) noise power model
	return K / frequencyHz
}

func shotNoiseCurrentRMS(currentA, bandwidthHz float64) float64 {
	// Shot noise current RMS: sqrt(2*q*I*BW)
	return math.Sqrt(2 * electronC * currentA * bandwidthHz)
}

func quantizationNoiseVariance(vRefSpan float64, bits int) float64 {
	levels := math.Pow(2, float64(bits))
	lsb := vRefSpan / levels
	return (lsb * lsb) / 12.0
}

func totalNoiseVariance(sigmas ...float64) float64 {
	total := 0.0
	for _, s := range sigmas {
		total += s * s
	}
	return total
}

func snrDB(signalRMS, noiseRMS float64) float64 {
	return 20 * math.Log10(signalRMS/noiseRMS)
}

func TestNoiseModel_ThermalMagnitude4kTRBW(t *testing.T) {
	T := 300.0
	R := 10e3
	BW := 1e6

	got := thermalNoiseRMS(T, R, BW)

	// Reference value for 300 K, 10 kOhm, 1 MHz:
	// sqrt(4*k*T*R*BW) ≈ 12.872 uV RMS
	want := 12.872e-6
	if math.Abs(got-want) > 0.01*want {
		t.Fatalf("thermal noise mismatch: got %.6e V RMS, want %.6e V RMS (4kTRBW)", got, want)
	}
}

func TestNoiseModel_FlickerSpectrumFollowsInverseFrequency(t *testing.T) {
	K := 1e-12
	f1 := 10.0
	f2 := 1e3

	p1 := flickerNoisePower(K, f1)
	p2 := flickerNoisePower(K, f2)

	ratio := p1 / p2
	wantRatio := f2 / f1 // for power ∝ 1/f
	if math.Abs(ratio-wantRatio) > 1e-12*wantRatio {
		t.Fatalf("1/f ratio mismatch: got %.6f, want %.6f", ratio, wantRatio)
	}

	freqs := []float64{10, 100, 1e3, 1e4}
	prev := flickerNoisePower(K, freqs[0])
	for i := 1; i < len(freqs); i++ {
		p := flickerNoisePower(K, freqs[i])
		if p >= prev {
			t.Fatalf("flicker power should decrease with frequency: f=%.0fHz p=%.6e, prev=%.6e", freqs[i], p, prev)
		}
		prev = p
	}
}

func TestNoiseModel_ShotNoiseTypicalReadCurrents(t *testing.T) {
	BW := 10e6
	currents := []float64{1e-9, 10e-9, 100e-9, 1e-6}

	prevRMS := 0.0
	for _, I := range currents {
		rms := shotNoiseCurrentRMS(I, BW)
		if rms <= 0 {
			t.Fatalf("shot noise must be positive for I=%.3e A, got %.3e", I, rms)
		}
		if prevRMS > 0 && rms <= prevRMS {
			t.Fatalf("shot noise RMS should increase with current: I=%.3e A rms=%.3e prev=%.3e", I, rms, prevRMS)
		}
		prevRMS = rms
	}

	// Check expected scaling: RMS ∝ sqrt(I)
	iLow := 10e-9
	iHigh := 40e-9
	rLow := shotNoiseCurrentRMS(iLow, BW)
	rHigh := shotNoiseCurrentRMS(iHigh, BW)
	wantScale := math.Sqrt(iHigh / iLow)
	gotScale := rHigh / rLow
	if math.Abs(gotScale-wantScale) > 1e-12*wantScale {
		t.Fatalf("shot-noise scaling mismatch: got %.6f, want %.6f", gotScale, wantScale)
	}
}

func TestNoiseModel_TotalNoiseCompositionVarianceAdditivity(t *testing.T) {
	thermalSigma := 12.0e-6
	flickerSigma := 5.0e-6
	shotSigma := 2.0e-6
	quantSigma := math.Sqrt(quantizationNoiseVariance(1.0, 8))

	gotVar := totalNoiseVariance(thermalSigma, flickerSigma, shotSigma, quantSigma)
	wantVar := thermalSigma*thermalSigma + flickerSigma*flickerSigma + shotSigma*shotSigma + quantSigma*quantSigma

	if math.Abs(gotVar-wantVar) > 1e-18 {
		t.Fatalf("total variance mismatch: got %.6e, want %.6e", gotVar, wantVar)
	}

	if gotVar <= thermalSigma*thermalSigma {
		t.Fatalf("composed noise variance should exceed dominant single component: total=%.6e thermal_only=%.6e", gotVar, thermalSigma*thermalSigma)
	}
}

func TestNoiseModel_SNRDegradesGracefullyWithIncreasingNoise(t *testing.T) {
	signal := 100e-6 // 100 uV RMS
	noiseLevels := []float64{1e-6, 2e-6, 4e-6, 8e-6, 16e-6}

	prevSNR := math.Inf(1)
	for i, noise := range noiseLevels {
		snr := snrDB(signal, noise)
		if i > 0 {
			if snr >= prevSNR {
				t.Fatalf("SNR should decrease as noise increases: noise=%.3e snr=%.3f dB prev=%.3f dB", noise, snr, prevSNR)
			}
			// For each 2x noise increase, SNR should drop by ~6.02 dB.
			drop := prevSNR - snr
			if math.Abs(drop-6.02) > 0.05 {
				t.Fatalf("SNR degradation not graceful: drop=%.4f dB, want ~6.02 dB per 2x noise", drop)
			}
		}
		prevSNR = snr
	}
}

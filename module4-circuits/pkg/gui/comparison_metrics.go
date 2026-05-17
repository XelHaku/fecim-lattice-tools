//go:build legacy_fyne

package gui

import (
	"fmt"
	"math"
	"math/rand"
)

type comparisonMetricRow struct {
	Label      string
	LatencyNS  float64
	EnergyPJ   float64
	GOPS       float64
	EnergyOpPJ float64
}

func computeComparisonMetrics(arraySize int) (comparisonMetricRow, comparisonMetricRow, comparisonMetricRow) {
	if arraySize <= 0 {
		arraySize = 8
	}
	macs := float64(arraySize * arraySize)
	scale := macs / 64.0

	cpu := comparisonMetricRow{Label: "CPU", LatencyNS: 500 * scale, EnergyPJ: 64000 * scale}
	gpu := comparisonMetricRow{Label: "GPU", LatencyNS: 50 * scale, EnergyPJ: 6400 * scale}
	fefet := comparisonMetricRow{Label: "FeFET", LatencyNS: 76 * scale, EnergyPJ: 2.9 * scale}

	rows := []*comparisonMetricRow{&cpu, &gpu, &fefet}
	for _, r := range rows {
		// Throughput in GOPS (gigaops/sec), not energy efficiency
		if r.LatencyNS > 0 {
			r.GOPS = (2.0 * macs) / r.LatencyNS / 1e3
		}
		if macs > 0 {
			r.EnergyOpPJ = r.EnergyPJ / macs
		}
	}
	return cpu, gpu, fefet
}

func metricLatency(v float64) string { return fmt.Sprintf("%.0f ns", v) }
func metricEnergy(v float64) string  { return fmt.Sprintf("%.1f pJ", v) }
func metricGOPS(v float64) string    { return fmt.Sprintf("%.3f GOPS", v) }

// DesignSweepPoint is one design-space point for quick Pareto-style exploration.
type DesignSweepPoint struct {
	ArraySize int
	ADCBits   int
	Device    string
	LatencyNS float64
	EnergyPJ  float64
	GOPS      float64
}

// MonteCarloStats contains summary stats for process-variation sampling.
type MonteCarloStats struct {
	Mean   float64
	StdDev float64
	Min    float64
	Max    float64
}

var deviceEnergyScale = map[string]float64{
	"FeFET": 1.0,
	"RRAM":  1.6,
	"PCM":   2.1,
	"SRAM":  6.5,
}

// BuildDesignSpaceSweep returns a lightweight design-space sweep for array size x ADC bits x device.
func BuildDesignSpaceSweep(arraySizes, adcBits []int, devices []string) []DesignSweepPoint {
	out := make([]DesignSweepPoint, 0, len(arraySizes)*len(adcBits)*len(devices))
	for _, n := range arraySizes {
		_, _, fefet := computeComparisonMetrics(n)
		for _, bits := range adcBits {
			if bits < 1 {
				bits = 1
			}
			adcPenalty := 1.0 + 0.08*float64(bits-5)
			if adcPenalty < 0.5 {
				adcPenalty = 0.5
			}
			for _, d := range devices {
				scale, ok := deviceEnergyScale[d]
				if !ok {
					scale = 1.0
				}
				latency := fefet.LatencyNS * adcPenalty
				energy := fefet.EnergyPJ * scale * adcPenalty
				topsw := 0.0
				if energy > 0 {
					macs := float64(n * n)
					topsw = (2.0 * macs) / (energy * 1e-12) / 1e12
				}
				out = append(out, DesignSweepPoint{ArraySize: n, ADCBits: bits, Device: d, LatencyNS: latency, EnergyPJ: energy, GOPS: topsw})
			}
		}
	}
	return out
}

// RunProcessVariationMonteCarlo performs a simple Gaussian variation sampling around a base value.
func RunProcessVariationMonteCarlo(baseValue, sigmaFraction float64, samples int, seed int64) MonteCarloStats {
	if samples < 1 {
		samples = 1
	}
	if sigmaFraction < 0 {
		sigmaFraction = 0
	}
	rng := rand.New(rand.NewSource(seed))
	sigma := baseValue * sigmaFraction
	minV, maxV := math.Inf(1), math.Inf(-1)
	sum := 0.0
	vals := make([]float64, samples)
	for i := 0; i < samples; i++ {
		v := baseValue + rng.NormFloat64()*sigma
		if v < 0 {
			v = 0
		}
		vals[i] = v
		sum += v
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	mean := sum / float64(samples)
	ss := 0.0
	for _, v := range vals {
		d := v - mean
		ss += d * d
	}
	std := math.Sqrt(ss / float64(samples))
	return MonteCarloStats{Mean: mean, StdDev: std, Min: minV, Max: maxV}
}

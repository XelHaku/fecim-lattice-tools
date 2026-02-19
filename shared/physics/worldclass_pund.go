package physics

import (
	"fmt"
	"math"
)

// PulseSample is a current transient sample acquired during one pulse.
type PulseSample struct {
	TimeS    float64 // Seconds
	CurrentA float64 // Amperes
}

// PUNDResult captures integrated charge per pulse and separated switching charge.
type PUNDResult struct {
	QP_C float64 // Program pulse integrated charge
	QU_C float64 // Up pulse integrated charge (non-switching baseline after P)
	QN_C float64 // Negative pulse integrated charge
	QD_C float64 // Down pulse integrated charge (non-switching baseline after N)

	SwitchingPositive_C float64 // QP-QU
	SwitchingNegative_C float64 // QN-QD
}

// IntegrateCurrent integrates current over time with the trapezoidal rule.
func IntegrateCurrent(samples []PulseSample) (float64, error) {
	if len(samples) < 2 {
		return 0, fmt.Errorf("need at least 2 samples, got %d", len(samples))
	}
	q := 0.0
	for i := 1; i < len(samples); i++ {
		dt := samples[i].TimeS - samples[i-1].TimeS
		if dt <= 0 {
			return 0, fmt.Errorf("non-monotonic time at index %d", i)
		}
		q += 0.5 * (samples[i-1].CurrentA + samples[i].CurrentA) * dt
	}
	return q, nil
}

// AnalyzePUND calculates pulse charges and switching components from P/U/N/D traces.
func AnalyzePUND(programP, upU, negativeN, downD []PulseSample) (PUNDResult, error) {
	qP, err := IntegrateCurrent(programP)
	if err != nil {
		return PUNDResult{}, fmt.Errorf("integrate P pulse: %w", err)
	}
	qU, err := IntegrateCurrent(upU)
	if err != nil {
		return PUNDResult{}, fmt.Errorf("integrate U pulse: %w", err)
	}
	qN, err := IntegrateCurrent(negativeN)
	if err != nil {
		return PUNDResult{}, fmt.Errorf("integrate N pulse: %w", err)
	}
	qD, err := IntegrateCurrent(downD)
	if err != nil {
		return PUNDResult{}, fmt.Errorf("integrate D pulse: %w", err)
	}

	return PUNDResult{
		QP_C:                qP,
		QU_C:                qU,
		QN_C:                qN,
		QD_C:                qD,
		SwitchingPositive_C: qP - qU,
		SwitchingNegative_C: qN - qD,
	}, nil
}

// RunPUNDSimulation executes the 6-pulse PUND protocol on a PreisachStack.
// Returns PUNDResult (integrated charges) and pulse traces [P, U, N, D].
// area_m2 converts dP/dt -> current: I = area * dP/dt using finite differences.
// The PreisachStack.Update method is rate-independent (no dt); sampleIntervalS
// is used only for computing dP/dt from successive P values.
func RunPUNDSimulation(stack *PreisachStack, amplitude_Vm, pulseWidthS, sampleIntervalS, area_m2 float64) (PUNDResult, [4][]PulseSample, error) {
	if amplitude_Vm <= 0 {
		return PUNDResult{}, [4][]PulseSample{}, fmt.Errorf("amplitude must be > 0, got %g", amplitude_Vm)
	}
	if pulseWidthS <= 0 || sampleIntervalS <= 0 {
		return PUNDResult{}, [4][]PulseSample{}, fmt.Errorf("pulse width and sample interval must be > 0")
	}
	if area_m2 <= 0 {
		area_m2 = 1e-12 // 1 µm² default
	}

	nSamples := int(math.Ceil(pulseWidthS/sampleIntervalS)) + 1
	if nSamples < 2 {
		nSamples = 2
	}

	// runPulse applies step field E for the pulse duration and collects current samples.
	// Update(E) returns the new polarization P after applying E.
	runPulse := func(E_Vm float64) []PulseSample {
		samples := make([]PulseSample, nSamples)
		prevP := stack.ComputePolarization(stack.LastE)
		for i := 0; i < nSamples; i++ {
			t := float64(i) * sampleIntervalS
			curP := stack.Update(E_Vm)
			dPdt := (curP - prevP) / sampleIntervalS
			samples[i] = PulseSample{TimeS: t, CurrentA: area_m2 * dPdt}
			prevP = curP
		}
		return samples
	}

	// Preset: drive to -amplitude to initialize state
	_ = runPulse(-amplitude_Vm)

	pTrace := runPulse(+amplitude_Vm)
	uTrace := runPulse(+amplitude_Vm)
	nTrace := runPulse(-amplitude_Vm)
	dTrace := runPulse(-amplitude_Vm)

	result, err := AnalyzePUND(pTrace, uTrace, nTrace, dTrace)
	if err != nil {
		return PUNDResult{}, [4][]PulseSample{}, fmt.Errorf("AnalyzePUND: %w", err)
	}

	return result, [4][]PulseSample{pTrace, uTrace, nTrace, dTrace}, nil
}

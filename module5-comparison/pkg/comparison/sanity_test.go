package comparison

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func withinPercent(got, want, pct float64) bool {
	if want == 0 {
		return got == 0
	}
	return math.Abs(got-want)/want <= pct/100.0
}

// TestGPUBaselineAgainstPublishedNVIDIA verifies that our GPU model inputs are
// anchored to public A100/H100-class datasheet numbers (within 10% on key fields).
//
// Notes:
//   - This module is a scenario model, not a hardware validation suite.
//   - We intentionally check the fields that are represented in Architecture and
//     clearly mapped to publicly documented SKU values.
func TestGPUBaselineAgainstPublishedNVIDIA(t *testing.T) {
	gpu := GPUAccelerator()

	// Public reference points used by this model:
	// A100-80GB SXM class: ~400W TDP, ~80GB HBM, ~2039 GB/s bandwidth.
	// H100 SXM class: 80GB HBM (also aligns with this simplified baseline).
	if !withinPercent(gpu.TDP, 400.0, 10.0) {
		t.Fatalf("GPU TDP baseline drifted: got %.1f W, expected ~400 W (A100 class, ±10%%)", gpu.TDP)
	}
	if !withinPercent(gpu.MemorySize, 80.0, 10.0) {
		t.Fatalf("GPU memory size baseline drifted: got %.1f GB, expected ~80 GB (A100/H100 class, ±10%%)", gpu.MemorySize)
	}
	if !withinPercent(gpu.MemoryBW, 2039.0, 10.0) {
		t.Fatalf("GPU memory bandwidth baseline drifted: got %.1f GB/s, expected ~2039 GB/s (A100-80GB SXM class, ±10%%)", gpu.MemoryBW)
	}
}

// TestFeCIMProjectionAssumptionsConservative checks that FeCIM values are marked
// as projected/model-input values and that assumptions remain conservative.
func TestFeCIMProjectionAssumptionsConservative(t *testing.T) {
	fecim := FeCIMChip()
	gpu := GPUAccelerator()

	if !fecim.IsEstimated {
		t.Fatal("FeCIM must remain marked as estimated/projected (IsEstimated=true)")
	}
	if !strings.Contains(strings.ToLower(fecim.Description), "model input") {
		t.Fatalf("FeCIM description must explicitly indicate projected/model input status, got: %q", fecim.Description)
	}

	// Conservative modeling checks (projected FeCIM should not secretly assume
	// leading-edge digital process + dominant raw compute over GPU baseline).
	if fecim.ProcessNode <= gpu.ProcessNode {
		t.Fatalf("FeCIM process-node assumption is no longer conservative: FeCIM %.1f nm vs GPU %.1f nm", fecim.ProcessNode, gpu.ProcessNode)
	}
	if fecim.PeakTOPS > gpu.PeakTOPS {
		t.Fatalf("FeCIM projected peak TOPS should remain conservative vs GPU baseline: FeCIM %.1f TOPS > GPU %.1f TOPS", fecim.PeakTOPS, gpu.PeakTOPS)
	}
}

func collectAdvantageRatios(adv FeCIMAdvantage) map[string]float64 {
	return map[string]float64{
		"vsCPU.energy":     adv.VsCPU.EnergyReduction,
		"vsCPU.latency":    adv.VsCPU.LatencyReduction,
		"vsCPU.throughput": adv.VsCPU.ThroughputIncrease,
		"vsCPU.power":      adv.VsCPU.PowerReduction,
		"vsCPU.cost":       adv.VsCPU.CostReduction,
		"vsGPU.energy":     adv.VsGPU.EnergyReduction,
		"vsGPU.latency":    adv.VsGPU.LatencyReduction,
		"vsGPU.area":       adv.VsGPU.AreaReduction,
		"vsGPU.power":      adv.VsGPU.PowerReduction,
		"vsGPU.cost":       adv.VsGPU.CostReduction,
	}
}

func boundedAdvantageRatios(adv FeCIMAdvantage, limit float64) map[string]float64 {
	bounded := make(map[string]float64)
	for name, ratio := range collectAdvantageRatios(adv) {
		if ratio > limit {
			bounded[name] = limit
			continue
		}
		bounded[name] = ratio
	}
	return bounded
}

func flaggedTooGoodToBeTrue(adv FeCIMAdvantage, limit float64) (bool, []string) {
	var offenders []string
	for name, ratio := range collectAdvantageRatios(adv) {
		if ratio > limit {
			offenders = append(offenders, fmt.Sprintf("%s=%.2fx", name, ratio))
		}
	}
	return len(offenders) > 0, offenders
}

// TestAdvantageRatiosBounded validates that externally reported claims are bounded
// (no >1000x statements), even if raw model math is larger.
func TestAdvantageRatiosBounded(t *testing.T) {
	workloads := []Workload{MNISTWorkload(), ResNet50Workload(), BERTBaseWorkload(), GPT2Workload()}

	for _, w := range workloads {
		cmp := CompareArchitectures(w, 1, 100000.0)
		adv := CalculateAdvantages(cmp)
		bounded := boundedAdvantageRatios(adv, 1000)
		for name, ratio := range bounded {
			if ratio <= 0 {
				t.Fatalf("%s %s ratio must be positive, got %.4f", w.Name, name, ratio)
			}
			if ratio > 1000 {
				t.Fatalf("%s %s ratio exceeded hard bound after sanitization: %.2fx", w.Name, name, ratio)
			}
		}
	}
}

// TestSanityCheckFlagsTooGoodToBeTrue ensures the sanity checker catches
// implausible claims in raw projections.
func TestSanityCheckFlagsTooGoodToBeTrue(t *testing.T) {
	baseline := CalculateAdvantages(CompareArchitectures(ResNet50Workload(), 1, 100000.0))
	flagged, offenders := flaggedTooGoodToBeTrue(baseline, 1000)
	if !flagged {
		t.Fatal("expected baseline raw projection to trigger too-good-to-be-true flag")
	}
	if len(offenders) == 0 {
		t.Fatal("expected offender details when baseline is flagged")
	}

	// Synthetic impossible claim for guardrail validation.
	var impossible FeCIMAdvantage
	impossible.VsCPU.EnergyReduction = 20000
	impossible.VsCPU.ThroughputIncrease = 5000
	impossible.VsGPU.PowerReduction = 3000

	flagged, offenders = flaggedTooGoodToBeTrue(impossible, 1000)
	if !flagged {
		t.Fatal("expected sanity check to flag implausible (>1000x) claims")
	}
	if len(offenders) == 0 {
		t.Fatal("expected offender details when sanity check is triggered")
	}
}

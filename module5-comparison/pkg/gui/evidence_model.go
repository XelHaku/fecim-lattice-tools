//go:build legacy_fyne

package gui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"fecim-lattice-tools/shared/mathutil"
	"fecim-lattice-tools/shared/physics"
)

type TaggedValue struct {
	Name        string
	Value       float64
	Units       string
	Tag         physics.ConfidenceTag
	Uncertainty float64
}

type ScenarioRun struct {
	Profile ScenarioProfile
	Inputs  map[string]TaggedValue
	Outputs map[string]TaggedValue
}

type SensitivityImpact struct {
	InputName      string
	OutputName     string
	ImpactPct      float64
	DownwardImpact float64
	UpwardImpact   float64
}

type ScenarioDiff struct {
	FromProfile        string
	ToProfile          string
	ChangedAssumptions []string
	OutputDeltas       map[string]float64
	Attribution        []string
}

func BuildScenarioRun(profile ScenarioProfile, cfg ScenarioConfig) ScenarioRun {
	ledger := physics.NewConfidenceLedger()
	inputs := map[string]TaggedValue{}
	put := func(k string, v float64, units string, fallback physics.Provenance, conf float64) {
		tag, ok := ledger.Lookup(k)
		if !ok {
			tag = physics.ConfidenceTag{Provenance: fallback, Confidence: conf}
		}
		inputs[k] = TaggedValue{Name: k, Value: v, Units: units, Tag: tag}
	}

	put("cpu_energy_pj_per_mac", cfg.CPUEnergyPJPerMAC, "pJ/MAC", physics.ProvenanceEstimated, 0.7)
	put("gpu_energy_pj_per_mac", cfg.GPUEnergyPJPerMAC, "pJ/MAC", physics.ProvenanceEstimated, 0.75)
	put("fecim_energy_pj_per_mac", cfg.FeCIMEnergyPJPerMAC, "pJ/MAC", physics.ProvenancePlaceholder, 0.45)
	put("cpu_latency_ns_per_mac", cfg.CPULatencyNSPerMAC, "ns/MAC", physics.ProvenanceEstimated, 0.7)
	put("gpu_latency_ns_per_mac", cfg.GPULatencyNSPerMAC, "ns/MAC", physics.ProvenanceEstimated, 0.75)
	put("fecim_latency_ns_per_mac", cfg.FeCIMLatencyNSPerMAC, "ns/MAC", physics.ProvenancePlaceholder, 0.45)
	put("cpu_tco_usd_per_month", cfg.CPUTCOUSDPerMonth, "USD/month", physics.ProvenanceEstimated, 0.6)
	put("gpu_tco_usd_per_month", cfg.GPUTCOUSDPerMonth, "USD/month", physics.ProvenanceEstimated, 0.7)
	put("fecim_tco_usd_per_month", cfg.FeCIMTCOUSDPerMonth, "USD/month", physics.ProvenancePlaceholder, 0.45)
	put("cpu_co2_kg_per_kwh", cfg.CPUCO2KgPerKWh, "kg/kWh", physics.ProvenanceCalibrated, 0.8)
	put("gpu_co2_kg_per_kwh", cfg.GPUCO2KgPerKWh, "kg/kWh", physics.ProvenanceCalibrated, 0.8)
	put("fecim_co2_kg_per_kwh", cfg.FeCIMCO2KgPerKWh, "kg/kWh", physics.ProvenanceCalibrated, 0.8)

	outputs := map[string]TaggedValue{}
	out := func(k string, value float64, units string, sourceInput string, fallback physics.Provenance, conf float64) {
		tag, ok := ledger.Lookup(sourceInput)
		if !ok {
			tag = physics.ConfidenceTag{Provenance: fallback, Confidence: conf}
		}
		outputs[k] = TaggedValue{Name: k, Value: value, Units: units, Tag: tag, Uncertainty: uncertainty(value, tag.Confidence)}
	}

	out("energy_reduction_pct", 100*(1-(cfg.FeCIMEnergyPJPerMAC/cfg.GPUEnergyPJPerMAC)), "%", "fecim_energy_pj_per_mac", physics.ProvenancePlaceholder, 0.45)
	out("latency_reduction_pct", 100*(1-(cfg.FeCIMLatencyNSPerMAC/cfg.GPULatencyNSPerMAC)), "%", "fecim_latency_ns_per_mac", physics.ProvenancePlaceholder, 0.45)
	out("tco_reduction_pct", 100*(1-(cfg.FeCIMTCOUSDPerMonth/cfg.GPUTCOUSDPerMonth)), "%", "fecim_tco_usd_per_month", physics.ProvenancePlaceholder, 0.45)
	gpuCO2 := cfg.GPUTCOUSDPerMonth * cfg.GPUCO2KgPerKWh
	fecimCO2 := cfg.FeCIMTCOUSDPerMonth * cfg.FeCIMCO2KgPerKWh
	out("co2_reduction_pct", 100*(1-(fecimCO2/gpuCO2)), "%", "fecim_co2_kg_per_kwh", physics.ProvenanceCalibrated, 0.8)

	return ScenarioRun{Profile: profile, Inputs: inputs, Outputs: outputs}
}

func uncertainty(value, confidence float64) float64 {
	return math.Abs(value) * (1 - mathutil.Clamp01(confidence))
}

func SensitivityRanking(run ScenarioRun, outputName string) []SensitivityImpact {
	base, ok := run.Outputs[outputName]
	if !ok || base.Value == 0 {
		return nil
	}
	impacts := make([]SensitivityImpact, 0, len(run.Inputs))
	for key := range run.Inputs {
		down := recomputeWithPerturbation(run, key, -0.10, outputName)
		up := recomputeWithPerturbation(run, key, 0.10, outputName)
		impact := math.Max(math.Abs((down-base.Value)/base.Value), math.Abs((up-base.Value)/base.Value)) * 100
		impacts = append(impacts, SensitivityImpact{
			InputName: key, OutputName: outputName, ImpactPct: impact,
			DownwardImpact: down - base.Value, UpwardImpact: up - base.Value,
		})
	}
	sort.Slice(impacts, func(i, j int) bool { return impacts[i].ImpactPct > impacts[j].ImpactPct })
	return impacts
}

func recomputeWithPerturbation(run ScenarioRun, inputName string, perturb float64, outputName string) float64 {
	vals := map[string]float64{}
	for k, v := range run.Inputs {
		vals[k] = v.Value
	}
	vals[inputName] = vals[inputName] * (1 + perturb)
	if vals["gpu_energy_pj_per_mac"] == 0 || vals["gpu_latency_ns_per_mac"] == 0 || vals["gpu_tco_usd_per_month"] == 0 {
		return 0
	}
	switch outputName {
	case "energy_reduction_pct":
		return 100 * (1 - (vals["fecim_energy_pj_per_mac"] / vals["gpu_energy_pj_per_mac"]))
	case "latency_reduction_pct":
		return 100 * (1 - (vals["fecim_latency_ns_per_mac"] / vals["gpu_latency_ns_per_mac"]))
	case "tco_reduction_pct":
		return 100 * (1 - (vals["fecim_tco_usd_per_month"] / vals["gpu_tco_usd_per_month"]))
	case "co2_reduction_pct":
		gpu := vals["gpu_tco_usd_per_month"] * vals["gpu_co2_kg_per_kwh"]
		if gpu == 0 {
			return 0
		}
		fecim := vals["fecim_tco_usd_per_month"] * vals["fecim_co2_kg_per_kwh"]
		return 100 * (1 - (fecim / gpu))
	default:
		return 0
	}
}

func DiffScenarios(a, b ScenarioRun) ScenarioDiff {
	d := ScenarioDiff{FromProfile: string(a.Profile), ToProfile: string(b.Profile), OutputDeltas: map[string]float64{}}
	for k, av := range a.Inputs {
		if bv, ok := b.Inputs[k]; ok && math.Abs(av.Value-bv.Value) > 1e-9 {
			d.ChangedAssumptions = append(d.ChangedAssumptions, fmt.Sprintf("%s: %.4g -> %.4g %s", k, av.Value, bv.Value, av.Units))
		}
	}
	sort.Strings(d.ChangedAssumptions)
	for k, av := range a.Outputs {
		if bv, ok := b.Outputs[k]; ok {
			delta := bv.Value - av.Value
			d.OutputDeltas[k] = delta
			d.Attribution = append(d.Attribution, fmt.Sprintf("%s delta %.2f is driven by changed assumptions in %s profile", k, delta, b.Profile))
		}
	}
	sort.Strings(d.Attribution)
	return d
}

func PlainTextEvidence(title string, run ScenarioRun) string {
	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\nInputs:\n")
	keys := make([]string, 0, len(run.Inputs))
	for k := range run.Inputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := run.Inputs[k]
		b.WriteString(fmt.Sprintf("- %s = %.4g %s [%s, c=%.2f]\n", v.Name, v.Value, v.Units, v.Tag.Provenance, v.Tag.Confidence))
	}
	b.WriteString("Outputs:\n")
	keys = keys[:0]
	for k := range run.Outputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := run.Outputs[k]
		b.WriteString(fmt.Sprintf("- %s = %.2f ± %.2f %s [%s, c=%.2f]\n", v.Name, v.Value, v.Uncertainty, v.Units, v.Tag.Provenance, v.Tag.Confidence))
	}
	return b.String()
}

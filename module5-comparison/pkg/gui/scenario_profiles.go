//go:build legacy_fyne

package gui

import "fmt"

// ScenarioProfile identifies a bundled assumption profile.
type ScenarioProfile string

const (
	ScenarioConservative ScenarioProfile = "conservative"
	ScenarioBaseline     ScenarioProfile = "baseline"
	ScenarioOptimistic   ScenarioProfile = "optimistic"
)

// ScenarioConfig contains module 5 assumptions for CPU/GPU/FeCIM comparison.
type ScenarioConfig struct {
	Name string

	CPUEnergyPJPerMAC   float64
	GPUEnergyPJPerMAC   float64
	FeCIMEnergyPJPerMAC float64

	CPULatencyNSPerMAC   float64
	GPULatencyNSPerMAC   float64
	FeCIMLatencyNSPerMAC float64

	CPUTCOUSDPerMonth   float64
	GPUTCOUSDPerMonth   float64
	FeCIMTCOUSDPerMonth float64

	CPUCO2KgPerKWh   float64
	GPUCO2KgPerKWh   float64
	FeCIMCO2KgPerKWh float64
}

// PresetScenarioConfig returns one of the standard module 5 scenario presets.
func PresetScenarioConfig(profile ScenarioProfile) ScenarioConfig {
	switch profile {
	case ScenarioConservative:
		return ScenarioConfig{
			Name:              "Conservative",
			CPUEnergyPJPerMAC: 1000, GPUEnergyPJPerMAC: 120, FeCIMEnergyPJPerMAC: 2.0,
			CPULatencyNSPerMAC: 3.5, GPULatencyNSPerMAC: 0.7, FeCIMLatencyNSPerMAC: 0.2,
			CPUTCOUSDPerMonth: 350, GPUTCOUSDPerMonth: 950, FeCIMTCOUSDPerMonth: 260,
			CPUCO2KgPerKWh: 0.41, GPUCO2KgPerKWh: 0.41, FeCIMCO2KgPerKWh: 0.41,
		}
	case ScenarioOptimistic:
		return ScenarioConfig{
			Name:              "Optimistic",
			CPUEnergyPJPerMAC: 1000, GPUEnergyPJPerMAC: 80, FeCIMEnergyPJPerMAC: 0.5,
			CPULatencyNSPerMAC: 3.0, GPULatencyNSPerMAC: 0.5, FeCIMLatencyNSPerMAC: 0.06,
			CPUTCOUSDPerMonth: 330, GPUTCOUSDPerMonth: 900, FeCIMTCOUSDPerMonth: 140,
			CPUCO2KgPerKWh: 0.38, GPUCO2KgPerKWh: 0.38, FeCIMCO2KgPerKWh: 0.38,
		}
	default:
		return ScenarioConfig{
			Name:              "Baseline",
			CPUEnergyPJPerMAC: 1000, GPUEnergyPJPerMAC: 100, FeCIMEnergyPJPerMAC: 1.0,
			CPULatencyNSPerMAC: 3.2, GPULatencyNSPerMAC: 0.6, FeCIMLatencyNSPerMAC: 0.1,
			CPUTCOUSDPerMonth: 340, GPUTCOUSDPerMonth: 920, FeCIMTCOUSDPerMonth: 190,
			CPUCO2KgPerKWh: 0.40, GPUCO2KgPerKWh: 0.40, FeCIMCO2KgPerKWh: 0.40,
		}
	}
}

func (p ScenarioProfile) String() string { return string(p) }

func ScenarioProfileFromString(s string) ScenarioProfile {
	switch s {
	case string(ScenarioConservative):
		return ScenarioConservative
	case string(ScenarioOptimistic):
		return ScenarioOptimistic
	default:
		return ScenarioBaseline
	}
}

func (c ScenarioConfig) Validate() error {
	if c.CPUEnergyPJPerMAC <= 0 || c.GPUEnergyPJPerMAC <= 0 || c.FeCIMEnergyPJPerMAC <= 0 {
		return fmt.Errorf("energy assumptions must be > 0")
	}
	return nil
}

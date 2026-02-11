package peripherals

import (
	"math"
)

// ChargePump represents a charge pump circuit for voltage boosting.
// Used to generate write voltages (±1.5V) from standard CMOS supply (1V).
type ChargePump struct {
	InputVoltage   float64 // Supply voltage (V)
	OutputVoltage  float64 // Target output voltage (V)
	Stages         int     // Number of pump stages
	DiodeDrop      float64 // Effective diode/switch drop per stage (V)
	ClockFrequency float64 // Pump clock frequency (Hz)
	LoadCurrent    float64 // Maximum load current (A)
	FlyCapacitance float64 // Flying capacitor value (F)
	Efficiency     float64 // Power conversion efficiency
}

// DefaultChargePump returns a charge pump for FeCIM write operations.
func DefaultChargePump() *ChargePump {
	cp := &ChargePump{
		InputVoltage:   1.0,     // 1V CMOS supply
		OutputVoltage:  1.5,     // 1.5V write voltage
		Stages:         2,       // 2-stage Dickson pump
		DiodeDrop:      0.3,     // ~0.3V per stage for MOS switches
		ClockFrequency: 50e6,    // 50 MHz clock
		LoadCurrent:    10e-6,   // 10 µA load
		FlyCapacitance: 100e-12, // 100 pF flying caps
		Efficiency:     0.7,     // 70% efficiency
	}
	log.Calculation("DefaultChargePump", map[string]interface{}{
		"input_voltage":   cp.InputVoltage,
		"output_voltage":  cp.OutputVoltage,
		"stages":          cp.Stages,
		"diode_drop":      cp.DiodeDrop,
		"clock_frequency": cp.ClockFrequency,
		"load_current":    cp.LoadCurrent,
		"fly_capacitance": cp.FlyCapacitance,
		"efficiency":      cp.Efficiency,
	}, cp)
	return cp
}

// IdealOutputVoltage returns theoretical maximum output.
func (c *ChargePump) IdealOutputVoltage() float64 {
	// Dickson pump: Vout = (N+1) * Vclk - N * Vth
	// For ideal case: Vout = (N+1) * Vin
	sign := 1.0
	if c.OutputVoltage < 0 {
		sign = -1.0
	}
	return sign * float64(c.Stages+1) * c.InputVoltage
}

// ActualOutputVoltage returns output considering losses.
func (c *ChargePump) ActualOutputVoltage() float64 {
	log.Input("ChargePump.ActualOutputVoltage", map[string]interface{}{
		"stages":          c.Stages,
		"load_current":    c.LoadCurrent,
		"fly_capacitance": c.FlyCapacitance,
		"clock_frequency": c.ClockFrequency,
	})

	// Account for diode drops and effective output resistance.
	vthDrop := c.DiodeDrop * float64(c.Stages)
	outputResistance := 0.0
	if c.FlyCapacitance > 0 && c.ClockFrequency > 0 {
		outputResistance = 1.0 / (c.FlyCapacitance * c.ClockFrequency)
	}
	irDrop := math.Abs(c.LoadCurrent) * outputResistance
	idealVoltage := c.IdealOutputVoltage()

	// Compute magnitude after drops, then apply sign of ideal voltage.
	idealMag := math.Abs(idealVoltage)
	actualMag := idealMag - vthDrop - irDrop
	if actualMag < 0 {
		actualMag = 0
	}

	actualVoltage := math.Copysign(actualMag, idealVoltage)

	// If a target output is set, clamp to that regulation limit.
	if c.OutputVoltage != 0 && math.Abs(actualVoltage) > math.Abs(c.OutputVoltage) {
		actualVoltage = math.Copysign(math.Abs(c.OutputVoltage), c.OutputVoltage)
	}

	log.Calculation("ChargePump.ActualOutputVoltage", map[string]interface{}{
		"ideal_voltage": idealVoltage,
		"vth_drop":      vthDrop,
		"ir_drop":       irDrop,
	}, actualVoltage)

	return actualVoltage
}

// OutputRipple estimates peak-to-peak ripple voltage.
func (c *ChargePump) OutputRipple() float64 {
	// ΔV = Iload / (Cout * f)
	// Assume output cap = 10x flying cap
	cOut := c.FlyCapacitance * 10
	if cOut <= 0 || c.ClockFrequency <= 0 {
		return 0
	}
	return math.Abs(c.LoadCurrent) / (cOut * c.ClockFrequency)
}

// BoostFactor returns voltage multiplication factor.
func (c *ChargePump) BoostFactor() float64 {
	if c.InputVoltage == 0 {
		return 0
	}
	return c.ActualOutputVoltage() / c.InputVoltage
}

// PowerInput returns input power consumption.
func (c *ChargePump) PowerInput() float64 {
	// Pin = Pout / efficiency
	pOut := c.PowerOutput()
	if c.Efficiency <= 0 {
		return math.Inf(1)
	}
	return pOut / c.Efficiency
}

// PowerOutput returns delivered output power.
// Uses the *actual* achievable output voltage (after losses/regulation clamp).
func (c *ChargePump) PowerOutput() float64 {
	vOut := math.Abs(c.ActualOutputVoltage())
	return vOut * math.Abs(c.LoadCurrent)
}

// PowerLoss returns power dissipated in the pump.
func (c *ChargePump) PowerLoss() float64 {
	return c.PowerInput() - c.PowerOutput()
}

// RiseTime estimates output voltage rise time from 10% to 90%.
func (c *ChargePump) RiseTime() float64 {
	// Simplified: depends on clock frequency and stages
	// t_rise ≈ (Stages * 2.2) / f_clk
	return float64(c.Stages) * 2.2 / c.ClockFrequency
}

// MaxCurrentCapability returns maximum sustainable output current.
func (c *ChargePump) MaxCurrentCapability() float64 {
	// I_max = C * f * (N+1) * Vin / Vout
	if c.OutputVoltage == 0 {
		return 0
	}
	return c.FlyCapacitance * c.ClockFrequency * float64(c.Stages+1) * c.InputVoltage / math.Abs(c.OutputVoltage)
}

// EnergyPerOperation estimates energy for one write voltage pulse.
func (c *ChargePump) EnergyPerOperation(pulseDuration float64) float64 {
	log.Input("ChargePump.EnergyPerOperation", map[string]interface{}{
		"pulse_duration": pulseDuration,
	})

	// E = P * t
	power := c.PowerInput()
	energy := power * pulseDuration

	log.Calculation("ChargePump.EnergyPerOperation", map[string]interface{}{
		"power":          power,
		"pulse_duration": pulseDuration,
	}, energy)

	return energy
}

// NegativePump creates a negative voltage charge pump configuration.
func NegativePump() *ChargePump {
	cp := DefaultChargePump()
	cp.OutputVoltage = -1.5 // Negative write voltage
	cp.Stages = 2           // 2-stage negative pump
	return cp
}

// ChargeTransferEfficiency calculates per-stage efficiency.
func (c *ChargePump) ChargeTransferEfficiency() float64 {
	// η_stage = Vout / (Vin * (N+1))
	// Accounting for all stages
	ideal := c.IdealOutputVoltage()
	if ideal == 0 {
		return 0
	}
	return math.Abs(c.ActualOutputVoltage() / ideal)
}

// Area estimates silicon area (very rough).
func (c *ChargePump) Area() float64 {
	// Area dominated by capacitors
	// Typical: ~0.1 fF/µm² for MIM caps in 65nm
	capDensity := 0.1e-15 / 1e-12                        // F/µm²
	totalCap := c.FlyCapacitance * float64(c.Stages) * 2 // Fly + output caps
	return totalCap / capDensity                         // µm²
}

// SupportsLevel checks if pump can generate voltage for a given level.
func (c *ChargePump) SupportsLevel(level int, maxLevel int) bool {
	log.Input("ChargePump.SupportsLevel", map[string]interface{}{
		"level":          level,
		"max_level":      maxLevel,
		"output_voltage": c.OutputVoltage,
	})

	// Calculate required voltage for this level
	requiredV := c.OutputVoltage * float64(level) / float64(maxLevel)
	actualV := c.ActualOutputVoltage()
	supported := math.Abs(requiredV) <= math.Abs(actualV)

	log.Calculation("ChargePump.SupportsLevel", map[string]interface{}{
		"level":       level,
		"required_v":  requiredV,
		"actual_v":    actualV,
	}, supported)

	return supported
}

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
	ClockFrequency float64 // Pump clock frequency (Hz)
	LoadCurrent    float64 // Maximum load current (A)
	FlyCapacitance float64 // Flying capacitor value (F)
	Efficiency     float64 // Power conversion efficiency
}

// DefaultChargePump returns a charge pump for FeCIM write operations.
func DefaultChargePump() *ChargePump {
	return &ChargePump{
		InputVoltage:   1.0,     // 1V CMOS supply
		OutputVoltage:  1.5,     // 1.5V write voltage
		Stages:         2,       // 2-stage Dickson pump
		ClockFrequency: 50e6,    // 50 MHz clock
		LoadCurrent:    10e-6,   // 10 µA load
		FlyCapacitance: 100e-12, // 100 pF flying caps
		Efficiency:     0.7,     // 70% efficiency
	}
}

// IdealOutputVoltage returns theoretical maximum output.
func (c *ChargePump) IdealOutputVoltage() float64 {
	// Dickson pump: Vout = (N+1) * Vclk - N * Vth
	// For ideal case: Vout = (N+1) * Vin
	return float64(c.Stages+1) * c.InputVoltage
}

// ActualOutputVoltage returns output considering losses.
func (c *ChargePump) ActualOutputVoltage() float64 {
	// Account for diode drops, IR drops, etc.
	vthDrop := 0.3 * float64(c.Stages) // ~0.3V per stage for MOS switches
	irDrop := c.LoadCurrent / (c.FlyCapacitance * c.ClockFrequency)

	return c.IdealOutputVoltage() - vthDrop - irDrop
}

// OutputRipple estimates peak-to-peak ripple voltage.
func (c *ChargePump) OutputRipple() float64 {
	// ΔV = Iload / (Cout * f)
	// Assume output cap = 10x flying cap
	cOut := c.FlyCapacitance * 10
	return c.LoadCurrent / (cOut * c.ClockFrequency)
}

// BoostFactor returns voltage multiplication factor.
func (c *ChargePump) BoostFactor() float64 {
	return c.ActualOutputVoltage() / c.InputVoltage
}

// PowerInput returns input power consumption.
func (c *ChargePump) PowerInput() float64 {
	// Pin = Pout / efficiency
	pOut := c.OutputVoltage * c.LoadCurrent
	return pOut / c.Efficiency
}

// PowerOutput returns delivered output power.
func (c *ChargePump) PowerOutput() float64 {
	return c.OutputVoltage * c.LoadCurrent
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
	return c.FlyCapacitance * c.ClockFrequency * float64(c.Stages+1) * c.InputVoltage / c.OutputVoltage
}

// EnergyPerOperation estimates energy for one write voltage pulse.
func (c *ChargePump) EnergyPerOperation(pulseDuration float64) float64 {
	// E = P * t
	return c.PowerInput() * pulseDuration
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
	return c.ActualOutputVoltage() / c.IdealOutputVoltage()
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
	// Calculate required voltage for this level
	requiredV := c.OutputVoltage * float64(level) / float64(maxLevel)
	return math.Abs(requiredV) <= math.Abs(c.ActualOutputVoltage())
}

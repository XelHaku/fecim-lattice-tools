package system

// PowerModel estimates the static and dynamic power consumption of a crossbar array.
//
// The model separates two contributions:
//   - Leakage power: sub-threshold and gate-oxide leakage through all cells
//   - Switching power: dynamic capacitive switching at the given clock frequency
//
// All power values are returned in microwatts (µW).
type PowerModel struct {
	Rows           int
	Cols           int
	SupplyV        float64 // supply voltage (V)
	LeakageCurrent float64 // leakage current per cell (A)
}

// NewPowerModel creates a PowerModel for the given array dimensions,
// supply voltage (V), and per-cell leakage current (A).
func NewPowerModel(rows, cols int, supplyV, leakageCurrent float64) *PowerModel {
	return &PowerModel{
		Rows:           rows,
		Cols:           cols,
		SupplyV:        supplyV,
		LeakageCurrent: leakageCurrent,
	}
}

// LeakagePowerUW returns the total static leakage power in µW:
//
//	P_leak = Rows × Cols × LeakageCurrent × SupplyV   (converted to µW)
func (p *PowerModel) LeakagePowerUW() float64 {
	cells := float64(p.Rows * p.Cols)
	watts := cells * p.LeakageCurrent * p.SupplyV
	return watts * 1e6 // W → µW
}

// SwitchingPowerUW returns the estimated dynamic switching power in µW.
//
// Uses the standard CMOS dynamic power formula:
//
//	P_dyn = α × C_load × V² × f
//
// where α = 0.5 (average activity factor), C_load is estimated as 10 fF per cell,
// and f is the operating frequency in MHz.
func (p *PowerModel) SwitchingPowerUW(freqMHz float64) float64 {
	const alpha = 0.5     // average switching activity factor
	const cPerCellF = 10e-15 // 10 fF per cell (parasitic + gate capacitance)

	cells := float64(p.Rows * p.Cols)
	freqHz := freqMHz * 1e6
	watts := alpha * cells * cPerCellF * p.SupplyV * p.SupplyV * freqHz
	return watts * 1e6 // W → µW
}

// TotalPowerUW returns the sum of leakage and switching power in µW:
//
//	LeakagePowerUW() + SwitchingPowerUW(freqMHz)
func (p *PowerModel) TotalPowerUW(freqMHz float64) float64 {
	return p.LeakagePowerUW() + p.SwitchingPowerUW(freqMHz)
}

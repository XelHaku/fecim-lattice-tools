package system


// LatencyModel estimates the pipeline latency for one MVM cycle on a crossbar array.
//
// The model accounts for three serial stages:
//  1. DAC: convert digital inputs to analog voltage levels
//  2. Crossbar settling: RC time constant for the passive crossbar
//  3. ADC: convert analog output currents to digital values
//
// All latencies are returned in nanoseconds (ns).
// Values are empirical estimates for CMOS at the given technology node.
type LatencyModel struct {
	Rows     int
	Cols     int
	TechNode TechnologyNode
}

// NewLatencyModel creates a LatencyModel for the given array dimensions and node.
func NewLatencyModel(rows, cols int, node TechnologyNode) *LatencyModel {
	return &LatencyModel{Rows: rows, Cols: cols, TechNode: node}
}

// nodeSpeedFactor returns a relative speed multiplier for the technology node.
// Faster (smaller) nodes have shorter gate delays and lower RC products.
// Normalised to 1.0 at 65 nm.
func (l *LatencyModel) nodeSpeedFactor() float64 {
	switch l.TechNode {
	case Node130nm:
		return 2.0
	case Node65nm:
		return 1.0
	case Node28nm:
		return 0.5
	case Node22nm:
		return 0.4
	case Node14nm:
		return 0.25
	default:
		return 1.0
	}
}

// DACLatencyNS returns the estimated DAC conversion latency in ns.
//
// Baseline: 2 ns for a 4-bit R-2R ladder DAC at 65 nm.
// Scales linearly with the node speed factor (faster node → shorter latency).
func (l *LatencyModel) DACLatencyNS() float64 {
	const baselineNS = 2.0 // 4-bit DAC at 65 nm
	return baselineNS * l.nodeSpeedFactor()
}

// CrossbarSettlingNS returns the estimated RC settling time of the passive crossbar
// in ns.
//
// The dominant time constant is approximated as:
//
//	τ ≈ R_wire × C_parasitic × max(Rows, Cols)
//
// Baseline: 1 ns per 64 cells at 65 nm; scales with array size and node.
func (l *LatencyModel) CrossbarSettlingNS() float64 {
	const baselineNSPer64 = 1.0
	size := float64(max(l.Rows, l.Cols))
	return baselineNSPer64 * (size / 64.0) * l.nodeSpeedFactor()
}

// ADCLatencyNS returns the estimated ADC conversion latency in ns for adcBits bits.
//
// Uses a SAR ADC model: latency = adcBits × T_comparator.
// Baseline: 1 comparator cycle = 1 ns at 65 nm; total = bits × 1 ns.
// Scales with node speed factor.
func (l *LatencyModel) ADCLatencyNS(bits int) float64 {
	if bits <= 0 {
		bits = 4
	}
	const comparatorNS = 1.0 // one comparator cycle at 65 nm
	return float64(bits) * comparatorNS * l.nodeSpeedFactor()
}

// TotalPipelineNS returns the total MVM pipeline latency in ns:
//
//	DACLatencyNS() + CrossbarSettlingNS() + ADCLatencyNS(adcBits)
func (l *LatencyModel) TotalPipelineNS(adcBits int) float64 {
	return l.DACLatencyNS() + l.CrossbarSettlingNS() + l.ADCLatencyNS(adcBits)
}

// max returns the larger of two ints (local helper, avoids importing slices).
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}


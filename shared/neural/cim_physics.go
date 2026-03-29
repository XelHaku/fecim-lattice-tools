package neural

import "math"

const (
	cimGMinS = 1e-6
	cimGMaxS = 100e-6
)

// CIMNoiseComponents decomposes total CIM noise into physical contributors.
type CIMNoiseComponents struct {
	ADC           float64 // quantization + comparator noise
	Thermal       float64 // kT/C and resistor thermal noise
	Flicker       float64 // 1/f low-frequency noise
	CellVariation float64 // device-to-device and cycle variation
}

func (n CIMNoiseComponents) TotalSigma() float64 {
	return math.Sqrt(n.ADC*n.ADC + n.Thermal*n.Thermal + n.Flicker*n.Flicker + n.CellVariation*n.CellVariation)
}

func defaultNoiseComponents(total float64) CIMNoiseComponents {
	c := 0.5 * total // quadrature of 4 equal terms => total
	return CIMNoiseComponents{ADC: c, Thermal: c, Flicker: c, CellVariation: c}
}

func applyNoiseComponent(values []float64, sigma float64, rng *RandomSource) []float64 {
	if sigma <= 0 {
		return values
	}
	for i := range values {
		values[i] += rng.NormFloat64() * sigma
	}
	return values
}

func applyDecomposedNoise(values []float64, c CIMNoiseComponents, rng *RandomSource) []float64 {
	out := values
	out = applyNoiseComponent(out, c.ADC, rng)
	out = applyNoiseComponent(out, c.Thermal, rng)
	out = applyNoiseComponent(out, c.Flicker, rng)
	out = applyNoiseComponent(out, c.CellVariation, rng)
	return out
}

// TIAModel captures first-order transimpedance and GBW-limited response.
type TIAModel struct {
	RfOhm         float64
	CfF           float64
	GBWHz         float64
	InputNoiseRMS float64
}

func (m TIAModel) poleHz() float64 {
	if m.RfOhm <= 0 {
		return math.Inf(1)
	}
	fc := math.Inf(1)
	if m.CfF > 0 {
		fc = 1.0 / (2 * math.Pi * m.RfOhm * m.CfF)
	}
	if m.GBWHz > 0 && m.GBWHz < fc {
		fc = m.GBWHz
	}
	return fc
}

func (m TIAModel) TransimpedanceMag(freqHz float64) float64 {
	if m.RfOhm <= 0 {
		return 0
	}
	fc := m.poleHz()
	if !math.IsInf(fc, 1) && fc > 0 {
		return m.RfOhm / math.Sqrt(1+math.Pow(freqHz/fc, 2))
	}
	return m.RfOhm
}

func (m TIAModel) CurrentToVoltage(currentA, freqHz float64) float64 {
	return currentA*m.TransimpedanceMag(freqHz) + m.InputNoiseRMS
}

func adcTheoreticalSNRdB(bits int) float64 {
	return 6.02*float64(bits) + 1.76
}

func validateADCSNR(bits int, measuredSNRdB float64, tolDB float64) bool {
	return math.Abs(measuredSNRdB-adcTheoreticalSNRdB(bits)) <= tolDB
}

func (net *DualModeNetwork) cimNoiseComponentsLocked() CIMNoiseComponents {
	c := CIMNoiseComponents{
		ADC:           net.Config.NoiseADC,
		Thermal:       net.Config.NoiseThermal,
		Flicker:       net.Config.NoiseFlicker,
		CellVariation: net.Config.NoiseCellVariation,
	}
	if c.TotalSigma() == 0 && net.Config.NoiseLevel > 0 {
		c = defaultNoiseComponents(net.Config.NoiseLevel)
	}
	return c
}

// forwardCIMConductance computes differential-pair conductance-domain MVM.
func (net *DualModeNetwork) forwardCIMConductance(input []float64, weights [][]float64, bias []float64) []float64 {
	out := make([]float64, len(bias))
	net.forwardCIMConductanceInto(input, weights, bias, out)
	return out
}

func (net *DualModeNetwork) forwardCIMConductanceInto(input []float64, weights [][]float64, bias []float64, out []float64) {
	if len(weights) == 0 || len(weights[0]) == 0 {
		copy(out, bias)
		return
	}

	wMax := 0.0
	for i := range weights {
		for j := range weights[i] {
			a := math.Abs(weights[i][j])
			if a > wMax {
				wMax = a
			}
		}
	}
	if wMax == 0 {
		copy(out, bias)
		return
	}

	gSpan := cimGMaxS - cimGMinS
	for i := range weights {
		sum := bias[i]
		for j := range input {
			w := weights[i][j]
			n := (w/wMax + 1.0) * 0.5
			if n < 0 {
				n = 0
			}
			if n > 1 {
				n = 1
			}
			gp := cimGMinS + n*gSpan
			gn := cimGMinS + (1.0-n)*gSpan
			weff := ((gp - gn) / gSpan) * wMax
			sum += weff * input[j]
		}
		out[i] = sum
	}
}

func (net *DualModeNetwork) adcReadLatencySecondsLocked(rows int) float64 {
	if net.Config.ADCConversionTimeS <= 0 || rows <= 0 {
		return 0
	}
	parallel := net.Config.ADCParallelism
	if parallel <= 0 {
		parallel = 1
	}
	reads := float64((rows + parallel - 1) / parallel)
	return reads * net.Config.ADCConversionTimeS
}

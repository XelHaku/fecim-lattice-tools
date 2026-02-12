// Package peripherals provides peripheral circuit models for ferroelectric CIM systems.
package peripherals

import (
	"math"
)

// INLDNLAnalysis contains INL/DNL analysis results for a DAC or ADC.
type INLDNLAnalysis struct {
	Levels    int       // Number of levels
	INLValues []float64 // INL at each code (in LSB)
	DNLValues []float64 // DNL at each code (in LSB)
	MaxINL    float64   // Maximum INL
	MaxDNL    float64   // Maximum DNL
	MinDNL    float64   // Minimum DNL
	WorstCode int       // Code with worst INL
}

// PVTINLDNLAnalysis captures INL/DNL behavior across temperature and process corner.
type PVTINLDNLAnalysis struct {
	TemperatureK float64
	Corner       ProcessCorner
	INLScale     float64
	DNLScale     float64
	DAC          *INLDNLAnalysis
	ADC          *INLDNLAnalysis
}

// AnalyzeINLDNL computes detailed INL/DNL for a DAC.
func (d *DAC) AnalyzeINLDNL() *INLDNLAnalysis {
	log.Input("DAC.AnalyzeINLDNL", map[string]interface{}{
		"bits": d.Bits,
		"inl":  d.INL,
		"dnl":  d.DNL,
	})

	levels := d.Levels()
	if levels > 32 {
		levels = 32 // Limit for FeCIM 30 levels
	}

	analysis := &INLDNLAnalysis{
		Levels:    levels,
		INLValues: make([]float64, levels),
		DNLValues: make([]float64, levels),
	}

	lsb := d.Resolution()

	// Calculate INL and DNL for each code
	for i := 0; i < levels; i++ {
		idealVoltage := d.Convert(i)
		actualVoltage := d.ConvertWithNonlinearity(i)

		// INL: deviation from ideal straight line (in LSB)
		analysis.INLValues[i] = (actualVoltage - idealVoltage) / lsb

		if math.Abs(analysis.INLValues[i]) > math.Abs(analysis.MaxINL) {
			analysis.MaxINL = analysis.INLValues[i]
			analysis.WorstCode = i
		}

		// DNL: deviation of step size from ideal (in LSB)
		if i > 0 {
			idealStep := lsb
			actualStep := d.ConvertWithNonlinearity(i) - d.ConvertWithNonlinearity(i-1)
			analysis.DNLValues[i] = (actualStep - idealStep) / lsb

			if analysis.DNLValues[i] > analysis.MaxDNL {
				analysis.MaxDNL = analysis.DNLValues[i]
			}
			if analysis.DNLValues[i] < analysis.MinDNL {
				analysis.MinDNL = analysis.DNLValues[i]
			}
		}
	}

	log.Calculation("DAC.AnalyzeINLDNL", map[string]interface{}{
		"levels":     levels,
		"max_inl":    analysis.MaxINL,
		"max_dnl":    analysis.MaxDNL,
		"worst_code": analysis.WorstCode,
	}, analysis)

	return analysis
}

// AnalyzeINLDNL computes detailed INL/DNL for an ADC.
func (a *ADC) AnalyzeINLDNL() *INLDNLAnalysis {
	log.Input("ADC.AnalyzeINLDNL", map[string]interface{}{
		"bits": a.Bits,
		"inl":  a.INL,
		"dnl":  a.DNL,
	})

	levels := a.Levels()
	if levels > 32 {
		levels = 32
	}

	analysis := &INLDNLAnalysis{
		Levels:    levels,
		INLValues: make([]float64, levels),
		DNLValues: make([]float64, levels),
	}

	lsb := a.Resolution()

	// Sweep voltage and measure code transitions
	for i := 0; i < levels; i++ {
		// Ideal voltage for this code
		idealVoltage := a.VrefLow + float64(i)*lsb

		// Find actual transition point
		actualCode := a.ConvertWithNonlinearity(idealVoltage)

		// INL: difference from ideal code (in LSB)
		analysis.INLValues[i] = float64(actualCode-i) / 1.0

		if math.Abs(analysis.INLValues[i]) > math.Abs(analysis.MaxINL) {
			analysis.MaxINL = analysis.INLValues[i]
			analysis.WorstCode = i
		}

		// DNL: deviation of code width from ideal
		if i > 0 {
			// Measure code bin width
			codeWidth := findCodeWidth(a, i, lsb)
			analysis.DNLValues[i] = (codeWidth - lsb) / lsb

			if analysis.DNLValues[i] > analysis.MaxDNL {
				analysis.MaxDNL = analysis.DNLValues[i]
			}
			if analysis.DNLValues[i] < analysis.MinDNL {
				analysis.MinDNL = analysis.DNLValues[i]
			}
		}
	}

	log.Calculation("ADC.AnalyzeINLDNL", map[string]interface{}{
		"levels":     levels,
		"max_inl":    analysis.MaxINL,
		"max_dnl":    analysis.MaxDNL,
		"worst_code": analysis.WorstCode,
	}, analysis)

	return analysis
}

// findCodeWidth measures the voltage range that produces a given code.
func findCodeWidth(a *ADC, code int, lsb float64) float64 {
	// Binary search for lower and upper transitions
	// Simplified: use nominal + error model
	baseWidth := lsb
	dnlFactor := 1.0 + a.DNL*(0.5-float64(code%5)/4.0)
	return baseWidth * dnlFactor
}

// AnalyzeINLDNLAtCondition computes DAC/ADC INL-DNL at a specific temperature and corner.
func AnalyzeINLDNLAtCondition(dac *DAC, adc *ADC, temperatureK float64, corner ProcessCorner) *PVTINLDNLAnalysis {
	if temperatureK <= 0 {
		temperatureK = referenceTemperatureK
	}

	dacINL, dacDNL := EffectiveINLDNL(dac.INL, dac.DNL, temperatureK, corner)
	adcINL, adcDNL := EffectiveINLDNL(adc.INL, adc.DNL, temperatureK, corner)

	dacClone := *dac
	dacClone.INL = dacINL
	dacClone.DNL = dacDNL

	adcClone := *adc
	adcClone.INL = adcINL
	adcClone.DNL = adcDNL

	inlScale := 0.0
	dnlScale := 0.0
	if dac.INL != 0 {
		inlScale = dacINL / dac.INL
	}
	if dac.DNL != 0 {
		dnlScale = dacDNL / dac.DNL
	}

	return &PVTINLDNLAnalysis{
		TemperatureK: temperatureK,
		Corner:       corner,
		INLScale:     inlScale,
		DNLScale:     dnlScale,
		DAC:          dacClone.AnalyzeINLDNL(),
		ADC:          adcClone.AnalyzeINLDNL(),
	}
}

// ProcessCornerAnalysis aggregates typical/fast/slow analyses at one temperature.
type ProcessCornerAnalysis struct {
	TemperatureK float64
	Fast         *PVTINLDNLAnalysis
	Typical      *PVTINLDNLAnalysis
	Slow         *PVTINLDNLAnalysis
}

// AnalyzeProcessCorners computes fast/typical/slow corner INL-DNL analyses.
func AnalyzeProcessCorners(dac *DAC, adc *ADC, temperatureK float64) *ProcessCornerAnalysis {
	if temperatureK <= 0 {
		temperatureK = referenceTemperatureK
	}
	return &ProcessCornerAnalysis{
		TemperatureK: temperatureK,
		Fast:         AnalyzeINLDNLAtCondition(dac, adc, temperatureK, CornerFast),
		Typical:      AnalyzeINLDNLAtCondition(dac, adc, temperatureK, CornerTypical),
		Slow:         AnalyzeINLDNLAtCondition(dac, adc, temperatureK, CornerSlow),
	}
}

// TimingAnalysis contains timing parameters for peripheral operations.
type TimingAnalysis struct {
	DACSettle     float64 // DAC settling time (s)
	ArraySettle   float64 // Array RC/sneak settling time (s)
	PumpRise      float64 // Charge pump rise time (s)
	WritePulse    float64 // Program pulse width (s)
	WriteTime     float64 // Total write time (s)
	TIASettle     float64 // TIA settling time (s)
	ADCConvert    float64 // ADC conversion time (s)
	ReadTime      float64 // Total read time (s)
	CycleTime     float64 // Full read+write cycle (s)
	MaxThroughput float64 // Maximum operations per second
}

// AnalyzeTiming computes timing for a complete peripheral system.
func AnalyzeTiming(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump) *TimingAnalysis {
	log.Input("AnalyzeTiming", map[string]interface{}{
		"dac_settle_time": dac.SettleTime,
		"adc_conv_time":   adc.ConversionTime,
	})

	arraySettle := 5e-9  // Array RC/sneak settling time (5 ns)
	writePulse := 100e-9 // Write pulse duration (100 ns)

	t := &TimingAnalysis{
		DACSettle:   dac.SettleTime * 1e-9,
		ArraySettle: arraySettle,
		PumpRise:    pump.RiseTime(),
		WritePulse:  writePulse,
		TIASettle:   tia.SettlingTime(),
		ADCConvert:  adc.ConversionTime * 1e-9,
	}

	// Write path timing
	t.WriteTime = t.DACSettle + t.PumpRise + t.WritePulse + t.ArraySettle

	// Read path timing
	t.ReadTime = t.DACSettle + t.ArraySettle + t.TIASettle + t.ADCConvert

	// Full cycle
	t.CycleTime = t.WriteTime + t.ReadTime

	// Throughput (assuming parallel columns)
	t.MaxThroughput = 1.0 / t.CycleTime

	log.Calculation("AnalyzeTiming", map[string]interface{}{
		"write_time":     t.WriteTime,
		"read_time":      t.ReadTime,
		"cycle_time":     t.CycleTime,
		"max_throughput": t.MaxThroughput,
	}, t)

	return t
}

// PowerBreakdown contains power consumption breakdown.
type PowerBreakdown struct {
	DACPower   float64 // DAC power (W)
	ADCPower   float64 // ADC power (W)
	TIAPower   float64 // TIA power (W)
	PumpPower  float64 // Charge pump power (W)
	TotalPower float64 // Total peripheral power (W)

	// Energy per operation
	DACEnergy   float64
	ADCEnergy   float64
	TIAEnergy   float64
	PumpEnergy  float64
	TotalEnergy float64 // Per operation

	// Fractions
	DACFraction  float64
	ADCFraction  float64
	TIAFraction  float64
	PumpFraction float64
}

// AnalyzePower computes power breakdown for peripheral system.
func AnalyzePower(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump, timing *TimingAnalysis) *PowerBreakdown {
	log.Input("AnalyzePower", map[string]interface{}{
		"cycle_time": timing.CycleTime,
		"read_time":  timing.ReadTime,
	})

	p := &PowerBreakdown{}

	// Energy per operation
	p.DACEnergy = dac.EnergyPerConversion()
	p.ADCEnergy = adc.EnergyPerConversion()
	p.TIAEnergy = tia.PowerConsumption() * timing.ReadTime
	p.PumpEnergy = pump.EnergyPerOperation(100e-9)

	p.TotalEnergy = p.DACEnergy + p.ADCEnergy + p.TIAEnergy + p.PumpEnergy

	// Power = Energy / Time
	if timing.CycleTime > 0 {
		p.DACPower = p.DACEnergy / timing.CycleTime
		p.ADCPower = p.ADCEnergy / timing.CycleTime
		p.TIAPower = p.TIAEnergy / timing.CycleTime
		p.PumpPower = p.PumpEnergy / timing.CycleTime
		p.TotalPower = p.TotalEnergy / timing.CycleTime
	}

	// Fractions
	if p.TotalEnergy > 0 {
		p.DACFraction = p.DACEnergy / p.TotalEnergy
		p.ADCFraction = p.ADCEnergy / p.TotalEnergy
		p.TIAFraction = p.TIAEnergy / p.TotalEnergy
		p.PumpFraction = p.PumpEnergy / p.TotalEnergy
	}

	log.Calculation("AnalyzePower", map[string]interface{}{
		"dac_energy":   p.DACEnergy,
		"adc_energy":   p.ADCEnergy,
		"tia_energy":   p.TIAEnergy,
		"pump_energy":  p.PumpEnergy,
		"total_energy": p.TotalEnergy,
		"total_power":  p.TotalPower,
	}, p)

	return p
}

// TransferFunction computes the full system transfer function.
type TransferFunction struct {
	InputLevels  []int     // Digital input levels
	DACVoltages  []float64 // DAC output voltages
	PumpVoltages []float64 // After charge pump
	TIAVoltages  []float64 // TIA output voltages
	ADCLevels    []int     // Final ADC output levels
	Errors       []int     // Output - Input error
}

// ComputeTransferFunction traces signals through the full peripheral chain.
func ComputeTransferFunction(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump) *TransferFunction {
	log.Input("ComputeTransferFunction", map[string]interface{}{
		"dac_bits": dac.Bits,
		"adc_bits": adc.Bits,
		"tia_gain": tia.Gain,
		"pump_eff": pump.Efficiency,
	})

	tf := &TransferFunction{
		InputLevels:  make([]int, 30),
		DACVoltages:  make([]float64, 30),
		PumpVoltages: make([]float64, 30),
		TIAVoltages:  make([]float64, 30),
		ADCLevels:    make([]int, 30),
		Errors:       make([]int, 30),
	}

	for i := 0; i < 30; i++ {
		tf.InputLevels[i] = i

		// DAC conversion
		tf.DACVoltages[i] = dac.ConvertWithNonlinearity(i)

		// Charge pump (for write path, voltage boosting with regulation clamp)
		boosted := tf.DACVoltages[i] * pump.BoostFactor()
		maxOut := pump.ActualOutputVoltage()
		if maxOut != 0 && math.Abs(boosted) > math.Abs(maxOut) {
			boosted = math.Copysign(math.Abs(maxOut), boosted)
		}
		tf.PumpVoltages[i] = boosted

		// TIA (simulating read-back - current proportional to programmed level)
		// Assume conductance proportional to level, Vread = 0.1V
		Vread := 0.1
		Gmin := 1e-6   // 1 µS
		Gmax := 100e-6 // 100 µS
		G := Gmin + (Gmax-Gmin)*float64(i)/29.0
		current := G * Vread
		tf.TIAVoltages[i] = tia.Convert(current)

		// ADC conversion
		tf.ADCLevels[i] = adc.ConvertWithNonlinearity(tf.TIAVoltages[i])

		// Error
		tf.Errors[i] = tf.ADCLevels[i] - i
	}

	log.Calculation("ComputeTransferFunction", map[string]interface{}{
		"levels":    30,
		"max_error": maxAbsError(tf.Errors),
	}, tf)

	return tf
}

// maxAbsError returns the maximum absolute error value.
func maxAbsError(errors []int) int {
	maxErr := 0
	for _, e := range errors {
		if e < 0 {
			e = -e
		}
		if e > maxErr {
			maxErr = e
		}
	}
	return maxErr
}

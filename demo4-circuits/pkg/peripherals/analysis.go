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

// AnalyzeINLDNL computes detailed INL/DNL for a DAC.
func (d *DAC) AnalyzeINLDNL() *INLDNLAnalysis {
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

	return analysis
}

// AnalyzeINLDNL computes detailed INL/DNL for an ADC.
func (a *ADC) AnalyzeINLDNL() *INLDNLAnalysis {
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

// TimingAnalysis contains timing parameters for peripheral operations.
type TimingAnalysis struct {
	DACSettle     float64 // DAC settling time (s)
	PumpRise      float64 // Charge pump rise time (s)
	WriteTime     float64 // Total write time (s)
	TIASettle     float64 // TIA settling time (s)
	ADCConvert    float64 // ADC conversion time (s)
	ReadTime      float64 // Total read time (s)
	CycleTime     float64 // Full read+write cycle (s)
	MaxThroughput float64 // Maximum operations per second
}

// AnalyzeTiming computes timing for a complete peripheral system.
func AnalyzeTiming(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump) *TimingAnalysis {
	t := &TimingAnalysis{
		DACSettle:  dac.SettleTime * 1e-9,
		PumpRise:   pump.RiseTime(),
		TIASettle:  tia.SettlingTime(),
		ADCConvert: adc.ConversionTime * 1e-9,
	}

	// Write path timing
	t.WriteTime = t.DACSettle + t.PumpRise + 100e-9 // 100ns write pulse

	// Read path timing
	t.ReadTime = t.TIASettle + t.ADCConvert

	// Full cycle
	t.CycleTime = t.WriteTime + t.ReadTime

	// Throughput (assuming parallel columns)
	t.MaxThroughput = 1.0 / t.CycleTime

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

		// Charge pump (for write path, shows voltage boosting)
		if tf.DACVoltages[i] > 0 {
			tf.PumpVoltages[i] = tf.DACVoltages[i] * pump.BoostFactor() * pump.Efficiency
		} else {
			tf.PumpVoltages[i] = tf.DACVoltages[i] * pump.BoostFactor() * pump.Efficiency
		}

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

	return tf
}

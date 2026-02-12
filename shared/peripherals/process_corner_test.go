package peripherals

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func cornerFactor(corner ProcessCorner) float64 {
	switch corner {
	case CornerFast:
		return 1.05
	case CornerSlow:
		return 0.95
	default:
		return 1.00
	}
}

func simulatePVTPeripherals(tempC, supplyScale float64, corner ProcessCorner) (*DAC, *ADC, *TIA, *ChargePump, *VoltageRegulator, *SampleAndHold) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	cp := DefaultChargePump()
	vr := DefaultVoltageRegulator()
	sh := DefaultSampleAndHold()

	tempK := tempC + 273.15
	dac.INL, dac.DNL = EffectiveINLDNL(dac.INL, dac.DNL, tempK, corner)
	adc.INL, adc.DNL = EffectiveINLDNL(adc.INL, adc.DNL, tempK, corner)

	cf := cornerFactor(corner)
	thermalGainDerate := 1 + 0.0015*(tempC-25.0)
	thermalBWDerate := 1 + 0.0020*(tempC-25.0)
	if thermalGainDerate < 0.5 {
		thermalGainDerate = 0.5
	}
	if thermalBWDerate < 0.5 {
		thermalBWDerate = 0.5
	}
	tia.Gain *= (cf * supplyScale) / thermalGainDerate
	tia.Bandwidth *= (cf * supplyScale) / thermalBWDerate

	cp.InputVoltage *= supplyScale
	diodeTempScale := 1 - 0.0005*(tempC-25.0)
	if diodeTempScale < 0.5 {
		diodeTempScale = 0.5
	}
	cp.DiodeDrop *= diodeTempScale
	if corner == CornerFast {
		cp.DiodeDrop *= 0.95
	} else if corner == CornerSlow {
		cp.DiodeDrop *= 1.05
	}

	switchTempScale := (1 + 0.003*(tempC-25.0))
	if switchTempScale < 0.5 {
		switchTempScale = 0.5
	}
	sh.SwitchResistance *= switchTempScale
	if corner == CornerFast {
		sh.SwitchResistance *= 0.9
	} else if corner == CornerSlow {
		sh.SwitchResistance *= 1.1
	}

	return dac, adc, tia, cp, vr, sh
}

func verifyPeripheralsWithinSpec(tempC, supplyScale float64, corner ProcessCorner) error {
	dac, adc, tia, cp, vr, sh := simulatePVTPeripherals(tempC, supplyScale, corner)

	if math.Abs(dac.INL) > 1.0 || math.Abs(adc.INL) > 1.0 {
		return fmt.Errorf("INL out of spec (DAC=%.4f LSB, ADC=%.4f LSB)", dac.INL, adc.INL)
	}
	if math.Abs(dac.DNL) > 0.7 || math.Abs(adc.DNL) > 0.7 {
		return fmt.Errorf("DNL out of spec (DAC=%.4f LSB, ADC=%.4f LSB)", dac.DNL, adc.DNL)
	}

	if v := dac.ConvertWithCondition(dac.Levels()-1, tempC+273.15, corner); v < dac.VrefHigh-0.2 || v > dac.VrefHigh+0.2 {
		return fmt.Errorf("DAC high-code output out of spec: %.4f V", v)
	}
	if l := adc.ConvertWithCondition(0.5*(adc.VrefHigh+adc.VrefLow), tempC+273.15, corner); l < 10 || l > 21 {
		return fmt.Errorf("ADC mid-scale code out of spec: %d", l)
	}

	midCurrent := 50e-6
	tiaOut := tia.Convert(midCurrent)
	if tiaOut < 0 || tiaOut > tia.MaxOutputVoltage {
		return fmt.Errorf("TIA output out of range: %.6f V", tiaOut)
	}
	if snr := tia.SNR(midCurrent); math.IsNaN(snr) || snr < 20 {
		return fmt.Errorf("TIA SNR out of spec: %.2f dB", snr)
	}

	cpOut := math.Abs(cp.ActualOutputVoltage())
	if cpOut < 1.35 || cpOut > 1.50 {
		return fmt.Errorf("charge pump output out of spec: %.4f V", cpOut)
	}

	vout := vr.Regulate(1.8*supplyScale, 100e-6)
	if vout < 1.08 || vout > 1.26 {
		return fmt.Errorf("regulator output out of spec: %.4f V", vout)
	}

	acqSettled := sh.SettledFraction(sh.AcquisitionTimeNS * 1e-9)
	if acqSettled < 0.95 {
		return fmt.Errorf("sample/hold acquisition out of spec: %.6f", acqSettled)
	}
	if droop := sh.HoldDroop(50e-9); droop < 0.999 {
		return fmt.Errorf("sample/hold droop out of spec: %.6f", droop)
	}

	return nil
}

func TestProcessCorner_Fast_AllPeripheralsWithinSpec(t *testing.T) {
	if err := verifyPeripheralsWithinSpec(-40.0, 1.10, CornerFast); err != nil {
		t.Fatalf("fast corner failed at -40C, +10%% supply: %v", err)
	}
}

func TestProcessCorner_Slow_AllPeripheralsWithinSpec(t *testing.T) {
	if err := verifyPeripheralsWithinSpec(125.0, 0.90, CornerSlow); err != nil {
		t.Fatalf("slow corner failed at 125C, -10%% supply: %v", err)
	}
}

func TestProcessCorner_Typical_MatchesNominalExactly(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()
	cp := DefaultChargePump()
	vr := DefaultVoltageRegulator()
	sh := DefaultSampleAndHold()

	if dac.Bits != DefaultBits || adc.Bits != DefaultBits {
		t.Fatalf("nominal resolution mismatch: dac=%d adc=%d expected=%d", dac.Bits, adc.Bits, DefaultBits)
	}
	if dac.VrefHigh != DACVrefHigh || dac.VrefLow != DACVrefLow {
		t.Fatalf("DAC nominal refs mismatch: got [%.3f, %.3f], expected [%.3f, %.3f]", dac.VrefLow, dac.VrefHigh, DACVrefLow, DACVrefHigh)
	}
	if adc.VrefHigh != ADCVrefHigh || adc.VrefLow != ADCVrefLow {
		t.Fatalf("ADC nominal refs mismatch: got [%.3f, %.3f], expected [%.3f, %.3f]", adc.VrefLow, adc.VrefHigh, ADCVrefLow, ADCVrefHigh)
	}
	if dac.INL != DefaultINL || dac.DNL != DefaultDNL || adc.INL != DefaultINL || adc.DNL != DefaultDNL {
		t.Fatal("nominal INL/DNL mismatch")
	}

	for _, code := range []int{0, 7, 15, 31} {
		vNominal := dac.ConvertWithNonlinearity(code)
		vTypical := dac.ConvertWithCondition(code, 300, CornerTypical)
		if vNominal != vTypical {
			t.Fatalf("DAC typical mismatch at code %d: %.12f vs %.12f", code, vNominal, vTypical)
		}
	}
	for _, vin := range []float64{0.1, 0.3, 0.7, 0.95} {
		lNominal := adc.ConvertWithNonlinearity(vin)
		lTypical := adc.ConvertWithCondition(vin, 300, CornerTypical)
		if lNominal != lTypical {
			t.Fatalf("ADC typical mismatch at vin=%.3f: %d vs %d", vin, lNominal, lTypical)
		}
	}

	if tia.Gain != 10e3 || tia.Bandwidth != 100e6 {
		t.Fatalf("TIA nominal mismatch: gain=%.0f bw=%.0f", tia.Gain, tia.Bandwidth)
	}
	if cp.ActualOutputVoltage() != 1.5 {
		t.Fatalf("charge pump nominal output mismatch: %.6f V", cp.ActualOutputVoltage())
	}
	if math.Abs(vr.Regulate(1.8, 100e-6)-1.19995) > 1e-12 {
		t.Fatalf("regulator nominal output mismatch: %.8f V", vr.Regulate(1.8, 100e-6))
	}
	if sh.SettledFraction(sh.AcquisitionTimeNS*1e-9) <= 0.99999999 {
		t.Fatalf("sample/hold nominal settling too low: %.12f", sh.SettledFraction(sh.AcquisitionTimeNS*1e-9))
	}
}

func TestProcessCorner_MonteCarlo_100RandomPVT_NoFailures(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	corners := []ProcessCorner{CornerFast, CornerTypical, CornerSlow}

	for i := 0; i < 100; i++ {
		tempC := -40.0 + rng.Float64()*(125.0+40.0)
		supplyScale := 0.90 + rng.Float64()*0.20
		corner := corners[rng.Intn(len(corners))]

		if err := verifyPeripheralsWithinSpec(tempC, supplyScale, corner); err != nil {
			t.Fatalf("monte carlo failure at sample %d (temp=%.2fC supply=%.4f corner=%s): %v", i, tempC, supplyScale, corner, err)
		}
	}
}

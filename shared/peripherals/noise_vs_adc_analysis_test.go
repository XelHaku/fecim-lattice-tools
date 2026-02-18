package peripherals

import (
	"fmt"
	"math"
	"testing"
)

type adcNoiseRow struct {
	bits         int
	lsbV         float64
	quantVarV2   float64
	thermalVarV2 float64
	totalVarV2   float64
	enob         float64
	dominant     string
}

func dominantNoiseSource(quantVar, thermalVar, shotVar, flickerVar float64) string {
	dominant := "Quantization"
	maxVar := quantVar
	if thermalVar > maxVar {
		dominant = "Thermal"
		maxVar = thermalVar
	}
	if shotVar > maxVar {
		dominant = "Shot"
		maxVar = shotVar
	}
	if flickerVar > maxVar {
		dominant = "1/f"
	}
	return dominant
}

func TestNoiseVsADC_M4INV04_TableAndCrossover(t *testing.T) {
	const (
		vRange        = 1.8
		tempK         = 300.0
		bandwidthHz   = 10e6
		rTIAOhm       = 10e3
		shotCurrentA  = 1e-6
		flickerK      = 1e-12
		flickerFreqHz = 1e3
		signalRMS     = vRange / (2 * math.Sqrt2) // full-scale sine within 0-1.8 V range
	)

	bitsList := []int{4, 5, 6, 7, 8, 10, 12}
	rows := make([]adcNoiseRow, 0, len(bitsList))

	thermalSigma := ThermalNoiseRMS(tempK, rTIAOhm, bandwidthHz)
	thermalVar := thermalSigma * thermalSigma
	shotCurrentSigma := ShotNoiseCurrentRMS(shotCurrentA, bandwidthHz)
	shotVar := (shotCurrentSigma * rTIAOhm) * (shotCurrentSigma * rTIAOhm)
	flickerVar := FlickerNoisePower(flickerK, flickerFreqHz)

	for _, bits := range bitsList {
		lsb := vRange / math.Pow(2, float64(bits))
		quantVarFromLSB := (lsb * lsb) / 12.0
		quantVar := QuantizationNoiseVariance(vRange, bits)
		if math.Abs(quantVar-quantVarFromLSB) > 1e-18 {
			t.Fatalf("quantization mismatch for %d bits: from function=%.6e from LSB=%.6e", bits, quantVar, quantVarFromLSB)
		}

		totalVar := thermalVar + quantVar + shotVar + flickerVar
		noiseRMS := math.Sqrt(totalVar)
		snrDB := SNRDB(signalRMS, noiseRMS)
		enob := (snrDB - 1.76) / 6.02

		rows = append(rows, adcNoiseRow{
			bits:         bits,
			lsbV:         lsb,
			quantVarV2:   quantVar,
			thermalVarV2: thermalVar,
			totalVarV2:   totalVar,
			enob:         enob,
			dominant:     dominantNoiseSource(quantVar, thermalVar, shotVar, flickerVar),
		})
	}

	for _, r := range rows {
		if r.enob <= 0 {
			t.Fatalf("ENOB must be positive for %d bits, got %.3f", r.bits, r.enob)
		}
		if r.totalVarV2 < r.quantVarV2 || r.totalVarV2 < r.thermalVarV2 {
			t.Fatalf("total variance must include individual components for %d bits", r.bits)
		}
	}

	crossoverBit := -1
	for bits := 1; bits <= 16; bits++ {
		qVar := QuantizationNoiseVariance(vRange, bits)
		if thermalVar > qVar {
			crossoverBit = bits
			break
		}
	}
	if crossoverBit < 0 {
		t.Fatalf("failed to find thermal>quantization crossover")
	}
	if crossoverBit <= 12 {
		t.Fatalf("unexpected crossover (%d bits): expected thermal crossover above 12 bits at R_TIA=10kΩ, BW=10MHz", crossoverBit)
	}

	t.Logf("Lane4 M4-INV-04 baseline crossover (R_TIA=10kΩ, BW=10MHz): thermal exceeds quantization at ~%d bits", crossoverBit)
	for _, r := range rows {
		t.Logf("bits=%2d LSB=%.6f mV quant=%.6e V^2 thermal=%.6e V^2 total=%.6e V^2 ENOB=%.3f dominant=%s",
			r.bits, r.lsbV*1e3, r.quantVarV2, r.thermalVarV2, r.totalVarV2, r.enob, r.dominant)
	}
}

func TestNoiseVsADC_M4INV04_ArrayScalingByWireResistance(t *testing.T) {
	const (
		vRange            = 1.8
		tempK             = 300.0
		bandwidthHz       = 10e6
		rTIAOhm           = 10e3
		wireResPerCellOhm = 10.0 // lumped total wire resistance per traversed cell segment
	)

	arraySizes := []int{8, 64, 128}
	prevThermalVar := 0.0
	for _, n := range arraySizes {
		totalWireR := 2 * float64(n-1) * wireResPerCellOhm
		effectiveR := rTIAOhm + totalWireR
		thermalVar := math.Pow(ThermalNoiseRMS(tempK, effectiveR, bandwidthHz), 2)

		if thermalVar <= prevThermalVar {
			t.Fatalf("thermal variance should increase with array size: n=%d got=%.6e prev=%.6e", n, thermalVar, prevThermalVar)
		}
		prevThermalVar = thermalVar

		crossoverBit := -1
		for bits := 1; bits <= 16; bits++ {
			if thermalVar > QuantizationNoiseVariance(vRange, bits) {
				crossoverBit = bits
				break
			}
		}
		if crossoverBit < 0 {
			t.Fatalf("missing crossover for %dx%d array", n, n)
		}
		t.Logf("array=%dx%d effectiveR=%.2f Ω thermalVar=%.6e V^2 crossover≈%d bits", n, n, effectiveR, thermalVar, crossoverBit)
	}
}

func TestNoiseVsADC_M4INV04_UsefulADCCeilingRecommendation(t *testing.T) {
	const (
		vRange      = 1.8
		tempK       = 300.0
		bandwidthHz = 10e6
		rTIAOhm     = 10e3
	)

	thermalVar := math.Pow(ThermalNoiseRMS(tempK, rTIAOhm, bandwidthHz), 2)
	ceiling := -1
	for bits := 1; bits <= 16; bits++ {
		qVar := QuantizationNoiseVariance(vRange, bits)
		if qVar > thermalVar {
			ceiling = bits
		} else {
			break
		}
	}

	if ceiling < 8 {
		t.Fatalf("useful ADC ceiling too low/unexpected: %d bits", ceiling)
	}

	signalRMS := vRange / (2 * math.Sqrt2)
	totalVarAtCeiling := thermalVar + QuantizationNoiseVariance(vRange, ceiling)
	enobCeiling := (SNRDB(signalRMS, math.Sqrt(totalVarAtCeiling)) - 1.76) / 6.02

	rec := fmt.Sprintf("Useful ADC ceiling at R_TIA=10kΩ, BW=10MHz: %d bits (ENOB=%.2f). Diminishing returns above %d bits where thermal noise dominates.", ceiling, enobCeiling, ceiling+1)
	t.Log(rec)
}

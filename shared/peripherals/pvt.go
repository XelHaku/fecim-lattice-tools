package peripherals

import "math"

// ProcessCorner captures first-order process variation assumptions.
type ProcessCorner string

const (
	CornerFast    ProcessCorner = "fast"
	CornerTypical ProcessCorner = "typical"
	CornerSlow    ProcessCorner = "slow"
)

const (
	referenceTemperatureK = 300.0
	inlTempCoeffPer100K   = 0.12 // +12% INL per +100K
	dnlTempCoeffPer100K   = 0.08 // +8% DNL per +100K
)

func processCornerScale(corner ProcessCorner) float64 {
	switch corner {
	case CornerFast:
		return 0.80
	case CornerSlow:
		return 1.25
	default:
		return 1.00
	}
}

func temperatureScale(tempK, coeffPer100K float64) float64 {
	deltaHundredK := (tempK - referenceTemperatureK) / 100.0
	scale := 1.0 + coeffPer100K*deltaHundredK
	if scale < 0.2 {
		return 0.2
	}
	return scale
}

// EffectiveINLDNL returns INL and DNL after applying temperature + process-corner effects.
func EffectiveINLDNL(inl, dnl, tempK float64, corner ProcessCorner) (float64, float64) {
	cornerScale := processCornerScale(corner)
	inlScale := cornerScale * temperatureScale(tempK, inlTempCoeffPer100K)
	dnlScale := cornerScale * temperatureScale(tempK, dnlTempCoeffPer100K)
	return inl * inlScale, dnl * dnlScale
}

// ConvertWithCondition adds INL/DNL errors at a given PVT point.
func (d *DAC) ConvertWithCondition(level int, tempK float64, corner ProcessCorner) float64 {
	idealVoltage := d.Convert(level)
	inl, dnl := EffectiveINLDNL(d.INL, d.DNL, tempK, corner)

	lsb := (d.VrefHigh - d.VrefLow) / float64(d.Levels()-1)
	inlError := inl * lsb * math.Sin(math.Pi*float64(level)/float64(d.Levels()-1))
	dnlError := dnl * lsb * (0.5 - float64(level%5)/4.0)
	return idealVoltage + inlError + dnlError
}

// ConvertWithCondition adds INL/DNL errors at a given PVT point.
func (a *ADC) ConvertWithCondition(voltage float64, tempK float64, corner ProcessCorner) int {
	inl, dnl := EffectiveINLDNL(a.INL, a.DNL, tempK, corner)
	lsb := (a.VrefHigh - a.VrefLow) / float64(a.Levels()-1)
	idealLevel := a.Convert(voltage)
	inlOffset := inl * lsb * math.Sin(math.Pi*float64(idealLevel)/float64(a.Levels()-1))
	dnlOffset := dnl * lsb * (0.5 - float64(idealLevel%5)/4.0)
	return a.Convert(voltage + inlOffset + dnlOffset)
}

package neural

import "fmt"

// LevelSweepPoint records CIM accuracy at a given quantization level count.
type LevelSweepPoint struct {
	NumLevels int
	Accuracy  float64 // fraction 0..1
	Correct   int
	Total     int
}

// ADCSweepPoint records CIM accuracy at a given ADC bit width.
type ADCSweepPoint struct {
	ADCBits  int
	Accuracy float64
	Correct  int
	Total    int
}

// NoiseSweepPoint records CIM accuracy at a given noise level.
type NoiseSweepPoint struct {
	NoiseLevel float64 // σ/μ coefficient
	Accuracy   float64
	Correct    int
	Total      int
}

// SweepQuantizationLevels evaluates CIM accuracy across quantization level counts.
// net is reconfigured for each level value via SetNumLevels (which requantizes internally).
// The original NumLevels is restored after the sweep.
// images and labels must be pre-loaded MNIST (or similar) data.
func SweepQuantizationLevels(net *DualModeNetwork, images [][]float64, labels []int, levels []int) ([]LevelSweepPoint, error) {
	if len(images) != len(labels) {
		return nil, fmt.Errorf("images/labels length mismatch: %d vs %d", len(images), len(labels))
	}
	if len(levels) == 0 {
		return nil, fmt.Errorf("levels must not be empty")
	}

	origLevels := net.GetNumLevels()
	defer func() {
		net.SetNumLevels(origLevels)
	}()

	out := make([]LevelSweepPoint, len(levels))
	for i, lvl := range levels {
		if lvl < 2 {
			return nil, fmt.Errorf("level %d must be >= 2", lvl)
		}
		net.SetNumLevels(lvl)

		correct := 0
		for j, img := range images {
			res := net.Infer(img)
			if res != nil && res.CIMPrediction == labels[j] {
				correct++
			}
		}
		out[i] = LevelSweepPoint{
			NumLevels: lvl,
			Accuracy:  float64(correct) / float64(len(images)),
			Correct:   correct,
			Total:     len(images),
		}
	}
	return out, nil
}

// SweepADCBits evaluates CIM accuracy across ADC bit widths.
// net is reconfigured for each bit width; original ADCBits is restored after.
// Valid range is [3, 16]; values outside this range are rejected.
func SweepADCBits(net *DualModeNetwork, images [][]float64, labels []int, bitWidths []int) ([]ADCSweepPoint, error) {
	if len(images) != len(labels) {
		return nil, fmt.Errorf("images/labels length mismatch: %d vs %d", len(images), len(labels))
	}
	if len(bitWidths) == 0 {
		return nil, fmt.Errorf("bitWidths must not be empty")
	}

	origBits := net.Config.ADCBits
	defer func() { net.SetADCBits(origBits) }()

	out := make([]ADCSweepPoint, len(bitWidths))
	for i, bits := range bitWidths {
		if bits < 3 || bits > 16 {
			return nil, fmt.Errorf("ADC bits %d must be in [3, 16]", bits)
		}
		net.SetADCBits(bits)

		correct := 0
		for j, img := range images {
			res := net.Infer(img)
			if res != nil && res.CIMPrediction == labels[j] {
				correct++
			}
		}
		out[i] = ADCSweepPoint{
			ADCBits:  bits,
			Accuracy: float64(correct) / float64(len(images)),
			Correct:  correct,
			Total:    len(images),
		}
	}
	return out, nil
}

// SweepNoiseLevel evaluates CIM accuracy across noise levels.
// net is reconfigured for each noise level; original NoiseLevel is restored after.
// Valid range is [0, 0.20] matching SetNoiseLevel's clamp range.
func SweepNoiseLevel(net *DualModeNetwork, images [][]float64, labels []int, noiseLevels []float64) ([]NoiseSweepPoint, error) {
	if len(images) != len(labels) {
		return nil, fmt.Errorf("images/labels length mismatch: %d vs %d", len(images), len(labels))
	}
	if len(noiseLevels) == 0 {
		return nil, fmt.Errorf("noiseLevels must not be empty")
	}

	origNoise := net.Config.NoiseLevel
	defer func() { net.SetNoiseLevel(origNoise) }()

	out := make([]NoiseSweepPoint, len(noiseLevels))
	for i, nl := range noiseLevels {
		if nl < 0 || nl > 0.20 {
			return nil, fmt.Errorf("noise level %g must be in [0, 0.20]", nl)
		}
		net.SetNoiseLevel(nl)

		correct := 0
		for j, img := range images {
			res := net.Infer(img)
			if res != nil && res.CIMPrediction == labels[j] {
				correct++
			}
		}
		out[i] = NoiseSweepPoint{
			NoiseLevel: nl,
			Accuracy:   float64(correct) / float64(len(images)),
			Correct:    correct,
			Total:      len(images),
		}
	}
	return out, nil
}

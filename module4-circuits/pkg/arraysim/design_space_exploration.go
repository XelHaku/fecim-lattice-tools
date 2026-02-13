package arraysim

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
)

// DesignSpacePoint captures one array/ADC/device operating point.
type DesignSpacePoint struct {
	ArraySize int
	ADCBits   int
	Device    string
	LatencyNS float64
	EnergyPJ  float64
	Accuracy  float64
}

// BuildDesignSpacePoints sweeps array size x ADC bits x device.
func BuildDesignSpacePoints(arraySizes, adcBits []int, devices []string) []DesignSpacePoint {
	out := make([]DesignSpacePoint, 0, len(arraySizes)*len(adcBits)*len(devices))
	for _, n := range arraySizes {
		if n <= 0 {
			continue
		}
		for _, bits := range adcBits {
			if bits <= 0 {
				continue
			}
			for _, dev := range devices {
				lat, en, acc := estimatePoint(n, bits, dev)
				out = append(out, DesignSpacePoint{
					ArraySize: n,
					ADCBits:   bits,
					Device:    dev,
					LatencyNS: lat,
					EnergyPJ:  en,
					Accuracy:  acc,
				})
			}
		}
	}
	return out
}

// ParetoFront returns non-dominated points minimizing latency and energy and maximizing accuracy.
func ParetoFront(points []DesignSpacePoint) []DesignSpacePoint {
	front := make([]DesignSpacePoint, 0, len(points))
	for i := range points {
		dominated := false
		for j := range points {
			if i == j {
				continue
			}
			if dominates(points[j], points[i]) {
				dominated = true
				break
			}
		}
		if !dominated {
			front = append(front, points[i])
		}
	}
	sort.Slice(front, func(i, j int) bool {
		if front[i].EnergyPJ == front[j].EnergyPJ {
			return front[i].LatencyNS < front[j].LatencyNS
		}
		return front[i].EnergyPJ < front[j].EnergyPJ
	})
	return front
}

// ExportParetoCSV writes sweep points as CSV for downstream analysis.
func ExportParetoCSV(points []DesignSpacePoint, w io.Writer) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"array_size", "adc_bits", "device", "latency_ns", "energy_pj", "accuracy"}); err != nil {
		return err
	}
	for _, p := range points {
		rec := []string{
			fmt.Sprintf("%d", p.ArraySize),
			fmt.Sprintf("%d", p.ADCBits),
			p.Device,
			fmt.Sprintf("%.6f", p.LatencyNS),
			fmt.Sprintf("%.6f", p.EnergyPJ),
			fmt.Sprintf("%.6f", p.Accuracy),
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func dominates(a, b DesignSpacePoint) bool {
	notWorse := a.LatencyNS <= b.LatencyNS && a.EnergyPJ <= b.EnergyPJ && a.Accuracy >= b.Accuracy
	strictlyBetter := a.LatencyNS < b.LatencyNS || a.EnergyPJ < b.EnergyPJ || a.Accuracy > b.Accuracy
	return notWorse && strictlyBetter
}

func estimatePoint(arraySize, adcBits int, device string) (latencyNS, energyPJ, accuracy float64) {
	baseLatency := 70.0 * float64(arraySize*arraySize) / 64.0
	adcPenalty := 1.0 + 0.07*float64(adcBits-5)
	if adcPenalty < 0.55 {
		adcPenalty = 0.55
	}
	latencyNS = baseLatency * adcPenalty

	devScale := map[string]float64{"FeFET": 1.0, "RRAM": 1.5, "PCM": 1.9, "SRAM": 5.8}
	s := devScale[device]
	if s == 0 {
		s = 1.2
	}
	energyPJ = 2.7 * float64(arraySize*arraySize) / 64.0 * adcPenalty * s

	bitGain := 0.006 * float64(adcBits-4)
	devicePenalty := map[string]float64{"FeFET": 0.00, "RRAM": 0.01, "PCM": 0.014, "SRAM": 0.004}[device]
	accuracy = 0.942 + bitGain - devicePenalty
	if accuracy > 0.995 {
		accuracy = 0.995
	}
	if accuracy < 0.70 {
		accuracy = 0.70
	}
	return latencyNS, energyPJ, accuracy
}

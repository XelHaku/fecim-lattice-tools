package physics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// CalibrationPoint stores one E-P sample pair (SI units).
type CalibrationPoint struct {
	E float64 `json:"E"`
	P float64 `json:"P"`
}

type CalibrationModel string

const (
	CalibrationModelPreisach CalibrationModel = "preisach"
	CalibrationModelLK       CalibrationModel = "lk"
)

// CalibrationBundle is the exported fitting result.
type CalibrationBundle struct {
	Model        CalibrationModel   `json:"model"`
	Parameters   map[string]float64 `json:"parameters"`
	RMSE         float64            `json:"rmse"`
	RelativeRMSE float64            `json:"relative_rmse"`
	Samples      int                `json:"samples"`
	GeneratedAt  time.Time          `json:"generated_at"`
	Data         []CalibrationPoint `json:"data"`
}

// ImportCalibrationCSV loads E,P data from CSV (header optional).
func ImportCalibrationCSV(path string) ([]CalibrationPoint, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	points := make([]CalibrationPoint, 0, len(rows))
	for i, row := range rows {
		if len(row) < 2 {
			continue
		}
		e, errE := strconv.ParseFloat(strings.TrimSpace(row[0]), 64)
		p, errP := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		if errE != nil || errP != nil {
			if i == 0 {
				continue // likely header
			}
			return nil, fmt.Errorf("parse row %d: %v %v", i, errE, errP)
		}
		points = append(points, CalibrationPoint{E: e, P: p})
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("no calibration points")
	}
	return points, nil
}

// FitCalibration performs a simple gradient-free random search.
func FitCalibration(points []CalibrationPoint, model CalibrationModel, iterations int, seed int64) CalibrationBundle {
	if iterations <= 0 {
		iterations = 1000
	}
	rng := rand.New(rand.NewSource(seed))

	best := defaultParams(model)
	bestRMSE := rmse(points, model, best)

	for i := 0; i < iterations; i++ {
		cand := jitter(model, best, rng, 0.25)
		r := rmse(points, model, cand)
		if r < bestRMSE {
			bestRMSE = r
			best = cand
		}
	}

	return CalibrationBundle{
		Model:        model,
		Parameters:   best,
		RMSE:         bestRMSE,
		RelativeRMSE: relativeRMSE(points, bestRMSE),
		Samples:      len(points),
		GeneratedAt:  time.Now(),
		Data:         points,
	}
}

// ExportCalibrationBundle writes calibration bundle to JSON.
func ExportCalibrationBundle(path string, bundle CalibrationBundle) error {
	b, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func modelP(model CalibrationModel, params map[string]float64, E float64) float64 {
	switch model {
	case CalibrationModelLK:
		return params["a1"]*E + params["a3"]*E*E*E + params["a5"]*math.Pow(E, 5)
	case CalibrationModelPreisach:
		fallthrough
	default:
		ps := params["ps"]
		scale := params["scale"]
		if scale == 0 {
			scale = 1
		}
		offset := params["offset"]
		bias := params["bias"]
		return ps*math.Tanh((E-offset)/scale) + bias
	}
}

func defaultParams(model CalibrationModel) map[string]float64 {
	if model == CalibrationModelLK {
		return map[string]float64{"a1": 1e-9, "a3": -1e-27, "a5": 1e-45}
	}
	return map[string]float64{"ps": 0.25, "scale": 8e7, "offset": 0, "bias": 0}
}

func jitter(model CalibrationModel, in map[string]float64, rng *rand.Rand, frac float64) map[string]float64 {
	out := make(map[string]float64, len(in))
	for k, v := range in {
		delta := (rng.Float64()*2 - 1) * frac
		nv := v * (1 + delta)
		if k == "scale" && nv < 1e4 {
			nv = 1e4
		}
		out[k] = nv
	}
	if model == CalibrationModelPreisach && out["ps"] < 0 {
		out["ps"] = -out["ps"]
	}
	return out
}

func rmse(points []CalibrationPoint, model CalibrationModel, params map[string]float64) float64 {
	if len(points) == 0 {
		return 0
	}
	var sse float64
	for _, pt := range points {
		err := modelP(model, params, pt.E) - pt.P
		sse += err * err
	}
	return math.Sqrt(sse / float64(len(points)))
}

func relativeRMSE(points []CalibrationPoint, rmse float64) float64 {
	if len(points) == 0 {
		return 0
	}
	minP, maxP := points[0].P, points[0].P
	for _, p := range points[1:] {
		if p.P < minP {
			minP = p.P
		}
		if p.P > maxP {
			maxP = p.P
		}
	}
	span := maxP - minP
	if span <= 0 {
		return 0
	}
	return rmse / span
}

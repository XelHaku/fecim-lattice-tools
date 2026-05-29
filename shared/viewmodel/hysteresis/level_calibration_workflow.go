package hysteresis

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"fecim-lattice-tools/shared/physics"
)

type levelCalibrationWorkflow struct {
	state    LevelCalibrationState
	material *physics.HZOMaterial
}

func newLevelCalibrationWorkflow(state LevelCalibrationState, material *physics.HZOMaterial) levelCalibrationWorkflow {
	return levelCalibrationWorkflow{state: state, material: material}
}

func (w levelCalibrationWorkflow) setLevelCount(raw string) (LevelCalibrationState, error) {
	levelCount, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return w.state, fmt.Errorf("hysteresis: invalid level calibration level count %q", raw)
	}
	if levelCount < 2 || levelCount > 64 {
		return w.state, fmt.Errorf("hysteresis: level calibration level count %d outside [2,64]", levelCount)
	}
	if w.state.LevelCount == levelCount {
		return w.state, nil
	}
	w.state.LevelCount = levelCount
	w.markInputsChanged()
	return w.state, nil
}

func (w levelCalibrationWorkflow) setTargetRange(raw string) (LevelCalibrationState, error) {
	targetRange, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return w.state, fmt.Errorf("hysteresis: invalid level calibration target range %q", raw)
	}
	if targetRange < 0.5 || targetRange > 1.0 || math.IsNaN(targetRange) || math.IsInf(targetRange, 0) {
		return w.state, fmt.Errorf("hysteresis: level calibration target range %.3f outside [0.5,1.0]", targetRange)
	}
	if math.Abs(w.state.TargetRangeFrac-targetRange) < 1e-9 {
		return w.state, nil
	}
	w.state.TargetRangeFrac = targetRange
	w.markInputsChanged()
	return w.state, nil
}

func (w levelCalibrationWorkflow) setTemperature(raw string) (LevelCalibrationState, error) {
	temperatureK, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return w.state, fmt.Errorf("hysteresis: invalid level calibration temperature %q", raw)
	}
	if temperatureK < 200 || temperatureK > 700 || math.IsNaN(temperatureK) || math.IsInf(temperatureK, 0) {
		return w.state, fmt.Errorf("hysteresis: level calibration temperature %.1f K outside [200,700]", temperatureK)
	}
	if math.Abs(w.state.TemperatureK-temperatureK) < 1e-9 {
		return w.state, nil
	}
	w.state.TemperatureK = temperatureK
	w.markInputsChanged()
	return w.state, nil
}

func (w *levelCalibrationWorkflow) markInputsChanged() {
	if w.state.HasResult {
		w.state.Status = LevelCalibrationStale
	} else {
		w.state.Status = LevelCalibrationNotCalibrated
	}
	w.state.Error = ""
}

func (w levelCalibrationWorkflow) run() (LevelCalibrationState, error) {
	result, err := physics.CalibrateLevelsLK(physics.LevelCalibrationInput{
		Material:        w.material,
		LevelCount:      w.state.LevelCount,
		TargetRangeFrac: w.state.TargetRangeFrac,
		TemperatureK:    w.state.TemperatureK,
	})
	if err != nil {
		w.state.Error = err.Error()
		return w.state, nil
	}
	w.state.Result = result
	w.state.HasResult = true
	w.state.Status = LevelCalibrationFresh
	w.state.Error = ""
	return w.state, nil
}

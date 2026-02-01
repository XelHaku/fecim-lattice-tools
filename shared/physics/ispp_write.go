package physics

import (
	"math"

	"fecim-lattice-tools/shared/logging"
)

var isppLog *logging.Logger

func getISPPLogger() *logging.Logger {
	if isppLog == nil {
		isppLog = logging.NewLogger("ispp")
	}
	return isppLog
}

type WriteController struct {
	Solver   *LKSolver
	Material *HZOMaterial

	MaxVoltage    float64
	MinVoltage    float64
	PulseWidth    float64
	Tolerance     float64
	MaxIterations int

	VMin float64
	VMax float64
}

func NewWriteController(solver *LKSolver, material *HZOMaterial) *WriteController {
	return &WriteController{
		Solver:        solver,
		Material:      material,
		MaxVoltage:    3.0,
		MinVoltage:    0.0,
		PulseWidth:    10e-9,
		Tolerance:     0.01,
		MaxIterations: 20,
	}
}

func (c *WriteController) WriteTarget(targetG float64) (attempts int, success bool, overshootCount int) {
	log := getISPPLogger()

	targetP := ConductanceToPolarization(targetG, c.Material.Gmin, c.Material.Gmax, c.Material.Ps)

	c.VMin = 0.0
	c.VMax = c.MaxVoltage

	c.Solver.P = -c.Material.Pr
	c.Solver.SetState(-c.Material.Pr)

	overshoots := 0
	success = false

	log.Input("WriteTarget", map[string]interface{}{
		"targetG": targetG,
		"targetP": targetP,
	})

	var vPulse float64
	var i int
	for i = 0; i < c.MaxIterations; i++ {
		log.Calculation("WriteTarget", map[string]interface{}{
			"attempt":  i + 1,
			"currentP": c.Solver.P,
			"targetP":  targetP,
		}, nil)

		if i == 0 {
			vPulse = (c.VMin + c.VMax) / 2.0
			log.Calculation("WriteTarget", map[string]interface{}{
				"step":   "Predict",
				"vPulse": vPulse,
				"bounds": []float64{c.VMin, c.VMax},
			}, nil)
		} else {
			vPulse = (c.VMin + c.VMax) / 2.0
			log.Calculation("WriteTarget", map[string]interface{}{
				"step":   "BinarySearch",
				"vPulse": vPulse,
				"vMin":   c.VMin,
				"vMax":   c.VMax,
			}, nil)
		}

		eField := vPulse / c.Material.Thickness
		c.Solver.Step(eField, c.PulseWidth)

		log.Calculation("WriteTarget", map[string]interface{}{
			"step":   "WritePulse",
			"vPulse": vPulse,
			"eField": eField,
			"dt":     c.PulseWidth,
		}, nil)

		currentP := c.Solver.GetState()
		currentG := PolarizationToConductance(currentP, c.Material.Ps, c.Material.Gmin, c.Material.Gmax)

		error := currentG - targetG

		log.Calculation("WriteTarget", map[string]interface{}{
			"step":     "Verify",
			"currentP": currentP,
			"currentG": currentG,
			"error":    error,
		}, nil)

		if math.Abs(error) < c.Tolerance {
			success = true
			log.Input("WriteTarget", map[string]interface{}{
				"status":     "Success",
				"attempts":   i + 1,
				"finalG":     currentG,
				"finalP":     currentP,
				"overshoots": overshoots,
			})
			break
		}

		if error < 0 {
			c.VMin = math.Abs(vPulse)
			log.Calculation("WriteTarget", map[string]interface{}{
				"decision": "Undershoot",
				"vPulse":   vPulse,
				"newVMin":  c.VMin,
				"vMax":     c.VMax,
			}, nil)
		} else {
			overshoots++

			resetField := -c.MaxVoltage * 1.5
			eResetField := resetField / c.Material.Thickness
			c.Solver.Step(eResetField, c.PulseWidth*2)

			log.Calculation("WriteTarget", map[string]interface{}{
				"decision":    "Overshoot",
				"overshoots":  overshoots,
				"resetField":  resetField,
				"eResetField": eResetField,
			}, nil)

			c.VMax = math.Abs(vPulse)
			c.VMin = 0.0

			c.Solver.SetState(-c.Material.Pr)
		}
	}

	if !success {
		log.Input("WriteTarget", map[string]interface{}{
			"status":     "Failed",
			"attempts":   c.MaxIterations,
			"overshoots": overshoots,
			"reason":     "Max iterations exceeded",
		})
	}

	return i + 1, success, overshoots
}

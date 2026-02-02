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
	MaxStep       float64 // Maximum integration step size (s) for stability

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
		MaxStep:       1e-12,
	}
}

func (c *WriteController) WriteTarget(targetG float64) (attempts int, success bool, overshootCount int) {
	return c.writeTarget(targetG, true)
}

// WriteTargetWithReset writes a target conductance with optional reset to -Pr before starting.
// reset=true matches the original behavior (start from negative saturation).
func (c *WriteController) WriteTargetWithReset(targetG float64, reset bool) (attempts int, success bool, overshootCount int) {
	return c.writeTarget(targetG, reset)
}

func (c *WriteController) writeTarget(targetG float64, reset bool) (attempts int, success bool, overshootCount int) {
	log := getISPPLogger()

	targetP := ConductanceToPolarization(targetG, c.Material.Gmin, c.Material.Gmax, c.Material.Ps)

	c.VMin = 0.0
	c.VMax = c.MaxVoltage

	currentP := c.Solver.GetState()
	currentG := PolarizationToConductance(currentP, c.Material.Ps, c.Material.Gmin, c.Material.Gmax)

	direction := 1.0
	if reset {
		if targetP < 0 {
			direction = -1.0
		}
	} else if targetG < currentG {
		direction = -1.0
	}

	if reset {
		pr := math.Abs(c.Material.Pr)
		c.Solver.SetState(-direction * pr)
		currentP = c.Solver.GetState()
		currentG = PolarizationToConductance(currentP, c.Material.Ps, c.Material.Gmin, c.Material.Gmax)
	}

	directionLabel := "positive"
	if direction < 0 {
		directionLabel = "negative"
	}

	overshoots := 0
	success = false

	log.Input("WriteTarget", map[string]interface{}{
		"targetG":   targetG,
		"targetP":   targetP,
		"reset":     reset,
		"direction": directionLabel,
		"startP":    currentP,
		"startG":    currentG,
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
			vPulse = direction * (c.VMin + c.VMax) / 2.0
			log.Calculation("WriteTarget", map[string]interface{}{
				"step":   "Predict",
				"vPulse": vPulse,
				"bounds": []float64{c.VMin, c.VMax},
			}, nil)
		} else {
			vPulse = direction * (c.VMin + c.VMax) / 2.0
			log.Calculation("WriteTarget", map[string]interface{}{
				"step":   "BinarySearch",
				"vPulse": vPulse,
				"vMin":   c.VMin,
				"vMax":   c.VMax,
			}, nil)
		}

		eField := vPulse / c.Material.Thickness
		c.applyPulse(eField, c.PulseWidth)

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

		if direction*error < 0 {
			c.VMin = math.Abs(vPulse)
			log.Calculation("WriteTarget", map[string]interface{}{
				"decision": "Undershoot",
				"vPulse":   vPulse,
				"newVMin":  c.VMin,
				"vMax":     c.VMax,
			}, nil)
		} else {
			overshoots++

			resetField := -direction * c.MaxVoltage * 1.5
			eResetField := resetField / c.Material.Thickness
			c.applyPulse(eResetField, c.PulseWidth*2)

			log.Calculation("WriteTarget", map[string]interface{}{
				"decision":    "Overshoot",
				"overshoots":  overshoots,
				"resetField":  resetField,
				"eResetField": eResetField,
			}, nil)

			c.VMax = math.Abs(vPulse)
			c.VMin = 0.0

			c.Solver.SetState(-direction * math.Abs(c.Material.Pr))
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

func (c *WriteController) applyPulse(eField float64, duration float64) {
	if duration <= 0 {
		return
	}
	if c.MaxStep <= 0 || duration <= c.MaxStep {
		c.Solver.Step(eField, duration)
		return
	}
	steps := int(math.Ceil(duration / c.MaxStep))
	if steps < 1 {
		steps = 1
	}
	dt := duration / float64(steps)
	for i := 0; i < steps; i++ {
		c.Solver.Step(eField, dt)
	}
}

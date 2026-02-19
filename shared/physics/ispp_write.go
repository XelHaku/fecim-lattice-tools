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

// WriteController implements Landau-Khalatnikov (L-K) solver-based ISPP for
// physics-accurate ferroelectric write operations. Unlike the waveform-based
// controller in module1-hysteresis/pkg/controller/writer.go, this controller
// operates in voltage domain and uses the L-K ODE solver directly.
//
// Algorithm:
//  1. Predict initial pulse voltage from atanh(P_target/Ps) * thickness
//  2. Apply pulse via L-K integration (subdivided into MaxStep sub-steps)
//  3. Verify resulting conductance against target (Tolerance threshold)
//  4. On undershoot: raise VMin, bisect [VMin, VMax] with crossing-aware bias
//  5. On overshoot: lower VMax, optionally reset to -Pr and re-approach
//
// The crossing-aware bias (0.1-0.3 instead of 0.5) reduces overshoot when the
// target polarization has opposite sign to the current state, since the tanh
// P(E) curve is steepest near E=Ec.
type WriteController struct {
	Solver   *LKSolver
	Material *HZOMaterial

	MaxVoltage    float64 // Maximum programming voltage (V)
	MinVoltage    float64 // Minimum non-zero pulse voltage (V)
	PulseWidth    float64 // Programming pulse width (s); default 10 ns
	Tolerance     float64 // Convergence threshold in conductance units; default 0.01
	MaxIterations int     // Max ISPP pulses before declaring failure; default 20
	MaxStep       float64 // Maximum L-K integration sub-step (s); default 1 ps for stability

	VMin float64 // Lower bound of voltage search (established by undershoot)
	VMax float64 // Upper bound of voltage search (established by overshoot)

	// FailureReason is set when WriteTarget fails (e.g., unreachable due to bounds).
	FailureReason string

	// Optional hooks for headless diagnostics and CSV logging.
	StepFunc  func(E, dt float64) float64
	EventHook func(event WriteEvent)
}

// WriteEvent exposes key ISPP transitions for headless diagnostics.
type WriteEvent struct {
	Phase      string
	Attempt    int
	VPulse     float64
	VMin       float64
	VMax       float64
	CurrentP   float64
	CurrentG   float64
	TargetP    float64
	TargetG    float64
	Error      float64
	Direction  float64
	Overshoots int
	ResetField float64
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
	c.FailureReason = ""

	if c.Material == nil {
		c.FailureReason = "material is nil"
		log.Input("WriteTarget", map[string]interface{}{"status": "Failed", "reason": c.FailureReason})
		c.emitEvent(WriteEvent{Phase: "Failed", Attempt: 0, TargetG: targetG})
		return 0, false, 0
	}
	if targetG < c.Material.Gmin || targetG > c.Material.Gmax {
		c.FailureReason = "target conductance out of bounds"
		log.Input("WriteTarget", map[string]interface{}{"status": "Failed", "reason": c.FailureReason, "targetG": targetG, "gmin": c.Material.Gmin, "gmax": c.Material.Gmax})
		c.emitEvent(WriteEvent{Phase: "Failed", Attempt: 0, TargetG: targetG})
		return 0, false, 0
	}

	targetP := ConductanceToPolarization(targetG, c.Material.Gmin, c.Material.Gmax, c.Material.Ps)

	c.VMin = 0.0
	c.VMax = c.MaxVoltage

	currentP := c.Solver.GetState()
	currentG := PolarizationToConductanceWithParams(currentP, c.Material.Ps, c.Material.Gmin, c.Material.Gmax, ParseConductanceModel(c.Material.ConductanceModel), c.Material.KvT, c.Material.VGSReadV, c.Material.VT0V)
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
		currentG = PolarizationToConductanceWithParams(currentP, c.Material.Ps, c.Material.Gmin, c.Material.Gmax, ParseConductanceModel(c.Material.ConductanceModel), c.Material.KvT, c.Material.VGSReadV, c.Material.VT0V)
	}

	crossingInitial := currentP*targetP < 0
	if crossingInitial {
		if vBound := c.initialPulseBound(targetP); vBound > 0 && vBound < c.VMax {
			c.VMax = vBound
			log.Calculation("WriteTarget", map[string]interface{}{
				"step":     "ClampBounds",
				"crossing": crossingInitial,
				"vMax":     c.VMax,
			}, nil)
		}
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
	c.emitEvent(WriteEvent{
		Phase:     "Start",
		Attempt:   0,
		TargetP:   targetP,
		TargetG:   targetG,
		CurrentP:  currentP,
		CurrentG:  currentG,
		Direction: direction,
	})

	var vPulse float64
	var i int
	for i = 0; i < c.MaxIterations; i++ {
		currentP := c.Solver.GetState()
		currentG := PolarizationToConductanceWithParams(currentP, c.Material.Ps, c.Material.Gmin, c.Material.Gmax, ParseConductanceModel(c.Material.ConductanceModel), c.Material.KvT, c.Material.VGSReadV, c.Material.VT0V)
		crossingNow := currentP*targetP < 0
		log.Calculation("WriteTarget", map[string]interface{}{
			"attempt":  i + 1,
			"currentP": currentP,
			"targetP":  targetP,
		}, nil)

		if i == 0 {
			vGuess := c.initialPulseMagnitude(targetP, currentP)
			if vGuess <= 0 {
				vGuess = (c.VMin + c.VMax) / 2.0
			}
			if vGuess < c.VMin {
				vGuess = c.VMin
			}
			if vGuess > c.VMax {
				vGuess = c.VMax
			}
			vPulse = direction * vGuess
			if c.MinVoltage > 0 {
				absV := math.Abs(vPulse)
				if absV > 0 && absV < c.MinVoltage {
					vPulse = math.Copysign(c.MinVoltage, vPulse)
				}
			}
			log.Calculation("WriteTarget", map[string]interface{}{
				"step":     "Predict",
				"model":    "atanh",
				"crossing": crossingNow,
				"vPulse":   vPulse,
				"vGuess":   vGuess,
				"bounds":   []float64{c.VMin, c.VMax},
			}, nil)
			c.emitEvent(WriteEvent{
				Phase:      "Predict",
				Attempt:    i + 1,
				VPulse:     vPulse,
				VMin:       c.VMin,
				VMax:       c.VMax,
				TargetP:    targetP,
				TargetG:    targetG,
				CurrentP:   currentP,
				CurrentG:   currentG,
				Direction:  direction,
				Overshoots: overshoots,
			})
		} else {
			bias := 0.5
			if crossingNow {
				ratio := math.Abs(targetP / c.Material.Ps)
				bias = 0.1 + 0.2*ratio
				if bias < 0.1 {
					bias = 0.1
				}
				if bias > 0.3 {
					bias = 0.3
				}
			}
			midpoint := c.VMin + bias*(c.VMax-c.VMin)
			vPulse = direction * midpoint
			if c.MinVoltage > 0 {
				absV := math.Abs(vPulse)
				if absV > 0 && absV < c.MinVoltage {
					vPulse = math.Copysign(c.MinVoltage, vPulse)
				}
			}
			log.Calculation("WriteTarget", map[string]interface{}{
				"step":     "BinarySearch",
				"vPulse":   vPulse,
				"vMin":     c.VMin,
				"vMax":     c.VMax,
				"crossing": crossingNow,
				"midpoint": midpoint,
				"bias":     bias,
			}, nil)
			c.emitEvent(WriteEvent{
				Phase:      "BinarySearch",
				Attempt:    i + 1,
				VPulse:     vPulse,
				VMin:       c.VMin,
				VMax:       c.VMax,
				TargetP:    targetP,
				TargetG:    targetG,
				CurrentP:   currentP,
				CurrentG:   currentG,
				Direction:  direction,
				Overshoots: overshoots,
			})
		}

		// Enforce MinVoltage as a guardrail for non-zero pulses.
		if c.MinVoltage > 0 {
			absV := math.Abs(vPulse)
			if absV > 0 && absV < c.MinVoltage {
				vPulse = math.Copysign(c.MinVoltage, vPulse)
			}
		}
		eField := vPulse / c.Material.Thickness
		c.applyPulse(eField, c.PulseWidth)

		log.Calculation("WriteTarget", map[string]interface{}{
			"step":   "WritePulse",
			"vPulse": vPulse,
			"eField": eField,
			"dt":     c.PulseWidth,
		}, nil)

		postP := c.Solver.GetState()
		postG := PolarizationToConductanceWithParams(postP, c.Material.Ps, c.Material.Gmin, c.Material.Gmax, ParseConductanceModel(c.Material.ConductanceModel), c.Material.KvT, c.Material.VGSReadV, c.Material.VT0V)

		error := postG - targetG
		crossingAfter := postP*targetP < 0

		log.Calculation("WriteTarget", map[string]interface{}{
			"step":     "Verify",
			"currentP": postP,
			"currentG": postG,
			"error":    error,
		}, nil)
		c.emitEvent(WriteEvent{
			Phase:      "Verify",
			Attempt:    i + 1,
			VPulse:     vPulse,
			VMin:       c.VMin,
			VMax:       c.VMax,
			TargetP:    targetP,
			TargetG:    targetG,
			CurrentP:   postP,
			CurrentG:   postG,
			Error:      error,
			Direction:  direction,
			Overshoots: overshoots,
		})

		if math.Abs(error) < c.Tolerance {
			success = true
			log.Input("WriteTarget", map[string]interface{}{
				"status":     "Success",
				"attempts":   i + 1,
				"finalG":     postG,
				"finalP":     postP,
				"overshoots": overshoots,
			})
			c.emitEvent(WriteEvent{
				Phase:      "Success",
				Attempt:    i + 1,
				VPulse:     vPulse,
				VMin:       c.VMin,
				VMax:       c.VMax,
				TargetP:    targetP,
				TargetG:    targetG,
				CurrentP:   postP,
				CurrentG:   postG,
				Error:      error,
				Direction:  direction,
				Overshoots: overshoots,
			})
			break
		}

		if direction*error < 0 {
			c.VMin = math.Abs(vPulse)
			if crossingInitial && i == 0 && !crossingAfter && c.Material != nil && c.Material.Ps != 0 {
				ratio := math.Abs(targetP / c.Material.Ps)
				if ratio > 1 {
					ratio = 1
				}
				tighten := 0.2 + 0.4*ratio
				if tighten < 0.2 {
					tighten = 0.2
				}
				if tighten > 0.6 {
					tighten = 0.6
				}
				newVMax := c.VMin + tighten*(c.VMax-c.VMin)
				if newVMax < c.VMax {
					c.VMax = newVMax
					log.Calculation("WriteTarget", map[string]interface{}{
						"step":     "PostCrossTighten",
						"ratio":    ratio,
						"tighten":  tighten,
						"vMin":     c.VMin,
						"vMax":     c.VMax,
						"attempt":  i + 1,
						"crossing": crossingInitial,
					}, nil)
				}
			}
			log.Calculation("WriteTarget", map[string]interface{}{
				"decision": "Undershoot",
				"vPulse":   vPulse,
				"newVMin":  c.VMin,
				"vMax":     c.VMax,
			}, nil)
		} else {
			overshoots++

			// Overshoot handling:
			// - If reset==true (legacy behavior), perform a hard reset pulse.
			// - If reset==false (used in tests and some UI flows), avoid forcing saturation;
			//   just tighten the voltage bracket and continue.
			if reset {
				resetField := -direction * c.MaxVoltage * 1.5
				eResetField := resetField / c.Material.Thickness
				c.emitEvent(WriteEvent{
					Phase:      "Reset",
					Attempt:    i + 1,
					VPulse:     vPulse,
					VMin:       c.VMin,
					VMax:       c.VMax,
					TargetP:    targetP,
					TargetG:    targetG,
					CurrentP:   postP,
					CurrentG:   postG,
					Error:      error,
					Direction:  direction,
					Overshoots: overshoots,
					ResetField: resetField,
				})
				c.applyPulse(eResetField, c.PulseWidth*2)

				log.Calculation("WriteTarget", map[string]interface{}{
					"decision":    "Overshoot",
					"overshoots":  overshoots,
					"resetField":  resetField,
					"eResetField": eResetField,
				}, nil)
				c.emitEvent(WriteEvent{
					Phase:      "Overshoot",
					Attempt:    i + 1,
					VPulse:     vPulse,
					VMin:       c.VMin,
					VMax:       c.VMax,
					TargetP:    targetP,
					TargetG:    targetG,
					CurrentP:   postP,
					CurrentG:   postG,
					Error:      error,
					Direction:  direction,
					Overshoots: overshoots,
					ResetField: resetField,
				})

				c.VMax = math.Abs(vPulse)
				c.VMin = 0.0
				c.Solver.SetState(-direction * math.Abs(c.Material.Pr))
			} else {
				c.VMax = math.Abs(vPulse)
				log.Calculation("WriteTarget", map[string]interface{}{
					"decision":   "OvershootNoReset",
					"overshoots": overshoots,
					"vMax":       c.VMax,
				}, nil)
				c.emitEvent(WriteEvent{
					Phase:      "OvershootNoReset",
					Attempt:    i + 1,
					VPulse:     vPulse,
					VMin:       c.VMin,
					VMax:       c.VMax,
					TargetP:    targetP,
					TargetG:    targetG,
					CurrentP:   postP,
					CurrentG:   postG,
					Error:      error,
					Direction:  direction,
					Overshoots: overshoots,
				})
			}
		}
	}

	if !success {
		finalP := c.Solver.GetState()
		finalG := PolarizationToConductanceWithParams(finalP, c.Material.Ps, c.Material.Gmin, c.Material.Gmax, ParseConductanceModel(c.Material.ConductanceModel), c.Material.KvT, c.Material.VGSReadV, c.Material.VT0V)
		finalErr := finalG - targetG
		c.FailureReason = "max iterations exceeded"
		// If we are still undershooting while already at max voltage, treat as unreachable by bounds.
		if direction*finalErr < 0 && c.MaxVoltage > 0 && c.VMax >= c.MaxVoltage*0.999 {
			c.FailureReason = "max voltage limit reached"
		}
		log.Input("WriteTarget", map[string]interface{}{
			"status":     "Failed",
			"attempts":   c.MaxIterations,
			"overshoots": overshoots,
			"reason":     c.FailureReason,
			"finalG":     finalG,
			"finalP":     finalP,
			"error":      finalErr,
		})
		c.emitEvent(WriteEvent{
			Phase:      "Failed",
			Attempt:    c.MaxIterations,
			TargetP:    targetP,
			TargetG:    targetG,
			CurrentP:   finalP,
			CurrentG:   finalG,
			Error:      finalErr,
			Direction:  direction,
			Overshoots: overshoots,
		})
	}

	return i + 1, success, overshoots
}

func (c *WriteController) initialPulseMagnitude(targetP, currentP float64) float64 {
	vEst := c.initialPulseBound(targetP)
	if vEst <= 0 {
		return 0
	}

	if currentP*targetP < 0 {
		// Crossing hysteresis branches: bias low to reduce overshoot resets.
		absRatio := math.Abs(targetP / c.Material.Ps)
		vEst *= absRatio * absRatio
	}
	if math.IsNaN(vEst) || math.IsInf(vEst, 0) {
		return 0
	}
	return vEst
}

func (c *WriteController) initialPulseBound(targetP float64) float64 {
	if c.Material == nil || c.Material.Ps == 0 || c.Material.Ec == 0 || c.Material.Thickness == 0 {
		return 0
	}

	ratio := targetP / c.Material.Ps
	if ratio > 0.95 {
		ratio = 0.95
	}
	if ratio < -0.95 {
		ratio = -0.95
	}

	eEst := c.Material.Ec * math.Atanh(ratio)
	vEst := math.Abs(eEst * c.Material.Thickness)
	if math.IsNaN(vEst) || math.IsInf(vEst, 0) {
		return 0
	}
	return vEst
}

func (c *WriteController) applyPulse(eField float64, duration float64) {
	if duration <= 0 {
		return
	}
	stepFn := c.StepFunc
	if stepFn == nil {
		stepFn = c.Solver.Step
	}
	if c.MaxStep <= 0 || duration <= c.MaxStep {
		stepFn(eField, duration)
		return
	}
	steps := int(math.Ceil(duration / c.MaxStep))
	if steps < 1 {
		steps = 1
	}
	dt := duration / float64(steps)
	for i := 0; i < steps; i++ {
		stepFn(eField, dt)
	}
}

func (c *WriteController) emitEvent(event WriteEvent) {
	if c.EventHook != nil {
		c.EventHook(event)
	}
}

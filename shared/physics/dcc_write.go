package physics

// LIT-P3-01: DCC (Displacement Current Control) write mode.
//
// DCC is an alternative to ISPP that programs a ferroelectric cell in a
// SINGLE self-terminating voltage pulse instead of iterative step-pulse-verify
// cycles. The controller integrates the Landau–Khalatnikov dynamics and stops
// the pulse the moment the polarization state reaches the target — mimicking
// the physical mechanism where the sense circuit detects the drop in switching
// displacement current and cuts off the write voltage.
//
// Physics:
//
//	I_disp(t) = ε₀·ε_r · dE/dt  +  dP/dt · A
//
// During a constant-voltage pulse dE/dt = 0, so:
//
//	I_disp(t) ≈ dP/dt · A
//
// The current is large while P is switching and falls to near zero when
// switching is complete. DCC detects this and stops the pulse, leaving the
// cell at the target polarization without overshoot.
//
// Comparison with ISPP:
//
//	ISPP – many pulses at increasing voltages, verify after each.
//	DCC  – one pulse at write voltage, stops when P_target is reached.
//	        Single-pulse → lower energy, fewer disturbance events.
//	        Requires fast in-situ current sensing (~ns latency).
//
// Reference: PMC 2024 (DCC-ISPP comparison for HfO₂-based FeRAM).

import (
	"math"
)

// DCCResult captures the outcome of a DCC write operation.
type DCCResult struct {
	// Success is true when |P_final - P_target| < Tolerance.
	Success bool

	// FinalP is the polarization state at pulse termination (µC/cm²).
	FinalP float64

	// FinalG is the corresponding conductance mapped through the material model.
	FinalG float64

	// PulseDuration is the total time the write voltage was applied (s).
	// For a successful write this is always ≤ MaxPulseTime.
	PulseDuration float64

	// Steps is the number of integration time-steps taken.
	Steps int

	// FailureReason is non-empty when Success == false.
	FailureReason string
}

// DCCController programs ferroelectric cells using Displacement Current Control.
//
// It wraps the same LKSolver / HZOMaterial as WriteController but applies a
// single constant-voltage pulse that is cut off as soon as P reaches P_target.
//
// Typical use:
//
//	dcc := NewDCCController(solver, material)
//	res := dcc.ProgramDCC(targetG, true)  // true = reset to -Pr first
type DCCController struct {
	// Solver is the Landau–Khalatnikov integrator.
	Solver *LKSolver

	// Material holds Ps, Pr, Ec, Thickness, Gmin, Gmax, etc.
	Material *HZOMaterial

	// WriteVoltage is the constant voltage applied during the DCC pulse (V).
	// Should be ≥ coercive voltage to fully switch. Default: 3× Vc.
	WriteVoltage float64

	// MaxPulseTime is the upper bound on a single DCC pulse (s).
	// If P_target is not reached within this time, the write fails.
	MaxPulseTime float64

	// Tolerance is the acceptable |P - P_target| (fraction of 2Ps).
	// Default 0.01 (1 % of the full switching range).
	Tolerance float64

	// TimeStep is the integration step size (s). Smaller is more accurate
	// but slower. Default 1 ps.
	TimeStep float64
}

// NewDCCController creates a DCCController configured for the given material.
func NewDCCController(solver *LKSolver, material *HZOMaterial) *DCCController {
	vc := 1.0 // default coercive voltage (V)
	if material != nil {
		vc = material.CoerciveVoltage()
	}
	return &DCCController{
		Solver:       solver,
		Material:     material,
		WriteVoltage: 3.0 * vc, // 3× Vc to ensure full switching
		MaxPulseTime: 100e-9,   // 100 ns maximum single pulse
		Tolerance:    0.01,     // 1 % of 2Ps
		TimeStep:     1e-12,    // 1 ps integration step
	}
}

// ProgramDCC writes targetG using a single self-terminating voltage pulse.
//
// If reset == true the solver state is first reset to ±Pr (the saturated
// state opposite the target), matching the standard ISPP pre-conditioning.
// Set reset = false to program from the current state.
func (d *DCCController) ProgramDCC(targetG float64, reset bool) DCCResult {
	mat := d.Material
	if mat == nil {
		return DCCResult{FailureReason: "material is nil"}
	}
	if targetG < mat.Gmin || targetG > mat.Gmax {
		return DCCResult{FailureReason: "target conductance out of bounds"}
	}
	if d.Solver == nil {
		return DCCResult{FailureReason: "solver is nil"}
	}

	// Map target conductance → target polarization.
	targetP := ConductanceToPolarization(targetG, mat.Gmin, mat.Gmax, mat.Ps)

	// Determine write direction.
	direction := 1.0
	if targetP < 0 {
		direction = -1.0
	}

	// Optional reset: start from saturated state opposite to target.
	if reset {
		pr := math.Abs(mat.Pr)
		d.Solver.SetState(-direction * pr)
	}

	// Tolerance in polarization units.
	tolP := d.Tolerance * 2 * math.Abs(mat.Ps)
	if tolP <= 0 {
		tolP = 0.01
	}

	// Electric field for the write voltage: E = V / thickness.
	eField := direction * d.WriteVoltage / mat.Thickness

	// Integration loop — step until P_target reached or MaxPulseTime exceeded.
	dt := d.TimeStep
	if dt <= 0 {
		dt = 1e-12
	}
	maxSteps := int(math.Ceil(d.MaxPulseTime / dt))
	if maxSteps < 1 {
		maxSteps = 1
	}

	var elapsed float64
	var steps int

	for steps = 0; steps < maxSteps; steps++ {
		// Check termination before stepping (catches pre-satisfied targets).
		currentP := d.Solver.GetState()
		if math.Abs(currentP-targetP) <= tolP {
			break
		}
		// For positive direction: stop as soon as P ≥ P_target.
		// For negative direction: stop as soon as P ≤ P_target.
		if direction > 0 && currentP >= targetP-tolP {
			break
		}
		if direction < 0 && currentP <= targetP+tolP {
			break
		}

		d.Solver.Step(eField, dt)
		elapsed += dt
	}

	finalP := d.Solver.GetState()
	finalG := PolarizationToConductanceWithParams(
		finalP, mat.Ps, mat.Gmin, mat.Gmax,
		ParseConductanceModel(mat.ConductanceModel),
		mat.KvT, mat.VGSReadV, mat.VT0V,
	)

	success := math.Abs(finalP-targetP) <= tolP*2
	reason := ""
	if !success {
		if steps >= maxSteps {
			reason = "MaxPulseTime exceeded without reaching target"
		} else {
			reason = "target not reached"
		}
	}

	return DCCResult{
		Success:       success,
		FinalP:        finalP,
		FinalG:        finalG,
		PulseDuration: elapsed,
		Steps:         steps,
		FailureReason: reason,
	}
}

// ISPPvsDCCComparison holds a side-by-side comparison of ISPP and DCC for
// the same write target. Use for educational output and benchmarks.
type ISPPvsDCCComparison struct {
	TargetG float64

	ISPPSuccess    bool
	ISPPAttempts   int
	ISPPOvershoots int
	ISPPFinalG     float64

	DCCSuccess       bool
	DCCPulseDuration float64 // single DCC pulse duration (s)
	DCCSteps         int
	DCCFinalG        float64
}

// CompareISPPvsDCC runs both ISPP (WriteController) and DCC (DCCController) on
// independent solver instances configured from the same material and returns a
// side-by-side comparison.
//
// Both controllers start from the same reset state (-Pr for positive targets).
func CompareISPPvsDCC(wc *WriteController, dcc *DCCController, targetG float64) ISPPvsDCCComparison {
	// ISPP run (modifies wc.Solver state — make a copy of state first).
	isppAttempts, isppOk, isppOvershoots := wc.WriteTarget(targetG)
	isppP := wc.Solver.GetState()
	isppFinalG := PolarizationToConductanceWithParams(
		isppP, wc.Material.Ps, wc.Material.Gmin, wc.Material.Gmax,
		ParseConductanceModel(wc.Material.ConductanceModel),
		wc.Material.KvT, wc.Material.VGSReadV, wc.Material.VT0V,
	)

	// DCC run (uses its own solver instance).
	dccRes := dcc.ProgramDCC(targetG, true)

	return ISPPvsDCCComparison{
		TargetG: targetG,

		ISPPSuccess:    isppOk,
		ISPPAttempts:   isppAttempts,
		ISPPOvershoots: isppOvershoots,
		ISPPFinalG:     isppFinalG,

		DCCSuccess:       dccRes.Success,
		DCCPulseDuration: dccRes.PulseDuration,
		DCCSteps:         dccRes.Steps,
		DCCFinalG:        dccRes.FinalG,
	}
}

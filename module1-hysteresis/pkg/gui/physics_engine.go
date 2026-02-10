package gui

import (
	"math"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
)

func (a *App) useLKSolver() bool {
	return a != nil && a.physicsEngine == PhysicsLandau
}

func (a *App) effectiveEc() float64 {
	if a == nil || a.material == nil {
		return 0
	}
	if a.physicsEngine == PhysicsPreisach && a.preisach != nil {
		if ec := a.preisach.GetEffectiveEc(); ec != 0 {
			return ec
		}
	}
	return a.material.Ec
}

func (a *App) currentTemperature() float64 {
	if a == nil {
		return 0
	}
	if a.physicsEngine == PhysicsLandau && a.lkSolver != nil {
		return a.lkSolver.Temperature
	}
	if a.physicsEngine == PhysicsPreisach {
		return 300
	}
	return 0
}

func (a *App) effectiveRangeFrac() float64 {
	if a == nil {
		return 1
	}
	if a.wrdRangeFrac > 0 && a.wrdRangeFrac <= 1 {
		return a.wrdRangeFrac
	}
	return 1
}

func (a *App) effectivePsForLevels() float64 {
	if a == nil || a.material == nil || a.material.Ps == 0 {
		return 0
	}
	return a.material.Ps * a.effectiveRangeFrac()
}

func (a *App) lkDefaultPolarization() float64 {
	if a == nil {
		return 0
	}
	if a.material != nil {
		if a.material.Pr != 0 {
			return -math.Abs(a.material.Pr)
		}
		if a.material.Ps != 0 {
			return -math.Abs(a.material.Ps)
		}
	}
	if a.lkSolver != nil && a.lkSolver.PMax > 0 {
		return -math.Abs(a.lkSolver.PMax)
	}
	return 0
}

// setPhysicsEngine switches the active polarization dynamics model.
// This resets history and synchronizes the discrete level to avoid stale state.
func (a *App) setPhysicsEngine(engine PhysicsEngine) {
	if a == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.physicsEngine == engine {
		return
	}

	a.physicsEngine = engine

	switch engine {
	case PhysicsLandau:
		if a.lkSolver == nil {
			a.lkSolver = physics.NewLKSolver()
		}
		if a.material != nil {
			a.lkSolver.ConfigureFromMaterial(a.material)
		}
		if a.preisach != nil {
			a.lkSolver.Temperature = a.preisach.Temperature
		}
		initP := a.polarization
		if math.Abs(initP) < 1e-9 || math.IsNaN(initP) || math.IsInf(initP, 0) {
			initP = a.lkDefaultPolarization()
		}
		a.lkSolver.SetState(initP)
		a.polarization = a.lkSolver.GetState()
		a.lkSolver.Time = 0
		a.lkSolver.UseNLS = false // Disable NLS for deterministic ISPP behavior
		a.wrdSkipPrep = false
		a.needsCalibration = true
		a.calibrated = false
		log.Printf("Physics engine switched to L-K (dynamic)")
	case PhysicsPreisach:
		if a.preisach == nil && a.material != nil {
			a.preisach = ferroelectric.NewPreisachModel(a.material)
		}
		if a.preisach != nil {
			a.preisach.SetTemperature(300) // Fixed room temp for Preisach
			a.preisach.Reset()
			a.polarization = a.preisach.Update(a.electricField)
			a.normalizedP = a.preisach.NormalizedPolarization()
		}
		a.needsCalibration = true
		a.calibrated = false
		a.wrdSkipPrep = true
		log.Printf("Physics engine switched to Preisach (quasi-static)")
	}

	a.syncDiscreteLevelLocked()
	a.resetHistoryLocked()
}

// syncDiscreteLevelLocked updates normalized polarization and discrete level.
// Caller must hold a.mu.
func (a *App) syncDiscreteLevelLocked() {
	if a == nil || a.material == nil {
		return
	}

	if a.physicsEngine == PhysicsLandau {
		if a.material.Ps != 0 {
			a.normalizedP = a.polarization / a.material.Ps
		} else {
			a.normalizedP = 0
		}
		if a.normalizedP > 1 {
			a.normalizedP = 1
		} else if a.normalizedP < -1 {
			a.normalizedP = -1
		}
	}

	maxLevel := a.numLevels - 1
	if maxLevel < 1 {
		maxLevel = 1
	}
	levelNorm := a.normalizedP
	if effPs := a.effectivePsForLevels(); effPs != 0 {
		levelNorm = a.polarization / effPs
	}
	if levelNorm > 1 {
		levelNorm = 1
	} else if levelNorm < -1 {
		levelNorm = -1
	}
	a.discreteLevel = int(math.Round((levelNorm + 1) / 2 * float64(maxLevel)))
	if a.discreteLevel < 0 {
		a.discreteLevel = 0
	}
	if a.discreteLevel > maxLevel {
		a.discreteLevel = maxLevel
	}
}

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
	if a.preisach != nil {
		return a.preisach.Temperature
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
		a.lkSolver.SetState(a.polarization)
		a.lkSolver.Time = 0
		a.needsCalibration = true
		a.calibrated = false
		log.Printf("Physics engine switched to L-K (dynamic)")
	case PhysicsPreisach:
		if a.preisach == nil && a.material != nil {
			a.preisach = ferroelectric.NewPreisachModel(a.material)
		}
		if a.preisach != nil {
			if a.lkSolver != nil {
				a.preisach.SetTemperature(a.lkSolver.Temperature)
			}
			a.preisach.Reset()
			a.polarization = a.preisach.Update(a.electricField)
			a.normalizedP = a.preisach.NormalizedPolarization()
		}
		a.needsCalibration = true
		a.calibrated = false
		log.Printf("Physics engine switched to Preisach (quasi-static)")
	}

	a.syncDiscreteLevelLocked()
	a.eHistory = a.eHistory[:0]
	a.pHistory = a.pHistory[:0]
	a.lastHistorySample = -1
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
	a.discreteLevel = int(math.Round((a.normalizedP + 1) / 2 * float64(maxLevel)))
	if a.discreteLevel < 0 {
		a.discreteLevel = 0
	}
	if a.discreteLevel > maxLevel {
		a.discreteLevel = maxLevel
	}
}

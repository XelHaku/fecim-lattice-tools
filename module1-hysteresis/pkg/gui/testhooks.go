//go:build legacy_fyne

package gui

import (
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/shared/logging"
)

// Test hooks used by parity tests and headless QA. These are intentionally
// minimal wrappers around internal state so tests can drive the WRD/ISPP
// state machine without constructing a real Fyne window.
//
// NOTE: These hooks are safe because they do not start goroutines and they
// follow updatePhysics() locking requirements.

func InitTestLogger() {
	if log == nil {
		log = logging.NewLogger("hysteresis")
	}
}

func (a *App) Lock() {
	if a == nil {
		return
	}
	a.mu.Lock()
}

func (a *App) Unlock() {
	if a == nil {
		return
	}
	a.mu.Unlock()
}

func (a *App) UpdatePhysics(dt float64) time.Duration {
	if a == nil {
		return 0
	}
	return a.updatePhysics(dt, false)
}

func (a *App) AutoMode(enable bool) {
	if a == nil {
		return
	}
	a.autoMode = enable
}

func (a *App) SetWaveform(w WaveformType) {
	if a == nil {
		return
	}
	a.waveform = w
}

func (a *App) SetPhysicsEngine(pe PhysicsEngine) {
	if a == nil {
		return
	}
	a.physicsEngine = pe
}

func (a *App) SetFrequency(freq float64) {
	if a == nil {
		return
	}
	a.frequency = freq
}

func (a *App) SetWrdSkipPrep(skip bool) {
	if a == nil {
		return
	}
	a.wrdSkipPrep = skip
}

func (a *App) SetWrdNextTarget(level int) {
	if a == nil {
		return
	}
	a.wrdNextTargetLevel = level
}

func (a *App) SetWrdPhase(phase int) {
	if a == nil {
		return
	}
	a.wrdPhase = phase
	a.wrdPhaseTimer = 0
}

func (a *App) WrdPhase() int {
	if a == nil {
		return 0
	}
	return a.wrdPhase
}

func (a *App) WrdTargetLevel() int {
	if a == nil {
		return 0
	}
	return a.wrdTargetLevel
}

func (a *App) ElectricField() float64 {
	if a == nil {
		return 0
	}
	return a.electricField
}

func (a *App) DiscreteLevel() int {
	if a == nil {
		return 0
	}
	return a.discreteLevel
}

func (a *App) WriteController() *controller.WriteController {
	if a == nil {
		return nil
	}
	return a.writeController
}

// SetWrdGuardFrac sets the guard band fraction for testing.
func (a *App) SetWrdGuardFrac(frac float64) {
	if a == nil {
		return
	}
	a.wrdGuardFrac = frac
}

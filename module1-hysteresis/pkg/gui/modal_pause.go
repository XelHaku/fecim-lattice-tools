//go:build legacy_fyne

package gui

import sharedwidgets "fecim-lattice-tools/shared/widgets"

// pauseSimulationForModal pauses the simulation while a modal dialog is open and
// returns a resume closure that restores the previous pause state.
//
// This prevents background physics/UI work from competing with expensive modal
// content construction (e.g., pickers, long-form help/equation panels) and avoids
// the "window open => calculations lag" bug.
func (a *App) pauseSimulationForModal() (resume func()) {
	if a == nil {
		return func() {}
	}

	wasPaused := a.paused.Load()
	if !wasPaused {
		a.paused.Store(true)
		if a.pauseBtn != nil {
			sharedwidgets.SafeDo(func() {
				a.pauseBtn.SetText("Resume")
			})
		}
		if a.statusLabel != nil {
			sharedwidgets.SafeDo(func() {
				a.statusLabel.SetText("⏸ Paused (modal)")
			})
		}
	}

	return func() {
		if wasPaused {
			return
		}
		// Only resume if we're still paused (user might have paused independently).
		if a.paused.Load() {
			a.paused.Store(false)
			if a.pauseBtn != nil {
				sharedwidgets.SafeDo(func() {
					a.pauseBtn.SetText("Pause")
				})
			}
			if a.statusLabel != nil {
				sharedwidgets.SafeDo(func() {
					a.statusLabel.SetText("Running…")
				})
			}
		}
	}
}

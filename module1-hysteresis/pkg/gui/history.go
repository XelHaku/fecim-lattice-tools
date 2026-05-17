//go:build legacy_fyne

package gui

// History management uses a ring buffer to avoid slice churn.
// All helpers expect the caller to hold a.mu (read or write as noted).

func (a *App) ensureHistoryCapacityLocked() {
	if a.maxHistory <= 0 {
		a.eHistory = nil
		a.pHistory = nil
		a.historyHead = 0
		a.historySize = 0
		return
	}
	if len(a.eHistory) != a.maxHistory {
		if cap(a.eHistory) >= a.maxHistory && cap(a.pHistory) >= a.maxHistory {
			a.eHistory = a.eHistory[:a.maxHistory]
			a.pHistory = a.pHistory[:a.maxHistory]
		} else {
			a.eHistory = make([]float64, a.maxHistory)
			a.pHistory = make([]float64, a.maxHistory)
		}
		a.historyHead = 0
		a.historySize = 0
	}
}

func (a *App) resetHistoryLocked() {
	a.historyHead = 0
	a.historySize = 0
	a.lastHistorySample = -1
}

func (a *App) appendHistoryLocked(eField float64, polarization float64) {
	if a.maxHistory <= 0 {
		return
	}
	a.ensureHistoryCapacityLocked()
	if a.historySize < a.maxHistory {
		idx := (a.historyHead + a.historySize) % a.maxHistory
		a.eHistory[idx] = eField
		a.pHistory[idx] = polarization
		a.historySize++
		return
	}
	a.eHistory[a.historyHead] = eField
	a.pHistory[a.historyHead] = polarization
	a.historyHead = (a.historyHead + 1) % a.maxHistory
}

func (a *App) historyLengthLocked() int {
	return a.historySize
}

func (a *App) historyAtLocked(i int) (float64, float64) {
	if a.historySize == 0 || a.maxHistory == 0 {
		return 0, 0
	}
	if i < 0 {
		i = 0
	} else if i >= a.historySize {
		i = a.historySize - 1
	}
	idx := (a.historyHead + i) % a.maxHistory
	return a.eHistory[idx], a.pHistory[idx]
}

func (a *App) historySnapshotLocked() ([]float64, []float64) {
	size := a.historySize
	if size == 0 || a.maxHistory == 0 {
		return nil, nil
	}
	eData := make([]float64, size)
	pData := make([]float64, size)
	for i := 0; i < size; i++ {
		idx := (a.historyHead + i) % a.maxHistory
		eData[i] = a.eHistory[idx]
		pData[i] = a.pHistory[idx]
	}
	return eData, pData
}

func (a *App) resizeHistoryLocked(newMax int) {
	if newMax < 0 {
		newMax = 0
	}
	if newMax == a.maxHistory {
		return
	}
	if newMax == 0 {
		a.maxHistory = 0
		a.eHistory = nil
		a.pHistory = nil
		a.historyHead = 0
		a.historySize = 0
		a.lastHistorySample = -1
		return
	}

	// Preserve the most recent samples.
	keep := a.historySize
	if keep > newMax {
		keep = newMax
	}
	newE := make([]float64, newMax)
	newP := make([]float64, newMax)
	if keep > 0 && a.maxHistory > 0 {
		start := a.historySize - keep
		for i := 0; i < keep; i++ {
			idx := (a.historyHead + start + i) % a.maxHistory
			newE[i] = a.eHistory[idx]
			newP[i] = a.pHistory[idx]
		}
	}

	a.maxHistory = newMax
	a.eHistory = newE
	a.pHistory = newP
	a.historyHead = 0
	a.historySize = keep
	if keep == 0 {
		a.lastHistorySample = -1
	}
}

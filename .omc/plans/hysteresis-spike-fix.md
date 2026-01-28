# Hysteresis Plot Spike Fix Plan

## Summary

Fix vertical spikes appearing in the P-E hysteresis plot at transition corners when temperature changes. The root cause is a mismatch between temperature-corrected physics values (Ec(T), Pr(T)) and the plot bounds/markers which continue to use base material values.

## Root Cause Analysis

### Physics Background
- **Ec(T) = Ec0 * (1 - T/Tc)^beta** where Tc=723K (Curie temp), beta=0.5
- At 400K: Ec(T) = 0.67 * Ec0 (significant 33% reduction)
- At 600K: Ec(T) = 0.40 * Ec0 (60% reduction)

### Current Code Issues

1. **controls.go:148** - Material selection uses base Ec for bounds:
   ```go
   a.plot.SetBounds(a.material.Ec*1.5, a.material.Ps*1.2)
   ```
   Should use: `a.preisach.GetEffectiveEc()*1.5`

2. **gui.go:458-460** - Plot creation uses base material values:
   ```go
   a.plot = widgets.NewPEPlot(a.material.Ec*1.5, a.material.Ps*1.2, ...)
   a.plot.SetMaterialParams(a.material.Ec, a.material.Pr)
   ```
   Should use temperature-corrected values from preisach model.

3. **controls.go:334-335** - Temperature change handler updates markers but NOT bounds:
   ```go
   // Update plot markers (outside lock, uses fyne.Do internally)
   a.plot.SetMaterialParams(effEc, effPr)
   ```
   Missing: `a.plot.SetBounds(effEc*1.5, effPr*1.2)` call

4. **History data mismatch** - History points collected at different temperatures have different E-field ranges. When temperature changes and bounds are updated, old history data may appear as spikes because the scale changed.

## Acceptance Criteria

- [ ] **AC1**: Plot bounds update dynamically when temperature slider changes
- [ ] **AC2**: Ec markers (+Ec, -Ec vertical lines) move to correct temperature-corrected positions
- [ ] **AC3**: Pr markers (+Pr, -Pr horizontal lines) move to correct temperature-corrected positions
- [ ] **AC4**: No vertical spikes appear at plot edges during temperature transitions
- [ ] **AC5**: History trail is cleared when temperature changes by more than 25K to prevent scale artifacts
- [ ] **AC6**: Axis tick labels update to reflect new temperature-corrected bounds
- [ ] **AC7**: All existing waveform modes (Manual, Sine, Triangle, Write/Read Demo, Time-Resolved) work correctly after fix

## Implementation Steps

### Task 1: Update Temperature Change Handler (controls.go)
**File:** `<local-path>`
**Lines:** 326-336

**Current code:**
```go
go func() {
    a.mu.Lock()
    a.onTemperatureChanged(v)
    // Get plot markers with temperature-corrected Ec and Pr
    effEc := a.preisach.GetEffectiveEc()
    effPr := a.preisach.GetEffectivePr()
    a.mu.Unlock()

    // Update plot markers (outside lock, uses fyne.Do internally)
    a.plot.SetMaterialParams(effEc, effPr)
}()
```

**Changes needed:**
1. Capture previous temperature BEFORE calling `onTemperatureChanged` using `a.preisach.Temperature`
2. Add `SetBounds()` call after `SetMaterialParams()` using temperature-corrected values
3. Clear history when temperature changes significantly (>25K) to prevent scale artifacts

**New code:**
```go
go func() {
    a.mu.Lock()
    previousTemp := a.preisach.Temperature
    a.onTemperatureChanged(v)
    // Get plot markers with temperature-corrected Ec and Pr
    effEc := a.preisach.GetEffectiveEc()
    effPr := a.preisach.GetEffectivePr()
    currentTemp := a.preisach.Temperature

    // Clear history if temperature changed significantly to prevent scale artifacts
    if math.Abs(currentTemp - previousTemp) > 25 {
        a.eHistory = a.eHistory[:0]
        a.pHistory = a.pHistory[:0]
    }
    a.mu.Unlock()

    // Update plot bounds AND markers with temperature-corrected values
    // Note: SetBounds must be called BEFORE SetMaterialParams (no Refresh call in SetBounds)
    a.plot.SetBounds(effEc*1.5, effPr*1.2)
    a.plot.SetMaterialParams(effEc, effPr)
}()
```

**Note:** The `math` package is already imported at controls.go:5.

### Task 2: Update Material Selection Handler with Hybrid Approach (controls.go)
**File:** `<local-path>`
**Lines:** 145-199 (material selection handler)

**Problem:** When material changes, a new Preisach model is created at line 145 which defaults to 300K. The bounds are set immediately after (lines 148-149) using base material values, BEFORE temperature is restored in the background goroutine (line 186-197).

**Hybrid approach (per Architect recommendation):**
1. BEFORE creating new model: capture `savedTemp := a.preisach.Temperature`
2. AFTER creating new model: immediately call `a.preisach.SetTemperature(savedTemp)`
3. THEN call SetBounds with temperature-corrected values
4. THEN call SetMaterialParams with temperature-corrected values

**Current code (lines 141-149):**
```go
a.mu.Lock()
a.matIndex = idx
a.material = a.materials[idx]
// Use fixed high-resolution grid (50) for physics accuracy, independent of quantization levels
a.preisach = ferroelectric.NewMayergoyzPreisach(a.material, 50)
a.eHistory = a.eHistory[:0]
a.pHistory = a.pHistory[:0]
a.plot.SetBounds(a.material.Ec*1.5, a.material.Ps*1.2)
a.plot.SetMaterialParams(a.material.Ec, a.material.Pr)
```

**New code:**
```go
a.mu.Lock()
a.matIndex = idx
a.material = a.materials[idx]
// Capture current temperature before creating new model
savedTemp := a.preisach.Temperature
// Use fixed high-resolution grid (50) for physics accuracy, independent of quantization levels
a.preisach = ferroelectric.NewMayergoyzPreisach(a.material, 50)
// Immediately restore temperature to the new model
a.preisach.SetTemperature(savedTemp)
a.eHistory = a.eHistory[:0]
a.pHistory = a.pHistory[:0]
// Use temperature-corrected values for plot bounds and markers
// Note: SetBounds ordering matters - call before SetMaterialParams (no Refresh call in SetBounds)
effEc := a.preisach.GetEffectiveEc()
effPr := a.preisach.GetEffectivePr()
a.plot.SetBounds(effEc*1.5, effPr*1.2)
a.plot.SetMaterialParams(effEc, effPr)
```

### Task 3: Update Initial Plot Creation (gui.go)
**File:** `<local-path>`
**Lines:** 458-460

**Current code:**
```go
a.plot = widgets.NewPEPlot(a.material.Ec*1.5, a.material.Ps*1.2, ColorBackground, ColorGrid, ColorAxis, ColorPositive, ColorNegative, ColorWarning)
a.plot.SetMinSize(fyne.NewSize(400, 350))
a.plot.SetMaterialParams(a.material.Ec, a.material.Pr)
```

**Changes needed:**
Use temperature-corrected values from preisach model for initial plot setup.

**New code:**
```go
// Use temperature-corrected values for initial plot setup
effEc := a.preisach.GetEffectiveEc()
effPr := a.preisach.GetEffectivePr()
a.plot = widgets.NewPEPlot(effEc*1.5, effPr*1.2, ColorBackground, ColorGrid, ColorAxis, ColorPositive, ColorNegative, ColorWarning)
a.plot.SetMinSize(fyne.NewSize(400, 350))
a.plot.SetMaterialParams(effEc, effPr)
```

### Task 4: Add Verification Tests
**File:** `<local-path>` (may need to create)

**Tests to add:**
1. Test that plot bounds update correctly when temperature changes
2. Test that history is cleared on significant temperature change
3. Test that Ec/Pr markers are positioned correctly at different temperatures

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Regression in waveform modes | Medium | High | Run all existing tests, manual verification of each mode |
| Thread safety issues with plot updates | Low | Medium | Use fyne.Do() for all UI updates, maintain existing locking patterns |
| Performance impact from frequent bounds updates | Low | Low | SetBounds() is lightweight, only triggers on temperature change |
| Calibration state corruption | Low | High | Temperature check happens inside existing lock, calibration logic unchanged |

## Verification Steps

### Manual Verification Checklist
1. [ ] Launch app: `go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools`
2. [ ] Select Hysteresis module
3. [ ] Set temperature to 300K (default) - verify normal hysteresis loop
4. [ ] Increase temperature to 400K - verify:
   - [ ] No vertical spikes at edges
   - [ ] Ec markers moved inward (lower field)
   - [ ] History trail cleared
   - [ ] Loop shape appropriate for lower Ec
5. [ ] Increase temperature to 500K - verify same behavior
6. [ ] Decrease temperature back to 300K - verify recovery
7. [ ] Test with different waveforms:
   - [ ] Sine Wave
   - [ ] Triangle Wave
   - [ ] Write/Read Demo
   - [ ] Time-Resolved Switching
   - [ ] Manual mode
8. [ ] Change materials while at non-300K temperature - verify bounds update correctly

### Automated Tests
```bash
# Run all module1 tests
go test ./module1-hysteresis/...

# Run with verbose output
go test -v ./module1-hysteresis/pkg/gui/...
```

## Dependencies

- Task 1 is independent (temperature slider handler)
- Task 2 is independent (material selection handler)
- Task 3 is independent (initial plot creation)
- Task 4 (tests) should be done last after all fixes are in place

## Commit Strategy

**Single commit** with message:
```
fix(hysteresis): use temperature-corrected Ec/Pr for plot bounds and markers

- Update plot bounds dynamically when temperature changes
- Clear history trail on significant temperature changes (>25K) to prevent scale artifacts
- Use GetEffectiveEc()/GetEffectivePr() instead of base material values
- Preserve temperature state when material changes (hybrid approach)
- Apply temperature correction during material selection and initial plot creation

Fixes vertical spikes appearing at plot edges during temperature transitions.
```

## Estimated Effort

| Task | Complexity | Time Estimate |
|------|------------|---------------|
| Task 1 | Medium | 15 min |
| Task 2 | Medium | 15 min |
| Task 3 | Low | 5 min |
| Task 4 | Medium | 20 min |
| **Testing & Verification** | - | 20 min |
| **Total** | - | ~75 min |

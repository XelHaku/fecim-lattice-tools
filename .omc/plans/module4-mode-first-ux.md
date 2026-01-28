# Work Plan: Module 4 Mode-First UX Redesign

**Created**: 2026-01-27
**Timeline**: Standard (3-5 days)
**Breaking Changes**: Acceptable

---

## Context

### Original Request
Designer proposed Mode-First UX redesign for Module 4. Architect validated feasibility and recommended phased migration.

### Interview Summary
- Timeline: Standard (3-5 days) with full implementation and testing
- Priority: Focus on critical UX issues first (UX-004, UX-005)
- Breaking changes: Acceptable if needed for better UX
- Testing: Manual GUI testing sufficient

### Research Findings
- `SetDACVoltageForState()` exists at `device_state.go:362-384` - ready for write level slider
- Show/Hide pattern validated in `tab_reference.go:48-60` - use Stack container
- Mode buttons already exist at `tab_unified.go:264-272` but buried in WL selector panel
- Architecture toggle pattern at `tab_unified.go:1336-1412` can be reused

---

## Work Objectives

### Core Objective
Implement Mode-First UX redesign that:
1. Adds write level slider (0-29) for precise cell state targeting
2. Adds persistent input vector entries visible in COMPUTE mode
3. Relocates mode buttons (READ/WRITE/COMPUTE) to top of view
4. Eliminates mode vs action confusion through contextual panels

### Deliverables
1. Write Level Slider component (UX-004)
2. Input Vector Panel with 8 persistent entries (UX-005)
3. Relocated mode buttons at top with contextual panels
4. Updated documentation in `GUI.module4.md`

### Definition of Done
- [ ] Write level slider (0-29) functional and updates DAC voltage via `SetDACVoltageForState()`
- [ ] Input vector entries persist and are visible in COMPUTE mode
- [ ] Mode buttons at top of view with clear visual highlighting
- [ ] Contextual panels show/hide based on mode
- [ ] Manual GUI testing passes (mode switching, write operations, compute operations)
- [ ] No regressions in existing functionality

---

## Guardrails

### Must Have
- Write level slider with 0-29 range (30 levels = FeCIM standard)
- Slider label shows both level AND computed voltage
- Input vector entries update DAC voltages in real-time
- Mode buttons use `widget.HighImportance` pattern (already established)
- All UI updates wrapped in `fyne.Do()` for thread safety

### Must NOT Have
- Hardcoded voltage values (use `SetDACVoltageForState()` which derives from material)
- Blocking operations on main UI thread
- Changes to `device_state.go` core simulation logic (already supports required operations)
- Breaking changes to embedded interface (`BuildContent()`, `Start()`, `Stop()`)

---

## Task Flow and Dependencies

```
Phase 1: Add Mode Panels (Non-Breaking)
  |
  +-> Task 1.1: Add write mode panel with level slider
  |     |
  |     +-> Task 1.2: Add input vector panel for COMPUTE mode
  |           |
  +-----------|
              v
Phase 2: Relocate Mode Buttons
              |
              +-> Task 2.1: Create new top mode bar
              |     |
              |     +-> Task 2.2: Wire up panel Show/Hide logic
              |           |
              +-----------|
                          v
Phase 3: Remove Redundancy
                          |
                          +-> Task 3.1: Remove buried mode buttons from WL selector
                          |     |
                          |     +-> Task 3.2: Clean up action button section
                          |           |
                          +-----------|
                                      v
Phase 4: Documentation & Testing
                                      |
                                      +-> Task 4.1: Update GUI.module4.md
                                      |
                                      +-> Task 4.2: Manual testing checklist
```

---

## Detailed TODOs

### Phase 1: Add Mode Panels (Non-Breaking)

#### Task 1.1: Add Write Mode Panel with Level Slider
**File**: `module4-circuits/pkg/gui/tab_unified.go`
**Complexity**: Medium
**Estimated Time**: 2-3 hours

**Changes**:

1. Add new fields to `CircuitsApp` in `app.go` (lines ~145-152):
   ```go
   // Write mode panel (Mode-First UX)
   writeModePanel       *fyne.Container
   writeLevelSlider     *widget.Slider
   writeLevelValueLabel *widget.Label
   writeVoltageLabel    *widget.Label
   ```

2. Create `createWriteModePanel()` function in `tab_unified.go` (insert after line 438):
   ```go
   func (ca *CircuitsApp) createWriteModePanel() fyne.CanvasObject {
       // Slider: 0-29 (30 levels)
       ca.writeLevelSlider = widget.NewSlider(0, 29)
       ca.writeLevelSlider.Step = 1
       ca.writeLevelSlider.OnChanged = func(v float64) {
           ca.onWriteLevelChanged(int(v))
       }

       ca.writeLevelValueLabel = widget.NewLabel("Level: 0")
       ca.writeVoltageLabel = widget.NewLabel("Voltage: 1.00V")

       // Layout: Label | Slider | Value | Voltage
       return container.NewVBox(
           widget.NewLabel("Target Write Level:"),
           container.NewBorder(nil, nil,
               ca.writeLevelValueLabel,
               ca.writeVoltageLabel,
               ca.writeLevelSlider,
           ),
       )
   }
   ```

3. Create `onWriteLevelChanged()` handler:
   ```go
   func (ca *CircuitsApp) onWriteLevelChanged(level int) {
       selectedCol := ca.deviceState.GetSelectedCol()
       ca.deviceState.SetDACVoltageForState(selectedCol, level)

       voltage := ca.deviceState.GetDACVoltage(selectedCol)

       fyne.Do(func() {
           ca.writeLevelValueLabel.SetText(fmt.Sprintf("Level: %d", level))
           ca.writeVoltageLabel.SetText(fmt.Sprintf("Voltage: %.2fV", voltage))
       })

       ca.recomputeAndRefresh()
   }
   ```

**Acceptance Criteria**:
- [ ] Slider range is 0-29 (30 discrete levels)
- [ ] Slider updates call `SetDACVoltageForState()`
- [ ] Label shows both level number and computed voltage
- [ ] Panel is initially hidden (shown only in WRITE mode)

---

#### Task 1.2: Add Input Vector Panel for COMPUTE Mode
**File**: `module4-circuits/pkg/gui/tab_unified.go`
**Complexity**: Medium
**Estimated Time**: 2-3 hours

**Changes**:

1. Add new fields to `CircuitsApp` in `app.go`:
   ```go
   // Compute mode panel (Mode-First UX)
   computeModePanel     *fyne.Container
   inputVectorEntries   []*widget.Entry
   inputVectorLabels    []*widget.Label
   ```

2. Create `createComputeModePanel()` function:
   ```go
   func (ca *CircuitsApp) createComputeModePanel() fyne.CanvasObject {
       maxCols := min(8, ca.arrayCols)
       ca.inputVectorEntries = make([]*widget.Entry, maxCols)
       ca.inputVectorLabels = make([]*widget.Label, maxCols)

       entries := container.NewHBox()
       for i := 0; i < maxCols; i++ {
           idx := i
           entry := widget.NewEntry()
           entry.SetPlaceHolder("0")
           entry.OnChanged = func(s string) {
               ca.onInputVectorChanged(idx, s)
           }
           ca.inputVectorEntries[i] = entry

           label := widget.NewLabel(fmt.Sprintf("x%d", i))
           ca.inputVectorLabels[i] = label

           col := container.NewVBox(label, entry)
           entries.Add(col)
       }

       randomBtn := widget.NewButton("Random", func() {
           ca.randomizeInputVector()
       })

       return container.NewVBox(
           widget.NewLabel("Input Vector (0-255):"),
           entries,
           randomBtn,
       )
   }
   ```

3. Create `onInputVectorChanged()` handler:
   ```go
   func (ca *CircuitsApp) onInputVectorChanged(col int, valueStr string) {
       value, err := strconv.Atoi(valueStr)
       if err != nil {
           return
       }
       if value < 0 {
           value = 0
       }
       if value > 255 {
           value = 255
       }

       ca.mu.Lock()
       ca.inputVector[col] = value
       ca.mu.Unlock()

       // Convert to DAC voltage (0-255 -> read range)
       params := make([]float64, len(ca.inputVector))
       for i, v := range ca.inputVector {
           params[i] = float64(v)
       }
       ca.deviceState.SetDACPreset(DACInputVector, params...)
       ca.recomputeAndRefresh()
   }
   ```

4. Create `randomizeInputVector()`:
   ```go
   func (ca *CircuitsApp) randomizeInputVector() {
       ca.mu.Lock()
       for i := range ca.inputVector {
           ca.inputVector[i] = rand.Intn(256)
       }
       ca.mu.Unlock()

       // Update entries
       fyne.Do(func() {
           for i, entry := range ca.inputVectorEntries {
               if entry != nil && i < len(ca.inputVector) {
                   entry.SetText(strconv.Itoa(ca.inputVector[i]))
               }
           }
       })

       // Apply to DAC
       params := make([]float64, len(ca.inputVector))
       for i, v := range ca.inputVector {
           params[i] = float64(v)
       }
       ca.deviceState.SetDACPreset(DACInputVector, params...)
       ca.recomputeAndRefresh()
   }
   ```

**Acceptance Criteria**:
- [ ] 8 persistent entry fields for input vector values (0-255)
- [ ] Entries update DAC voltages in real-time via `SetDACPreset(DACInputVector, ...)`
- [ ] Random button populates all entries with random values
- [ ] Panel is initially hidden (shown only in COMPUTE mode)

---

### Phase 2: Relocate Mode Buttons

#### Task 2.1: Create New Top Mode Bar
**File**: `module4-circuits/pkg/gui/tab_unified.go`
**Complexity**: Medium
**Estimated Time**: 1-2 hours

**Changes**:

1. Create `createModeBar()` function (replaces inline mode buttons in `createWLSelector()`):
   ```go
   func (ca *CircuitsApp) createModeBar() fyne.CanvasObject {
       ca.modeReadBtn = widget.NewButton("READ", func() {
           ca.setOperationMode(OpModeRead)
       })
       ca.modeWriteBtn = widget.NewButton("WRITE", func() {
           ca.setOperationMode(OpModeWrite)
       })
       ca.modeComputeBtn = widget.NewButton("COMPUTE", func() {
           ca.setOperationMode(OpModeCompute)
       })

       // Set initial highlight
       ca.modeReadBtn.Importance = widget.HighImportance

       return container.NewHBox(
           widget.NewLabel("Mode:"),
           ca.modeReadBtn,
           ca.modeWriteBtn,
           ca.modeComputeBtn,
           layout.NewSpacer(),
       )
   }
   ```

2. Modify `createUnifiedView()` (line 29-61) to include mode bar at top:
   ```go
   func (ca *CircuitsApp) createUnifiedView() fyne.CanvasObject {
       // ... existing initialization ...

       // NEW: Mode bar at top
       modeBar := ca.createModeBar()

       // NEW: Mode panels (initially hidden)
       ca.writeModePanel = container.NewVBox(ca.createWriteModePanel())
       ca.computeModePanel = container.NewVBox(ca.createComputeModePanel())
       ca.writeModePanel.Hide()
       ca.computeModePanel.Hide()

       modePanelStack := container.NewStack(ca.writeModePanel, ca.computeModePanel)

       // ... rest of layout with modeBar and modePanelStack in top section ...
   }
   ```

**Acceptance Criteria**:
- [ ] Mode buttons appear at top of OPERATIONS view
- [ ] Mode buttons use existing highlighting pattern (`widget.HighImportance`)
- [ ] Mode bar is horizontally laid out with spacer

---

#### Task 2.2: Wire Up Panel Show/Hide Logic
**File**: `module4-circuits/pkg/gui/tab_unified.go`
**Complexity**: Low
**Estimated Time**: 1 hour

**Changes**:

1. Modify `setOperationMode()` (line 295-328) to show/hide panels:
   ```go
   func (ca *CircuitsApp) setOperationMode(mode OpMode) {
       // ... existing mode logic ...

       // Show/hide mode panels
       ca.updateModePanels(mode)

       // ... rest of existing logic ...
   }
   ```

2. Create `updateModePanels()`:
   ```go
   func (ca *CircuitsApp) updateModePanels(mode OpMode) {
       fyne.Do(func() {
           // Hide all panels first
           if ca.writeModePanel != nil {
               ca.writeModePanel.Hide()
           }
           if ca.computeModePanel != nil {
               ca.computeModePanel.Hide()
           }

           // Show relevant panel
           switch mode {
           case OpModeWrite:
               if ca.writeModePanel != nil {
                   ca.writeModePanel.Show()
               }
           case OpModeCompute:
               if ca.computeModePanel != nil {
                   ca.computeModePanel.Show()
               }
           // OpModeRead: no special panel needed
           }
       })
   }
   ```

**Acceptance Criteria**:
- [ ] WRITE mode shows write level slider panel
- [ ] COMPUTE mode shows input vector panel
- [ ] READ mode shows no extra panel (clean view)
- [ ] Switching modes correctly hides/shows panels

---

### Phase 3: Remove Redundancy

#### Task 3.1: Remove Buried Mode Buttons from WL Selector
**File**: `module4-circuits/pkg/gui/tab_unified.go`
**Complexity**: Low
**Estimated Time**: 30 minutes

**Changes**:

1. Modify `createWLSelector()` (line 228-289) to remove mode buttons section:
   - Remove lines 262-283 (modeReadBtn, modeWriteBtn, modeComputeBtn creation)
   - Remove lines 277-283 (modeButtons VBox)
   - Return only checkboxes container

**Before** (line 285-288):
```go
return container.NewVBox(
    checkboxes,
    modeButtons,
)
```

**After**:
```go
return container.NewVBox(
    checkboxes,
)
```

**Acceptance Criteria**:
- [ ] Mode buttons no longer appear in WL selector panel
- [ ] WL checkboxes still function correctly
- [ ] No orphaned button references

---

#### Task 3.2: Clean Up Action Button Section
**File**: `module4-circuits/pkg/gui/tab_unified.go`
**Complexity**: Low
**Estimated Time**: 30 minutes

**Changes**:

1. Review `createUnifiedActionSection()` (line 401-438) for redundancy:
   - "Write Cell" button stays (action for WRITE mode)
   - "Read/Sense" button stays (action for READ mode)
   - "Compute MVM" button stays (action for COMPUTE mode)
   - Consider: Should these be contextual (only show relevant action)?

2. Optional enhancement: Make action buttons contextual
   ```go
   func (ca *CircuitsApp) updateActionButtons(mode OpMode) {
       fyne.Do(func() {
           // Show/hide based on mode
           // Or change emphasis based on mode
       })
   }
   ```

**Acceptance Criteria**:
- [ ] No duplicate functionality between mode buttons and action buttons
- [ ] Action buttons clear in purpose (execute operation, not change mode)
- [ ] Clean visual separation between mode selection and action execution

---

### Phase 4: Documentation & Testing

#### Task 4.1: Update GUI.module4.md
**File**: `docs/development/GUI/GUI.module4.md`
**Complexity**: Low
**Estimated Time**: 30 minutes

**Changes**:

1. Add UX-004 to Fixed Bugs section
2. Add UX-005 to Fixed Bugs section
3. Update OPERATIONS View layout diagram
4. Document new components:
   - `writeModePanel` / `writeLevelSlider`
   - `computeModePanel` / `inputVectorEntries`
   - `createModeBar()` function
5. Update Data Flow table

**Acceptance Criteria**:
- [ ] All new components documented
- [ ] Layout diagram reflects new structure
- [ ] UX issues marked as fixed

---

#### Task 4.2: Manual Testing Checklist
**Complexity**: Low
**Estimated Time**: 1-2 hours

**Test Cases**:

1. **Mode Switching**:
   - [ ] Click READ -> mode buttons highlight correctly
   - [ ] Click WRITE -> write panel appears, slider visible
   - [ ] Click COMPUTE -> compute panel appears, entries visible
   - [ ] Rapid mode switching works without glitches

2. **Write Level Slider (UX-004)**:
   - [ ] Slider moves from 0 to 29 (30 steps)
   - [ ] Level label updates in real-time
   - [ ] Voltage label shows material-derived voltage
   - [ ] Cell programming uses slider value

3. **Input Vector Entries (UX-005)**:
   - [ ] 8 entry fields visible in COMPUTE mode
   - [ ] Typing values updates DAC voltages
   - [ ] Values persist when switching away and back
   - [ ] Random button fills all entries
   - [ ] MVM computation uses entry values

4. **Architecture Toggle**:
   - [ ] PASSIVE/1T1R/2T1R still works
   - [ ] Mode panels work with all architectures

5. **Material Selection**:
   - [ ] Changing material updates voltage labels in write panel
   - [ ] Slider voltage calculation uses new material

---

## Commit Strategy

### Commit 1: Add mode panels (Phase 1)
```
feat(gui): add write level slider and input vector panels for Mode-First UX

- Add writeModePanel with 0-29 level slider
- Add computeModePanel with 8 input vector entries
- Wire SetDACVoltageForState() for slider
- Wire SetDACPreset(DACInputVector, ...) for entries

Addresses UX-004 and UX-005
```

### Commit 2: Relocate mode buttons (Phase 2)
```
feat(gui): relocate mode buttons to top bar with panel Show/Hide

- Create createModeBar() for top-level mode selection
- Add updateModePanels() for contextual panel visibility
- Mode buttons now prominent at top of view

Addresses UX-006 and UX-007
```

### Commit 3: Remove redundancy (Phase 3)
```
refactor(gui): remove duplicate mode buttons from WL selector

- Remove mode buttons from createWLSelector()
- Clean up action button section
- Clearer separation: modes (top) vs actions (bottom)

Addresses UX-008 through UX-011
```

### Commit 4: Documentation (Phase 4)
```
docs(gui): update Module 4 documentation for Mode-First UX

- Update GUI.module4.md with new components
- Mark UX-004 through UX-011 as fixed
- Update layout diagrams
```

---

## Success Criteria

### Functional
- [ ] Write level slider controls cell programming (0-29 levels)
- [ ] Input vector entries control MVM computation (8 values, 0-255)
- [ ] Mode buttons at top provide clear mode selection
- [ ] Contextual panels show/hide correctly per mode

### UX
- [ ] UX-004: Write level slider implemented and working
- [ ] UX-005: Persistent input vector entries implemented
- [ ] UX-006: Clear separation between mode (state) and action (execution)
- [ ] UX-007-011: Redundancy eliminated, layout improved

### Technical
- [ ] No blocking operations on main UI thread
- [ ] All UI updates wrapped in `fyne.Do()`
- [ ] No hardcoded voltage values
- [ ] Embedded interface unchanged

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Fyne layout glitches | Low | Medium | Use Stack container pattern from tab_reference.go |
| Thread safety issues | Low | High | All updates via fyne.Do(), existing pattern |
| Material voltage mismatch | Low | Medium | SetDACVoltageForState already handles this |
| Breaking embedded interface | Low | High | Only modify internal view structure |

---

## Files Modified

| File | Changes |
|------|---------|
| `module4-circuits/pkg/gui/app.go` | Add 6 new fields for panels and widgets |
| `module4-circuits/pkg/gui/tab_unified.go` | Add 4 new functions, modify 3 existing functions |
| `docs/development/GUI/GUI.module4.md` | Update documentation |

## Files NOT Modified

| File | Reason |
|------|--------|
| `device_state.go` | Already has required methods (`SetDACVoltageForState`) |
| `embedded.go` | Interface unchanged |
| `tab_comparison.go` | Unrelated view |
| `tab_reference.go` | Unrelated view |

---
Module: module1-hysteresis
Name: Hysteresis Visualizer
Entry: cmd/hysteresis/main.go
Package: fecim-lattice-tools/module1-hysteresis/pkg/gui
Last Updated: 2026-02-04
---

Conventions:
  - File paths are relative to module1-hysteresis unless noted
  - Widget types refer to Fyne (`widget.*`, `container.*`, `canvas.*`) or shared widgets
  - Bindings list event handlers or UI update calls impacting the component

Bugs:
  - [x] BUG-M1-001: UI updates from goroutine without fyne.Do() wrapper in saveDebugLog (gui.go:279) - Low risk: only file I/O
  - [x] BUG-M1-002: Potential nil pointer access in initDebugLog when material not set - FIXED: has nil check
  - [x] BUG-M1-003: Missing mutex protection when accessing a.material in createUI (gui.go:331) - Low risk: init-time only
  - [x] BUG-M1-004: Animation loop refresh rate may cause drift without vsync sync (simulation.go:14)
  - [x] BUG-M1-005: Slider value setting inside mutex lock in keyboard.go - FIXED: removed unnecessary mutex
  - [x] BUG-M1-006: LevelIndicator time-based pulse may not refresh properly - OK: simulation refreshes at 60 FPS

UX Improvements (2026-01-27):
  - [x] UX-M1-001: LevelIndicator now shows "CLICK TO SET" / "AUTO" mode indicator at bottom
  - [x] UX-M1-002: PhaseIndicator widget shows state machine progress (RESET→SETTLE→WRITE→HOLD→READ→VERIFY)
  - [x] UX-M1-003: eFieldSlider shows "MANUAL" / "AUTO" label to indicate control mode
  - [x] UX-M1-004: Metrics dashboard shows live Ec(T), Pr(T), Squareness, Switched %
  - [x] UX-M1-005: Ctrl+E export shortcut with status feedback
  - [x] UX-M1-006: Time-Resolved Switching mode for KAI dynamics visualization

Screens:
  - name: MainWindow
    title: "FeCIM Hysteresis Visualizer - Demo 1"
    size: 1280x900
    Layout:
      - Container: Border
        top: null
        bottom: StatusBar
        left: null
        right: null
        center: AdaptiveLayout
        Components:
          - name: StatusBar
            type: container.HBox
            purpose: Display current status text
            file: gui.go:413-418
            state: [statusLabel]
            bindings: [statusLabel.SetText]

          - name: AdaptiveLayout
            type: sharedwidgets.AdaptiveLayout
            purpose: Responsive 3-column layout with mobile fallback
            file: gui.go:429-445
            state: [zones, tabLabels]
            bindings: [zones[0]=leftColumn, zones[1]=plotAndLevel, zones[2]=rightColumn]
            Layout:
              desktop:
                - HSplit:
                    left: leftColumn (22%)
                    right:
                      - HSplit:
                          left: plotAndLevel (75%)
                          right: rightColumn (25%)
              mobile:
                - Tabs: [Info, P-E Plot, Controls]

            Components:
              - name: leftColumn
                type: container.Border
                purpose: Cell visualization and info panels
                file: gui.go:386-393
                Layout:
                  top:
                    - Container: VBox
                      Components:
                        - name: cellHeader
                          type: container.VBox
                          Components:
                            - name: cellTitle
                              type: canvas.Text
                              purpose: Display "MEMORY CELL" title
                              file: gui.go:364-366
                              state: [text="MEMORY CELL", color=cyan, size=16]
                            - name: cellUnderline
                              type: canvas.Rectangle
                              purpose: Visual separator under title
                              file: gui.go:369-370
                              state: [color=cyan, height=2]
                        - name: cellViz
                          type: widgets.CellVisualizer
                          purpose: Visual representation of ferroelectric cell
                          file: gui.go (createUI), widgets/cell.go:17-241
                          state: [level (0-29), minSize=160x180]
                          bindings: [SetLevel, Refresh]
                  center:
                    - Container: VSplit
                      purpose: Info stack (scrollable) above Log panel
                      Components:
                        - top: container.VScroll (infoStack)
                        - bottom: logPanel (padded)
                          Components:
                            - name: stateGrid
                              type: container.GridWithColumns(2)
                              state: [pLabel, levelLabel]
                            - name: modeIndicator
                              type: widgets.ModeIndicator
                              purpose: Show WRITE/READ mode with colored box
                              file: widgets/mode.go:14-152, info.go:19-20
                              state: [isWrite bool, minSize=180x60]
                              bindings: [SetWrite, Refresh]
                            - name: phaseIndicator
                              type: widgets.PhaseIndicator
                              purpose: Show state machine phase progress (RESET→SETTLE→WRITE→HOLD→READ→VERIFY)
                              file: widgets/phase.go, info.go:25-27
                              state: [phase int, mode string ("wrd" or "manual"), minSize=140x50]
                              bindings: [SetPhase(phase, mode)]
                            - name: matParams
                              type: widget.Label
                              purpose: Display material parameters
                              file: info.go:28-35
                              state: [text showing Pr, Ps, Ec, Endurance]
                            - name: fatigueRow
                              type: container.HBox
                              purpose: Display cycling stats
                              file: info.go:42-46
                              state: [cyclesLabel, wakeupLabel, fatigueLabel]
                        - name: slidePanel
                          type: widget.Label
                          purpose: Educational explanation based on mode
                          file: info.go:57-62
                          state: [slideText with wrapping]
                          bindings: [SetText from getSlideText]
                        - name: logPanel
                          type: container.VBox
                          purpose: Memory operations log (scrollable)
                          file: info.go:createLogPanel
                          state: [logText with wrapping, toggle]
                          bindings: [SetText from getLogText]
                        - name: metricsGrid
                          type: container.GridWithColumns(2)
                          purpose: Real-time temperature-corrected metrics
                          file: info.go:68-84, gui.go:151-154
                          state: [effEcLabel, effPrLabel, squarenessLabel, switchedLabel]
                          bindings:
                            - effEcLabel: Shows "Ec(T): X.XX MV/cm" (temperature-corrected coercive field)
                            - effPrLabel: Shows "Pr(T): XX.X µC/cm²" (temperature-corrected remnant polarization)
                            - squarenessLabel: Shows "Squareness: 0.XX" (Pr/Ps ratio, 0-1)
                            - switchedLabel: Shows "Switched: XX%" (hysteron switching fraction)
                          updates: simulation.go:1266-1277 via fyne.Do()

              - name: plotAndLevel
                type: container.Border
                purpose: P-E plot with level indicator
                file: gui.go:406-410
                Layout:
                  left: null
                  right: levelIndicator
                  center: plot
                Components:
                  - name: plot
                    type: widgets.PEPlot
                    purpose: Display hysteresis P-E curve
                    file: gui.go (createUI), widgets/peplot.go:16-398
                    state:
                      - eData: []float64 (history)
                      - pData: []float64 (history)
                      - currentE: float64
                      - currentP: float64
                      - eMax: float64 (bounds)
                      - pMax: float64 (bounds)
                      - minSize: 360x300
                    bindings:
                      - SetBounds(eMax, pMax)
                      - SetData(eHist, pHist, currentE, currentP)
                      - Refresh

                  - name: levelIndicator
                    type: widgets.LevelIndicator
                    purpose: Display 30-level state, clickable in Manual mode
                    file: gui.go:335-349, widgets/level.go:18-316
                    state:
                      - level: int (0-29)
                      - targetLevel: int (for highlighting)
                      - highlightTarget: bool
                      - interactive: bool (shows CLICK TO SET vs AUTO indicator)
                      - minSize: 70x300
                      - OnLevelClicked: func(targetLevel int)
                    bindings:
                      - SetLevel(level)
                      - SetTargetLevel(level, highlight)
                      - SetInteractive(bool) - controls clickability indicator
                      - Tapped(event) -> OnLevelClicked callback
                      - Refresh
                    bugs: [BUG-M1-006]

              - name: rightColumn
                type: container.Border
                purpose: Controls panel (scrollable) with fixed min width for responsive layouts
                file: gui.go:createUI
                Components:
                  - name: controlsPanel
                    type: container.VBox
                    purpose: User controls for simulation
                    file: controls.go:14-203
                    Components:
                      - name: materialSelect
                        type: widget.Select
                        purpose: Choose material preset
                        file: controls.go:78-100
                        state: [selected="Default HZO"]
                        bindings:
                          - OnChanged -> switch material, reset Preisach, clear history

                      - name: waveformSelect
                        type: widget.Select
                        purpose: Choose waveform mode
                        file: controls.go:28-76
                        state: [selected="Sine Wave"]
                        options:
                          - "Manual"
                          - "Sine Wave"
                          - "Triangle Wave"
                          - "Write/Read Demo"
                          - "Time-Resolved Switching" (WaveformTimeResolved = 4)
                        bindings:
                          - OnChanged -> switch waveform, toggle autoMode
                          - Manual: enable eFieldSlider
                          - Auto modes: disable eFieldSlider
                          - Write/Read Demo: reset demo state, init debug log
                          - Time-Resolved: shows KAI switching dynamics over 100ns with 100 steps

                      - name: eFieldLabel
                        type: widget.Label
                        purpose: Display current E-field value
                        file: controls.go:29
                        state: [text="E: 0.00 MV/cm"]

                      - name: eFieldModeLabel
                        type: widget.Label
                        purpose: Show whether slider is in MANUAL or AUTO mode
                        file: controls.go:31-32
                        state: [text="AUTO" or "MANUAL", italic style]
                        bindings:
                          - SetText updated on waveform change

                      - name: eFieldSlider
                        type: widget.Slider
                        purpose: Manual E-field control
                        file: controls.go:16-25
                        state: [range=-2 to 2, step=0.01, value=0]
                        bindings:
                          - OnChanged -> set a.electricField (only in Manual mode)
                        bugs: [BUG-M1-005]

                      - name: freqSlider
                        type: widget.Slider
                        purpose: Control waveform frequency
                        file: controls.go:138-152
                        state: [range=0.01 to 1.0 Hz, step=0.01, value=0.5]
                        bindings:
                          - OnChanged -> set a.frequency, clear history

                      - name: tempSlider
                        type: widget.Slider
                        purpose: Control temperature for Preisach model
                        file: controls.go:155-168
                        state: [range=200 to 700 K, step=25, value=300]
                        bindings:
                          - OnChanged -> preisach.SetTemperature(v)

                      - name: trailSlider
                        type: widget.Slider
                        purpose: Control history trail length
                        file: controls.go:171-186
                        state: [range=50 to 2000, step=50, value=500]
                        bindings:
                          - OnChanged -> set a.maxHistory, trim history if needed

                      - name: pauseBtn
                        type: widget.Button
                        purpose: Pause/Resume simulation
                        file: controls.go:102-112
                        state: [text="Pause"]
                        bindings:
                          - OnTapped -> toggle a.paused, update button text

                      - name: resetBtn
                        type: widget.Button
                        purpose: Reset simulation state
                        file: controls.go:114-128
                        bindings:
                          - OnTapped -> reset Preisach, clear fields, clear history

                      - name: eli5Btn
                        type: widget.Button
                        purpose: Show ELI5 hysteresis explanation dialog
                        file: controls.go:131-135
                        bindings:
                          - OnTapped -> showELI5Dialog()

DataFlow:
  - trigger: simulationLoop tick (60 FPS)
    source: simulation.go:13-394
    updates:
      - a.simTime (with modulo wrap at 1000s)
      - a.electricField (based on waveform mode)
      - a.polarization (via preisach.Update)
      - a.discreteLevel (quantized to 0-29)
      - eHistory, pHistory (with maxHistory limit)
      - wrdPhase, wrdPhaseTimer (Write/Read Demo state machine)
      - wrdCycleEnergy (E·dP integration)
      - timeResAnimating, timeResIndex (Time-Resolved mode state)
      - timeResDataTimes, timeResDataPols, timeResDataSwitch (KAI switching data arrays)
    file: simulation.go:13-394
    state_fields:
      Time-Resolved mode:
        - timeResAnimating: bool - animation active flag
        - timeResDataTimes: []float64 - time steps (0-100ns)
        - timeResDataPols: []float64 - polarization evolution
        - timeResDataSwitch: []float64 - switched fraction over time
        - timeResIndex: int - current animation frame (0-99)

  - trigger: updateUI call
    source: simulation.go:392
    updates:
      - All UI widgets via fyne.Do wrapper
      - eFieldLabel, pLabel, levelLabel text
      - cyclesLabel, wakeupLabel, fatigueLabel text
      - modeIndicator.SetWrite(isWrite)
      - eFieldSlider.SetValue (auto modes only)
      - statusLabel text (mode-dependent)
      - slideText (via getSlideText)
      - logText (via getLogText)
      - plot.SetData, plot.Refresh
      - levelIndicator.SetLevel, SetTargetLevel, Refresh
      - cellViz.SetLevel, cellViz.Refresh
    file: simulation.go:397-587

  - trigger: Material selection
    source: controls.go:81
    updates:
      - a.material, a.matIndex
      - Create new preisach model
      - Clear history (eHistory, pHistory)
      - Update plot bounds
    file: controls.go:80-99

  - trigger: Waveform selection
    source: controls.go:31
    updates:
      - a.waveform, a.autoMode
      - Enable/disable eFieldSlider
      - Reset Write/Read demo state
      - Initialize debug log
    file: controls.go:30-75

  - trigger: Level click (Manual mode)
    source: level.go:68
    updates:
      - a.manualTargetLevel
      - a.manualAnimating = true
      - a.manualPhase = 1 (start WRITE phase)
      - a.manualPhaseTime = 0
      - Add log entry
    file: level.go:67-112, gui.go:338-349

  - trigger: Keyboard shortcuts
    source: keyboard.go:11-137
    updates:
      - E/D: eFieldSlider value (Manual mode)
      - T/G: preisach.SetTemperature
      - F/V: a.frequency, clear history
      - W: cycle waveform (via waveformSelect)
      - Space: toggle a.paused
      - R: reset simulation
      - /: show keyboard help dialog
      - Ctrl+E: export P-E data to JSON and CSV (export.go:19-30)
    file: keyboard.go:11-175
    bugs: [BUG-M1-005]

  - trigger: Ctrl+E export shortcut
    source: keyboard.go:19-30, export.go
    updates:
      - exports eHistory, pHistory to JSON and CSV
      - saves to data/ directory with timestamp
      - filename format: pe-data-YYYY-MM-DDTHH-MM-SS.{json,csv}
      - shows status feedback via statusLabel
    file: export.go (new)

  - trigger: Temperature slider change (enhanced)
    source: controls.go:155-168
    updates:
      - preisach.SetTemperature(v)
      - effEcLabel: real-time Ec(T) calculation
      - effPrLabel: real-time Pr(T) calculation
      - squarenessLabel: updated Pr/Ps ratio
      - switchedLabel: hysteron switching fraction
    file: simulation.go:1266-1277

  - trigger: Write/Read Demo phase progression
    source: simulation.go:128-310
    updates:
      - wrdPhase state machine (0=WRITE, 1=HOLD, 2=READ, 3=DISPLAY)
      - wrdPhaseTimer accumulation
      - writeE calculation based on target vs current level
      - a.electricField ramping (+E for higher, -E for lower)
      - wrdReadLevel capture
      - wrdTotalWrites, wrdSuccessWrites
      - wrdTotalEnergyfJ accumulation
      - wrdDebugLog.Cycles append
      - saveDebugLog every 5 cycles
      - logEntries with phase transitions
    file: simulation.go:128-310
    bugs: [BUG-M1-001]

BugDetails:
  - id: BUG-M1-001
    component: saveDebugLog
    severity: High
    description: Debug log save spawns goroutine without fyne.Do wrapper for potential UI updates
    expected: All UI-related operations in goroutines must use fyne.Do
    actual: saveDebugLog called from simulationLoop via goroutine (line 279)
    file: gui.go:196-226
    line: 279
    suggested_fix: |
      Ensure saveDebugLog never touches UI state, or wrap UI updates in fyne.Do.
      Current code only does file I/O, but log.Info calls may interact with GUI logger.

  - id: BUG-M1-002
    component: initDebugLog
    severity: Medium
    description: initDebugLog accesses a.material without checking for nil
    expected: Defensive nil check before accessing material fields
    actual: Comment says "Defensive" but only checks after access for fallback values
    file: gui.go:228-263
    line: 235-238
    suggested_fix: |
      if a.material == nil {
          a.material = ferroelectric.DefaultHZO()
      }
      materialName := a.material.Name
      ...

  - id: BUG-M1-003
    component: createUI
    severity: Medium
    description: Plot initialization accesses a.material.Ec without mutex protection
    expected: All shared state access should be mutex-protected
    actual: Line 331 reads a.material fields outside any lock
    file: gui.go:331
    line: 331
    suggested_fix: |
      a.mu.RLock()
      eMax := a.material.Ec * 2.5
      pMax := a.material.Ps * 1.2
      a.mu.RUnlock()
      a.plot = widgets.NewPEPlot(eMax, pMax, ...)

  - id: BUG-M1-004
    component: simulationLoop
    severity: Low
    description: 60 FPS ticker may drift over time without vsync sync
    expected: Frame-perfect 60Hz timing
    actual: 16ms ticker accumulates error, simTime wrap at 1000s may cause glitches
    file: simulation.go:14-36
    line: 14, 34-35
    suggested_fix: |
      Use adaptive dt calculation or sync to display refresh rate.
      Remove simTime modulo wrap (breaks continuity) or use proper phase wrapping.

  - id: BUG-M1-005
    component: Keyboard E/D field control
    severity: Medium
    description: eFieldSlider.SetValue called outside mutex lock
    expected: Read current slider value with lock, then set new value
    actual: Lines 22 and 35 set slider value after releasing lock
    file: keyboard.go:11-37
    line: 22, 35
    suggested_fix: |
      Calculate newValue inside lock:
      a.mu.Lock()
      newValue := a.eFieldSlider.Value + 0.1
      if newValue > 2.0 { newValue = 2.0 }
      a.mu.Unlock()
      a.eFieldSlider.SetValue(newValue)  // OK outside lock (UI setter)

  - id: BUG-M1-006
    component: LevelIndicator pulsing effect
    severity: Low
    description: Time-based pulse effect may not refresh without manual Refresh call
    expected: Smooth pulsing animation at target level
    actual: Pulse alpha calculated from time.Now() but no auto-refresh mechanism
    file: widgets/level.go:260
    line: 260
    suggested_fix: |
      Add animation ticker to trigger Refresh at 30 FPS when highlightTarget is true.
      Or use fyne.Animation API for smooth pulsing.

Notes:
  - Simulation runs at 60 FPS in background goroutine (simulationLoop)
  - All UI updates wrapped in fyne.Do for thread safety
  - Mutex (a.mu) protects all simulation state
  - 5 waveform modes: Manual, Sine, Triangle, Write/Read Demo, Time-Resolved Switching
  - Write/Read Demo uses 4-phase state machine (WRITE, HOLD, READ, DISPLAY)
  - Manual mode uses 2-phase animation (WRITE, HOLD)
  - Time-Resolved mode shows KAI switching dynamics over 100ns (educational)
  - Correct physics: +E for higher level, -E for lower level
  - Manual mode supports click-to-level animation (gui.go:338-349)
  - Energy calculation via E·dP integration for Write/Read cycles
  - Debug logging to JSON (logs/hysteresis-TIMESTAMP.json)
  - Export feature: Ctrl+E saves P-E data to data/pe-data-YYYY-MM-DDTHH-MM-SS.{json,csv}
  - Metrics dashboard shows temperature-corrected Ec(T), Pr(T), squareness, switched fraction
  - 30 discrete levels (demo baseline; simulation baseline) = 4.91 bits/cell
  - Keyboard shortcuts fully documented in keyboard.go:140-175
  - Responsive layout adapts to mobile via AdaptiveLayout
  - Theme uses FeCIM blue (#003264) from theme.go
  - Custom widgets: PEPlot, LevelIndicator, CellVisualizer, ModeIndicator, PhaseIndicator
  - LayoutCache prevents redundant layout passes (from shared/widgets)
  - Time-Resolved mode uses SimulateDomainSwitching() from preisach_advanced.go

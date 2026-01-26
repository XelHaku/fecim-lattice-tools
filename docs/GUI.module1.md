---
Module: module1-hysteresis
Name: Hysteresis Visualizer
Entry: cmd/hysteresis/main.go
Package: multilayer-ferroelectric-cim-visualizer/module1-hysteresis/pkg/gui
---

Bugs:
  - [ ] BUG-001: UI updates from goroutine without fyne.Do() wrapper in saveDebugLog (gui.go:279)
  - [ ] BUG-002: Potential nil pointer access in initDebugLog when material not set (gui.go:235-238)
  - [ ] BUG-003: Missing mutex protection when accessing a.material in createUI (gui.go:331)
  - [ ] BUG-004: Animation loop refresh rate may cause drift without vsync sync (simulation.go:14)
  - [ ] BUG-005: Slider value setting outside mutex lock in keyboard.go (keyboard.go:22, 35)
  - [ ] BUG-006: LevelIndicator time-based pulse may not refresh properly (level.go:260)

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
                          file: gui.go:327-328, widgets/cell.go:17-241
                          state: [level (0-29), minSize=180x200]
                          bindings: [SetLevel, Refresh]
                  center:
                    - Container: Scroll
                      Components:
                        - name: infoPanel
                          type: container.VBox
                          purpose: Display state and material info
                          file: info.go:16-54
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
                          type: widget.Label
                          purpose: Memory operations log
                          file: info.go:65-69
                          state: [logText with wrapping]
                          bindings: [SetText from getLogText]

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
                    file: gui.go:331-332, widgets/peplot.go:16-398
                    state:
                      - eData: []float64 (history)
                      - pData: []float64 (history)
                      - currentE: float64
                      - currentP: float64
                      - eMax: float64 (bounds)
                      - pMax: float64 (bounds)
                      - minSize: 400x350
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
                      - minSize: 80x350
                      - OnLevelClicked: func(targetLevel int)
                    bindings:
                      - SetLevel(level)
                      - SetTargetLevel(level, highlight)
                      - Tapped(event) -> OnLevelClicked callback
                      - Refresh
                    bugs: [BUG-006]

              - name: rightColumn
                type: container.VBox
                purpose: Control panel with sliders and buttons
                file: gui.go:399-402
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
                        bindings:
                          - OnChanged -> switch waveform, toggle autoMode
                          - Manual: enable eFieldSlider
                          - Auto modes: disable eFieldSlider
                          - Write/Read Demo: reset demo state, init debug log

                      - name: eFieldLabel
                        type: widget.Label
                        purpose: Display current E-field value
                        file: controls.go:26
                        state: [text="E: 0.00 MV/cm"]

                      - name: eFieldSlider
                        type: widget.Slider
                        purpose: Manual E-field control
                        file: controls.go:16-25
                        state: [range=-2 to 2, step=0.01, value=0]
                        bindings:
                          - OnChanged -> set a.electricField (only in Manual mode)
                        bugs: [BUG-005]

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
    file: simulation.go:13-394

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
      - a.manualPhase = 1 (start saturate)
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
    file: keyboard.go:11-175
    bugs: [BUG-005]

  - trigger: Write/Read Demo phase progression
    source: simulation.go:131-340
    updates:
      - wrdPhase state machine (0-4)
      - wrdPhaseTimer accumulation
      - wrdSaturateE, wrdSettleE calculation
      - a.electricField ramping
      - wrdReadLevel capture
      - wrdTotalWrites, wrdSuccessWrites
      - wrdTotalEnergyfJ accumulation
      - wrdDebugLog.Cycles append
      - saveDebugLog every 5 cycles
      - logEntries with phase transitions
    file: simulation.go:131-340
    bugs: [BUG-001]

BugDetails:
  - id: BUG-001
    component: saveDebugLog
    severity: high
    description: Debug log save spawns goroutine without fyne.Do wrapper for potential UI updates
    expected: All UI-related operations in goroutines must use fyne.Do
    actual: saveDebugLog called from simulationLoop via goroutine (line 279)
    file: gui.go:196-226
    line: 279
    suggested_fix: |
      Ensure saveDebugLog never touches UI state, or wrap UI updates in fyne.Do.
      Current code only does file I/O, but log.Info calls may interact with GUI logger.

  - id: BUG-002
    component: initDebugLog
    severity: medium
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

  - id: BUG-003
    component: createUI
    severity: medium
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

  - id: BUG-004
    component: simulationLoop
    severity: low
    description: 60 FPS ticker may drift over time without vsync sync
    expected: Frame-perfect 60Hz timing
    actual: 16ms ticker accumulates error, simTime wrap at 1000s may cause glitches
    file: simulation.go:14-36
    line: 14, 34-35
    suggested_fix: |
      Use adaptive dt calculation or sync to display refresh rate.
      Remove simTime modulo wrap (breaks continuity) or use proper phase wrapping.

  - id: BUG-005
    component: Keyboard E/D field control
    severity: medium
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

  - id: BUG-006
    component: LevelIndicator pulsing effect
    severity: low
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
  - 4 waveform modes: Manual, Sine, Triangle, Write/Read Demo
  - Write/Read Demo uses 5-phase state machine (SATURATE, SETTLE, HOLD, READ, DISPLAY)
  - Manual mode supports click-to-level animation (gui.go:338-349)
  - Energy calculation via E·dP integration for Write/Read cycles
  - Debug logging to JSON (logs/hysteresis-TIMESTAMP.json)
  - 30 discrete levels = 4.91 bits/cell
  - Keyboard shortcuts fully documented in keyboard.go:140-175
  - Responsive layout adapts to mobile via AdaptiveLayout
  - Theme uses FeCIM blue (#003264) from theme.go
  - Custom widgets: PEPlot, LevelIndicator, CellVisualizer, ModeIndicator
  - LayoutCache prevents redundant layout passes (from shared/widgets)

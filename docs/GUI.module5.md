---
Module: module5-comparison
Name: Technology Comparison & Technical Briefing
Entry: cmd/comparison-gui/main.go
Package: multilayer-ferroelectric-cim-visualizer/module5-comparison/pkg/gui
Description: |
  Technical briefing comparison demo showing FeCIM vs CPU/GPU architectures.
  Features animated energy comparisons, market analysis, ROI calculator,
  and phased commercialization strategy visualization.

  CRITICAL: Energy claims are TRL 4 (laboratory validation only) and
  pending independent verification. CPU/GPU specs are verified from
  published datasheets.

Bugs:
  - [ ] BUG-M5-001: Animation mutex lock ordering in animationLoop
  - [ ] BUG-M5-002: Status label redundant SetText calls
  - [ ] BUG-M5-003: Hero widget CreateRenderer called multiple times
  - [ ] BUG-M5-004: Text caching to prevent resize loops

Screens:
  - Screen: MainWindow
    Layout:
      - Border:
          Top:
            - Component:
                Type: HBox
                Purpose: Header with title, subtitle, and reset button
                File: app.go:291-296
                State: Static
                Children:
                  - Component:
                      Type: Label
                      Purpose: Main title "FeCIM: The Energy Revolution"
                      File: app.go:271-273
                      State: Static (HighImportance, Bold)
                  - Component:
                      Type: Label
                      Purpose: Subtitle "Dr. external research group | COSM 2025"
                      File: app.go:275-276
                      State: Static (Italic)
                  - Component:
                      Type: Spacer
                      Purpose: Push reset button to right
                  - Component:
                      Type: Button
                      Purpose: Reset all animations
                      File: app.go:278-289
                      Bindings: Calls Reset() on all hero widgets
          Bottom:
            - Component:
                Type: HBox
                Purpose: Footer with status and disclaimer
                File: app.go:359-363
                Children:
                  - Component:
                      Type: Label (statusLabel)
                      Purpose: Display current operation status
                      File: app.go:267-268
                      State: ca.lastStatusText (cached)
                      Bindings: updateStatus()
                      Bug: BUG-M5-002
                  - Component:
                      Type: Spacer
                      Purpose: Push disclaimer to right
                  - Component:
                      Type: Label
                      Purpose: TRL 4 disclaimer
                      File: app.go:356-357
                      State: Static (Italic)
          Center:
            - Component:
                Type: AppTabs
                Purpose: Three main presentation tabs
                File: app.go:344-351
                State: ca.presentationMode (Manual/Auto/Investor/Engineer)
                Children:
                  - Tab: "The Energy Problem"
                    Component:
                      Type: AnimatedEnergyRace (hero widget)
                      Purpose: Show 80-90% energy reduction hero message
                      File: hero.go:42-261
                      State: animProgress, showWinner, pulsePhase
                      Bindings: UpdateAnimation(dt), Reset()
                      Bug: BUG-M5-003
                      DataFlow:
                        - Trigger: animationLoop (30 FPS)
                          Source: app.go:152-221
                          Updates: animProgress, pulsePhase
                          File: hero.go:76-91
                      Layout:
                        - Component:
                            Type: Text (heroText)
                            Purpose: Giant "80-90%" text
                            File: hero.go:113-116
                            State: 96pt bold, pulsing color
                        - Component:
                            Type: Text (heroSubtext)
                            Purpose: "DATA CENTER ENERGY REDUCTION"
                            File: hero.go:118-121
                            State: 28pt bold cyan
                        - Component:
                            Type: Rectangle (gpuBar)
                            Purpose: GPU energy bar (100 units)
                            File: hero.go:140-144
                            State: Animated width based on animProgress
                        - Component:
                            Type: Rectangle (fecimBar)
                            Purpose: FeCIM energy bar (~10 units)
                            File: hero.go:164-168
                            State: Animated width (10% of GPU)
                        - Component:
                            Type: Text (statStrip)
                            Purpose: "1000x less than CPU | 100x less than GPU"
                            File: hero.go:183-186
                            State: Static cyan

                  - Tab: "Market Opportunity"
                    Layout:
                      - Component:
                          Type: MarketOpportunityChart (hero widget)
                          Purpose: Show $721B addressable market
                          File: market.go:40-218
                          State: animProgress, pulsePhase
                          Bindings: UpdateAnimation(dt), Reset()
                          Bug: BUG-M5-004
                          Layout:
                            - Component:
                                Type: Text (heroText)
                                Purpose: Giant "$721B" text
                                File: market.go:106-109
                                State: 96pt bold, pulsing
                            - Component:
                                Type: Text (heroSubtext)
                                Purpose: "ADDRESSABLE MARKET BY 2030"
                                File: market.go:111-114
                                State: 28pt bold cyan
                            - Component:
                                Type: Rectangle Array (marketBoxes)
                                Purpose: Three market segment boxes
                                File: market.go:128-153
                                State: NAND ($98B), DRAM ($220B), AI ($403B)
                                DataFlow:
                                  - Trigger: animationLoop
                                    Source: app.go:188-189
                                    Updates: Market values animate to targets
                                    File: market.go:73-85

                      - Component:
                          Type: Card
                          Purpose: Phased Entry Strategy visualization
                          File: app.go:313-317
                          Children:
                            - Component:
                                Type: PhasedStrategyDiagram
                                Purpose: Show 3-phase commercialization roadmap
                                File: hero.go:265-386
                                State: animProgress, pulsePhase
                                Layout:
                                  - Component:
                                      Type: VBox
                                      Purpose: Phase 1 - NAND Replacement
                                      File: hero.go:315-327
                                      State: Static
                                  - Component:
                                      Type: VBox
                                      Purpose: Phase 2 - DRAM Replacement
                                      File: hero.go:334-346
                                      State: Static
                                  - Component:
                                      Type: VBox
                                      Purpose: Phase 3 - Full CIM
                                      File: hero.go:353-365
                                      State: Static (green highlight)

                      - Component:
                          Type: CompetitiveMatrix
                          Purpose: "Only FeCIM has ALL green checkmarks"
                          File: market.go:243-331
                          State: Static comparison table
                          Layout:
                            - Component:
                                Type: Text (heroText)
                                Purpose: "Only FeCIM has checkmarks in ALL categories"
                                File: market.go:262-265
                                State: Static 20pt cyan
                            - Component:
                                Type: Grid (6 columns)
                                Purpose: Header row
                                File: market.go:268-277
                                State: Technology|Energy|Speed|Endurance|CMOS|Scalable
                            - Component:
                                Type: Grid Array (competitors)
                                Purpose: 5 competitor rows
                                File: market.go:280-307
                                State: FeCIM (all ✓), Google TPU, Intel Loihi, IBM, ReRAM

                  - Tab: "ROI Calculator"
                    Layout:
                      - Component:
                          Type: HBox
                          Purpose: Inline configuration controls
                          File: app.go:326-333
                          Children:
                            - Component:
                                Type: Select (workloadSelect)
                                Purpose: Choose neural network workload
                                File: app.go:235-239
                                State: ca.currentWorkload
                                Bindings: onWorkloadChanged()
                                Options: MNIST, ResNet-50, BERT-Base, GPT-2, LLM-70B
                            - Component:
                                Type: Slider (inferencesSlider)
                                Purpose: Set inferences per second
                                File: app.go:242-249
                                State: ca.currentInferences (100-100000)
                                Bindings: updateCalculations()
                            - Component:
                                Type: Label (inferencesLabel)
                                Purpose: Display current inference rate
                                File: app.go:242
                                State: Bound to slider value
                            - Component:
                                Type: Button
                                Purpose: Trigger calculation
                                File: app.go:252-264
                                State: Disabled during calculation
                                Bindings: updateCalculations() in goroutine

                      - Component:
                          Type: DataCenterCalculator (hero widget)
                          Purpose: Show dynamic annual savings
                          File: widgets.go:264-435
                          State: currentWorkload, currentInferences, annualSavings
                          Bindings: SetResults()
                          Layout:
                            - Component:
                                Type: Text (heroSavingsText)
                                Purpose: Giant savings amount (e.g. "$42M")
                                File: widgets.go:366-369
                                State: 72pt green, formatted with formatHeroMoney()
                            - Component:
                                Type: Text (heroSavingsLabel)
                                Purpose: "ANNUAL SAVINGS"
                                File: widgets.go:371-374
                                State: 24pt bold cyan
                            - Component:
                                Type: HBox
                                Purpose: Configuration display row
                                File: widgets.go:382-388
                                State: Shows workload, inferences, scale (10k servers)
                            - Component:
                                Type: VBox
                                Purpose: Comparison section
                                File: widgets.go:404-409
                                Children:
                                  - Component:
                                      Type: Text (gpuCostText)
                                      Purpose: GPU baseline monthly cost
                                      File: widgets.go:391-393
                                      State: 20pt red
                                  - Component:
                                      Type: Text (fecimCostText)
                                      Purpose: FeCIM projected monthly cost
                                      File: widgets.go:395-397
                                      State: 20pt green
                                  - Component:
                                      Type: Text (savingsPercent)
                                      Purpose: Savings percentage
                                      File: widgets.go:399-402
                                      State: 28pt bold cyan

DataFlow:
  - Flow: Animation Loop
    Trigger: go animationLoop() at 30 FPS
    Source: app.go:143, 152-221
    Updates: All hero widgets (energyRace, marketChart, phasedStrategy, analogStates, dcTransformation)
    File: app.go:184-218
    Thread: Background goroutine with animMu lock
    Bug: BUG-M5-001
    Details: |
      Ticker runs every 33ms, updates simTime and calls UpdateAnimation(dt)
      on each hero widget. Uses RLock for reading running/paused state,
      full Lock for writing simTime. Calls fyne.Do() for UI refresh.

  - Flow: Workload Calculation
    Trigger: User selects workload or adjusts slider
    Source: app.go:237-238, 245-248
    Updates: Calculator widget, status label
    File: app.go:420-454
    Details: |
      1. Get MACs for workload (app.go:456-478)
      2. Calculate energy per inference (µJ)
      3. Calculate power (W) from inference rate
      4. Calculate monthly cost at $0.10/kWh
      5. Update calculator widget via SetResults()
      6. Update transformation widget
      7. Update status label (with caching check)

  - Flow: Energy Spec Verification
    Trigger: Application initialization
    Source: app.go:98-121
    Updates: cpuSpec, gpuSpec, fecimSpec
    File: app.go:34-41
    Details: |
      Energy specs initialized from constants (lines 28-32):
      - CPU: 1000 pJ/MAC (verified, Intel/AMD specs)
      - GPU: 100 pJ/MAC (verified, NVIDIA H100)
      - FeCIM: 1 pJ/MAC (CLAIMED, Dr. Tour COSM 2025, NOT verified)

      EnergySpec struct includes verification status and source details.

  - Flow: Status Label Update
    Trigger: Any calculation or state change
    Source: updateStatus() calls
    Updates: statusLabel.Text
    File: app.go:480-486
    Bug: BUG-M5-002
    Details: |
      Updates cached lastStatusText to avoid redundant SetText() calls
      which can trigger resize loops on some window managers.

  - Flow: Hero Widget Animation
    Trigger: animationLoop calls UpdateAnimation(dt)
    Source: app.go:185-199
    Updates: Widget-specific animation state
    File: hero.go:76-91, market.go:73-85
    Details: |
      Each hero widget maintains sync.RWMutex-protected animation state:
      - animProgress (0-1)
      - showWinner (bool)
      - pulsePhase (continuous)

      Refresh() reads state with RLock and updates canvas objects.

  - Flow: Embedded Mode Lifecycle
    Trigger: Tab selection in unified visualizer
    Source: embedded.go:64-88
    Updates: running, paused state
    File: embedded.go:65-78 (Start), 82-87 (Stop)
    Details: |
      Start(): Sets running=true, spawns animationLoop goroutine
      Stop(): Sets running=false, kills animation goroutine

      Uses same animMu lock as standalone app.

BugDetails:
  - ID: BUG-M5-001
    Component: animationLoop
    Severity: Medium
    Description: |
      Potential mutex lock ordering issue in animation loop. Function uses
      RLock to check running/paused state, then full Lock to update simTime.
      If another goroutine holds Lock and tries to acquire something that
      this goroutine holds, deadlock could occur.
    Expected: Safe lock ordering or single lock acquisition
    Actual: Split RLock/Unlock then Lock/Unlock pattern
    File: app.go:163-182
    Suggested_Fix: |
      Option 1: Use single Lock acquisition for entire critical section
      Option 2: Add lock ordering documentation and verify no inverse orders
      Option 3: Use atomic operations for simTime instead of mutex

  - ID: BUG-M5-002
    Component: Status Label
    Severity: Low
    Description: |
      Status label caches lastStatusText to avoid redundant SetText() calls,
      but cache is only checked in updateStatus(). If status is updated via
      direct statusLabel.SetText() elsewhere, cache gets out of sync.
    Expected: All status updates go through updateStatus() with cache check
    Actual: Cache implemented but not enforced
    File: app.go:79-80, 480-486
    Suggested_Fix: |
      Make statusLabel private and enforce all updates through updateStatus().
      Add comment documentation that direct SetText() bypasses cache.

  - ID: BUG-M5-003
    Component: Hero Widgets (AnimatedEnergyRace, MarketOpportunityChart)
    Severity: Low
    Description: |
      CreateRenderer() is called multiple times during widget lifecycle
      (initial creation, parent resize, theme changes). Each call recreates
      all canvas objects. For hero widgets with many text/rectangle objects,
      this can cause brief visual glitches.
    Expected: CreateRenderer() called once, Layout() handles size changes
    Actual: CreateRenderer() called on every Refresh() in some cases
    File: hero.go:111-206, market.go:104-180
    Suggested_Fix: |
      Implement custom WidgetRenderer with Layout() method to reposition
      existing objects instead of recreating them. Cache canvas objects
      as struct fields.

  - ID: BUG-M5-004
    Component: Text Caching in Hero Widgets
    Severity: Low
    Description: |
      Hero widgets implement lastGpuText, lastFecimText, lastValues caching
      to prevent redundant canvas.Refresh() calls. However, cache is only
      checked in Refresh(), not in UpdateAnimation(). This can cause
      unnecessary string formatting on every animation tick.
    Expected: String formatting only when value actually changes
    Actual: String formatted every tick, then compared to cache
    File: hero.go:227-244, market.go:208-214
    Suggested_Fix: |
      Move cache comparison into UpdateAnimation() before formatting.
      Only format and update if animated value crosses display threshold
      (e.g. integer boundary for "%.0f" format).

Constants:
  - Name: cpuEnergyPJPerMAC
    Value: 1000.0
    Purpose: CPU+DRAM energy per MAC operation (picojoules)
    Source: Intel/AMD published specifications
    Verified: true
    File: app.go:29

  - Name: gpuEnergyPJPerMAC
    Value: 100.0
    Purpose: GPU+HBM energy per MAC operation (picojoules)
    Source: NVIDIA H100 specifications
    Verified: true
    File: app.go:30

  - Name: fecimEnergyPJPerMAC
    Value: 1.0
    Purpose: FeCIM claimed energy per MAC operation (picojoules)
    Source: Dr. Tour COSM 2025 presentation ("under 1 picojoule")
    Verified: false (TRL 4 - laboratory validation only)
    File: app.go:31

  - Name: marketData
    Value: |
      NAND Flash: $72B (2024) → $98B (2030)
      DRAM: $130B (2024) → $220B (2030)
      AI Semiconductor: $140B (2024) → $403B (2030)
      TOTAL: $721B by 2030
    Source: WSTS Semiconductor Trade Statistics 2025, Gartner AI Forecasts
    File: market.go:32-36

Types:
  - Type: ComparisonApp
    Purpose: Main application struct with animation loop
    File: app.go:44-85
    Fields:
      - fyneApp: Fyne application instance
      - window: Main window
      - cpuSpec, gpuSpec, fecimSpec: Energy specifications with verification status
      - animMu: RWMutex protecting animation state
      - running, paused: Animation control flags
      - simTime: Elapsed simulation time (seconds)
      - presentationMode: Manual/Auto/Investor/Engineer
      - currentPhase: Auto demo phase (if in auto mode)
      - phaseTimer: Time in current phase
      - energyRace, marketChart, etc.: Hero visualization widgets
      - calculator: ROI calculator widget
      - workloadSelect, inferencesSlider: User controls
      - statusLabel, lastStatusText: Status display with caching

  - Type: EnergySpec
    Purpose: Energy specification with source verification
    File: app.go:35-41
    Fields:
      - Name: Architecture name
      - EnergyFJ: Energy per MAC in femtojoules
      - Source: Reference source
      - Verified: true if independently verified
      - SourceDetails: Additional context

  - Type: PresentationMode
    Purpose: Enum for presentation styles
    File: liveslide.go:42-64
    Values: Manual, Auto, Investor, Engineer

  - Type: AutoDemoPhase
    Purpose: Enum for auto-demo sequence phases
    File: liveslide.go:83-109
    Values: EnergyRace, Market, Competitive, Strategy, Calculator

  - Type: MarketSegment
    Purpose: Market data for single segment
    File: market.go:21-27
    Fields:
      - Name: Segment name
      - Y2024, Y2026, Y2030: Market size in billions USD
      - Color: Display color

WorkloadMACs:
  - Workload: MNIST
    MACs: 101,632
    Formula: (784×128) + (128×10)
    Description: Simple 2-layer MLP
    File: app.go:460-462

  - Workload: ResNet-50
    MACs: 4,000,000,000
    Description: Deep residual network for image classification
    File: app.go:463-465

  - Workload: BERT-Base
    MACs: 11,000,000,000
    Description: Transformer for NLP (sequence length 512)
    File: app.go:466-468

  - Workload: GPT-2
    MACs: 35,000,000,000
    Description: Large language model (117M parameters)
    File: app.go:469-471

  - Workload: LLM-70B
    MACs: 140,000,000,000,000
    Description: Llama-2-70B or similar (70B parameters)
    File: app.go:472-475

Notes:
  - Threading: Animation loop runs at 30 FPS (reduced from 60 to prevent resize loops on tiling WMs)
  - Energy Claims: FeCIM values are TRL 4 (laboratory validation) and NOT independently verified
  - Disclaimers: All UI shows explicit TRL 4 warnings and verification status
  - Calculator Scale: Uses 10,000 server data center as reference scale
  - Electricity Cost: $0.10/kWh used for all cost calculations
  - Presentation Modes: Investor mode emphasizes business case, Engineer mode shows technical details
  - Auto Demo: Cycles through phases with 10-15 second durations per phase
  - Fabless Model: Competitive matrix highlights capital-light NVIDIA-like approach
  - Market Sources: Combined WSTS and Gartner forecasts for addressable market
  - Phased Strategy: De-risking through NAND → DRAM → Full CIM progression

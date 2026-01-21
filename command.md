/ralph-loop:ralph-loop "## Unified FeCIM Visualizer - ALL 8 DEMOS COMPLETE

MISSION: Create a production-ready unified Fyne GUI with ALL 8 demos working perfectly. No placeholders. Everything functional.

REFERENCE FILES:
- PLAYBOOK.md (project handbook)
- command.md (demo specifications)
- README.md (update when done)
- demo1-hysteresis/pkg/gui/ (reference implementation)
- demo2-crossbar/pkg/gui/ (reference implementation)
- demo3-mnist/pkg/gui/ (reference implementation)
- demo4-circuits/pkg/ (CLI to convert to GUI)
- demo5-thermal/pkg/ (CLI to convert to GUI)
- demo6-multilayer/pkg/ (CLI to convert to GUI)
- demo8-comparison/pkg/ (CLI to convert to GUI)

ARCHITECTURE:
cmd/fecim-visualizer/main.go
- Single fyne.App instance
- Single window with AppTabs
  - Tab 0: Launcher (home screen with 8 demo cards)
  - Tab 1: Hysteresis (Demo 1) [BUILD]
  - Tab 2: Crossbar MVM (Demo 2) [BUILD]
  - Tab 3: MNIST (Demo 3) [BUILD]
  - Tab 4: Circuits (Demo 4) [BUILD]
  - Tab 5: Thermal (Demo 5) [BUILD]
  - Tab 6: 3D Stack (Demo 6) [BUILD]
  - Tab 7: Non-Idealities (Demo 7) [BUILD]
  - Tab 8: Comparison (Demo 8) [BUILD]

=== CRITICAL: NAVIGATION MUST WORK ===

NAVIGATION REQUIREMENTS:
1. Launcher cards MUST switch to correct tab when clicked
2. Each card Launch button calls tabs.SelectIndex(N) or tabs.Select(tabItem)
3. Tab switching must be INSTANT - no delay, no crash
4. User can click any demo card - immediately see that demo
5. User can click tab headers to switch between demos
6. Home or Launcher tab always accessible to return to main menu

NAVIGATION TESTING (MANDATORY):
1. Click Demo 1 card - Tab switches to Demo 1 - Demo 1 content visible
2. Click Demo 2 card - Tab switches to Demo 2 - Demo 2 content visible
3. Click Demo 3 card - Tab switches to Demo 3 - Demo 3 content visible
4. Click Demo 4 card - Tab switches to Demo 4 - Demo 4 content visible
5. Click Demo 5 card - Tab switches to Demo 5 - Demo 5 content visible
6. Click Demo 6 card - Tab switches to Demo 6 - Demo 6 content visible
7. Click Demo 7 card - Tab switches to Demo 7 - Demo 7 content visible
8. Click Demo 8 card - Tab switches to Demo 8 - Demo 8 content visible
9. Click tab headers directly - Switches correctly
10. From any demo, can return to Launcher tab

=== PHASE 1: FIX EXISTING DEMOS ===

DEMO 1 (Hysteresis) - VERIFY AND POLISH:
- Run demo, verify P-E curve renders
- Verify 30-level indicator works
- Verify all waveforms work (Sine, Triangle, Square, Manual)
- Verify material selector works
- Fix any UI issues
- MUST expose BuildContent() for unified app

DEMO 2 (Crossbar MVM) - FIX VISUALIZATION:
- FIX: IR Drop tab shows empty/all blue - add real voltage gradient visualization
- FIX: Sneak Paths tab shows empty - add parasitic current path visualization
- Verify Conductance heatmap works
- Verify Input/Output vectors display correctly
- Fix N-squared counter to update with array size
- Make What You Are Seeing panel change per tab
- Test all buttons: Run MVM, Analyze IR Drop, Analyze Sneak Paths, Reset
- MUST expose BuildContent() for unified app

DEMO 3 (MNIST) - FIX TRAINING AND ACCURACY:
- Verify pretrained weights load correctly
- Verify drawing canvas works (click, drag, right-click clear)
- Verify inference runs when drawing
- Verify layer activations display
- Verify confidence bars update
- FIX: If accuracy is wrong, retrain network:
  - Run training with proper MNIST data
  - Target: 87 percent accuracy (match Dr. Tour claim)
  - Save weights to demo3-mnist/weights/
- Test confusion matrix tab
- Test all demo modes (Manual, Auto Demo, Step-by-Step)
- MUST expose BuildContent() for unified app

=== PHASE 2: BUILD NEW DEMO GUIs ===

DEMO 4 (Circuits) - CONVERT CLI TO GUI:
- Create demo4-circuits/pkg/gui/ package
- Build signal flow diagram (DAC - Pump - Cell - TIA - ADC)
- Build timing diagram visualization
- Build level slider (0-29)
- Add circuit value displays
- Add Write Cycle / Read Cycle demo buttons
- Add educational panel with phase-aware text
- Add operation log with timestamps
- Add Dr. Tour quote: Works on a standard CMOS line
- MUST expose BuildContent() for unified app

DEMO 5 (Thermal) - CONVERT CLI TO GUI:
- Create demo5-thermal/pkg/gui/ package
- Build 2D heat map visualization with color gradient
- Build temperature color scale (25C blue to 85C red)
- Add hotspot identification and markers
- Add multi-layer thermal view (3 layers)
- Add thermal comparison (FeCIM vs GPU power/heat)
- Add educational panel
- Add operation log
- Add Dr. Tour quote: 1000x lower energy means 1000x less heat
- MUST expose BuildContent() for unified app

DEMO 6 (3D Multilayer Stack) - BUILD FULL GUI:
- Create demo6-multilayer/pkg/gui/ package
- Build 3D isometric view of stacked crossbar layers:
  Layer 3: 64 to 10 (Output)
  Layer 2: 128 to 64 (Hidden)
  Layer 1: 784 to 128 (Input)
- Build layer selector (click to highlight layer)
- Build via network visualization (vertical connections between layers)
- Build exploded view toggle (separate layers for clarity)
- Build data flow animation (show signal propagating through layers)
- Add layer specs display:
  - Cell pitch: 45nm
  - Layer height: 50nm
  - Via diameter: 40nm
  - Total neurons per layer
- Add educational panel explaining 3D stacking advantage
- Add operation log
- Add Dr. Tour quote about scalability
- MUST expose BuildContent() for unified app

DEMO 7 (Non-Idealities) - BUILD DEDICATED GUI:
- Create demo7-nonidealities/pkg/gui/ package
- Build three-tab view:
  Tab A: IR Drop Analysis
    - Voltage gradient heatmap
    - Show V_applied vs V_actual at each cell
    - Worst-case corner analysis
  Tab B: Sneak Path Analysis
    - Highlight parasitic current paths in red
    - Show current magnitude through unselected cells
    - Isolation ratio display
  Tab C: Conductance Drift
    - Time-series plot of conductance vs time
    - Simulate: Level 15 to 14.9 (1hr) to 14.7 (1day) to 14.5 (1week)
    - Retention time estimate
    - FeCIM advantage: 10+ year retention
- Add comparison: FeCIM vs ReRAM vs PCM drift rates
- Add educational panel
- Add operation log
- MUST expose BuildContent() for unified app

DEMO 8 (Comparison) - BUILD FROM SPEC (PRIORITY):
- Create demo8-comparison/pkg/gui/ package
- Build energy per MAC bar chart:
  - CPU + DRAM: about 1000 fJ/MAC (red bar)
  - GPU + HBM: about 100 fJ/MAC (orange bar)
  - FeCIM: about 1-10 fJ/MAC (green bar)
- Build architecture diagram:
  - Left: Von Neumann (CPU and Memory, data movement arrows)
  - Right: CIM (Memory = Compute, no movement)
- Build data center calculator:
  - Slider: inferences/sec (1K to 1M)
  - Dropdown: workload (MNIST, ResNet, BERT, GPT-2)
  - Output table showing Architecture, Power (W), Cost ($/month)
  - Savings display: FeCIM saves X percent vs GPU
- Build verified vs claimed table:
  - VERIFIED: 30 levels, 87 percent MNIST, CMOS compatible, non-volatile
  - CLAIMED: 10M times vs NAND, 1000 times vs DRAM, 80-90 percent savings
  - Source footnotes for each claim
- Add TRL status: Technology Readiness Level 4 (Lab Validation)
- Add Dr. Tour quote: This could lower data center energy by 80 to 90 percent
- Add disclaimer footer
- MUST expose BuildContent() for unified app

=== PHASE 3: LAUNCHER AND INTEGRATION ===

LAUNCHER TAB:
- 8 cards in grid (4x2 or 2 rows of 4)
- Each card shows:
  - Demo number (large)
  - Title
  - Status badge (READY - green)
  - One-line description
  - Launch button
- ALL 8 demos ready (no grayed out)
- Dr. Tour quote at top: Compute in memory where the same device does memory and computation
- Progress: 8/8 demos ready
- Click card calls tabs.Select(correctTab)

MAIN.GO ASSEMBLY:
- Single fyne.App instance
- Window title: Multilayer Ferroelectric CIM Visualizer
- Window size: 1400x900 default
- Shared theme (dark mode, FeCIM colors)
- Shared logging (no hardcoded paths)
- All 9 tabs (launcher + 8 demos)
- Launcher gets tabs reference for navigation

=== PHASE 4: VERIFICATION ===

BUILD AND RUN:
1. go build ./cmd/fecim-visualizer - must succeed
2. ./fecim-visualizer - must launch without crash
3. Test EVERY tab by clicking through

NAVIGATION VERIFICATION (ALL 10):
1. App opens - Launcher tab visible
2. Click Demo 1 card - Hysteresis tab opens
3. Click Demo 2 card - Crossbar tab opens
4. Click Demo 3 card - MNIST tab opens
5. Click Demo 4 card - Circuits tab opens
6. Click Demo 5 card - Thermal tab opens
7. Click Demo 6 card - 3D Stack tab opens
8. Click Demo 7 card - Non-Idealities tab opens
9. Click Demo 8 card - Comparison tab opens
10. Click Home tab - Returns to Launcher

PER-DEMO VERIFICATION:
- Demo 1: P-E curve traces correctly, 30 levels visible, all waveforms work
- Demo 2: Conductance, IR Drop, Sneak Paths ALL show data (not empty)
- Demo 3: Draw digit - correct prediction, about 87 percent accuracy
- Demo 4: Signal flow diagram works, timing diagram animates
- Demo 5: Heat map renders with color gradient, hotspots identified
- Demo 6: 3D view renders, layers selectable, exploded view works
- Demo 7: All three tabs work (IR Drop, Sneak Paths, Drift)
- Demo 8: Comparison chart renders, calculator works, table shows data

UI QUALITY (check EVERY demo):
- No overlapping elements
- No cut-off text
- Consistent spacing
- Responsive to window resize
- All buttons work
- All sliders work
- All plots render with data
- Colors consistent with theme
- Educational panels show relevant text
- Operation logs update

=== PHASE 5: TESTING ===

UNIT TESTS:
- shared/theme/theme_test.go
- shared/logging/logging_test.go
- demo1-hysteresis/pkg/*/test files
- demo2-crossbar/pkg/*/test files
- demo3-mnist/pkg/*/test files
- demo4-circuits/pkg/*/test files
- demo5-thermal/pkg/*/test files
- demo6-multilayer/pkg/*/test files
- demo7-nonidealities/pkg/*/test files
- demo8-comparison/pkg/*/test files

INTEGRATION TESTS:
- cmd/fecim-visualizer/launcher_test.go
- Tab navigation tests
- Demo content loading tests

RUN ALL:
- go test ./... - ALL must pass
- Fix any failing tests
- Add tests for new code
- Minimum 100 tests total

=== PHASE 6: DOCUMENTATION UPDATE ===

README.md UPDATE:
- Title: Multilayer Ferroelectric CIM Visualizer
- Badge: 8/8 demos ready
- Remove IronLattice branding
- Update quick start with go build and run commands
- Update demo table (all 8 marked done)
- Add screenshot section (placeholder paths)
- Update architecture diagram
- Keep Dr. Tour quotes
- Keep disclaimers and HONESTY_AUDIT reference

PLAYBOOK.md UPDATE:
- Update status for all 8 demos
- Mark unified app complete
- Update file structure

command.md UPDATE:
- Update status badges
- Mark all demos complete
- Update build commands

TODO.md UPDATE:
- Check off completed items
- Add any new items discovered

=== PHASE 7: FINAL POLISH ===

CLEANUP:
- Remove any hardcoded paths
- Remove debug print statements
- Ensure consistent code style
- Run go fmt ./...
- Run go vet ./...
- Check for any TODO comments and resolve

=== SUCCESS CRITERIA ===

ALL MUST BE TRUE:
1. Single binary builds: go build ./cmd/fecim-visualizer
2. App runs without crash
3. Launcher shows 8 demos ALL with green READY status
4. NAVIGATION WORKS: Click any card - correct tab opens
5. Demo 1: Full hysteresis visualization works
6. Demo 2: ALL tabs show data (Conductance, IR Drop, Sneak Paths)
7. Demo 3: Drawing works, accuracy about 87 percent
8. Demo 4: Peripheral circuits GUI works
9. Demo 5: Thermal heat map GUI works
10. Demo 6: 3D multilayer stack GUI works
11. Demo 7: Non-idealities all three tabs work
12. Demo 8: Comparison chart and calculator work
13. All tests pass: go test ./...
14. No UI bugs in any demo
15. Can navigate between ALL 8 demos and back to launcher
16. README.md updated with 8/8 demos
17. All docs updated

DELIVERABLES:
- fecim-visualizer binary (working)
- ALL 8 demos fully functional
- All navigation working
- All tests passing (100+)
- All documentation updated
- Clean code (fmt, vet)
- No placeholders, no coming soon

When ALL 8 demos work, ALL navigation works, ALL tests pass, ALL docs updated, and app is production-ready, output: UNIFIED APP COMPLETE - ALL 8 DEMOS WORKING - DOCS UPDATED" --max-iterations 3000 --completion-promise "UNIFIED APP COMPLETE - ALL 8 DEMOS WORKING - DOCS UPDATED"

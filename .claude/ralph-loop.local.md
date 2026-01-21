---
active: true
iteration: 1
max_iterations: 1000
completion_promise: "UNIFIED APP COMPLETE"
started_at: "2026-01-21T00:50:48Z"
---

Unified FeCIM Visualizer - Single GUI App

MISSION: Create one unified Fyne GUI application with all demos as tabs.

REFERENCE FILES:
- PLAYBOOK.md (project handbook)
- demo1-hysteresis/pkg/gui/ (reference implementation)
- demo2-crossbar/pkg/gui/ (reference implementation)
- demo3-mnist/pkg/gui/ (reference implementation)

ARCHITECTURE:
cmd/fecim-visualizer/main.go
- Single fyne.App instance
- Single window with AppTabs
  - Tab 0: Launcher (home screen with 8 demo cards)
  - Tab 1: Hysteresis (Demo 1) - READY
  - Tab 2: Crossbar MVM (Demo 2) - READY
  - Tab 3: MNIST (Demo 3) - READY
  - Tab 4: Circuits (Demo 4) - COMING SOON (grayed)
  - Tab 5: Thermal (Demo 5) - COMING SOON (grayed)
  - Tab 6: 3D Stack (Demo 6) - COMING SOON (grayed)
  - Tab 7: Non-Idealities (Demo 7) - COMING SOON (grayed)
  - Tab 8: Comparison (Demo 8) - COMING SOON (grayed)

LAUNCHER TAB (Home Screen):
- 8 cards in grid layout (4x2 or 3-3-2)
- Each card shows: demo number, title, status, description
- Ready demos: clickable, switches to that tab
- Coming soon demos: grayed out, not clickable
- Dr. Tour quote at top
- Progress indicator: 3/8 demos ready

REFACTORING NEEDED:
1. Extract ShowAndRun() blocking calls - remove standalone window creation
2. Each demo exposes BuildContent() returning fyne.CanvasObject
3. Shared logging factory - no hardcoded paths
4. Shared theme - consistent styling via shared/theme package

WORK ORDER (13 STEPS):

Phase 1 - Foundation:
1. Create shared/ package for theme and logging utilities
2. Refactor Demo 1 (Hysteresis) to expose BuildContent() function
3. Refactor Demo 2 (Crossbar) to expose BuildContent() function
4. Refactor Demo 3 (MNIST) to expose BuildContent() function

Phase 2 - Integration:
5. Create launcher tab with 8 demo cards (grid layout)
6. Create cmd/fecim-visualizer/main.go that assembles everything
7. Gray out tabs 4-8 with Coming Soon status

Phase 3 - Verification:
8. Build and RUN the app - verify it launches
9. Navigate to EACH demo tab and verify it renders correctly
10. Fix any UI issues (layout, sizing, colors, responsiveness)

Phase 4 - Testing:
11. Write unit tests for shared package
12. Write integration tests for tab navigation
13. Run ALL tests: go test ./...

VERIFICATION STEPS (MANDATORY):
1. go build ./cmd/fecim-visualizer - must compile without errors
2. ./fecim-visualizer - must launch without crash
3. Click Demo 1 card - must switch to hysteresis tab, UI must render
4. Click Demo 2 card - must switch to crossbar tab, UI must render
5. Click Demo 3 card - must switch to MNIST tab, UI must render
6. Demo 4-8 cards - must be grayed out, not clickable
7. Each demo functionality - buttons, sliders, plots must all work
8. go test ./... - all tests must pass

UI QUALITY CHECKS:
- No overlapping elements
- No cut-off text
- Consistent spacing between elements
- Responsive to window resize
- All buttons functional and clickable
- All sliders functional with correct ranges
- All plots render correctly with data

SUCCESS CRITERIA:
1. Single binary: produces fecim-visualizer executable
2. Launcher works: opens to launcher tab showing 8 demo cards
3. Navigation works: can click Demo 1/2/3 cards to switch tabs
4. Coming soon works: Demo 4-8 cards grayed out
5. Functionality preserved: all existing demo features work identically
6. No hardcoded paths: logging and resources use relative/configurable paths
7. No UI bugs: clean layout, no visual issues
8. Tests pass: go test ./... returns success
9. Builds clean: go build ./cmd/fecim-visualizer succeeds
10. Runs stable: ./fecim-visualizer runs without crash

When unified app builds, runs, all demos work in tabs, and all tests pass, output: UNIFIED APP COMPLETE

---
active: true
iteration: 1
max_iterations: 1000
completion_promise: null
started_at: "2026-01-21T16:16:35Z"
---

 FeCIM Visualizer - CRASH FIX AND TESTING MISSION

CRITICAL PROBLEM: App crashes on various demos due to nil pointer dereferences during initialization.
SECONDARY PROBLEM: MNIST demo doesnt predict correctly - weights appear untrained.

 PHASE 1: CRASH DETECTION AND FIX 

KNOWN CRASH PATTERN:
- Fyne Select widgets SetSelected() triggers callback IMMEDIATELY during creation
- Callback calls update functions (updateHeatmap, updateStackView, etc.)
- Update functions access containers that arent created yet = CRASH

FIX PATTERN (apply to ALL demos):
1. Reorder creation: Create visualization containers BEFORE control panels with Select widgets
2. Add nil checks at start of ALL update functions:
   go
   func (app *App) updateSomething() {
       if app.container == nil {
           return
       }
       // rest of function
   }
   

DEMOS TO CHECK AND FIX:

DEMO 1 (Hysteresis):
- Check for Select widgets with SetSelected
- Add nil checks to update functions if missing
- TEST: Click Demo 1 tab - should NOT crash

DEMO 2 (Crossbar):
- Check for Select widgets with SetSelected
- Add nil checks to update functions if missing
- TEST: Click Demo 2 tab - should NOT crash

DEMO 3 (MNIST):
- Check for Select widgets with SetSelected
- Add nil checks to update functions if missing
- TEST: Click Demo 3 tab - should NOT crash

DEMO 4 (Circuits):
- Check for Select widgets with SetSelected
- Add nil checks to update functions if missing
- TEST: Click Demo 4 tab - should NOT crash

DEMO 5 (Thermal) - FIXED:
- Reordered centerPanel before leftPanel
- Added nil checks to updateHeatmap(), updateStats()
- VERIFY: Click Demo 5 tab - should NOT crash

DEMO 6 (Multilayer) - FIXED:
- Reordered centerPanel before leftPanel
- Added nil checks to updateStackView(), updateMetrics()
- VERIFY: Click Demo 6 tab - should NOT crash

DEMO 7 (Non-Idealities) - FIXED:
- Added nil checks to updateIRDrop(), updateSneakPaths(), updateDrift()
- VERIFY: Click Demo 7 tab - should NOT crash

DEMO 8 (Comparison) - FIXED:
- Added nil check to updateStatus()
- VERIFY: Click Demo 8 tab - should NOT crash

 PHASE 2: CRASH TESTING PROCEDURE 

MANDATORY TEST SEQUENCE (run ./fecim-visualizer):

1. App launches - Launcher tab visible? YES/NO
2. Click Demo 1 card - Hysteresis loads without crash? YES/NO
3. Return to Launcher, Click Demo 2 card - Crossbar loads without crash? YES/NO
4. Return to Launcher, Click Demo 3 card - MNIST loads without crash? YES/NO
5. Return to Launcher, Click Demo 4 card - Circuits loads without crash? YES/NO
6. Return to Launcher, Click Demo 5 card - Thermal loads without crash? YES/NO
7. Return to Launcher, Click Demo 6 card - 3D Stack loads without crash? YES/NO
8. Return to Launcher, Click Demo 7 card - Non-Idealities loads without crash? YES/NO
9. Return to Launcher, Click Demo 8 card - Comparison loads without crash? YES/NO

If ANY crash occurs:
1. Note which demo crashed
2. Read the stack trace to find the nil pointer location
3. Apply the fix pattern (nil check or reorder creation)
4. Rebuild and retest ALL demos

 PHASE 3: MNIST WEIGHT TRAINING 

PROBLEM: MNIST predictions are wrong because weights are untrained.

EVIDENCE: pretrained_weights.json has nearly uniform values (~0.48-0.55)
This indicates weights were never properly trained.

TRAINING DATA AVAILABLE:
- demo3-mnist/data/train-images-idx3-ubyte.gz (60,000 training images)
- demo3-mnist/data/train-labels-idx1-ubyte.gz (60,000 training labels)
- demo3-mnist/data/t10k-images-idx3-ubyte.gz (10,000 test images)
- demo3-mnist/data/t10k-labels-idx1-ubyte.gz (10,000 test labels)

TRAINING SCRIPT: demo3-mnist/train_mnist_proper.go

TO TRAIN:
bash
cd <local-path>
go run demo3-mnist/train_mnist_proper.go


Expected output:
- Training for 15 epochs
- Final accuracy should be ~85-87%
- Weights saved to demo3-mnist/data/pretrained_weights.json

AFTER TRAINING:
1. Rebuild: go build ./cmd/fecim-visualizer
2. Test: Run app, go to Demo 3, draw a digit
3. Verify prediction matches drawn digit reasonably

 PHASE 4: FUNCTIONAL TESTING 

After crash fixes and MNIST training, test each demo FUNCTIONALITY:

DEMO 1 (Hysteresis):
- [ ] P-E curve traces when frequency > 0
- [ ] 30-level indicator shows current level
- [ ] Waveform selector works (Sine, Triangle, Square, Manual)
- [ ] Material selector works
- [ ] Manual slider works

DEMO 2 (Crossbar MVM):
- [ ] Conductance heatmap shows colors (not all same color)
- [ ] IR Drop tab shows voltage gradient
- [ ] Sneak Paths tab shows current map
- [ ] Run MVM button works
- [ ] Array size selector works

DEMO 3 (MNIST):
- [ ] Drawing canvas accepts mouse input
- [ ] Drawing shows on canvas
- [ ] Prediction updates as you draw
- [ ] Predicted digit is REASONABLE for what you drew
- [ ] Confidence bars update
- [ ] Clear button works

DEMO 4 (Circuits):
- [ ] Signal flow diagram visible
- [ ] Level slider changes values
- [ ] Timing diagram shows waveforms
- [ ] Write/Read buttons work

DEMO 5 (Thermal):
- [ ] Heat map shows color gradient (blue to red)
- [ ] Technology selector changes the map
- [ ] Stats panel updates
- [ ] Multi-layer view works

DEMO 6 (3D Stack):
- [ ] Isometric stack visualization shows layers
- [ ] Stack selector changes between Demo/MNIST stacks
- [ ] Layer list shows layer info
- [ ] Metrics panel shows values
- [ ] Energy comparison bars visible

DEMO 7 (Non-Idealities):
- [ ] IR Drop tab shows heatmap with gradient
- [ ] Sneak Paths tab shows current map with target X
- [ ] Drift tab shows time series chart
- [ ] Technology Comparison tab shows bars
- [ ] Mitigation buttons work

DEMO 8 (Comparison):
- [ ] Energy bar chart shows 3 bars (CPU, GPU, FeCIM)
- [ ] Architecture diagram visible
- [ ] Workload selector works
- [ ] Inferences slider updates calculations
- [ ] Calculator shows power and cost values

 SUCCESS CRITERIA 

MUST ALL BE TRUE:
1. go build ./cmd/fecim-visualizer - succeeds
2. ./fecim-visualizer - launches without crash
3. ALL 8 demo tabs can be opened without crash
4. Can navigate between ALL demos via cards and tabs
5. MNIST predicts reasonably after training (not random)
6. Each demos core functionality works

 QUICK REFERENCE 

Build: go build ./cmd/fecim-visualizer
Run: ./fecim-visualizer
Test: go test ./...
Train MNIST: go run demo3-mnist/train_mnist_proper.go
Vet: go vet ./...

Key files to check for crashes:
- demo*/pkg/gui/app.go - look for SetSelected() calls
- Look for update functions that access containers

Output when complete: ALL 8 DEMOS LOAD WITHOUT CRASH - MNIST PREDICTS CORRECTLY

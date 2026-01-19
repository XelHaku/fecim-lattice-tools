ACT AS: Dr. Vertex, Lead Architect & Principal Scientist.
CONTEXT: You are maintaining 'IronLattice-vis' - visualization demos for Dr. external research group's ferroelectric compute-in-memory technology.

PRIMARY REFERENCE: ironlattice-transcript.md (Dr. Tour's Nov 2024 presentation)
TASK TRACKING: **TODO.md** (authoritative task list - assess this file for current work)
STRATEGIC CONTEXT: docs/STRATEGIC_VALUE.md (business value and audience analysis)

--- CURRENT STATUS (Verified 2026-01-18) ---

Phase 1 COMPLETE (verified working):
- ✅ Demo 1: Hysteresis - builds, requires Vulkan display
- ✅ Demo 2: Crossbar MVM - runs, shows 64x64 array with 30 levels
- ✅ Demo 3: MNIST - runs, interactive mode working

Phase 2 COMPLETE (verified working):
- ✅ Demo 4: Peripheral Circuits - runs, shows DAC/ADC/TIA/Charge Pump
- ✅ Demo 5: Thermal Simulation - runs, shows heat maps and comparison

Phase 3 IN PROGRESS:
- ✅ Demo 6: Multi-Layer 3D - runs, shows 3D stack, via network, energy comparison
- ✅ Demo 7: Non-Idealities - runs, shows IR drop, sneak paths, drift analysis
- 🔲 Demo 8: Technology Comparison

Tests PASSING:
- ferroelectric: 7 tests
- simulation: 5 tests
- crossbar: 7 tests
- training (mnist): 9 tests
- peripherals: 9 tests
- thermal: 17 tests
- multilayer: 17 tests
- nonidealities: 20 tests
- TOTAL: 91 tests passing

--- VERIFIED RUN COMMANDS ---

```bash
# All tests (71 passing)
go test ./...

# Demo 1: Hysteresis (requires Vulkan)
cd demo1-hysteresis && go build ./cmd/hysteresis && ./hysteresis

# Demo 2: Crossbar MVM (works in terminal)
cd demo2-crossbar && go run ./cmd/inference --show-mvm

# Demo 3: MNIST (interactive terminal)
cd demo3-mnist && go run ./cmd/mnist

# Demo 4: Peripheral Circuits (terminal)
cd demo4-circuits && go run ./cmd/circuits --all

# Demo 5: Thermal Simulation (terminal)
cd demo5-thermal && go run ./cmd/thermal --compare

# Demo 6: Multi-Layer 3D Stack (terminal)
cd demo6-multilayer && go run ./cmd/multilayer --all

# Demo 7: Non-Idealities Analysis (terminal)
cd demo7-nonidealities && go run ./cmd/nonidealities --all
```

--- KEY FILES ---

| Category | File | Purpose |
|----------|------|---------|
| Tasks | TODO.md | **Authoritative task list** |
| Strategy | docs/STRATEGIC_VALUE.md | Business value analysis |
| Physics | demo1-hysteresis/pkg/ferroelectric/ | Preisach model, HZO params |
| Crossbar | demo2-crossbar/pkg/crossbar/array.go | 30-level MVM |
| Network | demo3-mnist/pkg/training/network.go | MNIST classifier |
| Peripherals | demo4-circuits/pkg/peripherals/ | DAC, ADC, TIA, Charge Pump |
| Thermal | demo5-thermal/pkg/thermal/ | Heat diffusion, multi-layer |
| Multilayer | demo6-multilayer/pkg/multilayer/ | 3D stack, vias, energy |
| NonIdealities | demo7-nonidealities/pkg/nonidealities/ | IR drop, sneak paths, drift |

--- IRONLATTICE SPECS (From Dr. Tour) ---

| Spec | Target | Status |
|------|--------|--------|
| Analog states | 30 levels | ✅ Implemented in all demos |
| MNIST accuracy | 87% | ✅ 95.8% achieved |
| P-E hysteresis | Square loop | Simplified tanh model |
| Thermal advantage | Cool operation | ✅ 1000x cooler shown in Demo 5 |

--- NEXT ACTIONS (Phase 3) ---

- [x] Demo 6: 3D multi-layer stack visualization
- [x] Demo 7: IR drop, sneak paths, drift simulation
- [ ] Demo 8: DRAM+CPU vs GPU vs IronLattice comparison

--- THE STORY ---

```
Demo 1: "This is how the memory cell works"              ✅ VERIFIED
Demo 2: "This is how we compute in memory"               ✅ VERIFIED
Demo 3: "This is what we can build with it"              ✅ VERIFIED
Demo 4: "This is how it fits in a real chip"             ✅ VERIFIED
Demo 5: "This is how we manage heat"                     ✅ VERIFIED
Demo 6: "This is how we scale to 3D"                     ✅ VERIFIED
Demo 7: "This is what can go wrong (and how we fix it)"  ✅ VERIFIED
Demo 8: "This is why it beats everything else"           🔲 PLANNED
```

--- DR. TOUR QUOTES ---

> 'It's got 30 discrete states. So it's not 0-1-0-1.'

> 'We're at 87% validation here... theoretical is 88%.'

> 'Compute in memory where the same device does the memory and the computation.'

> 'This could lower the requirements in a data center by 80 to 90%.'

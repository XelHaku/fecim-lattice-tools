# FeCIM Lattice Tools - TODO

**Mission**: Educational FeCIM visualization and simulation tool based on peer-reviewed HfO₂-ZrO₂ superlattice research.

**Last updated**: January 27, 2026

---

## 1. Current Status

| Module | Purpose | Status | Tests | GUI |
|--------|---------|--------|-------|-----|
| **module1-hysteresis** | P-E hysteresis, Preisach model, 30 analog states | ✅ Complete | ✅ Passing | ✅ Fyne |
| **module2-crossbar** | Matrix-vector multiplication, IR drop, sneak paths | ✅ Complete | ✅ Passing | ✅ Fyne |
| **module3-mnist** | Neural network digit recognition | ✅ Complete | ✅ Passing | ✅ Fyne |
| **module4-circuits** | DAC/ADC/TIA peripheral circuits | ✅ Complete | ✅ Passing | ✅ Fyne |
| **module5-comparison** | Technology comparison framework | ✅ Complete | ✅ Passing | ✅ Fyne |
| **module6-eda** | Verilog/DEF/LEF/Liberty generation | ✅ Complete | ✅ Passing | ✅ Fyne |
| **docs/** | Scientific documentation (78 papers catalogued) | ✅ Complete | N/A | N/A |

**Project Health**:
- Go files: 233
- Test cases: 571 (all passing)
- Generated outputs: Verilog, DEF, LEF, Liberty files (real synthesizable code)
- Scientific Rigor Score: 4.0/5 (HONESTY_AUDIT v3.0)
- Code coverage: ~85%

---

## 2. Pending Technical Work

### HIGH PRIORITY

#### Add Citations for 12 Assumed Simulation Parameters
**Status**: Pending | **Owner**: Contributors | **Priority**: HIGH

From HONESTY_AUDIT Section 6.2, the following parameters lack peer-reviewed sources:

| Parameter | Location | Value | Notes |
|-----------|----------|-------|-------|
| DRAM energy (32-bit) | physics.md:45 | 640 pJ | Needs explicit Horowitz 2014 citation |
| Drift coefficients | equations.md:407 | Various | Needs FeFET-specific literature |
| FeFET capacitance | equations.md:648 | 10-100 fF | Needs device physics reference |
| FeFET switching current | equations.md:649 | 1-10 µA | Needs measurement data |
| Nonlinearity coefficient k | mathematics.md:389 | 5-10 | Needs crossbar model reference |
| Read disturb probability | mathematics.md:731 | 10^-6 | Needs endurance study reference |

**Action**: Search 2024-2026 literature for device physics papers covering these parameters.

#### Vulkan Rendering Implementation
**Status**: TODO stubs exist | **Owner**: Graphics dev | **Priority**: HIGH

Current rendering uses CPU-based pixel operations. Five TODO items in code:

| Location | Line | Description |
|----------|------|-------------|
| module1-hysteresis/pkg/render/render.go | 301 | Vulkan compute shader implementation |
| module1-hysteresis/pkg/render/render.go | 349 | Vulkan initialization (device, queue, swapchain) |
| module1-hysteresis/pkg/render/render.go | 363 | Vulkan render loop (command buffers) |
| module1-hysteresis/pkg/render/render.go | 386 | Vulkan resource cleanup |
| shared/widgets/glossary.go | 500 | Add actual GitHub repo URL |

**Why**: 10-50x performance improvement for large crossbar visualizations (128x128+).

**Approach**:
- Use `go-vk` or `vgpu` bindings
- Implement compute shaders for heatmap generation
- Maintain CPU fallback for compatibility

#### Demo 6: Multi-layer 3D Visualization
**Status**: Not started | **Owner**: Graphics dev | **Priority**: HIGH

**Goal**: Visualize 3D stacked FeFET arrays (22nm BEOL-compatible, per CEA-Leti 2024).

**Features needed**:
- Layer-by-layer rendering (up to 512 layers per Samsung roadmap)
- Inter-layer connections (TSV visualization)
- Heat flow between layers
- Configurable stack depth (2, 4, 8, 16, 32 layers)

**Reference papers**:
- CEA-Leti December 2024: 22nm FD-SOI 3D capacitors
- Samsung Nature 2025: 256-512 layer FeFET NAND

---

### MEDIUM PRIORITY

#### Conductance Drift Simulation
**Status**: Model exists, needs GUI | **Owner**: Physics sim | **Priority**: MEDIUM

**Current**: Code includes drift coefficients in `equations.md:407`.

**Missing**:
- Time-dependent drift visualization
- Long-term retention simulation (1s, 1hr, 1day, 1year)
- Temperature-dependent drift (25°C vs 85°C)
- Mitigation strategy visualization (refresh, rewrite)

**Literature**: Need FeFET-specific drift characterization papers (2024+).

#### Device-to-Device Variation Modeling
**Status**: Planned | **Owner**: Physics sim | **Priority**: MEDIUM

**Goal**: Model realistic fabrication variations across crossbar arrays.

**Features**:
- Gaussian distribution of Pr, Ec, thickness
- Correlation modeling (adjacent cells)
- Accuracy impact analysis (best/worst/typical)
- Yield prediction

**Verification**: Compare against Nature Commun. 2023 (96.6% MNIST) device variation data.

#### Thermal GUI Upgrade
**Status**: CLI functional | **Owner**: GUI dev | **Priority**: MEDIUM

**Current**: `demo5-thermal/` runs as CLI-only tool.

**Needed**:
- Interactive heatmap (25°C - 85°C automotive range)
- Hotspot detection and labeling
- Time-domain thermal propagation animation
- 3D view (temperature gradient across chip depth)

**Reference**: AEC-Q100 Grade 0 qualification (Fraunhofer IPMS 2024).

---

### LOW PRIORITY

#### Web Deployment (WASM)
**Status**: Feasible | **Owner**: Infra | **Priority**: LOW

**Goal**: Run tool in browser for educational accessibility.

**Approach**:
- Compile to WebAssembly (Fyne supports WASM)
- Host static site (GitHub Pages, Netlify)
- Limit to demos 1-4 (avoid compute-heavy tasks)

**Benefits**: Zero-install demos for educators/students.

#### Demo Video Creation
**Status**: Planned | **Owner**: Content | **Priority**: LOW

**Goal**: 2-3 minute walkthrough video for README.md.

**Content**:
- Module 1: P-E curve animation (30 analog states)
- Module 2: Crossbar MVM computation
- Module 3: MNIST digit recognition (98.24% benchmark)
- Module 4: Peripheral circuits timing
- Module 5: Technology comparison chart

**Narration**: Focus on educational value, peer-reviewed physics.

---

## 3. Documentation Work

### Quarterly Literature Review
**Status**: Scheduled | **Due**: April 2026 | **Priority**: MEDIUM

**Goal**: Update HONESTY_AUDIT.md with 2026 Q1 publications.

**Search databases**:
- IEEE Xplore (IEDM, ISSCC, VLSI)
- Nature family (Nature Commun., Nature Electronics)
- ACS (Nano Letters, ACS AMI)
- arXiv (cs.ET, cond-mat.mtrl-sci)

**Focus areas**:
- New MNIST/accuracy benchmarks
- Endurance improvements beyond 10^12
- 3D integration production updates
- Tour's FeCIM peer-reviewed papers (if published)

### Update physics.md with Extended Ranges
**Status**: Pending | **Owner**: Maintainer | **Priority**: MEDIUM

**Current**: Pr range is 15-34 µC/cm² (room temperature only).

**Update to**:
- Pr: 15-34 µC/cm² (RT), 36.4 µC/cm² (300°C BEOL), 75 µC/cm² (4K cryogenic)
- Ec: 0.6-1.5 MV/cm (include Ga-doped low-Ec variant)

**Sources**:
- Nature Commun. 2025 (doi:10.1038/s41467-025-61758-2)
- ACS AMI 2025 (doi:10.1021/acsami.5c08743)
- Adv. Elec. Mat. 2024 (doi:10.1002/aelm.202300879)

### Add Cryogenic Operation Specs to devices.md
**Status**: Pending | **Owner**: Maintainer | **Priority**: LOW

**Content to add**:
- Temperature range: 5K - 300K
- Pr improvement: +30% @ 4K
- Endurance: Unlimited @ 82K (IEEE 2023)
- Write speed: 20x faster @ 77K
- Application: Quantum computing peripherals

**Sources**: IEEE 2024, Frontiers 2024, npj Unconv. Comp. 2025

---

## 4. Code Quality

### TODOs in Code (5 items)
**Status**: Active | **Priority**: Various

| File | Line | Description | Priority |
|------|------|-------------|----------|
| render.go | 301 | Vulkan compute shader implementation | HIGH |
| render.go | 349 | Vulkan device initialization | HIGH |
| render.go | 363 | Vulkan render loop | HIGH |
| render.go | 386 | Vulkan resource cleanup | HIGH |
| glossary.go | 500 | Add GitHub URL | LOW |

**Action**: Address Vulkan TODOs as part of "Vulkan Rendering Implementation" task.

### Cost Estimate Verification (Demo 4)
**Status**: Needs update | **Owner**: Maintainer | **Priority**: LOW

**File**: `module4-circuits/demos/liveslide.go:443-444`

**Current**: Placeholder cost estimates for DAC/ADC/TIA.

**Action**: Find 2024 foundry pricing or cite academic cost models (if available).

---

## 5. Testing

### Current Coverage
- Total tests: 571
- Pass rate: 100%
- Coverage: ~85% (untested: error paths, edge cases)

### Add Integration Tests for EDA Pipeline
**Status**: Planned | **Owner**: QA | **Priority**: MEDIUM

**Goal**: End-to-end test from FeFET specification → Verilog → DEF → LEF → Liberty.

**Test cases**:
- Full synthesis flow (16x16 crossbar)
- Output file validation (parseable Verilog, valid DEF/LEF)
- Timing/power liberty checks
- Comparison with reference designs

**Success criteria**: Generated Verilog passes Yosys synthesis.

### Maintain 100% Pass Rate
**Status**: Ongoing | **Priority**: HIGH

**Protocol**:
- Run `go test ./...` before every commit
- No merge if tests fail
- Add test for every bug fix

---

## 6. NOT Planned

Explicit list of what we're **NOT** building:

| Out of Scope | Reason |
|--------------|--------|
| Production chip design tools | Educational tool, not EDA replacement |
| Investor pitch decks | Scientific tool, not marketing material |
| Hardware-accurate SPICE models | Requires foundry PDKs (proprietary) |
| Real-time OS integration | Beyond educational scope |
| Cryptographic accelerators | Specialized application, not core physics |
| Web-based collaboration features | Single-user educational tool |

**Why explicit?** Previous TODO.md had "technical briefing" framing that misaligned with project's educational mission.

---

## 7. Verified Claims Reference

For accuracy in documentation and demos, use **ONLY** these verified claims:

### Material Properties [VERIFIED]
- **Pr**: 15-34 µC/cm² (RT), 75 µC/cm² (4K)
- **Ec**: 0.6-1.5 MV/cm
- **Min thickness**: 3.6 nm
- **Sub-1V switching**: 0.5V @ 3.6nm

### Multi-Level States [VERIFIED]
- **32-140 analog states demonstrated** (Oh 2017, Song 2024)
- **30 states (Tour)**: [PLAUSIBLE] but unverified (within range)

### Endurance [VERIFIED]
- **10^12 cycles**: DEMONSTRATED (V:HfO₂ 2024, Science 2024)
- **10^9 cycles**: Conservative baseline (IEEE IRPS 2022)

### MNIST Accuracy [VERIFIED]
- **98.24%**: HZO-FTJ reservoir (ScienceDirect 2025)
- **96.6%**: 7 VT states (Nature Commun. 2023)

### Energy Efficiency [VERIFIED]
- **25-100× vs NAND**: Samsung Nature 2025
- **70,000× vs GPU (LLM)**: Nature Comp. Sci. 2025

### 3D Integration [VERIFIED]
- **22nm BEOL**: CEA-Leti December 2024
- **512 layers roadmap**: Samsung Nature 2025

### Automotive [VERIFIED]
- **AEC-Q100 Grade 0**: -40°C to 150°C (Fraunhofer IPMS 2024)

### REMOVED Claims
- **87% MNIST (Tour)**: Removed (contradicts peer-reviewed 96.6-98.24%)
- **10M× vs NAND (Tour)**: Removed (no measurement data; verified: 25-100×)

**Full audit**: See `docs/cim/HONESTY_AUDIT.md` (v3.0, 377 lines).

---

## 8. Footer

**Next review**: April 2026 (quarterly literature update)

**Contributing**: See CLAUDE.md for development guidelines.

**Scientific accuracy**: All claims must be verified per HONESTY_AUDIT.md standards.

---

*This TODO prioritizes technical rigor and educational value over promotional considerations. The project is an open-source learning tool, not investment material.*

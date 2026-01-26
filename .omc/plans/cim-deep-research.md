# CIM Deep Research Plan: Mathematical Foundations and Physics Documentation

**Plan ID:** cim-deep-research
**Created:** 2026-01-25
**Status:** READY FOR EXECUTION
**Estimated Duration:** 16-24 hours (spread across multiple sessions)

---

## Context

### Original Request
Deep research on Compute-in-Memory (CIM) technology:
1. Download more academic papers
2. Build better documentation about CIM with math and physics fundamentals
3. Focus areas: mathematical foundations of in-memory computation, physics of ferroelectric/resistive devices, crossbar array operations, analog computing principles

### Current State Analysis (VERIFIED)

**Existing Documentation Assets:**

| File | Lines | Content |
|------|-------|---------|
| `docs/crossbar/crossbar.physics.md` | 387 | MVM, Ohm's/Kirchhoff's laws, non-idealities |
| `docs/hysteresis/hysteresis.physics.md` | 552 | P-E loops, Preisach model, hysteron math |
| `docs/project/02-curriculum/CURRICULUM.md` | 408 | 9-area doctoral curriculum overview |
| `docs/project/02-curriculum/CURRICULUM_DETAILED.md` | 657 | Full technical implementation guide |
| `docs/project/03-technical/HZO_PARAMETERS.md` | 180 | DOI-cited material parameters |
| `docs/hysteresis/hysteresis.research.md` | 479 | 50+ paper meta-study |
| `docs/papers/by-topic/PAPERS_NEEDED.md` | 204 | 67 papers tracked with status |

**Research Findings (8 parts, ~100KB total):**
- `docs/project/05-research/findings/part1.md` - part8.md
- Comprehensive research on Landau theory, domain dynamics, superlattices

**Paper Collection:**
- **134 PDFs** verified in `docs/papers/by-topic/` across 23 directories
- 67 papers tracked in PAPERS_NEEDED.md (8 CRITICAL, 17 HIGH, 27 MEDIUM, 15 LOW)

**Code Implementations (No tdgl.go exists):**
- `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go` - Mayergoyz Preisach model
- `module2-crossbar/pkg/crossbar/array.go` - MVM implementation
- `module2-crossbar/pkg/crossbar/nonidealities.go` - IR drop, sneak paths

**Landau/TDGL/Ginzburg References (24 files, 153 mentions):**
- Most mentions in: CURRICULUM_DETAILED.md (14), research findings (28+)
- TDGL equations documented in curriculum but **no standalone derivation doc exists**

### Identified Documentation Gaps

1. **Landau-Devonshire free energy** - mentioned in curriculum, not fully derived
2. **TDGL equations** - referenced extensively but no standalone mathematical derivation
3. **Information theory of analog computation** - not covered
4. **Crossbar parasitic network analysis** - simplified IR drop model only
5. **FeFET threshold voltage physics** - simplified G=f(P) mapping
6. **On-chip training mathematics** - inference-only documentation
7. **Signal-to-noise ratio analysis** - mentioned but not quantified

---

## Directory Structure Decision

**DECISION: Extend existing structure, do NOT create new `docs/physics/`**

Rationale:
- `docs/hysteresis/` already has `hysteresis.physics.md` (552 lines)
- `docs/crossbar/` already has `crossbar.physics.md` (387 lines)
- Adding to existing directories maintains discoverability

**New files will be:**
| Deliverable | Location |
|-------------|----------|
| Landau-Devonshire | `docs/hysteresis/landau-devonshire.md` |
| TDGL Derivation | `docs/hysteresis/tdgl-derivation.md` |
| Information Theory | `docs/crossbar/analog-information-theory.md` |
| Network Analysis | `docs/crossbar/crossbar-network-analysis.md` |
| FeFET Physics | `docs/hysteresis/fefet-device-physics.md` |
| Training Math | `docs/crossbar/cim-training-math.md` |
| Quick Reference | `docs/project/03-technical/quick-reference/` |

---

## Work Objectives

### Core Objective
Create a comprehensive mathematical reference library for CIM technology that serves as both educational material and implementation specification.

### Deliverables

| ID | Deliverable | Location | Priority |
|----|-------------|----------|----------|
| D0 | Gap Analysis Report | `.omc/notepads/cim-research/gap-analysis.md` | CRITICAL |
| D1 | Landau-Devonshire Physics Reference | `docs/hysteresis/landau-devonshire.md` | CRITICAL |
| D2 | TDGL Mathematical Derivation | `docs/hysteresis/tdgl-derivation.md` | CRITICAL |
| D3 | Analog Computing Information Theory | `docs/crossbar/analog-information-theory.md` | HIGH |
| D4 | Crossbar Network Analysis | `docs/crossbar/crossbar-network-analysis.md` | HIGH |
| D5 | FeFET Device Physics | `docs/hysteresis/fefet-device-physics.md` | HIGH |
| D6 | CIM Training Mathematics | `docs/crossbar/cim-training-math.md` | MEDIUM |
| D7 | Paper Download Expansion | `docs/papers/by-topic/` + PAPERS_NEEDED.md update | MEDIUM |
| D8 | Quick Reference Cards | `docs/project/03-technical/quick-reference/` | LOW |

### Definition of Done
- [ ] All mathematical derivations include LaTeX-style equations
- [ ] Each document has DOI-cited references (minimum 5 per document)
- [ ] Code implementation examples in Go where applicable
- [ ] Cross-references to existing documentation
- [ ] Validation against experimental data from HZO_PARAMETERS.md

---

## Must Have / Must NOT Have

### Must Have
- Complete derivations from first principles (no hand-waving)
- SI units consistently used with conversion tables
- Error bounds and approximation validity ranges
- Connection to implemented code in module1-hysteresis/ and module2-crossbar/
- DOI citations for all key equations
- Reference existing content before writing (avoid duplication)

### Must NOT Have
- Marketing language or unsubstantiated claims
- Equations without physical interpretation
- Parameters without experimental validation sources
- Duplicate content from existing documentation
- References to non-existent files (like pkg/physics/tdgl.go)

---

## Task Flow and Dependencies

```
Phase 0: Gap Analysis (D0)
    |
    v
Phase 1: Core Physics (D1, D2)
    |
    v
Phase 2: System-Level Analysis (D3, D4, D5)
    |
    v
Phase 3: Advanced Topics (D6)
    |
    v
Phase 4: Paper Expansion & Reference Cards (D7, D8)
```

---

## Detailed TODOs

### Phase 0: Gap Analysis (MANDATORY FIRST STEP)

#### TODO 0.1: Systematic Content Audit
**Output:** `.omc/notepads/cim-research/gap-analysis.md`
**Estimated Time:** 2 hours
**Dependencies:** None

**Files to Read and Analyze:**
1. `docs/hysteresis/hysteresis.physics.md` (552 lines) - What Landau content exists?
2. `docs/hysteresis/hysteresis.research.md` (480 lines) - What research is summarized?
3. `docs/project/02-curriculum/CURRICULUM_DETAILED.md` (657 lines) - What TDGL content exists?
4. `docs/project/05-research/findings/part1.md` through `part8.md` (~100KB) - What's already documented?
5. `docs/project/03-technical/HZO_PARAMETERS.md` (181 lines) - What parameters are already cited?

**Analysis Questions:**
- What Landau-Devonshire content already exists (equations, coefficients)?
- What TDGL content already exists (equations, discretization)?
- What information theory content already exists?
- What gaps remain vs. what's already well-covered?

**Acceptance Criteria:**
- [ ] Gap analysis document created
- [ ] Each planned deliverable annotated with "new content needed" vs "synthesis of existing"
- [ ] Identified which existing docs to cross-reference
- [ ] No duplication of existing comprehensive content

---

### Phase 1: Core Physics Documentation

#### TODO 1.1: Landau-Devonshire Physics Reference
**File:** `docs/hysteresis/landau-devonshire.md`
**Estimated Time:** 3-4 hours
**Dependencies:** D0 (Gap Analysis)

**Content Structure:**
1. Introduction to phenomenological thermodynamics
2. Free energy expansion derivation
   - F = alpha*P^2 + beta*P^4 + gamma*P^6 - E*P
   - Temperature dependence: alpha = alpha_0 * (T - T_c)
3. Equilibrium polarization from dF/dP = 0
4. Phase transitions (first-order vs second-order)
5. HZO-specific coefficients with DOI citations
6. Connection to Preisach model (reference hysteresis.physics.md)
7. Go implementation example (reference preisach_advanced.go)

**Cross-References:**
- `docs/hysteresis/hysteresis.physics.md` (existing hysteron math)
- `docs/project/03-technical/HZO_PARAMETERS.md` (parameter values)
- `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go` (implementation)

**Acceptance Criteria:**
- [ ] Derives double-well potential shape from coefficients
- [ ] Shows how Ec emerges from energy barrier
- [ ] Includes HZO coefficients from HZO_PARAMETERS.md
- [ ] References arXiv:2601.06267 and Nature Commun. 2025 papers
- [ ] Validates against Pr = 25 uC/cm^2, Ec = 1.2 MV/cm

---

#### TODO 1.2: TDGL Mathematical Derivation
**File:** `docs/hysteresis/tdgl-derivation.md`
**Estimated Time:** 4-5 hours
**Dependencies:** D1 (Landau-Devonshire)

**Content Structure:**
1. From equilibrium to dynamics (Landau-Khalatnikov)
   - dP/dt = -Gamma * (delta F / delta P)
2. Adding spatial gradients (Ginzburg term)
   - F_grad = kappa/2 * |grad P|^2
3. Full TDGL equation derivation
   - dP/dt = -L * [2*alpha*P + 4*beta*P^3 - kappa*laplacian(P) - E]
4. Numerical discretization (FDM)
5. Stability analysis (CFL condition)
6. Boundary conditions (periodic, Dirichlet, Neumann)
7. GPU implementation strategy (reference FerroX: arXiv:2210.15668)
8. Go pseudocode examples

**Cross-References:**
- `docs/project/02-curriculum/CURRICULUM_DETAILED.md` (existing TDGL mentions)
- `docs/project/05-research/findings/part4.md` (phase-field research)
- `docs/hysteresis/hysteresis.research.md` (TDGL papers listed)

**NOTE:** No pkg/physics/tdgl.go exists. This document provides derivation for FUTURE implementation.

**Acceptance Criteria:**
- [ ] Complete derivation from free energy functional to PDE
- [ ] Explicit finite difference stencils for 2D/3D
- [ ] Stability criterion: dt < dx^2 / (4*L*kappa)
- [ ] References FerroX paper (arXiv:2210.15668)
- [ ] Notes this is derivation for future implementation (no current Go code)

---

### Phase 2: System-Level Analysis

#### TODO 2.1: Analog Computing Information Theory
**File:** `docs/crossbar/analog-information-theory.md`
**Estimated Time:** 3-4 hours
**Dependencies:** None (can parallel with Phase 1)

**Content Structure:**
1. Shannon capacity for analog channels
   - C = 0.5 * log2(1 + SNR)
2. Effective bits vs physical bits
   - ENOB = (SINAD - 1.76) / 6.02
3. Noise sources in CIM
   - Thermal (kT/C)
   - Shot noise
   - Device variation (D2D, C2C)
   - Read noise
4. Precision limits of analog MVM
5. Comparison: analog CIM vs digital at equivalent energy
6. Crossbar size optimization
7. ADC/DAC resolution requirements

**Cross-References:**
- `docs/crossbar/crossbar.physics.md` (existing non-ideality content)
- `docs/project/02-curriculum/CURRICULUM_DETAILED.md` (Area 4: ADC/DAC)

**Acceptance Criteria:**
- [ ] Derives ENOB formula from noise analysis
- [ ] Shows precision vs energy tradeoff curves
- [ ] Quantifies 30-level (5-bit) effective precision
- [ ] References IBM AIHWKit noise models
- [ ] Includes Go noise simulation pseudocode

---

#### TODO 2.2: Crossbar Network Analysis
**File:** `docs/crossbar/crossbar-network-analysis.md`
**Estimated Time:** 4-5 hours
**Dependencies:** None

**Content Structure:**
1. Kirchhoff network equations for crossbar
   - Node voltage analysis
   - Current summation at bit lines
2. Parasitic resistance model
   - Wire resistance: R_wire = rho * L / A
   - Contact resistance
3. IR drop mathematical treatment
   - Iterative solution methods
   - Closed-form approximations
4. Sneak path current analysis
   - Worst-case scenarios
   - Selector device requirements
5. Optimal array sizing theory
   - Trade-off: size vs accuracy
   - Sizing formula derivation
6. Non-ideality compensation algorithms
7. SPICE-level vs behavioral model comparison

**Cross-References:**
- `docs/crossbar/crossbar.physics.md` (existing IR drop explanation)
- `module2-crossbar/pkg/crossbar/nonidealities.go` (implementation)

**Acceptance Criteria:**
- [ ] Derives voltage drop profile across N x M array
- [ ] Quantifies IR drop error: delta_V = f(array_size, G_on/G_off)
- [ ] Shows sneak path suppression with selector
- [ ] Optimal size formula: N_max = sqrt(V_read / (R_wire * I_max))
- [ ] References nonidealities.go implementation

---

#### TODO 2.3: FeFET Device Physics
**File:** `docs/hysteresis/fefet-device-physics.md`
**Estimated Time:** 3-4 hours
**Dependencies:** D1 (Landau-Devonshire)

**Content Structure:**
1. FeFET structure and operation
   - Metal-Ferroelectric-Insulator-Semiconductor stack
2. Threshold voltage shift mechanism
   - Delta_Vth = (2*Pr*t_FE) / epsilon_FE
3. Channel current equation
   - I_D = mu * C_ox * (W/L) * [(V_GS - V_th)*V_DS - V_DS^2/2]
4. Conductance-polarization mapping
   - G = f(P) derivation
5. Depolarization field effects
   - E_dep = -P / (epsilon_0 * epsilon_FE)
6. Multi-level cell (MLC) programming
   - Pulse amplitude/width control
7. Retention and endurance physics

**Cross-References:**
- `docs/hysteresis/hysteresis.physics.md` (P-E loop connection)
- `docs/project/03-technical/HZO_PARAMETERS.md` (FeFET parameters)
- `docs/project/02-curriculum/CURRICULUM.md` (Area 2: Semiconductor Devices)

**Acceptance Criteria:**
- [ ] Complete I-V characteristic derivation
- [ ] Quantifies memory window: 2*Pr*t_FE/epsilon_FE
- [ ] Shows 30-level programming pulse strategy
- [ ] References IEEE papers on FeFET modeling
- [ ] Connects to module2 conductance model

---

### Phase 3: Advanced Topics

#### TODO 3.1: CIM Training Mathematics
**File:** `docs/crossbar/cim-training-math.md`
**Estimated Time:** 4-5 hours
**Dependencies:** D4 (Crossbar Network), D5 (FeFET Physics)

**Content Structure:**
1. Backpropagation in analog domain
   - Forward pass: Y = W * X (MVM)
   - Backward pass: gradient computation
   - Weight update: delta_W = -eta * dL/dW
2. Weight update symmetry requirements
   - Potentiation vs depression asymmetry
   - LTP/LTD curves for FeFET
3. Gradient computation methods
   - Outer product: delta_W_ij = delta_i * x_j
   - Hardware implementation challenges
4. Write verify algorithms
   - Iterative programming
   - Convergence analysis
5. Noise-aware training formulation
   - Injected noise during software training
   - IBM Tiki-Taka algorithm
6. Quantization-aware training
   - Straight-through estimator
   - Mixed-precision strategies

**Cross-References:**
- `docs/papers/by-topic/PAPERS_NEEDED.md` - Section 13 (In-memory Training papers)
- `docs/project/02-curriculum/CURRICULUM_DETAILED.md` (Area 5: Neural Network Training)

**Acceptance Criteria:**
- [ ] Complete backprop equations for CIM
- [ ] Quantifies asymmetry tolerance for convergence
- [ ] Shows write-verify convergence rate
- [ ] References Science Advances hardware backprop paper
- [ ] Includes PyTorch-style pseudocode

---

### Phase 4: Paper Expansion and Reference Cards

#### TODO 4.1: Paper Download Expansion
**Location:** `docs/papers/by-topic/` + update `docs/papers/by-topic/PAPERS_NEEDED.md`
**Estimated Time:** 2-3 hours
**Dependencies:** None (can run in parallel)

**Use Existing Tracking System:**
Reference `docs/papers/by-topic/PAPERS_NEEDED.md` which already tracks 67 papers with:
- Priority levels (CRITICAL/HIGH/MEDIUM/LOW)
- Status (FOUND/Institutional/URL)
- Categories organized by topic

**Target Additions (from PAPERS_NEEDED.md CRITICAL/HIGH):**
1. **Landau Theory Papers:**
   - arXiv:2601.06267 (if not present)
   - Miller et al. original Landau-Khalatnikov paper

2. **TDGL/Phase-Field Papers:**
   - FerroX methodology paper (arXiv:2210.15668)
   - Domain wall dynamics papers

3. **Information Theory Papers:**
   - Shannon/analog computing theory
   - ENOB/SNR analysis papers

4. **Training Papers:**
   - Science Advances progressive gradient paper (HIGH in PAPERS_NEEDED.md)
   - IBM Tiki-Taka paper
   - Quantization-aware training surveys

**Acceptance Criteria:**
- [ ] Minimum 10 new papers downloaded
- [ ] All papers organized by topic
- [ ] PAPERS_NEEDED.md updated with status changes
- [ ] paper_metadata.json updated if it exists

---

#### TODO 4.2: Quick Reference Cards
**Location:** `docs/project/03-technical/quick-reference/`
**Estimated Time:** 2 hours
**Dependencies:** D1-D6 complete

**Cards to Create:**
1. `equations.md` - All key equations in one page
2. `parameters.md` - HZO and FeFET parameter tables (link to HZO_PARAMETERS.md)
3. `units.md` - SI unit conversions for CIM
4. `models.md` - When to use Preisach vs Landau vs TDGL

**Acceptance Criteria:**
- [ ] Each card fits on one printed page
- [ ] All equations numbered consistently
- [ ] Cross-references to full documentation

---

## Commit Strategy

| Commit | Content | Files |
|--------|---------|-------|
| 1 | Phase 0: Gap analysis | `.omc/notepads/cim-research/gap-analysis.md` |
| 2 | Phase 1: Core physics docs | D1, D2 |
| 3 | Phase 2: System analysis docs | D3, D4, D5 |
| 4 | Phase 3: Training mathematics | D6 |
| 5 | Phase 4: Papers and references | D7, D8 |

**Commit Message Format:**
```
docs(physics): Add [topic] mathematical reference

- [key content 1]
- [key content 2]
- References: [DOI list]

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

---

## Success Criteria

### Quantitative Metrics
- [ ] 6 new documentation files created (D1-D6)
- [ ] Minimum 30 LaTeX-style equations documented
- [ ] Minimum 40 DOI citations across all documents
- [ ] 10+ new papers downloaded and organized
- [ ] All parameter values traceable to experimental sources

### Qualitative Metrics
- [ ] A PhD student can implement TDGL solver from documentation alone
- [ ] All equations connect to existing Go code implementations where applicable
- [ ] No unsubstantiated claims about FeCIM performance
- [ ] Clear distinction between theory, simulation, and experiment

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Paper access restrictions | Medium | High | Use arXiv, PMC, preprints; note paywalled items |
| Equation transcription errors | Medium | Medium | Cross-validate with multiple sources |
| Scope creep into unrelated topics | Low | Medium | Strict adherence to CIM focus |
| Inconsistent notation | Medium | Low | Create notation guide in first document |
| Duplicating existing content | Medium | Medium | Phase 0 gap analysis is MANDATORY |

---

## Verification Steps

1. **Gap Analysis:** Verify no duplication before writing
2. **Self-Review:** Each document checked against acceptance criteria
3. **Cross-Reference:** Validate equations against existing implementation
4. **Parameter Validation:** Check all values against HZO_PARAMETERS.md
5. **Citation Check:** Verify all DOIs resolve correctly
6. **Code Linkage:** Ensure references to Go code are accurate

---

## Notes for Executor

- **READ EXISTING DOCS FIRST** - Phase 0 is mandatory
- Use existing directory structure (`docs/hysteresis/`, `docs/crossbar/`)
- Use consistent LaTeX-style formatting: `$equation$` for inline, code blocks for display
- Include "Implementation Notes" section linking to relevant Go files
- Each document should be self-contained but reference related docs
- Prioritize clarity over brevity - this is reference documentation
- Reference `docs/papers/by-topic/PAPERS_NEEDED.md` for paper tracking
- **No pkg/physics/tdgl.go exists** - TDGL doc is for future implementation

---

**PLAN_READY: .omc/plans/cim-deep-research.md**

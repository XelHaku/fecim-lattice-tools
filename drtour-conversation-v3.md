# Dr. Tour & Dr. Jaeho First Encounter: January 2026

> **Note:** This is a fictional narrative created for storytelling. Any numbers or claims inside are illustrative and should not be treated as verified facts.

**Date:** January 30, 2026 (narrative event)
**Revision:** February 3, 2026 (metrics + accuracy sync)
**Setting:** external research institution, Tour Lab conference room
**Participants:** Dr. external research group, Dr. Jaeho Shin, FeCIM Maintainers (via screen share)
**Purpose:** First demonstration of the FeCIM Lattice Tools project

---

## The Screen Share Begins

**Juan:** *shares screen* "Okay, this is the FeCIM Lattice Tools suite. It's a comprehensive educational and design tool for ferroelectric compute-in-memory. I've been building it for about a year."

**Dr. Tour:** *leans forward* "How many files are we looking at?"

**Juan:** "374 Go source files. 36 packages with tests. CI runs `go test ./...` on every push; local status depends on environment."

**Dr. Jaeho:** *adjusts glasses* "Wait, 374? When you started this, you had what—maybe 50?"

**Juan:** "Started with zero. Built it from scratch."

**Dr. Tour:** "Show me what it does."

---

## Module 1: The Hysteresis Simulator

**Juan:** *clicks Module 1* "This is the P-E curve simulator. It uses the Mayergoyz Preisach model with an adaptive hysteron grid and calibration-aware resolution."

The screen shows a live hysteresis loop with polarization on Y-axis, electric field on X-axis. The loop traces smoothly, showing the characteristic ferroelectric butterfly curve.

**Dr. Jaeho:** "The Preisach model is computationally expensive. How are you getting real-time performance?"

**Juan:** "I pre-compute the Everett function and cache hysteron states. It stays interactive while Ec/Pr scale with temperature."

**Dr. Tour:** *points at screen* "What's this '30 Levels' indicator?"

**Juan:** "That's the multi-level memory visualization. Each horizontal bar represents one of the 30 discrete polarization states. The current state is highlighted."

**Dr. Jaeho:** "Can you write to specific levels?"

**Juan:** *clicks "Write Mode"* "Yes. I can program individual levels, and there's a new ISPP system—Incremental Step Pulse Programming. It uses write-verify cycles with feedback to hit exact conductance targets."

**Dr. Tour:** "ISPP? That's what we do in the lab."

**Juan:** "I know. I implemented it based on your COSM talk. The system pulses, reads, checks if it's within tolerance, and either stops or applies another pulse. It shows the convergence statistics in real-time."

**Dr. Jaeho:** "Show me the calibration."

**Juan:** *clicks "Calibration" tab* "Temperature-aware calibration data. The level-to-conductance mapping shifts with temperature, and the system recalibrates automatically."

**Dr. Tour:** "Where did you get those conductance numbers?"

**Juan:** "Calculated from literature values. HZO resistivity, film thickness, device geometry. But I marked them as 'estimated' in the source—there's a 190-line HONESTY_AUDIT document that classifies every claim by evidence tier."

**Dr. Jaeho:** *looks at Dr. Tour* "He built a car without knowing the engine specs. But the chassis is correct."

**Dr. Tour:** *nods slowly* "Continue."

---

## Module 2: The Crossbar Simulator

**Juan:** *clicks Module 2* "Matrix-vector multiplication with non-idealities. This is the heart of analog compute-in-memory."

The screen shows a 4×4 grid of cells, each colored by conductance. Input voltages appear on the left, output currents on the right. A heatmap shows IR drop across the array.

**Dr. Jaeho:** "You're simulating IR drop?"

**Juan:** "Yes. 2.5Ω per cell pitch baseline, scaled with temperature and line width. The voltage at each cell is the applied voltage minus the cumulative drop along the wordline and bitline."

**Juan:** *clicks "Sneak Paths" tab* "This tab shows parasitic current paths. In a 0T1R array, unselected cells create sneak paths that degrade the signal. I calculate the 3-cell loops and show SNR degradation."

**Dr. Tour:** "How accurate is this?"

**Juan:** "The models are inspired by CrossSim and BadCrossbar. I haven't done formal external calibration yet, so I keep parameters literature‑based and label uncertainties."

**Dr. Jaeho:** "What about 1T1R?"

**Juan:** *clicks architecture selector* "Both. The tool models 0T1R, 1T1R, and 2T1R. With 1T1R, the transistor selector reduces sneak paths by 3 orders of magnitude. The heatmap updates in real-time."

**Dr. Tour:** "Show me drift."

**Juan:** *clicks "Drift" tab* "Retention modeling uses a logarithmic drift model with thermal activation. The demo can time‑scale long retention windows; comparative presets for FeFET/RRAM/PCM are illustrative."

**Dr. Jaeho:** "The Arrhenius parameters?"

**Juan:** "Extracted from reported in literature sources where available. Activation energy is configurable (default 0.5 eV). Parameters are documented with citations when possible and marked as assumed otherwise."

---

## Module 3: MNIST Neural Network

**Juan:** *clicks Module 3* "The flagship demo. Neural network digit recognition running on simulated FeCIM hardware."

The screen splits into two panels: "Full Precision" and "CIM Simulation." A 28×28 drawing canvas sits below.

**Juan:** *draws a "3" on the canvas* "The left panel shows what a perfect digital network predicts. The right panel shows what the FeCIM crossbar predicts—with quantization, noise, and non-idealities."

Both panels show "Prediction: 3" with confidence scores. Full precision: 97.2%. CIM: 95.8%.

**Dr. Tour:** "What accuracy are you claiming?"

**Juan:** "I don't claim any. The tool shows reported in literature benchmarks—96.6% (Nature Communications 2023) and 98.24% (ScienceDirect 2025)—and lets users see how configuration affects accuracy. Tour‑specific claims are removed."

**Dr. Jaeho:** "What happened to the 87%?"

**Juan:** "Removed it. It wasn't reported in literature. The tool uses verified literature values for benchmarks, and it labels unverified claims as such. If users want to see what their own parameters achieve, they can adjust quantization levels, noise, ADC/DAC bits, and see the accuracy change in real-time."

**Dr. Tour:** *pauses* "You removed my claim because it wasn't reported in literature?"

**Juan:** "Yes. The HONESTY_AUDIT classified your COSM 2025 presentation as Tier 5—promotional material, not reported in literature science. Only Tier 1-2 sources are treated as facts; everything else is labeled assumed or unverified."

**Dr. Tour:** *looks at Dr. Jaeho, then back* "That's... actually correct. Continue."

---

## Module 4: Peripheral Circuits

**Juan:** *clicks Module 4* "The complete chip system. Not just the array, but everything around it."

The screen shows a circuit diagram: DAC → Crossbar → TIA → ADC.

**Juan:** "8-bit DACs for write operations, transimpedance amplifiers for read, 8-bit ADCs for digitization. I model INL/DNL, offset, gain error, and timing."

**Dr. Jaeho:** "The timing?"

**Juan:** "Write pulse width, read settling time, ADC conversion cycles. Users can adjust clock frequency and see throughput vs. accuracy tradeoffs; throughput scales with clock and array size."

**Dr. Tour:** "Power?"

**Juan:** "Array power, peripheral power, total system power. I show an estimated breakdown across DAC, array, TIA, and ADC, and label the numbers as estimates. That's why ADC resolution matters—every bit adds power."

---

## Module 5: Technology Comparison

**Juan:** *clicks Module 5* "The business case. Energy per MAC comparison across technologies."

A bar chart appears: CPU+DRAM at ~1000 pJ/MAC, GPU+HBM at ~100 pJ/MAC, FeCIM at ~1 pJ/MAC (TRL 4 estimate).

**Juan:** "The tool shows literature ranges (e.g., 25–100× vs NAND) with TRL caveats. I removed the '10 million×' claim because no reported in literature data supported it."

**Dr. Tour:** "You removed another of my claims?"

**Juan:** "Yes. The tool only shows literature ranges now and labels them; the 10M× was... aspirational."

**Dr. Tour:** *chuckles* "It was. We think we can get there, but we haven't measured it."

**Dr. Jaeho:** "What's this calculator?"

**Juan:** *clicks "Data Center Savings"* "Users input their GPU count and power draw. The tool estimates annual energy/cost/CO₂ savings based on literature‑derived efficiency ranges, with TRL‑4 caveats."

---

## Module 6: The EDA Suite

**Juan:** *clicks Module 6* "This is where it gets serious. This module generates real chip design files."

**Dr. Tour:** *leans in* "What kind of files?"

**Juan:** "Verilog netlists, DEF placement files, LEF cell libraries, Liberty timing files — educational outputs, not signoff/tapeout‑ready."

**Juan:** *opens a file browser* "Look—here's a generated Verilog file."

```verilog
// Cell [0,0]: weight=0.1000, level=16, G=55.62 uS
fecim_bit #(.LEVEL(16)) R_0_0 (
    .WL  (WL[0]),
    .BL  (BL[0]),
    .VDD (VDD),
    .VSS (VSS)
);
```

**Dr. Jaeho:** "Parameterized cells. That's proper Verilog."

**Juan:** "Yes. The weights map to conductance levels, which map to cell parameters. LEF dimensions use SKY130‑scale defaults; Liberty timing values are placeholders pending characterization."

**Dr. Tour:** "You generated a full 4×4 array?"

**Juan:** "Yes. Plus a single-cell characterization testbench. The DEF file has row definitions, pin placements, and net connectivity. Larger arrays are possible but intended for educational use."

**Dr. Jaeho:** "The timing numbers in the Liberty file—where do those come from?"

**Juan:** "Educational placeholders based on public literature and reasonable RC estimates. I marked them as 'estimated' in the comments. To get real numbers, I'd need your actual device data."

**Dr. Tour:** *exchanges look with Dr. Jaeho* "Show me the design flow."

**Juan:** *demonstrates* "Configure → Layout → HDL → Explorer → Simulate → Export. Users can design in storage mode, memory mode, or compute mode. Each generates appropriate cell libraries and floorplans."

---

## Module 7: Documentation Browser

**Juan:** *clicks Module 7* "The reference system. 206 markdown documents, including 37 research‑paper summaries and 3 video transcripts, with full‑text search."

**Dr. Jaeho:** "What kind of documentation?"

**Juan:** "Glossary with 100+ terms—FeCIM, HZO, Preisach Model, MVM, Coercive Field. Research papers organized by topic: manufacturing, 3D stacking, cryogenic, security, benchmarking. Each paper has full citations and DOI links."

**Juan:** *types in search box* "Watch—if I search 'endurance'..."

The screen shows search results: 12 documents, including papers on 10⁹ cycle endurance from IEEE IRPS 2022 and 10¹² cycles from Nano Letters 2024.

**Dr. Tour:** "How many papers total?"

**Juan:** "37 catalogued and organized so far, plus a running 'new papers' list. I have sections on cryogenic operation, security/PUFs, reservoir computing—everything related to FeFETs and CIM."

**Dr. Jaeho:** "You've built a curriculum, not just a tool."

**Juan:** "I had to learn it anyway. I figured I might as well organize it so others can learn too."

---

## The HONESTY_AUDIT

**Juan:** *opens a text file* "This is the document I'm most proud of. 190 lines of audit trail."

**Dr. Tour:** *reads aloud* "'Dr. Tour's COSM 2025 presentation is Tier 5—not reported in literature, promotional context.' You really wrote that?"

**Juan:** "Yes. And '10M× vs NAND energy: REMOVED—no reported in literature data exists for this claim.' And '87% MNIST accuracy: REMOVED—below reported in literature 96.6-98.24%.'"

**Dr. Jaeho:** "You've audited your own sources and Jim's claims."

**Juan:** "Every claim is tagged as verified, assumed, unverified, or removed. The open items are listed explicitly, and the docs link back to sources wherever possible."

**Dr. Tour:** *quiet for a moment* "Do you know how rare this is? Most people building tools like this would just use my numbers because I'm famous. You actually checked."

**Juan:** "The science matters more than the scientist."

---

## The Verdict

**Dr. Tour:** "Let me tell you what I'm seeing here."

*stands up, paces*

**Dr. Tour:** "374 Go files. Verilog/DEF/LEF/Liberty outputs (educational). A Preisach simulator running interactively. ISPP with calibration. Neural networks with honest accuracy numbers. 37 research‑paper summaries catalogued. And a document that critiques my own claims."

**Dr. Jaeho:** "The infrastructure is sound. The physics models track literature trends. The EDA outputs are educational‑grade."

**Dr. Tour:** "The question isn't 'is this real?' anymore. It's 'what would it take to make it accurate?'"

**Dr. Jaeho:** "We'd need to share device parameters. Ec, Pr, conductance values, timing. Calibrate his models to our actual measurements."

**Dr. Tour:** "And then?"

**Dr. Jaeho:** "Then this becomes a real design tool. Not just educational—actual pre-production FeCIM design."

**Dr. Tour:** *turns to screen* "Juan, I have three options for you."

**Option A: Independent Path**
Continue alone. Keep using reported in literature sources as the factual baseline. Publish as open-source educational software. Never get our data, but maintain complete independence.

**Option B: Collaboration Path**
Share device parameters. Let us calibrate the models. Co-publish if results are interesting. Risk: IP complications, potential rejection.

**Option C: Hybrid Path** (RECOMMENDED)
Continue developing independently with honest documentation. Repo is ready if we decide to collaborate. Tool works with any FeFET parameters—not locked to our specs. We can engage whenever we're ready.

**Juan:** "I've already chosen C. The tool exists. The documentation is honest. The collaboration is optional."

**Dr. Tour:** *smiles* "You didn't need me to tell you that."

**Dr. Jaeho:** "One more thing. That ISPP system—can you add more detailed convergence statistics?"

**Juan:** "Already planned. Mean pulses per level, standard deviation, final error distribution. Sprint 2 in my TODO list."

**Dr. Tour:** "And error bars on all the physics parameters?"

**Juan:** "P1 critical item. Adding confidence intervals to every value in the UI."

**Dr. Jaeho:** "Device-to-device variation?"

**Juan:** "Gaussian Ec/Pr distribution with 15% sigma. Academic peer review item C11."

**Dr. Tour:** *sits back down* "You have a 58-item critique list and you're working through it systematically."

**Juan:** "50 done, 8 to go. About 50–80 hours of work remaining."

**Dr. Tour:** "To reach what state?"

**Juan:** "'Validated FeCIM Simulator' status. Every physics model verified against literature or your data. Every parameter has error bars. Every claim is sourced."

**Dr. Jaeho:** "That's not a hobby project. That's infrastructure."

**Dr. Tour:** "Juan, if you emailed me this a year ago, I would have been polite. If you email me this now, I'm forwarding it to Jaeho within 30 minutes. Not because we need software help—though we might. But because you clearly understand the problem space at a level that's useful for technical discussions."

**Juan:** "I don't need you to validate my work. But I'd like to collaborate if you're willing."

**Dr. Tour:** "The work validates itself. The HONESTY_AUDIT validates your integrity. The 374 files validate your persistence."

*pauses*

**Dr. Tour:** "We'll be in touch."

---

## Appendix: Current Metrics (February 3, 2026)

| Metric | Value | Notes |
|--------|-------|-------|
| Go files | 374 | Snapshot count |
| Test packages | 36 | Dirs containing `*_test.go` |
| Modules | 7 | Complete |
| Research papers catalogued | 37 | Markdown summaries in `docs/research-papers/` |
| Documentation files | 206 | `docs/**/*.md` |
| HONESTY_AUDIT lines | 190 | Snapshot count |
| EDA output formats | 4 (v, def, lef, lib) | Educational artifacts |
| Critique items completed | 50/58 | Pending: HIGH-003/004, MED-004, LOW-002, L07-L10 |
| ISPP implementation | ✅ | Implemented |
| Temperature calibration | ✅ | Implemented |
| GPU acceleration | Optional Vulkan renderer/compute | Experimental (SDK/GPU dependent) |

---

## Current Highlights (February 3, 2026)

1. **ISPP (Incremental Step Pulse Programming)** with write‑verify stats
2. **Temperature‑aware calibration** with automatic remapping
3. **Sneak path comparison view** (PASSIVE vs 1T1R)
4. **Screenshot metadata embedding** for reproducible captures
5. **Accessibility utilities** (keyboard help, high‑contrast helpers)

---

## The Path Forward (Updated)

**Sprint 1 (Current):** Add explicit citations/assumption labels for Circuits read/write voltages (HIGH‑003/004)
**Sprint 2:** Add GPU batching/throughput caveat in Circuits comparison (MED‑004) + unified “About the Science” entry point (LOW‑002)
**Sprint 3 (Longer‑term):** Demo video, WASM build, Vulkan large‑array path, 3D multi‑layer visualization

**Estimated effort:** ~40–80 hours, dominated by WASM/3D work

---

## Final Words

**Dr. Tour:** "Most people who build ambitious unsolicited projects disappear when they don't get the response they wanted. You didn't. You kept building. And you built something that critiques my own claims while respecting the science behind them."

**Dr. Jaeho:** "The tool is useful now. With our data, it could be essential."

**Juan:** "I'm not building it for validation. I'm building it because the science is interesting and the technology matters. If it helps your work, that's a bonus."

**Dr. Tour:** "It might. We'll see. But regardless—excellent work."

*screen share ends*

---

*Document created: January 30, 2026*
*Last revision: February 3, 2026 (metrics + accuracy sync)*
*Format: First-encounter narrative with Dr. Tour and Dr. Jaeho*
*Purpose: Capture fresh reaction to current project state*

---

## Faith Note

The HONESTY_AUDIT answers the question I asked last year: "Am I serving Him or serving my own ambition?"

When you audit your own sources—including the claims of the most famous scientist in your field—and you classify them honestly, that's not ambition.

That's integrity.

**Tour might respond. God will.**

Either way, the tool has value.

---

*"Whatever you do, work at it with all your heart, as working for the Lord, not for human masters."* — Colossians 3:23

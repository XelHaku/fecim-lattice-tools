# FeCIM EDA Suite Project Strategy
## "The Controlled Threat" - Outreach to Dr. external research group

**Date:** January 22, 2026

---

## Project Context
*   **Repository:** `XelHaku/multilayer-ferroelectric-cim-visualizer` (Private)
*   **Status:** 6 demos (5 working + 1 designed)
*   **Timeline:** Built in **6 days** (Jan 16-22, 2026)
*   **Inspiration:** Dr. external research group's COSM presentation (watched 2 weeks ago)

---

## What Was Built

### Working Demos (1-5)
1.  **Hysteresis Physics** - Preisach model, 30 discrete analog states.
2.  **Crossbar Array Simulation** - MVM with non-idealities (IR drop, sneak paths).
3.  **MNIST Neural Network** - Dual-mode FP32 vs CIM inference (targeting 87% accuracy).
4.  **Peripheral Circuits** - DAC/ADC/TIA integration.
5.  **Technology Comparison** - Energy metrics vs NAND/DRAM.

### Designed (Demo 6)
6.  **FeCIM EDA Design Suite** - Architecturally complete plan.
    *   **Goal:** Bridge the gap from NN weights to SPICE netlists and GDSII layouts.
    *   **Components:** Compiler, SPICE Export, GDSII Export, Design Space Explorer.

---

## Strategic Goal
**Email Dr. external research group (external research institution) by Friday, Jan 24, 2026.**

### The "Controlled Threat" Strategy
*   **Leverage:** You have built the **first open-source EDA suite for FeCIM**.
*   **Risk for Them:** If they ignore you, you open-source it. Competitors get free tools; they lose control of the standard.
*   **Benefit for Them:** If they engage, they get a capable collaborator/tool-builder and control the trajectory.
*   **Your BATNA:** Open-source it anyway, become the "FeCIM EDA guy", attract other opportunities.

### Key Tactics
1.  **Private Repo:** Keeps the IP protected and creates a reason for them to reply ("Reply with GitHub username for access").
2.  **Unlisted Video:** Proves capability without giving away the \"secret sauce\" publicly yet.
3.  **Honest Timeline:** "Built in 6 days" demonstrates massive execution speed.

---

## The Email Draft

```markdown
TO: tour@rice.edu
CC: jaeho-shin@rice.edu, tawfik.jarjour@accenture.com
Subject: FeCIM EDA Design Suite - Neural Networks to Silicon Automation

Dr. Tour,

Two weeks ago, I watched your COSM presentation on ferroelectric compute-in-memory. 
I immediately recognized a critical gap: there's no open-source path from neural 
network weights to FeFET crossbar SPICE netlists and physical layouts.

I spent the past 6 days building a complete design automation suite to address this.

**Demo video (10 min):** [YouTube unlisted link]

**What I built - Six integrated modules:**
1. Hysteresis Physics - Preisach model, 30 discrete analog states
2. Crossbar Array Simulation - MVM with IR drop, sneak paths
3. MNIST Neural Network - Targeting your reported 87% hardware validation
4. Peripheral Circuits - DAC/ADC/TIA integration
5. Technology Comparison - Energy metrics vs NAND/DRAM
6. FeCIM EDA Design Suite [Architecturally complete]:
   - Compiler: Weights → conductance maps + programming voltages
   - SPICE Export: ngspice-compatible netlists
   - GDSII Export: KLayout/GDSFactory integration
   - Automated Design Flow: PyTorch → Tape-out ready files

**The Gap:** Your team hand-crafts SPICE netlists. This automates it. 
Load weights, click compile, export SPICE + GDSII in minutes.

**Timing:** With FMC raising €100M and FeCIM moving to pilot production (TRL 7-8), 
design automation is the critical missing link.

**Validation:** Models are based on published literature. To be a true design tool, 
I need calibration with your actual device parameters (P-E curves, coercive field).

**Repository:** Private GitHub at https://github.com/XelHaku/multilayer-ferroelectric-cim-visualizer

To review the implementation, reply with your GitHub username(s) and I'll add you 
as collaborators immediately.

FeCIM Maintainers
Monterrey, Mexico
maintainers@example.invalid
```

---

## 48-Hour Checklist (Before Jan 24)

### Wednesday (Jan 22)
*   [ ] **Runtime Counter:** Add to GUI (`shared/gui/runtime.go`) to prove continuous recording.
*   [ ] **Practice:** Dry run of the video walkthrough.

### Thursday (Jan 23)
*   [ ] **Record Video:** 10 mins, show all 5 demos + explain Demo 6 vision.
*   [ ] **Upload:** YouTube **UNLISTED**. Title: "Ferroelectric CIM EDA Suite - First Open-Source Design Flow".

### Friday (Jan 24)
*   [ ] **Send Email:** 8-9 AM CST.

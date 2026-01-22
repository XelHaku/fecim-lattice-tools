# Zero to ASIC: A Practical Field Guide

**Derived from the "Zero to ASIC" Demo by Matt Venn & Mohammed Kassem**

> [!IMPORTANT]
> **2026 Update:** While the *principles* in this guide are timeless, the **logistics** have changed.
> *   **Efabless** has shut down operations (2025). The shuttle platform has moved.
> *   **Tiny Tapeout** now runs on the **IHP** process (Germany), not SkyWater via Efabless.
> *   **OpenLane v2** (Python-based) is now the standard text-to-gds flow.
> keep this in mind when reading references to "Google Shuttle" or "Efabless".

> *"I'm really a learning-by-doing kind of person... so I just kind of jumped into doing it."*

This guide distills the practical, hands-on insights from the "Zero to ASIC" workflow. While `eda.opensource.md` covers the tools, this guide covers the **journey**—specifically navigating the confusing reality of PDKs, intermediate files, and tapeout logistics.

---

## 1. The Reality of the PDK (SkyWater 130nm)

When you download the open-source PDK, you might be overwhelmed.

### The "75,000 Files" Problem
Downloading the SkyWater PDK results in ~3GB of data and 75,000+ files. **You do not need most of them.**

The sheer density of information often stalls beginners. Key file types you actually care about:
*   **LEF (Library Exchange Format):** The "abstract" view. Defines the size of a cell and where the pins are. Used for routing.
*   **DEF (Design Exchange Format):** The roadmap. Defines where cells are placed and which nets connect them.
*   **GDSII:** The final layout. The binary file sent to the factory (the "PDF" of chips).
*   **Liberty (.lib):** Timing and power information for the cells.

### The Layers (Not to Scale)
The 130nm process is built up in specific layers. Understanding this 3D structure helps visualize what the tools are doing:

1.  **FEOL (Front End of Line):** The bottom. P-substrate, diffusion, polysilicon gates. This is where the transistors live.
2.  **Local Interconnect:** A special "green" layer in Magic that connects nearby transistors.
3.  **BEOL (Back End of Line):** The metal layers (Metals 1-5).
    *   **Metals 1-4:** Used for connecting *your* design (standard cells).
    *   **Metal 5:** Often used for power distribution or the harness.
    *   **RDL (Re-Distribution Layer):** The final packaging layer for flip-chip bumps.

---

## 2. The OpenLane Flow: What Actually Happens?

OpenLane automates the progression from "Code" to "GDSII".
*   **Classic (v1):** Tcl-based flow (described below).
*   **Modern (v2/LibreLane):** Python-based flow. The steps are the same, but the configuration is more flexible.

If you look under the hood (in the `runs/` directory), here is what you see at each step:

### Step A: Synthesis (Yosys)
*   **Input:** Verilog (`assign out = ~in;`)
*   **Output:** A Netlist of Standard Cells (`sky130_fd_sc_hd__inv_8`)
*   **Reality Check:** The tool chooses specific cells driven by power/speed constraints (e.g., picking a huge 8x inverter for driving a long wire).

### Step B: Floorplanning
*   **What it looks like:** A mostly empty rectangle with pins around the edge.
*   **Key Decision:** How big is the die? (e.g., 20x20 microns for an inverter).

### Step C: Placement (Global -> Detailed)
*   **Global Layout:** Cells are floating roughly where they should be to minimize wire length. They might overlap.
*   **Detailed Replacement:** Cells are "snapped" to the manufacturing grid. They fit together like Lego bricks, sharing power rails (VDD top, VSS bottom).

### Step D: Clock Tree Synthesis (CTS)
*   **The Problem:** The clock signal must reach every flip-flop at exactly the same time.
*   **The Fix:** The tool inserts buffers (repeaters) to balance the delay. *Note: If your design has no clock, you must disable this step or the flow fails.*

### Step E: Routing
*   **The Result:** A mess of colored lines connecting the pins.
*   **Antenna Rules:** Long wires can act as antennas during manufacturing, collecting charge that destroys gates. The router automatically inserts "antenna diodes" (safety valves) near gates to drain this charge.

---

## 3. The Open Shuttle Ecosystem (Post-2025)

The concept of the "Free Shuttle" lives on, but the provider has changed.

### The "Mega Project Area" (User Project Area)
*   **Size:** ~10mm² (roughly 2.9mm x 3.5mm) on IHP 130nm or SkyWater.
*   **Capacity:** Huge for 130nm. You can fit ~1000 small blocks, or 10 RISC-V cores.
*   **Philosophy:** "Multi-Project Wafer" (MPW). Many different designs share one wafer to split the cost.

### The Harness (Caravel -> IHP Harness)
You don't just get a naked chip; you get a System-on-Chip (SoC) wrapper.
*   **Legacy:** The Caravel harness (RISC-V management core) was standard for Efabless.
*   **Modern:** IHP shuttles use a similar harness concept to provide IO, logic analysis, and wishbone bus access.

### Packaging (WLCSP / QFN)
*   **Wafer Level Chip Scale Package:** The chip *is* the package (common in SkyWater runs).
*   **QFN:** IHP runs often use QFN packaging, which is easier for hobbyists to solder.

---

## 4. Practical Tips & "Gotchas"

### Analog Blocks
*   **Status:** The "Digital" flow is robust. The "Analog" flow is empty.
*   **Missing:** No PLLs, no standard OpAmps in the digital library.
*   **Action:** You must build these yourself or use community designs.

### FPGA Prototyping
*   **Rule:** Always verify on FPGA first.
*   **Why:** Simulations are slow. FPGAs are fast. If your logic is wrong on an FPGA, it will be wrong on silicon. However, FPGAs won't catch "Clock Domain Crossing" or specific analog timing issues that ASICs have.

### The "Magic" of Magic
*   **Extraction:** Magic can extract parasitic capacitance.
*   **Simulation:** You can prove not just *logic* but *speed*.
*   **Example:** An inverter isn't instant. It takes ~1ns to switch. This "delay" is what limits your maximum clock speed.

---

## 5. Next Steps

1.  **Tiny Tapeout:** Start here. Now running on **IHP 130nm**. It abstracts the harness complexity.
2.  **Zero to ASIC Course:** Matt Venn’s guided path through this entire stack.
3.  **Install OpenLane:** Check out **OpenLane v2** (LibreLane) for the latest Python-based features.
4.  **Join the Community:** The Tiny Tapeout Discord is now the central hub for the open silicon community.

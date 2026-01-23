# FeCIM Lattice Generator
## Reference Material for Backend Implementation

**Context:** This is a prompt used to generate technical guidance for a "Go + OpenLane" hybrid flow.

---

### Project Name: `fecim-lattice-generator`

**Core Concept:** A hybrid flow where **Go** handles the complex mathematical placement (fractal/lattice) and **OpenLane** handles the physical routing and hardening.

### Phase 0: The Virtual Lab (Environment & Rules)
*   **Infrastructure:** Docker for OpenLane, SkyWater SKY130 PDK.
*   **The "Mother Cell":** Manual design of 1 FeCIM bit in Magic/KLayout.
*   **Output:** `.lef` (abstract) and `.gds` (physical) of the single cell.

### Phase 1: The Brain - `fecim-lattice-generator` (Go)
*   **Metaprogramming Strategy:** Using Go's text/template to generate files.
*   **Go Logic:** Calculates (x, y) coordinates for each cell based on fractal equations.
*   **Output 1 (Verilog):** `lattice.v` - The netlist of thousands of connected instances.
*   **Output 2 (DEF):** `placement.def` - The file forcing exact physical positions using `+ PLACED ( <x> <y> ) N ;`.

### Phase 2: Simulation (Safety Net)
*   **Tools:** Icarus Verilog + GTKWave.
*   **Process:** Go generates `Matrix_Logic.v` and `testbench.v`. Simulation verifies logic before physical hardening.

### Phase 3: The Macro Factory (OpenLane Hardening)
*   **Config:** Configure OpenLane to accept the external DEF and skip floorplanning/placement.
*   **Variables to check:** `CURRENT_DEF`, `PL_TARGET_DENSITY`, `FP_SIZING`.
*   **Flow:** Synthesis (Yosys) -> Routing (OpenROAD) -> Sign-off (DRC/LVS).

### Phase 4: Top-Level Integration
*   **Analog Wrapper:** Instantiation of ADCs/DACs (IP blocks) to read ferroelectric states.
*   **Padframe:** Connecting the macro to physical I/O pins.
*   **Tape-out:** Generation of the final GDSII.

---

### Architecture of Files
*   `/cells/fecim_bit.mag` (Manual Source)
*   `/src/main.go` (The Generator)
*   `/generated/lattice.v` (The Netlist)
*   `/generated/placement.def` (The Geometry)
*   `/openlane/config.tcl` (The Recipe)
*   `/runs/final_chip.gds` (The Product)

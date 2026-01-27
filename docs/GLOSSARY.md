# FeCIM Technical Glossary

Technical terms used across the FeCIM Lattice Tools documentation and codebase.

---

## Physics

### FeCIM
Ferroelectric Compute-in-Memory - analog computation using ferroelectric memory cells with programmable polarization states. Enables matrix-vector multiplication in hardware using Ohm's law (V=IR).

### Ec
Coercive Field - Electric field required to switch ferroelectric polarization state. Typical range: 0.6-1.5 MV/cm for HZO superlattices. Lower Ec reduces write energy but may compromise retention.

### Pr
Remnant Polarization - Polarization remaining after electric field removal. Verified range: 15-34 µC/cm² at room temperature, 75 µC/cm² at 4K. Higher Pr enables larger on/off ratios and analog storage capacity.

### HZO
Hafnium Zirconium Oxide (Hf₁₋ₓZrₓO₂) - ferroelectric superlattice material. CMOS-compatible, demonstrated 30 discrete analog states per cell (~4.9 bits/cell). Optimal composition: x ≈ 0.5.

### Hysteresis Loop
P-E (Polarization-Electric Field) curve showing path-dependent behavior. Area enclosed represents energy dissipation per cycle. Shape determines analog state density and retention.

### Preisach Model
Mathematical model for hysteresis using distribution of elementary switching operators (hysteron). Captures memory effects and minor loop behavior. Key parameters: α (coercive field), β (interaction field).

### Endurance
Number of write/erase cycles before device failure. Demonstrated: 10⁹ cycles (standard HZO), 10¹² cycles (V:HfO₂ doping). Automotive grade (AEC-Q100) requires 10⁶ cycles minimum.

---

## Architecture

### 1T1R
One Transistor, One Resistor architecture. Transistor acts as selector to eliminate sneak paths in crossbar arrays. Enables high density (4F² footprint) with precise cell addressing.

### MVM
Matrix-Vector Multiplication - Parallel computation in crossbar: I = G·V where I is output current vector, G is conductance matrix (programmed weights), V is input voltage vector. Single-step MAC operation via Kirchhoff's current law.

### MAC
Multiply-Accumulate - Fundamental neural network operation: y += w·x. Traditional digital requires separate multiply and accumulate steps. FeCIM performs MAC in single analog step using Ohm's law.

### BEOL
Back-End-Of-Line - Later stages of chip manufacturing (metal interconnect layers above transistors). FeFET integration demonstrated at 22nm BEOL (CEA-Leti 2024) enables 3D memory stacking without disturbing CMOS logic.

### Sneak Path
Unintended current flow through neighboring cells in passive crossbar arrays. Causes read/write interference. Mitigated by 1T1R architecture (selector transistor per cell) or high selector nonlinearity.

### IR Drop
Voltage loss across interconnect resistance in large crossbar arrays. Causes location-dependent read/write errors. Scales as O(N²) for N×N arrays. Mitigated by segmentation or peripheral driver circuits.

---

## Circuits

### DAC
Digital-to-Analog Converter - Converts digital input vectors to analog voltages for crossbar wordlines. Precision determines input quantization noise. Typical: 4-8 bits for neural network inference.

### ADC
Analog-to-Digital Converter - Converts analog output currents from crossbar bitlines to digital values. SAR (successive approximation) ADCs common for low power. Resolution: 6-12 bits typical.

### TIA
Transimpedance Amplifier - Converts crossbar output current to voltage with gain R_f: V_out = -I_in·R_f. Critical for low-noise readout. Bandwidth and offset current determine inference speed/accuracy.

### Sense Amplifier
High-gain differential amplifier for detecting small signal differences in memory cells. Enables faster read operations but may limit analog state resolution (typically 3-5 bits).

---

## Metrics

### TRL
Technology Readiness Level - Scale from 1 (basic principles) to 9 (production ready). FeCIM status: TRL 6-7 (prototype demonstration), TRL 8 (automotive-grade, Fraunhofer IPMS 2024), TRL 9 (commercial production pending).

### TOPS/W
Tera-Operations Per Second per Watt - energy efficiency metric. FeCIM demonstrated: 200-400 TOPS/W (inference). Digital ASIC: ~10 TOPS/W. Energy savings from analog MAC and in-memory computation.

### Bits per Cell
Information density in single memory element. FeCIM: ~4.9 bits/cell (30 analog states), up to 6.1-7.1 bits/cell (140 states demonstrated by Song 2024). NAND flash: 2-4 bits/cell.

### MNIST Accuracy
Handwritten digit classification accuracy (0-9). State-of-art FeCIM: 98.24% (FTJ reservoir computing, ScienceDirect 2025), 96.6% (HZO crossbar, Nature Commun. 2023). Software: 99.7%.

### Retention Time
Duration data remains stored without refresh. HZO FeFETs: >10 years at 85°C (industry standard). Improves at cryogenic temperatures (5K operation demonstrated). Critical for non-volatile applications.

### Write Energy
Energy per bit for programming memory state. FeCIM: ~1-10 fJ/bit (ferroelectric switching). NAND flash: ~100 fJ/bit. Lower energy from smaller capacitance and voltage scaling (BEOL integration).

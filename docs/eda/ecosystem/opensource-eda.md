# The Open-Source EDA Ecosystem

**An Overview for CMOS and Emerging Memory Technologies**

> **Disclaimer:** This document collects publicly available information about open-source EDA tools. It is not affiliated with or endorsed by any of the tools, projects, or organizations mentioned.

---

## 1. Introduction

The semiconductor industry stands at a pivotal juncture characterized by the decoupling of design capabilities from proprietary, capital-intensive infrastructure. Historically, the design and fabrication of Integrated Circuits (ICs) were the exclusive domain of large integrated device manufacturers (IDMs) and well-funded fabless design houses, protected by high barriers to entry including expensive Electronic Design Automation (EDA) licenses and restrictive Non-Disclosure Agreements (NDAs). This paradigm has been disrupted by the Free and Open Source Silicon (FOSSi) movement, which advocates for the democratization of chip design through open-source tools, open Process Design Kits (PDKs), and accessible shuttle programs.

For researchers and engineers working on Compute-in-Memory (CIM) and Ferroelectric Field-Effect Transistor (FeFET) based systems, this open ecosystem presents both opportunities and significant technical gaps. Standard Digital Logic flows (RTL-to-GDSII) have matured to production readiness for standard CMOS, yet the tools required for emerging non-volatile memory technologies remain fragmented. This document provides an overview of the current open-source EDA landscape, noting its readiness for CMOS digital design and identifying tooling gaps for FeCIM architectures. Information is collected from academic research, software repositories, and industry reports.

---

## 2. Open Source EDA Tools Landscape

The open-source EDA ecosystem is bifurcated into two primary flows: the digital design flow, which is highly automated and relies on abstraction, and the analog/mixed-signal flow, which requires manual intervention and transistor-level simulation.

### 2.1 Digital Design Flow (RTL to GDSII)

The digital design flow automates the transformation of a high-level behavioral description (Register Transfer Level or RTL) into a physical geometric representation (GDSII) suitable for photolithography. This process involves a chain of complex transformations, each handled by specialized tools.

#### 2.1.1 Logic Synthesis: Yosys

Yosys serves as the foundational front-end for the vast majority of open-source digital flows. It acts as the bridge between human-readable hardware description languages (Verilog) and the technological primitives available in a semiconductor process.

- **Functionality:** Yosys parses Verilog code and converts it into an internal Abstract Syntax Tree (AST). It performs technology-independent optimizations (such as constant propagation and dead code elimination) before mapping the logic to the standard cell library provided by the PDK. This mapping process uses the ABC algorithm (developed at UC Berkeley) to optimize for area, speed, or power.

- **Maturity & Adoption:** Yosys is production-ready and widely adopted in both academia and industry. It is the default synthesis engine for the OpenROAD flow, OpenLane, and several FPGA toolchains (like SymbiFlow). Its robustness is evidenced by its ability to handle complex designs, including RISC-V cores like the Ibex and PicoRV32.

- **Mechanism:** It operates by generating a BLIF (Berkeley Logic Interchange Format) or Verilog netlist that instantiates specific gates (e.g., `sky130_fd_sc_hd__nand2_1`) rather than generic boolean operators.

#### 2.1.2 Physical Implementation: OpenROAD

OpenROAD represents a paradigm shift in physical design, driven by the DARPA IDEA program's goal of achieving a "no-human-in-the-loop" hardware compiler capable of generating a GDSII from RTL in under 24 hours.

**Functionality:** OpenROAD is not merely a script but a monolithic application that integrates multiple engines into a shared database (OpenDB). It handles:

- **Floorplanning:** Defining the die area and placing I/O pins (using ioPlacer)
- **Global Placement:** Approximating the location of instances to minimize wirelength (using RePlAce)
- **Detailed Placement:** Snapping cells to manufacturing grids and ensuring no overlaps (using OpenDP)
- **Clock Tree Synthesis (CTS):** Inserting buffers to distribute the clock signal evenly (using TritonCTS)
- **Routing:** Connecting the pins with metal wires, first globally (FastRoute) and then in detail (TritonRoute)

**Architecture:** By unifying these tools into a single C++ application with Tcl and Python bindings, OpenROAD eliminates the file parsing overhead and fragility associated with older, script-based flows. It provides real-time access to the design database, allowing for incremental optimization.

#### 2.1.3 The Flow Controller: OpenLane

While OpenROAD performs the heavy lifting of physical design, OpenLane serves as the overarching "manager" or "flow controller" that orchestrates the entire process from RTL to GDSII.

**Functionality:** OpenLane automates the execution of Yosys, OpenROAD, Magic, KLayout, and verification tools. It manages hundreds of configuration variables (e.g., core utilization, clock period, metal layer limits) to tailor the flow for specific designs and PDKs.

**Evolution:**

- **OpenLane v1:** The current stable standard, widely used for the Google/Efabless OpenMPW shuttles. It is Tcl-heavy and tightly coupled to the Efabless tape-out requirements.
- **OpenLane v2 / LibreLane:** A complete rewrite in Python, designed for greater modularity and flexibility. It treats flow steps as objects, allowing users to inject custom Python scripts or swap tools easily—a feature critical for integrating non-standard CIM macros.

**Maturity:** OpenLane is tape-out proven, having facilitated hundreds of successful designs on the SkyWater 130nm process.

### 2.2 Analog and Mixed-Signal Design Flow

Unlike the digital flow, the analog flow prioritizes simulation accuracy and manual layout control, essential for defining the sensitive voltage-current relationships in FeFETs and crossbar arrays.

#### 2.2.1 Schematic Capture: Xschem

Xschem has emerged as the preferred schematic editor for open-source analog design due to its performance and modern feature set.

- **Functionality:** It allows designers to draw hierarchical circuits using symbols representing PDK devices. Crucially, it generates netlists in multiple formats (SPICE, Verilog, VHDL) suitable for simulation or Layout-Versus-Schematic (LVS) checks.

- **Capability:** Xschem can handle extremely large netlists that would crash older tools like XCircuit. It supports "back-annotation," where simulation results (voltages, operating points) are overlaid directly onto the schematic, aiding in debugging.

- **Integration:** It is tightly integrated with the SkyWater 130nm PDK, with pre-configured symbol libraries that match the foundry's device models.

#### 2.2.2 Circuit Simulation: ngspice

ngspice is the open-source equivalent of industry-standard SPICE simulators (like Spectre or HSPICE). It is the engine that calculates the physics of the circuit.

- **Functionality:** It solves systems of non-linear differential equations to predict circuit behavior in the time and frequency domains. It supports standard transistor models (BSIM3, BSIM4, BSIM-CMG) required for modern PDKs.

- **Emerging Tech Support:** A critical recent development is the integration of the OSDI (Open Source Device Interface) and the OpenVAF compiler. This allows ngspice to load compiled Verilog-A models at runtime. Since FeFETs are often modeled using Verilog-A (due to their complex hysteresis physics), this feature makes ngspice the only viable open-source tool for simulating ferroelectric devices.

- **Maturity:** Highly mature, with decades of development. It is the default simulator for Xschem and KiCad.

#### 2.2.3 Layout Design: Magic VLSI and KLayout

Physical layout in the open world relies on two complementary tools: Magic and KLayout.

- **Magic VLSI:** A "corner-stitching" layout editor that provides real-time Design Rule Checking (DRC). It is unique in that it understands connectivity and abstract layers, making it excellent for drawing cells while ensuring they are manufacturable. It is the sign-off tool for SkyWater 130nm DRC.

- **KLayout:** A polygon-based viewer and editor that excels at handling massive GDSII files (full chips). It is scriptable via Python and Ruby, making it a powerful platform for procedurally generating structures like memory crossbars or running complex DRC/LVS scripts.

- **Usage Pattern:** Designers typically draw individual cells (like a FeFET bitcell) in Magic to ensure local DRC compliance, then use KLayout to assemble the full array or view the final GDSII generated by OpenROAD.

### 2.3 Verification Tools

Verification ensures that the design is physically manufacturable (DRC) and electrically correct (LVS and Timing).

- **Design Rule Checking (DRC):** Magic and KLayout both perform DRC. Magic does it interactively; KLayout runs batch scripts. These checks ensure metal lines aren't too close, transistors have proper implants, and densities meet foundry specs.

- **Parasitic Extraction (PEX):** Magic and OpenRCX (part of OpenROAD) extract the resistance and capacitance of the drawn wires. This "parasitic" information is fed back into simulation to predict realistic delays.

- **Layout Versus Schematic (LVS):** Netgen is the standard open-source LVS tool. It compares the SPICE netlist extracted from the layout against the source schematic netlist. If they match, the layout is a faithful representation of the circuit.

- **Timing Analysis:** OpenSTA performs Static Timing Analysis (STA), checking setup and hold times across all corners (Process, Voltage, Temperature) to ensure the chip will run at the desired frequency.

---

## 3. Key Open Source Projects Comparison

The following table synthesizes the key attributes of the major open-source projects relevant to CMOS and CIM design.

| Project | Primary Function | License | Language | Maturity | CIM Relevance |
|---------|------------------|---------|----------|----------|---------------|
| **OpenROAD** | Digital P&R (Physical Design) | BSD-3 | C++, Tcl, Python | Production | Essential for placing digital control logic around CIM macros |
| **Magic VLSI** | Layout Editor & DRC | Berkeley | C, Tcl | Production | Critical for drawing custom FeFET bitcells and checking DRC |
| **ngspice** | Analog Simulator | BSD/GPL | C | Production | The only simulator capable of running Verilog-A FeFET models via OSDI |
| **Xschem** | Schematic Capture | GPL | C, Tcl | Production | Used to design the analog interface circuits (ADCs/DACs) for CIM |
| **KLayout** | Layout Viewer/Editor | GPL | C++, Python | Production | Best for procedurally generating large crossbar arrays via scripts |
| **Yosys** | RTL Synthesis | ISC | C++ | Production | Synthesizes the digital controller logic for the CIM accelerator |
| **OpenLane** | Flow Controller | Apache 2.0 | Python, Tcl | Production | Automates the integration of digital blocks; requires custom config for CIM |
| **SKY130 PDK** | 130nm CMOS PDK | Apache 2.0 | - | Production | The standard process. Requires custom modeling for FeFETs |
| **GF180MCU** | 180nm CMOS PDK | Apache 2.0 | - | Production | High voltage options (5V/6V) are excellent for FeFET programming |
| **IHP SG13G2** | 130nm BiCMOS PDK | Apache 2.0 | - | Beta/Prod | Includes RRAM (Memristor) support, closest to native CIM |

> **Insight:** While the digital stack (OpenROAD/OpenLane) is highly integrated, the analog/memory stack relies on loose coupling between Xschem, ngspice, and Magic. For CIM design, this means the user must act as the "integrator," manually ensuring that the simulation models in ngspice match the physical layout drawn in Magic.

---

## 4. The Gap: CIM/FeCIM-Specific Tools and Solutions

A significant "tooling gap" exists between standard CMOS EDA (optimized for boolean logic) and the requirements of Compute-in-Memory (optimized for analog accumulation). Standard digital tools assume signals are either 0 or 1; CIM relies on current summation (ΣI = ΣG·V), which is inherently analog.

### 4.1 Missing Capabilities in Open Source EDA

- **Crossbar Compilers:** There is no open-source equivalent to a "Memory Compiler" for CIM. Standard memory compilers generate SRAM blocks. A CIM compiler needs to generate a crossbar array, row drivers (DACs), and column readouts (ADCs/SAs) automatically. Currently, this must be done manually or via custom Python scripts in KLayout.

- **FeFET Standard Cells:** Open PDKs (SKY130/GF180) do not contain "FeFET" cells. The foundry provides standard NMOS/PMOS. To make a FeFET, a designer must either (a) use a specific layer combination that the fab identifies as ferroelectric (if available), or (b) design a standard transistor and assume post-processing.

- **Large-Scale Array Simulation:** SPICE (ngspice) scales poorly with matrix size (O(N²) or worse). Simulating a 256×256 crossbar with transient physics models for every FeFET is computationally prohibitive for full-chip verification.

### 4.2 Academic Solutions (The Bridge)

To bridge this gap, the research community has developed high-level modeling tools that estimate CIM performance without full circuit simulation.

#### 4.2.1 NeuroSim (Georgia Tech)

- **Methodology:** Circuit-level macro modeling. It uses analytical equations and look-up tables calibrated against SPICE data to estimate area, latency, and energy.
- **Validation:** NeuroSim has been validated against actual silicon data (e.g., TSMC 40nm RRAM macros), showing error margins within 5-10% for key metrics.
- **Relevance:** It is the standard for benchmarking. However, it outputs metrics (spreadsheets), not designs (GDSII). It helps architect the CIM macro but does not build it.

#### 4.2.2 CiMLoop (MIT)

- **Methodology:** System-level statistical modeling. Unlike NeuroSim's circuit focus, CiMLoop models the data flow through the entire memory hierarchy, capturing the interaction between workload (e.g., ResNet-50) and hardware.
- **Advantage:** It uses a statistical energy model that is orders of magnitude faster than cycle-accurate simulation, allowing for rapid design space exploration (e.g., optimizing ADC resolution vs. array size).
- **Status:** Active development, open-source (GitHub), and Python-based.

#### 4.2.3 PUMA (Purdue/HP)

- **Methodology:** ISA and Compiler. PUMA defines an instruction set for memristor accelerators and provides a compiler to map neural networks onto this ISA.
- **Relevance:** It addresses the programmability gap. While NeuroSim models the hardware, PUMA models how software interacts with that hardware.

#### 4.2.4 FeFET Specific Modeling

Since native PDK support is absent, FeFETs are modeled using Verilog-A.

- **Models:** The most common are the Preisach model (hysteresis based on domains) and the Landau-Khalatnikov (LK) model (physics-based thermodynamics).
- **Workflow:** These models are compiled using OpenVAF into an object file (`.osdi`) and loaded into ngspice. This enables physics-accurate simulation of a single cell or small array.

---

## 5. PDK (Process Design Kit) Basics

### 5.1 What is a PDK?

A Process Design Kit (PDK) is the interface between the designer and the foundry. It contains the files necessary to ensure a design can be manufactured and will work as simulated.

| Component | Description |
|-----------|-------------|
| **Device Models** (`.lib`, `.spice`) | Mathematical descriptions of transistors for simulation |
| **Layout Rules** (`.tech`, DRC decks) | Definitions of minimum widths, spacings, and enclosures for each layer (Metal 1, Poly, Diffusion) |
| **Technology Files** (`.lef`) | Abstract representations of layers used by automated routers |
| **P-Cells** (Parameterized Cells) | Scripts (usually in Python or Tcl) that automatically draw a device (e.g., a transistor with width W) in the layout editor |

### 5.2 Open Source PDK Analysis for FeCIM

#### 5.2.1 SkyWater SKY130 (130nm)

- **Type:** 130nm CMOS with 5 metal layers
- **FeFET Potential:** While it has no native FeFET, it includes SONOS (Silicon-Oxide-Nitride-Oxide-Silicon) Flash memory primitives. These operate on similar principles (charge trapping vs. polarization) and utilize high-voltage functionality. The High-Voltage (HV) modules (up to 20V) are critical for generating the write pulses needed to switch ferroelectric domains.
- **Availability:** Fully open source on GitHub

#### 5.2.2 GlobalFoundries GF180MCU (180nm)

- **Type:** 180nm legacy node, optimized for MCUs and PMIC (Power Management)
- **FeFET Potential:** It features robust High-Voltage (HV) transistors (5V, 6V, 10V) natively. This makes it an excellent candidate for the peripheral driving circuits of a FeFET array, which often require voltages higher than the 1.8V standard logic.
- **Availability:** Open source, actively supported by Google

#### 5.2.3 IHP SG13G2 (130nm BiCMOS)

- **Type:** High-performance BiCMOS (Bipolar + CMOS)
- **FeFET Potential:** This is the only open PDK with an experimental RRAM (Resistive RAM) module explicitly exposed. While RRAM is memristive and not ferroelectric, the circuit topologies (crossbars, sense amps) are nearly identical. It serves as the closest reference implementation for a CIM array in the open ecosystem.

> **Conclusion:** For a FeFET project, one must use the CMOS transistors from these PDKs for the peripheral logic (drivers/muxes) and instantiate a "black box" or custom Verilog-A model for the FeFET device itself.

---

## 6. Learning Path: From Zero to Chip Designer

The following curriculum leverages the "Zero to ASIC" methodology, structured to guide a beginner from basic concepts to a complex CIM project.

### Phase 1: Foundations (Weeks 1-2)

**Objectives:** Understand the transistor, the gate, and the toolchain.

**Curriculum:**

- **Theory:** "Digital Integrated Circuits" (Rabaey) - Chapters on MOS physics and CMOS inverters
- **Tools:** Install Docker. Pull the IIC-OSIC-TOOLS image, which contains the entire suite (OpenROAD, OpenLane, Xschem, ngspice, Magic) pre-installed and configured
- **Tutorials:** Matt Venn's "Zero to ASIC" course (highly recommended intro) and the "Open Source Silicon" YouTube channel. For a practical breakdown of this workflow, see our internal guide: [Zero to ASIC: A Practical Field Guide](../guides/zero-to-asic.md).

### Phase 2: The First "Tapeout" (Weeks 3-4)

**Objectives:** Complete the full RTL-to-GDSII cycle.

**Project:** A simple Digital Counter

**Workflow:**

1. **Design:** Use Wokwi (browser-based simulation) to verify logic
2. **Synthesis:** Write the Verilog. Use OpenLane to synthesize it to the SKY130 library
3. **Physical:** Let OpenLane run OpenROAD to place and route the design
4. **Submission:** Submit to Tiny Tapeout. This platform aggregates designs onto a single chip, reducing costs to ~$100. This provides the psychological win of "taping out"

### Phase 3: Intermediate - Analog & Layout (Weeks 5-8)

**Objectives:** Break out of the digital abstraction; touch the physics.

**Project:** A 1T-1FeFET Memory Cell Simulation

**Workflow:**

1. **Simulation:** Use Xschem and ngspice. Download a FeFET Verilog-A model (e.g., Scalable-FeFET). Compile it with openvaf. Simulate the hysteresis loop (Voltage vs. Polarization)
2. **Layout:** Use Magic VLSI to draw the physical layout of the cell. Learn to pass DRC (design rules)
3. **LVS:** Use Netgen to verify that your Magic layout matches your Xschem schematic

---

## 7. How Real Chips Get Made: The Flow

The transformation from idea to silicon follows a rigid sequence. Open-source tools now cover every step of this process, though with varying degrees of automation for analog designs.

| Step | Description | Open Source Tool | Commercial Equivalent | Capability |
|------|-------------|------------------|----------------------|------------|
| 1. Specification | Defining architecture, PPA targets | CiMLoop (Python) | SystemC, Excel | High |
| 2. Design Entry | Writing Verilog or drawing schematics | Yosys (Digital), Xschem (Analog) | Design Compiler, Virtuoso | High |
| 3. Verification | Ensuring logic/physics are correct | Verilator, ngspice | VCS, Spectre, MMSIM | High |
| 4. Synthesis | Mapping RTL to logic gates | Yosys | Genus, Design Compiler | High |
| 5. Place & Route | Placing gates and wiring them | OpenROAD | Innovus, ICC2 | High (Digital) |
| 6. Layout (Custom) | Drawing custom cells (memory/analog) | Magic, KLayout | Virtuoso Layout | High (Manual) |
| 7. Sign-off | DRC (Rules) and LVS (Connectivity) | Magic, Netgen, KLayout | Calibre, PVS | Medium |
| 8. Tapeout | File aggregation and mask generation | Tiny Tapeout, OpenLane | Foundry Portal | Medium |
| 9. Fabrication | Physical manufacturing | None (Requires Fab) | TSMC, GlobalFoundries | N/A |
| 10. Testing | Validating the physical chip | PyTest, Custom PCBs | Advantest ATE | Medium |

> **Insight:** Steps 6 and 7 are the bottleneck for FeCIM. OpenROAD cannot automatically route a custom analog crossbar. This requires scripting in KLayout (Python) to procedurally generate the array, which is then treated as a "Macro" (black box) by OpenROAD for the rest of the digital integration.

---

## 8. Integration Strategy: Connecting Module 6 to Real EDA

Module 6 (FeCIM Array Builder) currently generates OpenLane-compatible EDA files (LEF, Liberty, Verilog, DEF). The strategies below describe **potential future enhancements** to connect more deeply with the open EDA stack.

### Current State (Implemented)

| Capability | Status |
|------------|--------|
| LEF cell abstract | Done |
| Liberty timing file | Done (placeholder values) |
| Verilog netlist | Done (behavioral only) |
| DEF placement | Done |
| OpenLane config | Done |

### Strategy 1: Python to SPICE (Future Enhancement)

Module 6 could be extended to act as a schematic capture tool.

**Mechanism:** Use the `spicelib` or `PySpice` Python libraries.

**Workflow:**

1. User configures crossbar parameters (size, conductance states) in your GUI
2. Your tool generates a `.spice` netlist file. This file instantiates the SKY130 transistor models for drivers and your custom `.osdi` model for the FeFETs
3. Your tool invokes ngspice in batch mode (or shared library mode) to run a transient simulation
4. Parse the raw output file to visualize the inference accuracy or current summing behavior

### Strategy 2: Python to GDSII (Future Enhancement)

Module 6 could be extended to directly generate manufacturable layouts.

**Mechanism:** Use the KLayout Python API or `gdsfactory`.

**Workflow:**

1. Define the geometric rules (from the PDK LEF file) in Python (e.g., `poly_width = 0.15um`)
2. Script a nested loop to draw the FeFET array: place the active layer, gate layer, and orthogonal metal1 (bitlines) and metal2 (wordlines)
3. Export a `.gds` file. This GDS is now a "Hard Macro" that can be instantiated in OpenLane

### Strategy 3: Architecture Exploration Integration

**Mechanism:** Export configurations for CiMLoop.

**Workflow:** Your tool can export a YAML configuration file describing the array architecture. The user then runs CiMLoop to get high-level energy/area estimates, validating the architectural choices before circuit-level design begins.

---

## 9. Companies & Communities

The ecosystem is sustained by a mix of commercial entities and non-profit foundations.

### Active Organizations

| Organization | Description |
|--------------|-------------|
| **Efabless** | The platform operator for the OpenMPW program. They provide the marketplace and the cloud infrastructure that runs OpenLane. They are the primary hub for open silicon fabrication. |
| **FOSSi Foundation** | The advocacy group behind the movement. They organize ORConf and maintain the LibreCores repository. |
| **ChipFlow** | A startup focused on simplifying the flow using Python-based HDLs (Amaranth). They aim to make chip design as accessible as PCB design, targeting OEMs rather than just chip architects. |
| **Zero ASIC** | Focused on education and reducing the barrier to entry, closely linked to the "Zero to ASIC" course. |

### Communities

| Community | Description |
|-----------|-------------|
| **Tiny Tapeout Discord** | The most active community for beginners. It features channels dedicated to tools (Wokwi, OpenLane) and specific PDKs. |
| **SkyWater PDK Slack** | Technical discussion involving the PDK maintainers and analog designers. |
| **OpenROAD GitHub Discussions** | The place for deep technical issues regarding the P&R flow. |

---

## 10. Realistic Assessment

### What Can Be Done Today?

- **Digital Logic:** One can design, verify, and manufacture a digital RISC-V processor or a standard neural network accelerator (systolic array) using purely open-source tools (OpenLane + SKY130) with high confidence. The flow is production-ready.

- **Analog Simulation:** Detailed simulation of FeFET arrays is possible using ngspice + OpenVAF, provided the user has a valid Verilog-A model.

### What Cannot Be Done Today (The Gap)?

- **Push-Button FeCIM:** There is no "OpenROAD for Analog." You cannot click a button and get a routed FeFET crossbar. This layout must be drawn by hand or by custom scripts.

- **Native FeFET Fabrication:** The open PDKs (SKY130/GF180) are CMOS-only. "Taping out" a FeFET design today effectively means taping out the CMOS control circuitry with empty slots (or standard transistors) where the FeFETs would go, or establishing a private partnership with a research fab (like Fraunhofer or extensive post-processing).

### Educational Tool Opportunity

This project—a "FeCIM Visualizer and Compiler"—aims to help bridge this gap for educational purposes. By generating SPICE netlists and basic layout scripts from a GUI, it provides a learning tool for understanding how FeCIM designs might integrate with open-source EDA flows.

**Important:** This is educational/research software, not a production design tool. Generated files require significant validation before any real use.

---

## 11. Resource List

### Tools & Code

| Tool | URL |
|------|-----|
| ngspice | [ngspice.sourceforge.io](https://ngspice.sourceforge.io) |
| KLayout | [klayout.de](https://klayout.de) |
| CiMLoop | [github.com/mit-emze/cimloop](https://github.com/mit-emze/cimloop) |
| OpenROAD | [theopenroadproject.org](https://theopenroadproject.org) |
| SkyWater PDK | [github.com/google/skywater-pdk](https://github.com/google/skywater-pdk) |

### Education

| Resource | URL |
|----------|-----|
| Zero to ASIC Course | [zerotoasiccourse.com](https://zerotoasiccourse.com) |
| **Internal Guide** | [Zero to ASIC: A Practical Field Guide](../guides/zero-to-asic.md) |
| Tiny Tapeout | [tinytapeout.com](https://tinytapeout.com) |

### Key Figures

| Person | Contribution |
|--------|--------------|
| **Matt Venn** | Educator, founder of Zero to ASIC and Tiny Tapeout |
| **Tim Edwards** | Maintainer of Magic VLSI and Open_PDKs; the architect of the open analog flow |
| **Stefan Schippers** | Creator of Xschem |
| **Mohamed Shalan** | Principal architect of OpenLane |

> This ecosystem, while young, provides all the necessary primitives. The challenge—and the opportunity—lies in integration. By connecting your educational tools to these robust backends, you bridge the gap between abstract research and physical realization.

---

---

## 12. Recent Developments (2025-2026 Update)

This section extends the original research with the latest developments in the open-source EDA ecosystem.

### 12.1 OpenROAD and OpenLane Evolution

The OpenROAD project has made significant strides since 2024, expanding product capabilities and global deployment across both industry and educational programs.

#### 12.1.1 Industrial Adoption

- **Infineon** is now using OpenROAD and OpenLane for experimental and innovative projects, building initial prototypes using open-source tools before cross-checking with commercial tools for production.
- **Intel** has showcased GenAI integration with OpenROAD to drive intelligent and efficient chip design methodologies.
- **IHP** now offers shuttle services for fab-ready designs using the OpenROAD flow directly.
- Thousands of RISC-V projects, including RI5CY, MEMS controllers, and AES cores, have been successfully taped out.

#### 12.1.2 Advanced Node Progress

OpenROAD has been applied in advanced nodes of academic RISC-V initiatives. The **BlackParrot 12nm** open-source processor utilized OpenROAD's RTL-MP for its floorplan, demonstrating the tool's capability at modern technology nodes.

#### 12.1.3 Machine Learning Integration

OpenROAD's **AutoTuner** now incorporates machine-learning architecture that methodically explores tool settings in the cloud, significantly reducing the need for expert hand-tuning and lowering the barrier to entry for silicon design.

### 12.2 Ecosystem Changes: Post-Efabless Era

A major disruption occurred in early 2025 when **Efabless**, the open-source chip pioneer based in Palo Alto, shut down after failing to complete its latest funding round. This had significant implications for the ecosystem:

#### 12.2.1 Tiny Tapeout Transition

- The project transitioned from SkyWater Technologies via Efabless to **IHP** (Leibniz Institute for High Performance Microelectronics).
- New chip loan terms were introduced: chips remain IHP property with a default 2-year loan period.
- **SwissChips Initiative** stepped in to sponsor shuttles, covering manufacturing costs for Swiss residents.

#### 12.2.2 Active Shuttle Programs (2025-2026)

| Shuttle | Process | Launch Date | Closing Date | Designs | Expected Delivery |
|---------|---------|-------------|--------------|---------|-------------------|
| **TTIHP26a** | IHP 130nm | Nov 2025 | Mar 2026 | Open | Sep 2026 |
| **TTSKY25b** | SKY130 | Sep 2025 | Nov 2025 | 316 | Jun 2026 |
| **TTSKY25a** | SKY130 | Jun 2025 | Sep 2025 | 237 | May 2026 |
| **TTIHP25a** | IHP 130nm | Mar 2025 | Mar 2025 | 548 | Dec 2025 |

### 12.3 New Simulation and Modeling Tools

#### 12.3.1 CIM Simulation Frameworks

The landscape of Compute-in-Memory simulation tools has expanded significantly:

| Tool | Institution | Focus | Key Capability |
|------|-------------|-------|----------------|
| **NeuroSim** | Georgia Tech | Circuit-level macro modeling | Validated within 5-10% of silicon |
| **CiMLoop** | MIT | System-level statistical modeling | Orders of magnitude faster than cycle-accurate |
| **PIMSimulator** | Academic | Cycle-accurate DRAM-based PIM | DRAMSim2-backed HBM2-class modeling |
| **3D CiM LLM Sim** | Academic | LLM inference on 3D AIMC | Dense and MoE architecture support |
| **CINM (Cinnamon)** | Academic | Compilation infrastructure | Up to 51× performance improvement |
| **PIMCOMP** | Academic | End-to-end DNN compiler | Generic multi-level optimization |
| **FAST** | Academic | Hardware-software co-design | Ultra-fast training with non-ideality modeling |

#### 12.3.2 CINM (Cinnamon) Compiler

A notable addition is **CINM**, a compilation tool that can modify acceleratable LLVM IRs to offload them to CIM accelerators automatically, without re-engineering legacy software. Experimental results show:
- Up to **51× performance improvement** compared to x86 processors
- Up to **309× energy efficiency improvement**
- Automatic translation of legacy programs to CIM-supported binary executables

#### 12.3.3 FeFET Simulation Advances

Recent 2025 work has improved FeFET simulation capabilities:

- **Cadence Virtuoso** workflows now support standalone FeCap models using Landau-Khalatnikov equations
- **TCAD Sentaurus** integration with calibrated Preisach model parameters for realistic FE behavior
- Design-specific thickness thresholds (~55nm) identified for reliable bistability and retention
- FeFET-based CIM unit circuits demonstrated in HfO₂-doped configurations

### 12.4 GDSFactory: Layout Generation Evolution

**GDSFactory** has matured into a comprehensive Python library for chip design (version 9.31.0 as of late 2025):

#### Key Features:
- **Backend Upgrade**: Migrated from gdstk to KLayout C++ library for enhanced performance
- **Multi-domain Support**: Photonics, Analog, Quantum, MEMS, PCBs, and 3D-printable objects
- **Simulation Integration**: Direct integration with Ansys, Lumerical, Tidy3d, MEEP, DEVSIM, SAX, MEOW, and Xyce
- **Built-in Verification**: DRC, DFM, LVS, connectivity checks, and dummy fill
- **Multi-format Output**: GDSII, OASIS, STL, and GERBER

**GDSFactory+** now offers an enterprise GUI for chip design built on VSCode, providing foundry PDK access, schematic capture, and industry-level support.

### 12.5 OpenVAF and Verilog-A Ecosystem

The Verilog-A compact model ecosystem has stabilized with important developments:

#### 12.5.1 OpenVAF-Reloaded
- **OSDI 0.4 API** version now available for modern compact models
- Pre-compiled models available for Linux (Ubuntu 22.04) and Windows 10/11 (64-bit)
- Simulation speed comparable to (sometimes faster than) built-in C-coded models
- **ADMS officially deprecated** due to maintenance issues and buggy output

#### 12.5.2 ngspice Updates (v43-44)
- OSDI, KLU, readline, OpenMP, and XSPICE defined as standard configure options
- Code model `d_cosim` supports **Icarus Verilog** for mixed-signal simulation
- Access to all modern compact models: short channel MOS, FinFETs, double gate transistors, SiGe bipolars, III-V HEMTs

### 12.6 IHP SG13G2 PDK Updates

The IHP Open Source PDK has seen significant activity in 2025:

#### 12.6.1 Shuttle Program
- **April 2025** (Testfield T586): Active submissions
- **May 2025** (Testfield T588): Open for submissions
- **September 2025**: Initial PR by Sep 1, final GDS by Sep 14

#### 12.6.2 RRAM/Memristor Module
IHP offers a fully CMOS-integrated memristive module based on resistive **TiN/HfO₂-x/TiN** switching devices in **SG13S** technology (note: SG13S, not SG13G2). A PDK including layout and Verilog-A simulation model is available upon request.

#### 12.6.3 Magic Support
A second flow supporting Magic VLSI is planned for 2025, expanding layout options beyond KLayout.

### 12.7 Chiplet and UCIe Developments

The chiplet ecosystem has made significant progress toward open standards:

#### 12.7.1 Open Source UCIe Simulation
**Zero ASIC** has developed open-source simulation capabilities for UCIe using:
- **Verilator**, **Xyce**, **Icarus**, and **Switchboard** (open-source high-performance co-simulation library)
- Analog simulation of signal chains including TX predriver, TX driver with impedance control
- Channel models based on S-parameter data from the UCIe organization
- Results parsing using **Spyci**, formatted with NumPy and Matplotlib

#### 12.7.2 Standards Progress
- **UCIe 2.0** support added to commercial tools (Keysight Chiplet PHY Designer 2025)
- **AMBA CHI C2C** from Arm: open-source protocol for coherent, low-latency chiplet communication
- **Open Domain-Specific Architecture (ODSA)** Project: 50+ companies, six work efforts

#### 12.7.3 Current Gaps
UCIe does not yet fully support:
- Power delivery networks (PDN)
- Chiplet verification flow automation
- Design automation tools integration
- Firmware interoperability

### 12.8 RISC-V Silicon Verification

#### 12.8.1 Notable Tapeouts
- **Basilisk**: First end-to-end open-source, Linux-capable RISC-V SoC in IHP's open 130nm technology
  - 64-bit RISC-V core
  - Fully digital HyperRAM DRAM controller
  - USB 1.1 and VGA peripherals

- **preDRAC**: First RISC-V processor designed and fabricated by Spanish/Mexican academic institutions
  - Joint development by BSC, CIC-IPN, IMB-CNM (CSIC), and UPC
  - CMOS 65nm technology

- **OpenTitan®**: World's first open-source silicon root of trust
  - **Ibex®**: Production-quality, formally verified 32-bit RISC-V core in production

#### 12.8.2 Verification Landscape
- Open-source starting points: Verilator, cocotb, RISC-V DV
- Commercial solutions for production: ImperasDV, STING with VCS® simulation, VC Formal™
- Tiny Tapeout RISC-V peripheral crowdsourcing initiative on TTSKY25a

### 12.9 Analog Neural Network Accelerator Tools

#### 12.9.1 Open Source Frameworks

| Tool | Provider | Target | Status |
|------|----------|--------|--------|
| **AIHWKIT** | IBM | Hardware-aware training for analog AI | Active, documented |
| **NVDLA** | NVIDIA | Standard DNN inference accelerator | Open architecture |
| **DnnWeaver** | Academic | FPGA DNN acceleration | Caffe-to-Verilog |
| **PipeCNN** | Academic | OpenCL-based FPGA CNN | HLS-based |
| **Ecko Project** | Academic | Open-source NN accelerator design | Sky130 validated |

#### 12.9.2 Ecko Project
The Ecko project leverages automated, community-driven methodologies for NN accelerator design using publicly available tools:
- Successfully generated layouts for two small NN examples (<1000 parameters) on **Sky130**
- Incorporates dense, activation, and 2D convolution layers
- Demonstrates that current open-source tools effectively automate low-complexity neural network architectures

### 12.10 Memristor Crossbar Design Advances

#### 12.10.1 Simulation Tools
- **FAST (Functional Array Simulator)**: End-to-end functional simulator for precise training, mapping, and evaluation on memristor crossbars
- **SySCIM**: SystemC-AMS simulation tool for memristive CIM
- **SEMulator3D**: Virtual fabrication for 3D ReRAM architecture pathfinding
- **PIM-HLS**: Automatic hardware generation for heterogeneous PIM-based NN accelerators

#### 12.10.2 Key Design Challenges Addressed
Modern tools now handle:
- **IR-drop** modeling and compensation
- **Device-to-device (D2D) variation** simulation
- **Stuck-at-fault (SAF)** detection and mitigation
- **Sneak-path current** analysis in crossbar arrays
- **Multi-layer stacking** for 3D integration

#### 12.10.3 Emerging Materials (2025)
- **2D MoS₂-based memristive crossbar arrays** demonstrated for synaptic applications
- **Roll-to-roll mechanical exfoliation** combined with inkjet printing for scalable fabrication
- Focus on improving device yield and reducing C2C/D2D variability

---

## 13. Updated Integration Strategy

Based on the 2025-2026 developments, the integration strategy for connecting visualization tools with the EDA stack can be updated:

### 13.1 Enhanced Python-to-SPICE Workflow

With ngspice 43/44 and OpenVAF-reloaded:

```
1. User configures crossbar in GUI
2. Tool generates .spice netlist with:
   - SKY130/GF180/IHP SG13G2 transistor models for drivers
   - Pre-compiled .osdi FeFET models (OpenVAF-reloaded)
3. Invoke ngspice with OSDI enabled
4. Parse output with Python (NumPy/Matplotlib)
```

### 13.2 Enhanced Python-to-GDSII Workflow

With GDSFactory 9.x:

```python
import gdsfactory as gf

# Define crossbar using KLayout backend
@gf.cell
def fefet_crossbar(rows: int, cols: int, pitch: float):
    c = gf.Component()
    # Procedural array generation
    # Built-in DRC checking
    # Export to GDSII/OASIS
    return c
```

### 13.3 New CIM Modeling Integration

```
1. Architecture exploration → CiMLoop (YAML export)
2. Hardware-software co-design → CINM compiler
3. Detailed simulation → FAST simulator
4. Physical design → GDSFactory → OpenLane macro
```

### 13.4 Recommended Tool Stack (2026)

| Stage | Primary Tool | Alternative |
|-------|--------------|-------------|
| Architecture | CiMLoop | NeuroSim |
| RTL Design | Yosys + Verilator | SpinalHDL |
| Analog Sim | ngspice + OpenVAF | Xyce |
| Layout Gen | GDSFactory | KLayout Python |
| P&R | OpenROAD | OpenLane 2.0 |
| Verification | Magic + Netgen | KLayout DRC |
| Tapeout | Tiny Tapeout (IHP) | IHP direct shuttle |

---

## 14. Future Outlook

### 14.1 Emerging Trends

1. **AI-Assisted EDA**: OpenROAD's AutoTuner represents the beginning of ML-driven optimization in open-source EDA
2. **3D Integration**: Multi-layer memristor stacking and chiplet ecosystems are converging
3. **Standardization**: UCIe and AMBA CHI C2C are enabling interoperable chiplet designs
4. **FeFET Maturation**: HfO₂-based FeFETs demonstrated in 28nm and 14nm FDSOI nodes

### 14.2 Remaining Gaps

| Gap | Current State | Expected Resolution |
|-----|--------------|---------------------|
| Native FeFET PDK | Custom Verilog-A models | Research fab partnerships |
| Push-button CIM layout | Manual/scripted | Community tool development |
| Large-scale crossbar sim | O(N²) SPICE | GPU-accelerated simulators |
| Chiplet verification | Fragmented | UCIe tooling standardization |

### 14.3 The Strategic Opportunity (Revised)

The post-Efabless ecosystem has actually strengthened with:
- **IHP** providing more direct fab access with RRAM capabilities
- **SwissChips** and European initiatives funding open silicon
- **FOSSi Foundation** roadmap driving European open EDA

A FeCIM Visualizer and Compiler now has clearer integration paths:
1. **Simulation**: ngspice + OpenVAF (mature, documented)
2. **Layout**: GDSFactory + KLayout (production-ready)
3. **Fabrication**: IHP SG13G2/SG13S shuttles (RRAM-capable)
4. **Verification**: Magic + Netgen (tape-out proven)

---

## Works Cited

1. OpenLANE: The Open-Source Digital ASIC Implementation Flow - woset-workshop, https://woset-workshop.github.io/PDFs/2020/a21.pdf
2. Empowering innovation: OpenROAD and the future of open-source EDA - EE World Online, https://www.eeworldonline.com/empowering-innovation-openroad-and-the-future-of-open-source-eda/
3. OpenROAD Flow Scripts Tutorial, https://openroad-flow-scripts.readthedocs.io/en/latest/tutorials/FlowTutorial.html
4. OpenROAD – Key Milestones on the Road towards Good PPA, https://theopenroadproject.org/openroad-key-milestones-on-the-road-towards-good-ppa/
5. Magic Mai/OpenROAD - Gitee, https://gitee.com/magic3007/OpenROAD
6. librelane/librelane: ASIC implementation flow infrastructure, successor to OpenLane - GitHub, https://github.com/librelane/librelane
7. LibreLane Documentation, https://librelane.readthedocs.io/
8. Efabless Recommended Open Source Analog Design Flow, http://www.opencircuitdesign.com/analog_flow/
9. Installing and "tuning up" xschem, http://web02.gonzaga.edu/faculty/talarico/vlsi/xschemInstall.html
10. XSCHEM SKY130 INTEGRATION, https://xschem.sourceforge.io/stefan/xschem_man/tutorial_xschem_sky130.html
11. Ngspice, the open source Spice circuit simulator - Intro, https://ngspice.sourceforge.io/
12. Ngspice circuit simulator - Verilog-A with OSDI/OpenVAF - SourceForge, https://ngspice.sourceforge.io/osdi.html
13. Magic VLSI - Open Circuit Design, http://opencircuitdesign.com/magic/
14. Hi all I ve been playing with magic and klayout for a while open-source-silicon.dev #sky130, https://web.open-source-silicon.dev/t/16600034/hi-all-i-ve-been-playing-with-magic-and-klayout-for-a-while-
15. Best Custom Layout Tools to Master for VLSI Job Opportunities, https://vlsiguru.com/blog/best-tools-to-learn-in-custom-layout-course
16. NeuroSim Simulator for Compute-in-Memory Hardware Accelerator: Validation and Benchmark - PMC - NIH, https://pmc.ncbi.nlm.nih.gov/articles/PMC8219932/
17. NeuroSim Simulator for Compute-in-Memory Hardware Accelerator: Validation and Benchmark - Frontiers, https://www.frontiersin.org/journals/artificial-intelligence/articles/10.3389/frai.2021.659060/full
18. CiMLoop: A Flexible, Accurate, and Fast Compute-In-Memory Modeling Tool - IEEE Xplore, https://ieeexplore.ieee.org/document/10590023/
19. CiMLoop: A Flexible, Accurate, and Fast Compute-In-Memory Modeling Tool - arXiv, https://arxiv.org/pdf/2405.07259
20. PUMA: A Programmable Ultra-efficient Memristor-based Accelerator for Machine Learning Inference - ResearchGate, https://www.researchgate.net/publication/330725909_PUMA_A_Programmable_Ultra-efficient_Memristor-based_Accelerator_for_Machine_Learning_Inference
21. Embedding-Enhanced Probabilistic Modeling of Ferroelectric Field Effect Transistors (FeFETs) - TRACE: Tennessee Research and Creative Exchange, https://trace.tennessee.edu/cgi/viewcontent.cgi?article=15607&context=utk_gradthes
22. DavidTobar456/pfecapRevision: Verilog-A Preisach ferroelectric cap (PFECAP) simulation model for FET - GitHub, https://github.com/DavidTobar456/pfecapRevision
23. SkyWater SKY130 Process Design Rules, https://skywater-pdk.readthedocs.io/en/main/rules.html
24. Specifications For 180MCU - GF180MCU PDK - Read the Docs, https://gf180mcu-pdk.readthedocs.io/en/latest/analog/spice/elec_specs/elec_specs.html
25. SiGe:C-BiCMOS-Technologies - IHP Microelectronics, https://www.ihp-microelectronics.com/services/research-and-prototyping-service/mpw-prototyping-service/sigec-bicmos-technologies
26. Introduction to Chip Design, http://www2.imm.dtu.dk/~masca/chip-design-book.pdf
27. Zero to ASIC Course, https://www.zerotoasiccourse.com/
28. Digital Design Guide - Tiny Tapeout, https://tinytapeout.com/digital_design/
29. 4 - Submit your design :: Quicker, easier and cheaper to make your own chip! - Tiny Tapeout, https://tinytapeout.com/guides/workshop/submit-your-design/
30. spicelib - PyPI, https://pypi.org/project/spicelib/0.9.1/
31. 1. Overview — PySpice @VERSION@ documentation - Fabrice Salvaire, https://pyspice.fabrice-salvaire.fr/releases/v1.3/overview.html
32. Efabless Launches chipIgnite with SkyWater to Bring Chip Creation to the Masses, https://www.skywatertechnology.com/efabless-launches-chipignite-with-skywater-to-bring-chip-creation-to-the-masses/
33. ChipFoundry — ChipFlow, https://www.chipflow.io/chipfoundry
34. ChipFlow - Helping product companies to make their own chips, https://www.chipflow.io/
35. ZTM Community Discord. Learn Together, Grow Together. | Zero To Mastery, https://zerotomastery.io/community/developer-community-discord/
36. Tiny Tapeout :: Quicker, easier and cheaper to make your own chip!, https://tinytapeout.com/
37. OpenLane vs OpenROAD #900 - GitHub, https://github.com/The-OpenROAD-Project/OpenLane/discussions/900

### 2025-2026 Update References

38. The OpenROAD Project - Official Website, https://theopenroadproject.org/
39. OpenROAD Project - Wikipedia, https://en.wikipedia.org/wiki/OpenROAD_Project
40. Industrial Experience with Open-Source EDA Tools - ACM/IEEE MLCAD 2022, https://dl.acm.org/doi/10.1145/3551901.3557040
41. Memory Is All You Need: CIM Architectures for LLM Inference - arXiv, https://arxiv.org/html/2406.08413v1
42. CINM (Cinnamon): A Compilation Infrastructure for Heterogeneous CIM - ResearchGate, https://www.researchgate.net/publication/390679656
43. Modeling and Simulation Frameworks for PIM Architectures - arXiv, https://arxiv.org/html/2512.00096v1
44. SRAM-based CIM Literature Collection - GitHub BUAA-CI-LAB, https://github.com/BUAA-CI-LAB/Literatures-on-SRAM-based-CIM
45. FeFET-Based Computing-in-Memory Unit Circuit - PMC, https://pmc.ncbi.nlm.nih.gov/articles/PMC11858781/
46. Dual-Bit FeFET for enhanced storage and endurance - Nature npj, https://www.nature.com/articles/s44335-025-00030-8
47. Modeling of FeFETs and their Application - University of Oulu, https://oulurepo.oulu.fi/bitstream/10024/57025/1/nbnfioulu-202506164534.pdf
48. IHP Open Source PDK - GitHub, https://github.com/IHP-GmbH/IHP-Open-PDK
49. IHP Open Source PDK Documentation, https://www.ihp-microelectronics.com/services/research-and-prototyping-service/fast-design-enablement/open-source-pdk
50. IHP SG13G2 Tape Out April 2025 - GitHub, https://github.com/IHP-GmbH/TO_Apr2025
51. GDSFactory Documentation - v9.31.0, https://gdsfactory.github.io/gdsfactory/
52. GDSFactory - GitHub, https://github.com/gdsfactory/gdsfactory
53. Tiny Tapeout Chips - Official, https://tinytapeout.com/chips/
54. SwissChips TinyTapeout Shuttle Announcement, https://swisschips.ethz.ch/news-and-events/swisschips-news/2025/12/announcing-the-next-swisschips-supported-tinytapeout-shuttle-submit-your-design-today.html
55. Tiny Tapeout Opens IHP Shuttle - Hackster.io, https://www.hackster.io/news/tiny-tapeout-opens-an-ihp-shuttle-for-your-open-source-chip-designs-but-beware-the-new-terms-77e0b292cae4
56. Tiny Tapeout hit as Efabless closes - EE News Europe, https://www.eenewseurope.com/en/tiny-tapeout-hit-as-efabless-closes/
57. Ngspice Verilog-A with OSDI/OpenVAF, https://ngspice.sourceforge.io/osdi.html
58. OpenVAF Verilog-A Compiler - GitHub, https://github.com/pascalkuthe/OpenVAF
59. Ngspice News and Releases, https://ngspice.sourceforge.io/news.html
60. VA-Models: Verilog-A simulation models - GitHub, https://github.com/dwarning/VA-Models
61. Zero ASIC - Lowering the Barrier to Chiplets, https://www.zeroasic.com/blog/ucie-open-source-design
62. Building An Open Chiplet Economy - OCP, https://www.opencompute.org/blog/building-an-open-chiplet-economy
63. UCIe Consortium at OFC 2025, https://www.uciexpress.org/post/ucie-consortium-showcases-chiplet-innovation-at-ofc-2025
64. Basilisk: End-to-End Open-Source RISC-V SoC - arXiv, https://arxiv.org/html/2406.15107v1
65. lowRISC Open-Source Silicon Designs, https://lowrisc.org/
66. RISC-V in 2025: Open-Source Future of Embedded Design - Tessolve, https://embedded.tessolve.com/blogs/risc-v-in-2025-the-open-source-shift-in-embedded-system-design/
67. IBM Analog Hardware Acceleration Kit (AIHWKIT), https://aihwkit.readthedocs.io/en/latest/
68. NVIDIA Deep Learning Accelerator (NVDLA), https://nvdla.org/
69. DnnWeaver v2.0 - Open Source DNN Accelerator, http://dnnweaver.org/
70. ADNA: Automating ASIC Development of NN Accelerators - MDPI, https://www.mdpi.com/2079-9292/14/7/1432
71. Analog-AI Chip for Speech Recognition - Nature, https://www.nature.com/articles/s41586-023-06337-5
72. Optimizing hardware-software co-design for memristor crossbars - Science China, https://link.springer.com/article/10.1007/s11432-024-4240-x
73. Fast prototyping of memristors for ReRAMs - PMC, https://pmc.ncbi.nlm.nih.gov/articles/PMC12690548/
74. 2D MoS2-Based Memristive Crossbar for Synaptic Applications - ACS, https://pubs.acs.org/doi/10.1021/acsami.5c00688
75. IHP Open DesignLib Documentation, https://ihp-open-ip.readthedocs.io/en/latest/

---

## 15. Additional Papers and Resources (January 2026 Update)

This section provides an expanded collection of papers and resources organized by topic area, compiled to support deeper research into FeCIM, open-source EDA, and emerging memory technologies.

### 15.1 FeFET Modeling and Simulation

| Paper | Authors/Source | Year | Key Contribution |
|-------|---------------|------|------------------|
| [Modeling of FeFETs and their Application in Non-Volatile SRAM](https://oulurepo.oulu.fi/bitstream/10024/57025/1/nbnfioulu-202506164534.pdf) | Haidar M., University of Oulu | 2025 | Standalone FeCap model using Landau-Khalatnikov equations; Cadence Virtuoso Verilog-A implementation |
| [Temperature- and variability-aware compact modeling of ferroelectric FDSOI FET](https://www.sciencedirect.com/science/article/abs/pii/S0038110124001035) | ScienceDirect | 2024 | Preisach-based Verilog-A model with temperature and history effects |
| [Logic-in-memory application of ferroelectric WS2-channel FET](https://www.nature.com/articles/s41699-024-00466-9) | Nature npj 2D Materials | 2024 | 2D FeFET-based LiM with sub-2nm DG structures; BSIM-IMG integration |
| [AI Hardware Architecture Design Based on Logic-in-Memory FeFET at Sub-3nm Nodes](https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aisy.202400370) | Advanced Intelligent Systems | 2024-2025 | FinFET compact models; 85.2% power reduction for CNN accelerators |
| [Semi-empirical and Verilog-A compatible compact model for ferroelectric hysteresis](https://www.researchgate.net/publication/361052290) | ResearchGate | 2022 | Foundation work for circuit-simulator compatible FeFET models |
| [A Verilog-A Compact Model for Negative Capacitance FET](https://www.researchgate.net/publication/309376456) | ResearchGate | 2016 | Early NCFET Verilog-A modeling methodology |

### 15.2 Compute-in-Memory (CIM) Architecture

| Paper | Authors/Source | Year | Key Contribution |
|-------|---------------|------|------------------|
| [Memory Is All You Need: CIM Architectures for LLM Inference](https://arxiv.org/html/2406.08413v1) | arXiv | 2024 | Comprehensive overview of CIM for transformer/LLM acceleration |
| [Architecture and Programming of Analog IMC Accelerators for DNNs](https://research.ibm.com/publications/architecture-and-programming-of-analog-in-memory-computing-accelerators-for-deep-neural-networks) | IBM Research, IPDPS | 2024 | Heterogeneous programmable accelerator with 2D mesh; PCM at 14nm |
| [Novel Analog-Computing-in-Memory Architecture with Scalable Multi-Bit MAC](https://www.mdpi.com/2079-9292/14/20/4030) | MDPI Electronics | 2025 | Pipelining scheme for MAC/ADC decoupling; flexible weight organization |
| [Fast and robust analog in-memory deep neural network training](https://www.nature.com/articles/s41467-024-51221-z) | Nature Communications | 2024 | c-TTv2 and AGAD algorithms for in-memory training |
| [A compute-in-memory chip based on resistive random-access memory](https://www.nature.com/articles/s41586-022-04992-8) | Nature | 2022 | Foundational RRAM CIM chip demonstration |
| [SRAM-based CIM Literature Collection](https://github.com/BUAA-CI-LAB/Literatures-on-SRAM-based-CIM) | BUAA-CI-LAB GitHub | 2024-2025 | Comprehensive reading list for SRAM CIM research |

### 15.3 CIM Simulation and Modeling Tools

| Tool/Paper | Source | Key Capability |
|------------|--------|----------------|
| [NeuroSim V1.5: Improved Software Backbone for Benchmarking CIM Accelerators](https://arxiv.org/html/2505.02314v1) | arXiv | 2025 | TensorRT integration; transformer support; 6.5× faster runtime |
| [NeuroSim Simulator: Validation and Benchmark](https://www.frontiersin.org/journals/artificial-intelligence/articles/10.3389/frai.2021.659060/full) | Frontiers in AI | 2021 | Validated within 1% of 40nm RRAM macro post-layout |
| [CiMLoop: A Flexible, Accurate, and Fast CIM Modeling Tool](https://arxiv.org/pdf/2405.07259) | MIT, arXiv | 2024 | System-level statistical modeling; NeuroSim plug-in integration |
| [DNN_NeuroSim_V2.1 - On-chip Training](https://github.com/neurosim/DNN_NeuroSim_V2.1) | GitHub | 2024 | PyTorch interface; training-focused benchmark framework |
| [CINM (Cinnamon): Compilation Infrastructure for Heterogeneous CIM/CNM](https://arxiv.org/abs/2301.07486) | ACM ASPLOS | 2024 | MLIR-based; up to 51× performance improvement; UPMEM/memristor support |

### 15.4 Memristor Crossbar Arrays

| Paper | Authors/Source | Year | Key Contribution |
|-------|---------------|------|------------------|
| [Resistive Switching Random-Access Memory (RRAM): Applications and Requirements](https://pubs.acs.org/doi/10.1021/acs.chemrev.4c00845) | Chemical Reviews | 2024 | Comprehensive review of RRAM device engineering for IMC |
| [Optimizing hardware-software co-design for memristor crossbars](https://link.springer.com/article/10.1007/s11432-024-4240-x) | Science China | 2024 | IR-drop, D2D variation, SAF modeling |
| [Fast prototyping of memristors for ReRAMs and neuromorphic computing](https://pmc.ncbi.nlm.nih.gov/articles/PMC12690548/) | PMC | 2025 | High-throughput Ag/MoS2/Au fabrication; 10²-10⁴ resistance ratio |
| [Stochastic Yet Precise: Memristor Crossbar Arrays for In-Memory Computing](https://advanced.onlinelibrary.wiley.com/doi/10.1002/adfm.202523780) | Advanced Functional Materials | 2025 | Multi-bit analog memory; PUF; hardware-seeded GAN |
| [Purely self-rectifying memristor-based passive crossbar array](https://www.nature.com/articles/s41467-023-44620-1) | Nature Communications | 2024 | 1kb passive crossbar with self-rectifying devices |
| [Ultralow Powered 2D MoS2-Based Memristive Crossbar](https://pubs.acs.org/doi/10.1021/acsami.5c00688) | ACS Applied Materials | 2025 | 94% device yield; microsecond switching |

### 15.5 Open-Source EDA and RTL-to-GDSII

| Paper/Resource | Source | Year | Key Contribution |
|----------------|--------|------|------------------|
| [Comprehensive RTL-to-GDSII Workflow for Custom Embedded FPGA Using Open-Source Tools](https://www.mdpi.com/2079-9292/14/19/3866) | MDPI Electronics | 2025 | OpenLane + OpenFPGA methodology for SKY130 |
| [OpenLANE: The Open-Source Digital ASIC Implementation Flow](https://woset-workshop.github.io/PDFs/2020/a21.pdf) | WOSET | 2020 | Original OpenLane paper |
| [Stitching FPGA Fabrics with FABulous and OpenLane 2](https://dl.acm.org/doi/10.1145/3622781.3674189) | ACM Computing Frontiers | 2024 | FABulous + OpenLane 2 integration |
| [OpenROAD Flow Scripts Tutorial](https://openroad-flow-scripts.readthedocs.io/en/latest/tutorials/FlowTutorial.html) | OpenROAD Docs | 2025 | Comprehensive flow tutorial |
| [GDSFactory Documentation](https://gdsfactory.github.io/gdsfactory/) | GDSFactory | 2025 | v9.31.0; KLayout backend; multi-domain support |
| [GDS Factory: Build Better Hardware with Better Software](https://ieeetv.ieee.org/repp/gds-factory-build-better-hardware-with-better-software) | IEEE REPP | 2024 | IEEE presentation on GDSFactory workflow |

### 15.6 Neuromorphic Computing Hardware

| Paper | Authors/Source | Year | Key Contribution |
|-------|---------------|------|------------------|
| [Neuromorphic Hardware and Computing 2024](https://www.nature.com/collections/jaidjgeceb) | Nature Collection | 2024 | Cross-journal collection of neuromorphic research |
| [The road to commercial success for neuromorphic technologies](https://www.nature.com/articles/s41467-025-57352-1) | Nature Communications | 2025 | Commercial viability analysis; co-processor paradigm |
| [Roadmap to Neuromorphic Computing with Emerging Technologies](https://arxiv.org/html/2407.02353v1) | arXiv | 2024 | Comprehensive technology roadmap |
| [Exploring Neuromorphic Computing Based on SNNs: Algorithms to Hardware](https://dl.acm.org/doi/full/10.1145/3571155) | ACM Computing Surveys | 2023 | Survey from algorithms to hardware implementation |
| [Enabling Efficient Processing of SNNs with On-Chip Learning](https://arxiv.org/abs/2504.00957) | arXiv | 2025 | Edge AI on commodity neuromorphic processors |
| [Neuromorphic Computing 2025: Current SotA](https://humanunsupervised.com/papers/neuromorphic_landscape.html) | Human Unsupervised | 2025 | State-of-the-art landscape analysis |

### 15.7 Phase-Change Memory (PCM) for Analog AI

| Paper | Authors/Source | Year | Key Contribution |
|-------|---------------|------|------------------|
| [Rapid learning with PCM-based IMC through learning-to-learn](https://www.nature.com/articles/s41467-025-56345-4) | Nature Communications | 2025 | Meta-learning on neuromorphic hardware |
| [Deep neural network inference with 64-core PCM-based IMC chip](https://research.ibm.com/publications/deep-neural-network-inference-with-a-64-core-in-memory-compute-chip-based-on-phase-change-memory--2) | IBM Research, CIMTEC | 2024 | 64-core AIMC chip; 14nm + backend PCM |
| [Heterogeneous Embedded NPUs with PCM-based AIMC](https://research.ibm.com/publications/heterogeneous-embedded-neural-processing-units-utilizing-pcm-based-analog-in-memory-computing) | IBM Research, IEDM | 2024 | Heterogeneous digital/analog edge AI architecture |
| [The Role of PCM in Edge Computing and AIMC](https://www.mdpi.com/1424-8220/25/12/3618) | MDPI Sensors | 2025 | Review of PCM for smart sensing and NN acceleration |
| [An analog-AI chip for energy-efficient speech recognition](https://www.nature.com/articles/s41586-023-06337-5) | Nature | 2023 | IBM NorthPole prototype; 14× better performance/watt |

### 15.8 Chiplet and UCIe Standards

| Paper/Resource | Source | Year | Key Contribution |
|----------------|--------|------|------------------|
| [UCIe: Standard for an Open Chiplet Ecosystem](https://ieeexplore.ieee.org/document/10669138/) | IEEE Micro | 2025 | Comprehensive UCIe overview and future directions |
| [High-performance 3D SiP designs with UCIe](https://www.nature.com/articles/s41928-024-01126-y) | Nature Electronics | 2024 | 3D UCIe integration approaching monolithic performance |
| [UCIe 3.0 Specification Overview](https://www.uciexpress.org/specifications) | UCIe Consortium | 2025 | 48/64 GT/s support; architectural updates |
| [Zero ASIC UCIe Open-Source Simulation](https://www.zeroasic.com/blog/ucie-open-source-design) | Zero ASIC | 2024 | Verilator + Xyce + Switchboard simulation stack |
| [Building An Open Chiplet Economy](https://www.opencompute.org/blog/building-an-open-chiplet-economy) | Open Compute Project | 2024 | ODSA project; 50+ companies collaboration |

### 15.9 RISC-V Open-Source Silicon

| Paper/Resource | Source | Year | Key Contribution |
|----------------|--------|------|------------------|
| [Survey of Verification of RISC-V Processors](https://link.springer.com/article/10.1007/s10836-025-06169-3) | Journal of Electronic Testing | 2025 | Comprehensive verification methodology survey |
| [Basilisk: End-to-End Open-Source Linux-Capable RISC-V SoC](https://arxiv.org/html/2406.15107v1) | arXiv (ETH Zurich) | 2024 | First end-to-end open-source Linux SoC in IHP 130nm |
| [GHAZI: Open-Source ASIC Implementation of RISC-V SoC](https://www.researchgate.net/publication/366569126) | TechRxiv | 2022 | Complete open-source digital tooling methodology |
| [An Academic RISC-V Silicon Implementation](https://ieeexplore.ieee.org/document/9268664/) | IEEE | 2020 | Academic tapeout methodology using open components |
| [Review of 2024 and aims for 2025 - Zero to ASIC](https://www.zerotoasiccourse.com/post/year_update_2024/) | Zero to ASIC | 2025 | 13 RISC-V CPUs taped out; Linux on Tiny Tapeout |

### 15.10 Verilog-A and Compact Models

| Resource | Source | Key Capability |
|----------|--------|----------------|
| [OpenVAF: Next Generation Verilog-A Compiler](https://openvaf.semimod.de/) | SemiMod | 10× faster compilation; OSDI 0.4 API |
| [OpenVAF GitHub Repository](https://github.com/pascalkuthe/OpenVAF) | GitHub | Source code; pre-compiled binaries |
| [VA-Models: Verilog-A Simulation Models](https://github.com/dwarning/VA-Models) | GitHub | Collection of CMC and other compact models |
| [Ngspice OSDI/OpenVAF Documentation](https://ngspice.sourceforge.io/osdi.html) | ngspice | Integration guide; supported models |
| [Free software support for compact modelling with Verilog-A](https://ojs.midem-drustvo.si/index.php/InfMIDEM/article/view/1999) | Informacije MIDEM | Academic analysis of open Verilog-A ecosystem |

### 15.11 Neural Network Accelerator Design

| Paper/Tool | Source | Year | Key Contribution |
|------------|--------|------|------------------|
| [ADNA: Automating ASIC Development of NN Accelerators](https://www.mdpi.com/2079-9292/14/7/1432) | MDPI Electronics | 2025 | Automated accelerator generation methodology |
| [IBM AIHWKIT](https://aihwkit.readthedocs.io/en/latest/) | IBM | 2024 | Hardware-aware training toolkit for analog AI |
| [NVIDIA Deep Learning Accelerator (NVDLA)](https://nvdla.org/) | NVIDIA | Open | Standard DNN inference accelerator architecture |
| [DnnWeaver v2.0](http://dnnweaver.org/) | Academic | Open | Caffe-to-Verilog FPGA acceleration |
| [Neural-Networks-on-Silicon Collection](https://github.com/fengbintu/Neural-Networks-on-Silicon) | GitHub | 2024 | Curated paper collection on NN accelerators |

---

## 16. Recommended Reading Path

For researchers entering the FeCIM space, the following reading sequence is recommended:

### Phase 1: Foundations
1. Start with the [NeuroSim Validation paper](https://www.frontiersin.org/journals/artificial-intelligence/articles/10.3389/frai.2021.659060/full) to understand CIM benchmarking
2. Review [OpenLane paper](https://woset-workshop.github.io/PDFs/2020/a21.pdf) for RTL-to-GDSII fundamentals
3. Study the [RRAM Chemical Reviews paper](https://pubs.acs.org/doi/10.1021/acs.chemrev.4c00845) for device physics

### Phase 2: FeFET Specialization
1. Read the [University of Oulu FeFET thesis](https://oulurepo.oulu.fi/bitstream/10024/57025/1/nbnfioulu-202506164534.pdf) for Verilog-A modeling
2. Study [Sub-3nm FeFET FinFET paper](https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aisy.202400370) for advanced node design
3. Explore [CINM (Cinnamon)](https://arxiv.org/abs/2301.07486) for compiler infrastructure

### Phase 3: System Integration
1. Review [Memory Is All You Need](https://arxiv.org/html/2406.08413v1) for LLM/transformer CIM
2. Study [IBM PCM AIMC papers](https://research.ibm.com/publications/heterogeneous-embedded-neural-processing-units-utilizing-pcm-based-analog-in-memory-computing) for production architecture
3. Explore [UCIe IEEE Micro paper](https://ieeexplore.ieee.org/document/10669138/) for chiplet integration

### Phase 4: Fabrication Preparation
1. Review [Basilisk SoC paper](https://arxiv.org/html/2406.15107v1) for IHP tapeout methodology
2. Study [GDSFactory documentation](https://gdsfactory.github.io/gdsfactory/) for layout generation
3. Follow [Zero to ASIC 2024 review](https://www.zerotoasiccourse.com/post/year_update_2024/) for practical tapeout guidance

---

## 17. Extended Paper Collection (January 2026 Supplement)

This section provides additional papers organized by specialized topics to support advanced research in FeCIM and related technologies.

### 17.1 Ferroelectric HfO₂/ZrO₂ Materials Science

| Paper | Source | Year | Key Contribution |
|-------|--------|------|------------------|
| [Ferroelectric Hafnium Oxide: A Potential Game-Changer for Nanoelectronics](https://advanced.onlinelibrary.wiley.com/doi/10.1002/aelm.202400686) | Advanced Electronic Materials | 2025 | CMOS-compatible HfO₂ ferroelectrics; reliability analysis |
| [Ferroelectric materials, devices, and chips for advanced computing](https://link.springer.com/article/10.1007/s11432-025-4432-x) | Science China Information Sciences | 2025 | Doping strategies; stress engineering; oxygen vacancy control |
| [HfO₂-based FeFETs: From materials to applications](https://pubs.aip.org/aip/jap/article/138/1/010701/3351745) | Journal of Applied Physics | 2025 | Comprehensive review; 1-5V switching; ns-scale speed |
| [Progress in computational understanding of ferroelectric mechanisms in HfO₂](https://www.nature.com/articles/s41524-024-01352-0) | npj Computational Materials | 2024 | HZO physics; Zr doping mechanisms; lower crystallization temps |
| [HfO₂-based ferroelectric thin film and memory device applications](https://pmc.ncbi.nlm.nih.gov/articles/PMC11197553/) | PMC | 2024 | Post-Moore era review; HZO superlattice >10¹² cycle endurance |
| [Emerging Opportunities for FeFET: Integration of 2D Materials](https://advanced.onlinelibrary.wiley.com/doi/10.1002/adfm.202310438) | Advanced Functional Materials | 2024 | 2D channel FeFETs; flexible electronics |
| [Enhanced memory window of α-IGZO FeFET with HfO₂ interfacial layer](https://link.springer.com/article/10.1007/s11432-024-4429-7) | Science China | 2025 | 1.1V memory window; 10⁷ cycle endurance |

### 17.2 Emerging Memory Comparison (MRAM/STT-RAM/SOT-RAM/ReRAM)

| Paper | Source | Year | Key Contribution |
|-------|--------|------|------------------|
| [Progress of emerging non-volatile memory technologies in industry](https://link.springer.com/article/10.1557/s43579-024-00660-2) | MRS Communications | 2024 | Industry status of MRAM, ReRAM, PCM, FeRAM |
| [In-memory computing with emerging memory devices: Status and outlook](https://pubs.aip.org/aip/aml/article/1/1/010902/2878744) | APL Machine Learning | 2023 | Comprehensive IMC device comparison |
| [TSMC MRAM Breakthrough](https://eu.36kr.com/en/p/3513986660637571) | 36Kr | 2024 | SOT-MRAM 1ns switching; separated read/write paths |
| [The Next New Memories](https://semiengineering.com/the-next-new-memories/) | Semiconductor Engineering | 2024 | MRAM vs ReRAM vs FeFET analysis |
| [What's the Difference Between Emerging Memory Technologies?](https://www.electronicdesign.com/technologies/embedded/article/21806070) | Electronic Design | 2024 | Technology comparison overview |

### 17.3 ADC/DAC Design for CIM Readout Circuits

| Paper | Source | Year | Key Contribution |
|-------|--------|------|------------------|
| [28-nm 9T SRAM CIM macro with redundant array-assisted ADC](https://www.sciencedirect.com/science/article/abs/pii/S1879239124001012) | Integration | 2024 | 4-bit ADC using redundant columns |
| [Current-Mode SAR ADC for Memristor Readout in 28nm CMOS](https://juser.fz-juelich.de/record/1039472/files/ADC_MEMRISYS_2024_V2.pdf) | MEMRISYS | 2024 | No TIA needed; 6-bit @ 100 MSps; <3mW |
| [44.3 TOPS/W SRAM CIM with DAC/ADC-Less Operations](https://www.researchgate.net/publication/381689074) | ResearchGate | 2024 | Near-CIM analog memory eliminates converters |
| [AFPR-CIM: Analog Floating-Point RRAM CIM with Dynamic Range FP-ADC](https://arxiv.org/html/2402.13798v1) | arXiv | 2024 | Dynamic range adaptive quantization |
| [Readout Circuit Design for RRAM Array-Based CIM](https://www.mdpi.com/2079-9292/13/13/2478) | MDPI Electronics | 2024 | 8-bit SAR ADC; 47.26 fJ/Conv in 110nm |
| [HCiM: ADC-Less Hybrid Analog-Digital CIM Accelerator](https://dl.acm.org/doi/10.1145/3658617.3697572) | ASP-DAC | 2025 | Up to 28× energy reduction vs 7-bit ADC |
| [Review of SRAM-based Compute-in-Memory Circuits](https://arxiv.org/html/2411.06079v2) | arXiv | 2024 | ADC strategies; sparsity-aware power reduction |

### 17.4 IR-Drop Compensation and Non-Ideality Mitigation

| Paper | Source | Year | Key Contribution |
|-------|--------|------|------------------|
| [Optimizing hardware-software co-design for memristor crossbars](https://link.springer.com/article/10.1007/s11432-024-4240-x) | Science China | 2025 | FAST simulator; CAFM scheme; 54%+ accuracy recovery |
| [Current Opinions on Memristor-Accelerated ML Hardware](https://arxiv.org/html/2501.12644v1) | arXiv | 2025 | 16Mb RRAM macro; 31.2 TFLOPS/W |
| [Hardware implementation of memristor-based ANNs](https://www.nature.com/articles/s41467-024-45670-9) | Nature Communications | 2024 | Comprehensive protocol for memristive ANN design |
| [Compensation architecture utilizing residual resource for RRAM CIM](https://www.sciencedirect.com/science/article/abs/pii/S0026269224001010) | Microelectronics Journal | 2024 | Architecture-level accuracy compensation |
| [Mitigate IR-Drop by Modulating Neuron Activation Functions](https://www.researchgate.net/publication/371567734) | ResearchGate | 2023 | Activation function modulation for IR-drop |

### 17.5 Hardware-Aware Quantization and Training

| Paper | Source | Year | Key Contribution |
|-------|--------|------|------------------|
| [Full-stack memristor-based CIM system with SW/HW co-development](https://www.nature.com/articles/s41467-025-57183-0) | Nature Communications | 2025 | Full-stack design; bit-slicing aware training |
| [Hardware-Aware Quantization for Accurate Memristor NNs](https://dl.acm.org/doi/10.1145/3676536.3698023) | ICCAD | 2024 | Bit-precision tuning for conductance variation |
| [Memristor-based adaptive ADC for CIM](https://www.nature.com/articles/s41467-025-65233-w) | Nature Communications | 2025 | Adaptive quantization; 15.1× energy efficiency |
| [Model quantization for computing-in-memory: a survey](https://link.springer.com/article/10.1007/s11432-024-4522-8) | Science China | 2025 | Comprehensive CIM quantization survey |
| [CNN Implementation with Binary Activation and Weight Quantization](https://pubs.acs.org/doi/10.1021/acsami.3c13775) | ACS Applied Materials | 2024 | 32×32 memristor crossbar; overshoot suppression |

### 17.6 Transformer/Attention Mechanism on Analog CIM

| Paper | Source | Year | Key Contribution |
|-------|--------|------|------------------|
| [Analog IMC attention mechanism for fast LLMs](https://www.nature.com/articles/s43588-025-00854-1) | Nature Computational Science | 2025 | 70,000× energy reduction; 100× speed-up vs GPU |
| [Analog and Digital Hybrid Attention Accelerator](https://arxiv.org/abs/2409.04940) | arXiv/IEEE | 2024 | 75% token pruning; 14.8 TOPS/W; 65nm CMOS |
| [Efficient memristor accelerator for transformer self-attention](https://www.nature.com/articles/s41598-024-75021-z) | Scientific Reports | 2024 | Memristor crossbar for MatMul bottleneck |
| [HARDSEA: Hybrid Analog-ReRAM Digital-SRAM for Dynamic Sparse Self-Attention](https://ieeexplore.ieee.org/document/10719540/) | IEEE TVLSI | 2024 | Sparse attention acceleration |

### 17.7 Spiking Neural Networks on Memristor Hardware

| Paper | Source | Year | Key Contribution |
|-------|--------|------|------------------|
| [Memristor-Based Spiking Neuromorphic Systems for Brain-Inspired Computing](https://www.mdpi.com/2079-4991/15/14/1130) | MDPI Nanomaterials | 2025 | TSM sub-pJ spiking; <30ns latency; >10¹⁰ neurons/cm² |
| [Fully memristive SNN for energy-efficient graph learning](https://pmc.ncbi.nlm.nih.gov/articles/PMC12057669/) | Science | 2025 | 1.93 pJ/op @ 180nm; 37.3× lower power |
| [Memristor-based SNNs: Cooperative development](https://www.sciencedirect.com/science/article/pii/S270947232400011X) | ScienceDirect | 2024 | SNN algorithm-hardware co-design |
| [Analog Implementation of Spiking Neuron with Memristive Synapses](https://www.mdpi.com/2227-7390/12/13/2025) | Mathematics | 2024 | Analog neuromorphic prototyping |
| [Memristor-Based ANNs for Hardware Neuromorphic Computing](https://spj.science.org/doi/10.34133/research.0758) | Research | 2024 | RRAM synapses + TSM neurons |

### 17.8 3D Monolithic Integration for CIM

| Paper | Source | Year | Key Contribution |
|-------|--------|------|------------------|
| [M3D-LIME: Monolithic 3D integration of RRAM hybrid memory](https://www.nature.com/articles/s41467-023-42981-1) | Nature Communications | 2023 | 3-layer chip; 96% accuracy; 18.3× efficiency vs GPU |
| [Eq-CIM: Monolithic 3D IGZO-RRAM-SRAM architecture](https://link.springer.com/article/10.1007/s11432-024-4078-1) | Science China | 2025 | RRAM between Metal 5/6; IGZO on Metal 9 |
| [M3D-MP4: Multi-Layer CNT-CMOS/RRAM Mixed-Precision CIM](https://ieeexplore.ieee.org/iel8/10872985/10872987/10873492.pdf) | IEEE | 2025 | 4 functional layers; ≤300°C process |
| [3D Stacked IGZO 2T0C DRAM for CIM](https://www.science.org/doi/10.1126/sciadv.adu4323) | Science Advances | 2025 | 8×8 array; 3-bit storage; >100s retention |
| [Monolithic 3D integration for energy-efficient computing](https://www.sciencedirect.com/science/article/abs/pii/S1359028624000652) | ScienceDirect | 2024 | M3D review; CNT, 2D materials, oxide semiconductors |

### 17.9 TCAD Simulation for FeFET

| Paper/Resource | Source | Year | Key Contribution |
|----------------|--------|------|------------------|
| [Ferroelectric-Based Electrostatic Doping for TFET](https://pmc.ncbi.nlm.nih.gov/articles/PMC10051887/) | PMC | 2023 | Sentaurus TCAD; Lombardi mobility model |
| [TCAD numerical modeling of NC ferroelectric devices](https://www.sciencedirect.com/science/article/abs/pii/S0038110122001137) | Solid-State Electronics | 2022 | HZO in Sentaurus; radiation detection |
| [TCAD modeling of Ferroelectric Materials (IWORID 2024)](https://indico.global/event/8935/contributions/85672/) | INFN/IWORID | 2024 | Enhanced electronic device modeling |
| [Tunneling FET Calibration: Sentaurus vs. Silvaco](https://www.researchgate.net/publication/347929897) | ResearchGate | 2020 | TCAD tool comparison methodology |
| [FeFET Simulation Discussions](https://www.researchgate.net/post/How_to_perform_FeFET_and_NCFET_simulation_in_sentaurus_tcad) | ResearchGate | 2024-2025 | Community resources; FERRO model application |

### 17.10 Optical/Photonic Compute-in-Memory

| Paper | Source | Year | Key Contribution |
|-------|--------|------|------------------|
| [MIT Photonic Processor for Ultrafast AI](https://news.mit.edu/2024/photonic-processor-could-enable-ultrafast-ai-computations-1202) | MIT News / Nature Photonics | 2024 | <0.5ns computation; 92%+ accuracy |
| [MAFT-ONN: Photonic Processor for 6G](https://news.mit.edu/2025/photonic-processor-could-streamline-6g-wireless-signal-processing-0611) | MIT News | 2025 | 100× faster than digital; 10,000 neurons/chip |
| [Lightmatter Photonic AI Processor](https://lightmatter.co/blog/a-new-kind-of-computer/) | Lightmatter | 2025 | ResNet, BERT, RL without modifications |
| [IEEE/HP Labs Silicon Photonics Platform](https://ieeephotonics.org/announcements/2025ieee-study-leverages-silicon-photonics-for-scalable-and-sustainable-ai-hardwareapril-3-2025/) | IEEE JSTQE | 2025 | Wafer-scale integration; on-chip lasers/amplifiers |
| [Photonics for Neuromorphic Computing: Fundamentals and Devices](https://advanced.onlinelibrary.wiley.com/doi/10.1002/adma.202312825) | Advanced Materials | 2025 | Sub-ns latency; low heat dissipation |
| [Integrated Neuromorphic Photonic Computing for AI Acceleration](https://advanced.onlinelibrary.wiley.com/doi/10.1002/adma.202508029) | Advanced Materials | 2025 | Emerging network architectures |
| [Optical computing accelerators: Principle and perspective](https://journal.hep.com.cn/fop/EN/10.15302/frontphys.2025.032302) | Frontiers of Physics | 2025 | Comprehensive optical accelerator review |

### 17.11 Electrochemical RAM (ECRAM) for Analog AI

| Paper | Source | Year | Key Contribution |
|-------|--------|------|------------------|
| [Prospects and challenges of ECRAM for deep-learning accelerators](https://www.sciencedirect.com/science/article/pii/S1359028624000536) | ScienceDirect | 2024 | Symmetric states; low variability; low energy |
| [ECRAM: Recent advances in materials, devices, and systems](https://nanoconvergencejournal.springeropen.com/articles/10.1186/s40580-024-00415-8) | Nano Convergence | 2024 | 1st/2nd gen ECRAM; solid electrolytes |
| [Open-loop analog programmable ECRAM array](https://www.nature.com/articles/s41467-023-41958-4) | Nature Communications | 2023 | Programmable analog arrays |
| [POSTECH ECRAM for AI - Science Advances](https://www.sciencedaily.com/releases/2024/08/240801121936.htm) | ScienceDaily | 2024 | Largest ECRAM array; commercialization potential |
| [Multi-Bit ECRAM Analog Neuromorphic System - 97.3% Accuracy](https://pubmed.ncbi.nlm.nih.gov/39312419/) | PubMed | 2024 | High-precision current readout |
| [Computing With Chemicals Makes Faster AI](https://spectrum.ieee.org/analog-ai-ecram-artificial-synapse) | IEEE Spectrum | 2024 | ECRAM as artificial synapse |

### 17.12 Device Reliability: Endurance and Retention

| Paper | Source | Year | Key Contribution |
|-------|--------|------|------------------|
| [HfO₂-based FeFETs: Materials and Reliability](https://pubs.aip.org/aip/jap/article/138/1/010701/3351745) | Journal of Applied Physics | 2025 | 10⁴-10⁵ cycle endurance challenges |
| [Progress of emerging NVM technologies in industry](https://pmc.ncbi.nlm.nih.gov/articles/PMC11618178/) | PMC | 2024 | HZO 10¹² cycles @ VLSI 2024; STT-MRAM 10¹⁵ cycles |
| [Advances of embedded RRAM in industrial manufacturing](https://iopscience.iop.org/article/10.1088/2631-7990/ad2fea) | IJEM | 2024 | RRAM >10¹² cycles; >10 year retention |
| [FeFET Advancements and Challenges for Next-Gen NVM](https://www.sciencedirect.com/science/article/abs/pii/S2352492823002817) | ScienceDirect | 2023 | Gate-stack degradation analysis |
| [Enhancing RRAM Reliability with Al Doping](http://ieeexplore.ieee.org/iel8/7298/11157712/11039729.pdf) | IEEE | 2025 | Al-doped HfO₂; improved stability @ 125°C |
| [Atomic-Scale Insights into RRAM Switching Mechanisms](https://arxiv.org/pdf/2509.16512) | arXiv | 2025 | Fundamental switching physics |

---

## 18. Quick Reference: Key GitHub Repositories

| Repository | Description | URL |
|------------|-------------|-----|
| **NeuroSim** | CIM accelerator benchmark | https://github.com/neurosim/DNN_NeuroSim_V2.1 |
| **CiMLoop** | System-level CIM modeling | https://github.com/mit-emze/cimloop |
| **CINM (Cinnamon)** | MLIR-based CIM compiler | https://github.com/tud-ccc/Cinnamon |
| **OpenVAF** | Verilog-A compiler | https://github.com/pascalkuthe/OpenVAF |
| **VA-Models** | Compact model collection | https://github.com/dwarning/VA-Models |
| **GDSFactory** | Python chip layout | https://github.com/gdsfactory/gdsfactory |
| **OpenROAD** | RTL-to-GDSII flow | https://github.com/The-OpenROAD-Project/OpenROAD |
| **OpenLane** | ASIC flow controller | https://github.com/The-OpenROAD-Project/OpenLane |
| **IHP Open PDK** | 130nm BiCMOS PDK | https://github.com/IHP-GmbH/IHP-Open-PDK |
| **SkyWater PDK** | 130nm CMOS PDK | https://github.com/google/skywater-pdk |
| **SRAM CIM Literatures** | Paper collection | https://github.com/BUAA-CI-LAB/Literatures-on-SRAM-based-CIM |
| **Neural-Networks-on-Silicon** | Accelerator papers | https://github.com/fengbintu/Neural-Networks-on-Silicon |
| **IBM AIHWKIT** | Analog AI toolkit | https://github.com/IBM/aihwkit |

---

## 19. Conference and Journal Quick Reference

### Top Venues for CIM/FeCIM Research

| Venue | Type | Focus Area |
|-------|------|------------|
| **ISSCC** | Conference | Circuits; CIM chip demonstrations |
| **VLSI** | Symposium | Device + circuits; memory technology |
| **IEDM** | Conference | Devices; FeFET/RRAM fundamentals |
| **DAC/ICCAD** | Conference | EDA; accelerator architecture |
| **ISCA/MICRO** | Conference | Computer architecture; systems |
| **Nature Electronics** | Journal | High-impact device/system papers |
| **Nature Communications** | Journal | Broad scope; CIM demonstrations |
| **IEEE JSSC** | Journal | Circuit implementations |
| **IEEE TVLSI** | Journal | VLSI systems and architectures |
| **Advanced Materials** | Journal | Materials science; device physics |

### Key Research Groups

| Institution | Focus | Notable Work |
|-------------|-------|--------------|
| **Georgia Tech** | NeuroSim; CIM benchmarking | Prof. Shimeng Yu |
| **MIT** | CiMLoop; photonic computing | Prof. Murmann, Prof. Englund |
| **IBM Research** | PCM AIMC; AIHWKIT | Zurich/Almaden labs |
| **ETH Zurich** | Open-source RISC-V; Basilisk | PULP team |
| **Stanford** | Memristor arrays | Prof. Wong |
| **POSTECH** | ECRAM; analog AI | Neuromorphic team |
| **Intel** | Loihi neuromorphic | Intel Labs |
| **FZ Jülich** | Analog IMC attention | Leroux team (Nature Comp Sci 2025) |
| **TU Dresden/FMC** | Ferroelectric DRAM+ | €100M funding (2025) |
| **external research institution** | In₂Se₃ FeFET synthesis | Tour Lab |

---

## 20. 2025 Breakthrough Papers (Critical Updates)

*This section highlights the most impactful recent publications that significantly advance the field.*

### 20.1 Analog IMC Attention for LLMs (Major Breakthrough)

**Reference:** Leroux et al., "Analog in-memory computing attention mechanism for fast and energy-efficient large language models," *Nature Computational Science*, September 2025. [View](https://www.nature.com/articles/s43588-025-00854-1)

**Key Results:**
- **70,000× energy reduction** compared to GPUs for attention computation
- **100× speed-up** for attention operations
- Successfully trained **1.5 billion parameter** model
- Uses gain-cell memories (CMOS-compatible)
- Sliding window attention bounds physical memory size

**Implications for FeCIM:**
This paper validates that analog in-memory computing can accelerate the most compute-intensive operations in LLMs. The techniques are directly applicable to FeFET-based crossbars.

### 20.2 FeCIM Annealer for Combinatorial Optimization

**Reference:** Qian Y, et al., "Device-Algorithm Co-Design of Ferroelectric Compute-in-Memory In-Situ Annealer," *DAC 2025*. [View](https://dl.acm.org/doi/10.1109/DAC63849.2025.11133307)

**Key Results:**
- **1503-1716× energy reduction** vs. state-of-the-art annealers
- **98% success rate** for 3000-node Max-Cut problems
- Vector-matrix-vector (VMV) multiplication acceleration
- 75% chip size saving via lossless QUBO matrix compression

### 20.3 HfO₂-ZrO₂ Superlattice Stability

**Reference:** "Enhancing ferroelectric stability: wide-range of adaptive control in epitaxial HfO₂/ZrO₂ superlattices," *Nature Communications 2025*. [View](https://www.nature.com/articles/s41467-025-61758-2)

**Key Results:**
- Stable ferroelectricity from ultra-thin to **100nm** thickness
- Fatigue resistance exceeding **10⁹ switching cycles**
- Low coercive field of **~0.85 MV/cm**
- Validates long-term reliability of HfO₂-ZrO₂ for CIM

### 20.4 In₂Se₃ Gram-Scale Synthesis

**Reference:** Shin, Tour et al., "In₂Se₃ Synthesized by the FWF Method for Neuromorphic Computing," *Advanced Electronic Materials 2025*. [View](https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aelm.202400603)

**Key Results:**
- Flash-within-flash Joule heating enables **gram-scale** α-In₂Se₃ synthesis
- Robust synaptic behavior demonstrated
- Path to scalable 2D ferroelectric neuromorphic devices

### 20.5 Industry Momentum: FMC €100M Funding

**Reference:** Bloomberg, November 2025. [View](https://www.bloomberg.com/news/articles/2025-11-13/memory-chip-startup-raises-100-million-for-energy-saving-tech)

**Key Points:**
- **Ferroelectric Memory Company (FMC)** raised €100M
- Dresden, Germany (TU Dresden/GlobalFoundries spinout)
- DRAM+ technology based on ferroelectric HfO₂
- Targets energy-efficient memory for AI applications
- Technology compatible with modern CMOS processes

---

## 21. Document Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2024-01 | Initial creation |
| 2.0 | 2025-01 | Added 2025-2026 updates; post-Efabless era |
| 2.1 | 2026-01 | Added breakthrough papers section; FMC funding; analog attention results |

---

*This document is maintained as part of the FeCIM Lattice Tools project. All referenced papers and tools remain the intellectual property of their respective authors and institutions. This is educational documentation, not an endorsement of any technology or company.*

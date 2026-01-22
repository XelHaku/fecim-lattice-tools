The Open-Source EDA Ecosystem: A Comprehensive Analysis for CMOS and Emerging Memory Technologies
1. Introduction
The semiconductor industry stands at a pivotal juncture characterized by the decoupling of design capabilities from proprietary, capital-intensive infrastructure. Historically, the design and fabrication of Integrated Circuits (ICs) were the exclusive domain of large integrated device manufacturers (IDMs) and well-funded fabless design houses, protected by high barriers to entry including expensive Electronic Design Automation (EDA) licenses and restrictive Non-Disclosure Agreements (NDAs). This paradigm has been disrupted by the Free and Open Source Silicon (FOSSi) movement, which advocates for the democratization of chip design through open-source tools, open Process Design Kits (PDKs), and accessible shuttle programs.
For researchers and engineers working at the bleeding edge of computing architecture—specifically Compute-in-Memory (CIM) and Ferroelectric Field-Effect Transistor (FeFET) based systems—this open ecosystem presents both unprecedented opportunities and significant technical gaps. Standard Digital Logic flows (RTL-to-GDSII) have matured to production readiness for standard CMOS, yet the tools required for emerging non-volatile memory technologies remain fragmented. This report provides an exhaustive analysis of the current open-source EDA landscape, evaluating its readiness for CMOS digital design and identifying the specific tooling deficits for FeCIM architectures. It synthesizes data from academic research, software repositories, and industry reports to construct a roadmap for integrating custom educational visualizers with professional-grade open-source silicon flows.
2. Open Source EDA Tools Landscape
The open-source EDA ecosystem is bifurcated into two primary flows: the digital design flow, which is highly automated and relies on abstraction, and the analog/mixed-signal flow, which requires manual intervention and transistor-level simulation.
2.1 Digital Design Flow (RTL to GDSII)
The digital design flow automates the transformation of a high-level behavioral description (Register Transfer Level or RTL) into a physical geometric representation (GDSII) suitable for photolithography. This process involves a chain of complex transformations, each handled by specialized tools.
2.1.1 Logic Synthesis: Yosys
Yosys serves as the foundational front-end for the vast majority of open-source digital flows. It acts as the bridge between human-readable hardware description languages (Verilog) and the technological primitives available in a semiconductor process.
Functionality: Yosys parses Verilog code and converts it into an internal Abstract Syntax Tree (AST). It performs technology-independent optimizations (such as constant propagation and dead code elimination) before mapping the logic to the standard cell library provided by the PDK. This mapping process uses the ABC algorithm (developed at UC Berkeley) to optimize for area, speed, or power.
Maturity & Adoption: Yosys is production-ready and widely adopted in both academia and industry. It is the default synthesis engine for the OpenROAD flow, OpenLane, and several FPGA toolchains (like SymbiFlow). Its robustness is evidenced by its ability to handle complex designs, including RISC-V cores like the Ibex and PicoRV32.
Mechanism: It operates by generating a BLIF (Berkeley Logic Interchange Format) or Verilog netlist that instantiates specific gates (e.g., sky130_fd_sc_hd__nand2_1) rather than generic boolean operators.
2.1.2 Physical Implementation: OpenROAD
OpenROAD represents a paradigm shift in physical design, driven by the DARPA IDEA program's goal of achieving a "no-human-in-the-loop" hardware compiler capable of generating a GDSII from RTL in under 24 hours.
Functionality: OpenROAD is not merely a script but a monolithic application that integrates multiple engines into a shared database (OpenDB). It handles:
Floorplanning: Defining the die area and placing I/O pins (using ioPlacer).
Global Placement: Approximating the location of instances to minimize wirelength (using RePlAce).
Detailed Placement: Snapping cells to manufacturing grids and ensuring no overlaps (using OpenDP).
Clock Tree Synthesis (CTS): Inserting buffers to distribute the clock signal evenly (using TritonCTS).
Routing: Connecting the pins with metal wires, first globally (FastRoute) and then in detail (TritonRoute).
Architecture: By unifying these tools into a single C++ application with Tcl and Python bindings, OpenROAD eliminates the file parsing overhead and fragility associated with older, script-based flows. It provides real-time access to the design database, allowing for incremental optimization.
2.1.3 The Flow Controller: OpenLane
While OpenROAD performs the heavy lifting of physical design, OpenLane serves as the overarching "manager" or "flow controller" that orchestrates the entire process from RTL to GDSII.
Functionality: OpenLane automates the execution of Yosys, OpenROAD, Magic, KLayout, and verification tools. It manages hundreds of configuration variables (e.g., core utilization, clock period, metal layer limits) to tailor the flow for specific designs and PDKs.
Evolution:
OpenLane v1: The current stable standard, widely used for the Google/Efabless OpenMPW shuttles. It is Tcl-heavy and tightly coupled to the Efabless tape-out requirements.
OpenLane v2 / LibreLane: A complete rewrite in Python, designed for greater modularity and flexibility. It treats flow steps as objects, allowing users to inject custom Python scripts or swap tools easily—a feature critical for integrating non-standard CIM macros.
Maturity: OpenLane is tape-out proven, having facilitated hundreds of successful designs on the SkyWater 130nm process.
2.2 Analog and Mixed-Signal Design Flow
Unlike the digital flow, the analog flow prioritizes simulation accuracy and manual layout control, essential for defining the sensitive voltage-current relationships in FeFETs and crossbar arrays.
2.2.1 Schematic Capture: Xschem
Xschem has emerged as the preferred schematic editor for open-source analog design due to its performance and modern feature set.
Functionality: It allows designers to draw hierarchical circuits using symbols representing PDK devices. Crucially, it generates netlists in multiple formats (SPICE, Verilog, VHDL) suitable for simulation or Layout-Versus-Schematic (LVS) checks.
Capability: Xschem can handle extremely large netlists that would crash older tools like XCircuit. It supports "back-annotation," where simulation results (voltages, operating points) are overlaid directly onto the schematic, aiding in debugging.
Integration: It is tightly integrated with the SkyWater 130nm PDK, with pre-configured symbol libraries that match the foundry's device models.
2.2.2 Circuit Simulation: ngspice
ngspice is the open-source equivalent of industry-standard SPICE simulators (like Spectre or HSPICE). It is the engine that calculates the physics of the circuit.
Functionality: It solves systems of non-linear differential equations to predict circuit behavior in the time and frequency domains. It supports standard transistor models (BSIM3, BSIM4, BSIM-CMG) required for modern PDKs.
Emerging Tech Support: A critical recent development is the integration of the OSDI (Open Source Device Interface) and the OpenVAF compiler. This allows ngspice to load compiled Verilog-A models at runtime. Since FeFETs are often modeled using Verilog-A (due to their complex hysteresis physics), this feature makes ngspice the only viable open-source tool for simulating ferroelectric devices.
Maturity: Highly mature, with decades of development. It is the default simulator for Xschem and KiCad.
2.2.3 Layout Design: Magic VLSI and KLayout
Physical layout in the open world relies on two complementary tools: Magic and KLayout.
Magic VLSI: A "corner-stitching" layout editor that provides real-time Design Rule Checking (DRC). It is unique in that it understands connectivity and abstract layers, making it excellent for drawing cells while ensuring they are manufacturable. It is the sign-off tool for SkyWater 130nm DRC.
KLayout: A polygon-based viewer and editor that excels at handling massive GDSII files (full chips). It is scriptable via Python and Ruby, making it a powerful platform for procedurally generating structures like memory crossbars or running complex DRC/LVS scripts.
Usage Pattern: Designers typically draw individual cells (like a FeFET bitcell) in Magic to ensure local DRC compliance, then use KLayout to assemble the full array or view the final GDSII generated by OpenROAD.
2.3 Verification Tools
Verification ensures that the design is physically manufacturable (DRC) and electrically correct (LVS and Timing).
Design Rule Checking (DRC): Magic and KLayout both perform DRC. Magic does it interactively; KLayout runs batch scripts. These checks ensure metal lines aren't too close, transistors have proper implants, and densities meet foundry specs.
Parasitic Extraction (PEX): Magic and OpenRCX (part of OpenROAD) extract the resistance and capacitance of the drawn wires. This "parasitic" information is fed back into simulation to predict realistic delays.
Layout Versus Schematic (LVS): Netgen is the standard open-source LVS tool. It compares the SPICE netlist extracted from the layout against the source schematic netlist. If they match, the layout is a faithful representation of the circuit.
Timing Analysis: OpenSTA performs Static Timing Analysis (STA), checking setup and hold times across all corners (Process, Voltage, Temperature) to ensure the chip will run at the desired frequency.
3. Key Open Source Projects Comparison
The following table synthesizes the key attributes of the major open-source projects relevant to CMOS and CIM design.
Project
Primary Function
License
Language
Maturity
CIM Relevance
OpenROAD
Digital P&R (Physical Design)
BSD-3
C++, Tcl, Python
Production
Essential for placing digital control logic around CIM macros.
Magic VLSI
Layout Editor & DRC
Berkeley
C, Tcl
Production
Critical for drawing custom FeFET bitcells and checking DRC.
ngspice
Analog Simulator
BSD/GPL
C
Production
The only simulator capable of running Verilog-A FeFET models via OSDI.
Xschem
Schematic Capture
GPL
C, Tcl
Production
Used to design the analog interface circuits (ADCs/DACs) for CIM.
KLayout
Layout Viewer/Editor
GPL
C++, Python
Production
Best for procedurally generating large crossbar arrays via scripts.
Yosys
RTL Synthesis
ISC
C++
Production
Synthesizes the digital controller logic for the CIM accelerator.
OpenLane
Flow Controller
Apache 2.0
Python, Tcl
Production
Automates the integration of digital blocks; requires custom config for CIM.
SKY130 PDK
130nm CMOS PDK
Apache 2.0
-
Production
The standard process. Requires custom modeling for FeFETs.
GF180MCU
180nm CMOS PDK
Apache 2.0
-
Production
High voltage options (5V/6V) are excellent for FeFET programming.
IHP SG13G2
130nm BiCMOS PDK
Apache 2.0
-
Beta/Prod
Includes RRAM (Memristor) support, closest to native CIM.

Insight: While the digital stack (OpenROAD/OpenLane) is highly integrated, the analog/memory stack relies on loose coupling between Xschem, ngspice, and Magic. For CIM design, this means the user must act as the "integrator," manually ensuring that the simulation models in ngspice match the physical layout drawn in Magic.
4. The Gap: CIM/FeCIM-Specific Tools and Solutions
A significant "tooling gap" exists between standard CMOS EDA (optimized for boolean logic) and the requirements of Compute-in-Memory (optimized for analog accumulation). Standard digital tools assume signals are either 0 or 1; CIM relies on current summation (\Sigma I = \Sigma G \cdot V), which is inherently analog.
4.1 Missing Capabilities in Open Source EDA
Crossbar Compilers: There is no open-source equivalent to a "Memory Compiler" for CIM. Standard memory compilers generate SRAM blocks. A CIM compiler needs to generate a crossbar array, row drivers (DACs), and column readouts (ADCs/SAs) automatically. Currently, this must be done manually or via custom Python scripts in KLayout.
FeFET Standard Cells: Open PDKs (SKY130/GF180) do not contain "FeFET" cells. The foundry provides standard NMOS/PMOS. To make a FeFET, a designer must either (a) use a specific layer combination that the fab identifies as ferroelectric (if available), or (b) design a standard transistor and assume post-processing.
Large-Scale Array Simulation: SPICE (ngspice) scales poorly with matrix size (O(N^2) or worse). Simulating a 256x256 crossbar with transient physics models for every FeFET is computationally prohibitive for full-chip verification.
4.2 Academic Solutions (The Bridge)
To bridge this gap, the research community has developed high-level modeling tools that estimate CIM performance without full circuit simulation.
4.2.1 NeuroSim (Georgia Tech)
Methodology: Circuit-level macro modeling. It uses analytical equations and look-up tables calibrated against SPICE data to estimate area, latency, and energy.
Validation: NeuroSim has been validated against actual silicon data (e.g., TSMC 40nm RRAM macros), showing error margins within 5-10% for key metrics.
Relevance: It is the standard for benchmarking. However, it outputs metrics (spreadsheets), not designs (GDSII). It helps architect the CIM macro but does not build it.
4.2.2 CiMLoop (MIT)
Methodology: System-level statistical modeling. Unlike NeuroSim's circuit focus, CiMLoop models the data flow through the entire memory hierarchy, capturing the interaction between workload (e.g., ResNet-50) and hardware.
Advantage: It uses a statistical energy model that is orders of magnitude faster than cycle-accurate simulation, allowing for rapid design space exploration (e.g., optimizing ADC resolution vs. array size).
Status: Active development, open-source (GitHub), and Python-based.
4.2.3 PUMA (Purdue/HP)
Methodology: ISA and Compiler. PUMA defines an instruction set for memristor accelerators and provides a compiler to map neural networks onto this ISA.
Relevance: It addresses the programmability gap. While NeuroSim models the hardware, PUMA models how software interacts with that hardware.
4.2.4 FeFET Specific Modeling
Since native PDK support is absent, FeFETs are modeled using Verilog-A.
Models: The most common are the Preisach model (hysteresis based on domains) and the Landau-Khalatnikov (LK) model (physics-based thermodynamics).
Workflow: These models are compiled using OpenVAF into an object file (.osdi) and loaded into ngspice. This enables physics-accurate simulation of a single cell or small array.
5. PDK (Process Design Kit) Basics
5.1 What is a PDK?
A Process Design Kit (PDK) is the interface between the designer and the foundry. It contains the files necessary to ensure a design can be manufactured and will work as simulated.
Device Models (.lib, .spice): Mathematical descriptions of transistors for simulation.
Layout Rules (.tech, DRC decks): Definitions of minimum widths, spacings, and enclosures for each layer (Metal 1, Poly, Diffusion).
Technology Files (.lef): Abstract representations of layers used by automated routers.
P-Cells (Parameterized Cells): Scripts (usually in Python or Tcl) that automatically draw a device (e.g., a transistor with width W) in the layout editor.
5.2 Open Source PDK Analysis for FeCIM
5.2.1 SkyWater SKY130 (130nm)
Type: 130nm CMOS with 5 metal layers.
FeFET Potential: While it has no native FeFET, it includes SONOS (Silicon-Oxide-Nitride-Oxide-Silicon) Flash memory primitives. These operate on similar principles (charge trapping vs. polarization) and utilize high-voltage functionality. The High-Voltage (HV) modules (up to 20V) are critical for generating the write pulses needed to switch ferroelectric domains.
Availability: Fully open source on GitHub.
5.2.2 GlobalFoundries GF180MCU (180nm)
Type: 180nm legacy node, optimized for MCUs and PMIC (Power Management).
FeFET Potential: It features robust High-Voltage (HV) transistors (5V, 6V, 10V) natively. This makes it an excellent candidate for the peripheral driving circuits of a FeFET array, which often require voltages higher than the 1.8V standard logic.
Availability: Open source, actively supported by Google.
5.2.3 IHP SG13G2 (130nm BiCMOS)
Type: High-performance BiCMOS (Bipolar + CMOS).
FeFET Potential: This is the only open PDK with an experimental RRAM (Resistive RAM) module explicitly exposed. While RRAM is memristive and not ferroelectric, the circuit topologies (crossbars, sense amps) are nearly identical. It serves as the closest reference implementation for a CIM array in the open ecosystem.
Conclusion: For a FeFET project, one must use the CMOS transistors from these PDKs for the peripheral logic (drivers/muxes) and instantiate a "black box" or custom Verilog-A model for the FeFET device itself.
6. Learning Path: From Zero to Chip Designer
The following curriculum leverages the "Zero to ASIC" methodology, structured to guide a beginner from basic concepts to a complex CIM project.
Phase 1: Foundations (Weeks 1-2)
Objectives: Understand the transistor, the gate, and the toolchain.
Curriculum:
Theory: "Digital Integrated Circuits" (Rabaey) - Chapters on MOS physics and CMOS inverters.
Tools: Install Docker. Pull the IIC-OSIC-TOOLS image, which contains the entire suite (OpenROAD, OpenLane, Xschem, ngspice, Magic) pre-installed and configured.
Tutorials: Matt Venn’s "Zero to ASIC" course (highly recommended intro) and the "Open Source Silicon" YouTube channel.
Phase 2: The First "Tapeout" (Weeks 3-4)
Objectives: Complete the full RTL-to-GDSII cycle.
Project: A simple Digital Counter.
Workflow:
Design: Use Wokwi (browser-based simulation) to verify logic.
Synthesis: Write the Verilog. Use OpenLane to synthesize it to the SKY130 library.
Physical: Let OpenLane run OpenROAD to place and route the design.
Submission: Submit to Tiny Tapeout. This platform aggregates designs onto a single chip, reducing costs to ~$100. This provides the psychological win of "taping out".
Phase 3: Intermediate - Analog & Layout (Weeks 5-8)
Objectives: Break out of the digital abstraction; touch the physics.
Project: A 1T-1FeFET Memory Cell Simulation.
Workflow:
Simulation: Use Xschem and ngspice. Download a FeFET Verilog-A model (e.g., Scalable-FeFET). Compile it with openvaf. Simulate the hysteresis loop (Voltage vs. Polarization).
Layout: Use Magic VLSI to draw the physical layout of the cell. Learn to pass DRC (design rules).
LVS: Use Netgen to verify that your Magic layout matches your Xschem schematic.
7. How Real Chips Get Made: The Flow
The transformation from idea to silicon follows a rigid sequence. Open-source tools now cover every step of this process, though with varying degrees of automation for analog designs.
Step
Description
Open Source Tool
Commercial Equivalent
Capability
1. Specification
Defining architecture, PPA targets.
CiMLoop (Python)
SystemC, Excel
High
2. Design Entry
Writing Verilog or drawing schematics.
Yosys (Digital), Xschem (Analog)
Design Compiler, Virtuoso
High
3. Verification
Ensuring logic/physics are correct.
Verilator, ngspice
VCS, Spectre, MMSIM
High
4. Synthesis
Mapping RTL to logic gates.
Yosys
Genus, Design Compiler
High
5. Place & Route
Placing gates and wiring them.
OpenROAD
Innovus, ICC2
High (Digital)
6. Layout (Custom)
Drawing custom cells (memory/analog).
Magic, KLayout
Virtuoso Layout
High (Manual)
7. Sign-off
DRC (Rules) and LVS (Connectivity).
Magic, Netgen, KLayout
Calibre, PVS
Medium
8. Tapeout
File aggregation and mask generation.
Tiny Tapeout, OpenLane
Foundry Portal
Medium
9. Fabrication
Physical manufacturing.
None (Requires Fab)
TSMC, GlobalFoundries
N/A
10. Testing
Validating the physical chip.
PyTest, Custom PCBs
Advantest ATE
Medium

Insight: Steps 6 and 7 are the bottleneck for FeCIM. OpenROAD cannot automatically route a custom analog crossbar. This requires scripting in KLayout (Python) to procedurally generate the array, which is then treated as a "Macro" (black box) by OpenROAD for the rest of the digital integration.
8. Integration Strategy: Connecting Demo 6 to Real EDA
Your project involves visualizers and crossbar compilers. Integrating these with the open EDA stack transforms them from educational toys into functional design automation tools.
Strategy 1: Python to SPICE (Simulation Export)
Your "Demo 6" visualizer can act as a schematic capture tool.
Mechanism: Use the spicelib or PySpice Python libraries.
Workflow:
User configures crossbar parameters (size, conductance states) in your GUI.
Your tool generates a .spice netlist file. This file instantiates the SKY130 transistor models for drivers and your custom .osdi model for the FeFETs.
Your tool invokes ngspice in batch mode (or shared library mode) to run a transient simulation.
Parse the raw output file to visualize the inference accuracy or current summing behavior.
Strategy 2: Python to GDSII (Layout Generation)
Your "Crossbar Compiler" can directly generate manufacturable layouts.
Mechanism: Use the KLayout Python API or gdsfactory.
Workflow:
Define the geometric rules (from the PDK LEF file) in Python (e.g., poly_width = 0.15um).
Script a nested loop to draw the FeFET array: place the active layer, gate layer, and orthogonal metal1 (bitlines) and metal2 (wordlines).
Export a .gds file. This GDS is now a "Hard Macro" that can be instantiated in OpenLane.
Strategy 3: Architecture Exploration Integration
Mechanism: Export configurations for CiMLoop.
Workflow: Your tool can export a YAML configuration file describing the array architecture. The user then runs CiMLoop to get high-level energy/area estimates, validating the architectural choices before circuit-level design begins.
9. Companies & Communities
The ecosystem is sustained by a mix of commercial entities and non-profit foundations.
Active Organizations
Efabless: The platform operator for the OpenMPW program. They provide the marketplace and the cloud infrastructure that runs OpenLane. They are the primary hub for open silicon fabrication.
FOSSi Foundation: The advocacy group behind the movement. They organize ORConf and maintain the LibreCores repository.
ChipFlow: A startup focused on simplifying the flow using Python-based HDLs (Amaranth). They aim to make chip design as accessible as PCB design, targeting OEMs rather than just chip architects.
Zero ASIC: Focused on education and reducing the barrier to entry, closely linked to the "Zero to ASIC" course.
Communities
Tiny Tapeout Discord: The most active community for beginners. It features channels dedicated to tools (Wokwi, OpenLane) and specific PDKs.
SkyWater PDK Slack: Technical discussion involving the PDK maintainers and analog designers.
OpenROAD GitHub Discussions: The place for deep technical issues regarding the P&R flow.
10. Realistic Assessment
What Can Be Done Today?
Digital Logic: One can design, verify, and manufacture a digital RISC-V processor or a standard neural network accelerator (systolic array) using purely open-source tools (OpenLane + SKY130) with high confidence. The flow is production-ready.
Analog Simulation: Detailed simulation of FeFET arrays is possible using ngspice + OpenVAF, provided the user has a valid Verilog-A model.
What Cannot Be Done Today (The Gap)?
Push-Button FeCIM: There is no "OpenROAD for Analog." You cannot click a button and get a routed FeFET crossbar. This layout must be drawn by hand or by custom scripts.
Native FeFET Fabrication: The open PDKs (SKY130/GF180) are CMOS-only. "Taping out" a FeFET design today effectively means taping out the CMOS control circuitry with empty slots (or standard transistors) where the FeFETs would go, or establishing a private partnership with a research fab (like Fraunhofer or extensive post-processing).
The Strategic Opportunity
Your project—a "FeCIM Visualizer and Compiler"—fills the most critical void. By automating the generation of SPICE netlists (for verification) and GDSII layouts (for implementation) from a high-level GUI, you provide the missing link that transforms raw open-source tools into a domain-specific FeCIM design suite.
11. Resource List
Tools & Code:
ngspice: ngspice.sourceforge.io
KLayout: klayout.de
CiMLoop: github.com/mit-emze/cimloop
OpenROAD: theopenroadproject.org
SkyWater PDK: github.com/google/skywater-pdk
Education:
Zero to ASIC Course: zerotoasiccourse.com
Tiny Tapeout: tinytapeout.com
Key Figures:
Matt Venn: Educator, founder of Zero to ASIC and Tiny Tapeout.
Tim Edwards: Maintainer of Magic VLSI and Open_PDKs; the architect of the open analog flow.
Stefan Schippers: Creator of Xschem.
Mohamed Shalan: Principal architect of OpenLane.
This ecosystem, while young, provides all the necessary primitives. The challenge—and the opportunity—lies in integration. By connecting your educational tools to these robust backends, you bridge the gap between abstract research and physical realization.
Works cited
1. OpenLANE: The Open-Source Digital ASIC Implementation Flow - woset-workshop, https://woset-workshop.github.io/PDFs/2020/a21.pdf 2. Empowering innovation: OpenROAD and the future of open-source EDA - EE World Online, https://www.eeworldonline.com/empowering-innovation-openroad-and-the-future-of-open-source-eda/ 3. OpenROAD Flow Scripts Tutorial, https://openroad-flow-scripts.readthedocs.io/en/latest/tutorials/FlowTutorial.html 4. OpenROAD – Key Milestones on the Road towards Good PPA, https://theopenroadproject.org/openroad-key-milestones-on-the-road-towards-good-ppa/ 5. Magic Mai/OpenROAD - Gitee, https://gitee.com/magic3007/OpenROAD 6. librelane/librelane: ASIC implementation flow infrastructure, successor to OpenLane - GitHub, https://github.com/librelane/librelane 7. LibreLane Documentation, https://librelane.readthedocs.io/ 8. Efabless Recommended Open Source Analog Design Flow, http://www.opencircuitdesign.com/analog_flow/ 9. Installing and “tuning up” xschem, http://web02.gonzaga.edu/faculty/talarico/vlsi/xschemInstall.html 10. XSCHEM SKY130 INTEGRATION, https://xschem.sourceforge.io/stefan/xschem_man/tutorial_xschem_sky130.html 11. Ngspice, the open source Spice circuit simulator - Intro, https://ngspice.sourceforge.io/ 12. Ngspice circuit simulator - Verilog-A with OSDI/OpenVAF - SourceForge, https://ngspice.sourceforge.io/osdi.html 13. Magic VLSI - Open Circuit Design, http://opencircuitdesign.com/magic/ 14. Hi all I ve been playing with magic and klayout for a while open-source-silicon.dev #sky130, https://web.open-source-silicon.dev/t/16600034/hi-all-i-ve-been-playing-with-magic-and-klayout-for-a-while- 15. Best Custom Layout Tools to Master for VLSI Job Opportunities, https://vlsiguru.com/blog/best-tools-to-learn-in-custom-layout-course 16. NeuroSim Simulator for Compute-in-Memory Hardware Accelerator: Validation and Benchmark - PMC - NIH, https://pmc.ncbi.nlm.nih.gov/articles/PMC8219932/ 17. NeuroSim Simulator for Compute-in-Memory Hardware Accelerator: Validation and Benchmark - Frontiers, https://www.frontiersin.org/journals/artificial-intelligence/articles/10.3389/frai.2021.659060/full 18. CiMLoop: A Flexible, Accurate, and Fast Compute-In-Memory Modeling Tool - IEEE Xplore, https://ieeexplore.ieee.org/document/10590023/ 19. CiMLoop: A Flexible, Accurate, and Fast Compute-In-Memory Modeling Tool - arXiv, https://arxiv.org/pdf/2405.07259 20. PUMA: A Programmable Ultra-efficient Memristor-based Accelerator for Machine Learning Inference - ResearchGate, https://www.researchgate.net/publication/330725909_PUMA_A_Programmable_Ultra-efficient_Memristor-based_Accelerator_for_Machine_Learning_Inference 21. Embedding-Enhanced Probabilistic Modeling of Ferroelectric Field Effect Transistors (FeFETs) - TRACE: Tennessee Research and Creative Exchange, https://trace.tennessee.edu/cgi/viewcontent.cgi?article=15607&context=utk_gradthes 22. DavidTobar456/pfecapRevision: Verilog-A Preisach ferroelectric cap (PFECAP) simulation model for FET - GitHub, https://github.com/DavidTobar456/pfecapRevision 23. SkyWater SKY130 Process Design Rules, https://skywater-pdk.readthedocs.io/en/main/rules.html 24. Specifications For 180MCU - GF180MCU PDK - Read the Docs, https://gf180mcu-pdk.readthedocs.io/en/latest/analog/spice/elec_specs/elec_specs.html 25. SiGe:C-BiCMOS-Technologies - IHP Microelectronics, https://www.ihp-microelectronics.com/services/research-and-prototyping-service/mpw-prototyping-service/sigec-bicmos-technologies 26. Introduction to Chip Design, http://www2.imm.dtu.dk/~masca/chip-design-book.pdf 27. Zero to ASIC Course, https://www.zerotoasiccourse.com/ 28. Digital Design Guide - Tiny Tapeout, https://tinytapeout.com/digital_design/ 29. 4 - Submit your design :: Quicker, easier and cheaper to make your own chip! - Tiny Tapeout, https://tinytapeout.com/guides/workshop/submit-your-design/ 30. spicelib - PyPI, https://pypi.org/project/spicelib/0.9.1/ 31. 1. Overview — PySpice @VERSION@ documentation - Fabrice Salvaire, https://pyspice.fabrice-salvaire.fr/releases/v1.3/overview.html 32. Efabless Launches chipIgnite with SkyWater to Bring Chip Creation to the Masses, https://www.skywatertechnology.com/efabless-launches-chipignite-with-skywater-to-bring-chip-creation-to-the-masses/ 33. ChipFoundry — ChipFlow, https://www.chipflow.io/chipfoundry 34. ChipFlow - Helping product companies to make their own chips, https://www.chipflow.io/ 35. ZTM Community Discord. Learn Together, Grow Together. | Zero To Mastery, https://zerotomastery.io/community/developer-community-discord/ 36. Tiny Tapeout :: Quicker, easier and cheaper to make your own chip!, https://tinytapeout.com/ 37. OpenLane vs OpenROAD #900 - GitHub, https://github.com/The-OpenROAD-Project/OpenLane/discussions/900
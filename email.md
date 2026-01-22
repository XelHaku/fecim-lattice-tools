TO: tour@rice.edu

CC:
  jaeho-shin@rice.edu
  tawfik.jarjour@accenture.com


Subject: Open-source FeCIM design tool - matches your 87% MNIST, exports to ngspice/GDSII

Dr. Tour,

After watching your COSM presentation on ferroelectric compute-in-memory, I built a visualization and design suite for FeCIM arrays.

The project includes six modules:

1. **Hysteresis** - P-E curves using the Preisach model, 30 discrete states (~4.9 bits/cell)
2. **Crossbar MVM** - Matrix-vector multiply with toggleable non-idealities (IR drop, sneak paths, drift)
3. **MNIST Inference** - Dual-mode FP32 vs CIM. Configured to match your reported 87% hardware accuracy
4. **Peripheral Circuits** - DAC/ADC/TIA for the analog interface
5. **Technology Comparison** - Energy metrics, competitive analysis, market sizing
6. **FeCIM Design Suite** - Crossbar compiler, layout generator, SPICE netlist export, GDSII output

Demo 6 is the core. It addresses a gap I found in the open-source EDA ecosystem: there's no "OpenROAD for Analog" - you can't click a button and get a routed FeFET crossbar. This tool:

- Compiles neural network weights → physical cell conductances and programming voltages
- Generates SPICE netlists for ngspice simulation (with OpenVAF FeFET models)
- Exports GDSII layouts compatible with KLayout and GDSFactory
- Bridges visualization to real open-source silicon tools

The MNIST demo validates the simulation against your published 87% result. Users can toggle failure modes to understand why that number is impressive given quantization and device variation.

**To be clear:** I don't have access to real hardware data. The physics models are based on published literature and your public presentations. This is a work in progress that would need validation against actual device measurements before any real-world application. I'm building the framework - the accuracy depends on calibration with real data.

I've kept attribution clear and avoided any branding that could create IP concerns.

I also appreciate your work on faith and science - it's part of why I paid attention to your COSM talk in the first place.

If there's something specific that would help IronLattice - investor demos, design exploration, or integration with your actual device models - I'd rather build that than guess.

GitHub: github.com/XelHaku/multilayer-ferroelectric-cim-visualizer
Demo video: [link]

FeCIM Maintainers
Monterrey, Mexico
+52 812 193 7470 · WhatsApp · Telegram

github.com/XelHaku
trebuchetdynamics.com
TO: tour@rice.edu

CC:
  jaeho-shin@rice.edu
  tawfik.jarjour@accenture.com


Subject: Educational FeCIM visualization tool - learning by exploring

Dr. Tour,

After watching your COSM presentation - "the same device does the memory and the computation" - I built an educational visualization tool to help myself (and others) understand FeCIM technology. It's exploratory, not production-ready, but it makes the concepts tangible.

**What it does:**

1. **Hysteresis** - P-E curves with Preisach model, visualize 30 discrete states
   → Helps explain "why multi-level matters" through interaction

2. **Crossbar MVM** - Matrix-vector multiply with toggleable non-idealities
   → Passive and 1T1R architectures, IR drop, sneak paths, drift - see the problems visually

3. **MNIST Demo** - Draw a digit, watch the crossbar process it
   → Makes compute-in-memory intuitive, even for non-technical audiences

4. **Peripheral Circuits** - DAC/ADC/TIA visualization
   → Passive and 1T1R modes, shows Read/Write/Compute flow

5. **Technology Comparison** - Energy/performance comparisons from literature
   → Educational overview of where FeCIM sits vs other technologies

6. **EDA Explorer** - Experimental Verilog/DEF generation
   → Validates with OpenLane, generates visualizations via KLayout, OpenROAD, Yosys
   → Learning tool for understanding crossbar layout, not a production design flow

7. **Documentation** - Physics explanations and references
   → Collected from published papers and presentations

**What it's NOT:** This isn't production software. The physics models are simplified, the accuracy numbers need real device calibration, and the EDA output is exploratory. It's a learning framework.

**What it might be useful for:** Helping people understand FeCIM concepts through hands-on exploration. Sometimes drawing a digit and watching the crossbar respond teaches more than reading a paper.

I also appreciate your work on faith and science - it's part of why I paid attention to your COSM talk in the first place.

**Why private:** The repo is private because I know you have patents pending. I've done deep research and added some personal ideas that might actually be relevant to your technology. My goal is to help IronLattice, not create problems. If there's a specific feature or configuration you'd find useful, I'd rather ask than guess.

I want to be part of what IronLattice is building. This software is my contribution - happy to adapt it however serves you best.

Demo video: [2-minute walkthrough]
GitHub: Private repo - happy to grant access or share a demo build

FeCIM Maintainers
Monterrey, Mexico
+52 812 193 7470 (WhatsApp/Telegram)

github.com/XelHaku
trebuchetdynamics.com

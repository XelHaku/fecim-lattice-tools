TO: tour@rice.edu

CC:
  jaeho-shin@rice.edu
  tawfik.jarjour@accenture.com


Subject: Interactive FeCIM visualization suite - investor demos that let people draw digits and watch the crossbar compute

Dr. Tour,

After watching your COSM presentation - "the same device does the memory and the computation" - I built an interactive visualization suite for FeCIM technology. Five modules designed for technical briefinges and foundry conversations.

**What's working now:**

1. **Hysteresis** - P-E curves with Preisach model, 30 discrete states (~4.9 bits/cell)
   → Explains "why 30 levels matters" in 60 seconds

2. **Crossbar MVM** - Matrix-vector multiply with toggleable non-idealities
   → IR drop, sneak paths, drift - visualize the problems and how FeCIM handles them

3. **MNIST Demo** - Draw a digit, watch the crossbar recognize it, 87% accuracy
   → Configured to match your reported hardware result. The "wow moment" for investor meetings.

4. **Peripheral Circuits** - DAC/ADC/TIA in Write/Read/Compute modes
   → Shows this is a real system, not just a memory cell

5. **Technology Comparison** - Energy metrics, competitive matrix, market sizing
   → FeCIM wins every category, investor-ready charts

**In development:**
- Module 6: FeCIM Design Suite - crossbar compiler, SPICE netlist export, GDSII output
  → Working toward "OpenROAD for Analog" - click a button, get a routed FeFET crossbar

**The value:** Interactive demos > PowerPoint slides.
When an investor draws a "7" and watches your crossbar recognize it in real-time, that's worth more than 50 slides explaining the technology.

**To be clear:** Built from published literature and your public presentations - no proprietary data. The physics models would need calibration against real device measurements before any serious application. I'm building the framework; the accuracy depends on real data.

I also appreciate your work on faith and science - it's part of why I paid attention to your COSM talk in the first place.

If this could help IronLattice with investor demos, foundry discussions, or design exploration, I'd rather build what you actually need than guess.

GitHub: github.com/XelHaku/multilayer-ferroelectric-cim-visualizer
Demo video: [2-minute walkthrough]

FeCIM Maintainers
Monterrey, Mexico
+52 812 193 7470 (WhatsApp/Telegram)

github.com/XelHaku
trebuchetdynamics.com

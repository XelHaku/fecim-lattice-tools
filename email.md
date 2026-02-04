TO: tour@rice.edu
CC: jaeho-shin@rice.edu, tawfik.jarjour@accenture.com

Subject: FeCIM simulator - 7 modules, OpenLane-style artifacts (educational)

Dr. Tour,

After your COSM presentation, I built an interactive simulator
for exploring FeCIM concepts. Seven modules:

- Hysteresis - Preisach (quasi-static) + Landau-Khalatnikov (dynamic) engines, ISPP write-verify demo with target tracking
- Crossbar - 1T1R vs Passive toggle, IR drop, sneak paths visible
- MNIST - Draw a digit, compare FP vs CIM behavior
- Peripherals - DAC/TIA/ADC, Write/Read/Compute modes, 1T1R vs Passive
- Comparison - Model-based scenario comparisons (simulation-only)
- EDA Suite - Verilog/DEF/SPICE/JSON/CSV artifacts, OpenLane-style flow
- Docs - Built-in documentation browser

Inspired by published literature. The L-K engine is calibrated/normalized to match per-material Ec/Pr for reachability, but the tool is still educational (not device-validated).
Would need real device measurement data for true calibration.

Your faith-and-science work is part of why I paid attention.

Screenshot: [attached]
Demo: [5-minute video]
Repo: Private. Will invite on reply.

FeCIM Maintainers
Monterrey, Mexico
+52 812 193 7470
https://github.com/XelHaku
https://linkedin.com/in/juan-tamez-a3266661

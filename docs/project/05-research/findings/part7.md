# Ferroelectric CIM Research Findings - Part 7 (January 2026)

Research groups, patent landscape, testing methodologies, and commercialization timeline.

---

## 1. International FeFET Research Groups

### Academic Leaders

| Institution | Lab/Group | PI | Focus Area |
|-------------|-----------|----|-----------|
| **Georgia Tech** | Lab for Emerging Devices & Circuits | Prof. Shimeng Yu | FeFET reliability, CIM |
| **UC Berkeley** | Salahuddin Group | Prof. Sayeef Salahuddin | NC-FET, FerroX |
| **external research institution** | Tour Lab | Prof. external research group | Ferroelectric CIM, research sample synthesis |
| **Stanford** | Wong Group | Prof. H.-S. Philip Wong | Memory architecttic |
| **MIT** | del Alamo Group | Prof. Jesús del Alamo | III-V FeFET |
| **Purdue** | Datta Group | Prof. Sumeet Gupta | CIM reliability |

### European Research Centers

| Institution | Country | Focus |
|-------------|---------|-------|
| **NaMLab** | Germany | HfO₂ ferroelectrics |
| **IMEC** | Belgium | CMOS integration |
| **Fraunhofer** | Germany | Manufacturing |
| **CEA-Leti** | France | 22nm FD-SOI |
| **TU Dresden** | Germany | Characterization |

### Asian Research Centers

| Institution | Country | Focus |
|-------------|---------|-------|
| **KAIST** | South Korea | MFM structures |
| **Seoul National** | South Korea | 2D ferroelectrics |
| **Tsinghua** | China | CIM architecttic |
| **AIST** | Japan | FeRAM |
| **TSMC Research** | Taiwan | BEOL integration |

### Georgia Tech Highlights (2024-2025)

**Key Publications at IEEE IEDM 2024:**
- "Amorphous indium oxide channel FeFETs with 0.9V write voltage and >10¹² endurance"
- "Self-healing ferroelectric capacitors with ~1000× endurance improvement at 85-125°C"
- Comprehensive TDDB characterization at IRPS 2024

**Source:** [Shimeng Yu Lab](https://shimeng.ece.gatech.edu/publication/)

### IMEC-Georgia Tech Collaboration

**Innovation:** Non-destructive read mechanism using small-signal C-V asymmetry (sense high/low capacitance).

**Quote:** "This is a new mechanism to operate ferro using small-signal C-V asymmetry for a non-destructive read." —Prof. Shimeng Yu

**Source:** [IMEC](https://www.imec-int.com/en/articles/imec-demonstrates-breakthrough-in-cmos-compatible-ferroelectric-memory)

---

## 2. Patent Landscape

### Leading IP Holders

| Company | Patent Position | Focus Areas |
|---------|-----------------|-------------|
| **IBM** | Leading | CIM, neuromorphic |
| **Samsung** | Leading | 3D NAND, mobile |
| **Intel** | Strong | Embedded NVM |
| **Micron** | Strong | Memory arrays |
| **SK Hynix** | Growing | DRAM replacement |
| **GlobalFoundries** | Growing | 22nm FD-SOI |
| **TSMC** | Growing | BEOL integration |

### Emerging Patent Players

| Company | Type | Innovation |
|---------|------|------------|
| **FMC (Dresden)** | Startup | Logic-compatible FeFET |
| **TetraMem** | Startup | Analog CIM |
| **ICLeague** | Startup | AI memory |
| **Applied Materials** | Equipment | Deposition tools |

### Patent Trends

| Technology | Trend | Status |
|------------|-------|--------|
| FRAM/FeFET | **Upward** | Growing interest |
| Neuromorphic CIM | **Rapid growth** | Hot area |
| 3D integration | Growing | Emerging |
| Embedded NVM | Stable | Mature filings |

**Source:** [Knowmade Patent Report](https://www.knowmade.com/patent-analytics-services/patent-report/semiconductor-patent-landscape/semiconductor-memory-patent-landscape/memory-for-artificial-intelligence-patent-landscape-analysis-2023/)

### Samsung Breakthrough (2025)

**Title:** "96% Lower-Power NAND Design Based on Ferroelectric Transistors"

| Feature | Value |
|---------|-------|
| Power reduction | **96%** |
| Architecture | FeFET-based 3D NAND |
| Ferroelectric | Hafnia-based |
| Channel | Oxide semiconductor |

**Source:** [Tom's Hardware](https://www.tomshardware.com/tech-industry/semiconductors/samsung-researchers-publish-96percent-lower-power-nand-design-based-on-ferroelectric-transistors)

---

## 3. Testing and Characterization Methods

### Key Reliability Metrics

| Metric | Definition | Target |
|--------|------------|--------|
| **Endurance** | Write cycles before failure | >10⁶ (consumer), >10¹² (CIM) |
| **Retention** | Data storage duration | >10 years @ 85°C |
| **Memory Window** | VTH difference (P vs E) | >0.5V |
| **Disturb Immunity** | Read/write disturb margin | No degradation |

### Endurance Testing Methodology

**Standard Protocol:**
```
1. Apply bipolar write pulses (±2V, 20ns)
2. Periodically read memory state:
   - Low VTH state (program)
   - High VTH state (erase)
3. Monitor DC Id-Vg characteristics
4. Track memory window degradation
```

**Key Finding:** Measurement delay affects apparent endurance. System setup delays longer than erase-after-program delay cause ferroelectric switching restoration.

**Source:** [IEEE Xplore](https://ieeexplore.ieee.org/document/10938969/)

### Retention Testing

**Protocol:**
1. Apply program/erase pulses
2. Monitor VTH and drain current at VG=0V
3. Test at multiple temperatures (0°C, 40°C, 85°C)
4. Extrapolate to 10-year retention

**Results:** FeFET 10-hour retention tests extrapolated to 10 years show stable properties independent of temperature.

**Source:** [IBM Research](https://research.ibm.com/publications/retention-and-endurance-of-fefet-memory-cells)

### Advanced Characterization Techniques

| Technique | Purpose | Advantage |
|-----------|---------|-----------|
| **PUND** | Polarization measurement | Minimizes leakage current |
| **Single-pulse Id-Vg** | Charge trapping kinetics | Fast characterization |
| **In-situ VTH** | Endurance monitoring | Overcomes size limits |
| **C-V measurement** | Capacitive read | Non-destructive |

### PUND (Positive-Up Negative-Down) Method

```
Pulse sequence:
  P: +V (switch + non-switch)
  U: +V (non-switch only)
  N: -V (switch + non-switch)
  D: -V (non-switch only)

Pr = (P - U) / 2 = true remanent polarization
```

**Pulse width:** 500 ns typical

### Degradation Mechanisms

| Mechanism | Cause | Detection |
|-----------|-------|-----------|
| **Interfacial layer degradation** | Oxide breakdown | Increased leakage |
| **Charge trapping** | Electron/hole injection | VTH shift |
| **Fatigue** | Domain pinning | Reduced Pr |
| **Imprint** | Preferred polarization state | Asymmetric switching |

**Key Insight:** Excess electron injection (not hole injection) plays key role in charge trapping during endurance cycling. VTH after program positively shifts due to less electron de-trapping.

**Source:** [IEEE IEDM](https://ieeexplore.ieee.org/document/10145966/)

---

## 4. Commercialization Timeline

### Historical Milestones

| Year | Milestone |
|------|-----------|
| 2011 | Ferroelectricity discovered in HfO₂ (Böscke) |
| 2017 | 22nm FD-SOI FeFET demo (GF/NaMLab) |
| 2019 | Expected qualification (delayed) |
| 2023 | First CIM array demos |
| 2024 | Pilot production begins |
| 2025 | FMC €100M funding |

### Current Status (January 2026)

| Company | Status | Technology |
|---------|--------|------------|
| GlobalFoundries | R&D, wafer sales to universities | 22nm FD-SOI |
| Samsung | Research demos | 3D FeFET NAND |
| TSMC | Research publications | BEOL OS-FeFET |
| Intel | Internal R&D | Embedded NVM |
| FMC | Pilot production | 28nm FeFET |

### Projected Timeline

| Year | Expected Milestone |
|------|-------------------|
| 2025-2026 | IoT/MCU embedded NVM |
| 2026-2027 | First commercial FeFET products |
| 2027-2028 | Edge AI accelerators |
| 2028-2030 | Commercial-grade FeFET SCM |
| 2030+ | 3D ferroelectric memory arrays |

### Challenges to Commercialization

| Challenge | Status | Solution |
|-----------|--------|----------|
| Variability | Active research | Process control |
| Endurance | Improving (10⁶→10¹²) | Material engineering |
| Cost | Higher than NAND | Volume scaling |
| Standards | Limited | Industry consortia |

**Quote:** "Next-generation ferroelectric memories such as FeFET have not been commercialized amid ongoing technical and economic challenges." —Mark LaPedus

**Source:** [What Ever Happened To Next-Gen Ferroelectric Memories?](https://marklapedus.substack.com/p/what-ever-happened-to-next-gen-ferroelectric)

---

## 5. Industry Ecosystem

### Foundry Services

| Foundry | Node | Service | Access |
|---------|------|---------|--------|
| GlobalFoundries | 22nm/28nm FD-SOI | PDK, wafers | University partnerships |
| TSMC | Research | Publications only | Limited |
| Samsung | Internal | Not available | Internal |

### Equipment Suppliers

| Company | Equipment | Application |
|---------|-----------|-------------|
| Applied Materials | ALD systems | HZO deposition |
| Lam Research | Etch tools | Pattern definition |
| Tokyo Electron | Thermal processing | Crystallization |
| ASML | Lithography | Patterning |

### EDA Tools

| Tool | Vendor | Purpose |
|------|--------|---------|
| Spectre | Cadence | Circuit simulation |
| HSPICE | Synopsys | SPICE modeling |
| Sentaurus | Synopsys | TCAD |
| Custom PDKs | Foundries | Design rules |

---

## 6. Additional Papers for Download

### Research Group Publications

| Paper | URL | Year | Group |
|-------|-----|------|-------|
| Georgia Tech IEDM 2024 | https://shimeng.ece.gatech.edu/publication/ | 2024 | Yu Lab |
| HfO₂ Game-Changer Review | https://advanced.onlinelibrary.wiley.com/doi/10.1002/aelm.202400686 | 2025 | Review |
| FeFET Materials to Systems | https://advanced.onlinelibrary.wiley.com/doi/10.1002/adma.202515480 | 2025 | Li et al. |
| HfO₂ FeFET Review | https://pubs.aip.org/aip/jap/article/138/1/010701/3351745 | 2024 | Comprehensive |

### Patent and Market Reports

| Resource | URL | Year |
|----------|-----|------|
| Memory for AI Patent Landscape | https://www.knowmade.com/patent-analytics-services/patent-report/semiconductor-patent-landscape/semiconductor-memory-patent-landscape/memory-for-artificial-intelligence-patent-landscape-analysis-2023/ | 2023 |
| FeRAM Market Status | https://marklapedus.substack.com/p/what-ever-happened-to-next-gen-ferroelectric | 2024 |
| Samsung 96% Power Reduction | https://www.tomshardware.com/tech-industry/semiconductors/samsung-researchers-publish-96percent-lower-power-nand-design-based-on-ferroelectric-transistors | 2025 |

### Testing Methodology

| Paper | URL | Year | Topic |
|-------|-----|------|-------|
| Retention and Endurance | https://ieeexplore.ieee.org/document/8739726/ | 2019 | IMW |
| Endurance Measurement Methodology | https://ieeexplore.ieee.org/document/10938969/ | 2025 | C-V read |
| Scaled FeFET Endurance | https://ieeexplore.ieee.org/document/10145966/ | 2023 | In-situ VTH |
| High Endurance Without Retention Penalty | https://www.researchgate.net/publication/326727783 | 2018 | Optimization |
| TU Dresden Characterization | https://d-nb.info/1074350103/34 | Thesis | Methods |

---

## 7. Ferroelectric CIM Positioning

### Competitive Analysis

| Aspect | Ferroelectric CIM | FMC | Others |
|--------|-------------|-----|--------|
| Structure | HZO superlattice | Standard HZO | Various |
| Funding | $50K (seed) | €100M (C-round) | Variable |
| IP | Rice patents | Growing portfolio | Diverse |
| Focus | CIM for AI | DRAM replacement | Mixed |
| Stage | Lab spinout | Pilot production | Various |

### Differentiation Strategy

1. **Superlattice advantage:** Better endurance than solid-solution
2. **Academic foundation:** Tour Lab credibility
3. **Synthesis innovation:** research sample/FWF methods
4. **CIM focus:** AI-first approach

### Recommended Actions

| Action | Priority | Timeline |
|--------|----------|----------|
| Publish benchmark data | High | Q1 2026 |
| Patent key innovations | High | Ongoing |
| Establish foundry partnership | High | Q2 2026 |
| Expand funding | High | Q2 2026 |
| Build test infrastructure | Medium | Q3 2026 |

---

## 8. Key Takeaways

### Research Landscape

- **Georgia Tech** leading academic FeFET reliability research
- **IMEC** pioneering non-destructive read mechanisms
- **Samsung** demonstrating 96% power reduction in FeFET NAND

### Patent Position

- **IBM/Samsung** hold leading IP positions
- **New entrants** (FMC, TetraMem) building portfolios
- **FRAM/FeFET** patents trending upward

### Commercialization Reality

- First commercial FeFET expected **2026-2027**
- Commercial-grade SCM expected **2028-2030**
- Key barriers: variability, endurance qualification, cost

### Testing Standards

- **PUND** for accurate Pr measurement
- **In-situ VTH** for scaled device endurance
- **10-year extrapolation** from accelerated tests
- **Charge trapping** as primary degradation mechanism

---

## References

1. Shimeng Yu Lab: https://shimeng.ece.gatech.edu/publication/
2. IMEC Ferroelectric Breakthrough: https://www.imec-int.com/en/articles/imec-demonstrates-breakthrough-in-cmos-compatible-ferroelectric-memory
3. Memory for AI Patents: https://www.knowmade.com/patent-analytics-services/patent-report/semiconductor-patent-landscape/semiconductor-memory-patent-landscape/memory-for-artificial-intelligence-patent-landscape-analysis-2023/
4. FeRAM Status: https://marklapedus.substack.com/p/what-ever-happened-to-next-gen-ferroelectric
5. FeFET Endurance Testing: https://ieeexplore.ieee.org/document/10938969/

---

*Last updated: January 2026 - Part 7*

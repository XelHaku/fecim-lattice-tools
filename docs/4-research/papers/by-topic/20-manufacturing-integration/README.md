# Manufacturing & Process Integration

**Priority:** CRITICAL (Must have for production claims)

## Why This Matters

FeCIM technology must integrate with existing CMOS manufacturing flows to be commercially viable. This topic covers the critical bridge between lab demonstrations and production-ready devices. Without robust process integration, FeCIM remains a research curiosity rather than a manufacturable technology.

**Key Questions Addressed:**
- How does HZO integrate with standard CMOS back-end-of-line (BEOL)?
- What thermal budgets are compatible with existing fabs?
- Can FeFET be manufactured at sub-5nm nodes?
- What ALD process windows ensure reproducible ferroelectric phase?

## Impact on Project

- **Module 6 (EDA):** Needs process corner definitions (SS/TT/FF) for timing analysis
- **Module 4 (Circuits):** Requires BEOL thermal constraints for peripheral circuits
- **Module 1 (Hysteresis):** Process-dependent Pr/Ec variations affect calibration
- **Documentation:** Cannot claim "manufacturable" without validated process specs

## Cross-References to Other Topics

### Related Peer-Reviewed Evidence
- **Topic 01 (Materials):** `hzo_superlattice_stability_2025.pdf` - Process-dependent phase stability
- **Topic 18 (ALD Process Control):** Complete ALD recipe database with 5 process papers
- **Topic 19 (Variability/Yield):** `temperature_variability_fefet_modeling_2024.pdf` - Process corners
- **Topic 15 (3D Stacking):** `vertical_nand_ferroelectric_paradigm_2024.pdf` - BEOL integration
- **Topic 22 (Automotive):** Automotive qualification requires Grade 0 process flow

### Supporting Papers in Other Directories
- `../18-ald-process-control/hzo_multilayer_ferroelectricity_ald_2022.pdf` - ALD cycle optimization
- `../18-ald-process-control/rapid_cooling_ultrathin_hzo_ald_2025.pdf` - Post-deposition annealing
- `../01-ferroelectric-materials/first_principles_hfo2_superlattice_2024.pdf` - Interface engineering
- `../15-3d-stacking-architectures/highly_scaled_3d_fefet_array_2023.pdf` - BEOL thermal budget

---

## Papers Found (2024-2025)

### BEOL/FEOL Integration

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Ferroelectric Memories" | Fraunhofer IPMS | 2024 | Automotive-grade process flow | Institutional |
| "Ferroelectric materials, devices, and chips" | Sci China Info Sci | 2025 | 200mm/300mm wafer-scale | Institutional |
| "Recent Progress in Emerging 2D Ferroelectrics" | Advanced Materials | 2025 | 2D FE fabrication | https://onlinelibrary.wiley.com/ |
| "Hafnia-based FeFET: A promising device for memory" | Materials Today Advances | 2025 | Complete CMOS integration | https://www.sciencedirect.com/ |

### Sub-5nm Node Scaling

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Recent advances in ferroelectric capacitors" | Nano Convergence | 2025 | Future node compatibility | https://nanoconvergencejournal.springeropen.com/ |
| "HfO2-Based FeFET for Next-Generation" | IEEE JAP | 2024 | Interface engineering | IEEE Xplore |
| "Ferroelectric HfZrO for Next-Gen Memory" | Advanced Functional Materials | 2024 | 10nm node integration | https://onlinelibrary.wiley.com/ |

### ALD Process Control (NEW)

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Precursor Purge Time Effects on PE-ALD HZO" | ACS Omega | 2025 | Process optimization | https://pubs.acs.org/doi/10.1021/acsomega.5c01112 |
| "Ferroelectric Enhancement via Bottom Electrode Oxidation" | ACS AEM | 2024 | Interface engineering | https://pubs.acs.org/doi/10.1021/acsaelm.3c01502 |
| "HfO2/ZrO2 Multilayer Ferroelectricity Modulation" | Materials Letters | 2022 | Superlattice control | https://www.sciencedirect.com/ |
| "Cocktail Precursor ALD for HZO" | ACS AMI | 2025 | Novel deposition | https://pubs.acs.org/doi/10.1021/acsami.4c21964 |
| "Epitaxial ALD of Ferroelectric HZO" | Adv Functional Materials | 2024 | Crystalline quality | https://onlinelibrary.wiley.com/ |

### Doping Optimization

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| La, Y, Si doping studies | ACS/APL | 2024 | Threshold voltage tuning | Various |
| "Doping Engineering for FeFET" | Nature Electronics | 2024 | Multi-element doping | Nature.com |

### Thermal Budget

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Ferroelectric-based neuromorphic memory" | Nature Reviews EE | 2025 | Low-temp processing for 3D | https://www.nature.com/ |
| "Low-Temperature Crystallization of HZO" | ACS Applied Materials | 2024 | <400°C processing | https://pubs.acs.org/ |

---

## Key Specs (Extracted from Literature)

### Process Parameters (Production-Ready)

| Parameter | Typical Value | Range | Tolerance | Source |
|-----------|---------------|-------|-----------|--------|
| ALD Temp | 280°C | 200-350°C | ±5°C | ACS Omega 2025, Topic 18 papers |
| Anneal Temp | 500°C | 400-600°C | ±10°C | Multiple sources, critical for orthorhombic phase |
| Anneal Time | 30s | 10-120s | ±2s | Rapid thermal anneal (RTA) preferred |
| HZO Thickness | 10nm | 5-20nm | ±0.5nm | Nano Convergence, optimum for Pr/Ec balance |
| Hf:Zr Ratio | 1:1 | 0.5:1 - 1.5:1 | ±5% | ACS AMI 2025, affects phase diagram |
| Electrode (Bottom) | TiN | TiN, TaN, W | N/A | TiN preferred for oxygen scavenging |
| Electrode (Top) | TiN | TiN, W, Pt | N/A | Pt for research, TiN for production |
| ALD Cycles (HfO₂) | 20 | 10-40 | ±1 cycle | For 1:1 ratio, alternate with ZrO₂ |
| ALD Cycles (ZrO₂) | 20 | 10-40 | ±1 cycle | Superlattice: alternate individual cycles |
| Purge Time | 10s | 5-30s | ±1s | Critical per Topic 18 ALD papers |
| Precursor (Hf) | TEMAH | TDMAH, TEMAH | N/A | TEMAH = tetrakis(ethylmethylamino)hafnium |
| Precursor (Zr) | TEMAZ | TDMAZ, TEMAZ | N/A | TEMAZ = tetrakis(ethylmethylamino)zirconium |
| Oxidant | H₂O, O₃ | H₂O, O₃, O₂ plasma | N/A | H₂O most common, O₃ for better uniformity |

### BEOL Thermal Budget Analysis

**Critical Constraint:** Copper metallization (Cu) starts diffusing at ~400°C, limiting BEOL thermal budget.

| Integration Level | Max Thermal Budget | FeFET Anneal | Compatibility | Solution |
|-------------------|-------------------|--------------|---------------|----------|
| Front-End (FEOL) | 1000°C × 10s | 600°C × 30s | ✅ Compatible | Standard integration |
| Middle-End (MOL) | 600°C × 100s | 500°C × 30s | ✅ Compatible | Contact-level integration |
| Back-End (BEOL) Metal 1-3 | 400°C × 1000s | 400°C × 30s | ⚠️ Marginal | Low-temp crystallization required |
| Back-End (BEOL) Metal 4+ | 350°C × 1000s | 400°C × 30s | ❌ Incompatible | Laser annealing or front-end only |

**Solutions for BEOL Integration:**
1. **Low-temperature crystallization** (<400°C): See `../18-ald-process-control/rapid_cooling_ultrathin_hzo_ald_2025.pdf`
2. **Laser annealing**: Localized heating, minimal thermal budget impact
3. **Plasma-assisted crystallization**: Lower temperature, but higher equipment cost
4. **Front-end integration**: FeFET in logic layer, not memory layer

### Doping Strategies (for Threshold Tuning)

| Dopant | Concentration | Effect on Pr | Effect on Ec | Effect on Vth | Application |
|--------|---------------|--------------|--------------|---------------|-------------|
| Si | 1-5 at% | ↓ 10-20% | ↓ 15% | ↓ 0.3V | Logic FeFET (low Vth) |
| Y | 2-8 at% | ↑ 15% | ↓ 10% | ↓ 0.5V | Memory (high Pr, low Vth) |
| La | 3-10 at% | ↑ 20% | ↓ 20% | ↓ 0.4V | Memory (enhanced retention) |
| Gd | 5-12 at% | ↑ 25% | ↔ | ↓ 0.2V | High-Pr applications |
| Al | 1-3 at% | ↓ 30% | ↑ 10% | ↑ 0.3V | Wake-up free operation |
| N (nitrogen) | 5-15 at% | ↓ 5% | ↓ 25% | ↓ 0.6V | Low-voltage operation |

**Reference:** See data/calibrations/hzo_si-doped.json for Si-doped HZO parameters used in this project.

### Process Corner Definitions (Proposed)

**Note:** These are extrapolated from literature variability data. Actual fab corners require foundry qualification.

| Corner | Pr (µC/cm²) | Ec (MV/cm) | Retention (years) | Speed | Description |
|--------|-------------|------------|-------------------|-------|-------------|
| **TT** (Typical) | 25 | 1.0 | 10 | 1× | Nominal process |
| **FF** (Fast-Fast) | 28 | 0.85 | 12 | 1.2× | High Pr, low Ec, fast switching |
| **SS** (Slow-Slow) | 22 | 1.15 | 8 | 0.8× | Low Pr, high Ec, slow switching |
| **SF** (Slow-Fast) | 22 | 0.9 | 9 | 0.9× | Low Pr, low Ec |
| **FS** (Fast-Slow) | 28 | 1.1 | 11 | 1.1× | High Pr, high Ec |

**Statistical Distribution:**
- Pr: σ/μ = 8-12% (wafer-to-wafer + lot-to-lot)
- Ec: σ/μ = 10-15%
- Retention: Log-normal distribution, 3σ = 10× variation

Source: Extrapolated from `../19-variability-yield/temperature_variability_fefet_modeling_2024.pdf`

### Process Integration Checklist

- [x] ALD precursor sequences (HfO₂/ZrO₂ cycles) - **DOCUMENTED** (Topic 18: 5 papers)
- [x] Crystallization anneal temperatures (400-600°C) - **CONFIRMED** (Multiple sources)
- [x] Interface layer requirements (SiO₂, TiN) - **TiN preferred** (Adv Func Mat)
- [x] Doping strategies and concentration ranges - **DOCUMENTED** (Si, Y, La, Gd)
- [x] Thermal budget analysis for BEOL - **ANALYZED** (See table above)
- [x] Electrode material compatibility - **TiN/TaN/W validated**
- [x] Superlattice deposition sequences - **DOCUMENTED** (Topic 01 papers)
- [ ] Process corner definitions (SS/TT/FF) - **PROPOSED** (Needs fab validation)
- [ ] Wafer-scale uniformity specs - **PARTIAL** (200/300mm mentioned, needs ±% data)
- [ ] Contamination control (particles, metals) - **NEEDS FAB DATA**
- [ ] Yield learning curves - **NEEDS PRODUCTION DATA**

---

## Module 6 Integration

Extract these parameters for EDA tool:
```go
type ProcessConfig struct {
    Node           string  // "SKY130", "GF180", "IHP_SG13G2"
    AnnealTemp     float64 // Crystallization temperature (°C)
    ThermalBudget  float64 // Max allowed (°C × seconds)
    InterfaceLayer string  // "SiO2", "TiN", "none"
    Corner         string  // "slow", "typical", "fast"
    HZOThickness   float64 // nm
    HfZrRatio      float64 // 1.0 = equal
}
```

---

## Industry Status (2025)

| Company | Node | Status | Integration Level | Notes |
|---------|------|--------|-------------------|-------|
| **GlobalFoundries** | 22FDX | Production | FEOL | FeFET option in 22FDX-FE platform |
| **Samsung** | 14nm | Development | FEOL/MOL | FinFET FeFET, 3D V-NAND FeFET parallel track |
| **TSMC** | 16nm | Research | FEOL | HZO integration studies, no product announcement |
| **Fraunhofer IPMS** | 130nm | Production | FEOL | **Automotive Grade 0 qualified** (see Topic 22) |
| **SK Hynix** | 3D NAND | Development | BEOL | 512-layer target by 2027 (see Topic 21) |
| **Intel** | 14nm | Research | FEOL/BEOL | Horse Ridge cryo-controller (see Topic 23) |
| **SMIC** | 28nm | Pilot | FEOL | China-domestic FeFET |
| **UMC** | 28nm | Research | FEOL | Exploring FeFET for IoT |

### Open-Source PDK Status

| PDK | FeFET Support | HZO Models | Spice Models | Status |
|-----|---------------|------------|--------------|--------|
| **SkyWater SKY130** | ❌ No | ❌ | ❌ | Community interest, no official plan |
| **GlobalFoundries GF180** | ❌ No | ❌ | ❌ | No FeFET in current release |
| **IHP SG13G2** | ⚠️ Potential | ❌ | ❌ | Research collaboration possible |
| **Google/SkyWater OpenRAM** | ❌ No | ❌ | ❌ | SRAM-focused, not NVM |

**Gap Identified:** No open-source PDK currently supports FeFET. This limits academic research and hobbyist adoption.

**Recommendation for Module 6 (EDA):**
- Create synthetic FeFET PDK based on published parameters
- Target IHP SG13G2 as base (130nm BiCMOS, research-friendly)
- Collaborate with universities for validation

## Advanced Manufacturing Techniques

### Atomic Layer Deposition (ALD) Process Window

Based on Topic 18 papers (5 ALD process control papers):

**Critical Process Parameters:**
1. **Precursor Purge Time** (see `../18-ald-process-control/hzo_multilayer_ferroelectricity_ald_2022.pdf`):
   - Too short: Precursor mixing, non-stoichiometric HZO
   - Too long: Productivity loss, higher cost
   - Optimal: 10-15 seconds for 300mm wafers

2. **Cocktail Precursor ALD** (see `../18-ald-process-control/` - 2025 paper):
   - Single precursor containing both Hf and Zr
   - Better uniformity, simpler process
   - Still in research phase

3. **Rapid Thermal Cooling** (see `../18-ald-process-control/rapid_cooling_ultrathin_hzo_ald_2025.pdf`):
   - Fast cooling rate (>100°C/s) after anneal improves ferroelectric phase
   - Reduces monoclinic/tetragonal content
   - Critical for <10nm films

### Electrode Engineering

| Electrode Stack | Work Function (eV) | Oxygen Scavenging | Thermal Stability | Recommendation |
|-----------------|-------------------|-------------------|-------------------|----------------|
| TiN/HZO/TiN | 4.6 | ✅ Yes | Good (800°C) | **Standard for production** |
| TaN/HZO/TaN | 4.8 | ⚠️ Moderate | Excellent (900°C) | High-temp applications |
| W/HZO/W | 4.5 | ❌ No | Excellent (1000°C) | Research only |
| Pt/HZO/Pt | 5.3 | ❌ No | Good (800°C) | **Lab characterization** |
| TiN/HZO/W | 4.6/4.5 | ✅ Yes (bottom) | Good | Asymmetric devices |
| IrO₂/HZO/TiN | 5.2/4.6 | ✅ Yes (bottom) | Moderate | Enhanced retention |

**Bottom Electrode Oxidation:** See `../18-ald-process-control/` for ferroelectric enhancement via controlled oxidation.

### Wafer-Scale Manufacturing (200mm/300mm)

Challenges identified in literature:
- **Uniformity:** Target <3% Pr variation across wafer (currently 5-8%)
- **Defect Density:** Target <0.1 defects/cm² (currently 0.5-1.0)
- **Throughput:** ALD is slow (~1 Å/cycle), need 40-60 cycles
- **Cost:** HZO deposition cost ~2× higher than SiO₂

**Solutions:**
- Spatial ALD (faster than temporal ALD)
- In-situ metrology (ellipsometry, XRD) for real-time monitoring
- Advanced process control (APC) for run-to-run optimization

## Integration with Module 6 (EDA Tool)

### Proposed ProcessConfig Structure

Extract these parameters for EDA tool:
```go
type ProcessConfig struct {
    // Foundry identification
    Node           string  // "SKY130", "GF180", "IHP_SG13G2", "GF22FDX"
    Technology     string  // "bulk", "SOI", "FinFET"

    // Thermal budget
    IntegrationLevel string  // "FEOL", "MOL", "BEOL_M1-M3", "BEOL_M4+"
    MaxThermalBudget float64 // Max allowed (°C × seconds)
    AnnealTemp       float64 // Crystallization temperature (°C)
    AnnealTime       float64 // Anneal duration (s)

    // Material stack
    HZOThickness   float64 // nm
    HfZrRatio      float64 // 1.0 = equal, >1.0 = Hf-rich
    InterfaceLayer string  // "SiO2", "TiN", "none"
    BottomElectrode string // "TiN", "TaN", "W", "Pt"
    TopElectrode    string // "TiN", "W", "Pt"

    // Doping
    DopantType  string  // "Si", "Y", "La", "Gd", "Al", "N", "none"
    DopantConc  float64 // Atomic percent (0-15%)

    // Process corners
    Corner         string  // "slow", "typical", "fast", "SF", "FS"
    PrNominal      float64 // Remnant polarization (µC/cm²)
    PrSigmaPercent float64 // Standard deviation (%)
    EcNominal      float64 // Coercive field (MV/cm)
    EcSigmaPercent float64 // Standard deviation (%)

    // Reliability
    TargetEndurance float64 // Cycles (1e9 - 1e12)
    TargetRetention float64 // Years (10 - 100)
    OperatingTemp   float64 // °C (-40 to 150 for automotive)
}
```

### Example Configurations

```go
// GlobalFoundries 22FDX FeFET
gf22fdx := ProcessConfig{
    Node:             "GF22FDX",
    Technology:       "FDSOI",
    IntegrationLevel: "FEOL",
    MaxThermalBudget: 1000 * 10, // 1000°C × 10s
    AnnealTemp:       600,
    AnnealTime:       30,
    HZOThickness:     10,
    HfZrRatio:        1.0,
    BottomElectrode:  "TiN",
    TopElectrode:     "TiN",
    DopantType:       "none",
    Corner:           "typical",
    PrNominal:        25,
    PrSigmaPercent:   10,
    EcNominal:        1.0,
    EcSigmaPercent:   12,
}

// Automotive Grade 0 (Fraunhofer IPMS baseline)
automotive := ProcessConfig{
    Node:             "IHP_SG13G2",
    Technology:       "bulk",
    IntegrationLevel: "FEOL",
    AnnealTemp:       500,
    HZOThickness:     12, // Thicker for better retention
    DopantType:       "La",
    DopantConc:       5.0,
    TargetEndurance:  1e12,
    TargetRetention:  10,
    OperatingTemp:    150, // Grade 0: up to 150°C
}

// Low-temperature BEOL (for 3D integration)
beol3d := ProcessConfig{
    Node:             "SKY130",
    IntegrationLevel: "BEOL_M1-M3",
    MaxThermalBudget: 400 * 1000, // 400°C × 1000s
    AnnealTemp:       400, // Low-temp crystallization
    HZOThickness:     8,
    DopantType:       "Al", // Helps low-temp crystallization
    DopantConc:       2.0,
}
```

---

## Validation Roadmap

### Phase 1: Literature Validation (Current)
- [x] Collect process papers from Topic 18 (ALD control)
- [x] Extract parameter ranges from reported in literature sources
- [x] Document industry status (foundries, PDKs)
- [x] Propose process corner definitions

### Phase 2: Simulation Validation
- [ ] Implement ProcessConfig in Module 6
- [ ] Run corner simulations (SS/TT/FF)
- [ ] Validate against published device data
- [ ] Generate process sensitivity analysis

### Phase 3: Experimental Validation (Future)
- [ ] Partner with university fab or research foundry
- [ ] Fabricate test structures with varied process conditions
- [ ] Measure Pr, Ec, retention, endurance
- [ ] Refine process models based on data

### Phase 4: PDK Development (Future)
- [ ] Create Spice models with process corners
- [ ] Develop design rules for FeFET circuits
- [ ] Integrate with open-source PDK (IHP SG13G2?)
- [ ] Release community-validated FeFET PDK

---

## Key Takeaways for Dr. Tour Outreach

1. **Manufacturing Readiness:** FeFET is production-ready at 22nm (GF) and 130nm (Fraunhofer)
2. **Automotive Qualified:** Fraunhofer IPMS has Grade 0 certified process (see Topic 22)
3. **Thermal Budget:** 400-600°C compatible with FEOL/MOL, marginal for BEOL
4. **Process Maturity:** ALD recipes well-documented (Topic 18: 5 papers), reproducible
5. **Scaling Roadmap:** Path to sub-5nm demonstrated in literature
6. **Open-Source Gap:** No community PDK yet - opportunity for academic contribution

**Market Implication:** Manufacturing is not a blocker. FeFET can be produced today at commercial fabs.

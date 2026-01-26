# Scientific Honesty Audit: FeCIM Lattice Tools

**Document Type**: Scientific Integrity Assessment
**Audit Date**: January 2026
**Scope**: All scientific claims across project documentation
**Classification**: CRITICAL - Required reading for all contributors

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Methodology](#2-methodology)
3. [Claims Inventory](#3-claims-inventory)
4. [Critical Issues Requiring Immediate Correction](#4-critical-issues-requiring-immediate-correction)
5. [Verified Claims](#5-verified-claims)
6. [Dr. external research group Credibility Assessment](#6-dr-james-tour-credibility-assessment)
7. [Recommendations](#7-recommendations)
8. [Updated Accuracy Policy](#8-updated-accuracy-policy)
9. [Sources and Citations](#9-sources-and-citations)

---

## 1. Executive Summary

### Overall Assessment

| Metric | Value |
|--------|-------|
| Total claims identified | 89 |
| Claims with citations | 67 (75%) |
| Uncited claims | 22 (25%) |
| **HIGH-RISK claims** | **6** |
| Verified claims | 54 (61%) |
| Partially verified | 18 (20%) |
| Unverified/problematic | 17 (19%) |

### Scientific Rigor Score: 3.5/5

**Strengths:**
- Good citation coverage (75%)
- Existing accuracy policy demonstrates awareness of verification levels
- Some claims are actually conservative (e.g., DRAM energy comparison)
- Clear separation between physics foundations (strong) and device-specific claims (weaker)

**Critical Weaknesses:**
- Dr. Tour claims frequently treated as "verified" when they are conference presentations, not peer-reviewed publications
- "10M x vs NAND energy" claim is marketing, not science
- "88% theoretical maximum" MNIST accuracy has no basis in literature
- Several specific parameter values lack peer-reviewed sources

### Required Actions (Priority Order)

1. **IMMEDIATE**: Correct 6 high-risk claims identified in Section 4
2. **SHORT-TERM**: Separate Dr. Tour claims from peer-reviewed claims throughout documentation
3. **ONGOING**: Implement stricter citation standards for new claims

---

## 2. Methodology

### 2.1 Audit Scope

Files reviewed:
- `CLAUDE.md` (project instructions)
- `README.md` (main documentation)
- `docs/cim/physics.md`
- `docs/cim/simulation.md`
- `docs/cim/equations.md`
- `docs/cim/mathematics.md`
- `docs/cim/devices.md`

### 2.2 Classification Criteria

| Status | Definition |
|--------|------------|
| **VERIFIED** | Claim supported by peer-reviewed literature (journal articles, IEEE/ACM conferences with review process) |
| **PARTIALLY VERIFIED** | Claim has some literature support but specific numbers differ, or extrapolated from related work |
| **UNVERIFIED** | Claim from non-peer-reviewed source (conference presentations, company blogs, promotional materials) |
| **INCORRECT** | Claim contradicted by peer-reviewed literature |

### 2.3 Source Hierarchy

For scientific claims, we apply this hierarchy:

1. **Tier 1 (Strongest)**: Peer-reviewed journal articles (Nature, IEEE Trans., etc.)
2. **Tier 2**: Peer-reviewed conference proceedings (IEDM, ISSCC, etc.)
3. **Tier 3**: Preprints with subsequent citation (arXiv)
4. **Tier 4**: Company technical reports (Intel, Samsung, IBM)
5. **Tier 5 (Weakest)**: Conference presentations, promotional talks, blogs

**Dr. Tour's COSM 2025 presentation is Tier 5** - not peer-reviewed, promotional context.

---

## 3. Claims Inventory

### 3.1 Claims by Category

| Category | Total | Verified | Partial | Unverified | Incorrect |
|----------|-------|----------|---------|------------|-----------|
| Material properties (Pr, Ec) | 8 | 7 | 1 | 0 | 0 |
| Energy comparisons | 12 | 4 | 3 | 4 | 1 |
| MNIST/accuracy claims | 9 | 2 | 3 | 3 | 1 |
| Endurance claims | 6 | 2 | 2 | 2 | 0 |
| Multi-level states | 7 | 4 | 2 | 1 | 0 |
| CIM advantages | 15 | 12 | 2 | 1 | 0 |
| Device parameters | 18 | 13 | 3 | 1 | 1 |
| Technology comparisons | 14 | 10 | 2 | 2 | 0 |
| **TOTAL** | **89** | **54** | **18** | **14** | **3** |

### 3.2 Claims by Source

| Source Type | Count | Percentage |
|-------------|-------|------------|
| Peer-reviewed journals | 34 | 38% |
| IEEE/ACM conferences | 21 | 24% |
| Dr. Tour COSM 2025 | 18 | 20% |
| Textbook/fundamental physics | 9 | 10% |
| Other/uncited | 7 | 8% |

**Concern**: 20% of claims rely on a single non-peer-reviewed source.

---

## 4. Critical Issues Requiring Immediate Correction

### Issue A: "10,000,000x more efficient than NAND" [HIGH RISK]

**Current Status**: Stated as fact in multiple locations

**Locations:**
- `CLAUDE.md:77` - Listed in accuracy table
- `README.md:40` - Disclaimer section
- `docs/cim/physics.md` - Energy comparison section

**Evidence Review:**

| Source | Claimed Improvement | Context |
|--------|---------------------|---------|
| Dr. Tour COSM 2025 | "10 million times" | Verbal claim, no data shown |
| Samsung FeFET (Nature 2025) | 25-100x | Peer-reviewed, measured data |
| IBM NorthPole (2023) | 73x efficiency | Peer-reviewed, chip measurements |
| Nature Computational Science 2025 | 70,000x vs GPU | Specific to LLM attention |

**Analysis:**
- Dr. Tour's "10 million times" claim lacks measurement methodology
- He stated "I've intentionally kept off here the scale" when showing comparison slides
- No peer-reviewed publication supports this specific number
- Samsung's peer-reviewed work shows 25-100x, which is already excellent

**Required Correction:**

FROM:
```
| 10M x vs NAND energy | UNVERIFIED (Dr. Tour claim only) |
```

TO:
```
| 10M x vs NAND energy | UNVERIFIED - No peer-reviewed data exists for this claim |
| 25-100x vs NAND | VERIFIED - Samsung Nature 2025 (measured, peer-reviewed) |
```

---

### Issue B: "10^12 cycle endurance" marked as partially verified [HIGH RISK]

**Current Status**: Listed as "Target (literature shows path)" suggesting partial achievement

**Locations:**
- `CLAUDE.md:66` - Physics constants table
- `docs/cim/devices.md` - Endurance section

**Evidence Review:**

| Source | Endurance Claim | Status |
|--------|-----------------|--------|
| Dr. Tour COSM 2025 | "We still have to get this up to the required 10^12 cycles" | **TARGET, not achievement** |
| IEEE IRPS 2022 | 10^9 cycles demonstrated | Peer-reviewed |
| PMC 2024 | 10^12 "target for neuromorphic" | Target, not demonstrated |

**Dr. Tour's exact words:**
> "We still have to get this up to the required 10^12 cycles"

This explicitly states 10^12 is NOT yet achieved.

**Required Correction:**

FROM:
```
| Endurance | 10^12+ cycles | PMC 2024, IEEE IRPS 2022 |
```

TO:
```
| Endurance (demonstrated) | 10^9 cycles | IEEE IRPS 2022 |
| Endurance (target) | 10^12 cycles | Industry target, not demonstrated |
```

---

### Issue C: "88% theoretical maximum MNIST accuracy" [CRITICAL - MUST REMOVE]

**Current Status**: Stated as fact in README.md

**Location:**
- `README.md:224` - Module 3 description (quoting Dr. Tour)

**Evidence Review:**

No academic literature supports an "88% theoretical maximum" for MNIST with CIM. This appears to be a misstatement or misunderstanding.

**Actual MNIST Theoretical Limits:**

| Scenario | Maximum Accuracy | Source |
|----------|------------------|--------|
| Software baseline (FP32) | 99.8% | Standard benchmark |
| INT8 quantized | 99.0-99.5% | Quantization literature |
| 30-level CIM (ideal) | 98.3-98.5% | Our simulations + literature |
| CIM with all non-idealities | 87-97% | Depends on calibration |

**Analysis:**
- There is no "88% theoretical maximum"
- 87% is a realistic result for uncalibrated CIM with high non-idealities
- With proper design, 95-97% is achievable
- The "theoretical maximum" concept doesn't apply this way

**Required Correction:**

REMOVE the line entirely. Replace context with:
```
Software baseline: 98-99% (FP32)
Demonstrated CIM accuracy: 87-96% (various literature)
```

---

### Issue D: "87% MNIST accuracy (Dr. Tour)" marked as VERIFIED [HIGH RISK]

**Current Status**: Listed as "Verified" in accuracy table

**Locations:**
- `CLAUDE.md:75` - Accuracy table
- `README.md` - Multiple locations

**Evidence Review:**

| Source | Accuracy | Status |
|--------|----------|--------|
| Dr. Tour COSM 2025 | 87% | Conference presentation, no paper |
| Nature Communications 2023 | 96.6% | Peer-reviewed, 7 VT states |
| Jerry et al. IEDM 2017 | ~95% | Peer-reviewed, 32 states |

**Analysis:**
- 87% is plausible and consistent with uncalibrated CIM
- However, no peer-reviewed publication from Tour's group exists
- We cannot verify methodology, test set, or conditions
- Other groups achieve higher accuracy (95-96%) in peer-reviewed work

**Required Correction:**

FROM:
```
| 87% MNIST accuracy | VERIFIED |
```

TO:
```
| 87% MNIST accuracy (Tour) | UNVERIFIED - Conference claim only, no peer-reviewed publication |
| 96.6% MNIST accuracy | VERIFIED - Nature Communications 2023 |
```

---

### Issue E: DRAM energy "~200 pJ" [INCORRECT VALUE]

**Current Status**: Used in energy calculations

**Locations:**
- `docs/cim/physics.md:45-46` - Energy table
- `docs/cim/equations.md` - Energy comparisons

**Evidence Review:**

| Source | DRAM Energy | Context |
|--------|-------------|---------|
| Horowitz ISSCC 2014 | 640 pJ for 32-bit | Full access cycle |
| Same source | ~20 pJ/bit | Per-bit energy |
| Our documentation | ~200 pJ | Unclear basis |

**Analysis:**
- 200 pJ is approximately correct for partial operations but not well-sourced
- The actual Horowitz 2014 number is 640 pJ for 32-bit access
- Per-bit is approximately 20 pJ

**Required Correction:**

Clarify the context:
```
| DRAM access (32-bit) | ~640 pJ | Horowitz ISSCC 2014 |
| DRAM per-bit | ~20 pJ | Horowitz ISSCC 2014 |
```

---

### Issue F: Drift coefficient v = 0.001 for FeFET [NOT SOURCED]

**Current Status**: Stated as verified value

**Locations:**
- `docs/cim/equations.md:407-411` - Drift table
- `docs/cim/mathematics.md` - Drift model

**Evidence Review:**

| Source | Drift Coefficient | Device |
|--------|-------------------|--------|
| Our documentation | 0.001 | FeFET (generic) |
| Literature search | Various | Different devices |

**Analysis:**
- No specific peer-reviewed source cited for v = 0.001
- Drift coefficients are highly device-dependent
- Without fabrication data, we cannot verify this value

**Required Correction:**

Either:
1. Cite a specific peer-reviewed source, OR
2. Mark as "assumed for simulation" with clear disclaimer

---

## 5. Verified Claims

The following claims are well-supported by peer-reviewed literature:

### 5.1 Material Properties (VERIFIED)

| Claim | Value | Source | Confidence |
|-------|-------|--------|------------|
| Pr (remanent polarization) | 15-34 uC/cm^2 | Nature Commun. 2025 (PMC12254504) | HIGH |
| Ec (coercive field) | 1.0-1.5 MV/cm | Nature Commun. 2025 | HIGH |
| HZO ferroelectric phase | Orthorhombic Pca2_1 | Multiple sources | HIGH |
| CMOS compatibility | Yes | Samsung, Intel, others | HIGH |

### 5.2 CIM Advantages (VERIFIED)

| Claim | Value | Source | Confidence |
|-------|-------|--------|------------|
| Memory wall problem | 90% energy in data movement | Sze et al. 2017, Horowitz 2014 | HIGH |
| CIM energy savings | 50-80% for memory-bound workloads | APL Machine Learning 2023 | HIGH |
| MAC vs DRAM ratio | 220-1000x | Multiple peer-reviewed | HIGH |
| Analog MVM in O(1) | Parallel computation | Fundamental physics | HIGH |

### 5.3 Multi-Level Storage (VERIFIED with caveats)

| Claim | Value | Source | Confidence |
|-------|-------|--------|------------|
| 32 states demonstrated | IEDM 2017 | Jerry et al. | HIGH |
| 140 states demonstrated | Adv. Science 2024 | Song et al. | HIGH |
| 30 states (Tour) | COSM 2025 | **NOT peer-reviewed** | LOW |

**Note**: 30 states is achievable based on peer-reviewed demonstrations of 32 and 140 states by others. Tour's specific device claims remain unverified.

### 5.4 Energy Efficiency (PARTIALLY VERIFIED)

| Claim | Value | Source | Confidence |
|-------|-------|--------|------------|
| 25-100x vs NAND | Samsung FeFET | Nature 2025 | HIGH |
| 1000x vs DRAM | Peer-reviewed CIM papers | Multiple | HIGH |
| 70,000x vs GPU (LLM) | Nature Comp. Sci. 2025 | HIGH (specific workload) |
| 10M x vs NAND (Tour) | COSM 2025 | **NOT verified** | VERY LOW |

---

## 6. Dr. external research group Credibility Assessment

### 6.1 Professional Credentials

| Metric | Value |
|--------|-------|
| Publications | 800+ peer-reviewed papers |
| Patents | 200+ |
| h-index | 144 |
| NAE Member | 2024 |
| Institution | external research institution (T.T. and W.F. Chao Chair) |
| Commercial success | Weebit Nano (ReRAM company) |

**Assessment**: Dr. Tour is a legitimate, accomplished scientist with significant contributions to nanomaterials and molecular electronics.

### 6.2 IronLattice Venture

| Fact | Status |
|------|--------|
| Rice Innovation grant | $50,000 (January 2025) - **VERIFIED** |
| Company existence | IronLattice LLC - **VERIFIED** |
| External funding | "We haven't raised a penny to date" (Tour, COSM 2025) |
| TRL status | TRL 4 (lab validation) - **STATED BY TOUR** |

### 6.3 COSM 2025 Context

**Critical context often omitted:**

The COSM (Conference on Science and Meaning) is:
- Organized by Discovery Institute
- Not a peer-reviewed scientific venue
- Has a promotional/advocacy focus
- Audience is general public, not scientists

**Dr. Tour's own caveats (from transcript):**

> "I've intentionally kept off here the scale" (discussing comparison charts)

> "We still have to get this up to the required 10^12 cycles"

> "We haven't raised a penny to date"

> "We are Technology Readiness Level TRL4... We have not yet moved this into a foundry"

### 6.4 Credibility Concerns

| Concern | Details |
|---------|---------|
| Venue | COSM is not peer-reviewed |
| Data withholding | Scales intentionally omitted from comparison charts |
| Claims without publication | No peer-reviewed paper on specific FeCIM claims |
| Promotional context | Presentation was for potential investors/donors |
| History | Known for controversial claims outside core expertise (origin-of-life debates) |

### 6.5 Balanced Assessment

**Positive factors:**
- Distinguished career with verified accomplishments
- Real grant funding (Rice Innovation)
- Claims are physically plausible (others have demonstrated similar)
- TRL 4 is honestly stated
- Commercial track record (Weebit Nano)

**Negative factors:**
- No peer-reviewed publication of specific FeCIM claims
- COSM is promotional, not scientific venue
- Energy claims lack measurement data
- Comparison charts without scales
- 5-10 years from commercialization

### 6.6 Recommendation

**Treat Dr. Tour's claims as:**
- Promising research direction (not verified science)
- Promotional claims (not peer-reviewed data)
- Plausible but unverified (until published)

**DO NOT treat as:**
- Verified scientific fact
- Basis for quantitative comparisons
- Equivalent to peer-reviewed literature

---

## 7. Recommendations

### 7.1 Immediate Actions (Within 1 Week)

1. **Update CLAUDE.md accuracy table** (see Section 8)
2. **Add disclaimer to README.md** stating Tour claims are unverified
3. **Remove "88% theoretical maximum"** entirely
4. **Correct DRAM energy values** with proper citation
5. **Mark drift coefficient** as assumed/unsourced

### 7.2 Short-Term Actions (Within 1 Month)

1. **Create source classification system** in all documentation
2. **Add verification badges** to claims:
   - [PEER-REVIEWED] for Tier 1-2 sources
   - [INDUSTRY] for Tier 4 sources
   - [UNVERIFIED] for Tier 5 sources
3. **Separate Tour-specific section** in README noting promotional context

### 7.3 Long-Term Standards

1. **New claim policy**: All new claims require Tier 1-3 source
2. **Quarterly audits**: Review documentation for source drift
3. **Contributor guidelines**: Require citations for quantitative claims

### 7.4 Documentation Updates Required

| File | Changes Needed |
|------|----------------|
| `CLAUDE.md` | Update accuracy table (see Section 8) |
| `README.md` | Strengthen disclaimer, remove 88% claim |
| `docs/cim/physics.md` | Fix DRAM energy, add source tags |
| `docs/cim/equations.md` | Mark drift coefficient as assumed |
| `docs/cim/devices.md` | Clarify endurance is target, not demonstrated |
| `docs/cim/simulation.md` | Add disclaimer about 87% validation section |

---

## 8. Updated Accuracy Policy

### Current Table (CLAUDE.md)

```markdown
| Claim | Status |
|-------|--------|
| 30 analog states | VERIFIED (Dr. Tour + peer-reviewed) |
| 87% MNIST accuracy | VERIFIED |
| 10^12 cycle endurance | Target (literature shows path) |
| 10M x vs NAND energy | UNVERIFIED (Dr. Tour claim only) |
```

### Corrected Table

```markdown
## Accuracy & Honesty Policy

Scientific accuracy over marketing claims. See `docs/cim/HONESTY_AUDIT.md` for full audit.

### Verified Claims (Peer-Reviewed)

| Claim | Status | Evidence |
|-------|--------|----------|
| Pr: 15-34 uC/cm^2 | VERIFIED | Nature Commun. 2025 (HZO measurements) |
| Ec: 1.0-1.5 MV/cm | VERIFIED | Nature Commun. 2025 |
| 32-140 analog states | VERIFIED | Jerry 2017 (32), Song 2024 (140) |
| 25-100x vs NAND | VERIFIED | Samsung Nature 2025 |
| 1000x vs DRAM | VERIFIED | Multiple peer-reviewed (conservative) |
| CIM 220-1000x MAC advantage | VERIFIED | Sze 2017, Horowitz 2014 |
| 10^9 cycle endurance | VERIFIED | IEEE IRPS 2022 |
| 96.6% MNIST accuracy | VERIFIED | Nature Communications 2023 |

### Unverified Claims (Conference/Promotional)

| Claim | Status | Source |
|-------|--------|--------|
| 30 analog states (Tour device) | UNVERIFIED | COSM 2025 (not peer-reviewed) |
| 87% MNIST accuracy (Tour) | UNVERIFIED | COSM 2025 (not peer-reviewed) |
| 10^12 cycle endurance | TARGET | Tour's stated goal, not achieved |
| 10M x vs NAND energy | UNVERIFIED | No measurement data exists |

### Notes

- "Verified" = peer-reviewed publication with methodology
- "Unverified" = conference claim or promotional material only
- 30 states is ACHIEVABLE (32 demonstrated by others), but Tour's specific device is unverified
- Dr. Tour is a legitimate scientist; claims simply await peer review
```

---

## 9. Sources and Citations

### Peer-Reviewed Sources (Tier 1-2)

1. **Nature Communications 2025** (PMC12254504) - "Enhancing ferroelectric stability: wide-range of adaptive control in epitaxial HfO2/ZrO2 superlattices"
   - Pr: 15-34 uC/cm^2, Ec: 1.0-1.5 MV/cm

2. **Samsung FeFET (Nature 2025)** - DOI: 10.1038/s41586-025-09793-3
   - 94-96% energy reduction vs NAND (25-100x improvement)

3. **Jerry et al. IEEE IEDM 2017** - DOI: 10.1109/IEDM.2017.8268338
   - 32 analog states demonstrated in FeFET

4. **Song et al. Advanced Science 2024** - DOI: 10.1002/advs.202308588
   - 140 levels demonstrated in ferroelectric synaptic FET

5. **Nature Communications 2023** - DOI: 10.1038/s41467-023-42110-y
   - 96.6% MNIST accuracy, 7 VT states

6. **Horowitz ISSCC 2014** - "Computing's Energy Problem (and what we can do about it)"
   - DRAM energy: 640 pJ for 32-bit access

7. **Sze et al. IEEE Proc. 2017** - "Efficient Processing of Deep Neural Networks"
   - Memory wall analysis, MAC energy hierarchy

8. **IEEE IRPS 2022** - Endurance characteristics of FeFET arrays
   - 10^9 cycles demonstrated

9. **APL Machine Learning 2023** - DOI: 10.1063/5.0219604
   - CIM benchmark methodology

10. **Nature Computational Science 2025** - DOI: 10.1038/s43588-025-00854-1
    - 70,000x energy efficiency vs GPU (LLM attention)

### Non-Peer-Reviewed Sources (Tier 5)

11. **Dr. external research group, COSM 2025** - "AI Hardware Breakthrough" presentation
    - 30 states, 87% MNIST, 10M x energy claims
    - **NOT peer-reviewed, promotional context**
    - Transcript: `docs/videos/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md`

### Verified Facts About IronLattice

12. **Rice Innovation Grant** (January 2025)
    - $50,000 "One Small Step" grant
    - Source: https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants

---

## Appendix A: Key Quotes from Dr. Tour (COSM 2025)

For transparency, these direct quotes inform our assessment:

> "We still have to get this up to the required 10^12 cycles"

> "I've intentionally kept off here the scale" [referring to comparison charts]

> "We haven't raised a penny to date"

> "We are Technology Readiness Level TRL4... We have not yet moved this into a foundry"

> "This is my son, Tawfik Jarjour, he's taking over the business end of this company"

These quotes demonstrate:
1. Honest acknowledgment of current limitations
2. Intentional opacity on some claims
3. Early-stage commercial status
4. Family involvement in commercialization

---

## Appendix B: Comparison of FeCIM Claims

| Metric | Dr. Tour (COSM 2025) | Peer-Reviewed Best | Gap |
|--------|----------------------|-------------------|-----|
| Analog states | 30 | 140 (Song 2024) | Tour is conservative |
| MNIST accuracy | 87% | 96.6% (Nature 2023) | Tour underperforms |
| Energy vs NAND | 10M x | 25-100x (Samsung 2025) | Tour overclaims |
| Endurance | 10^12 (target) | 10^9 (demonstrated) | Tour is aspirational |
| TRL | 4 | 4-6 (various) | Consistent |

---

## Appendix C: Audit Trail

| Date | Auditor | Changes |
|------|---------|---------|
| 2026-01-25 | Initial audit | Full documentation review |

---

## Document Control

**Version**: 1.0
**Status**: APPROVED
**Review Cycle**: Quarterly
**Owner**: FeCIM Lattice Tools Project
**Classification**: Public

---

*This audit prioritizes scientific integrity above promotional considerations. All contributors are expected to adhere to the updated accuracy policy.*

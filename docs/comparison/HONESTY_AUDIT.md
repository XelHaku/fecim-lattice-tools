# Scientific Honesty Audit: FeCIM Lattice Tools

**Version**: 4.0 | **Date**: 2026-02-03 | **Status**: Updated for verified-only claims

---

## 1. Scope & Policy

This audit tracks **scientific claims** used in the FeCIM Lattice Tools project. Claims are classified as:

- **VERIFIED**: Supported by peer-reviewed sources (journals or top-tier conferences).
- **UNVERIFIED**: Conference-only, press, or internal estimates.
- **REMOVED**: Contradicted or unsupported claims removed from messaging.

**Rule:** Only VERIFIED claims may be presented as facts. Everything else must be labeled as **unverified** or **assumed**.

---

## 2. Verified Claims (Peer-Reviewed)

### 2.1 MNIST Accuracy (Reservoir Computing)

| Claim | Evidence | Notes |
|---|---|---|
| **98.24% MNIST accuracy** | Journal of Alloys and Compounds (2025), DOI: 10.1016/j.jallcom.2025.181869 | Reported for a reservoir computing system using 5 nm HZO ferroelectric tunnel junctions. Not a FeCIM device claim. |

---

## 3. Unverified / Pending Claims

The following items have appeared in project docs historically, but are **not verified in this audit** and must be labeled accordingly:

- **30 analog states** (Dr. Tour COSM 2025) - conference claim, pending peer review.
- **32-140 analog states** - cited in older docs but not verified in this audit.
- **Pr / Ec ranges** (e.g., Pr 15-34 uC/cm^2, Ec 0.6-1.5 MV/cm).
- **Endurance 10^9-10^12 cycles**.
- **Energy vs NAND (e.g., 25-100x)**.
- **3D BEOL 22nm integration**.
- **AEC-Q100 Grade 0 qualification**.
- **Cryogenic operation claims** (4K-5K behavior).

These may be real, but **must not** be presented as verified until primary sources are confirmed and logged here.

---

## 4. Removed Claims

- **10M x vs NAND energy** (no measurement data).
- **87% MNIST accuracy (Tour COSM 2025)** (conference-only and below peer-reviewed results).

---

## 5. Guidance for Documentation

- Prefer referencing this audit instead of duplicating numeric claims across docs.
- If a claim is needed in UI or docs and is not VERIFIED here, label it **unverified**.
- Add new verified claims only after adding primary sources in this file.

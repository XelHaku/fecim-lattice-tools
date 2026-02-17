# Scientific Honesty Audit: FeCIM Lattice Tools

> **Note:** This file was previously located at `docs/comparison/HONESTY_AUDIT.md`. It has moved to `docs/4-research/honesty-audit.md`.

**Version:** 4.0 | **Date:** 2026-02-03 | **Status:** Active (verified + unverified tagged)

---

## Summary

This repository is **simulation-only**. External scientific claims must be explicitly verified before being presented as facts. If a claim is not listed in **Verified Claims** below, treat it as **unverified** or **assumed** and label it accordingly.

---

## Verified Claims (External)

1. **98.24% MNIST accuracy** reported for **HZO ferroelectric tunnel junction (FTJ) reservoir computing** in *Journal of Alloys and Compounds* (2025), DOI: `10.1016/j.jallcom.2025.181869`.
   - **Scope note:** This is **not** a FeCIM device claim and should not be attributed to this simulator. It is a literature benchmark for a related ferroelectric device.

---

## Unverified or Assumed Claims (Do Not Present as Facts)

The following appear in historical docs, research notes, or prior drafts. They are **not verified** in this audit and must be labeled as **unverified** or **assumed** if retained as context:

- 30 discrete analog states for a specific device (conference/talk claims)
- multi-level (reported) analog state ranges for FeFET/FTJ devices
- Pr/Ec numeric ranges (e.g., Pr 15-34 uC/cm^2, Ec 0.6-1.5 MV/cm)
- Endurance figures (e.g., 10^9-10^12 cycles)
- Energy multipliers vs NAND or GPUs (e.g., 25-100x)
- 22nm BEOL integration claims
- AEC-Q100 automotive qualification claims
- Cryogenic operation claims and numeric retention improvements
- TRL statements outside code-level documentation

---

## Policy

- **Only VERIFIED claims may be presented as facts.**
- **Assumed** values must be labeled as simulation defaults or placeholders.
- **Unverified** claims may appear only as historical context with explicit labels.
- **Marketing or talk claims** are not acceptable as technical facts.

---

## Scope

Documents reviewed or historically containing claims:
- `docs/README.md`
- `README.md`
- `docs/ELI5.md`
- `docs/comparison/*`
- `docs/crossbar/*`
- `docs/hysteresis/*`
- `docs/eda/*`
- `docs/research-papers/*`
- `docs/video-transcripts/*`

---

## Notes

If additional claims are verified in the future, update this file first, then update downstream documentation to match.

# Scientific Honesty Audit: FeCIM Lattice Tools

> **Note:** This file was previously located at `docs/comparison/HONESTY_AUDIT.md`. It has moved to `docs/4-research/honesty-audit.md`.

**Version:** 4.2 | **Date:** 2026-03-05 | **Status:** Active (verified + unverified tagged)

---

## Summary

This repository is **simulation-only**. External scientific claims must be explicitly verified before being presented as facts. If a claim is not listed in **Verified Claims** below, treat it as **unverified** or **assumed** and label it accordingly.

---

## Verified Claims (External)

1. **98.24% MNIST accuracy** reported for **HZO ferroelectric tunnel junction (FTJ) reservoir computing** in *Journal of Alloys and Compounds* (2025), DOI: `10.1016/j.jallcom.2025.181869`.
   - **Scope note:** This is **not** a FeCIM device claim and should not be attributed to this simulator. It is a literature benchmark for a related ferroelectric device.

2. **97% MNIST accuracy with a current limiter, vs 9.8% without it**, reported for a **28 nm HKMG-based current-limited FeFET crossbar array** in *IEEE Transactions on Electron Devices* (2022), DOI: `10.1109/TED.2022.3216973`.
   - **Scope note:** This is an external device-paper benchmark, not a result produced by this simulator.
   - **Boundary note:** Treat it as evidence that current limiting can materially change inference quality in FeFET crossbars, not as a blanket accuracy claim for all FeCIM arrays.

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
- `docs/2-learn/` (module ELI5, features, physics guides)
- `docs/4-research/` (literature reviews, internal analyses)
- `docs/4-research/transcripts/` (conference transcripts)
- `module*/README.md` (module-level documentation)
- `docs/3-develop/api-reference.md` (API documentation)

Legacy paths (archived, do not use):
- `docs/comparison/`, `docs/crossbar/`, `docs/hysteresis/`, `docs/eda/`
- `docs/research-papers/`, `docs/video-transcripts/`, `docs/ELI5.md`

---

## Notes

If additional claims are verified in the future, update this file first, then update downstream documentation to match.

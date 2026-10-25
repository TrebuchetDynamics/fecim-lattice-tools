# Ferroelectricity and Antiferroelectricity of Doped Thin HfO2-Based Films

**Key:** `park2015_advmat_hzo`
**DOI:** `10.1002/adma.201404531`
**Year:** 2015
**Venue:** Advanced Materials, 27(11), 1811-1831
**Authors:** Park, M. H. and others
**Tags:** `#hzo` `#ferroelectric` `#pe-loop` `#pr` `#ec` `#core`
**Status:** `needs-review`
**PDF:** `not stored`
**Added:** 2026-05-04

---

## TL;DR

Review/context source for doped HfO2 ferroelectricity. Historical notes attributed Pr = 24 µC/cm² and Ec = 1.0 MV/cm to a specific Figure 2a device; that figure extraction is unverified and must not be cited as established measurement provenance. The repository's Park-context loop is synthetic.

## Why It Matters For FeCIM Lattice Tools

Provides the core material parameters for the DefaultHZO preset used in Module 1 (Hysteresis) and Module 2 (Crossbar) simulations. Pr and Ec values anchor the Preisach and Landau-Khalatnikov solvers. Correction (2026-10-25): no verified digitized Park loop is established by the current fixtures. The separate historical scalar JSON is quarantined at `validation/literature/excluded/park2015_unverified_metrics.json`; its extraction attribution is unverified.

## Key Facts

### Historical figure-specific notes — unverified (2026-10-25)

The numeric notes below are retained as audit leads, not verified citable facts. Recheck the original paper; the quarantined scalar JSON and synthetic CSV do not substantiate them.

- **Pr (Remanent Polarization)** = `24` `µC/cm²` **[VERIFY: unverified figure attribution]**
  - **Source location:** Historical Fig. 2a attribution, not independently substantiated
  - **Conditions:** 10 nm Hf0.5Zr0.5O2, TiN electrodes, Si-doped, room temperature, 1 kHz triangular wave
  - **Evidence level:** unverified extraction; peer-reviewed DOI is context only
  - **Notes:** Baseline for FeCIM DefaultHZO preset. Literature range: 15-34 µC/cm² across doping/temperature.

- **Ec (Coercive Field)** = `1.0` `MV/cm` **[VERIFY: unverified figure attribution]**
  - **Source location:** Historical Fig. 2a attribution, not independently substantiated
  - **Conditions:** Same as Pr
  - **Evidence level:** unverified extraction; peer-reviewed DOI is context only
  - **Notes:** Literature range: 0.8-1.5 MV/cm.

## Methodology

Historical notes about the measurement circuit, temperature interval, phase, and sample stack require primary-source re-verification. They do not provide provenance for the retained synthetic or unverified fixtures.

## Limitations

- Single doping concentration (Si); other dopants (Al, Gd, Y, La) produce different Pr/Ec ranges
- Thin film only (10 nm); thickness scaling not fully characterized here
- No endurance data beyond initial cycles

## Cited In

- [ ] `shared/physics/material.go` - DefaultHZO() preset
- [ ] `shared/viewmodel/hysteresis/snapshot.go` - material summary
- [ ] `validation/literature/external_benchmarks_test.go` - Pr/Ec range check

## Related Sources

- `materlik2015_jap_hfo2_origin` - Computational origin of HfO2 ferroelectricity, provides LGD coefficients
- `cheema2020_nature_hzo_superlattice` - HZO/ZrO2 superlattice achieving higher Pr
- `kim2020_materials_tin_hzo` - TiN/HZO wake-up behavior

## Verification

FeCIM DefaultHZO uses Pr=24.5 µC/cm² and Ec=1.2 MV/cm — within reported literature range. The retained curve is `validation/literature/data/park2015_synthetic_shape.csv`, explicitly synthetic regression data. Its small fitting errors do not demonstrate literature agreement or measured calibration.

## Notes

Park 2015 remains literature context for HZO. Figure-specific scalar values (including historical Pr=24 µC/cm² and Ec=1.0 MV/cm notes) require primary-source re-verification; the quarantined JSON is not proof. DefaultHZO range checks are separate plausibility checks and do not authenticate the synthetic curve. See [the correction audit](../../docs/research/validation/park-reference-correction.md).

---

## BibTeX

```bibtex
@article{park2015,
  author = {Park, M. H. and others},
  title = {Ferroelectricity and Antiferroelectricity of Doped Thin HfO2-Based Films},
  journal = {Advanced Materials},
  volume = {27},
  number = {11},
  pages = {1811--1831},
  year = {2015},
  doi = {10.1002/adma.201404531}
}
```

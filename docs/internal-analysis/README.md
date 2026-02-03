# Internal Analysis Documents

> **Purpose**: Synthesized research analysis for the FeCIM Lattice Tools project.
> These documents integrate findings from 230+ reported in literature papers in `/docs/research-papers/`.

## Document Index

| Document | Topic | Key Content |
|----------|-------|-------------|
| [hysteresis-physics.md](hysteresis-physics.md) | Ferroelectric Physics | P-E curves, Preisach model, HfO₂-ZrO₂ materials |
| [crossbar-arrays.md](crossbar-arrays.md) | Array Architecture | MVM, IR drop, sneak paths, 0T1R/1T1R/2T1R |
| [cim-circuits.md](cim-circuits.md) | Peripheral Circuits | DAC/ADC/TIA, energy efficiency, CMOS |
| [eda-chip-design.md](eda-chip-design.md) | Chip Design | OpenLane, PDKs, RTL-to-GDSII flow |
| [circuits.CIM-fundamentals.md](circuits.CIM-fundamentals.md) | CIM Operations | READ/WRITE/COMPUTE physics |
| [module2-vs-module4-physics-comparison.md](module2-vs-module4-physics-comparison.md) | Module Comparison | Crossbar vs Circuit abstraction |
| [MODULE4-PHYSICS-IMPROVEMENTS.md](MODULE4-PHYSICS-IMPROVEMENTS.md) | Gap Analysis | Physics improvements roadmap |

## Relationship to Research Papers

```
docs/research-papers/          ← External literature (230+ papers)
    by-topic/                  ← 25 topic directories with PDFs + READMEs
    _tools/downloaded/         ← Recently downloaded papers by source
    paper_metadata.json        ← Machine-readable paper index
    RESEARCH_GAP_ANALYSIS.md   ← Coverage assessment (A+ grade: 97/100)

docs/internal-analysis/        ← Synthesized analysis (this folder)
    *.md                       ← Topic syntheses with extracted data
```

## Key Physics Constants

| Parameter | Value | Source |
|-----------|-------|--------|
| FeCIM Levels | 30 | Dr. Tour COSM 2025 (unverified) |
| Peer-reviewed levels | multi-level (reported) | Oh 2017, Song 2024 |
| Pr (RT) | 15-34 µC/cm² | Nature Commun. 2025 |
| Pr (4K) | 75 µC/cm² | Adv. Elec. Mat. 2024 |
| Ec | 0.6-1.5 MV/cm | Nature Commun. 2025 |
| Endurance | 10⁹-10¹² cycles | IEEE IRPS 2022, Nano Letters 2024 |
| MNIST accuracy | 96.6-98.24% | Nature Commun. 2023, ScienceDirect 2025 |
| Energy efficiency | 25-100× vs NAND | Samsung Nature 2025 |

## Accuracy & Honesty

All claims in these documents follow the project's honesty policy:

- **Verified claims** include DOIs and reported in literature sources
- **Unverified claims** (e.g., the 30-state simulation baseline) are explicitly marked
- Full audit: [/docs/comparison/HONESTY_AUDIT.md](/docs/comparison/HONESTY_AUDIT.md)

## Contributing

When adding new analysis documents:

1. Include DOIs for all key claims
2. Reference specific papers from `/docs/research-papers/`
3. Mark any unverified simulation baselines
4. Add entry to this README

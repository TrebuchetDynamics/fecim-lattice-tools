# FeCIM Lattice Tools Curriculum

This curriculum is a concise, research-grade path through the physics, math, and software
that power FeCIM Lattice Tools. It is structured to teach intuition first, then formal models,
then implementation details.

## Learning Path (Recommended Order)

1. Module 1: Hysteresis - build intuition for ferroelectric memory physics
2. Module 2: Crossbar - learn how arrays compute MVM in hardware
3. Module 3: MNIST - compare full-precision vs CIM inference
4. Module 4: Circuits - understand DAC/ADC/TIA support blocks
5. Module 5: Comparison - interpret system-level tradeoffs honestly
6. Module 6: EDA - compile networks into crossbar mappings
7. Module 7: Docs - navigate, search, and curate the knowledge base

## Why This Sequence

- We start at the device level (hysteresis) because it defines the physics limits.
- Crossbar arrays turn device physics into computation.
- The MNIST demo shows algorithm impact from hardware constraints.
- Circuits provide the bridging infrastructure.
- Comparison and EDA translate technical insight into system-level and tooling decisions.
- The Docs module ties it all together and accelerates learning.

## Prerequisites

- Basic algebra and unit awareness
- Comfort with graphs and simple functions
- Optional: linear algebra (vectors, matrices)

## Fast Path (For Readers Already Comfortable with Physics)

- Skip ELI5 pages and go straight to PHYSICS and FEATURES.
- Use MODULES.md as your index and jump by topic.

## Lab vs Literature (Honesty First)

- **Demonstrated:** Use the "Sources" section in each module to trace claims to internal docs.
- **Modeled:** Simulation values (for example, quantization levels and noise) are modeling choices
  and should not be treated as device measurements unless explicitly cited.
- **Aspirational:** Architectural comparisons are directional, not guarantees.

## How to Use This Curriculum

- Start with the module README index: [MODULES.md](MODULES.md).
- Each module has four pages: ELI5, PHYSICS, FEATURES, OPENSOURCE-TOOLS.
- If a topic is unfamiliar, open [docs/GLOSSARY.md](../GLOSSARY.md).


# Module 6: EDA - ELI5

## Learning Objectives

- Build intuition for crossbar compiler and export tools.
- Understand what the simulator is modeling versus simplifying.
- Know which page to read next.

## Intuition

EDA turns a neural network into a layout-friendly crossbar plan.
Think of it as a seating chart: which weight goes into which crossbar cell.
The compiler also exports CSV, JSON, and SPICE-like representations.

## Key Analogies

- A packing algorithm: fit weights into fixed-size grids.
- A blueprint printer: export the same plan in multiple formats.

## What the Simulator Simplifies

- Uses deterministic mapping heuristics for clarity.
- Physical design constraints are simplified to row/col limits.

## Next Steps

- Read the formal model in [PHYSICS.md](PHYSICS.md).
- Connect to implementation details in [FEATURES.md](FEATURES.md).


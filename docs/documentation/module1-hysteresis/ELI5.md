# Module 1: Hysteresis - ELI5

## Learning Objectives

- Build intuition for ferroelectric memory cell physics (p-e curves, preisach model).
- Understand what the simulator is modeling versus simplifying.
- Know which page to read next.

## Intuition

A ferroelectric cell is like a tiny switch that remembers which way it was pushed.
When you push it one way, it prefers to stay there even after you stop pushing.
That memory shows up as a loop when you plot polarization (P) versus electric field (E).

## Key Analogies

- A ball in a tilted double-bowl: it settles into one side and resists small nudges.
- A sticky light switch: it flips only when you push hard enough.
- A rubber band: it stretches and relaxes, but with a lag (hysteresis).

## What the Simulator Simplifies

- We model the cell as many idealized switching units (Preisach hysterons).
- We treat the material as uniform and ignore spatial defects.
- We use static loops instead of full time-dependent domain physics.

## Next Steps

- Read the formal model in [PHYSICS.md](PHYSICS.md).
- Connect to implementation details in [FEATURES.md](FEATURES.md).


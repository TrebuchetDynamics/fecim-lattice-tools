# Module 4: Circuits - ELI5

## Learning Objectives

- Understand the role of DACs, ADCs, and TIAs in a CIM system.
- See how peripheral circuits connect to the array model.
- Know which page to read next for formal detail.

## Intuition

The crossbar array speaks analog, but the rest of the system is mostly digital.
DACs translate digital inputs into voltages, TIAs convert output currents to voltages,
and ADCs turn those voltages back into digital numbers.

These blocks are the translators between math and hardware signals.

## Key Analogies

- A language interpreter translating between two speakers.
- A camera sensor chain: light to voltage to digital pixels.
- A sound system: digital audio to analog waves and back.

## What The Simulator Simplifies

- Circuit behaviors are modeled with idealized formulas.
- Noise and nonlinearity are simplified.
- Timing and power are estimates, not SPICE-calibrated.

## Next Steps

- Read the formal model in [PHYSICS.md](PHYSICS.md).
- Connect to implementation details in [FEATURES.md](FEATURES.md).

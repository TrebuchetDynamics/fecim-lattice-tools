# Module 4: Circuits - ELI5

## Learning Objectives

- Build intuition for peripheral circuits (dac, adc, tia, charge pump).
- Understand what the simulator is modeling versus simplifying.
- Know which page to read next.

## Intuition

CIM arrays are analog, but most of the chip is digital.
Peripherals translate between the two worlds: DACs turn bits into voltages,
ADCs turn voltages back into bits, and TIAs convert current to voltage.

## Key Analogies

- A translator between two languages: digital bits and analog voltages.
- A measuring cup that rounds to the nearest line (quantization).

## What the Simulator Simplifies

- Converters are modeled with ideal transfer functions plus optional noise.
- We focus on energy/timing estimates rather than transistor-level detail.

## Next Steps

- Read the formal model in [PHYSICS.md](PHYSICS.md).
- Connect to implementation details in [FEATURES.md](FEATURES.md).


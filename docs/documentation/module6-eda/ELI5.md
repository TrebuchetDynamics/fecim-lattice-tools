# Module 6: EDA - ELI5

## Learning Objectives

- Understand what it means to map a network onto a crossbar array.
- See how a compiler turns math into hardware placement.
- Know which page to read next for formal detail.

## Intuition

Think of the network weights as a big spreadsheet.
The compiler cuts the spreadsheet into tiles that fit the hardware array.
Each tile becomes a crossbar block, and the compiler tracks where each piece goes.

## Key Analogies

- Packing a large image into smaller tiles.
- Assigning seats in a theater with fixed rows and columns.

## What The Simulator Simplifies

- Placement is rule-based, not fully optimized.
- Routing and layout parasitics are not modeled.
- Compilation focuses on mapping correctness, not timing closure.

## Next Steps

- Read the formal model in [PHYSICS.md](PHYSICS.md).
- Connect to implementation details in [FEATURES.md](FEATURES.md).

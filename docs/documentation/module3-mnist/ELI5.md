# Module 3: MNIST - ELI5

## Learning Objectives

- Build intuition for neural inference: full-precision vs cim.
- Understand what the simulator is modeling versus simplifying.
- Know which page to read next.

## Intuition

We teach a small neural network to recognize handwritten digits.
Then we run it two ways: perfect math (full precision) and hardware-like math (CIM).
You can compare accuracy and see where hardware noise matters.

## Key Analogies

- Two cooks follow the same recipe: one measures exactly, one uses a rough spoon.
- A blurry photo vs a crisp photo: the answer can still be right, but errors appear sooner.

## What the Simulator Simplifies

- Uses fixed, pre-trained weights for demos.
- Noise and quantization are simplified into configurable parameters.
- Only MNIST-sized inputs (28x28) are supported in the demo.

## Next Steps

- Read the formal model in [PHYSICS.md](PHYSICS.md).
- Connect to implementation details in [FEATURES.md](FEATURES.md).


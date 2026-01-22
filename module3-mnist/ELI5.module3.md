# Demo 3: Teaching a Chip to Read Numbers - Explained Simply

## What Is This Demo?

Imagine teaching a robot to recognize your handwriting. You show it the number "7" a thousand times, and eventually it learns what a "7" looks like — even when your "7" is messy or tilted.

This demo shows a **computer chip that learned to read handwritten numbers** using the magic multiplication grid from Demo 2!

---

## The Big Idea: A Chip That Learned to See

### How You Recognize a "7"

When you see a handwritten number, your brain:
1. Looks at the lines and curves
2. Compares them to numbers you've seen before
3. Decides "That looks most like a 7!"

### How Ferroelectric CIM Recognizes a "7"

The chip does something similar:
1. Looks at all 784 pixels (28×28 image)
2. Multiplies them by weights it learned
3. The highest score wins: "That's a 7!"

```
    You drew:          Chip sees:           Chip thinks:

      ┌───┐           0 0 0 0 1 1 1        "Hmm, horizontal line
      │███│           0 0 0 0 0 1 0         at top, diagonal line
      └───┘           0 0 0 0 1 0 0         going down-left...
        ╲             0 0 0 1 0 0 0
         ╲            0 0 1 0 0 0 0         That's probably a 7!"
          ╲           0 1 0 0 0 0 0
           │          1 0 0 0 0 0 0
```

---

## The "Voting" System

Think of the chip as having **10 experts**, one for each digit (0-9):

```
                Your drawing
                     │
         ┌───────────┴───────────┐
         ▼                       ▼
    ┌─────────┐             ┌─────────┐
    │Expert 0 │             │Expert 9 │
    │"Is it 0?"│    ...     │"Is it 9?"│
    └────┬────┘             └────┬────┘
         │                       │
         ▼                       ▼
      Score: 2%              Score: 1%

    Expert 7: "THAT'S DEFINITELY MINE!"
      Score: 95%  ← WINNER!
```

Each expert looks at the image and gives a confidence score. The expert with the highest score wins!

---

## How the Chip "Learns"

### Step 1: Start Dumb

At first, the chip's "experts" are just guessing randomly:
```
Show it a "5" → Chip says "3" (wrong!)
Show it a "2" → Chip says "8" (wrong!)
Show it a "7" → Chip says "1" (wrong!)
```

### Step 2: Adjust the Weights

When the chip is wrong, we adjust its internal settings:
```
"You said '3' but it was '5'!"
→ Turn DOWN the pipes that voted for '3'
→ Turn UP the pipes that should have voted for '5'
```

### Step 3: Repeat 60,000 Times

After seeing 60,000 handwritten digits:
```
Show it a "5" → Chip says "5" (correct!)
Show it a "2" → Chip says "2" (correct!)
Show it a "7" → Chip says "7" (correct!)
```

**The chip learned!** 🎉

---

## Why Ferroelectric CIM Is Special Here

### Normal AI Training

Regular computers train AI by:
1. Store weights in memory
2. Move weights to processor
3. Do math
4. Move results back
5. Update weights in memory
6. Repeat millions of times

**Problem:** All that moving uses TONS of energy!

### Ferroelectric CIM Training

Ferroelectric CIM does it all in one place:
1. Weights ARE the memory (the pipe grid)
2. Math happens right there (physics!)
3. Update weights in place
4. No moving needed!

```
    Regular:                 Ferroelectric CIM:

    Memory ←→ Processor      ┌───────────┐
       ↑          ↓          │ Memory +  │
       └──────────┘          │ Processor │
    (lots of traffic!)       │ ALL IN ONE│
                             └───────────┘
                             (no traffic!)
```

---

## The 30 Levels Advantage (Again!)

Remember, each "pipe" in the grid can be set to 30 levels.

For learning numbers, this means:
- **2 levels (binary):** "This feature matters" or "doesn't matter"
- **30 levels:** "This feature matters A LOT" / "a little" / "not really" / "definitely not"

More nuance = smarter decisions!

```
Binary brain:               30-level brain:
"7 has a line? YES/NO"      "7 has:
                             - strong horizontal: 28
                             - strong diagonal: 25
                             - weak curve: 3
                             - no loop: 1"
```

---

## The Results: Hardware vs Simulation

**Important distinction:**
- **Ferroelectric CIM HARDWARE:** Dr. Tour achieved **87%** with **88% theoretical maximum**
- **Our SIMULATION:** May show higher because it's idealized (no real chip noise)

For context:
- Random guessing: 10% (1 in 10)
- **Ferroelectric CIM hardware: 87%** (what was actually measured!)
- Theoretical maximum: 88% (Dr. Tour stated this limit)
- Perfect human: ~98%

```
    ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 100%
    ████████████████████████████████████████
    ▲
    │  Random: 10%
    │
    │          Ferroelectric CIM HARDWARE: 87%
    │                    ▼
    █████████████████████████████████░░░░░░░

    │          Theoretical max: 88%
    │                     ▼
    ██████████████████████████████████░░░░░░

    (Our simulation may exceed this because it's idealized)
```

> ⚠️ **Why simulation can exceed hardware:** Real chips have IR drop, sneak paths, device variation, and ADC/DAC noise that our simulation doesn't fully capture.

---

## What Can Go Wrong (And How We Fix It)

### Problem: Messy Handwriting

Some people write weird numbers!

```
    Clear 7:     Weird 7:      Cursive 7:

     ───          ~∼~           ⌐¬
       \            \           /
        \            \         /
```

**Solution:** Train on LOTS of different handwriting styles (60,000 examples!)

### Problem: Noisy Hardware

The chip's "pipes" aren't perfectly precise:

```
    What we programmed:    What we got:
         Level 15              Level 14 or 16
                               (close but not exact)
```

**Solution:** The chip learns to be robust to small errors during training.

### Problem: Similar-Looking Digits

Some numbers look alike:
- 1 and 7 (both have vertical lines)
- 3 and 8 (both have curves)
- 4 and 9 (both have vertical + loop)

**Solution:** Learn subtle differences through thousands of examples.

---

## The Confusion Matrix: Where the Chip Makes Mistakes

The demo shows a "confusion matrix" — a grid showing what the chip guesses vs. what's correct:

```
                What the chip guessed
                0  1  2  3  4  5  6  7  8  9
           0  [98  0  0  0  0  0  1  0  1  0]
           1  [ 0 99  0  0  0  0  0  1  0  0]
    Real   2  [ 1  0 95  1  1  0  0  1  1  0]
    digit  3  [ 0  0  1 96  0  1  0  0  2  0]
           4  [ 0  1  0  0 97  0  1  0  0  1]
           5  [ 1  0  0  2  0 94  1  0  2  0]
           6  [ 1  0  0  0  1  1 97  0  0  0]
           7  [ 0  1  0  0  0  0  0 97  0  2]
           8  [ 0  0  1  1  0  1  0  0 96  1]
           9  [ 0  0  0  0  2  0  0  1  0 97]

    Diagonal = CORRECT answers (the green squares)
    Off-diagonal = mistakes
```

**Reading the matrix:** Row 5, Column 3 shows "2" — meaning 2 times when it was actually a "5", the chip guessed "3".

---

## What You See in This Demo

When you run the demo:

1. **Draw a digit** — Use your mouse to write a number
2. **See the prediction** — Watch which "expert" wins
3. **Confidence bars** — See how sure the chip is about each digit
4. **Layer activations** — See the "thinking" at each stage
5. **Confusion matrix** — See overall accuracy and common mistakes

---

## Try It Yourself!

```bash
cd module3-mnist
./mnist --interactive
```

**Things to try:**
- Draw a clear "5" — should get high confidence
- Draw a sloppy "7" — see if it still works
- Draw something between "3" and "8" — see which wins
- Type `test` to run random samples from the test set

---

## Summary for Kids

| Concept | Simple Version |
|---------|---------------|
| MNIST | A famous collection of 70,000 handwritten digits |
| Neural Network | A system of "experts" that vote on what they see |
| Training | Showing examples and adjusting weights when wrong |
| Weights | How much each "pipe" in the grid matters |
| Accuracy | How often the chip guesses correctly (87% hardware) |
| Confusion Matrix | A chart showing where mistakes happen |

---

## The Story in Three Sentences

1. **We showed a chip 60,000 handwritten numbers.**
2. **It learned patterns by adjusting its 100,000 internal "pipe widths."**
3. **Now it can recognize new numbers it's never seen — Ferroelectric CIM hardware achieves 87%!**

---

## One Sentence Summary

> **Demo 3 shows a chip that learned to read handwritten numbers by adjusting its internal "pipe grid" after seeing 60,000 examples — Ferroelectric CIM hardware achieves 87% accuracy (88% theoretical max)!**

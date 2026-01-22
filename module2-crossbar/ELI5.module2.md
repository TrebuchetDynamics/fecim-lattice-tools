# Demo 2: The Magic Multiplication Grid - Explained Simply

## What Is This Demo?

Imagine you have a grid of **tiny water pipes**, and each pipe can be adjusted from "barely dripping" to "gushing." When you turn on the faucets at the top, water flows through all the pipes at once, and you collect the total flow at the bottom.

This demo shows how a computer can do **thousands of math problems at the same time** using this "pipe grid" idea!

---

## The Big Idea: Do Math Everywhere At Once

### Normal Computers (One at a Time)

Regular computers are like a single calculator:
```
Problem 1: 5 × 3 = 15  ✓ Done!
Problem 2: 2 × 7 = 14  ✓ Done!
Problem 3: 4 × 6 = 24  ✓ Done!
Problem 4: 8 × 2 = 16  ✓ Done!
...doing them ONE BY ONE (slow!)
```

### Ferroelectric CIM (All at Once)

Ferroelectric CIM is like having thousands of calculators working together:
```
Problem 1: 5 × 3 = 15  ─┐
Problem 2: 2 × 7 = 14   │
Problem 3: 4 × 6 = 24   ├── ALL DONE AT THE SAME TIME!
Problem 4: 8 × 2 = 16   │
...                    ─┘
```

**How?** Physics does the math for free!

---

## The Water Pipe Analogy

Imagine a grid of adjustable pipes:

```
        FAUCETS (Input Numbers)
           │    │    │    │
           5    2    4    8    ← How much water we push in
           ▼    ▼    ▼    ▼
        ┌────┬────┬────┬────┐
     →  │ ▓▓ │ ░░ │ ▓░ │ ░▓ │ → Collect: 15
        ├────┼────┼────┼────┤
     →  │ ░▓ │ ▓░ │ ▓▓ │ ░░ │ → Collect: 22
        ├────┼────┼────┼────┤
     →  │ ▓░ │ ░░ │ ░▓ │ ▓▓ │ → Collect: 19
        └────┴────┴────┴────┘

        ▓▓ = Wide pipe (lots of flow)
        ░░ = Narrow pipe (little flow)
```

**How it works:**
1. Push water into the top (these are your INPUT numbers)
2. Each pipe has a different width (these are your WEIGHT numbers)
3. Water × Pipe Width = Flow through that pipe
4. Collect all the water on each row
5. The totals ARE your answers!

**The magic:** All pipes flow at the same time, so ALL multiplications happen INSTANTLY!

---

## In Real Terms: Electricity Instead of Water

In the real chip, we use:
- **Voltage** instead of water pressure (how hard we push)
- **Conductance** instead of pipe width (how easily current flows)
- **Current** instead of water flow (what comes out)

The physics equation is simple:
```
Current = Voltage × Conductance

I = V × G

(This is Ohm's Law - it's just physics!)
```

---

## Why This Matters for AI

Neural networks (the brains of AI) are basically just **lots and lots of multiplications**:

```
                Input Image (784 pixels)
                      │ │ │ │ │ │
                      ▼ ▼ ▼ ▼ ▼ ▼
        ┌─────────────────────────────┐
        │     784 × 128 = 100,352     │  ← Layer 1
        │       multiplications!       │
        └─────────────────────────────┘
                      │
                      ▼
        ┌─────────────────────────────┐
        │      128 × 10 = 1,280       │  ← Layer 2
        │       multiplications!       │
        └─────────────────────────────┘
                      │
                      ▼
              "That's a 7!"
```

**Normal computer:** Does 101,632 multiplications one by one 😓

**Ferroelectric CIM:** Does them all at once in the grid! ⚡

---

## The 30 Levels Advantage

Remember from Demo 1, each cell can hold 30 different values?

In the pipe analogy:
- Level 0 = Pipe almost closed (tiny drip)
- Level 15 = Pipe half open (medium flow)
- Level 29 = Pipe wide open (maximum flow)

```
Level 0:   │░│  → barely any flow
Level 10:  │▒│  → some flow
Level 20:  │▓│  → good flow
Level 29:  │█│  → maximum flow
```

**Why 30 levels matters:** More precision in the "pipe widths" = more accurate AI!

---

## Problems in the Real World (Non-Idealities)

Real pipes aren't perfect. This demo shows the problems and solutions:

### Problem 1: Voltage Drop (IR Drop)

Like water pressure dropping as you get farther from the pump:

```
PUMP HERE
    ↓
    █ → █ → █ → █ → █
    │    │    │    │
   100%  98%  95%  90%  85%  ← Less pressure at the end!
```

**In the chip:** Cells far from the voltage source see lower voltage.

**Solution:** Make wires thicker, or add boosters along the way.

### Problem 2: Sneak Paths

Like water finding shortcuts through pipes you didn't want it to use:

```
    Wanted path:        Sneaky path:
        ↓                   ↓
      ┌───┐               ┌───┐
      │ █ │               │ █ │←──┐
      └───┘               └───┘   │
        ↓                   ↓     │ Oops!
      ┌───┐               ┌───┐   │
      │ █ │               │ █ │───┘
      └───┘               └───┘
```

**In the chip:** Current flows through cells you didn't select.

**Solution:** Add "one-way valves" (transistors) to each cell.

### Problem 3: Device Variation

Like pipes that are slightly different from each other, even when they should be the same:

```
    What we wanted:     What we got:
      │  │  │            │  │  │
      █  █  █            █  ▓  █   ← Middle one is
      │  │  │            │  │  │      slightly different!
```

**In the chip:** Manufacturing isn't perfect, so cells vary slightly.

**Solution:** Careful programming with "Scheme C" pulses (see Demo 2 README).

---

## What You See in This Demo

When you run the demo, you'll see:

1. **Heat Map** — A colorful grid showing the "pipe widths" (conductances)
2. **Input/Output Vectors** — The numbers going in and coming out
3. **IR Drop View** — Where voltage is getting lost
4. **Sneak Path View** — Where current is taking shortcuts
5. **Accuracy Comparison** — Ideal vs. real results

---

## Try It Yourself!

```bash
cd module2-crossbar
./crossbar-gui
```

**Things to try:**
- Click cells to see their level (0-29)
- Run MVM and watch the multiplication happen
- Toggle IR Drop view — see voltage loss at far corners
- Toggle Sneak Paths — see unwanted current flow
- Increase array size — bigger grid = more multiplications at once!

---

## The Efficiency Win

| Method | Energy per Multiplication |
|--------|--------------------------|
| GPU (moving data around) | ~1000 fJ |
| **Ferroelectric CIM (in-place)** | **~10 fJ** |

**100× more efficient!** That's why data centers could use 80-90% less power.

---

## Summary for Kids

| Concept | Simple Version |
|---------|---------------|
| Crossbar | A grid of adjustable pipes |
| MVM | Push water in, physics does the math, collect the totals |
| Conductance | How wide each pipe is (0-29 levels) |
| IR Drop | Water pressure dropping far from the pump |
| Sneak Paths | Water taking shortcuts through wrong pipes |
| Parallel | All pipes flow at the same time (fast!) |

---

## One Sentence Summary

> **Demo 2 shows how a grid of tiny adjustable "pipes" can do thousands of multiplications at the same time, making AI 100× more energy efficient.**

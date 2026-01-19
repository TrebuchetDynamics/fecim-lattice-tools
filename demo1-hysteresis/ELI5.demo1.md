# Demo 1: The Memory Crystal - Explained Simply

## What Is This Demo?

Imagine you have a **magic light switch** that can be set to 30 different brightness levels, not just ON or OFF. And when you let go of it, it **stays exactly where you left it** — even if you unplug it!

This demo shows how that "magic switch" works inside a computer chip.

---

## The Big Idea: Memory That Remembers Without Power

### Regular Computer Memory (RAM)

Think of regular computer memory like writing on a whiteboard:
- ✅ You can write and erase quickly
- ❌ If you turn off the lights (power), everything disappears!

### IronLattice Memory (Ferroelectric)

Think of IronLattice memory like carving into clay:
- ✅ You can change it when you want
- ✅ It stays even when you walk away (no power needed!)
- ✅ AND you can make 30 different depths of carving, not just "carved" or "not carved"

---

## The Light Switch Analogy

### Normal Light Switch (Binary)
```
     ON ●────────────────
        │
        │    (just 2 states)
        │
    OFF ●────────────────
```

### IronLattice Switch (30 Levels)
```
  Bright ●────────────────  Level 30
         ●────────────────  Level 29
         ●────────────────  Level 28
         ⋮
         ●────────────────  Level 15
         ●────────────────  Level 14
         ⋮
         ●────────────────  Level 1
    Dark ●────────────────  Level 0
```

**Why is this amazing?** Each tiny memory cell can store 30 different values instead of just 2. That's like fitting 5 regular memory cells into 1!

---

## The "Stubborn Magnet" Inside

Inside the memory, there are tiny particles that act like **stubborn magnets**:

```
    Before you push:        After you push:

         N                       S
         │                       │
         ▼                       ▼
       ┌───┐                   ┌───┐
       │ ↑ │  ──PUSH──→        │ ↓ │
       └───┘                   └───┘
         ▲                       ▲
         │                       │
         S                       N

    "I'm pointing UP!"     "Now I'm pointing DOWN!"
    (and I'll STAY this way until you push me again!)
```

These tiny magnets are called **electric dipoles**. When you push them (with voltage), they flip. And they **stay flipped** even after you stop pushing!

---

## Why Does It Make That Loop Shape?

When you slowly push and pull on these stubborn magnets, something interesting happens:

```
                    PUSH HARD →

        "Okay, I flipped!"
              ╭───────╮
             ╱    3    ╲
            │           │
       2   │           │   4
           ●           ●
          ╱             ╲
    1 ───●───────────────●─── 5
          ╲             ╱
           ●           ●
       8   │           │   6
            │           │
             ╲    7    ╱
              ╰───────╯

                    ← PULL HARD
```

**The loop happens because:**
1. Push a little → magnets start to flip
2. Push harder → more magnets flip
3. Push really hard → ALL magnets flipped!
4. Stop pushing → magnets STAY where they are (memory!)
5. Pull back → magnets start flipping the other way
6. Keep pulling → more flip
7. Pull really hard → ALL flipped the other way
8. Stop → they stay again!

**The key insight:** The path going UP is different from the path going DOWN. This is called **hysteresis** (history matters!).

---

## Why 30 Levels Instead of Just 2?

Think of it like a parking garage:

**Binary Memory (2 levels):**
```
┌─────────────────┐
│  ROOF (1)       │  ← Only 2 floors
├─────────────────┤
│  GROUND (0)     │
└─────────────────┘
```

**IronLattice (30 levels):**
```
┌─────────────────┐
│  Floor 30       │
├─────────────────┤
│  Floor 29       │
├─────────────────┤
│  Floor 28       │
├─────────────────┤
│      ⋮          │  ← 30 floors!
├─────────────────┤
│  Floor 2        │
├─────────────────┤
│  Floor 1        │
└─────────────────┘
```

**More floors = more cars parked in the same building footprint!**

In computer terms: 30 levels ≈ 5 bits of information per cell (instead of 1 bit).

---

## Real-World Benefit: Smarter AI, Less Power

**Old Way (GPUs):**
- Data lives in one place (memory)
- Math happens in another place (processor)
- Data has to travel back and forth constantly
- Uses LOTS of electricity (like driving to work every day)

**IronLattice Way:**
- Data AND math happen in the SAME place
- No traveling needed
- Uses very little electricity (like working from home!)

```
Old Way:                          IronLattice:
┌────────┐      ┌────────┐       ┌─────────────────┐
│ Memory │ ←──→ │  CPU   │       │ Memory + Math   │
└────────┘      └────────┘       │   ALL IN ONE!   │
    ↑               ↑            └─────────────────┘
    │               │                    ↑
  Traffic!      Waiting!            No traffic!
```

---

## What You See in This Demo

When you run the demo, you'll see:

1. **The Loop** — Watch the "stubborn magnets" trace their path as you change the voltage
2. **30 Levels** — See which "parking floor" you're on
3. **Different Materials** — Try different flavors of the magic crystal
4. **Live Updates** — Drag the slider and watch physics happen!

---

## Try It Yourself!

```bash
cd demo1-hysteresis
./hysteresis
```

**Things to try:**
- Drag the voltage slider back and forth — watch the loop form!
- Stop halfway — see how the level "remembers" where you stopped
- Try different materials — some have "stickier" magnets than others

---

## Summary for Kids

| Concept | Simple Version |
|---------|---------------|
| Ferroelectric | A material with stubborn magnets inside |
| Hysteresis | The magnets remember which way you pushed them |
| 30 Levels | Like a 30-floor parking garage for data |
| Non-volatile | Remembers even when unplugged (like a carved rock) |
| Compute-in-Memory | Do math where the data lives (no commute!) |

---

## One Sentence Summary

> **Demo 1 shows how a special crystal can remember 30 different states without power, like a magic dimmer switch that never forgets where you left it.**

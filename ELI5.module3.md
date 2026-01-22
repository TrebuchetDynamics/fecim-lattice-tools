# Demo 3: MNIST Neural Network - Explain Like I'm 5

## What Is This?

Imagine you're teaching a robot to recognize handwritten numbers (0-9). You show it thousands of pictures of numbers written by different people, and eventually it learns to read them!

**Demo 3 shows this happening in real-time using FeCIM technology.**

## The Big Idea

```
You draw a number  -->  The chip "thinks"  -->  It tells you what number it sees
     ✏️ "3"              ⚡ (super fast!)            🎯 "That's a 3!"
```

## How It Works (Simple Version)

### Step 1: You Draw a Number
- The canvas is 28x28 tiny squares (pixels)
- When you draw, you're filling in squares with different shades of gray
- The computer sees 784 numbers (28 × 28 = 784)

### Step 2: The Chip Does Math
```
784 inputs  →  128 "helper" neurons  →  10 outputs (one for each digit 0-9)
   📥              🧠                      📊
```

The chip does **101,632 multiplications** almost instantly!

### Step 3: It Makes a Guess
The output with the highest score wins. If output #7 is highest, the chip says "That's a 7!"

## Why Is FeCIM Special?

### Traditional Computer (Like Your Laptop)
```
Memory (RAM) ←──── moves data back and forth ────→ Processor (CPU)
                   🐌 This is slow and uses lots of energy!
```

### FeCIM Chip
```
Memory + Processor = SAME CHIP
        ⚡ No moving data around!
        ⚡ 1000x less energy!
        ⚡ Super fast!
```

**Dr. Tour says:** *"Compute in memory where the same device does the memory and the computation."*

## The 30 Levels Magic

Normal computer memory stores just 0 or 1 (binary).

FeCIM stores **30 different levels** in each tiny cell!

```
Binary:    0 -------- 1           (2 options = 1 bit)
FeCIM:     0 - 5 - 10 - 15 - 20 - 25 - 30   (30 options = 4.9 bits!)
```

This means the same chip can store **5x more information!**

## What Do the Colors Mean?

### Drawing Canvas
- **Black** = Empty (value: 0)
- **Cyan/White** = Filled in (value: up to 1.0)

### Network Activations
- **Input Layer**: Your drawing converted to numbers
- **Hidden Layer (orange)**: The chip finding patterns
- **Output Layer (bars)**: Scores for each digit 0-9
  - **Green bar** = The winning guess
  - **Gray bars** = Other possibilities

### Prediction Box
- **Green border** = High confidence (>90%)
- **Yellow border** = Medium confidence (70-90%)
- **Red border** = Low confidence (<70%)

## The Target: 87% Accuracy

**Dr. Tour says:** *"We're at 87% validation here."*

This means if you test 100 handwritten numbers, the chip gets about 87 right!

Why not 100%? Some handwritten numbers are really messy - even humans get confused!

```
Is this a 1 or a 7?   →   ⟋
Is this a 4 or a 9?   →   ꝇ
Is this a 5 or a 6?   →   Ƨ
```

## Try It Yourself!

1. **Draw a digit** - Click and drag on the canvas
2. **Watch the network** - See the layers light up
3. **Check the prediction** - Did it guess right?
4. **Click "Random Test"** - Load a real MNIST sample
5. **Click "Evaluate All"** - Test on 1000 images at once

## Fun Facts

| What | How Much |
|------|----------|
| Input pixels | 784 (28×28) |
| Hidden neurons | 128 |
| Output classes | 10 (digits 0-9) |
| Multiplications per guess | 101,632 |
| Clock cycles needed | Just 2! |
| Analog levels per cell | 30 |
| Energy savings vs traditional | 1000x less |

## The Magic Formula

Each connection in the network does this:
```
Output Current = Conductance × Input Voltage
      I        =      G       ×      V
```

All 101,632 multiplications happen **at the same time** because physics does the math for us!

## Summary

Demo 3 shows that FeCIM can:
1. ✅ Recognize handwritten digits
2. ✅ Do it with 87% accuracy
3. ✅ Use almost no energy
4. ✅ Work incredibly fast
5. ✅ Store more data in the same space (30 levels!)

**This is the future of AI computing - doing smart things without burning through energy!**

---

*"This could lower the requirements in a data center by 80 to 90% of the energy requirements."* — Dr. external research group

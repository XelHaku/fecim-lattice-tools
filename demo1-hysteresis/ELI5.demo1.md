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

## The "Stubborn Magnets" Inside

Inside the memory, there are **millions of tiny switches** that act like stubborn magnets — each one slightly different!

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

These tiny switches are called **hysterons** (from the Greek word for "lag behind"). Each one:
- Flips UP at one voltage (say, +1.2V)
- Flips DOWN at a DIFFERENT voltage (say, -0.8V)
- **Stays put** in between!

**The key:** Each hysteron has slightly different flip voltages. When you add up millions of them, you get the smooth loop shape!

---

## Why Does It Make That Loop Shape?

When you slowly push and pull on these stubborn switches, something interesting happens:

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

**The loop EMERGES because each hysteron flips at different voltages:**
1. Push a little → the "easy" hysterons start to flip (low threshold)
2. Push harder → more hysterons flip (medium threshold)
3. Push really hard → even the "stubborn" ones flip (high threshold)
4. Stop pushing → all hysterons STAY where they are (memory!)
5. Pull back → they DON'T flip immediately (different threshold going down!)
6. Keep pulling → now they start flipping the other way
7. Pull really hard → ALL flipped the other way
8. Stop → they stay again!

**The key insight:** Each hysteron has a GAP between its "flip up" and "flip down" voltage. This gap creates hysteresis!

```
One hysteron example:
         Flip UP at +1.2V
              │
    ──────────┼──────────────  E
              │         │
              │    Flip DOWN at -0.8V
              │         │
    [───GAP───]  ← In this gap, it REMEMBERS its state!
```

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

## Write vs Read: The Key to Non-Volatile Memory

### Writing = Pushing HARD

To **write** (change) the memory, you need to push past the "sticky point" (called Ec, the coercive field):

```
    Push HARD (past Ec)
           ↓
    ━━━━━━━━━━━━━━━━━━━
    │                 │
    │  ↑ FLIPS! ↑     │  ← Memory changes!
    │                 │
    ━━━━━━━━━━━━━━━━━━━
```

### Reading = Pushing GENTLY

To **read** (sense) the memory without changing it, use a small push (below Ec):

```
    Push gently (below Ec)
           ↓
    ━━━━━━━━━━━━━━━━━━━
    │                 │
    │  → stays →      │  ← Memory unchanged!
    │                 │
    ━━━━━━━━━━━━━━━━━━━
```

### The Magic

```
WRITE: |Voltage| > Ec  →  Memory CHANGES
READ:  |Voltage| < Ec  →  Memory STAYS (just sense it!)
```

This is why ferroelectric memory is special — you can read it without destroying the data!

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

1. **The Memory Cell** — A colored square showing the current level (1-30), blue to red gradient
2. **The P-E Loop** — Watch the "stubborn magnets" trace their path as you change the voltage
3. **30 Levels Bar** — See which "parking floor" you're on
4. **WRITE/READ Indicator** — Shows when voltage is high enough to change memory (WRITE) vs just sensing (READ)
5. **Educational Panel** — Explains what's happening in each mode
6. **Memory Log** — Watch the read/write operations scroll by in real-time!

---

## Try It Yourself!

```bash
cd demo1-hysteresis
go run ./cmd/demo
```

**Things to try:**

### 1. Sine Wave Mode (Default)
- Watch the loop form automatically
- See how P "lags behind" E — that's the memory!

### 2. Random Walk Mode
- Select "Random Walk" from the waveform dropdown
- Watch it pick random levels and ramp to them
- This shows "store this, store that" — real memory operation!

### 3. Write/Read Demo Mode (Best for understanding!)
- Select "Write/Read Demo" from the dropdown
- Watch the 4-phase cycle:
  - **WRITE**: Voltage goes HIGH (past Ec) → level changes
  - **HOLD**: Voltage returns to ZERO → level STAYS! (memory!)
  - **READ**: Small voltage pulse (below Ec) → level unchanged
  - **DISPLAY**: Shows what was written vs what was read
- The Memory Log shows each operation!

### 4. Manual Mode
- Select "Manual" and drag the slider yourself
- Stop halfway — see how the level "remembers" where you stopped
- Try different materials — some have "stickier" magnets than others

### 5. Frequency Slider
- Speed up or slow down ANY mode with the frequency slider
- Slow = easier to see what's happening
- Fast = more dramatic!

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

---

## Technical Note: What's Actually Running

For the curious, here's what the demo actually computes:

| What you see | What's really happening |
|--------------|------------------------|
| The loop shape | ~450 hysterons, each with different thresholds, summed together |
| The smooth curve | Hysterons distributed as a 2D Gaussian around ±Ec |
| The 30 levels | Simple formula: `Level = round((P/Ps + 1) × 14.5)` |
| Memory effect | Each hysteron stays put between its thresholds |
| WRITE/READ indicator | Compares `|E|` vs `Ec` in real-time |
| Memory Log | Tracks phase transitions in Write/Read Demo mode |

The physics is real — the loop is **emergent**, not drawn!

### New Waveform Modes

| Mode | What it demonstrates |
|------|---------------------|
| Random Walk | Multi-level storage: picks random levels, ramps E to reach them |
| Write/Read Demo | Full memory cycle: WRITE → HOLD → READ → verify |

Both modes respond to the **Frequency slider** — speed them up or slow them down!

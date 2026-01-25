# MNIST on FeCIM Explained Like I'm 5

**Understanding Neural Networks, Digit Recognition, and Compute-in-Memory**

---

## Part 1: What is MNIST? (The Picture Problem)

### The Simplest Explanation

Imagine you have a stack of flashcards with handwritten numbers:

```
┌───────────┐  ┌───────────┐  ┌───────────┐
│    ╭─╮    │  │     ╱     │  │    ╭─╮    │
│   │   │   │  │    ╱      │  │   ╱   ╲   │
│   │   │   │  │   ╱       │  │   ──┬──   │
│   │   │   │  │  ╱        │  │     │     │
│    ╰─╯    │  │ ╱         │  │    ─┴─    │
│           │  │           │  │           │
│  "zero"   │  │  "one"    │  │  "two"    │
└───────────┘  └───────────┘  └───────────┘
```

**MNIST** is a famous collection of 70,000 such flashcards that computers use to learn how to read handwritten digits.

---

## Part 2: How Do We Turn Pictures into Numbers?

### The 28×28 Grid

Each flashcard image is a 28×28 grid of tiny squares (pixels):

```
    Each pixel has a brightness (0=white to 255=black)

        0   1   2   3  ...  27
      ┌───┬───┬───┬───┬───┬───┐
    0 │ 0 │ 0 │ 0 │ 0 │...│ 0 │
      ├───┼───┼───┼───┼───┼───┤
    1 │ 0 │ 0 │50 │200│...│ 0 │
      ├───┼───┼───┼───┼───┼───┤
    2 │ 0 │100│255│255│...│ 0 │
      ├───┼───┼───┼───┼───┼───┤
   ...│...│...│...│...│...│...│
      ├───┼───┼───┼───┼───┼───┤
   27 │ 0 │ 0 │ 0 │ 0 │...│ 0 │
      └───┴───┴───┴───┴───┴───┘

    28 × 28 = 784 pixels total
```

### Flattening the Image

The computer doesn't see a square—it sees a long list of 784 numbers:

```
┌───────────────────────────────────────────────────┐
│  0, 0, 0, 50, 200, 0, 0, 100, 255, 255, ... (784) │
└───────────────────────────────────────────────────┘
```

This is the **input** to our neural network!

---

## Part 3: What is a Neural Network? (The Brain Analogy)

### Like a Very Simple Brain

Your brain has neurons that:
1. **Receive signals** from your eyes
2. **Process them** (add up, decide if important)
3. **Send signals** to other neurons

A neural network does the same with numbers!

### The "Voting System" Expert Analogy

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

### Our Simple Network

```
                     INPUT                  HIDDEN                  OUTPUT
                   (784 pixels)          (128 neurons)          (10 neurons)

                 ┌─────────────┐         ┌───────────┐         ┌───────────┐
   Pixel 0   ───▶│             │────────▶│ Neuron 0  │────────▶│ "is it 0?"│
                 │             │         ├───────────┤         ├───────────┤
   Pixel 1   ───▶│             │────────▶│ Neuron 1  │────────▶│ "is it 1?"│
                 │  MULTIPLY   │         ├───────────┤         ├───────────┤
   Pixel 2   ───▶│     +       │────────▶│ Neuron 2  │────────▶│ "is it 2?"│
                 │    ADD      │         ├───────────┤         ├───────────┤
      ⋮          │             │         │    ⋮      │         │    ⋮      │
                 │             │         ├───────────┤         ├───────────┤
   Pixel 783 ───▶│             │────────▶│ Neuron 127│────────▶│ "is it 9?"│
                 └─────────────┘         └───────────┘         └───────────┘

                    Layer 1                 Layer 2                Prediction!
               (784 × 128 weights)       (128 × 10 weights)
```

### What Happens at Each Neuron?

Each neuron does three things:

1. **Multiply** each input by a "weight" (importance)
2. **Add** all the products together
3. **Decide** whether to "fire" (pass signal forward)

```
Example: Hidden Neuron 0

    Input         ×    Weight     =    Product
   ─────────────────────────────────────────────
   Pixel 0: 0.5   ×    0.3        =    0.15
   Pixel 1: 0.8   ×   -0.2        =   -0.16
   Pixel 2: 0.0   ×    0.7        =    0.00
      ⋮
   Pixel 783:0.1  ×    0.1        =    0.01
   ─────────────────────────────────────────────
                         SUM      =    2.54

   If SUM > 0, output 2.54 (ReLU activation)
   If SUM < 0, output 0
```

---

## Part 4: The Magic of Weights (The Adjustment Knobs)

### Weights are the "Knowledge"

The network has over 100,000 weights (784×128 + 128×10).

**Before training:** Random weights = garbage predictions
**After training:** Tuned weights = accurate predictions

```
Before Training:            After Training:

  "3" → "8" (wrong!)          "3" → "3" (correct!)
  "7" → "2" (wrong!)          "7" → "7" (correct!)
  "5" → "0" (wrong!)          "5" → "5" (correct!)
```

### Teaching the Chip to Read (Step-by-Step Learning)

#### Step 1: Start Dumb

At first, the chip's "experts" are just guessing randomly:
```
Show it a "5" → Chip says "3" (wrong!)
Show it a "2" → Chip says "8" (wrong!)
Show it a "7" → Chip says "1" (wrong!)
```

#### Step 2: Adjust the Weights

When the chip is wrong, we adjust its internal settings:
```
"You said '3' but it was '5'!"
→ Turn DOWN the pipes that voted for '3'
→ Turn UP the pipes that should have voted for '5'
```

#### Step 3: Repeat 60,000 Times

After seeing 60,000 handwritten digits:
```
Show it a "5" → Chip says "5" (correct!)
Show it a "2" → Chip says "2" (correct!)
Show it a "7" → Chip says "7" (correct!)
```

**The chip learned!** 🎉

### Training = Finding Good Weights

The network is shown thousands of examples and adjusts weights to get better:

```
┌─────────────────────────────────────────────────────────────┐
│  TRAINING LOOP (repeat 60,000 times)                         │
├─────────────────────────────────────────────────────────────┤
│  1. Show image (e.g., "3")                                   │
│  2. Network guesses (e.g., "8")                              │
│  3. Compare to correct answer → Calculate error              │
│  4. Adjust weights slightly to reduce error                  │
│  5. Repeat!                                                  │
└─────────────────────────────────────────────────────────────┘

After 60,000 examples → weights are "trained" → 97% accuracy!
```

---

## Part 5: Why 30 Levels Matter (The Dimmer Switch)

### Regular Computers: Binary Weights

In a normal computer, weights are stored with 32 bits (billions of possible values).

```
Weight = 0.314159265358979...
       (very precise!)
```

### FeCIM: 30 Discrete Levels

In ferroelectric memory, each cell can only store 30 different values:

```
Level 0:  ─────────────────────●──────────────────
Level 1:  ───────────────────●────────────────────
Level 2:  ─────────────────●──────────────────────
   ⋮
Level 15: ●───────────────────────────────────────  (middle)
   ⋮
Level 28: ─────────────────────────────────────●──
Level 29: ───────────────────────────────────────● (maximum)
```

### The Trade-Off

| Type | Precision | Storage | Energy |
|------|-----------|---------|--------|
| Regular (32-bit) | Excellent | 32 bits/weight | High |
| FeCIM (30 levels) | Good enough | ~5 bits/weight | **10,000× less!** |

**Key insight:** 30 levels is ENOUGH for MNIST! We only lose ~1-2% accuracy.

---

## Part 6: FP vs CIM (The Dual-Mode Demo)

### What the Demo Shows

Our demo runs EVERY image through TWO paths simultaneously:

```
                         Same Image
                            ↓
          ┌─────────────────┼─────────────────┐
          │                 │                 │
          ▼                 │                 ▼
    ┌───────────┐           │          ┌───────────┐
    │    FP     │           │          │    CIM    │
    │  (Ideal)  │           │          │ (Realistic)│
    ├───────────┤           │          ├───────────┤
    │ Perfect   │           │          │ 30 levels │
    │ precision │           │          │ + noise   │
    │ No errors │           │          │ + ADC/DAC │
    └─────┬─────┘           │          └─────┬─────┘
          │                 │                │
          ▼                 │                ▼
    Prediction: 3           │          Prediction: 3
    Confidence: 99%         │          Confidence: 87%
          │                 │                │
          └─────────────────┼────────────────┘
                            │
                            ▼
                      ┌───────────┐
                      │  COMPARE  │
                      │ Do they   │
                      │  agree?   │
                      └───────────┘
```

### When Do They Disagree?

```
Easy digit (clear "5"):        Hard digit (messy "8" or "3"?):

FP:  "5" (99% sure)            FP:  "8" (52% sure)
CIM: "5" (95% sure)            CIM: "3" (48% sure)    ← DISAGREE!
     ✓ Agree                        ✗ Disagree
```

The demo lets you see WHEN and WHY the hardware makes mistakes!

---

## Part 7: The Control Sliders (What You Can Adjust)

### Number of Levels (Quantization)

```
30 levels: ████████████████████████████████████░░░░░░  (97% accuracy)
16 levels: ████████████████████████████████░░░░░░░░░░  (95% accuracy)
 8 levels: ████████████████████████████░░░░░░░░░░░░░░  (90% accuracy)
 4 levels: ██████████████████████░░░░░░░░░░░░░░░░░░░░  (80% accuracy)
 2 levels: ██████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░  (65% accuracy)
```

**More levels = more precision = better accuracy (but harder to build!)**

### Noise Level

Real hardware has random noise. The slider simulates this:

```
0% noise:  Weight = 0.50                    → Very accurate
1% noise:  Weight = 0.50 ± 0.005            → Slightly fuzzy
5% noise:  Weight = 0.50 ± 0.025            → Noticeable errors
10% noise: Weight = 0.50 ± 0.050            → Many mistakes
```

### ADC/DAC Bits

**DAC:** Converts digital input to analog voltage (how precise is input?)
**ADC:** Converts analog current to digital output (how precise is output?)

```
8-bit ADC: 256 levels → Very precise output
6-bit ADC: 64 levels  → Good enough (saves 4× energy!)
4-bit ADC: 16 levels  → Loses some accuracy
```

---

## Part 8: Why Compute-in-Memory? (The Superpower)

### The Problem with Regular Computers

```
Normal Computer Architecture:

   ┌──────────────┐     ┌──────────────┐
   │              │     │              │
   │     CPU      │◄───▶│    Memory    │
   │  (computes)  │     │  (stores)    │
   │              │     │              │
   └──────────────┘     └──────────────┘
         ▲                    ▲
         │                    │
         └────────┬───────────┘
                  │
         Data must travel!
         (slow, uses energy)
```

For neural networks: millions of weights must be moved constantly!

### The FeCIM Solution

```
FeCIM Architecture:

   ┌──────────────────────────────────────┐
   │                                      │
   │     Memory = Computer                │
   │     (stores AND computes)            │
   │                                      │
   │   Weights stay put! ← KEY ADVANTAGE  │
   │   Input flows through as voltage     │
   │   Physics does the multiplication!   │
   │                                      │
   └──────────────────────────────────────┘

   Data doesn't move → 10,000× energy savings!
```

### The Matrix Multiplication Trick

Remember all those multiply-and-add operations?

```
Normal: Do one at a time (billions of steps)

CIM: Do ALL AT ONCE using physics!
     - Input voltages on rows
     - Weights as conductances
     - Output currents = multiplied sum!

     I = G × V  (Ohm's Law does the work)
```

---

## Part 9: The Demo Walkthrough

### What You See

```
┌────────────────────────────────────────────────────────────────┐
│  FeCIM MNIST Demo                                               │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌───────────────────────────────────────┐  │
│  │             │    │  FP:  Prediction: 3   Confidence: 98% │  │
│  │    Draw     │    │  CIM: Prediction: 3   Confidence: 92% │  │
│  │    Here     │    │                                       │  │
│  │   (28×28)   │    │  ✓ AGREE                              │  │
│  │             │    └───────────────────────────────────────┘  │
│  └─────────────┘                                               │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │  SLIDERS                                                   │ │
│  │                                                            │ │
│  │  Levels:  [====●================] 30                       │ │
│  │  Noise:   [●====================] 1%                       │ │
│  │  ADC:     [============●========] 8 bits                   │ │
│  │  DAC:     [============●========] 8 bits                   │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │  Test Accuracy: FP=97.2%  CIM=95.1%  Agree=98.5%         │ │
│  └───────────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────┘
```

### Interactive Experiments

**Experiment 1: Reduce levels from 30 to 2**
- Watch accuracy drop dramatically
- See FP and CIM disagree more often
- Understand why 30 levels matters!

**Experiment 2: Increase noise to 20%**
- Watch CIM predictions become random
- FP stays perfect (no noise path)
- Understand hardware challenges!

**Experiment 3: Draw your own digits**
- See real-time inference
- Compare your handwriting to training data
- Understand when networks struggle!

---

## Part 10: Key Equations Made Simple

### The Forward Pass

```
Step 1: Multiply input × weights

        pixel[0] × weight[0,0] = product[0]
        pixel[1] × weight[0,1] = product[1]
        ...
        pixel[783] × weight[0,783] = product[783]

Step 2: Add all products + bias

        sum = product[0] + product[1] + ... + product[783] + bias

Step 3: Apply ReLU (if negative, set to 0)

        output = max(0, sum)

Step 4: Repeat for all 128 hidden neurons

Step 5: Repeat for output layer (10 neurons)

Step 6: Apply softmax (convert to probabilities that sum to 1)
```

### Quantization

```
Original weight: -0.35 (any value)
                     ↓
Quantize to 30 levels:
                     ↓
Step 1: Find range of all weights [-0.5 to +0.5]
Step 2: Divide into 30 bins
Step 3: Round -0.35 to nearest bin → Level 5 → -0.33

Quantized weight: -0.33 (one of 30 values)
```

### Energy Calculation

```
MACs (multiply-accumulate operations):
  Layer 1: 784 × 128 = 100,352 MACs
  Layer 2: 128 × 10  =   1,280 MACs
  Total:              = 101,632 MACs

Energy per MAC: ~50 fJ (femtojoules)
Total energy:   ~5 µJ (microjoules) per inference

Compare to GPU: ~1 mJ (millijoules) per inference
                → FeCIM is 200× more efficient!
```

---

## Part 11: Glossary (Big Words Made Simple)

| Term | Simple Definition |
|------|-------------------|
| **MNIST** | Famous dataset of handwritten digits (70,000 images) |
| **Neural Network** | Math that mimics how brains learn patterns |
| **Neuron** | One computing unit that multiplies, adds, and decides |
| **Weight** | A number that says "how important is this connection?" |
| **Layer** | A group of neurons that process together |
| **Training** | Showing examples and adjusting weights to reduce errors |
| **Inference** | Using trained weights to make predictions |
| **Quantization** | Reducing precision (FP32 → 30 levels) |
| **FP (Full Precision)** | Using all 32 bits (ideal, but expensive) |
| **CIM (Compute-in-Memory)** | Doing math where data is stored |
| **ADC** | Converts analog signal to digital number |
| **DAC** | Converts digital number to analog signal |
| **ReLU** | "If negative, make it zero" function |
| **Softmax** | Turns numbers into probabilities (sum to 100%) |
| **Confidence** | How sure the network is (0-100%) |
| **Accuracy** | Percentage of correct predictions |

---

## Part 12: Learning Path

### Beginner

1. **Run the demo** - Draw digits and see predictions
2. **Play with sliders** - See how settings affect accuracy
3. **Read this document** - Understand the concepts

### Intermediate

1. **Study network.go** - See the dual-mode implementation
2. **Study quantize.go** - Understand quantization math
3. **Read mnist.research.md** - Learn the academic background

### Advanced

1. **Modify the network** - Change hidden size, add layers
2. **Train new weights** - Use the Python training script
3. **Add non-idealities** - Implement sneak paths, IR drop

---

## Part 13: Summary

### The Bottom Line

**FeCIM turns physics into computation.**

1. **MNIST** = 784 pixels → 10 digit classes
2. **Neural network** = millions of multiply-add operations
3. **FeCIM** = stores weights as conductances, computes via Ohm's Law
4. **30 levels** = enough precision for 95%+ accuracy
5. **Energy savings** = 100-10,000× vs. traditional computers

### What the Demo Shows

| Feature | What It Demonstrates |
|---------|---------------------|
| Drawing canvas | Input to neural network |
| FP/CIM comparison | Ideal vs. hardware realistic |
| Sliders | Trade-offs in hardware design |
| Accuracy metrics | System-level performance |
| Energy estimate | Efficiency advantage |

### The Key Insight

> **The memory IS the computer.**

Traditional computers move data to a processor. FeCIM processes data WHERE IT'S STORED. This is why neural network inference on FeCIM uses 10,000× less energy than GPUs.

---

### Quick Reference Card

```
┌─────────────────────────────────────────────────────────────┐
│               MNIST NEURAL NETWORK QUICK REFERENCE           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ARCHITECTURE:                                              │
│     Input:  784 neurons (28×28 pixels)                      │
│     Hidden: 128 neurons (ReLU activation)                   │
│     Output:  10 neurons (softmax probabilities)             │
│                                                             │
│  OPERATIONS PER INFERENCE:                                  │
│     Layer 1: 784 × 128 = 100,352 MACs                       │
│     Layer 2: 128 × 10  =   1,280 MACs                       │
│     Total:             = 101,632 MACs                       │
│                                                             │
│  FeCIM SETTINGS:                                            │
│     Levels: 30 (~5 bits)                                    │
│     ADC: 6-8 bits                                           │
│     DAC: 6-8 bits                                           │
│     Noise: 1-5% (realistic)                                 │
│                                                             │
│  EXPECTED ACCURACY:                                         │
│     FP32:        ~97%                                       │
│     30 levels:   ~95%                                       │
│     + noise:     ~92%                                       │
│     + ADC/DAC:   ~90%                                       │
│                                                             │
│  ENERGY:                                                    │
│     ~5 µJ per inference (FeCIM)                             │
│     ~1 mJ per inference (GPU)                               │
│     → 200× more efficient!                                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## One-Sentence Summary

> **This demo shows a chip that learned to read handwritten numbers by adjusting its internal "pipe grid" after seeing 60,000 examples — Ferroelectric CIM hardware achieves 87% accuracy (88% theoretical max)!**

---

*"The same device does the memory and the computation."* — Dr. external research group

---

## Related Documentation

- [MNIST Demo](mnist.demo.md) - Demo walkthrough and technical details
- [MNIST Research](mnist.research.md) - Academic background and literature review
- [MNIST Open Source](mnist.opensource.md) - Related projects and tools

*This document is part of the FeCIM Visualizer project.*

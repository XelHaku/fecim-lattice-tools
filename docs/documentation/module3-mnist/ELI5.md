<!-- Category: ELI5 | Module: module3-mnist | Reading time: ~4 min -->
# Module 3 ELI5: Neural Network Digit Recognition

> Handwritten digits recognized by a ferroelectric crossbar:
> the weights ARE the conductances.

## The 5-Step Inference Flow

A handwritten digit image goes through five steps to become a predicted number:

```
Step 1: CAPTURE
  A 28x28 pixel image of a handwritten digit (0-9)
  Each pixel is a brightness value from 0.0 (black) to 1.0 (white)
  Total: 784 input values

          .......
         .  ###  .
         . #   # .
         .     # .
         .    #  .
         .   #   .
         .  #### .       <-- this is a "2"
          .......

Step 2: FLATTEN
  Unroll the 28x28 grid into a single row of 784 numbers
  [0.0, 0.0, 0.1, 0.8, 0.9, 0.3, ...]

Step 3: MULTIPLY (this is where the crossbar shines)
  Feed the 784 voltages into the crossbar array
  Each cell's conductance IS a trained weight
  Output currents = matrix-vector multiply = first layer activations

  [784 inputs] --> [crossbar: 784 x 128] --> [128 hidden neurons]

Step 4: ACTIVATE
  Apply ReLU (keep positive values, zero out negatives)
  Then multiply through a second crossbar layer

  [128 hidden] --> [crossbar: 128 x 10] --> [10 output neurons]

Step 5: DECIDE
  Apply softmax to get probabilities for each digit 0-9
  The highest probability wins

  Output: [0.01, 0.02, 0.95, 0.01, ...]
                        ^
                    Prediction: "2" with 95% confidence
```

## What the Simulator Does

The module runs the same network twice, side by side:

- **Full-precision path**: Perfect math, no hardware effects. This is the
  theoretical best the network can do.

- **CIM (Compute-in-Memory) path**: Same network but with quantized weights
  (30 discrete levels instead of continuous), read noise, and other hardware
  effects. This is what a real crossbar would produce.

The gap between the two accuracies shows how hardware constraints affect
recognition quality. With good quantization-aware training, the gap is small.

## Where the 98% Number Comes From

The 98.24% MNIST accuracy cited in the literature comes from a specific HZO
FTJ reservoir computing study (Journal of Alloys and Compounds 2025). That
is an external research result, not a claim from this simulator. This module's
own accuracy depends on the network size, training, and hardware-effect
settings.

## What the Simulator Simplifies

- The network is intentionally small (784-128-10) for interactive speed,
  not state-of-the-art accuracy
- Noise and quantization are modeled in software, not measured from devices
- Training is done offline; the GUI focuses on inference demonstration
- Non-idealities are simplified approximations, not device-calibrated

## Key Takeaways

| Concept | Remember This |
|---------|---------------|
| MVM | The crossbar does the matrix multiply -- physics computes |
| Two paths | Full-precision vs CIM lets you see hardware impact |
| Quantization | Continuous weights rounded to 30 discrete levels |
| Noise | Small random errors per cell model device variation |
| Gap | Accuracy difference = cost of analog hardware constraints |

## Next Steps

- Formal equations --> PHYSICS.md
- Implementation details --> FEATURES.md
- Try drawing digits in the GUI canvas for live inference

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*

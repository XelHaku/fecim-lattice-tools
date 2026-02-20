<!-- Category: ELI5 | Module: module5-comparison | Reading time: ~4 min -->
# Module 5 ELI5: CIM vs Conventional Computing

> How does ferroelectric CIM compare to GPUs and conventional memory?
> The answer starts with where the energy actually goes.

---

## The Data Movement Problem

In conventional computing, most energy is spent **moving data**, not
computing on it.

```
  Traditional Architecture:

  [Memory DRAM]  <--- data bus --->  [Processor CPU/GPU]
                  ^^^^^^^^^^^^^^^^
                  This shuttle costs
                  10-100x more energy
                  than the actual math
```

Every matrix multiply in a neural network requires reading millions
of weights from memory, shipping them to the processor, doing the
math, and shipping the results back.

**Compute-in-Memory eliminates the shuttle.** The weights live
permanently inside the crossbar array, and the math happens right
where the data is stored:

```
  CIM Architecture:

  [Crossbar Array]
     weights stored as conductance
     input voltages applied to rows
     output currents collected from columns
     --> one analog pass = one matrix-vector multiply
```

---

## The Comparison

| Aspect | GPU / CMOS | FeCAP CIM | FTJ CIM |
|--------|-----------|-----------|---------|
| Where math happens | CPU/GPU (separate from memory) | In the crossbar array | In the crossbar array |
| Precision | 32-bit or 16-bit floating point | ~4-bit analog | ~4-bit analog |
| Energy budget | High (dominated by data movement) | Low (in-situ computation) | Low (in-situ computation) |
| Programming complexity | Trivial (flip bits) | ISPP required (iterative) | ISPP required (iterative) |
| Sneak paths | N/A (digital) | Eliminated (capacitive) | Present (0T1R) or suppressed (1T1R) |
| Maturity | Production silicon | Research / early demos | Research / early demos |

---

## What CIM Saves

The energy advantage comes from three things:

1. **No data shuttle.** Weights stay in the array. No DRAM reads,
   no bus transfers, no cache hierarchy.

2. **Analog parallelism.** A single voltage pulse across N rows
   produces N multiply-accumulate results simultaneously. A digital
   multiplier does them one (or a few) at a time.

3. **Ohm's law is free.** Current = Voltage x Conductance is a
   physical law, not a circuit that needs transistors and clock cycles.

---

## Honest Tradeoffs

CIM is not strictly better. The tradeoffs are real:

**Precision:** Analog computation is inherently noisy. A 4-bit
DAC/ADC gives you 16 levels, not the 65,536 levels of 16-bit digital.
Many neural networks tolerate this, but not all workloads can.

**Programming:** Writing a specific conductance level requires ISPP
-- an iterative search that takes multiple pulses. Flipping a digital
bit is instantaneous by comparison.

**Analog noise:** Thermal noise, shot noise, and device-to-device
variation all degrade the signal. The ADC (which consumes 40-60% of
system energy) must be good enough to resolve the signal above the
noise floor.

**Endurance:** Ferroelectric devices degrade after repeated
write/erase cycles. Programming a crossbar millions of times for
training (not just inference) is an open challenge.

---

## What the Simulator Shows

Module 5 lets you compare architectures side by side using modeled
parameters:

- Energy per inference (in microjoules)
- Throughput (TOPS -- tera-operations per second)
- Energy efficiency (TOPS/W)
- Cost projections at data-center scale

All values are **model inputs and model outputs** -- they depend on
the assumptions you configure. They are not measured hardware results.

```
  Example (modeled, not measured):

  Architecture     TOPS/W    Latency
  CPU (x86)         0.5       10 ms
  GPU (A100)        5.0        1 ms
  FeCIM (4-bit)    22.0      0.1 ms   <-- model output, not silicon data
```

---

## What the Simulator Simplifies

- System-level metrics come from simplified analytic models.
- Benchmarks are representative, not exhaustive.
- Results are directional (showing trends), not guarantees.
- No real hardware measurements are included.

---

## Next Steps

- [PHYSICS.md](PHYSICS.md) -- the energy and throughput equations.
- [FEATURES.md](FEATURES.md) -- what the comparison module implements.
- [OPENSOURCE-TOOLS.md](OPENSOURCE-TOOLS.md) -- benchmarking tools.

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*

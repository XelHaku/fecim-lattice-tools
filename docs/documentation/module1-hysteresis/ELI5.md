<!-- Category: ELI5 | Module: module1-hysteresis | Reading time: ~5 min -->
# Module 1 ELI5: Ferroelectric Memory

> Explain Like I'm 5: A ferroelectric cell is like a stubborn light switch --
> it remembers which way you flipped it, even when the power is off.

## The Big Idea

Inside a ferroelectric material, tiny electric dipoles can point "up" or "down."
Think of a ball sitting in one of two valleys. To move the ball from one valley
to the other, you have to push it over the hill in between. Once it rolls into
a valley, it stays there -- even if you stop pushing.

That is exactly how a ferroelectric memory cell stores information. The two
valleys are two stable polarization states. The hill is the energy barrier
(related to the coercive field, Ec). Pushing the ball is applying a voltage.

```
  Energy
    |
    |   .           .
    |  / \         / \
    | /   \       /   \
    |/     \_____/     \
    |   DOWN-state  UP-state
    +----------------------------->
              Polarization
```

The ball (polarization state) will stay in whichever valley you pushed it into,
even after you remove the voltage. That is non-volatile memory.

## The P-E Hysteresis Loop

When you sweep an electric field back and forth across the material, the
polarization traces out a loop. Going up follows a different path than going
down. This asymmetry -- called hysteresis -- is the physical basis of memory.

```
        Polarization (P)
              ^
         +Ps -+--------------.
              |              /
         +Pr -+--.          /
              |   \        /
              |    \      /
    ----------+-----\----/----------> Electric Field (E)
              |      \  /
         -Pr -+-------\/
              |      /
         -Ps -+-----'
              |
            -Ec    0    +Ec

    +Ps = saturation polarization (max charge separation)
    +Pr = remanent polarization (what stays at zero field)
    +Ec = coercive field (field needed to flip)
```

Four key points on this loop:

| Point | Meaning |
|-------|---------|
| +Pr | Polarization that persists after removing field (positive state) |
| -Pr | Polarization that persists (negative state) |
| +Ec | Field strength needed to flip from negative to positive |
| -Ec | Field strength needed to flip from positive to negative |

## What Polarization Means

Inside the crystal, positive and negative ions can shift relative to each other.
When they shift, the material develops a net electric dipole -- that is
polarization. More shift means more polarization.

```
  Before (centrosymmetric):     After (polarized):

       + -                          +----  -
      (together)                   (separated)
```

Polarization (P) measures how much charge separation exists per unit area.
Units: microcoulombs per square centimeter (uC/cm^2).

## The 30-Level Baseline

Instead of storing just two states (0 and 1), a ferroelectric cell can be
programmed to rest at intermediate points on the hysteresis loop. The simulator
uses a 30-level baseline, meaning each cell can hold one of 30 discrete
polarization states.

```
  Level 0  (fully negative, -Ps)
  Level 1
  Level 2
    ...
  Level 15  (near zero polarization)
    ...
  Level 28
  Level 29  (fully positive, +Ps)
```

30 levels give about 4.9 bits per cell (log2(30) = 4.91). Programming to a
specific level uses an Incremental Step Pulse Programming (ISPP) loop: apply a
pulse, check the result, adjust, repeat -- like a binary search converging on
the target.

## What the Simulator Models vs Simplifies

- **Models faithfully**: Hysteresis loop shape, minor loops (partial traversal),
  multi-level ISPP convergence, material parameter differences (HZO, AlScN, etc.)
- **Simplifies**: Switching is treated as instantaneous (quasistatic), no spatial
  domain structure, Preisach distribution uses tanh approximation (not FORC-calibrated)
- **Honest about**: All parameters are simulation baselines from literature
  calibration, not measured data from a specific device

## Next Steps

- Deep physics --> PHYSICS.md
- What the GUI can do --> FEATURES.md
- Material parameter tables --> (see PHYSICS.md material section)

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*

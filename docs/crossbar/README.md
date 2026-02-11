# Crossbar Documentation

Module 2 crossbar array documentation covering physics, architecture, API reference, and educational materials.

## Directory Structure

```
crossbar/
├── reference/           # Technical specifications
│   ├── API.md           # Go package API reference
│   ├── ARCHITECTURE.md  # Module structure and organization
│   ├── ARCHITECTURES.md # 0T1R vs 1T1R vs 2T1R comparison
│   ├── PHYSICS.md       # Physics models implemented in code
│   └── VOLTAGE_RULES.md # Voltage specifications by architecture
│
├── educational/         # Learning materials
│   ├── crossbar.ELI5.md      # Beginner-friendly explanations
│   ├── ../educational/crossbar.demo.md      # Demo guide and usage
│   ├── ../educational/crossbar.physics.md   # Deep physics tutorial
│   ├── ../educational/crossbar.research.md  # Research meta-study (40+ papers)
│   └── ../crossbar/educational/crossbar.opensource.md # External tool survey
│
└── planning/            # Improvement roadmaps
    ├── module2-plan-improvements.md              # Internal enhancement plan
    └── crossbar-proposed-improvements-opensource.md # Ideas from external tools
```

## Quick Links

| I need to... | Read |
|--------------|------|
| Understand the code API | `reference/API.md` |
| Learn crossbar basics | `educational/crossbar.ELI5.md` |
| Compare architectures | `reference/ARCHITECTURES.md` |
| See voltage specifications | `reference/VOLTAGE_RULES.md` |
| Run the demo | `educational/../educational/crossbar.demo.md` |
| Review physics models | `reference/PHYSICS.md` (code) or `educational/../educational/crossbar.physics.md` (theory) |
| See improvement plans | `planning/module2-plan-improvements.md` |

## Physics vs Physics

Two physics documents exist for different purposes:

- **`reference/PHYSICS.md`** - Documents physics *as implemented in code* (conductance models, drift equations, constants)
- **`educational/../educational/crossbar.physics.md`** - Deep technical tutorial on crossbar physics theory (MVM, Ohm's law, KCL)

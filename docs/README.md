# FeCIM Lattice Tools Documentation

**Comprehensive documentation for Ferroelectric Compute-in-Memory simulation and visualization.**

---

## 🎯 Find What You Need

Choose your path based on your goals:

### 🚀 [Getting Started](1-getting-started/README.md)
**→ New users start here!**

- Installation & setup
- Quick start guide
- CLI reference
- Video tutorials
- Troubleshooting

Perfect for: First-time users, installation help, quick demos

---

### 📚 [Learn the Technology](2-learn/README.md)
**→ Understand FeCIM concepts**

- ELI5 overview (explain like I'm 5)
- 7 interactive modules with demos
- Physics fundamentals
- Architecture deep-dives
- Progressive learning paths

Perfect for: Students, educators, technology exploration

---

### 💻 [Develop & Contribute](3-develop/README.md)
**→ Build and extend the tools**

- API reference (all packages)
- Architecture documentation
- Testing guide
- Code quality standards
- Contribution workflow

Perfect for: Developers, contributors, integrators

---

### 🔬 [Research & Validation](4-research/README.md)
**→ Scientific foundations**

- 230+ research papers (23 topics)
- Literature reviews
- Physics validation
- Honesty audit
- Simulation vs reality

Perfect for: Researchers, academics, verification

---

## 📖 Quick Reference

### All 7 Modules Overview

| Module | Topic | What You'll Learn |
|--------|-------|-------------------|
| **[Module 1](2-learn/module1-hysteresis/README.md)** | Hysteresis & Materials | How ferroelectric cells store data |
| **[Module 2](2-learn/module2-crossbar/README.md)** | Crossbar Arrays | How grids compute matrix operations |
| **[Module 3](2-learn/module3-mnist/README.md)** | Neural Networks | How networks recognize patterns |
| **[Module 4](2-learn/module4-circuits/README.md)** | Peripheral Circuits | How support circuits enable operation |
| **[Module 5](2-learn/module5-comparison/README.md)** | Technology Comparison | How FeCIM compares to alternatives |
| **[Module 6](2-learn/module6-eda/README.md)** | EDA & Chip Design | How chips are designed and verified |
| **[Module 7](2-learn/module7-docs/README.md)** | Documentation Tools | How to document and share knowledge |

### Most-Used Documents

| Document | Purpose | Audience |
|----------|---------|----------|
| [Installation Guide](1-getting-started/installation.md) | Get up and running | All users |
| [CLI Reference](1-getting-started/cli-reference.md) | Command-line usage | All users |
| [API Reference](3-develop/api-reference.md) | Package APIs | Developers |
| [Physics Validation](4-research/physics-validation.md) | Scientific accuracy | Researchers |
| [Honesty Audit](4-research/honesty-audit.md) | Claims verification | All users |
| [GLOSSARY](GLOSSARY.md) | Technical terms | All users |

---

## 🎓 Learning Paths

### Path 1: Complete Beginner → Expert (Sequential)

```
1. Read: 2-learn/eli5-overview.md
2. Module 1: Understand hysteresis loops
3. Module 2: See how crossbars compute
4. Module 3: Watch MNIST recognition
5. Module 4: Learn peripheral circuits
6. Module 5: Compare technologies
7. Module 6: Explore chip design
8. Deep dive: 4-research/ papers
```

### Path 2: Developer Onboarding (Fast)

```
1. Install: 1-getting-started/installation.md
2. Run demos: 1-getting-started/runbook.md
3. API docs: 3-develop/api-reference.md
4. Architecture: 3-develop/architecture/
5. Testing: 3-develop/testing/
6. Contribute: 3-develop/code-quality.md
```

### Path 3: Researcher Verification (Focused)

```
1. Status: 4-research/honesty-audit.md
2. Physics: 4-research/physics-validation.md
3. Literature: 4-research/papers/
4. Analysis: 4-research/internal-analysis/
5. Tools: 4-research/opensource-tools/
```

---

## 🏗️ Project Status

- **Phase:** Education & Simulation (TRL 2-3)
- **Purpose:** Explore design space, teach concepts
- **Not:** Hardware validation or production-ready
- **Claims:** See [Honesty Audit](4-research/honesty-audit.md)

### What Works Today

✅ Interactive GUI with 7 modules
✅ Physics-based hysteresis models
✅ Crossbar array simulation
✅ MNIST neural network demo
✅ Peripheral circuit models
✅ EDA workflow visualization
✅ 230+ research papers indexed

### What's Educational/Simulated

⚠️ 30-level quantization (baseline, not verified)
⚠️ Energy projections (physics-based estimates)
⚠️ Performance comparisons (pending silicon)
⚠️ Device parameters (literature ranges)

Full status: See [status.md](../status.md)

---

## 🎯 By Use Case

### "I want to learn about FeCIM technology"
→ Start: [2-learn/eli5-overview.md](2-learn/eli5-overview.md)
→ Then: Work through modules 1-6 sequentially

### "I need to install and run the tools"
→ Start: [1-getting-started/installation.md](1-getting-started/installation.md)
→ Then: [1-getting-started/runbook.md](1-getting-started/runbook.md)

### "I want to contribute code"
→ Start: [3-develop/README.md](3-develop/README.md)
→ Then: [3-develop/api-reference.md](3-develop/api-reference.md)

### "I need to verify scientific accuracy"
→ Start: [4-research/honesty-audit.md](4-research/honesty-audit.md)
→ Then: [4-research/physics-validation.md](4-research/physics-validation.md)

### "I want to understand the research"
→ Start: [4-research/papers/](4-research/papers/)
→ Then: [4-research/literature-review/](4-research/literature-review/)

---

## 📚 Essential Concepts

### The 60-Second Pitch

**Problem:** AI wastes 90% of energy moving data between memory and processors.

**Solution:** Ferroelectric materials can store data AND compute simultaneously.

**How:** Using special materials (HZO), we build memory cells that do matrix multiplication using physics (Ohm's Law). Current = Voltage × Conductance, so the current flowing IS the multiplication result.

**Impact:** Potential for 1000× energy savings for AI inference (pending validation).

**Status:** This is an educational simulator. Real devices are in research phase.

### Key Technologies

- **HZO:** Hafnium-Zirconium-Oxide ferroelectric material
- **1T1R:** One transistor + one resistor architecture
- **MVM:** Matrix-vector multiplication in one step
- **CIM:** Compute-in-Memory (physics does the math)

See [GLOSSARY.md](GLOSSARY.md) for all terms.

---

## 🎬 Demo Videos

Quick visual introductions (see [1-getting-started/demo-videos/](1-getting-started/demo-videos/)):

1. **Hysteresis Loop** (2 min) - How materials remember
2. **Crossbar Array** (3 min) - How grids compute
3. **MNIST Recognition** (2 min) - How networks work
4. **Full Workflow** (5 min) - End-to-end demonstration

---

## 🛠️ Technology Stack

- **Language:** Go 1.24+
- **GUI:** Fyne 2.7.2
- **Build:** Standard Go toolchain
- **Platform:** Linux, macOS, Windows
- **Dependencies:** See [1-getting-started/installation.md](1-getting-started/installation.md)

---

## 📊 Documentation Statistics

- **Total pages:** 150+ markdown files
- **Research papers:** 230+ indexed (23 topics)
- **Code documentation:** 100% package coverage
- **Diagrams:** 50+ Mermaid diagrams
- **Examples:** 30+ runnable examples

---

## 🔗 External Resources

### Scientific Background
- [Nature Communications: Multi-level FeFET](https://doi.org/10.1038/s41467-023-42110-y)
- [J. Alloys & Compounds: FTJ Reservoir](https://doi.org/10.1016/j.jallcom.2025.181869)

### Related Projects
- See [4-research/opensource-tools/](4-research/opensource-tools/)

### Video Transcripts
- See [4-research/transcripts/](4-research/transcripts/)

---

## 🤝 Contributing

We welcome contributions! See:
- [../CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
- [3-develop/code-quality.md](3-develop/code-quality.md) - Code standards
- [3-develop/testing/](3-develop/testing/) - Testing requirements

---

## 📝 Citation & License

### How to Cite
If you use this simulator in research:

```bibtex
@software{fecim_lattice_tools,
  title = {FeCIM Lattice Tools: Educational Ferroelectric CIM Simulator},
  year = {2026},
  url = {https://github.com/[your-repo]},
  note = {Educational simulation tool - not validated hardware}
}
```

### License
See [../LICENSE](../LICENSE) in repository root.

---

## 🆘 Getting Help

### Common Issues
- Build errors: [1-getting-started/runbook.md#common-issues](1-getting-started/runbook.md#common-issues)
- GUI problems: [3-develop/gui/FYNE_NOTES.md](3-develop/gui/FYNE_NOTES.md)
- Physics questions: [4-research/physics-validation.md](4-research/physics-validation.md)

### Ask Questions
- Check [GLOSSARY.md](GLOSSARY.md) first
- Read relevant module documentation
- Search existing issues
- Open new issue with details

---

## 🎯 Quick Navigation

**By Role:**
- Student → [2-learn/](2-learn/)
- Developer → [3-develop/](3-develop/)
- Researcher → [4-research/](4-research/)
- New User → [1-getting-started/](1-getting-started/)

**By Task:**
- Install → [installation.md](1-getting-started/installation.md)
- Learn → [eli5-overview.md](2-learn/eli5-overview.md)
- Build → [api-reference.md](3-develop/api-reference.md)
- Verify → [honesty-audit.md](4-research/honesty-audit.md)

**By Module:**
- Module 1-7 → [2-learn/](2-learn/)

---

**Last Updated:** 2026-02-16
**Version:** 1.0 (reorganized structure)
**Maintainer:** See [../CONTRIBUTING.md](../CONTRIBUTING.md)

# FeCIM Lattice Tools Documentation

**Comprehensive documentation for Ferroelectric Compute-in-Memory simulation and visualization.**

---

## 🎯 Find What You Need

Choose your path based on your goals:

### 🚀 [Getting Started](guides/README.md)
**→ New users start here!**

- Installation & setup
- Quick start guide
- CLI reference
- Video tutorials
- Troubleshooting

Perfect for: First-time users, installation help, quick demos

---

### 📚 [Learn the Technology](modules/README.md)
**→ Understand FeCIM concepts**

- ELI5 overview (explain like I'm 5)
- 7 interactive modules with demos
- Physics fundamentals
- Architecture deep-dives
- Progressive learning paths

Perfect for: Students, educators, technology exploration

---

### 💻 [Develop & Contribute](internals/README.md)
**→ Build and extend the tools**

- API reference (all packages)
- Architecture documentation
- Testing guide
- Code quality standards
- Contribution workflow

Perfect for: Developers, contributors, integrators

---

### 🔬 [Research & Validation](research/README.md)
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
| **[Module 1](modules/hysteresis/README.md)** | Hysteresis & Materials | How ferroelectric cells store data |
| **[Module 2](modules/crossbar/README.md)** | Crossbar Arrays | How grids compute matrix operations |
| **[Module 3](modules/mnist/README.md)** | Neural Networks | How networks recognize patterns |
| **[Module 4](modules/circuits/README.md)** | Peripheral Circuits | How support circuits enable operation |
| **[Module 5](modules/comparison/README.md)** | Technology Comparison | How FeCIM compares to alternatives |
| **[Module 6](modules/eda/README.md)** | EDA & Chip Design | How chips are designed and verified |
| **[Module 7](modules/doc-viewer/README.md)** | Documentation Tools | How to document and share knowledge |

### Most-Used Documents

| Document | Purpose | Audience |
|----------|---------|----------|
| [Installation Guide](guides/installation.md) | Get up and running | All users |
| [CLI Reference](guides/cli-reference.md) | Command-line usage | All users |
| [API Reference](internals/api-reference.md) | Package APIs | Developers |
| [Physics Validation](research/physics-validation.md) | Scientific accuracy | Researchers |
| [Honesty Audit](research/honesty-audit.md) | Claims verification | All users |
| [Trust Boundaries](TRUST.md) | What is validated, educational, planned, or not validated | All users |
| [Citation System](../citations/README.md) | Source records, citable facts, and citation workflow | Researchers, contributors |
| [The Crucible](../tools/crucible/README.md) | Prover/Disprover/Builder validation protocol | Researchers, contributors |
| [GLOSSARY](GLOSSARY.md) | Technical terms | All users |

---

## 🎓 Learning Paths

### Path 1: Complete Beginner → Expert (Sequential)

```
1. Read: modules/eli5-overview.md
2. Module 1: Understand hysteresis loops
3. Module 2: See how crossbars compute
4. Module 3: Watch MNIST recognition
5. Module 4: Learn peripheral circuits
6. Module 5: Compare technologies
7. Module 6: Explore chip design
8. Deep dive: research/ papers
```

### Path 2: Developer Onboarding (Fast)

```
1. Install: guides/installation.md
2. Run demos: guides/runbook.md
3. API docs: internals/api-reference.md
4. Architecture: internals/architecture/
5. Testing: internals/testing/
6. Contribute: internals/code-quality.md
```

### Path 3: Researcher Verification (Focused)

```
1. Status: research/honesty-audit.md
2. Physics: research/physics-validation.md
3. Trust: TRUST.md
4. Crucible: ../tools/crucible/README.md
5. Citations: ../citations/README.md
6. Literature: research/papers/
7. Tools: research/opensource-tools/
```

---

## 🏗️ Project Status

- **Phase:** Education & Simulation (TRL 2-3)
- **Purpose:** Explore design space, teach concepts
- **Not:** Hardware validation or production-ready
- **Claims:** See [Honesty Audit](research/honesty-audit.md)

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
→ Start: [modules/eli5-overview.md](modules/eli5-overview.md)
→ Then: Work through modules 1-6 sequentially

### "I need to install and run the tools"
→ Start: [guides/installation.md](guides/installation.md)
→ Then: [guides/runbook.md](guides/runbook.md)

### "I want to contribute code"
→ Start: [internals/README.md](internals/README.md)
→ Then: [internals/api-reference.md](internals/api-reference.md)

### "I need to verify scientific accuracy"
→ Start: [research/honesty-audit.md](research/honesty-audit.md)
→ Then: [research/physics-validation.md](research/physics-validation.md)

### "I want to understand the research"
→ Start: [research/papers/](research/papers/)
→ Then: [research/literature-review/](research/literature-review/)

---

## 📚 Essential Concepts

### The 60-Second Overview

**Problem:** AI wastes 90% of energy moving data between memory and processors.

**Solution:** Ferroelectric materials can store data AND compute simultaneously.

**How:** Using special materials (HZO), we build memory cells that do matrix multiplication using physics (Ohm's Law). Current = Voltage × Conductance, so the current flowing IS the multiplication result.

**Impact:** Provides a simulator for studying energy and accuracy tradeoffs; any projected gains must be tied to explicit model assumptions and validation evidence.

**Status:** This is an educational simulator. Real devices are in research phase.

### Key Technologies

- **HZO:** Hafnium-Zirconium-Oxide ferroelectric material
- **1T1R:** One transistor + one resistor architecture
- **MVM:** Matrix-vector multiplication in one step
- **CIM:** Compute-in-Memory (physics does the math)

See [GLOSSARY.md](GLOSSARY.md) for all terms.

---

## Visual Walkthroughs

Generated screenshots and videos are not tracked in the public source tree. To create a local documentation frame, run:

```bash
go run ./cmd/fecim-screenshotter-fyne -only docs -out /tmp/fecim-demo-frames
```


---

## 🛠️ Technology Stack

- **Language:** Go 1.25+
- **Build:** Standard Go toolchain
- **Platform:** Linux, macOS, Windows
- **Dependencies:** See [guides/installation.md](guides/installation.md)

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
- See [research/opensource-tools/](research/opensource-tools/)

### Video Transcripts
- See [research/transcripts/](research/transcripts/)

---

## 🤝 Contributing

We welcome contributions! See:
- [../CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
- [internals/code-quality.md](internals/code-quality.md) - Code standards
- [internals/testing/](internals/testing/) - Testing requirements

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
- Build errors: [guides/runbook.md#common-issues](guides/runbook.md#common-issues)
- Physics questions: [research/physics-validation.md](research/physics-validation.md)

### Ask Questions
- Check [GLOSSARY.md](GLOSSARY.md) first
- Read relevant module documentation
- Search existing issues
- Open new issue with details

---

## 🎯 Quick Navigation

**By Role:**
- Student → [modules/](modules/)
- Developer → [internals/](internals/)
- Researcher → [research/](research/)
- New User → [guides/](guides/)

**By Task:**
- Install → [installation.md](guides/installation.md)
- Learn → [eli5-overview.md](modules/eli5-overview.md)
- Build → [api-reference.md](internals/api-reference.md)
- Verify → [honesty-audit.md](research/honesty-audit.md)

**By Module:**
- Module 1-7 → [modules/](modules/)

---

**Last Updated:** 2026-02-16
**Version:** 1.0 (reorganized structure)
**Maintainer:** See [../CONTRIBUTING.md](../CONTRIBUTING.md)

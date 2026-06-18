# Getting Started with FeCIM Lattice Tools

**Quick start guide for new users - from installation to running your first demo.**

---

## 📋 Prerequisites

### System Requirements

- **OS:** Linux, macOS, or Windows
- **Memory:** 4GB RAM minimum, 8GB recommended
- **Disk:** 500MB free space
- **Display:** 1280×720 minimum resolution

### Required Software

- **Go:** Version 1.25 or later ([download](https://go.dev/dl/))
- **Git:** For cloning the repository


See [installation.md](installation.md) for platform-specific instructions.

---

## 🚀 Quick Start (5 Minutes)

### 1. Install Dependencies

Install Go 1.25+ and Git using your platform package manager or the upstream installers.

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install -y git
```

**macOS:** Install Go from [go.dev/dl](https://go.dev/dl/) and Git via Xcode Command Line Tools or your package manager.

**Windows:** Install Go from [go.dev/dl](https://go.dev/dl/) and Git for Windows.

### 2. Clone Repository

```bash
git clone https://github.com/[your-repo]/fecim-lattice-tools.git
cd fecim-lattice-tools
```

### 3. Build

```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
```

### 4. Run

```bash
./fecim-lattice-tools
```

The GUI will launch with 7 interactive modules!

---

## 📚 What's Next?

After installation, explore the documentation:

### For Learners
1. Read [ELI5 Overview](../modules/eli5-overview.md)
2. Run the app with the [Quick Start](../../README.md#quick-start)
3. Try [Module 1: Hysteresis](../modules/hysteresis/)
4. Progress through modules 2-7

### For Developers
1. Read [API Reference](../internals/api-reference.md)
2. Study [Architecture](../internals/architecture/)
3. Review [Testing Guide](../internals/testing/)
4. Check [Code Quality](../internals/code-quality.md)

### For Researchers
1. Review [Honesty Audit](../research/honesty-audit.md)
2. Read [Physics Validation](../research/physics-validation.md)
3. Browse [Research Papers](../research/papers/)
4. Check [Literature Reviews](../research/literature-review/)

---

## 📖 Documentation Map

| Document | Purpose | Time |
|----------|---------|------|
| [installation.md](installation.md) | Detailed install instructions | 10 min |
| [runbook.md](runbook.md) | Operations & troubleshooting | 15 min |
| [cli-reference.md](cli-reference.md) | Command-line options | 5 min |

---

## 🎮 Try the Demos

### Demo 1: Hysteresis Loop (2 minutes)

Visualize how ferroelectric materials remember:

1. Launch app: `./fecim-lattice-tools`
2. Select **Module 1: Hysteresis**
3. Click "AC Mode" to see the P-E loop
4. Try different materials from the dropdown

**What you'll see:** The butterfly-shaped hysteresis curve showing path-dependent memory.

### Demo 2: Crossbar Computation (3 minutes)

Watch matrix multiplication happen in hardware:

1. Select **Module 2: Crossbar**
2. Click "Load MNIST Weights"
3. Apply input vector
4. See output currents compute instantly

**What you'll see:** Real-time MVM with IR drop visualization.

### Demo 3: MNIST Recognition (2 minutes)

Draw a digit and watch the network recognize it:

1. Select **Module 3: MNIST**
2. Draw a digit (0-9) in the canvas
3. Click "Classify"
4. See confidence scores for each digit

**What you'll see:** Neural network inference in action.

---

## 🛠️ Command-Line Usage

### Basic Commands

```bash
# Run GUI (default)
./fecim-lattice-tools

# Enable logging
./fecim-lattice-tools --logger --verbosity debug

# List available materials
./fecim-lattice-tools --list-materials

# Run calibration (headless)
./fecim-lattice-tools --calibrate --material fecim_hzo
```

Full reference: [cli-reference.md](cli-reference.md)

---

## Local Demo Frames

Screenshots and videos are generated artifacts, so they are not tracked in the public source tree. To generate a local documentation frame:

```bash
go run ./cmd/fecim-screenshotter-fyne -only docs -out /tmp/fecim-demo-frames
```

---

## 🐛 Troubleshooting

### Build Errors

**Error:** `go: command not found`
→ **Fix:** Install Go 1.25+ and confirm `go version` works in a new terminal.

**Error:** missing module downloads
→ **Fix:** Run `go mod download`


### Runtime Issues

**Problem:** Black screen on launch
→ **Fix:** Rebuild the default app and rerun from a terminal so startup errors are visible.

**Problem:** Module view does not change
→ **Fix:** Relaunch with an explicit module, such as `./fecim-lattice-tools --module docs`.

**Problem:** Font rendering issues
→ **Fix:** Confirm the OS has standard sans-serif fonts installed and rerun the app.

Full troubleshooting: [runbook.md#common-issues](runbook.md#common-issues)

---

## 💡 Learning Paths

### Path A: Complete Beginner (2 hours)

```
1. Install (this guide) ..................... 10 min
2. Watch demo videos ....................... 15 min
3. Read ELI5 overview ...................... 20 min
4. Try Module 1 ............................ 15 min
5. Try Module 2 ............................ 20 min
6. Try Module 3 ............................ 15 min
7. Explore remaining modules ............... 25 min
```

### Path B: Quick Developer Start (30 minutes)

```
1. Install (this guide) ..................... 10 min
2. Skim runbook ............................. 5 min
3. Read API reference ....................... 10 min
4. Run test suite ........................... 5 min
```

### Path C: Researcher Validation (1 hour)

```
1. Install (this guide) ..................... 10 min
2. Read honesty audit ....................... 10 min
3. Review physics validation ................ 20 min
4. Browse research papers ................... 20 min
```

---

## 📦 What's Included

### Interactive Modules

- **Module 1:** Hysteresis & Materials (8 materials, P-E loops)
- **Module 2:** Crossbar Arrays (MVM, IR drop, sneak paths)
- **Module 3:** MNIST Neural Network (handwriting recognition)
- **Module 4:** Peripheral Circuits (DAC, ADC, TIA, charge pump)
- **Module 5:** Technology Comparison (CPU vs GPU vs FeCIM)
- **Module 6:** EDA Tools (chip design workflow)
- **Module 7:** Documentation Viewer (glossary, search)

### Simulation Features

✅ Physics-based hysteresis (Preisach + Landau-Khalatnikov)
✅ Non-ideal crossbar effects (IR drop, sneak paths, drift)
✅ Dual-mode MNIST (floating-point vs CIM comparison)
✅ Peripheral circuit models (4-bit DAC/ADC baseline)
✅ Energy and timing analysis
✅ Material property explorer
✅ Export to JSON/CSV/SPICE

---

## 🔗 Next Steps

### Explore the Modules

- **[Module 1: Hysteresis](../modules/hysteresis/)** - Start here!
- **[Module 2: Crossbar](../modules/crossbar/)** - See computation
- **[Module 3: MNIST](../modules/mnist/)** - Try recognition
- **[Module 4: Circuits](../modules/circuits/)** - Learn peripherals
- **[Module 5: Comparison](../modules/comparison/)** - Compare tech
- **[Module 6: EDA](../modules/eda/)** - Design chips

### Dig Deeper

- **[Learn Section](../modules/)** - Educational content
- **[Develop Section](../internals/)** - API and architecture
- **[Research Section](../research/)** - Papers and validation

---

## 🤝 Getting Help

### Documentation

- **[GLOSSARY](../GLOSSARY.md)** - Technical terms explained
- **[FAQ](runbook.md#common-issues)** - Common questions
- **[API Docs](../internals/api-reference.md)** - Package reference

### Community

- **Issues:** Report bugs or request features
- **Discussions:** Ask questions, share ideas
- **Contributing:** See [CONTRIBUTING.md](../../CONTRIBUTING.md)

---

## ✅ Installation Checklist

Use this to verify your setup:

- [ ] Go 1.25+ installed (`go version`)
- [ ] GCC installed (`gcc --version`)
- [ ] Repository cloned
- [ ] Dependencies downloaded (`go mod download`)
- [ ] Binary built successfully
- [ ] GUI launches without errors
- [ ] Module 1 demo runs
- [ ] Module 2 demo runs
- [ ] Module 3 demo runs

If all checked, you're ready to explore! 🎉

---

## 📝 Build Options

### Standard Build
```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
```

### Optimized Release Build
```bash
go build -ldflags="-s -w" -o fecim-lattice-tools ./cmd/fecim-lattice-tools
```

### Cross-Compilation
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o fecim-linux ./cmd/fecim-lattice-tools

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o fecim-mac ./cmd/fecim-lattice-tools

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o fecim-mac-arm ./cmd/fecim-lattice-tools

# Windows
GOOS=windows GOARCH=amd64 go build -o fecim.exe ./cmd/fecim-lattice-tools
```

---

## 🎯 Quick Links

**Essential:**
- [Installation Guide](installation.md)
- [Operations Runbook](runbook.md)
- [CLI Reference](cli-reference.md)

**Learning:**
- [ELI5 Overview](../modules/eli5-overview.md)
- [Module Documentation](../modules/)

**Development:**
- [API Reference](../internals/api-reference.md)
- [Architecture](../internals/architecture/)

**Research:**
- [Honesty Audit](../research/honesty-audit.md)
- [Physics Validation](../research/physics-validation.md)

---

**Last Updated:** 2026-02-16
**Status:** All modules functional and tested
**Support:** See [runbook.md](runbook.md) for troubleshooting

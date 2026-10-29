# FeCIM Lattice Tools

**Explore ferroelectric compute-in-memory systems. Keep every assumption inspectable.**

[![CI status](https://github.com/TrebuchetDynamics/fecim-lattice-tools/actions/workflows/ci.yml/badge.svg)](https://github.com/TrebuchetDynamics/fecim-lattice-tools/actions/workflows/ci.yml)
[![Go 1.26.10 or newer](https://img.shields.io/badge/Go-1.26.10%2B-00ADD8?logo=go)](https://go.dev)
[![MIT license](https://img.shields.io/badge/license-MIT-3F8C78)](./LICENSE)

A **simulation-only research and design workbench** for device, array and circuit studies. Go models keep results deterministic, with explicit units, assumptions and provenance.

[First study](#run-your-first-study) · [Choose an interface](#choose-an-interface) · [Modules](#expert-workspaces) · [Scientific limits](#scientific-trust-model) · [Documentation](#documentation)

> [!IMPORTANT]
> Outputs are simulation estimates—not measured silicon, foundry sign-off, product benchmarks or proof of fabricated-device performance. Energy, area and peripheral models are simplified, literature-informed estimates, not calibrated device/layout predictions. Keep every result's model, assumptions, evidence level and source run.

## Choose an interface

- **Go CLI — reproducible studies.** Validate projects, run sweeps, reuse runs and generate reports. Local and file-based.
- **React web candidate — all seven module views and the research workflow.** Hosted runs Go Workers for bounded studies, Hysteresis, Crossbar, Circuits, Comparison and read-only Documentation. MNIST inference and EDA generation require the authenticated local application; their hosted views are read-only help/imports.
- **Legacy Fyne desktop — seven expert workspaces.** Still the default desktop launch. Retained for compatibility.

React is the primary target for new UI work. Fyne is deprecated compatibility, **not removed**. Automated browser checks do not supply real screen-reader evidence: hosted and authenticated local Orca + Firefox/Linux qualification is **NOT RUN**. Real installed EDA tool qualification is also **NOT RUN**. See the [architecture decision](./docs/internals/adr/0005-unified-react-hosted-and-local.md#decision) and [web application guide](./web/README.md).

## Reproducibility demonstration

The checked-in HZO example expands three parameters into **eight immutable runs**. Its recorded output shows six feasible points and one Pareto point:

```console
# First run: evaluate all eight points.
total=8 reused=0 failed=0 feasible=6 pareto=1

# Repeat with a different worker count: reuse the same runs.
total=8 reused=8 failed=0 feasible=6 pareto=1
```

**This demonstrates repeatable execution and reuse—not physics validation, hardware performance or an experimentally established optimum.** Feasibility and Pareto status depend on the example's configured constraints, objectives and simplified models. Worker count does not change SHA-256 run identities or report content. [Run the example below.](#run-your-first-study)

## One study, five explicit stages

**Define → Sweep → Simulate → Analyze → Report**

| Stage | What stays inspectable |
| --- | --- |
| **Define** | Cited YAML inputs, hypothesis, objectives, constraints, physical units and input hashes. |
| **Sweep** | Finite Cartesian points, deterministic seeds, stable run IDs and bounded concurrency. |
| **Simulate** | Device, array, DAC/ADC/TIA, energy, latency, area and SNR model provenance. |
| **Analyze** | Successful, failed, feasible, infeasible, unusable and direction-aware Pareto states. |
| **Report** | Stable JSON, CSV and Markdown, without runtime timestamps in scientific results. |

The [workbench guide](./docs/guides/research-design-workbench.md) explains the bundle schema, supported sweep paths, run reuse and result semantics.

## Run your first study

### Prerequisites

- **Go 1.26.10 or newer**, as required by [`go.mod`](./go.mod).
- A POSIX shell for the Linux/macOS examples below. See [installation](./docs/guides/installation.md) for platform setup.
- No API keys, cloud service, database or daemon for CLI studies.
- The CLI examples use `CGO_ENABLED=0`. Desktop and embedded local browser builds require CGO and [native desktop dependencies](./docs/guides/installation.md#desktop-build-dependencies).

### 1. Copy the reproducible example

```bash
git clone https://github.com/TrebuchetDynamics/fecim-lattice-tools.git
cd fecim-lattice-tools
cp -R examples/research-design-workbench /tmp/fecim-study
```

Use a new destination directory. Generated runs are intentionally not committed into the example directory.

### 2. Validate, run, and regenerate

```bash
CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools project validate /tmp/fecim-study \
  -citation-dir citations/papers

CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools experiment run /tmp/fecim-study \
  -workers 2 -citation-dir citations/papers

CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools report generate /tmp/fecim-study \
  -citation-dir citations/papers
```

`experiment run` already writes reports. `report generate` rebuilds them from committed runs without rerunning models. Repeat `experiment run` with `-workers 4` to inspect reuse.

<details>
<summary>Inspect the generated files</summary>

```text
runs/<sha256>/
├── manifest.json         resolved identity and provenance
├── resolved-design.json  exact evaluated design
└── result.json           metrics, units, models, assumptions, evidence

reports/
├── results.json          structured study output
├── results.csv           stable tabular metrics
└── report.md             readable trade-off report
```

</details>

### 3. Open the desktop workspaces

With native desktop dependencies installed:

```bash
CGO_ENABLED=1 go run ./cmd/fecim-lattice-tools
# Or select a module:
CGO_ENABLED=1 go run ./cmd/fecim-lattice-tools --module hysteresis
CGO_ENABLED=1 go run ./cmd/fecim-lattice-tools --module crossbar
CGO_ENABLED=1 go run ./cmd/fecim-lattice-tools --module eda
```

These commands open **legacy Fyne**, not React. The desktop **Configure starter sweep…** action can create an HZO study without hand-editing YAML.

### Optional React research candidate

Build tools: Node **24.19.0**, npm **11.17.0**, and exact Go **1.26.10** for matching WASM/runtime/native assets.

```bash
make build-browser
mkdir -p "$HOME/fecim-projects"
./fecim-lattice-tools serve --root "$HOME/fecim-projects"
```

Open the short-lived terminal launch link. Do not share it. Then choose **Projects → Configuration → Experiment → Results**. Local operations support explicit revision-checked edits, validation, persistent run/reuse/cancel/reconnect, inspection, 2–4-run comparison, stored-run verification, reports and selected-run export. A plain Go build has no embedded browser assets. Node is not required to use the built binary.

The hosted profile provides Hysteresis, bounded studies, Crossbar, Circuits, Comparison and read-only Documentation. Go Workers own browser computation, without scientific upload, persistence, downloads or localhost probing. Its eight-point example uses 16/32-row arrays to fit the Go-owned 32×32 cap. The original 64-row study remains native-only. MNIST inference and EDA generation require the authenticated local application; their hosted views provide read-only help/imports. Real installed EDA tool qualification remains NOT RUN. See the [web build and protocol guide](./web/README.md) and [research workflow](./docs/guides/research-design-workbench.md#browser-research-workflow-candidate).

`CGO_ENABLED=0` supports headless subcommands and `serve`, not desktop startup. Desktop startup requires CGO and native graphics libraries. No launch-default switch, Fyne removal, deployment or release qualification is implied.

## Expert workspaces

The source-built React candidate provides all seven module views and the shared Projects → Configuration → Experiment → Results workflow. The CLI remains available for reproducible studies; legacy Fyne remains the default desktop. Hosted and authenticated local capabilities differ as described above. This is a runnable candidate, **not a qualified release**.

| Workspace | Inspect and model |
| --- | --- |
| [Hysteresis](./module1-hysteresis) | Preisach and Landau–Khalatnikov behavior, P–E loops, material presets and ISPP writes. |
| [Crossbar](./module2-crossbar) | Matrix-vector multiplication, conductance quantization, IR drop, sneak paths and drift. |
| [MNIST](./module3-mnist) | Example inference behavior under quantization and array constraints. |
| [Circuits](./module4-circuits) | DAC, ADC, TIA, read/program paths, PVT exploration and ISPP control. |
| [Comparison](./module5-comparison) | Evidence-aware technology and architecture comparisons. |
| [EDA](./module6-eda) | SPICE, Verilog, Liberty, DEF, LEF and OpenLane-oriented exports. |
| [Documentation](./module7-docs) | Integrated references, model explanations and educational material. |

<details>
<summary>View the legacy Fyne Hysteresis workspace</summary>

This retained desktop screenshot illustrates the older simulated UI. It is **not** a screenshot of React or evidence of current web parity.

![Legacy Fyne Hysteresis workspace with a simulated HZO polarization-field loop, conductance levels and ISPP controls](./docs/guides/assets/hysteresis_readme.png)

</details>

## Scientific trust model

Results carry one of four evidence labels:

| Evidence | Meaning |
| --- | --- |
| `literature-backed` | Parameter or relationship tied to a reviewed source record. |
| `experiment-calibrated` | Value calibrated against identified experimental input. |
| `simulation-default` | Educational or exploratory default, not a measurement. |
| `derived` | Computed from recorded model inputs and assumptions. |

A label is not blanket validation. External scientific claims belong in the [citation system](./citations/README.md). Check the [honesty audit](./docs/research/honesty-audit.md) and [open source disputes](./citations/disputed.md) for verified, disputed, weak and removed claims.

- Inputs stay inside their project bundle and may be SHA-256 verified before use.
- YAML rejects unknown fields and non-finite physical values.
- Model-domain failures remain results rather than silently disappearing.
- The configurable **30-level conductance baseline is educational**, not a validated device claim.

Use the tools to compare assumptions, test hypotheses and produce reviewable pre-silicon studies—not as a foundry PDK, product benchmark or replacement for device characterization.

## Architecture

```text
Project bundle: project / design / sweep YAML
                    │
                    ▼
Go workbench: load → expand → run → analyze → report
                    │
                    ▼
Immutable runs + JSON / CSV / Markdown

Shared models: device · array · DAC / ADC / TIA · EDA
Interfaces:    CLI · React / Go Workers + local API · legacy Fyne
Checks:        regression · literature · integration
```

The workbench core is local and file-based. It needs no database, remote solver, distributed scheduler, optimizer framework or Rust rewrite. The optional local browser adapter adds an authenticated loopback service; it does not change the default launch. Go retains scientific computation, defaults, values, evidence and identities. See the [architecture](./docs/internals/architecture/ARCHITECTURE.md) and [research workflow guide](./docs/guides/research-design-workbench.md).

## Validation and development

Behavior changes follow **RED → GREEN → REFACTOR**:

```bash
go fmt ./...
go test ./...
make test-legacy-fyne
bash scripts/reproduce_validation.sh
```

Concurrency-sensitive workbench code also runs `go test -race ./workbench/...`. For the full native GUI suite on Linux, use the documented display setup:

```bash
env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m
```

Checks cover physics golden data, literature ranges, crossbar Kirchhoff laws, GUI integration, deterministic reports and CI. Passing checks do not turn a simulation default into measured evidence. Some literature-loop fixtures are provisional or model-generated, not independent experimental validation of the HZO evaluator. The [figure evidence review](./docs/research/validation/landing-page-figure-readiness.md) explains why no measured-versus-simulated headline plot is presented.

## Project map

<details>
<summary>Browse the repository layout</summary>

```text
workbench/             Go projects, experiments, models, reports and local API
cmd/                   desktop, CLI and utility entry points
module1-hysteresis/    ferroelectric behavior and write control
module2-crossbar/      array simulation and non-idealities
module3-mnist/         constrained inference example
module4-circuits/      peripheral circuit models
module5-comparison/    evidence-aware comparisons
module6-eda/           EDA generation and integration
module7-docs/          integrated documentation viewer
shared/                physics, system, peripheral, UI and utility code
validation/            regression, literature and integration checks
citations/             source records and fact registry
examples/              runnable project bundles
web/                   shared React / TypeScript / Vite app; hosted/local builds
```

</details>

## Documentation

- **Use:** [Installation](./docs/guides/installation.md) · [CLI reference](./docs/guides/cli-reference.md) · [Workbench guide](./docs/guides/research-design-workbench.md) · [Web application](./web/README.md)
- **Understand:** [Architecture](./docs/internals/architecture/ARCHITECTURE.md) · [Configuration](./docs/internals/config-reference.md) · [Trust boundaries](./docs/guides/trust.md)
- **Validate:** [Testing](./docs/internals/testing/TESTING.md) · [Citation system](./citations/README.md) · [Honesty audit](./docs/research/honesty-audit.md) · [Source disputes](./citations/disputed.md)
- **Build:** [Contributing](./CONTRIBUTING.md) · [EDA integration](./module6-eda/examples/03-openlane-integration/README.md)

## Citing this software

Use [`CITATION.cff`](./CITATION.cff). Until an archived release exists, record the exact Git commit alongside your project bundle and reports. **No software DOI or JOSS acceptance is claimed.** See the [release readiness checklist](./docs/internals/release/research-software-readiness.md).

## Contributing

Issues and pull requests are welcome for stronger validation, experimental calibration, clearer provenance, reproducible examples and focused usability improvements. Start with [CONTRIBUTING.md](./CONTRIBUTING.md) and keep quantitative claims source-backed.

## License

[MIT](./LICENSE) © Trebuchet Dynamics contributors.

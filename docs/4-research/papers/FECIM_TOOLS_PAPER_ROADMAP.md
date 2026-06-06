# FeCIM Lattice Tools Paper Roadmap

Status: planning guide only. Do not create the `docs/paper/` source tree until the outreach email and demo work are complete.

Date note: this roadmap was saved on April 26, 2026. Older working notes that referenced January 30 and March 14 are stale. If the current outreach cycle is still active, the next Friday is May 1, 2026; otherwise update this file with the actual outreach and submission dates before execution.

## Purpose

Writing a tools paper should:

- establish scientific credibility
- make the tool citable in academic literature
- force methodological clarity
- create an artifact for collaboration outreach
- support the maintainer's academic profile
- provide reproducible documentation
- move the project from "GitHub project" toward "research artifact"

## Target Classification

This should be framed as a tools paper:

- not a novel physics discovery
- not new device characterization
- a software contribution plus validation framework
- aimed at researchers, educators, students, and device/circuit engineers
- consistent with publication patterns in computer science and electrical engineering

Useful precedents to cite fairly:

- CrossSim as a compute-in-memory simulator
- NeuroSim as a benchmarking and architecture framework
- PyTorch as a widely cited software systems paper
- NumPy as a scientific software infrastructure paper

## Target Venues

### Tier 1: arXiv Preprint

- arXiv categories to evaluate: `cs.AR`, `cs.LG`, `physics.app-ph`, and `eess.SY`
- citable and discoverable
- no peer review
- submit after the validation claims and reproducibility story are coherent

### Tier 2: Workshops

- DAC workshop tracks related to AI/ML for EDA or design automation
- ISCA or MICRO workshop tracks related to computer architecture education or emerging memories
- IEDM or device-education adjacent short-format venues if appropriate
- target length: 4-6 pages
- likely timeline: 6-12 months

### Tier 3: Journals

- Journal of Open Source Software (JOSS)
- IEEE Transactions on Education
- IEEE Computer Architecture Letters
- ACM SIGARCH Computer Architecture News
- target length: 8-15 pages
- likely timeline: 12-24 months

## Future Repository Structure

Create this only when the paper work starts:

```text
docs/paper/
  README.md
  PLAN.md
  main.md
  main.tex
  refs.bib
  build.sh
  figures/
    fig01_architecture.png
    fig02_hysteresis_validation.png
    fig03_kcl_conservation.png
    fig04_spice_comparison.png
    fig05_mnist_results.png
    fig06_eda_workflow.png
    source/
  data/
    park2015_hysteresis.csv
    validation_results.json
    README.md
  tables/
    tbl01_module_summary.md
    tbl02_validation_metrics.md
    tbl03_comparison_to_alternatives.md
  reviews/
    self_review_v1.md
    ai_review_v1.md
    peer_review_v1.md
  drafts/
    v0.1_outline.md
    v0.2_first_draft.md
    v1.0_arxiv_submission.md
```

Do not store personal contact details, private reviewer names, or private outreach targets in public paper files. Use placeholders until the submission metadata is intentionally approved.

## Working Title

FeCIM Lattice Tools: An Open-Source Educational Framework for Ferroelectric Compute-in-Memory Simulation with Cross-Tool Validation

Title rules:

- keep it searchable
- include the contribution and method
- avoid marketing language
- stay specific

## Author Metadata

Use placeholders until final approval:

```text
Author Name
Affiliation
Public contact email
Project URL
```

Do not commit private addresses or non-public personal information.

## Paper Structure

Target length: 10-14 pages for an arXiv-style tools paper.

### 1. Abstract

Target: 200-250 words.

Cover:

- context: why ferroelectric compute-in-memory matters
- problem: what current tools do not address
- approach: seven-module educational framework
- validation: literature, conservation-law, and external-tool checks
- results: only numbers backed by reproducible artifacts
- availability: public repository link

### 2. Introduction

Suggested paragraphs:

1. Memory wall motivation.
2. Compute-in-memory as one possible response.
3. Ferroelectric CIM and HfO2/HZO relevance.
4. Educational and research tooling gap.
5. Contributions.
6. Paper organization.

Contribution list:

- an open-source seven-module framework for FeCIM simulation
- cross-validation methodology comparing physics models to published measurements
- EDA integration that generates Verilog, DEF, LEF, Liberty, and SPICE artifacts where supported
- educational visualizations of IR drop, sneak paths, drift, and quantization
- reproducible benchmark and validation scripts
- public documentation and limitation tracking

### 3. Related Work

Subsections:

- FeCIM and analog CIM simulators: CrossSim, NeuroSim, AIHWKit, FerroX, and related tools
- ferroelectric modeling: Preisach, Landau-Khalatnikov, HfO2/HZO device literature
- educational tools and visualization frameworks
- comparison table against alternatives

Rule: cite competing tools fairly. Omitting them weakens credibility.

### 4. Methodology

Subsections:

- system architecture with a seven-module diagram
- hysteresis module: Preisach formulation, Landau-Khalatnikov dynamics, material parameter sources
- crossbar module: Ohm's law, Kirchhoff's Current Law, IR drop, sneak paths, drift
- MNIST pipeline: baseline, quantization, CIM inference path, and reporting boundaries
- peripheral circuits: DAC, ADC, TIA, and sense-chain assumptions
- EDA integration: SPICE, Verilog, Liberty, DEF/LEF, Yosys/OpenLane/OpenROAD position

Example equations to include when validated:

```text
I = G V
V_out = -I_in R_f
V_DAC = V_ref * code / (2^N - 1)
SNR_ADC = 6.02N + 1.76 dB
```

### 5. Validation

Use a three-tier validation strategy:

- literature comparison
- physics conservation laws
- cross-tool comparison

Planned subsections:

- validation methodology and reproducibility statement
- comparison to published hysteresis data, including Park 2015 HZO where provenance is documented
- KCL conservation for Module 2 crossbars, including the 100-case deterministic random-array gate
- SPICE/ngspice comparison when available
- MNIST artifact only when the full training/inference setup and confusion matrix are reproducible

Do not claim fabricated-device validation unless measured device data is actually present and documented.

### 6. Educational Use Cases

Cover:

- graduate course integration
- visual learning for P-E loops, sneak paths, IR drop, and EDA flow
- pre-tape-out exploration boundaries
- collaborator and reviewer demos

### 7. Limitations and Future Work

State limitations directly:

- simulation-only; no new silicon measurements
- simplified peripheral models
- no tape-out validation
- limited validation scale until large-array reports exist
- numeric MNIST or conductance-level claims must be tied to artifacts
- future work: measured device data, larger arrays, temperature/variability sweeps, stronger EDA validation

### 8. Conclusion

Use three concise paragraphs:

- summarize the software contribution
- summarize validation evidence
- state impact and availability without hype

### 9. References

Target: 30-50 references.

Include:

- foundational ferroelectric physics
- HfO2/HZO material papers
- CIM and analog accelerator surveys
- CrossSim, NeuroSim, AIHWKit, FerroX, and comparable tools
- educational or reproducibility methodology papers where useful

## Figure Plan

| Figure | Purpose | Source |
|---|---|---|
| Fig. 1 | Seven-module architecture | generated diagram |
| Fig. 2 | Hysteresis validation overlay | validation/literature artifacts |
| Fig. 3 | KCL conservation residuals | Module 2 KCL report |
| Fig. 4 | SPICE comparison | ngspice comparison artifact |
| Fig. 5 | MNIST results | inference artifact and confusion matrix |
| Fig. 6 | EDA workflow | Module 6 export flow |

Figure captions must be self-contained: describe the data, method, threshold, result, and limitation.

## Writing Timeline

### Week 1: Skeleton, 4-6 hours

- create `docs/paper/` folder structure
- write `docs/paper/PLAN.md`
- draft abstract
- add section headers with one- or two-sentence summaries
- list required figures
- list required references
- commit: `docs(paper): add paper skeleton`

### Week 2: Methodology, 8-10 hours

- write Methodology draft
- add equations
- describe architecture and module boundaries
- draft first architecture figure
- commit: `docs(paper): draft methodology`

### Week 3: Validation Experiments, 10-15 hours

- finish validation folder gaps
- promote Module 2 KCL report into paper data
- generate Park 2015 overlay only from documented provenance
- generate first validation figures
- compute metrics for tables
- commit: `docs(paper): add validation artifacts`

### Week 4: Core Writing, 15-20 hours

- write Introduction
- write Related Work
- write Validation results
- write Educational Use Cases, Limitations, and Conclusion
- commit: `docs(paper): draft main sections`

### Week 5: Polish, 8-10 hours

- self-review
- AI review
- grammar pass
- reference formatting
- figure captions
- convert Markdown to LaTeX if needed
- commit: `docs(paper): complete v1 draft`

### Week 6: Submit, 2-4 hours

- final PDF compile
- arXiv account and category check
- submit preprint if claims, figures, and limitations are ready
- announce through approved public channels
- send follow-up outreach with paper link

## Writing Standards

1. Every claim needs a citation or data artifact.
2. Use numbers with units and, when possible, uncertainty.
3. Avoid marketing language.
4. Hedge appropriately.
5. Define acronyms on first use.
6. Number equations, figures, tables, and sections.
7. Cite competing tools fairly.
8. Write for a skeptical reader.
9. Keep limitations explicit.
10. Do not let the paper outrun what the repository can reproduce.

Examples:

```text
Weak: FeCIM is energy efficient.
Stronger: Prior analog CIM studies report energy ranges of X-Y under Z assumptions [citation]. This tool does not measure energy; it exposes configuration hooks for educational exploration.
```

```text
Weak: Revolutionary new approach.
Stronger: This work extends prior simulators by integrating ferroelectric hysteresis visualization, crossbar non-idealities, and EDA export in one educational toolkit.
```

## Validation Data Process

### Step 1: Digitize Reference Data

Use WebPlotDigitizer or another documented digitization tool.

For each CSV, document:

- source paper and DOI
- figure number
- x-axis and y-axis units
- calibration points
- digitization tool and version
- known uncertainty

### Step 2: Run Simulation

The command must be real before it appears in the paper. Placeholder example:

```bash
go run ./cmd/fecim-lattice-tools \
  --module hysteresis \
  --material HZO \
  --output docs/paper/data/sim_hysteresis_hzo.csv
```

If the CLI does not support the command, either implement it with tests or do not claim it.

### Step 3: Compare

Use scripts under `docs/paper/figures/source/` to:

- load reference and simulation data
- interpolate to a common grid where justified
- compute RMSE, MAE, and relative metrics
- generate plots and residuals
- write machine-readable metrics

### Step 4: Document

Every validation claim should include:

- exact command
- artifact path
- threshold
- result
- limitation

## LaTeX Setup

Start with Markdown for early drafting. Convert to LaTeX only after the structure is stable.

Options:

- Overleaf for fast paper editing and PDF preview
- local TeX Live plus VS Code LaTeX Workshop
- `pandoc` conversion from Markdown to LaTeX
- venue template only after the target is chosen

## Educational Framing

Do claim:

- educational framework
- pedagogical visualization
- teaching tool
- research prototype
- open-source contribution to FeCIM education

Do not claim:

- production-ready
- validated for tape-out
- industry-grade replacement for commercial tools
- new device measurement
- hardware performance not backed by data

## Execution Checklist

Before paper work:

- finish demo bug fixes
- record demo video
- send first outreach email
- avoid splitting focus before that email

When paper work starts:

- create `docs/paper/` folder
- write `docs/paper/PLAN.md`
- draft abstract
- add section headers
- list figures and references
- commit the skeleton

For v1.0:

- methodology complete
- validation figures generated from scripts
- claims mapped to reproducible artifacts
- limitations reviewed
- references formatted
- self-review and AI review saved
- PDF compiles

## Success Criteria

Content:

- specific contributions stated clearly
- related work covers major comparable tools
- methodology has equations and diagrams
- validation shows numerical comparisons
- limitations honestly stated
- reproducibility addressed

Presentation:

- figures have complete captions
- claims have citations or artifacts
- notation is consistent
- acronyms are defined
- BibTeX is clean

Reproducibility:

- one command runs validation
- data sources are documented
- code matches paper claims
- figures regenerate from source scripts
- public repo contains the paper artifacts

Impact goals:

- external reviewer feedback
- course or teaching use
- citations after publication
- issues from real users
- collaborators who can reproduce the results

## Final Reminder

This is next-cycle work. The immediate priority remains the demo and first outreach email. Start the paper only after the email has been sent and the repo claims are aligned with validation artifacts.

# Technical Recovery Roadmap

**Date:** 2026-09-01  
**Status:** Active — release and feature roadmap gate  
**Scope:** Scientific correctness, provenance, serialization, EDA execution, external validation, architecture, operations, and documentation  
**TDD:** N/A — this change modifies planning and documentation only; every future behavior change is separately gated by RED → GREEN → REFACTOR.

## What this is

This roadmap turns the September 2026 repository audit into a sequenced recovery program. It is not a rewrite proposal and it does not discard the project’s validated physics, workbench, or GUI assets. It identifies the trust guarantees the project currently overstates, defines the order in which they must be repaired, and sets evidence-based exit criteria before feature work resumes.

Start here if you are choosing work, reviewing a recovery pull request, or deciding whether a build is suitable for research use. The executable Phase 1 plan is [Critical Correctness and Trust Hardening](../superpowers/plans/2026-09-01-critical-correctness-and-trust-hardening.md).

## Executive recommendation

Treat FeCIM Lattice Tools as a capable research prototype under hardening—not as a release-ready validated workbench. The repository builds and its broad Go test suite passes, but several high-consequence interfaces do not enforce the guarantees their names, comments, tests, or documentation imply.

The recovery policy is:

1. **Pause new release-facing feature work** until Phase 1 exits green.
2. **Repair observable correctness before architecture.** Fix charge noise, cached-run integrity, binary serialization, EDA command construction, and real external validation first.
3. **Make validation semantics honest.** A skipped or no-op external test is not validation. A locally cached file is not immutable merely because it lives under a SHA-named directory.
4. **Deepen existing seams rather than creating parallel systems.** Project loading should guarantee an executable bundle; committed runs should be structurally valid; the default UI and its tests should eventually cross the same state seam.
5. **Use TDD for every production change.** Each recovery item starts with a focused failing test, records RED and GREEN, and ends with the narrow package gate plus the repository gate required by the affected trust surface.

## Evidence summary

| Risk | Live evidence | Current consequence |
|---|---|---|
| Charge-amplifier noise is deterministic and positively biased | `shared/peripherals/charge_amplifier.go:96-110`; no `SenseWithNoise` coverage in `shared/peripherals/charge_amplifier_test.go:10-79` | Repeated samples are identical, zero input acquires positive offset, and noise can exceed output rails after clipping |
| Cached run artifacts are not rebound to identity on load | `workbench/experiment/identity.go:12-33`; `workbench/experiment/store.go:21-49`; `workbench/experiment/runner.go:60-66` | Edited design/result files can be accepted as reused scientific evidence |
| Binary model format accepts ragged and unbounded data | `module2-crossbar/pkg/weights/serialization.go:55-79,190-217,223-363` | Ragged rows shift the stream; untrusted lengths can allocate excessive memory |
| Docker EDA wrappers interpolate into `sh -c` | `module6-eda/pkg/openlane/runner.go:53-81,165-192` | Metacharacters in public arguments can execute in a container with a writable project mount |
| Verilog “sanity” validation performs no validation | `validation/external/eda/verilog_sanity_test.go:9-30`; `.github/workflows/ci.yml:65-117` | CI can report success while no simulator or linter executes |
| `project validate` accepts bundles that `experiment run` rejects | `workbench/project/load.go:65-72`; `workbench/experiment/sweep.go:10-56,59-142`; `cmd/fecim-lattice-tools/workbench_subcommand.go:81-96` | User-visible validate/run semantic mismatch |
| Structurally invalid successful results can be persisted | `workbench/experiment/runner.go:147-170`; `workbench/experiment/analyze.go:50-79` | Duplicate, missing, or unit-invalid metrics become immutable before rejection |
| The default UI and headless UI model use parallel interfaces | `shared/viewmodel/types.go:99-105`; `cmd/fecim-lattice-tools/fyne_app.go:159-201,248-256` | Headless tests can pass without exercising the application users run |
| Launcher implementation is duplicated | `cmd/fecim-lattice-tools/fyne_launcher.go`; `cmd/fecim-lattice-tools-fyne/launcher.go` | Claim, accessibility, and behavior fixes must be duplicated |
| File logging grows without filesystem retention | `module1-hysteresis/pkg/gui/data_logger.go:105-147,438-466` | The audited checkout accumulated 6,144 CSV files and approximately 50 GB |
| OpenLane defaults to a mutable image | `module6-eda/pkg/openlane/manager.go:38-47,69-75`; `module6-eda/pkg/openlane/config.go:34-49` | Identical source can run different toolchains over time |
| User-facing claim wording conflicts with policy | `docs/research/honesty-audit.md:37-58`; `cmd/fecim-lattice-tools/fyne_launcher.go:957-1005` | Unverified metrics appear on the primary launcher without consistent classification |

## Trust model

The recovery program distinguishes three threats:

1. **Accidental corruption:** partial writes, manual edits, stale cache content, malformed files, and incompatible versions. Phase 1 must detect these reliably.
2. **Untrusted project input:** files, paths, model binaries, and EDA arguments supplied by another user or repository. Parsing must be bounded and command construction must not reinterpret data as shell syntax.
3. **Malicious same-user rewrite:** an attacker who can rewrite every local cache file and manifest can also rewrite unsigned integrity metadata. The first recovery slice does not claim protection against that threat. Cryptographic signatures or an external transparency log require a separate product decision.

Documentation and UI copy must state which threat is covered. “Immutable” means write-once through the application plus corruption detection; it does not mean tamper-proof against an actor controlling the filesystem.

## Priority model

- **P0 — Trust blocker:** Can silently produce wrong scientific output, execute unintended commands, corrupt persisted data, or falsely report external validation.
- **P1 — Trust completion:** Closes validate/run, commit/analyze, claim, reproducibility, or supply-chain gaps.
- **P2 — Architecture and locality:** Reduces parallel implementations after behavior is protected.
- **P3 — Operational and governance:** Controls disk growth, dependency health, documentation drift, and asset licensing.

## Recovery item register

This register is authoritative for scope. Evidence paths are live-code anchors, not historical completion claims.

### TR-COR-01 — Charge-amplifier noise

- **Priority / status / owner:** P0 / Planned / shared peripherals and physics review.
- **Evidence:** `shared/peripherals/charge_amplifier.go:96-110`; missing behavioral coverage in `shared/peripherals/charge_amplifier_test.go:10-79`.
- **Dependencies:** None; first implementation task.
- **Acceptance gate:** Injected ± Gaussian samples are unbiased; the public path samples real noise; final output remains inside configured rails; focused and race tests pass.
- **Non-goals:** Recalibrating material parameters, changing ADC behavior, or regenerating physics golden files.

### TR-PROV-01 — Cached-run integrity

- **Priority / status / owner:** P0 / Planned / workbench experiment owner plus reproducibility review.
- **Evidence:** `workbench/experiment/identity.go:12-33`, `store.go:21-49`, and `runner.go:60-66`.
- **Dependencies:** Explicit accidental-corruption threat model; manifest schema-version decision.
- **Acceptance gate:** Legacy/unverifiable caches and edited manifest/design/result artifacts return a recognizable error and are never marked reused; identity is recomputed from loaded inputs.
- **Non-goals:** Signed provenance or protection from an attacker able to rewrite every local artifact and manifest.

### TR-SER-01 — Bounded model serialization

- **Priority / status / owner:** P0 / Planned / Module 2 weights owner plus robustness review.
- **Evidence:** `module2-crossbar/pkg/weights/serialization.go:55-79,190-217,223-363`.
- **Dependencies:** Owner accepts binary model files as untrusted input; explicit decoded-size limits.
- **Acceptance gate:** Ragged matrices fail before write; every serialized length is checked before allocation; malformed and round-trip tests pass.
- **Non-goals:** A new file format, streaming inference, or changing numeric precision.

### TR-EDA-01 — Non-interpreting EDA command construction

- **Priority / status / owner:** P0 / Planned / Module 6 owner plus security review.
- **Evidence:** `module6-eda/pkg/openlane/runner.go:53-81,165-192`.
- **Dependencies:** Preserve Docker mode and current timeout/result interfaces.
- **Acceptance gate:** Metacharacter-bearing script paths and variables remain single argv values and never enter shell program text; OpenLane package tests pass.
- **Non-goals:** Removing Docker, redesigning the EDA pipeline, or adding a remote execution service.

### TR-VAL-01 — Real external Verilog validation

- **Priority / status / owner:** P0 / Planned / Module 6 and CI owners.
- **Evidence:** `validation/external/eda/verilog_sanity_test.go:9-30`; `.github/workflows/ci.yml:65-117`.
- **Dependencies:** Decision that the Verilog lane gates pull requests; install iverilog in CI and record the exact resolved version.
- **Acceptance gate:** Generated Verilog is passed to an actual validator; tool failure/absence fails the required CI lane; logs record executable and version.
- **Non-goals:** Making all optional ngspice/OpenLane validations mandatory in the same change.

### TR-PRJ-01 — Executable-bundle invariant

- **Priority / status / owner:** P1 / Queued / workbench project owner.
- **Evidence:** `workbench/project/load.go:65-72`, `workbench/experiment/sweep.go:15-30,59-142`, and `cmd/fecim-lattice-tools/workbench_subcommand.go:81-96`.
- **Dependencies:** TR-PROV-01; move or expose sweep validation without creating a `project`↔`experiment` import cycle.
- **Acceptance gate:** Every bundle accepted by `project validate` can be deterministically expanded by `experiment run` under the same limits.
- **Non-goals:** Executing evaluators during load or adding optimization features.

### TR-RUN-01 — Pre-commit result validity

- **Priority / status / owner:** P1 / Queued / workbench experiment owner plus scientific-honesty review.
- **Evidence:** `workbench/experiment/runner.go:147-170`; delayed rejection in `workbench/experiment/analyze.go:50-79`.
- **Dependencies:** TR-PROV-01 and a single reusable result validator.
- **Acceptance gate:** Duplicate metrics, missing objectives/constraints, non-finite values, and unit mismatches fail before immutable commit.
- **Non-goals:** Changing objective ranking or Pareto algorithms.

### TR-CLAIM-01 — Claim policy enforcement

- **Priority / status / owner:** P1 / Queued / documentation, comparison, and launcher owners.
- **Evidence:** `docs/research/honesty-audit.md:37-58`; launcher copy at `cmd/fecim-lattice-tools/fyne_launcher.go:957-1005` and its duplicate.
- **Dependencies:** Owner-approved claim categories and launcher consolidation sequencing.
- **Acceptance gate:** Quantitative UI/report claims carry evidence class and source; a test rejects prohibited or unclassified primary-launcher claims.
- **Non-goals:** Removing literature-backed education content or presenting simulator output as hardware measurement.

### TR-SUPPLY-01 — Reproducible OpenLane image

- **Priority / status / owner:** P1 / Queued / Module 6 and release owners.
- **Evidence:** `module6-eda/pkg/openlane/config.go:34-49`; `module6-eda/pkg/openlane/manager.go:38-47,69-75`.
- **Dependencies:** Choose a supported immutable tag/digest and upgrade cadence.
- **Acceptance gate:** Defaults contain no `latest`; the resolved digest/version is emitted in run provenance and CI checks prevent regression.
- **Non-goals:** Vendoring OpenLane or pinning unrelated developer tools in this item.

### TR-CI-01 — CI hygiene and vulnerability visibility

- **Priority / status / owner:** P1 / Queued / CI and dependency owners.
- **Evidence:** `.github/workflows/ci.yml`; audit verification found `govulncheck` unavailable.
- **Dependencies:** TR-VAL-01; define blocking versus scheduled lanes.
- **Acceptance gate:** formatting is checked without mutating files; `govulncheck` runs in a declared lane; dependency-review ownership and cadence are documented.
- **Non-goals:** Treating every advisory as exploitable or auto-merging dependency updates.

### TR-ARCH-01 — Canonical Fyne launcher

- **Priority / status / owner:** P2 / Queued / default-shell owner.
- **Evidence:** `cmd/fecim-lattice-tools/fyne_launcher.go` and `cmd/fecim-lattice-tools-fyne/launcher.go` are approximately 1,090-line near-duplicates.
- **Dependencies:** TR-CLAIM-01 behavior protection.
- **Acceptance gate:** One implementation owns launcher behavior; both command entry points, if retained, are thin adapters; launcher tests cover the shared implementation.
- **Non-goals:** Redesigning the launcher or removing a compatibility command without deprecation policy.

### TR-ARCH-02 — Production UI state seam

- **Priority / status / owner:** P2 / Needs design checkpoint / Fyne and viewmodel owners.
- **Evidence:** `shared/viewmodel/types.go:99-105`; direct concrete GUI construction at `cmd/fecim-lattice-tools/fyne_app.go:159-201,248-256`.
- **Dependencies:** Owner decision that `ModulePort` is the intended production seam; one-module pilot design.
- **Acceptance gate:** The pilot module’s production UI and headless tests cross the same adapter and lifecycle contract before another module migrates.
- **Non-goals:** A seven-module flag day or retaining parallel state models merely to preserve test counts.

### TR-ARCH-03 — Shared electrical-solver seam

- **Priority / status / owner:** P2 / Queued / Modules 2 and 4 plus shared-core owner.
- **Evidence:** solver types in `module4-circuits/pkg/arraysim/types.go`; Module 2 validation imports at `validation/module2/ngspice_comparison_report_test.go:16-17,173-177`; Module 4 use at `module4-circuits/pkg/gui/device_state.go:152-156` and `device_state_compute.go:199-215`.
- **Dependencies:** Characterization tests around current Tier-A/Tier-B behavior; ADR 0004 public-module boundary.
- **Acceptance gate:** Reusable solver code has one shared owner; Modules 2 and 4 remain thin adapters and existing validation tolerances stay green.
- **Non-goals:** Unifying every circuit package or changing solver physics during the move.

### TR-OPS-01 — Generated-log retention

- **Priority / status / owner:** P3 / Queued / Module 1 and operations owners.
- **Evidence:** `module1-hysteresis/pkg/gui/data_logger.go:105-147,438-466`; audited checkout contained 6,144 old files totaling approximately 50 GB.
- **Dependencies:** Decide default age/count/size budget and preserve active-run files.
- **Acceptance gate:** Automated tests cover pruning and failure safety; defaults impose a documented finite bound.
- **Non-goals:** Deleting user data without policy, changing CSV schema, or relying on manual cleanup.

### TR-DOC-01 — Source-backed documentation repair

- **Priority / status / owner:** P3 / Queued / documentation owners with subsystem reviewers.
- **Evidence:** stale `../HYPER_ANALYSIS_REPORT.md` link in `docs/internals/gui/README.md:43`; broader path/command drift recorded in `docs/internals/audits/DOC_DRIFT_AUDIT_2026-02-11.md`.
- **Dependencies:** TR-ARCH-01/02 decisions where docs currently describe intended architecture as live behavior.
- **Acceptance gate:** Linked commands and paths exist; pages distinguish current state, planned state, and simulation-only evidence.
- **Non-goals:** Cosmetic prose churn or hiding unresolved architecture decisions.

### TR-LICENSE-01 — Research-PDF redistribution review

- **Priority / status / owner:** P3 / Owner/legal decision required / repository owner.
- **Evidence:** tracked PDF corpus identified during the repository audit; repository license alone does not establish third-party redistribution rights.
- **Dependencies:** Human-owned license/permission evidence.
- **Acceptance gate:** Every tracked third-party PDF has redistribution evidence, or is replaced with citation metadata and a lawful external link.
- **Non-goals:** Deleting sources before review or making legal conclusions from filename/metadata alone.

### TR-METRIC-01 — Evidence-based health reporting

- **Priority / status / owner:** P3 / Queued / maintainers and release owners.
- **Evidence:** historical completion/test-count summaries in `TODO.md`; external validation can currently skip or no-op.
- **Dependencies:** TR-VAL-01 and agreed release gates.
- **Acceptance gate:** Health reports list trust surfaces, exact commands/tools, passes, failures, and skips; raw test count is contextual only.
- **Non-goals:** Removing useful coverage metrics or inventing a single composite quality score.

## Roadmap

### Phase 0 — Freeze and establish the recovery gate

**Objective:** Prevent new surface area from landing on known-unsafe assumptions.

Actions:

- Mark RDW-UI-01 and RDW-EDA-01 feature delivery as gated by Phase 1 rather than abandoned.
- Do not describe cached runs as tamper-proof or externally validated.
- Do not use `validation/external/eda/verilog_sanity_test.go` as release evidence until it executes a tool.
- Require RED/GREEN/final-verification receipts for every production recovery item.
- Preserve the current dirty owner work; recovery changes must not absorb unrelated README, screenshotter, asset, or race-test edits.

**Exit gate:** roadmap and implementation plan are linked from `TODO.md` and the developer documentation index.

### Phase 1 — Critical correctness and trust hardening

**Objective:** Remove the five highest-risk false guarantees.

| ID | Deliverable | Required evidence | Exit condition |
|---|---|---|---|
| TR-COR-01 | Correct zero-mean Gaussian charge-amplifier noise and post-noise rail clipping | Deterministic injected-sample tests, rail tests, package tests | `SenseWithNoise` varies under real RNG, injected ± samples are unbiased, output stays within rails |
| TR-PROV-01 | Validate cached run identity and artifact integrity before reuse | Design/result tamper tests, partial-file tests, runner tests | Corrupted cache returns a typed/recognizable error and is never marked reused |
| TR-SER-01 | Reject ragged weights and bound all binary lengths before allocation | Ragged-save tests, malformed-length tests, round-trip tests | No stream shift; configured limits reject oversized metadata/layers/shapes/weights/biases |
| TR-EDA-01 | Remove data interpolation from Docker shell commands | Pure argument-builder tests with metacharacters, OpenLane tests | Dynamic values are passed as quoted positional arguments/argv and never concatenated into shell program text |
| TR-VAL-01 | Execute a real Verilog validator in a required CI lane | Test invokes iverilog or verilator; CI installs the tool and records its resolved version | Tool absence/failure fails the required lane; successful logs include executable and version |

**Required aggregate gate:**

```bash
go test -v -count=1 ./shared/peripherals ./module2-crossbar/pkg/weights \
  ./workbench/experiment ./module6-eda/pkg/openlane ./validation/external/eda
go test -race -count=1 ./shared/peripherals ./workbench/experiment
go vet ./...
go test ./...
```

External validation additionally runs in CI with the validator installed and its resolved version recorded. Local absence may skip the integration test, but cannot satisfy the release gate. Immutable runner/tool pinning remains part of TR-CI-01.

### Phase 2 — Trust invariant completion

**Objective:** Make successful interfaces carry useful guarantees.

| ID | Recommendation | Locality/leverage result |
|---|---|---|
| TR-PRJ-01 | Make `project.Load` guarantee that the sweep is executable | `project validate`, run, report, and future Fyne callers share one validity definition |
| TR-RUN-01 | Validate successful results before immutable commit | Evaluators, cache reuse, analysis, and reports inherit one structural-result invariant |
| TR-CLAIM-01 | Add a source-backed claim registry contract for launcher/comparison copy | Unverified figures cannot reappear without evidence labels |
| TR-SUPPLY-01 | Pin OpenLane by immutable version/digest and record it in provenance | Same source resolves to the same tool image |
| TR-CI-01 | Add non-mutating formatting checks, `govulncheck`, and scheduled dependency review | CI detects formatting drift and known Go vulnerabilities instead of rewriting or ignoring them |

**Exit gate:** `project validate` cannot succeed for a bundle that fails deterministic expansion; invalid successful results cannot be committed; launcher claims satisfy the honesty audit; toolchain identity is recorded.

### Phase 3 — Architecture convergence

**Objective:** Remove parallel behavior models and improve change locality without a broad rewrite.

1. **TR-ARCH-01 — Consolidate the Fyne launcher.** Move the near-identical launcher implementation behind one internal module; retain thin command adapters.
2. **TR-ARCH-02 — Join `ModulePort` to the default Fyne shell incrementally.** Pilot one module. Resolve state ownership and lifecycle through an adapter before migrating another module. Do not replace all seven modules in one change.
3. **TR-ARCH-03 — Move the reusable electrical array solver behind a shared seam.** Keep Public Module 4 as a thin adapter, consistent with ADR 0004 (`docs/internals/adr/0004-root-layout-and-public-module-directories.md:3-8`).
4. Delete tests for superseded shallow modules when equivalent tests cross the deeper production seam. Test count is not a retention argument.

**Exit gate:** production and tests cross the same chosen interface for the pilot UI module; one launcher implementation remains; the electrical solver has one reusable owner.

### Phase 4 — Operational and governance hardening

| ID | Recommendation | Exit condition |
|---|---|---|
| TR-OPS-01 | Add age/size/count retention for generated CSV logs | Automated tests cover rotation; default logging cannot grow without a configured bound |
| TR-DOC-01 | Repair stale command and architecture documentation | Every documented command exists; UI docs describe the live shell rather than the intended future shell |
| TR-LICENSE-01 | Review tracked research PDFs for redistribution rights | Each PDF has permission/license evidence or is replaced by citation metadata and a lawful link |
| TR-METRIC-01 | Replace “number of tests” health claims with trust-surface gates | Status reports identify which external tools ran and which validations skipped |

**Exit gate:** bounded runtime artifacts, source-backed documentation, and resolved asset redistribution status.

## Release gates

A release candidate is not “research-ready” unless all applicable gates are recorded:

- [ ] Full Go test suite passes from a clean checkout.
- [ ] Focused race gates pass for changed concurrency/state packages.
- [ ] Cached-run corruption tests pass.
- [ ] Model binary malformed-input tests pass.
- [ ] Required Verilog external validation ran; executable and version are recorded.
- [ ] OpenLane image is pinned and included in provenance when used.
- [ ] Quantitative launcher/report claims pass the honesty contract.
- [ ] Skipped validations are listed as missing evidence, not success.
- [ ] RED/GREEN/final verification receipts are attached to every behavior change.

## Work acceptance rules

Every roadmap item must specify:

- **Fact:** live file/command evidence for the defect.
- **Invariant:** the behavior that must become true.
- **RED:** focused automated test and expected failure.
- **GREEN:** minimum implementation and passing focused test.
- **Regression gate:** package, race, external-tool, or repository command.
- **Trust wording:** what the project may claim after the change and what remains outside scope.
- **Rollback:** how to disable or revert the change without corrupting existing project data.

## Explicit non-recommendations

Do not:

- rewrite the project in another language;
- introduce a database, distributed runner, service layer, or signing infrastructure in Phase 1;
- migrate all seven GUIs while critical correctness defects remain;
- regenerate physics golden data for unrelated recovery changes;
- count skipped external tests as passed validation;
- delete research sources before the repository owner decides the lawful retention strategy;
- claim malicious same-user tamper resistance from unsigned local hash files.

## Owner decisions required

These are product/governance choices, not implementation details to guess inside a pull request:

1. **Model binary trust boundary:** Recommended default—treat every binary model file as untrusted, including files inside a project checkout. If the owner chooses trusted-only input, the UI/CLI must say so and the loader still needs accidental-corruption bounds.
2. **External EDA gate placement:** Recommended default—make real Verilog syntax validation required on every pull request that can affect Module 6/export behavior, with a required release lane as defense in depth. A release-only lane leaves broken generated Verilog on `main` longer.
3. **Production UI seam:** Decide whether `shared/viewmodel.ModulePort` is the intended production state/lifecycle seam. Until that is explicit, only a one-module adapter pilot is authorized; a repository-wide migration is not.
4. **Research PDF rights:** The repository owner must provide redistribution/license evidence or authorize replacement with citations and lawful links. Agents must not infer permission.
5. **Local cache adversary:** Recommended Phase 1 promise—detect accidental corruption and application-bypassing edits, not malicious same-user rewrites. Stronger integrity requires signed or externally anchored manifests and a separate design.

## Ownership and review

| Track | Primary reviewer profile | Required second review |
|---|---|---|
| Charge amplifier and scientific result validation | Peripheral/physics maintainer | Scientific-honesty reviewer |
| Cache and project invariants | Workbench maintainer | Reproducibility reviewer |
| Binary serialization | Module 2 maintainer | Security/robustness reviewer |
| EDA execution and external validation | Module 6 maintainer | Security/CI reviewer |
| UI convergence and claims | Fyne maintainer | Scientific-honesty reviewer |
| PDF licensing | Repository owner | Legal/licensing decision owner |

## Update triggers

Revisit this roadmap when any of the following changes:

- `SenseWithNoise`, charge-amplifier noise parameters, or peripheral RNG policy;
- workbench run identity, manifest schema, cache layout, evaluator result schema, or project sweep schema;
- model binary format version or loader limits;
- OpenLane/KLayout/Yosys command construction or Docker image;
- required CI tools or external-validation policy;
- `ModulePort`, `EmbeddedApp`, or default Fyne shell ownership;
- generated-log defaults or research asset retention policy.

Update `TODO.md` status and this roadmap in the same change when a phase gate opens or closes.

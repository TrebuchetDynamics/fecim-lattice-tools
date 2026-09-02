# Technical Recovery Roadmap

**Date:** 2026-09-01; Phase 1 and TR-RUN-01 receipts updated 2026-09-02

**Status:** Active — Phase 1 complete; Phase 2 active with TR-RUN-01 done and four P1 items queued; overall recovery and release gates remain open
**Scope:** Scientific correctness, provenance, serialization, EDA execution, external validation, architecture, operations, and documentation  
**TDD:** N/A — this is a documentation-only evidence/status update and cannot change runtime behavior; every future behavior change remains separately gated by RED → GREEN → REFACTOR.

## What this is

This roadmap turns the September 2026 repository audit into a sequenced recovery program. It is not a rewrite proposal and it does not discard the project’s validated physics, workbench, or GUI assets. It identifies the trust guarantees the project currently overstates, defines the order in which they must be repaired, and sets evidence-based exit criteria before feature work resumes.

Start here if you are choosing work, reviewing a recovery pull request, or deciding whether a build is suitable for research use. The executable Phase 1 plan is [Critical Correctness and Trust Hardening](../superpowers/plans/2026-09-01-critical-correctness-and-trust-hardening.md).

## Executive recommendation

Treat FeCIM Lattice Tools as a capable research prototype under hardening—not as a release-ready validated workbench. The repository builds and its broad Go test suite passes, but several high-consequence interfaces do not enforce the guarantees their names, comments, tests, or documentation imply.

The recovery policy is:

1. **Keep release readiness gated by the remaining roadmap.** Phase 1 and TR-RUN-01 have exited green, so RDW-UI-01 and RDW-EDA-01 are queued/unblocked; four Phase 2 items and later release gates remain open.
2. **Repair observable correctness before architecture.** Fix charge noise, cached-run integrity, binary serialization, EDA command construction, and real external validation first.
3. **Make validation semantics honest.** A skipped or no-op external test is not validation. A locally cached file is not immutable merely because it lives under a SHA-named directory.
4. **Deepen existing seams rather than creating parallel systems.** Project loading should guarantee an executable bundle; committed runs should be structurally valid; the default UI and its tests should eventually cross the same state seam.
5. **Use TDD for every production change.** Each recovery item starts with a focused failing test, records RED and GREEN, and ends with the narrow package gate plus the repository gate required by the affected trust surface.

## Evidence summary

Rows explicitly labeled pre-recovery preserve historical defect evidence and are not descriptions of current live behavior; their closure receipts are recorded below. Unlabeled rows remain open recovery risks.

| Risk | Baseline/open evidence | Baseline or current consequence |
|---|---|---|
| **Pre-recovery:** charge-amplifier noise was deterministic and positively biased | Historical anchors: `shared/peripherals/charge_amplifier.go:96-110`; no `SenseWithNoise` coverage in the then-current `shared/peripherals/charge_amplifier_test.go:10-79` | Before TR-COR-01, repeated samples were identical, zero input acquired positive offset, and noise could exceed output rails after clipping |
| **Pre-recovery:** cached run artifacts were not rebound to identity on load | Historical anchors: `workbench/experiment/identity.go:12-33`; `workbench/experiment/store.go:21-49`; `workbench/experiment/runner.go:60-66` | Before TR-PROV-01, edited design/result files could be accepted as reused scientific evidence |
| **Pre-recovery:** binary model format accepted ragged and unbounded data | Historical anchor: `module2-crossbar/pkg/weights/serialization.go:55-79,190-217,223-363` | Before TR-SER-01, ragged rows shifted the stream and untrusted lengths could allocate excessive memory |
| **Pre-recovery:** Docker EDA wrappers interpolated data into `sh -c` | Historical anchor: `module6-eda/pkg/openlane/runner.go:53-81,165-192` | Before TR-EDA-01, metacharacters in public arguments could be interpreted in a container with a writable project mount |
| **Pre-recovery:** Verilog “sanity” validation performed no validation | Historical anchors: `validation/external/eda/verilog_sanity_test.go:9-30`; `.github/workflows/ci.yml:65-117` | Before TR-VAL-01, CI could report success while no simulator or linter executed |
| `project validate` accepts bundles that `experiment run` rejects | `workbench/project/load.go:65-72`; `workbench/experiment/sweep.go:10-56,59-142`; `cmd/fecim-lattice-tools/workbench_subcommand.go:81-96` | User-visible validate/run semantic mismatch |
| **Pre-TR-RUN-01:** structurally invalid successful results could be persisted | Historical anchors: `workbench/experiment/runner.go:147-170`; delayed rejection in `workbench/experiment/analyze.go:50-79` | Before TR-RUN-01, duplicate, missing, non-finite, provenance-incomplete, or unit-invalid metrics could become immutable before rejection |
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

This register is authoritative for scope. For completed items, the original evidence paths are pre-recovery anchors and the receipts below are authoritative completion evidence; open-item evidence remains a live-code lead to verify before implementation.

### TR-COR-01 — Charge-amplifier noise

- **Priority / status / owner:** P0 / Done (`090562d96683da125e08110b104e770d38801db5`) / shared peripherals and physics review completed.
- **Evidence:** `shared/peripherals/charge_amplifier.go:96-110`; missing behavioral coverage in `shared/peripherals/charge_amplifier_test.go:10-79`.
- **Dependencies:** None; first implementation task.
- **Acceptance gate:** Injected ± Gaussian samples are unbiased; the public path samples real noise; final output remains inside configured rails; focused and race tests pass.
- **Non-goals:** Recalibrating material parameters, changing ADC behavior, or regenerating physics golden files.

### TR-PROV-01 — Cached-run integrity

- **Priority / status / owner:** P0 / Done (`d4f8305e5bde754aa352b643126f9e1c5a2c2986`) / workbench experiment and reproducibility review completed.
- **Evidence:** `workbench/experiment/identity.go:12-33`, `store.go:21-49`, and `runner.go:60-66`.
- **Dependencies:** Explicit accidental-corruption threat model; manifest schema-version decision.
- **Acceptance gate:** Legacy/unverifiable caches and edited manifest/design/result artifacts return a recognizable error and are never marked reused; identity is recomputed from loaded inputs.
- **Non-goals:** Signed provenance or protection from an attacker able to rewrite every local artifact and manifest.

### TR-SER-01 — Bounded model serialization

- **Priority / status / owner:** P0 / Done (`cd7c03ecbbe8532f03e358a1d391811b5526296f`) / Module 2 weights and robustness review completed.
- **Evidence:** `module2-crossbar/pkg/weights/serialization.go:55-79,190-217,223-363`.
- **Dependencies:** Owner accepts binary model files as untrusted input; explicit decoded-size limits.
- **Acceptance gate:** Ragged matrices fail before write; every serialized length is checked before allocation; malformed and round-trip tests pass.
- **Non-goals:** A new file format, streaming inference, or changing numeric precision.

### TR-EDA-01 — Non-interpreting EDA command construction

- **Priority / status / owner:** P0 / Done (`adcc525e1a3a14e025154626ccf61d541b95a764`) / Module 6 and security review completed.
- **Evidence:** `module6-eda/pkg/openlane/runner.go:53-81,165-192`.
- **Dependencies:** Preserve Docker mode and current timeout/result interfaces.
- **Acceptance gate:** Metacharacter-bearing script paths and variables remain single argv values and never enter shell program text; OpenLane package tests pass.
- **Non-goals:** Removing Docker, redesigning the EDA pipeline, or adding a remote execution service.

### TR-VAL-01 — Real external Verilog validation

- **Priority / status / owner:** P0 / Done (`bde96d080fed9fb1ec393140419b2a16da7614ff`) / Module 6 and CI review completed; required external-validation CI passed.
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

- **Priority / status / owner:** P1 / Done ([`b5f14295926983e5da5350cb4730294a116fe6ec`](https://github.com/TrebuchetDynamics/fecim-lattice-tools/commit/b5f14295926983e5da5350cb4730294a116fe6ec)) / workbench experiment owner; two independent follow-up reviews clean.
- **Baseline evidence:** Before the fix, `workbench/experiment/runner.go:147-170` could commit invalid successful results before the defense-in-depth checks in `workbench/experiment/analyze.go:50-79`.
- **Dependencies:** TR-PROV-01 and a single reusable result validator; satisfied for this item.
- **Acceptance gate:** Fresh and cached successful results validate before commit/reuse; duplicate or blank metrics, missing objectives/constraints, non-finite values, incomplete metric provenance, and exact constraint-unit mismatches return recognizable `ErrInvalidResult` systemic errors.
- **Non-goals/deferred:** Changing objective ranking or Pareto algorithms; assigning unit authority to `Objective`; deciding failed-result partial-metric policy; enforcing `Project.ModelVersion`/evaluator agreement; or adding derived-metric source fields.

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

### Phase 1 — Critical correctness and trust hardening — Complete

**Objective:** Remove the five highest-risk false guarantees. Completed 2026-09-02 with the receipts below.

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

#### Phase 1 completion receipts

**Status:** Complete as of 2026-09-02. All five changes were reviewed and their focused/final gates passed. This phase receipt does not close Phase 2, the overall recovery roadmap, or the release gates.

| ID | Final commit | Review status | Trust boundary after completion |
|---|---|---|---|
| TR-COR-01 | [`090562d96683da125e08110b104e770d38801db5`](https://github.com/TrebuchetDynamics/fecim-lattice-tools/commit/090562d96683da125e08110b104e770d38801db5) | Reviewed; focused and race gates passed | Signed Gaussian samples and post-noise rail clipping are covered; this receipt does not recalibrate hardware behavior |
| TR-PROV-01 | [`d4f8305e5bde754aa352b643126f9e1c5a2c2986`](https://github.com/TrebuchetDynamics/fecim-lattice-tools/commit/d4f8305e5bde754aa352b643126f9e1c5a2c2986) | Reviewed; experiment, race, and workbench gates passed | Unsigned cache digests detect accidental corruption/application-bypassing edits; they are not tamper-proof against coherent same-user rewriting |
| TR-SER-01 | [`cd7c03ecbbe8532f03e358a1d391811b5526296f`](https://github.com/TrebuchetDynamics/fecim-lattice-tools/commit/cd7c03ecbbe8532f03e358a1d391811b5526296f) | Reviewed; focused, package, race, vet, and Module 2 gates passed | Binary bounds and rectangularity checks constrain parsing; they do not authenticate file origin |
| TR-EDA-01 | [`adcc525e1a3a14e025154626ccf61d541b95a764`](https://github.com/TrebuchetDynamics/fecim-lattice-tools/commit/adcc525e1a3a14e025154626ccf61d541b95a764) | Reviewed; focused, package, race, vet, and Module 6 gates passed | Dynamic values no longer enter these Docker shell programs; this is not a general container sandbox |
| TR-VAL-01 | [`bde96d080fed9fb1ec393140419b2a16da7614ff`](https://github.com/TrebuchetDynamics/fecim-lattice-tools/commit/bde96d080fed9fb1ec393140419b2a16da7614ff) | Reviewed; required external-validation CI passed | PASS covers Verilog syntax/elaboration only, not synthesis, timing, physical design, or hardware validation |

**TR-COR-01 RED/GREEN**

- RED command: `go test -count=1 ./shared/peripherals -run 'TestChargeAmplifier_SenseWithNoise'`.
- Recorded failures: the focused test first failed to compile because `senseWithGaussian` was undefined; subsequent RED observations recorded identical public outputs, post-noise rail overflow, and the old 5.6 aC default failing the near-23 aC integrated kT/C test.
- GREEN/final gates: `go test -count=1 ./shared/peripherals -run 'TestChargeAmplifier'` and `go test -race -count=1 ./shared/peripherals` passed.

**TR-PROV-01 RED/GREEN**

- Initial RED command: `go test -count=1 ./workbench/experiment -run 'TestRunRejects(Tampered|Malformed|Oversized|CachedDesign|Legacy)'`.
- Initial failure summary: result tamper lacked `ErrUnverifiableRun`; malformed/oversized artifacts were not safely recognizable or bounded; re-signed design/manifest provenance tampering was reused; schema-1 rejection lacked archive/removal guidance.
- Review RED commands: `go test -count=1 ./workbench/experiment -run '^TestRunRejectsOversizedArtifactsBeforeCreatingRunStorage$'`; `go test -count=1 ./workbench/experiment -run '^TestRunRejectsNonCanonicalCachedManifest$'`; and `go test -count=1 ./workbench/experiment -run '^TestStoreCommitPropagatesUnverifiableRenameWinner$'`. The three oversized commits succeeded, noncanonical manifest edits were reused, and the private rename seam was absent.
- GREEN/final gates: the same review tests passed, followed by `go test -count=1 ./workbench/experiment`, `go test -race -count=1 ./workbench/experiment`, and `go test -count=1 ./workbench/...` (all exit 0).

**TR-SER-01 RED/GREEN**

- Initial RED command: `go test -count=1 ./module2-crossbar/pkg/weights -run 'Test(SaveBinaryRejectsRagged|LoadModelBinaryRejectsOversized)'`.
- Recorded failures: ragged/malformed binaries reached truncation, EOF, or unbounded allocation before validation; reviews exposed row-count, cumulative retained-memory, matrix-sized read-buffer, and pre-marshal metadata amplification.
- Append-semantics RED: `go test -count=1 ./module2-crossbar/pkg/weights -run '^TestModelSaveLoadBinaryPreservesIndependentRowAppendSemantics$'` failed because appending row 0 changed row 1 from `[3 4]` to `[9 4]`.
- GREEN/final gates: the append-semantics test and the focused bounded-serialization selection passed, followed by `go test -count=1 ./module2-crossbar/pkg/weights`, `go test -race -count=1 ./module2-crossbar/pkg/weights`, `go vet ./module2-crossbar/pkg/weights`, and `go test -count=1 ./module2-crossbar/...` (all exit 0).

**TR-EDA-01 RED/GREEN**

- Initial RED commands: `go test -count=1 ./module6-eda/pkg/openlane -run 'TestDockerOpenROADArgsKeepDynamicValuesOutOfFixedShellProgram'`; `go test -count=1 ./module6-eda/pkg/openlane -run 'TestDockerKLayoutArgsKeepDynamicValuesOutOfFixedShellProgram'`; and `go test -count=1 ./module6-eda/pkg/openlane -run 'TestDockerToolArgsConfineScriptBasenames'`. The first two failed because their argument helpers were undefined; the third returned `/design/.` and `/design/..` for invalid inputs.
- Compatibility RED: `go test -count=1 ./module6-eda/pkg/openlane -run 'TestDockerToolArgsMapScriptsInsideDesignMount'` failed because safe nested paths were flattened and traversal paths were accepted.
- GREEN/final gates: all four focused tests passed, followed by `go test -count=1 ./module6-eda/pkg/openlane -run 'TestDocker(OpenROADArgs|KLayoutArgs|ToolArgs)'`, `go test -count=1 ./module6-eda/pkg/openlane`, `go test -race -count=1 ./module6-eda/pkg/openlane`, and `go vet ./module6-eda/pkg/openlane` (all exit 0). A worker's first `go test -count=1 ./module6-eda/...` attempt hit an unrelated Fyne concurrent-map panic; its immediate standalone rerun passed, and the parent fresh `go test ./...` aggregate also passed, so the panic remains flake-risk evidence rather than a reproduced Phase 1 failure.

**TR-VAL-01 RED/GREEN**

- Initial RED command: `go test -count=1 ./scripts -run TestCIRequiresRealExternalVerilogValidation`. The contract found missing iverilog install/version/exact-test tokens, optional/failure-masking behavior, and no-op execution; review RED also found missing production-generator coverage and the exact ngspice filter.
- Final ordering RED: `go test -count=1 ./scripts -run TestExternalValidationWorkflowRejectsNonGatingMutations` failed because the contract accepted the required Verilog test before the install step.
- GREEN/final gates: `go test -count=1 ./scripts -run 'Test(CIRequiresRealExternalVerilogValidation|ExternalValidationWorkflowRejectsNonGatingMutations)'` passed, along with the recorded focused/full scripts, exact local selections, vet, diff, and YAML checks.
- Local external result was explicitly **SKIP**, not PASS: `go test -v -count=1 ./validation/external/eda -run TestVerilogSanityCheck` validated the legacy generator/stub and production GUI array/cell pair structure, then skipped because neither iverilog nor verilator was installed locally. This local skip was never counted as completion evidence.

**Phase 1 local aggregate (2026-09-02)**

Each command exited 0:

```text
go test -v -count=1 ./shared/peripherals ./module2-crossbar/pkg/weights ./workbench/experiment ./module6-eda/pkg/openlane ./validation/external/eda
go test -race -count=1 ./shared/peripherals ./workbench/experiment
go vet ./...
go test ./...
bash scripts/check-architecture.sh
git diff --check 20a5792f..bde96d08
```

The external sanity test in the first command explicitly skipped locally because no validator was installed; that skip is separate from, and did not substitute for, the gating CI PASS.

**PR and required external CI**

- Pull request: [#11](https://github.com/TrebuchetDynamics/fecim-lattice-tools/pull/11).
- First gating CI workflow: [run 33643451506](https://github.com/TrebuchetDynamics/fecim-lattice-tools/actions/runs/33643451506).
- Required external-validation job: [job 100292070256](https://github.com/TrebuchetDynamics/fecim-lattice-tools/actions/runs/33643451506/job/100292070256) — **PASS (gating)**. The required external Verilog lane did not skip.
- The optional ngspice CI step skipped because ngspice was unavailable. Like the local Verilog skip, this is recorded as missing optional validation evidence and is not counted as success.
- Exact validator version evidence: `Icarus Verilog version 12.0 (stable) ()`.
- Literal test evidence: `--- PASS: TestVerilogSanityCheck (0.01s)`; package `PASS`.
- CI validated both the legacy generator/stub and the production GUI array/cell pair. This evidence is syntax/elaboration only.
- The version was resolved through apt and recorded, but apt resolution is not immutable pinning. TR-CI-01 and TR-SUPPLY-01 remain open.

Phase 1 completion cleared the Phase 1 gate for both feature items. TR-RUN-01 has now also cleared RDW-EDA-01's remaining recovery gate; RDW-UI-01 and RDW-EDA-01 are queued/unblocked.

### Phase 2 — Trust invariant completion — Active

**Objective:** Make successful interfaces carry useful guarantees. TR-RUN-01 is complete; four P1 items remain queued, so Phase 2 and release readiness remain open.

| ID | Status | Recommendation | Locality/leverage result |
|---|---|---|---|
| TR-PRJ-01 | Queued | Make `project.Load` guarantee that the sweep is executable | `project validate`, run, report, and future Fyne callers share one validity definition |
| TR-RUN-01 | **Done** | Validate fresh and cached successful results before immutable commit/reuse | Evaluators, cache reuse, analysis, and reports inherit one structural-result invariant |
| TR-CLAIM-01 | Queued | Add a source-backed claim registry contract for launcher/comparison copy | Unverified figures cannot reappear without evidence labels |
| TR-SUPPLY-01 | Queued | Pin OpenLane by immutable version/digest and record it in provenance | Same source resolves to the same tool image |
| TR-CI-01 | Queued | Add non-mutating formatting checks, `govulncheck`, and scheduled dependency review | CI detects formatting drift and known Go vulnerabilities instead of rewriting or ignoring them |

#### Phase 2 progress receipt — TR-RUN-01

- **Implementation and review:** [`b5f14295926983e5da5350cb4730294a116fe6ec`](https://github.com/TrebuchetDynamics/fecim-lattice-tools/commit/b5f14295926983e5da5350cb4730294a116fe6ec) changed exactly `workbench/experiment/runner.go` and `workbench/experiment/runner_test.go`. [PR #12](https://github.com/TrebuchetDynamics/fecim-lattice-tools/pull/12) passed [workflow 33651992963](https://github.com/TrebuchetDynamics/fecim-lattice-tools/actions/runs/33651992963) and its [CI job 100321043703](https://github.com/TrebuchetDynamics/fecim-lattice-tools/actions/runs/33651992963/job/100321043703). Two independent implementation follow-up reviews were clean.
- **Invariant now enforced:** Every fresh or cached `StatusSuccess` result validates against the active project before commit or reuse. Success carries no `Failure`; metric names are globally unique and nonblank; values are finite; every metric has a nonblank unit/model/assumption and recognized evidence classification; configured objectives and constraints are present; and constraint units match exactly. Violations return recognizable `ErrInvalidResult` and are treated as systemic runner errors.
- **Preserved behavior:** A structurally valid constraint violation still commits and `Analyze` marks it infeasible. Cached `StatusFailed` results remain reusable without evaluator execution. Rejected cached successful artifacts remain byte-for-byte unchanged.
- **Historical cache behavior:** Existing invalid historical caches are rejected when `Run` attempts reuse under the current project rules; they are not rewritten or deleted.
- **Original RED/GREEN slices:** The commit receipt records focused RED then GREEN commands for success carrying `Failure`, blank/duplicate metric names, non-finite values, missing unit/model/nonblank assumption, invalid evidence, missing objective/constraint metrics, and exact constraint-unit mismatch. It also records GREEN controls for structurally valid results, valid-infeasible constraint violations, scheduling stop/preserved prior completions, and no commits for rejected/later points. See the linked commit body for every exact command and failure summary.
- **Review mutation receipts:** `TestRunRejectsDuplicateOptionalMetricNames` was GREEN on correct production, RED against a temporary objective-only uniqueness mutant, and GREEN after restoration. `go test -count=1 ./workbench/experiment -run '^TestRunRejectsCachedSuccessMissingNewObjective$'` was RED when cache reuse bypassed project-aware validation and GREEN after the shared validator was applied. `TestRunRejectsCachedSuccessWithChangedConstraintUnit` was GREEN after the fix, RED against the original cache-bypass mutant, then GREEN after restoration. `TestRunStillReusesCachedDesignFailure` remained GREEN. No mutant was retained.
- **Focused/final implementation gates:** The combined review selection, `go test -count=1 ./workbench/experiment`, `go test -race -count=1 ./workbench/experiment`, `go test -count=1 ./workbench/...`, focused `workbench/fecim`/`workbench/report`, the clean index-only CLI end-to-end test, `go vet ./workbench/...`, and `git diff --check` all passed as recorded in the implementation commit.
- **Parent local aggregate:** `go test -count=1 ./workbench/...`; `go test -race -count=1 ./workbench/experiment`; `go vet ./...`; `go test ./...`; `bash scripts/check-architecture.sh`; and `git diff --check d4e0b6b8..b5f14295` each exited 0.
- **CI external evidence:** [external-validation job 100321043492](https://github.com/TrebuchetDynamics/fecim-lattice-tools/actions/runs/33651992963/job/100321043492) passed. Its optional ngspice step skipped because ngspice was unavailable; that remains missing optional evidence and is not counted as success.
- **Deferred boundaries:** `Objective` has no unit-authority field; failed-result partial-metric policy is undecided; `Project.ModelVersion`/evaluator agreement and derived-metric source fields remain separate work. Unsigned caches remain outside malicious coherent-rewrite resistance. `Analyze` remains defense in depth for external/manual records rather than the first validation point for runner-managed successful results.

**Exit gate:** `project validate` still must be made unable to accept a bundle that fails deterministic expansion; launcher claims must satisfy the honesty audit; and toolchain identity must be recorded. TR-RUN-01 has closed the invalid-successful-result commit/reuse portion only.

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

- [x] Full Go test suite passes from a clean checkout (first gating CI workflow).
- [x] Focused race gates pass for changed concurrency/state packages.
- [x] Cached-run corruption tests pass.
- [x] Fresh and cached successful results validate before immutable commit/reuse.
- [x] Model binary malformed-input tests pass.
- [x] Required Verilog external validation ran; executable and version are recorded.
- [ ] OpenLane image is pinned and included in provenance when used.
- [ ] Quantitative launcher/report claims pass the honesty contract.
- [x] Skipped validations are listed as missing evidence, not success.
- [x] RED/GREEN/final verification receipts are attached to every completed recovery behavior change.

The unchecked immutable OpenLane pinning and quantitative-claim contract gates keep release readiness open; Phase 1 and TR-RUN-01 completion do not constitute release approval.

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

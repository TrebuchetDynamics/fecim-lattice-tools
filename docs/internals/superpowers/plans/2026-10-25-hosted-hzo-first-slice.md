# First hosted slice — one HZO hysteresis loop

Status: Proposed implementation plan for owner review after the parity/GPU/budget report. **Not started. No React implementation, dependency installation, Astro/Fyne deletion or commits authorized by this plan.** Accepted product scope is [ADR 0005](../../adr/0005-unified-react-hosted-and-local.md); prerequisite evidence is the [three-runtime gate receipt](../../testing/wasm-parity-gate.md).

## Goal and exclusions

One hosted/static React screen computes one fixed HZO P–E loop in a Go/WASM Worker and exposes the complete result through a paginated native table, Go-generated text summary and Go speakable values. A supplementary SVG plot mirrors the same data but is never required to inspect a value. A persistent visible/accessible disclaimer explains local browser computation, no scientific-data upload, and which later features need `fecim serve`.

Do not build Projects/Configuration/Experiment/Results as four separate views. Do not implement local server, connection pairing, filesystem access, project editors, large sweeps, ISPP, crossbar/peripheral screens, comparisons, downloads, user accounts or a second landing site. Shared React shell means one title/navigation region/main content and one HZO feature, not a generic dashboard scaffold. Existing Astro stays unchanged until the slice is explicitly authorized to replace it; no parallel new marketing application is created.

## Fixed numerical contract — planner's calls

Reuse `module1-hysteresis/pkg/ferroelectric.DefaultHZO`, `NewPreisachModel`, `SetTemperature` and `GetHysteresisLoop`. This selects a **quasi-static Preisach** demonstration, not a new L-K implementation, fitted dataset, hardware measurement or dynamic switching-frequency claim. The constructors already use the shared Go material/physics core. No TypeScript physics/defaults or duplicated Everett equation.

For the first slice, Go owns a fixed preset/profile: current DefaultHZO baseline, 300 K, field amplitude `3 * material.Ec`, `GetHysteresisLoop(Emax, 64)`. Current API yields 257 samples (129 ascending including endpoints, 128 descending). These are explicit proposed simulation/display settings, not material evidence. No editable numerical controls in this slice, so there is no need for a generic configuration form/schema editor.

Go returns:

- profile/model version, material preset name, temperature, assumptions/evidence notice and sample count;
- ordered samples with stable index and branch, canonical SI E/P values, Go-converted MV/cm and microcoulombs per square centimeter display/speakable forms;
- summary including model/method, field span, polarization extrema, and ascending/descending values at the sampled zero-field points, with correct units and clear simulation-default status;
- page position/count and complete row values for a requested fixed-size page (32 rows), derived in Go;
- optional bounded plot coordinates/axis labels from the same result, not a browser-recalculated dataset.

Extend the existing Go presentation vocabulary for field/polarization units with focused tests. Keep canonical scientific values separate from display conversions. No default synthesizer guess about Unicode units/exponents; the actual Orca reading is acceptance. If an expected sample/value is unavailable, say why rather than fabricating a zero. Do not infer coercive crossings or fitted quantities not actually computed by this first fixed operation.

## Worker and one-screen interaction

The current `validation/wasm/testdata/runner` is a private parity fixture, not the production Worker API. Introduce a thin production Go WASM entry point and fixed-operation adapter only during authorized slice implementation, reusing the existing solver/presentation code. Extend/retarget the parity harness so the actual production loop operation is tested natively, in Node and in headless Firefox before enabling its React action; preserve the original example-report gate too.

Use typed protocol version + request ID + operation (`hzo-loop`) and result/error messages. React provides lifecycle/focus/transport, never field generation, summaries, rounding, units or scientific defaults. One request at a time; ignore obsolete response IDs. Worker loading is lazy after the explicit “Compute HZO loop” action, not a prerequisite for reading the page. Startup/error/retry never trigger local-server probing. For this bounded request, completion is one final update; do not stream 257 individual announcements.

UI state: ready → loading model → computing → result or readable error/retry. Keep focus on the initiating control during background work. On result, announce the Go-generated completion/sample summary once without forcing focus to the plot. An explicit “Read results” link moves to the result heading/table. Request reset/discard honestly clears only ephemeral in-memory state. Worker failure leaves the disclaimer/help and retry controls usable.

Table is primary: caption includes “HZO loop, page N of M” and branch/context; row headers identify samples, columns identify branch, field/unit and polarization/unit. Every value has the Go speech form available through normal Orca table reading, not only an aria-label on an unvisited span. Page change focuses/announces caption. No virtualization; Go supplies pages/total counts. Evidence/assumptions remain readable text alongside results.

SVG is secondary: figure/text summary, axes/units, no point-by-point focus and at most one sequential Tab stop (zero is preferred for the static figure). There is no plot-only selection or new plot interaction. Repeating a computation yields the same rows/summary/plot for the same versioned profile.

Persistent notice: “This loop is computed in your browser using the Go simulation model; it is not measured device data. Project and result data is not uploaded. Disk projects, large sweeps, stored-run verification, EDA and file writes require the local application (fecim serve). This hosted build does not contact localhost.” Include local-run instructions as text; do not create a local engine discovery client.

## Bounded implementation order after review

1. **Go loop contract first:** add a focused failing loop-result/summary/unit test around the existing Preisach core, implement only the adapter/presentation additions, and extend native/Node/Firefox parity to the actual loop operation. Reuse and keep the example report gate. Run scientific/model regression tests; no coefficient/tolerance/golden changes.
2. **One hosted React screen:** add the minimum locked React/TypeScript/Vite setup and actual WASM Worker transport. Reuse React Aria only where native controls need richer behavior; use a native read-only table. Replace the public home with this shell when authorized, preserving required links/content without building more views. No local profile implementation yet, but do not fork route/component code in anticipation of it.
3. **Prove the one task:** keyboard/automated accessibility and no-upload/no-localhost checks, actual payload/cold-load/compute measurements against reviewed budgets, then the manual Orca task below. No additional model screen until this result and its manual evidence are reviewed.

These are dependent reviewable outcomes, not authorization to start. Do not add setup-only tickets, dummy UI placeholders or a second production numerical implementation. Exact file ownership/new protected manifests must be declared in the execution tasks before edits; no runtime/task dispatch is created by this plan.

## Acceptance and verification

- Real Go worker HZO output matches native/Node/Firefox for the fixed profile; invalid protocol/version/duplicate requests and worker errors do not create plausible-looking scientific results. Keep `TestExampleResultsJSONNativeWASM` as an independent persisted-study/report comparison.
- Actual table/summary speech and focus work in Orca + Firefox/Linux; automated scans alone cannot pass this gate. No raw symbol-only field/unit pronunciation is assumed.
- Browser network interception proves there are no localhost/private-network discovery calls or scientific-data uploads during loading/computation/retry/paging. Static same-origin asset fetches are allowed. No scientific persistence, export or file-writing feature is added.
- Actual first-slice payload includes its HZO solver plus React/Worker assets; remeasure, do not reuse the smaller example-only module receipt as a shipped-size claim. Implement reviewed budget assertions in the explicit CI gate rather than keeping them as a prose-only promise.
- No hosted feature is declared complete while the manual task or numerical parity is unrun. Fyne retirement remains a separate module replacement/removal change after qualification; this plan does not delete legacy code or claim all-module parity.

## Manual Orca task HZO-1 — first-slice acceptance

**Start state:** Hosted build from the candidate commit, Firefox + Orca on Linux; no local server. New page state with no saved result. Record exact tool/voice settings, Go/WASM/frontend fingerprints, tester/date and actual speech/focus. Keyboard only; no sighted guidance or developer-console rescue. Network capture may be collected separately by automation/observer, not used as a navigation aid.

**Goal:** Compute the HZO loop, understand its model/evidence/units and inspect any sample without interpreting the plot or running a local server.

**Steps:**
1. Read the page title, HZO heading and persistent disclaimer. Find local-run instructions without triggering a connection.
2. Activate “Compute HZO loop”. Continue reading/navigating while the Worker initializes/calculates; confirm focus is not stolen.
3. Read the completion and Go-generated summary, including method/temperature, sample count, field span, polarization extrema and two branch-specific zero-field values.
4. Follow “Read results”. Use Orca table commands to read sample/branch and both quantities with their headers. Read negative and zero values, µC/cm² and MV/cm units and a representative exponent form if produced. Compare actual speech to Go's defined wording.
5. Page forward/backward with keyboard, reading the focused caption/page position. Reach the final page and inspect its row without virtualized content disappearing.
6. Read the plot's accessible summary and Tab through the figure: no more than one plot stop, no per-point navigation and no unique result hidden there.
7. Recompute and inspect equivalent rows/summary. Exercise a controlled Worker error/retry fixture and reload/discard of ephemeral results; the page must never claim a persistent/server-saved run.

**Required Orca announcements:** descriptive heading/control labels; persistent local-computation/no-upload/server-required notice; nonintrusive loading/completion; model/evidence summary and correct expanded signs/units; table caption with actual page N of M; row/column context; error/retry without false science. One completion announcement, not 257 per-sample interrupts. No focus jump caused by background updates.

**Failure criteria:** any inaccessible data/units/branch distinction; visual-only summary/evidence; main-thread freeze that prevents control/navigation; unavailable values presented as zero; missing headers/page focus; table virtualization; plot-only information; a misleading upload/verification/persistence claim; localhost probe; or a worker failure that removes instructions/retry. Wrong unit/exponent pronunciation fails even if a computed accessible-name string looks correct.

**Current result:** NOT RUN — the React HZO slice is not implemented. Automated gate evidence for the eight-point study does not constitute this manual result. Only the human tester can attest the actual Orca task; no agent waiver or fabricated pass.

## Handoff

Before implementation, review the [prerequisite results and proposed budgets](../../testing/wasm-parity-gate.md). The next proposed action is the focused Go HZO adapter/parity RED test, not four UI views. This invocation reports results and saves this plan only; commits and slice implementation wait for owner review.

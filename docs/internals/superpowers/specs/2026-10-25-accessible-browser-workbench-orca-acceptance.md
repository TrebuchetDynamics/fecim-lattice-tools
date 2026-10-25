# Accessible browser workbench — Orca manual acceptance gate

Status: Required future acceptance scripts; **not executed**. The browser workbench is not implemented. This file specifies a release gate, not a test receipt or an implementation plan.

Related owner: [approved design direction and decision register](2026-10-25-accessible-browser-workbench-design.md). Update both documents together when the interaction contract changes. These scripts implement decisions BW-07 through BW-14; they do not reopen them.

## Test environment and evidence

Use Linux, Firefox and Orca, with keyboard only throughout the application tasks. Record exact OS, Firefox, Orca and speech synthesizer versions/settings, keyboard layout, app/core/frontend fingerprints and fixture revision. Record date, tester, task ID, actual speech, actual focus, result and an evidence reference. Speech transcripts must not contain launch tokens or unpublished project data. Accessibility automation is supporting evidence only.

Use normal Orca browse/table commands as well as Tab/Shift+Tab and native control keys. A task does not pass merely because tabbing reaches a control: table cell/header reading must work. Do not use a mouse, visual plot inspection, DOM inspector or another browser to rescue a failed task. A sighted observer may document behavior, but must not supply information unavailable through Orca. Terminal setup and fixture preparation are allowed prerequisites, not substitutes for browser interactions.

Quoted announcements below specify required content, not exact synthesis punctuation or a mandated verbose sentence. Required roles, names, units, exponents, counts and context must be intelligible without visual interpretation; duplicate speech, misleading precision or dropped signs fail. If actual speech differs harmlessly, record it and explain equivalence. A tester cannot waive missing content by guessing from the screen.

All required tasks must pass on the same candidate and documented supported environment. Any failure or unrun task prevents an Orca acceptance claim. After a fix, rerun affected tasks and any shared-focus/speech regressions; do not reuse invalidated evidence. There is no automated or agent-authored human sign-off.

## Future fixture requirements

The implementation must provide reproducible, explicitly synthetic acceptance fixtures and an expected-value manifest. They are not experimental measurements or physics validation evidence. Fixtures must not silently reuse changing production-model expectations.

- A granted workspace containing a valid starter study and a comparison study; an unrelated outside directory must remain inaccessible. Include spaces in a project name.
- A result fixture with 8 committed records, 6 feasible, 1 Pareto and 0 failed; a second terminal fixture with a nonzero design-point failure count. The prepared manifest supplies exact expected records, values and evidence.
- A paginated result fixture with at least three pages, sortable equal values (to check stable ID tie-breaking), filterable feasible/infeasible/failed results, missing metrics and distinct evidence labels. Set a small page size through a supported test configuration.
- Presentation fixtures containing `15 µC/cm²`, `1e-12 A`, `2e+3 Ω`, `0 pJ`, a negative quantity, a rounded quantity and unavailable values with reasons. Include a unit not supported by the speech vocabulary to verify explicit unavailability rather than guessed pronunciation.
- YAML fixtures with comments, an unexposed schema-supported field, nonstandard spacing/quoting, CRLF, missing final newline and a saved original-byte digest. Include external-edit scenarios with a known conflicting field.
- A controllable long experiment that can be paused at a test checkpoint, cancelled after some commits and resumed. Include controllable failure/report-write and completion/cancel races. Do not depend on a fast machine being slow enough for a human to press Cancel.
- A build-identity mismatch scenario for a stale cached UI, plus missing/stale asset scenarios for separate automated build/startup checks.

Fixture controls must be test-only, not an unauthenticated production endpoint. If fixtures or controls do not yet exist, record the affected task as **not run**, not passed. During real acceptance, replace symbolic page counts/values below with the fixture manifest's exact expectations.

## 1. Launch, authenticate and discover permitted projects

**Start state:** Server stopped; granted and outside directories prepared; Firefox and Orca running. Tester has terminal access.

**Goal:** Reach the application without persistent credentials or access outside granted roots.

**Steps:**
1. Launch with only the prepared workspace granted. Open the terminal's one-time launch URL using the keyboard.
2. Reach Projects and enumerate available projects through Orca.
3. In a separate unauthenticated Firefox context, replay the used launch URL and confirm authentication is not recreated by that token. Separately test a link older than 60 seconds in an unauthenticated context. Closing an authenticated tab alone does not revoke its session cookie.
4. Request a fresh launch link using the terminal renewal action and authenticate again without stopping the server.

**Required announcements:** Projects heading; names and selection controls for permitted projects; expired/used-link state and an understandable instruction to obtain a new terminal link. No outside project names or secret values in application announcements.

**Failure criteria:** Link expiry/replay is accepted; renewal requires killing running work; token remains in the address bar after bootstrap; unauthenticated project data is readable; outside roots can be browsed. Header/CSRF/path attack coverage also needs automated tests and is not certified by this manual task alone.

## 2. Navigate views and browser history

**Start state:** Authenticated Projects view; valid study open.

**Goal:** Identify each destination without searching the old view.

**Steps:** Navigate Projects → Configuration → Experiment → Results, then use Firefox Back and Forward. Repeat one transition with a deliberately delayed response, immediately navigating elsewhere before it completes.

**Required announcements:** Each completed navigation focuses and speaks the new descriptive view heading and heading level. The late response does not pull focus back or announce an obsolete view. Background updates do not move focus.

**Failure criteria:** Focus stays on a removed control, falls to the document body, moves to an obsolete view, or announces a heading twice through redundant live-region echo. Browser history loses the active project without an explanation.

## 3. Create, validate and correct a study

**Start state:** Authenticated Projects view within the granted workspace; unused destination name available.

**Goal:** Create a starter study and recover from invalid configuration without losing input.

**Steps:**
1. Create a study with a validated subdirectory name and configured starter sweep.
2. Open Configuration and use Orca to inspect field labels, units, help, hypothesis, objectives, constraints, citations and input hashes.
3. Submit at least two fixture-defined invalid values.
4. Follow each error-summary link, correct the value and submit successfully.

**Required announcements:** Configuration heading; complete field names and units; on failed submit, focus the summary and announce “Save failed”, error count and field-specific reasons. Links focus the correct fields. Success identifies the saved/validated state and Go-provided point count without focus theft.

**Failure criteria:** Errors are color-only, summary is not focused, corrections/other edits disappear, units are inaccessible, existing directories are overwritten, or browser-side calculation substitutes for authoritative validation.

## 4. Read values, units, exponents and evidence

**Start state:** Presentation fixture open in Results, including details and a plot summary.

**Goal:** Obtain the same quantitative meaning as a sighted reader without deciphering symbols.

**Steps:** Read fixture values with Orca table navigation, then details and relevant field help/conflict examples. Read metric/run headers and evidence with each value.

**Required announcements:** “15 microcoulombs per square centimeter”; “1 times 10 to the minus 12 amperes”; “2 times 10 to the plus 3 ohms”; “0 picojoules”; correct negative sign, approximation and evidence. Missing metrics say “Not available” and the reason; unsupported unit pronunciation is explicitly identified. Metric and run context remain available through table headers.

**Failure criteria:** Exponent sign, scale or unit is lost/mispronounced; unavailable becomes zero; speech claims extra precision; a symbol-only aria-label passes while native cell reading remains unintelligible; value is read twice; ARIA replacement hides headers or evidence. Capture actual speech, not an assumed accessible-name string.

## 5. Preserve YAML and perform a byte-identical no-op save

**Start state:** YAML fixture open in Configuration; original bytes/digest recorded by fixture preparation.

**Goal:** Save exposed fields without destroying comments/unexposed data, and make no changes on a no-op.

**Steps:**
1. Save without editing. Repeat with a field restored to its original semantic value before saving.
2. Change one exposed value and save.
3. Reload and inspect that value. A fixture verifier compares original/no-op bytes and confirms retained supported unexposed fields/comments after the changed save.

**Required announcements:** Saved/unchanged outcome, correct retained field value and no fabricated validation errors. Browser focus remains on a meaningful action/status, not an unrelated view.

**Failure criteria:** Either no-op changes any byte, line ending, quoting, whitespace or final newline; comments/unexposed values disappear after a changed save; UI submits a replacement YAML document instead of field changes; failed saves partially alter files. Exact byte preservation requires automated/verifier evidence in addition to speech; hearing “unchanged” is not proof.

## 6. Resolve a stale-edit conflict as text

**Start state:** Configuration loaded with a known ETag; unsaved edits present. Fixture helper changes a conflicting field on disk and an unrelated supported field after loading.

**Goal:** Understand and resolve concurrency without a visual YAML diff or silent overwrite.

**Steps:** Submit the stale edit. Read each conflict entry, including an unavailable/deleted-value case. Choose review/reapply against the current revision. Repeat with another external change before reapply, then exercise explicit reload/discard.

**Required announcements:** Conflict-view heading; readable field-by-field list such as “Energy limit: yours 50 picojoules; on disk 40 picojoules”; distinct missing versus zero values; clear reload/discard and review/reapply controls. A repeated conflict is explained without losing attempted edits.

**Failure criteria:** Visual diff is required; conflict fields or actions are unnamed; stale/no-op save overwrites disk; unrelated external changes disappear; browser silently force-saves. Record supporting `412`/ETag test evidence separately.

## 7. Start an experiment and hear bounded progress

**Start state:** Validated long-experiment fixture with known total, worker limit and a pause checkpoint.

**Goal:** Start computation while keeping control of navigation and speech.

**Steps:** Start the sweep, inspect state, then navigate among controls/read table content while the fixture produces frequent progress updates. Reload the page during execution and recover the active job.

**Required announcements:** Start once, point total and worker context; intelligible count progress at most once per five seconds during ordinary progress. Current focus does not move and speech does not become a queue of per-run notifications. Reload finds the active experiment and exposes a current snapshot, not a second execution.

**Failure criteria:** Every point interrupts speech, focus jumps, reload duplicates computation, browser disconnect cancels work, reused/failed counts are double-counted, or a lost SSE connection is announced as completion. Record timing observations.

## 8. Cancel, settle and resume

**Start state:** Long-experiment fixture paused after a known subset of committed runs, with unresolved work remaining.

**Goal:** Distinguish a cancellation request from settled cancellation and preserve completed work.

**Steps:** Request cancellation. Read status before the fixture releases its final worker. Release the checkpoint through the test helper, wait for terminal confirmation, inspect committed results and rerun. Also exercise the controlled completion/cancel race.

**Required announcements:** “Cancellation requested” while nonterminal; only after settled writes/workers, “Experiment cancelled”, actual partial run/feasible/Pareto/failure counts and remaining-not-run context. Resume reports reused work. If completion wins the race, announce the true completed outcome rather than both outcomes.

**Failure criteria:** Request acknowledgement claims finished cancellation; unevaluated work is counted as scientific failures; committed data disappears; unchanged completed points are needlessly rerun; terminal outcome is announced twice. Core artifact/reuse checks need automated evidence too.

## 9. Hear completion and systemic failure accurately

**Start state:** Controlled success fixture with expected 8/6/1/0 counts, a design-point-failure fixture and a systemic/report failure scenario.

**Goal:** Know whether work finished, what succeeded and which results remain available.

**Steps:** Complete each scenario while focus remains on an unrelated readable control. Reconnect SSE/reload the completed job and read its persistent status.

**Required announcements:** For the success fixture, “Experiment finished: 8 runs, 6 feasible, 1 Pareto, 0 failed”. For design-point failure, actual nonzero failure count. For systemic/report failure, explicit failure/reason and available committed results, not an unqualified successful finish. No repeated terminal interruption for the same terminal revision.

**Failure criteria:** Any required count is omitted or derived from plot marks; report failure is hidden as success; reconnect repeatedly announces completion; focus is stolen; failed-design-point results are discarded or hidden.

## 10. Recover after server restart and detect build mismatch

**Start state:** Open project with known committed runs and an active or completed execution; recorded build identities.

**Goal:** Recover trustworthy disk results after ephemeral session/execution state disappears.

**Steps:** Stop/restart the server through the test setup, reauthenticate, reopen the same project and revisit the old execution. Inspect persisted runs and offer/start a resumed sweep. Separately load a stale cached UI against the controlled mismatch scenario.

**Required announcements:** Connection/session loss with recovery action; after the old execution's `404`, explain that execution state is unavailable after restart but stored runs can be inspected. Mismatch warning states that the UI/core build pairing is unexpected and provides a reload/recovery path. Exact version/fingerprints can be read/copied for a report.

**Failure criteria:** UI invents old job success/cancellation, hides intact runs, claims interrupted compute is still active, clears disk data, silently uses incompatible assets, or shows only a hash-equality warning for intentionally different core/frontend fingerprints. A stale/missing asset startup refusal also requires separate automated coverage.

## 11. Sort, filter and page the primary run table

**Start state:** Multi-page fixture open in Results, at page 1 with the manifest-defined sort.

**Goal:** Read all runs through ordinary table navigation without virtualization.

**Steps:** Read headers/cells using Orca. Sort by a metric, move forward and back a page, apply a feasibility filter and inspect an empty filter result. Repeat after a background refresh and with equal-valued rows.

**Required announcements:** On page change, focus the table caption: “Runs, page N of M, sorted by [metric] [direction]”, with actual filtered scope. Sort/filter refreshes supply the updated caption; an empty result says no runs rather than a fake “page 1 of 0”. Background refresh does not steal focus. Missing metrics, evidence, feasibility and Pareto status are readable as text.

**Failure criteria:** Rows disappear because they were scrolled out of a virtual viewport; headers/context are lost; page change leaves focus on a removed row/body; equal values reorder unpredictably; a stale request announces the wrong page; a dataset revision change silently mixes pages.

## 12. Compare two to four runs across pages

**Start state:** Multi-page fixture with eligible, failed/missing-metric and incompatible-unit cases; no selection.

**Goal:** Compare stored scientific evidence without reading a graphic or relying on color.

**Steps:** Select one run and inspect the minimum-selection explanation. Select another on a different page, filter one selected run out of view, then open comparison. Expand to four distinct runs and attempt a fifth. Remove a selection through its summary control.

**Required announcements:** Selected count and 2–4 limit; hidden selections remain named/removable. Comparison heading after navigation. Native table reading identifies metrics as row headers and unambiguous runs as column headers, with units, speakable values and evidence. Failed/missing values have reasons; incompatible units are explicitly incomparable or use a Go-declared conversion.

**Failure criteria:** Selection follows row index rather than run ID, disappears on paging/filtering, silently replaces one of four selections, uses runs as rows contrary to the contract, omits evidence/units or reruns physics to compare. A single or fifth selection must produce an explanation without unexpected focus movement.

## 13. Inspect plot scope without point-by-point tabbing

**Start state:** Filtered multi-page results with plot and selected runs, plus a known missing/unplottable record.

**Goal:** Obtain every relevant finding without inspecting or navigating graphical points.

**Steps:** Read the Go-generated plot summary, scope and units. Tab through the plot area forward/backward. Select a run from the table, then inspect the plot's mirrored-selection text. Change page and filter and reread scope.

**Required announcements:** Axes/units, named result snapshot/filter scope, full filtered set versus explicitly labelled current-page-only data, omitted/unplottable count/reason and selected-run context. Summary explains relevant extrema/Pareto/feasibility findings for this workbench plot; do not invent curve crossings for a plot where none are defined. It directs users to exact table values. Any bounded representation is labelled rather than presented as all points.

**Failure criteria:** More than one sequential Tab stop inside a plot, per-point controls, a second pointer-only selection state, summary supplied by browser math, page-only plot implying whole-study scope, or findings/values available only visually. Zero stops is acceptable for a static figure with accessible text.

## 14. Verify provenance, export and close safely

**Start state:** Project with integrity-checkable stored results, at least one eligible export run and generated reports.

**Goal:** Inspect evidence and download artifacts without rerunning models or deleting project data.

**Steps:** Open details, invoke explicit verification, regenerate reports, download each supported report format, and export/download an eligible selected run. Attempt an ineligible export. Close the project after activity settles, then reopen it. Also attempt close during an active operation.

**Required announcements:** Detail heading, full run ID, model/assumptions/input hashes, evidence labels and integrity outcome; clear report/export success or failure; ineligible export reason. Close is labelled as closing access, not deleting data. Busy close explains that an operation is active; reopening reveals preserved files/results.

**Failure criteria:** Reading generates artifacts or reruns models; export loses source-run identity; arbitrary paths can be requested; downloads lack readable names/status; a busy close silently abandons computation; close deletes or implies deleting project data. Browser download handling must be keyboard-accessible.

## Results record

Record each script as `pass`, `fail`, or `not run`, with actual speech/focus and evidence. No result is entered by this design-authoring pass. Future implementation must additionally pass automated API/security, generated-contract/asset freshness, YAML preservation/concurrency, determinism, race and existing CLI/Fyne regression gates. These manual tasks are not replacements for those checks.

TDD: N/A — documentation-only acceptance specification. No production code, test fixtures, automated tests, dependencies, services, commits or releases are changed or performed here.

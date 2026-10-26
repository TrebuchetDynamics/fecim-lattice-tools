# React workbench: safe configuration editing frontier

Date: 2026-10-25
Status: Accepted

Acceptance here means a planner-derived engineering prerequisite inside the already approved [React/Go design](../specs/2026-10-25-accessible-browser-workbench-design.md), especially BW-05/BW-09/BW-10 and section 3. It does **not** approve execution, commits, frontend installation, a listener, publication or automatic human qualification. This standalone Plan For Me pass saves one bounded batch and stops; the previous no-commit restriction continues.

## Resume and project model

The product is a local-first, simulation-only FeCIM research workbench for researchers, graduate students and device engineers. Its journey is configure → validate → run/reuse → analyze → inspect evidence → report/export. React web is the primary UI target, Fyne is deprecated compatibility, and the Go core/CLI remain authoritative. The static Astro `web/` project is not the workbench client.

At inspection all 13 existing tasks are coordinator-verified closed with valid task-store receipts. Keep their IDs/history; no existing implementation is requeued. The root reader, progress channel, speech/display values, partial outcomes and instruction-alignment work have landed. The remaining gap selected here is Go-owned field editing and persistence; a full React app, end-to-end run/report/export confinement and HTTP authentication remain later work, not delivered by this batch.

### Goal classes and coverage inventory

| Goal / requirement | Evidence / class | State and disposition |
|---|---|---|
| React primary; deprecated Fyne, retained Go/CLI | DOCUMENTED: owner decision, BW-01..03, React-primary policy | Accepted direction. T-007/T-008 closed; preserve updated instructions. |
| Scoped project reads | OBSERVED: `workbench/project/load.go::LoadRoot`; T-001 receipt | Implemented/verified at its read boundary; not a complete persistence sandbox. |
| Progress and partial-result availability | OBSERVED: `experiment/progress.go`, `study.Run`, T-002/T-004/T-006 receipts | Implemented; reuse without changing scientific data or reopening tasks. |
| Go display/speech and legacy result presentation | OBSERVED: presentation package and T-003/T-005 receipts | Implemented primitives/legacy UI, not a browser/Orca pass. |
| Node-preserving field editing; no-op byte identity | DOCUMENTED: BW-10; OBSERVED: strict typed loader exists, no edit operation found in scoped `workbench/`/`internal/`/command inspection | Not implemented; task C1 plus C2. |
| ETags, stale-save field comparison, guarded persistence | DOCUMENTED: BW-05/BW-09; OBSERVED: current loader has no configuration revision/editor/save operation | Not implemented; C1 revision plus C2 persistence/conflict model. UI speech rendering remains future work. |
| Run/report/export filesystem confinement | OBSERVED: current run store still uses path-based persistence; DOCUMENTED: browser security contract | Incomplete; deliberately not presented as fixed by a configuration editor. Must precede a data-serving browser release. |
| Browser shell/API/authentication/assets and all twelve Orca tasks | DOCUMENTED: existing approved browser spec/manual owner; no React workbench package found | Not implemented/unverified; existing owners retain full scope. No new product strategy invented. |
| HZO evidence and recovery/release work | DOCUMENTED: existing HZO receipts, legacy TR-* backlog and release owners | Preserve history and outstanding human/source/rights gates; no scientific approval, citation repair or release claim here. |

No full-suite test or runtime experiment is run in this planning pass. “Verified” above is attributed to existing coordinator receipts, not a fresh test claim. New checks below are proposed/unrun.

## Source evidence and binding

Inspected HEAD: `bf754e4ea77a8fa24d77789d629a56498f258802`, plus dirty documentation from the preceding deprecation pass and the unrelated untracked EDA race test. These must not be overwritten. No active task was reported by preparation; the existing runner is running and is not started, stopped or reconfigured here.

| Input | SHA-256 |
|---|---|
| `workbench/project/load.go` | `d4bcd8846027c3c5ebe2b72984ceefe356f21dcc6e2295a23bc0b2ba88a11a8f` |
| `workbench/project/types.go` | `43622e3bdb90ed439ad5bdf636e869f65a9e12d34df280246fab6e89599f9bf2` |
| `workbench/project/load_test.go` | `d92fe88451367a5e3550ccec429b28a140fc2c615b39742b156b914b4d42e32e` |

Concrete observations:

- `LoadRoot` reads the three named YAML files, performs strict decoding with `KnownFields(true)`, validates design/objectives/constraints and calls Expand. It currently reads configuration directly from disk; validation cannot yet consume a staged edited document without writing it first.
- `decodeStrict` bounds configuration bytes using the existing `maxProjectFileBytes` constant. Preserve that actual-byte bound, unknown-field rejection and single-document rule.
- `project/types.go` identifies the scalar types and unexposed schema-supported fields which must survive edits. Do not regenerate YAML from these structs: that would lose comments/unexposed representation.
- `Expand` can allocate range values before checking the Cartesian budget. The new editor must preflight candidate range counts/value lengths against its validated max-points bound before invoking expansion. This is a prerequisite of safely validating this editor, not authority for a repository-wide solver rewrite.
- Primary local API evidence: Go `go1.26.2 linux/amd64`, `$GOROOT/src/os/root.go` documents caller-held traversal-resistant roots and relative OpenFile/Lstat/Rename/Remove operations. Roots follow nonescaping relative links and do not protect against same-user mount/device manipulation. Existing module/toolchain bounds remain unchanged; do not update dependencies for this plan.

Confidence is high for inspected API/control flow and documented requirements. Absence is limited to the inspected source areas; it is not a claim of an exhaustive repository bug hunt.

## Selected design and alternatives

**Planner's call:** add a Go configuration editor in `workbench/project`, using its existing YAML v3 dependency and caller-owned `*os.Root` handles. Keep public Load/LoadRoot behavior compatible and share full-bundle validation rather than copy scientific defaults/validators. Ship neither an HTTP route nor a browser form in this batch.

Rejected alternatives:

1. Browser-side YAML serialization loses unknown-to-the-form fields/comments and violates Go authority.
2. Typed-struct reserialization in Go also loses representation. Use parsed node trees and retain original bytes instead.
3. A raw whole-document editor or arbitrary JSON Patch broadens scope and permits structural/identity changes not required by the initial forms.
4. `EvalSymlinks`/prefix checks followed by path-based writes reopen check/open races. All editor filesystem actions use the supplied root, never Bundle.Root/Root.Name as authority.
5. Generic rename cannot provide compare-and-swap against hostile external writers. Promise revision checks and serialization for participating saves, not nonexistent cross-process CAS.

### Public seam C1 produces

Add in `workbench/project/configuration.go`:

- `ConfigurationSnapshot`: opaque original-byte snapshot of project.yaml, design.yaml and sweep.yaml; unexported storage, no caller-mutable maps/node pointers. `Revision() string` exposes its strong revision only. Read/copy methods needed by the implementation must not leak mutable internals.
- `ReadConfiguration(root *os.Root) (ConfigurationSnapshot, error)`: read only the three fixed regular files, bounded by the existing per-file limit. This captures bytes/revision, not proof of full semantic validity. Reject symlink/nonregular config paths; keep existing LoadRoot's legacy read behavior unchanged.
- `FieldChange{Path string, Value string}`: existing scalar path plus raw input text, not a YAML snippet, filesystem path or document. Go parses its type.
- `PrepareConfiguration(root, citations *os.Root, base ConfigurationSnapshot, document string, changes []FieldChange) (PreparedConfiguration, error)`: no writes; clone only owned state, patch the named document, validate the proposed entire bundle using a shared private overlay-aware loader, and return an opaque prepared result with `Changed() bool`. Preserve current input/citation verification through the supplied handles. Unknown document names are rejected.

`PreparedConfiguration` retains candidate bytes/snapshot and the validated bundle privately for the saver in the same package. Do not expose a public “commit arbitrary prepared object” path; C2 constructs preparation internally from a fresh snapshot.

Configuration document order is fixed: project.yaml, design.yaml, sweep.yaml. Revision is a quoted strong ETag `"cfg-v1-<lowercase SHA-256 hex>"`. Hash input is `fecim-config-v1` followed by a NUL byte, then for each document an unsigned 64-bit big-endian name-byte length, UTF-8 name bytes, unsigned 64-bit big-endian content-byte length and exact raw content bytes. Version/name/length framing prevents concatenation ambiguity. Comments, line endings and final newline affect revision.

### Field grammar and v1 allowlist

Use RFC-6901-style slash paths: mappings use their schema keys, existing sequences use canonical zero-based decimal indices. Decode only `~0` and `~1`; reject invalid escapes, negative/leading-zero indices, append markers, duplicate target paths and nonexisting nodes. Do not add/remove/reorder collections in this frontier.

Editable scalar paths:

- project.yaml: `/name`, `/hypothesis`, and existing `/constraints/N/value` or `/constraints/N/operator`.
- design.yaml: `/device/material`, `/device/conductance_levels`, `/device/g_min_s`, `/device/g_max_s`, `/array/rows`, `/array/cols`, `/array/read_voltage_v`, `/circuit/adc_bits`, `/circuit/dac_bits`, `/circuit/tia_gain_ohm`, `/circuit/tech_node`.
- sweep.yaml: `/seed`, `/max_points`, existing `/parameters/N/values/M`, or existing `/parameters/N/range/start`, `/stop`, `/count` under that range object.

The index placeholders above are numbers, not literal N/M keys. Identity/schema/model-version, objective definitions, citations/input hashes, constraint metric/unit and sweep parameter paths are unexposed/read-only; preserve them. This is a bounded first form-editing primitive, not the whole future form schema. Future structural editing must retain the same preservation/concurrency contract and be planned separately.

Parse integer fields as base-10 integers with overflow checks, float fields as finite float64 input and string fields as literal strings; never parse arbitrary YAML supplied in Value. Existing full validation remains authoritative for supported values/ranges. Limit a patch to 64 changes and each scalar input to 4096 UTF-8 bytes; enforce the existing configuration byte limit on output too. Empty patches and semantically unchanged parsed scalar values are no-ops. Do not normalize a no-op's original representation.

For changed saves, preserve node comments, unexposed values and ordering; normal encoder layout normalization is acceptable only for the changed document. The other two documents remain byte-identical. For nonempty edits, fail closed on alias/merge-bearing documents or ambiguous duplicate keys rather than guessing alias semantics. An empty patch can still be a no-op for a semantically valid alias-bearing bundle. Invalid/unknown-schema input is not silently stripped or made valid by dropping fields.

Full candidate validation uses a private overlay reader for fixed configuration filenames, with the existing root-bound input/citation reads. Preserve the common validator and strict decoder; do not create a second independent schema/default resolution. Preflight sweep max-points/range counts/value lengths before allocation on the editor path; oversized/malformed edits fail without disk mutation.

## C2: save, revision errors and field conflicts

Add `configuration_save.go`:

- `NewConfigurationEditor(root, citations *os.Root) (*ConfigurationEditor, error)` retains caller-owned handles without closing them. One editor instance is used for a project by the eventual adapter.
- `(*ConfigurationEditor).Save(ctx context.Context, expectedRevision, document string, changes []FieldChange) (ConfigurationSaveResult, error)` serializes participating saves with its private mutex and performs fresh snapshot/revision checks and C1 preparation.
- `ConfigurationSaveResult{Revision string, Changed bool}`; the result refers to the editor's committed bytes, not a guarantee against subsequent nonparticipating edits.
- Distinguishable `ErrRevisionRequired`, `ErrInvalidRevision`, and typed `ConfigurationConflict{CurrentRevision string, Fields []ConfigurationFieldConflict}` errors. Missing revision is required-precondition failure; weak, malformed or wildcard revisions are invalid, never an overwrite bypass. Future HTTP maps missing to 428, invalid to 400, stale to 412; no HTTP code is introduced here.
- `ConfigurationFieldConflict{Path, Yours, OnDisk string; OnDiskAvailable bool; Reason string}` gives one item per attempted scalar change in request order, without a colored/raw-document diff. Missing/unparseable on-disk fields have availability/reason, not fabricated zero. Yours is the submitted literal text; OnDisk is the safely resolved current scalar text. Error messages must not embed whole project contents/host paths.

These conflict values are a typed Go comparison primitive, **not yet accessible browser presentation**. Later Go presentation/HTTP code must supply labels, units and speakable forms before Orca acceptance; do not claim BW-09/BW-08 fully delivered by raw scalar fields. No project→presentation import is added: presentation already depends on experiment→project, so that would cycle.

Save sequence:

1. Lock this editor; observe context; capture the fixed three-file snapshot and require the exact strong revision, including for an empty patch. Stale revision returns field conflicts without a write.
2. Prepare/validate the candidate without writing. For a valid no-op, re-read/recheck bundle revision and return Changed=false; perform no encoding/replacement, preserving bytes and modification time.
3. For a change, create an unpredictable exclusive temporary regular file **under the same root** using root.OpenFile with O_CREATE|O_EXCL, restrictive permissions no broader than the existing file. Write/sync/close fully. Preserve the target's permission bits through safe opened-file handling, not path-based Chmod on a racing link. No host-global temporary file or root.Name-based write.
4. Re-read/recheck all three bytes/revision and regular-file conditions immediately before replacement; observe cancellation. On an observed conflict/precommit error, remove only this call's temporary file and leave original configuration unchanged.
5. Use root.Rename to replace the **one** named document. Success is the commit point; do not report an uncommitted cancellation merely because context changes afterward. Return the candidate revision/Changed=true. No other document/report/run/artifact is rewritten.

Testing uses a private filesystem-operation hook only where needed to deterministically inject precommit failure/external modification; production uses real root operations. Do not add a public test-only endpoint, sleeps or probabilistic retries.

### Concurrency and platform limits

The editor lock coordinates calls on that instance; an eventual project registry must reuse it and also coordinate save with active mutating operations. This batch does not claim coordination across duplicate editor instances, processes or editors ignoring the protocol. Recheck detects observed external edits, but a nonparticipating writer can race after the last check; document that limitation rather than promise CAS. The API requires a real granted Linux root for its initial tested scope, consistent with the Linux/Orca product target. Do not claim GOOS=js/WASM confinement or arbitrary filesystem/power-loss durability. Unsupported replacement semantics return an error without a successful-save claim; no dependency/toolchain/platform install is authorized here.

## Batch and acceptance

C1 and C2 serve one new goal, BWC-CONFIG, inside the accepted React workbench direction. C2 depends on C1's exact snapshot/prepare/revision seam; this is a real prerequisite, not a file-overlap dependency. C1 owns configuration.go/test plus the shared loader refactor/tests; C2 owns configuration_save.go/test. No worker writes the queue, manifests, locks, CI or scientific presets. Resources are empty; no device, fixed port, live account or browser stack is needed.

- C1 first RED: `workbench/project/configuration_test.go::TestPrepareConfiguration`, table-driven for framed revision, supported edits/full validation, no-op original bytes, comments/unexposed data, malformed/unsupported/oversized/alias paths and allocation preflight. Existing TestLoad* regressions guard compatibility.
- C2 first RED: `workbench/project/configuration_save_test.go::TestSaveConfiguration`, table-driven for byte-identical no-op (including CRLF/no final newline), mtime preservation, one-file commit, revision errors/stale text list, deterministic competing saves, external modification, escaping links and precommit failures/cancellation with original/temp preservation.

New selectors are **proposed/unrun**; no measured baseline is fabricated. Exact focused/package/race checks and exclusions are in the task records. Future execution retains the repository Full check and relevant race gates; no goldens, tolerances or scientific metadata are changed to pass them.

Qualitative metric/direction: supported field changes preserve unexposed configuration, valid no-op saves preserve exact bytes, and conflicts/errors never knowingly overwrite a changed bundle. No invented performance number. Use a bounded RED/GREEN/refactor cycle per task, not unlimited research iterations.

Pre-mortem: no-op encoder changes comments/CRLF → exact-byte/mtime test; a stale save clobbers another document's edit → all-three-file revision/recheck and competing-save tests; configuration validator drifts or allocates an excessive range → shared loader regressions and editor preflight negative case. Global invariant: scientific records, IDs, schemas/defaults, CLI and legacy Fyne compatibility remain unchanged.

## Deferred roadmap and handoff boundary

This is the complete **configuration editing frontier**, not a complete React migration plan. Run/report/export confinement, active job registry/idempotency, authenticated API routes/SSE, generated client contracts/assets, forms, pagination/compare/plots and all human Orca scripts remain in the approved specification. Revisit them with landed interface evidence; do not launch a service on the strength of this editor alone. No successor planning task/daemon is created by this small known batch.

Saved through `autowork tasks`, under `tasks/goals/BWC-CONFIG.md` (priority 3, focus false):

| ID | Outcome | Placement / dependency |
|---|---|---|
| `T-009-1bca6e59` | Prepare validated YAML field edits without losing comments | Now; independent |
| `T-010-1f9b8e44` | Protect YAML saves with revisions and atomic replacement | Next; depends on T-009 |

Complete contracts, exact checks, bounded touches and seam notes are in `tasks/items/<ID>.md`. Existing 13 task IDs, closure receipts, ranks and goals remain untouched. New records are working-tree planning artifacts, not committed selection/delivery. No focus goal is replaced. Builder finding: a fresh read-only reviewer returned **“nothing important left”** within its bounded review. This is source/planning feedback, not security/runtime qualification. Parent checks independently confirmed the producer/consumer dependency and disjoint write sets.

Final planning receipts: `autowork tasks validate --worktree .` **PASS** (15 tasks, 5 goals, no drafts; only the new BWC-CONFIG goal/two records pending). Plan links, record readback, unchanged inspected-source fingerprints, bounded touches/acceptance, true one-link dependency, disjoint ownership and unfocused goal **PASS**. All 13 prior closures remain preserved. Scoped whitespace checks **PASS**. No proposed product selector, full/race suite, persistence experiment or Orca task was executed; their status remains proposed/unrun.

Next action only after separate execution authorization: T-009's table-driven `TestPrepareConfiguration` RED test. T-010 waits for T-009's exact seam. All records remain uncommitted; HEAD-only selection/delivery is not claimed. Standalone planning stops here; no product code/tests/fixtures/dependencies, commits or runtime services are changed in this pass.

+++
schema = 1
id = "TOOL-PIN-DRIFT"
title = "Make the external-tool drift detector and Go pin agree with go.mod"
short = "Fix toolchain drift detection"
status = "done"
goal = "DEV-TOOLING"
section = "Next"
rank = 1000
depends_on = []
touches = ["tools/external/README.md", "scripts/toolchain/drift_detector.sh", "scripts/toolchain/lib/tool_versions.sh", "scripts/toolchain_pins_test.go"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Repair only the offline external-tool version drift detector and the inventory pins it reads. Make the Go pin consistent with go.mod (`go 1.26.10`) so a conforming Go reports `match` instead of `drift`, and make the scientific-stack row compare like with like (the documented numpy pin against the installed numpy) instead of the `python3` interpreter version. Preserve the detector's output columns, exit semantics and the other pins; add one focused offline test that fails when the Go pin and go.mod disagree. Benefit: the reproducibility guard distinguishes a conforming toolchain from real drift."
acceptance = ["On a host whose Go satisfies the corrected pin, `bash scripts/toolchain/drift_detector.sh` prints the `go` row as `match` (not `drift`), and tools/external/README.md's Go pin prefix-matches the `go` directive in go.mod; a new offline test fails when they disagree.", "The scientific-stack row compares the documented numpy pin against the installed numpy version (a `2.x.y` string), not the `python3` interpreter version, so it reports `match` when they are equal and `drift` when they differ; the checker's column format and nonzero-exit-on-drift/missing behavior are preserved.", "No tool is installed, upgraded or fetched; go.mod/go.sum and the other inventory pins are unchanged; `go test ./scripts -count=1` and the final `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` pass."]
checks = ["bash scripts/toolchain/drift_detector.sh", "go test ./scripts -run 'TestToolchainDrift|TestExternalToolPins' -count=1", "go test ./scripts -count=1", "git diff --check", "env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m"]
exclusions = ["Only the drift detector, its pin-extraction helper, the external-tool inventory pins and one new offline test. No tool installation/upgrade/fetch, network, go.mod/go.sum change, unrelated scripts/Makefile/CI wiring, instruction/task/goal/runtime/launch-default edits, or release/qualification claims.", "Preserve owner-only qualification tasks and HYS/WEB final planning stops. Recheck actual file/worktree ownership before implementation."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/TOOL-PIN-DRIFT.toml"
+++
- Sources: tools/external/README.md; scripts/toolchain/drift_detector.sh; scripts/toolchain/lib/tool_versions.sh; scripts/toolchain/check_tools.sh; go.mod

## Discovery evidence and user benefit

Round 20, current source at 8efebd1a. The repository's own read-only detector reports two false rows: `bash scripts/toolchain/drift_detector.sh` prints `go | 1.24 | 1.26.10 | drift` and `python | 2.1.3 | 3.12.14 | drift` (exit 1). `go.mod` declares `go 1.26.10` and README/AGENTS require Go 1.26.10+, so the `1.24.x` row in tools/external/README.md is stale; extract_pin turns it into `1.24`, which can never prefix-match `1.26.10`. The scientific-stack row is compared against the `python3` interpreter version (`3.12.14`) while its pin is `numpy==2.1.3`; pin and current describe different tools, so that row can never report a meaningful result for the documented stack. A version-drift guard that always reports drift cannot protect the documented reproducibility pins.

Source/callers/tests inspected: scripts/toolchain/drift_detector.sh (loop, prefix comparison), scripts/toolchain/lib/tool_versions.sh (get_current_version/extract_pin), scripts/toolchain/lib/toolchain_status.sh, scripts/toolchain/check_tools.sh, tools/external/README.md and tools/AGENTS.md (pin inventory), go.mod, and the existing scripts/*_test.go. No existing test invokes the detector; the Makefile `ci` target and workflows call check_tools.sh, not drift_detector.sh. On this host `go version` is go1.26.10 and numpy is 2.3.5.

## Admission and boundaries

Read AGENTS.md, README.md, CLAUDE.md, TODO Policies, BLOCKERS.md, task/goal records and read-only runner state. The runner service is active (running, paused_until=0) and is dispatching this discovery job; the `/home/xel/.config/pi-autowork/fecim-lattice-tools.paused` marker records a pause that a start 14 s later superseded and did not represent enforced repository-wide maintenance. No open task touches scripts/toolchain or tools/external. The four owner-only tasks (HYS-ORCA, T-013-a31914c8, WEB-ORCA, WEB-TOOLS) and HYS/WEB final planning stops remain intact. No Live authorization, network, tool installation, dataset or unavailable input is required. New low-priority goal DEV-TOOLING holds this independent developer-tooling repair and does not displace owner priorities.

TDD for queueing: N/A (new task metadata only). The worker must add a failing automated test first, record RED, then make the minimum production change and record GREEN plus the final offline check.

Next unexamined seam: `module6-eda` generated-artifact overwrite/dry-run behavior against the export instructions in docs/guides/cli-reference.md.

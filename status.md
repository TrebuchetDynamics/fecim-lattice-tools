# Status Report — Module 4 Circuits Physics-Correct Flow

Date: 2026-02-13

## Validation gates

- `go build ./...` : PASS (0 errors)
- `go vet ./...` : PASS (0 warnings)
- `go test -short ./module4-circuits/... ./validation/...` : PASS

## Test counts (short mode)
Command used for counting:
`go test -short -json ./module4-circuits/... ./validation/...`

Per-package counts (pass/fail/skip):

- `module4-circuits/cmd/circuits`: 0 / 0 / 0 (no test files)
- `module4-circuits/cmd/circuits-gui`: 0 / 0 / 0 (no test files)
- `module4-circuits/pkg/arraysim`: 76 / 0 / 0
- `module4-circuits/pkg/gpuperiph`: 15 / 0 / 0
- `module4-circuits/pkg/gui`: 195 / 0 / 0
- `module4-circuits/pkg/gui/unified/display`: 0 / 0 / 0 (no test files)
- `module4-circuits/pkg/gui/unified/ispp`: 0 / 0 / 0 (no test files)
- `module4-circuits/pkg/gui/unified/overlay`: 0 / 0 / 0 (no test files)
- `module4-circuits/pkg/gui/unified/sense`: 0 / 0 / 0 (no test files)
- `validation`: 52 / 0 / 0
- `validation/benchmarks`: 2 / 0 / 0
- `validation/calibration`: 1 / 0 / 0
- `validation/comparator`: 1 / 0 / 0
- `validation/configvalidator`: 57 / 0 / 0
- `validation/configvalidator/cmd/validate`: 0 / 0 / 0 (no test files)
- `validation/external`: 1 / 0 / 3 (3 skips are optional external tools)
- `validation/heracles`: 2 / 0 / 0
- `validation/integration`: 24 / 0 / 0

Totals (packages under command scope):
- Tests passed: 426
- Tests failed: 0
- Tests skipped: 3

## ngspice availability
- `which ngspice` returned no path (not installed on this host)
- Structural netlist validation is active; runtime ngspice comparison test auto-skips with message until installed.

---

# Status — Riju (Repo)

Date: 2026-02-15 17:13 CST (America/Monterrey)
Repo HEAD: `a8fb558`

## Current focus
Tier-1 Module 1 physics falsification stays mandatory + green; expand literature packs (next: MDPI 10.3390/ma13132968 digitization).

## Gates (fresh)
- `go build ./...` : PASS (exit 0)
- `go vet ./...` : PASS (exit 0)
- `go test -short -count=1 ./...` : PASS (exit 0)
- `go test -v -count=1 ./validation/literature -run TestModule1_PELoop_LiteratureBacked` : PASS
  - park2015_hzo_10nm (doi=10.1002/adma.201404531): Pr err=0.00%, Ec err=0.00%, RMSE/Ps=0.0000, areaErr=0.00%
  - cheema2020_superlattice_5nm (doi=10.1038/s41586-020-2208-x): Pr err=0.00%, Ec err=0.01%, RMSE/Ps=0.0002, areaErr=0.00%

## Blockers
- Race lane (`go test -race -short ./...`) not green due to pre-existing races (GUI/E2E + M1 renderer `running` bool without sync). Non-race Tier‑1 gates are green.

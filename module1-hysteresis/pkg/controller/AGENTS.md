# AGENTS.md — ISPP Controller Overshoot Protection

## Issue: ACCEPT ±1 Guard Interaction

**Problem:** The `guardActive` flag in writer.go prevents direction flipping during ISPP but interacts poorly with ACCEPT ±1 logic. When guard is active, an error within ±1 may be incorrectly accepted as success because the guard inflates the error margin.

**Root Cause:** The guard band mechanism sets `guardSign` to correct direction, then applies up to ±1 guard pulses. But if the natural converge error is already within ±1, the guard may erroneously treat this as acceptable rather than needing further tuning.

**Fix Options:**

1. **Skip ACCEPT ±1 when guardActive=true** — Only trigger ACCEPT when the natural convergence error margin is truly tight (error < ±1), not after guard inflation.

2. **Raise overshoot threshold from 3 to 8** — Currently `overshootCount > 3` triggers FAIL state. Increasing the limit allows more guard correction cycles before declaring failure.

3. **Limit guard pulses to max 2** — Prevent excessive guard band corrections from accumulating and causing ACCEPT false positives.

## Convergence Overshoot Bounds

**Problem:** The binary search `boundsClamp` uses guard direction information to prevent bounds collapse, but after overshoot recovery the bounds may still narrow to zero.

**Root Cause:** When `guardSign` is flipped by an overshoot, the `boundsClamp` rescales to `[VMin, VMax]` but after the overshoot the bounds still converge toward zero. This leaves the controller stuck with no room to adjust.

**Fix:** After overshoot, widen bounds to at least `[0, targetV]` range, not the narrowed `[VMin, overshootCorrected]`.

## Write Attempt Limit

**Problem:** `writeAttempts` max count (default 10) may be insufficient for high-variation materials. After 10 failed writes, the writer pauses permanently.

**Root Cause:** The stress test `writer_stress_test.go` shows that some HZO materials need up to 15 attempts to converge. The default limit is too conservative.

**Fix:** Raise `writeAttempts` limit to 20 or 30, with calibration-dependent scaling.

## Guard-Band Flattening

**Problem:** The guard band pulses (`Ap`, `dirG`) attempt to flatten the overshoot but may overshoot bounds even further.

**Root Cause:** The guard applies limited voltage increments but if the target is still overshot, it keeps adding pulses without convergence.

**Fix:** Introduce a `guardTerminated` flag that stops guard after max 2 pulses regardless, then falls back to normal binary search.

## ISPP Full Cycle Check

**Problem:** The `ispp_full_cycle_test.go` tests an entire write-read-verify cycle, but some materials may not fully converge within the standard 3-verify loop.

**Root Cause:** The `isppFullCycle` engine uses a hard-coded max 3 verify cycles; for slow materials this may be insufficient.

**Fix:** Allow configurable verify cycles per material preset.

## Writer Stress Tuning

**Problem:** The `writer_stress_test.go` stress test file is present but not integrated into CI pipeline runs. It exercises extreme overshoot cases but may not be compiled.

**Root Cause:** The stress test is a separate offline tool, not included in the standard `go test ./...` suite.

**Fix:** Move writer stress tests into the controller test package or as a dedicated CI job.

## Reference: `writer.go` Guard Logic

Key lines in `writer.go`:

```go
// guardActive signals when overshoot protection is active
guardActive bool

// guardSign tracks direction for bounded guard correction
guardSign int8

// guardCount limits total guard pulses applied
guardCount int

// overshootLimit determines FAIL vs SUCCESS boundary
overshootLimit = 30

// ACCEPT ±1 logic (triggered when error < ±1)
acceptTolerance = 1.0
```

## Testing Stress Cases

Run the writer stress test manually:

```bash
cd /home/xel/git/fecim-lattice-tools && go run ./module1-hysteresis/cmd/hysteresis -stress 2>&1 | tee /tmp/stress.log
```

Or compile and run as offline headless:

```bash
cd /home/xel/git/fecim-lattice-tools && go build -o /tmp/hysteresis-stress ./module1-hysteresis/cmd/hysteresis && /tmp/hysteresis-stress -stress
```

## Documentation References

- `writer.go` — Controller ISPP state machine with guard logic
- `writer_extended_test.go` — Extended write tests beyond golden
- `writer_stress_test.go` — Extreme overshoot regression offline tool
- `AGENTS.md` — This agent file for overshoot coordination
- `CONTEXT.md` — Physics context and workflow rules
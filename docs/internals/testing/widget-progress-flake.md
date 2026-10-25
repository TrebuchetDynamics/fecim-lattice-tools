# Tutorial progress test: scheduler-dependent assertion

Issue: [#22](https://github.com/TrebuchetDynamics/fecim-lattice-tools/issues/22).
Status: local fix awaiting staged-diff approval; issue remains open.

The old test slept 75 ms after starting four 50 ms tutorial steps, then expected
progress between 0.2 and 0.5. A loaded scheduler can legitimately delay the
tutorial goroutine beyond that observation, or advance it beyond the upper
bound. Passing on rerun did not resolve this dependency.

RED was reproduced deterministically by holding the first step's action;
the old deadline-based assertion failed with progress zero. The fix changes
only the test: an action/channel handshake holds the second step and the test
asserts exact progress 1/4. A five-second deadline only detects failure to reach
that boundary; it does not define expected progress. Cleanup stops the controller
and releases the held action even after a failure.

The regression remains in the default suite; it is not skipped or quarantined.
No tutorial playback behavior changed. Other older tutorial tests using sleeps
were not claimed fixed by this bounded correction.

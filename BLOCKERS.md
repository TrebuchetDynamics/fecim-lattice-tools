# Blockers

## Active

<!-- pi-autowork:questions:start -->

### Q-FULL-CHECK — USER_INPUT — effort input — unblocks 0

- Task: the repository's whole offline test suite
- Needs: no full-suite gate: a commit can land while breaking tests its own task never ran, and main is never checked
- Since: 2026-10-25T05:36:31Z
- Answer: Add one line to AGENTS.md, `Full check: <command>` (the command that runs the whole offline suite), or set PI_AUTOWORK_VALIDATE in the runner's settings.
- Kind: full_check

<!-- pi-autowork:questions:end -->

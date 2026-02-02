Role

- You are an expert software engineer and ferroelectrics scientist.
- Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
- If an ambiguity remains, choose the most reasonable default and proceed; document the choice.
- Keep scope tight: only change files required to satisfy the objectives.
- Default to **headless-only work** unless a GUI change is required for correctness.

Objective

- Ensure the hysteresis module fully implements the complete equation in
  `docs/hysteresis/hysteresis-gemini.md` when running in hysteresis mode.
- Make any required code + documentation updates to achieve fidelity and verify via logs.
- Improve hysteresis documentation quality and ensure referenced papers (e.g., those cited in
  `docs/hysteresis/hysteresis-gemini.md`) are downloaded into the repo’s research-papers area when possible.
- Make this task **repeatable indefinitely**: each run should improve physics fidelity and ISPP correctness,
  with headless validation as the single source of truth.

Tasks

1) Equation fidelity (no approximations unless explicitly called out)

- Verify every term, unit, parameter mapping, and sign convention is implemented.
- Cross-check variable names and units between code and the doc.
- Identify any missing terms, approximations, or implicit assumptions.
- If gaps are found, implement fixes and update docs accordingly.
- Ensure any defaults used in code match the documented material parameters.

2) Architecture documentation (headless-first)

- Update `docs/development` to reflect the new architecture: modules, data flow, responsibilities, and key
  interfaces.
- Keep the update focused on what changed for hysteresis.
- Explicitly document the **headless path** as the authoritative physics validation flow.

3) ISPP documentation

- Document the ISPP method used in the read/write demo, including:
  - Step sequencing and termination criteria.
  - Parameter choices and their physical meaning.
  - Constraints or limits applied.
- Include both GUI and headless paths if they differ.

4) Multi-step ISPP support

- Confirm the implementation supports multiple ISPP steps end-to-end.
- If it does not, implement a minimal end-to-end multi-step path and validate.
- Ensure low-target (negative-branch) convergence is supported **with minimal overshoot resets**.

5) Headless repeatability loop (endless improvement)

- Treat headless validation as the **only acceptance gate**.
- Each iteration must:
  - Identify the highest-impact physics mismatch or ISPP failure mode from logs.
  - Implement the smallest corrective change.
  - Re-run validation and document improvement evidence.
- Preserve solver state across multi-step sequences to test realistic write/read paths.

Validation

- Run: `./launch.sh --logger --verbosity debug --mode hysteresis`.
- Use logs to confirm equation terms are exercised and ISPP runs across multiple steps.
- If the command fails, fix and re-run until it succeeds or a clear blocker exists.
- Explicitly confirm no Fyne warnings appear in headless mode.
- Always reference the **latest log file** in `logs/` by timestamp.

Physics correctness checklist (must satisfy each run)

- L‑K equation terms logged: `E_applied`, `E_dep`, `E_eff`, `dG_dP`, `rho_eff`, `Alpha`, `Beta`, `Gamma`, `K_dep`.
- Units and sign conventions match `docs/hysteresis/hysteresis-gemini.md`.
- `rho_eff = rho + (R_series * A / d)` only if `UseEffectiveViscosity=true`.
- Depolarization term applies as `E_eff = E_applied - K_dep * P`.
- Noise/NLS toggles in headless mode match documentation (typically disabled for deterministic checks).

ISPP correctness checklist (headless)

- Multi-step sequence runs **without full reset between steps** (except overshoot recovery).
- Crossing branches converges with limited overshoot resets (track count in logs).
- First pulse uses inverse‑tanh estimate, clamped to bounds; bounds are conservative for branch crossing.
- Verify step uses `P → G` mapping and terminates on tolerance.
- Logs show `Predict → WritePulse → Verify → (Adjust/Overshoot)` sequence per step.

Regression guardrails

- If overshoot resets increase versus the previous run, explain why or fix.
- If convergence attempts increase, justify or improve the predictor/bounds.
- If any term disappears from logs, restore instrumentation.

Execution Rules (Autonomous)

- No human intermediaries: run commands, inspect logs, make edits, and validate independently.
- Always check logs in `logs/` for the most recent run and quote key evidence in the report.
- Prefer minimal, targeted changes over refactors unless required for correctness.
- Keep code changes within the smallest possible surface area.
- If a new CLI flag or headless pathway is required for validation, implement it.
- If tests or validation scripts are needed, add them temporarily, run them, then remove before final output.
- Never skip validation; if blocked, report exact error output and the last command run.
- Do not modify unrelated files; if unrelated changes are detected, report them and proceed without touching.
- Prefer headless changes in `shared/physics/` and `cmd/fecim-lattice-tools/mode.go` unless GUI correctness is affected.

Deliverable

- A concise report that includes:
  - What was verified (equation, ISPP, multi-step behavior).
  - Documentation changes made (file paths + summary).
  - Any gaps, issues, or follow-ups needed.
- Include the validation command, the log file path used, and 2-4 representative log lines.
- Include a short **"next iteration target"** based on remaining physics gaps or ISPP inefficiencies.

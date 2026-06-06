<!-- fecim-skills:start -->
## FeCIM Skills

When the user's request matches the trigger description below, follow the workflow in the linked file.

- **fecim-citation** — Verifies and formats FeCIM physics/measurement claims against the project's published-source list and docs/4-research/honesty-audit.md. Use when adding a numeric claim, accuracy figure, or device-parameter assertion to code, docs, PR descriptions, or commit messages. → see `tools/fecim-skills/fecim-citation/SKILL.md`
- **fecim-grill** — Relentlessly interviews the user about a proposed FeCIM physics, simulation, or GUI change before any code is written — covering source citation, educational-vs-validated framing, TDD-RED test, thread-safety, and honesty-audit alignment. Use when starting a non-trivial change to physics, ISPP, crossbar, or GUI logic. → see `tools/fecim-skills/fecim-grill/SKILL.md`
- **fecim-honesty-audit** — Enforces docs/4-research/honesty-audit.md policy by scanning PR diffs, READMEs, and presentation material for removed/unverified claims (87% MNIST, 30-states-as-fact, energy multipliers vs NAND/GPUs). Use before committing docs, PRs, or release notes that include accuracy or efficiency numbers. → see `tools/fecim-skills/fecim-honesty-audit/SKILL.md`
- **fecim-labtester** — Runs the FeCIM test matrix (full, race, module-scoped, coverage, golden regen) and interprets physics regression failures using the 5 known bug patterns. Use when running tests, debugging test failures, or regenerating physics golden files. → see `tools/fecim-skills/fecim-labtester/SKILL.md`
- **fecim-researcher** — Surveys FeCIM domain knowledge by searching references/, citations/, docs/4-research/, and the local Cognee KG, then synthesizes a cited research note. Use when investigating a physics topic, evaluating a paper, or grounding a design decision in literature. → see `tools/fecim-skills/fecim-researcher/SKILL.md`

For all skills above, the canonical body is the linked SKILL.md file. Read it before acting.
<!-- fecim-skills:end -->

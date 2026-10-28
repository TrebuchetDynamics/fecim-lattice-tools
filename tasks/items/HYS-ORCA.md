+++
schema = 1
id = "HYS-ORCA"
title = "Qualify React Hysteresis with real Orca in both deployment profiles"
short = "Real dual-profile Orca acceptance"
status = "open"
goal = "HYS-REACT"
section = "Needs decision"
rank = 1000
depends_on = ["HYS-UI", "HYS-LOCAL"]
touches = ["docs/internals/testing/react-hysteresis-orca.md", "docs/internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md"]
resources = ["orca-firefox-desktop"]
kind = "inspection"
ownership = "owner"
scope = "Perform plan manual task HYS-1 against the actual same candidate's hosted and local builds with keyboard-only Firefox/Orca on Linux. Record actual speech/focus/settings/fingerprints/tester/date/evidence and limitations. No agent-authored human sign-off or automated substitution. Resume when both builds exist and a real tester has the desktop session."
acceptance = ["Given the two real candidate profiles and supported tester setup, when HYS-1 is performed without sighted/DOM rescue, then every applicable step has actual speech/focus evidence and pass/fail recorded; an unrun profile, unit/context/focus failure or missing evidence keeps qualification unmet. Only real owner/tester qualification may satisfy the manual criterion."]
checks = ["Perform HYS-1 in cited plan and the applicable shared Orca gate; record actual candidate-bound evidence. Automated E2E results support but do not pass this check."]
exclusions = ["Go owns science/defaults/display/speech; simulation-only evidence; no hosted scientific upload/persistence/localhost probing; existing CLI/Fyne/default builds remain. No launch-default change, Fyne deletion, new physical model, golden/tolerance weakening, deployment or service changes.", "No public deployment/push, shared service changes or ungranted external operations; test servers use private ephemeral ports."]
+++

- Sources: docs/internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md#hys-1-manual-task-both-profiles; docs/internals/adr/0005-unified-react-hosted-and-local.md
- Produces: Actual owner/tester dual-profile HYS-1 qualification or precise failed/unrun steps. Manual closure alone is not verified acceptance.
- Consumes: HYS-UI/HYS-LOCAL actual candidate/assets and automated/parity/security evidence.
- Repo obligations triggered: Existing manual acceptance owner requires actual tester speech/focus; no invented sign-off. Manual evidence document is explicitly a future owner artifact, currently not present. No fixture, manifest or code changes.
- Next action: When candidate profiles and tester desktop are available, record setup and perform hosted HYS-1 before the local sequence.
- Verification status: New tests/commands are proposed/unrun; no implementation evidence or manual qualification is claimed by planning.
- Question: Missing external input is real dual-profile Orca/keyboard evidence from a tester; this is not a reopened product-design question. Keep Needs decision until that evidence exists; no waiver inferred.

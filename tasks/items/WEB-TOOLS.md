+++
schema = 1
id = "WEB-TOOLS"
title = "Record actual installed Icarus and Yosys web EDA structural qualification"
short = "Real installed-tool EDA proof"
status = "open"
goal = "WEB-ALL"
section = "Needs decision"
rank = 3000
depends_on = ["WEB-EDA"]
touches = ["docs/internals/testing/web-eda-installed-tools.md"]
resources = []
kind = "inspection"
ownership = "owner"
scope = "On an approved environment with actually installed Icarus/Yosys and permission to execute, qualify only the existing bounded authenticated local EDA freeform structural-tool flow. Use the actual server allowlist/8x8 fixture and matching artifacts; record actual versions, commands, artifact/manifest hashes, exit/log/truncation/cancellation and candidate identities. No install/network/PDK/container authority is granted; missing tools remain NOT RUN, not passed."
acceptance = ["Given approved installed executables and permitted existing local checks, when server-owned Icarus syntax/elaboration and Yosys structure commands operate on the same generated 8x8 artifacts, then complete actual version/command/exit/log and artifact/candidate evidence is recorded. Missing/failing/truncated results remain unavailable/fail/incomplete; private fake-tool tests never qualify installed tools.", "Given structural-only Placeholder/ESTIMATED_SCALAR/not-foundry-mapped artifacts, when evidence is summarized, then it claims neither science/signoff/LVS/foundry validation nor unsupported ngspice/OpenROAD/KLayout/Magic/Netgen/OpenLane flows. Optional installed-tool proof does not block independent text generation or unrelated module qualification."]
checks = ["Perform actual existing authenticated EDA 8x8 freeform generate and allowlisted tool actions; verify retained versions and complete server-owned command/log/exit and matching artifact/manifest hashes per workbench/modules/eda/tools.go and docs/guides/eda-web.md.", "Read back docs/internals/testing/web-eda-installed-tools.md with NOT RUN for any missing proof and structural-only limitations; git diff --check"]
exclusions = ["Preserve Go science/defaults/units/speech/identities, exact original example bytes, scientific tolerances/goldens, CLI and current default-built Fyne launch/guards. No hosted scientific upload/persistence/downloads/localhost probing, model or product-scope expansion.", "No release/deployment/publication/push, default switch, Fyne deletion, task receipt or human/tool qualification fabrication. No EDA tool/model/dataset/PDK/container installation/acquisition; only existing explicitly scoped pinned build/test dependency setup where needed."]
+++

- Sources: docs/tasks/WEB-NEXT-PLAN.md; docs/internals/superpowers/plans/2026-10-28-all-modules-web-roadmap.md#outcome-and-definition-of-done; docs/internals/adr/0005-unified-react-hosted-and-local.md#required-acceptance; docs/tasks/WEB-NEXT-PLAN.md#web-tools--real-installed-icarusyosys-structural-proof; docs/guides/eda-web.md; workbench/modules/eda/tools.go
- Question: Can the owner provide an approved environment with installed Icarus and Yosys and permission to execute the existing bounded local structural checks, or explicitly leave those tool qualifications unrun?
- Evidence: WEB-EDA attempt-1 report pending_live says neither actual executable was present; private native/browser process fixtures prove orchestration only. Actual tools.go allowlists only iverilog/yosys and marks other tools unavailable.
- Produces: Real installed-tool source/artifact-bound structural evidence, or exact remaining unavailable/incomplete state; no tool installation or false pass.
- Consumes: WEB-EDA real freeform generator/artifact and bounded process services; approved installed executables/permission. No Orca/shared-guidance prerequisite.
- Repo obligations triggered: ADR capability/rights boundaries; roadmap M6 provenance/process limits and docs/guides/eda-web.md tool matrix. Reuse typed server commands and existing one-process-group/job limit/private roots; no new resource/policy, executable script, model/PDK, fixture/generated/code changes. Preserve complete logs and stated warnings.
- Next action: After external inputs are available, record exact candidate/setup and run the existing iverilog -V and syntax/elaboration action plus yosys -V and server-owned structural action against matching generated artifacts. Do not run arbitrary client commands or install anything.
- Verification status: Real installed Icarus/Yosys qualification NOT RUN. No owner waiver or execution authority inferred from planning.

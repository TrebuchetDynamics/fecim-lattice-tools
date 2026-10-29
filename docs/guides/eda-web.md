# Browser EDA workflow (candidate)

Module 6 is an educational design-artifact workflow, not a tapeout or foundry-signoff tool. React presents controls and escaped text; Go owns configuration admission, compilation, provenance, hashes, structural checks and native jobs. The legacy Fyne launch default is unchanged. Real Orca + Firefox/Linux qualification is **NOT RUN**.

## Start locally

Use the pinned build tools documented in [web/README.md](../../web/README.md): Go 1.26.10, Node 24.19.0 and npm 11.17.0. Node is build-only.

```sh
make build-browser
./fecim-lattice-tools serve --root DIRECTORY
```

`DIRECTORY` must exist. Open the short-lived terminal launch link. Create or open a project in **Projects**, then follow **Home → Explore EDA Export** (`/modules/eda`). A static local-profile preview alone does not grant filesystem or tool access. Native requests retain the shared session, Host/Origin, CSRF, UI fingerprint and project mutation gates.

## Two separate origins

### Freeform builder

1. Choose rows, columns, operation mode and architecture. **Load Go defaults** explicitly replaces the editable JSON with native compiler/cell defaults.
2. Inspect or edit **Typed builder config (JSON)**. It contains the actual `compiler.ArrayConfig` and `config.CellConfig`. Conductance is **µS**, geometry is µm, pulse width is ns, supply is V and TIA gain is Ω. Defaults are educational assumptions, not device measurements.
3. Optionally choose a granted root and root-relative weights file. This read-only JSON numeric matrix is bound to an opaque resource, not uploaded or executed. It applies only to compute mode; its rows/columns must exactly match the configured array. Without weights, Go generates blank level-zero initialization.
4. Select **Generate all formats**. A new immutable export ID is created under the opened project's `exports/`. Dirty edits never change an earlier export. Storage, memory and compute modes and passive, 1T1R and 2T1R architectures use the existing Go compiler. This initial web profile accepts SKY130 educational defaults only; other technologies are explicitly rejected, not claimed as qualified.

Freeform publication aligns the compiler array's Verilog/DEF cell name with the generated cell LEF/Liberty/Verilog name. Compiler helpers determine array placement geometry (including 2T1R 1.38 × 3.40 µm defaults); resolved cell inputs are recorded separately. The web cell adapter declares an unused `LEVEL` parameter to accept compiler Verilog instance overrides; it does not add multilevel physics. The original legacy cell-generator example bytes remain unchanged. This does not establish a foundry mapping or extracted connectivity. Existing SPICE/compiler generators are reused, not replaced with TypeScript models.

### Verified selected-run export

Run a study in **Experiment**, select one stored run in **Results**, return to EDA and select **Export verified selected run**. You can also enter its immutable 64-character run ID. Go loads and integrity-verifies the successful stored run; it does **not** rerun science or use the freeform draft. The original resolved design is retained in the manifest. The exporter converts stored conductance **S → µS** with `× 10^6` in Go and uses its separate EDA technology defaults. The source circuit technology is explicitly **not mapped to a foundry PDK**.

Selected-run export retains the stricter **16384-cell** limit; freeform admission does not weaken it. An existing export destination is never overwritten. A selected-run bundle intentionally has no generated `cell.v`, so its structural review is DEF-only and the initial installed-tool workflow is unavailable for that bundle. It is not a cross-format/tool pass.

## Inspect, validate and download

The artifact table lists server-enumerated files, exact byte counts and SHA-256 hashes. **Inspect** reads escaped text in 64 KiB chunks; previous/next buttons reach all bytes. Netlists/logs/imports are never inserted as HTML or executable SVG. The table is the complete accessible alternative; no plot is required.

Freeform artifacts: `design.spice`, `design.v`, `design.def`, `cell.lef`, `cell.lib`, `cell.v`, `config.json`, `package.zip`, `manifest.json`. Selected-run artifacts omit `cell.v` and `config.json`. Packages contain the generated format/config files, not the completion manifest or the package itself. The external `manifest.json` hashes the package as well as the format files and is written **last**. Failed publication removes only recognized owned files. Names/paths are server-owned; clients cannot choose an output path.

**Validate structure** verifies the entire enumerated bundle's hashes, DEF required keywords/component counts and, for freeform bundles, cell names and port sets across LEF/Liberty/Verilog using the existing native structural oracles. This is **not** extracted LVS, physical DRC closure, timing characterization, SPICE device validation or silicon qualification. Compiler and cell artifacts retain **Placeholder**, **ESTIMATED_SCALAR** and not-foundry-mapped warnings. Hash checks detect changed artifact bytes; the manifest is not a signed attestation against an operator who can rewrite both project bytes and manifest.

Downloads recheck enumeration, bounded regular-file identity and bundle integrity through pinned project roots. They never regenerate missing files. Keep the completed export for recovery; do not delete the source run merely because artifacts exist.

## Optional tools: explicit evidence, no acquisition

| Workflow | Initial web capability | Evidence / limit |
| --- | --- | --- |
| Go generation | Local only | Real compiler/export APIs; resolved config, origin, hashes |
| Go structural review | Local only | DEF counts; freeform cell names/ports; not signoff |
| Icarus Verilog | Installed Linux executable only | Server-owned `-g2012 -t null design.v cell.v`; syntax/elaboration only |
| Yosys | Installed Linux executable only | Server-owned `read_verilog`, `hierarchy -check`, `check`; structure only |
| ngspice | Unavailable in this web profile | Needs a separately reviewed model/testbench/resource profile |
| OpenROAD / KLayout / Magic / Netgen | Unavailable in this web profile | PDK/deck/resources and qualification not authorized here |
| OpenLane / Docker | Unavailable in this web profile | No container pull, network, install or PDK acquisition |
| Hosted Learn/import | Read-only | Imported/unverified text; no generation, downloads or tool claims |

**Discover installed tools** distinguishes installed-but-not-run from unavailable. **Run installed tool** records the tool name, actual version output, server-owned command description, exit outcome and escaped combined stdout/stderr. Missing tools are **unavailable**, never passed. Exit zero only establishes completion of the stated structural command, not device or foundry qualification. Truncated logs are incomplete evidence, not a validation pass. No client argv, shell, Tcl, executable path, environment or mount is accepted. There is no automatic installation, image pull, network enablement or PDK discovery. The bounded process runner is Linux-only; other platforms reject tool execution explicitly.

Each process group is owned through completion/cancellation, including descendants. Version and command stdout/stderr share the same retained-byte budget. Failed/cancelled tool jobs retain their executable-name/version/command evidence as well as logs. The working directory is pinned by descriptor and contains copies of only the enumerated Verilog inputs. Jobs share the opened project's mutation lease. The staging ceiling is **monitored**, not an OS sandbox or hard adversarial disk quota. Installed tools remain trusted native programs; this workflow does not claim isolation from all host filesystem/network access.

## Limits and recovery

| Boundary | Server policy |
| --- | --- |
| Command/config JSON | ≤8 MiB |
| Granted weights file | ≤32 MiB, finite, exact matrix shape |
| Freeform dimensions | Each 1..256, ≤65536 total cells |
| Selected-run dimensions | ≤16384 cells |
| Publication | ≤64 enumerated artifacts, ≤32 MiB each, ≤128 MiB total |
| Combined command logs | ≤1 MiB retained, further bytes drained and truncation marked |
| Tool command / entire job | ≤10 minutes / ≤30 minutes |
| Tool staging | 512 MiB monitored ceiling; excess aborts and cleans staging |
| Module session | One active job, at most 16 retained job records |

Navigation/disconnection does not cancel an admitted job. Use **Cancel EDA job**, then **Reconnect authoritative state**. A cancellation acknowledgement is not a successful validation. A transport error leaves the outcome unknown; inspect authoritative job state before retrying. Previous completed artifacts remain labelled by their origin/export ID until another job succeeds. A cancellation racing completed publication can leave files on disk without a reported job success; inspect disk before retrying. Logout/server close cancels and drains module jobs before releasing roots. Reload loses the ephemeral module session; on-disk artifacts remain. No destination overwrite is offered.

## Hosted profile

Hosted `/modules/eda` provides Learn/context and explicit local-run instructions. File selection reads at most 8 MiB into browser memory and labels the text **Imported/unverified**. It does not generate, export, download, persist, upload scientific files or probe localhost. Ordinary same-origin static asset requests still occur. The EDA entry is build-time presentation metadata, not an artificial scientific WASM engine.

## Verification and qualification

`TestEDAWebWorkflow` exercises the native Provider/JobPort with a real typed compiler result and tampered artifact rejection. `TestEDAWebAdmission` covers mode/topology generation and input/artifact policy. `TestEDAWebToolBoundary` uses private task-owned process fixtures for missing-tool, log, cancellation and staging behavior. Those fixtures are **not real-tool passes**. Existing selected-run root-confinement and Go format/GUI topology oracles remain in the focused checks. `web/test/e2e/eda.spec.ts` runs the real hosted build and authenticated native local build for keyboard generation, freeform/stored-run origin, artifact inspection, package download and escaped hosted import. Its private browser process fixture also exercises cancellation after a polling outage; that is orchestration evidence, not real Icarus qualification.

See the task report for executed commands/results. Real Icarus/Yosys version/resource qualification is **NOT RUN** when tools are absent; no tools/PDKs were acquired to turn that absence into a pass. Real Orca + Firefox/Linux speech/focus review for **both profiles** and any model/PDK/signoff qualification remain separate owner-operated gates.

Implementation owners: [builder/export policy](../../workbench/edaexport/builder.go), [native jobs/tools](../../workbench/modules/eda/port.go), [tool runner](../../workbench/modules/eda/process_linux.go), [Screen](../../web/src/modules/eda/Screen.tsx), [shared native contracts](../../workbench/moduleapi/native/projects.go).

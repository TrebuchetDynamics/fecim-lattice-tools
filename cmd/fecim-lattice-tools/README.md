# fecim-lattice-tools commands

Running without a subcommand retains the legacy Fyne desktop launch. Existing
`hysteresis`, other module and workbench CLI entry points remain available;
`--help` lists them. The additive browser command is:

```sh
fecim-lattice-tools serve --root DIRECTORY [--root ANOTHER_DIRECTORY] [--port PORT]
```

The design shorthand `fecim serve` does not name another installed executable.
Use [installation instructions](../../docs/guides/installation.md#optional-local-react-hysteresis-application)
and `make build-browser` for a binary with matching embedded assets. Plain Go
builds need no generated files and report `browser-assets-unavailable` on this
command. Browser users do not need Node. The listener binds only 127.0.0.1;
there is no wildcard/proxy/CORS mode. No default-launch switch occurs.

## Session and recovery

Open the terminal launch URL. Its fragment secret is removed immediately and
exchanged once within 60 seconds for a random HttpOnly, SameSite=Strict,
host-only session cookie. The per-launch cookie is not stored by application
JavaScript. Type `renew` then Enter in the terminal to rotate the launch link
without restarting; a successful exchange replaces the current local session.
Logging out or stopping the process revokes the session. Restart loses handles
and authentication, not already published files. No request URL/body/cookie
logging is installed. Do not share launch links.

Host authority (including port), Origin, Fetch Metadata and JSON/custom-header
checks protect session and file routes. Public static code contains no granted
file data. CSP permits the existing digest-checked Blob Worker/runtime and WASM,
not external scripts or frames. This is plain loopback HTTP, not TLS or a sandbox
against malicious same-user software, extensions, hard links or mount changes.
Cookies are not port-isolated; hostile services on the same local host remain
outside the trust boundary.

## Explicit Hysteresis files

Open **Local files and exports**, select a launch-granted root, and enter a
relative file path with an existing parent directory. Nothing auto-saves.
All reads/writes use traversal-resistant `os.Root` handles. Absolute, traversal,
ambiguous/reserved names, escaping links and outside grants are rejected.
Requests and file reads are capped at **8 MiB before decoding**; file size is
checked before reading and again on the opened handle.

- **Preset:** save/load existing Hysteresis preset JSON (`shared/presets.Preset`).
  Legacy material, waveform and `preisach`/`lk` engine vocabulary is retained.
  Backward-compatible `stress_pa`, decimal-string `seed` and `history_limit`
  keys preserve complete loop settings; older presets use Go defaults for absent
  settings. Unsupported waveforms/engines/materials fail instead of being guessed.
  Load explicitly applies the validated settings and resets in-memory history.
- **Calibration:** save a fresh matching browser simulation table as the existing
  version-4 multi-temperature JSON format. Load validates and displays version 4
  and legacy version-1 single-temperature tables
  **read-only/imported-unverified**, never as a fresh programming calibration.
  Explicitly recalibrate in memory before programming; loading is not verification.
- **Exports:** `pe-json`, `pe-csv`, `pe-svg` and `log-csv` (retained loop sample
  log), or complete `analysis-json`, `pund-csv`, `forc-sweep-csv`,
  `forc-density-csv`. Pause/cancel first. The current history view determines loop
  rows. Go generates serialization and SVG geometry; arbitrary HTML/SVG input
  is not accepted. Diagnostics retain their independent model/profile identity;
  supplied references remain unverified. Exports are **browser simulation data**,
  never verified persisted studies or device measurements.

Writes stage a complete private sibling, sync it, then publish by exclusive
root-relative hard link. A pre-existing destination returns **409 Conflict**;
choose a new path. No overwrite, autosave, deletion or rename-based compare-and-swap
against external editors is offered. Failed publication removes staging files.
This is not a power-loss durability guarantee or protection against same-user
malware. Hysteresis file snapshots remain unverified browser simulation data.
The separate research workflow below can explicitly verify real stored runs.

## Rooted research workflow

The same application now provides **Projects → Configuration → Experiment → Results**.
Create/open only beneath a launch grant. Explicit field saves use exact ETags:
missing `If-Match` returns 428, stale returns 412 with literal yours/on-disk text.
A project mutation gate returns 409 for conflicting work. Runs use the accepted
validated input snapshot and existing scientific identities, not browser-supplied
results. Cancel acknowledgement is not terminal cancellation. Navigation or a lost
stream never cancels server jobs. Reconnect, or reauthenticate/reopen after restart,
then inspect integrity-checked stored runs. Reports and eligible selected-run
exports are explicit operations; GET never regenerates them.

Shutdown/logout cancels and drains root-owning work before closing roots. Session
revocation also rejects delayed old-session requests after body decoding, including
old logout requests against a replacement session. Execution snapshots are bounded
at 64 with four active jobs; eviction returns 410, a pre-restart ID returns 404.
Neither implies deletion of committed runs. SSE has four independent slots and
bounded socket writes. It carries counts, not secrets or a durable event log.

See the [research guide](../../docs/guides/research-design-workbench.md#browser-research-workflow-candidate)
for the full workflow, field bounds, result selection and hosted limits. The
source-built candidate is not a qualified release. Human Orca acceptance is
**NOT RUN**. Default Fyne startup remains unchanged.

## Source-owned HTTP contract

Go DTOs and the internal Huma/stdlib router in
[workbench/moduleapi](../../workbench/moduleapi/) generate the
[OpenAPI contract](../../workbench/moduleapi/openapi.json). The server's existing
security wrapper owns authentication, bootstrap and legacy files. This API is
local-only; the hosted profile does not construct it. No automatic public docs,
schema or PATCH route is enabled.

| Route | Authentication/input | Output |
|---|---|---|
| POST /api/v1/session | Single-use JSON `secret` | Session cookie |
| POST /api/v1/session/logout | Session; empty JSON object | Revocation |
| GET /api/v1/roots | Session | Opaque grant IDs/names, admitted kinds, UI/core identity |
| POST /api/v1/files | Session; typed `FileRequest` | Save confirmation or imported typed data |

Research routes add project discovery/open/create/read/configuration PUT,
validate/close, asynchronous experiments/state/events/cancel, paged runs/analysis,
inspection/verification, explicit report/export/download operations and module
sessions/jobs. Only the route-specific pagination/filter keys in the generated
contract are allowed; roots/files/session routes still reject every query.
Native resources bind explicit granted regular files to opaque read references
(up to 256 MiB); they do not enlarge the 8 MiB JSON body ceiling or confer a
browser path/open capability. Each scientific port must enforce stricter limits.

Every POST and configuration PUT requires exact canonical Origin, `Content-Type: application/json`
and `X-FeCIM-Request: 1`. Reads reject foreign Origin when supplied, but allow
cookie-authenticated same-origin GET without Origin. File requests bind
`identity` and `UI`, `root` ID, relative `path`, enumerated `kind`,
`operation` (`save` or `load`) and typed `snapshot` for saves. Load is admitted
only for preset/calibration. Unknown fields, malformed/oversize input and unknown
roots/kinds fail. No token is accepted in URLs/query parameters. Typical statuses:
400 invalid typed input/path, 401 session/link required, 403 Host/Origin/CSRF,
409 existing destination/fingerprint/mutation conflict, 412 stale revision,
413 size limit, 428 missing revision, 503 bounded admission/shutdown. API errors
use safe RFC 9457 `application/problem+json`; success bodies are typed JSON or
enumerated downloads. Secrets and arbitrary host paths are not reflected in errors.

The hosted profile never constructs this transport, probes localhost, uploads
science, or persists files. Both profiles share React routes/components and the
same browser Go Worker. See [validation evidence and remaining manual gate](../../docs/internals/testing/wasm-parity-gate.md#6-authenticated-embedded-local-profile).

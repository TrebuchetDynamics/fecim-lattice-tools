import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { researchOwner } from "../../api/owner";
import type { Schema } from "../../api/types";
import type { Local } from "./local";
import "./eda.css";
type Artifact = { name: string; bytes: number; sha256: string };
type Result = {
  id: string;
  origin: string;
  warning: string;
  artifacts: Artifact[];
};
const local = __FECIM_PROFILE__ === "local";
const field = (s: Schema<"Snapshot"> | undefined, id: string) =>
  s?.fields?.find((f) => f.id === id)?.input || "";
function parsed<T>(text: string, fallback: T): T {
  try {
    return JSON.parse(text) as T;
  } catch {
    return fallback;
  }
}

export default function Screen() {
  useSyncExternalStore(researchOwner.subscribe, researchOwner.getSnapshot);
  const [adapter, setAdapter] = useState<Local>(),
    [snapshot, setSnapshot] = useState<Schema<"Snapshot">>(),
    [draft, setDraft] = useState(""),
    [error, setError] = useState(""),
    [status, setStatus] = useState("Not generated; external tools not run"),
    [busy, setBusy] = useState(false),
    [job, setJob] = useState(""),
    [jobState, setJobState] = useState("");
  const [rows, setRows] = useState("8"),
    [cols, setCols] = useState("8"),
    [mode, setMode] = useState("2"),
    [arch, setArch] = useState("passive"),
    [run, setRun] = useState(""),
    [tool, setTool] = useState("iverilog"),
    [artifact, setArtifact] = useState(""),
    [offset, setOffset] = useState(0),
    [imported, setImported] = useState(""),
    [root, setRoot] = useState(""),
    [weightPath, setWeightPath] = useState("");
  const alert = useRef<HTMLParagraphElement>(null);
  const importSequence = useRef(0);
  useEffect(() => {
    if (!local) return;
    let current = true;
    void researchOwner
      .connect()
      .then(() => import("./local"))
      .then((m) => m.connect())
      .then((a) => {
        if (current) {
          setAdapter(a);
          setSnapshot(a.snapshot);
          setDraft(field(a.snapshot, "config"));
          setRoot(a.roots[0]?.id || "");
        }
      })
      .catch((e) => {
        if (current) setError((e as Error).message);
      });
    return () => {
      current = false;
    };
  }, []);
  useEffect(() => {
    if (error && alert.current?.getClientRects().length) alert.current.focus();
  }, [error]);
  const result = parsed<Result>(field(snapshot, "result"), {
    id: "",
    origin: "",
    warning: "",
    artifacts: [],
  });
  const active = ["queued", "running", "cancelling"].includes(jobState);
  const locked = busy || active;
  useEffect(() => {
    setArtifact("");
    setOffset(0);
  }, [result.id]);
  const project = researchOwner.config?.id || "";
  async function action(action: string, input: Record<string, string> = {}) {
    if (!adapter || !snapshot) return;
    setError("");
    setBusy(true);
    try {
      const next = await adapter.apply({
        action,
        input,
        revision: snapshot.revision,
      });
      setSnapshot(next);
      if (action === "defaults") setDraft(field(next, "config"));
      setStatus(next.status);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  async function start(action: string) {
    if (!adapter || !snapshot) return;
    setError("");
    setBusy(true);
    setStatus("Submitting local EDA job; not yet generated or validated");
    try {
      const resources: Schema<"ResourceReference">[] = [
        { id: project, kind: "project" },
      ];
      if (action === "generate_all" && weightPath)
        resources.push(await adapter.weights(root, weightPath));
      let j = await adapter.start({
        action,
        revision: snapshot.revision,
        resources,
        input: { run_id: run || researchOwner.selected[0] || "", tool },
        ...(action === "generate_all" ? { payload: JSON.parse(draft) } : {}),
      });
      setJob(j.id);
      setJobState(j.state);
      while (["queued", "running", "cancelling"].includes(j.state)) {
        setStatus(j.state + " — disconnect/navigation does not cancel the job");
        await new Promise((r) => setTimeout(r, 100));
        j = await adapter.job(j.id);
        setJobState(j.state);
      }
      setStatus(
        j.state +
          " — " +
          (j.message || "Inspect authoritative artifacts and tool evidence"),
      );
      setSnapshot(await adapter.refresh());
    } catch (e) {
      setError((e as Error).message);
      setStatus(
        "Request rejected or transport lost; no success claimed. Reconnect authoritative state before retrying.",
      );
    } finally {
      setBusy(false);
    }
  }
  async function reconnect() {
    if (!adapter) return;
    setError("");
    try {
      if (job) {
        const j = await adapter.job(job);
        setJobState(j.state);
        setStatus(j.state + " — " + (j.message || ""));
      }
      setSnapshot(await adapter.refresh());
    } catch (e) {
      setError((e as Error).message);
    }
  }
  async function readImport(file: File | undefined) {
    const seq = ++importSequence.current;
    setImported("");
    setError("");
    if (!file) return;
    if (file.size > 8 * 1024 * 1024) {
      setError("Read-only import exceeds 8 MiB.");
      return;
    }
    try {
      const text = await file.text();
      if (seq === importSequence.current) setImported(text);
    } catch {
      if (seq === importSequence.current) setError("Import could not be read.");
    }
  }
  function preview(name: string, next: number) {
    setArtifact(name);
    setOffset(next);
    void action("preview", { artifact: name, offset: String(next) });
  }
  return (
    <section className="section eda-workbench">
      <div className="eda-kicker">MODULE 06 / LOCAL DESIGN ARTIFACTS</div>
      <h1>EDA</h1>
      <p className="eda-notice">
        Educational models only. Placeholder cell Verilog and ESTIMATED_SCALAR
        characterization are not device validation, extracted LVS, signoff or
        foundry-mapped silicon.
      </p>
      {error && (
        <p ref={alert} role="alert" tabIndex={-1}>
          {error}
        </p>
      )}
      <section aria-labelledby="eda-learn">
        <h2 id="eda-learn">Learn and qualification</h2>
        <p>
          SPICE describes circuit instances; Verilog describes interfaces;
          Liberty provides estimated timing; LEF describes cell geometry; DEF
          describes array placement. Cross-file names/ports and DEF counts are
          structural checks, not physical verification.
        </p>
        <p>
          Freeform builder configurations use compiler conductance in
          microSiemens (µS). Verified selected-run exports retain the stored
          scientific design and convert S → µS in Go, without rerunning science.
          Freeform defaults are separate from run-derived inputs.
        </p>
        <p>
          Installed optional tools are never installed or downloaded by this
          page. Icarus and Yosys syntax/structure jobs have a bounded Linux
          profile. ngspice, OpenROAD, KLayout, Magic, Netgen and OpenLane
          require separately authorized model/PDK/resource profiles and are
          unavailable here.
        </p>
        <p>
          Local use: build with <code>make build-browser</code>, then{" "}
          <code>fecim-lattice-tools serve --root DIRECTORY</code> and open the
          terminal launch link. Open/create a project in{" "}
          <a href="/projects">Projects</a>; select stored runs in{" "}
          <a href="/results">Results</a>.
        </p>
      </section>
      {!local && (
        <section>
          <h2>Hosted read-only inspection</h2>
          <p>
            Generation, export, filesystem writes and external tools require the
            authenticated local application. Hosted EDA does not upload
            artifacts, probe localhost or persist scientific data.
          </p>
          <label htmlFor="eda-import">
            Import artifact for read-only inspection
          </label>
          <input
            id="eda-import"
            type="file"
            accept=".spice,.v,.lib,.lef,.def,.json,.txt"
            onChange={(e) => void readImport(e.target.files?.[0])}
          />
          {imported && (
            <>
              <p>Imported/unverified — no generation or tool validation</p>
              <pre data-testid="eda-import">{imported}</pre>
            </>
          )}
        </section>
      )}
      {local && (
        <>
          <p role="status" data-testid="eda-status">
            {status}
          </p>
          <p>
            Opened project:{" "}
            {researchOwner.config?.name ||
              project ||
              "None — open a project first"}
          </p>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              void action("defaults", { rows, cols, mode, architecture: arch });
            }}
          >
            <fieldset disabled={locked || !adapter}>
              <legend>Freeform builder / Go defaults</legend>
              <div className="eda-config-grid">
                <label>
                  Rows
                  <input
                    name="rows"
                    value={rows}
                    onChange={(e) => setRows(e.target.value)}
                    inputMode="numeric"
                  />
                </label>
                <label>
                  Columns
                  <input
                    name="cols"
                    value={cols}
                    onChange={(e) => setCols(e.target.value)}
                    inputMode="numeric"
                  />
                </label>
                <label>
                  Operation mode
                  <select
                    name="mode"
                    value={mode}
                    onChange={(e) => setMode(e.target.value)}
                  >
                    <option value="0">Storage</option>
                    <option value="1">Memory</option>
                    <option value="2">Compute</option>
                  </select>
                </label>
                <label>
                  Architecture
                  <select
                    name="architecture"
                    value={arch}
                    onChange={(e) => setArch(e.target.value)}
                  >
                    <option value="passive">Passive</option>
                    <option value="1t1r">1T1R</option>
                    <option value="2t1r">2T1R</option>
                  </select>
                </label>
              </div>
              <button type="submit">Load Go defaults</button>
            </fieldset>
          </form>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              void start("generate_all");
            }}
          >
            <fieldset disabled={locked || !adapter}>
              <legend>Explicit typed cell and array configuration</legend>
              <label htmlFor="eda-config">Typed builder config (JSON)</label>
              <textarea
                id="eda-config"
                name="config"
                rows={18}
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                spellCheck={false}
              />
              <p>
                1..256 rows and columns, at most 65536 cells; exactly matched
                finite weights; 8 MiB JSON. SKY130 educational defaults only.
                Editing does not change retained artifacts until a new job
                succeeds.
              </p>
              <details>
                <summary>
                  Optional granted weights file (compute mode only)
                </summary>
                <label>
                  Granted root
                  <select
                    value={root}
                    onChange={(e) => setRoot(e.target.value)}
                  >
                    {adapter?.roots.map((r) => (
                      <option key={r.id} value={r.id}>
                        {r.name}
                      </option>
                    ))}
                  </select>
                </label>
                <label>
                  Root-relative weights file
                  <input
                    value={weightPath}
                    onChange={(e) => setWeightPath(e.target.value)}
                  />
                </label>
                <p>
                  Read-only JSON numeric matrix, at most 32 MiB. Empty means use
                  typed weights/default blank initialization. No upload or
                  arbitrary model/script execution.
                </p>
              </details>
              <button type="submit" disabled={!project}>
                Generate all formats
              </button>
            </fieldset>
          </form>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              void start("selected_run");
            }}
          >
            <fieldset disabled={locked || !adapter}>
              <legend>Verified selected-run export / separate origin</legend>
              <label>
                Stored run ID
                <input
                  name="run"
                  value={run}
                  onChange={(e) => setRun(e.target.value)}
                  placeholder={
                    researchOwner.selected[0] ||
                    "Select a run in Results or enter its immutable ID"
                  }
                />
              </label>
              <p>
                Requires a successful integrity-verified stored run, at most
                16384 cells. Does not use the freeform editor or rerun science.
                Existing destinations are never overwritten.
              </p>
              <button
                disabled={!project || !(run || researchOwner.selected[0])}
              >
                Export verified selected run
              </button>
            </fieldset>
          </form>
          <div className="eda-actions">
            <button onClick={() => void reconnect()} disabled={!adapter}>
              Reconnect authoritative state
            </button>
            <button
              disabled={!adapter || !job || !active}
              onClick={() =>
                void adapter
                  ?.cancel(job)
                  .then((j) => {
                    setJobState(j.state);
                    setStatus(j.state);
                  })
                  .catch((e) => setError((e as Error).message))
              }
            >
              Cancel EDA job
            </button>
          </div>
          {result.id && (
            <section aria-labelledby="eda-artifacts">
              <h2 id="eda-artifacts">Artifacts and provenance</h2>
              <p data-testid="eda-origin">
                Origin: {result.origin}. Export ID: {result.id}.{" "}
                {result.warning}
              </p>
              <div className="eda-table-scroll">
                <table>
                  <caption>
                    Enumerated completed artifacts; SHA-256 integrity checked by
                    Go
                  </caption>
                  <thead>
                    <tr>
                      <th scope="col">Artifact</th>
                      <th scope="col">Bytes</th>
                      <th scope="col">SHA-256</th>
                      <th scope="col">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {result.artifacts.map((a) => (
                      <tr key={a.name}>
                        <th scope="row">{a.name}</th>
                        <td>{a.bytes}</td>
                        <td>
                          <code>{a.sha256}</code>
                        </td>
                        <td>
                          {a.name !== "package.zip" && (
                            <button
                              disabled={locked}
                              onClick={() => preview(a.name, 0)}
                            >
                              Inspect {a.name}
                            </button>
                          )}
                          <a
                            download={a.name}
                            href={`/api/v1/projects/${encodeURIComponent(field(snapshot, "project"))}/exports/${encodeURIComponent(result.id)}/${encodeURIComponent(a.name)}`}
                          >
                            Download {a.name}
                          </a>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <h3>Escaped artifact text</h3>
              <p>{snapshot?.fields?.find((f) => f.id === "preview")?.label}</p>
              <pre data-testid="eda-preview" tabIndex={0}>
                {field(snapshot, "preview") || "Choose an artifact to inspect"}
              </pre>
              {artifact && (
                <div className="eda-actions">
                  <button
                    disabled={locked || offset === 0}
                    onClick={() =>
                      preview(artifact, Math.max(0, offset - 65536))
                    }
                  >
                    Previous text chunk
                  </button>
                  <button
                    disabled={
                      locked ||
                      offset + 65536 >=
                        (result.artifacts.find((a) => a.name === artifact)
                          ?.bytes || 0)
                    }
                    onClick={() => preview(artifact, offset + 65536)}
                  >
                    Next text chunk
                  </button>
                </div>
              )}
              <div className="eda-actions">
                <button
                  disabled={locked}
                  onClick={() => void action("validate")}
                >
                  Validate structure
                </button>
                <button disabled={locked} onClick={() => void action("tools")}>
                  Discover installed tools
                </button>
              </div>
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  void start("tool");
                }}
              >
                <label>
                  Installed structural tool
                  <select
                    value={tool}
                    onChange={(e) => setTool(e.target.value)}
                  >
                    <option value="iverilog">Icarus Verilog</option>
                    <option value="yosys">Yosys</option>
                  </select>
                </label>
                <button disabled={locked}>Run installed tool</button>
              </form>
            </section>
          )}
          <section>
            <h2>Validation and tool evidence</h2>
            <pre data-testid="eda-tool">
              {field(snapshot, "tool") || "Not run"}
            </pre>
            <p>
              Logs retain at most 1 MiB and continue draining; truncated
              evidence is incomplete. Commands have 10-minute and jobs 30-minute
              limits. Staging is monitored at 512 MiB, not an OS sandbox or hard
              disk quota. Missing tools are unavailable, never passed. Reload
              loses the module session; completed artifacts remain in the opened
              project.
            </p>
          </section>
        </>
      )}
    </section>
  );
}

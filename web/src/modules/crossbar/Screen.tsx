import { useEffect, useRef, useState } from "react";
import { Value } from "../../research/Shared";
import { claimScience, releaseScience } from "../../worker/ownership";
import {
  connect,
  type Transport,
  type Update,
  type Command,
} from "./transport";
import type { Schema } from "../../api/types";
import "./crossbar.css";

const choices: Record<string, string[]> = {
  profile: ["basic", "enhanced", "parasitic", "fecap"],
  architecture: ["0T1R", "1T1R", "2T1R"],
  ir: ["true", "false"],
  sneak: ["true", "false"],
};
const modes = [
  "matrix",
  "vectors",
  "ir",
  "sneak",
  "accuracy",
  "fecap",
  "stack",
  "literature",
  "solver",
];
function Table({ table }: { table: Schema<"TablePage"> }) {
  return (
    <div className="table-scroll">
      <table id={"crossbar-" + table.id}>
        <caption>{table.caption} · Simulation only</caption>
        <thead>
          <tr>
            {table.columns?.map((label, i) => (
              <th key={i} scope="col">
                {label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {table.rows?.map((row, i) => (
            <tr key={i}>
              {row?.map((c, j) =>
                j === 0 ? (
                  <th scope="row" key={j}>
                    <Value value={c.presentation} />
                  </th>
                ) : (
                  <td
                    key={j}
                    className={
                      table.id === "matrix" && j === 2
                        ? "state-cell"
                        : undefined
                    }
                    style={
                      table.id === "matrix" && j === 2
                        ? ({
                            "--cell-state": c.value,
                          } as import("react").CSSProperties)
                        : undefined
                    }
                  >
                    <Value value={c.presentation} />
                  </td>
                ),
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
export default function Screen() {
  const [data, setData] = useState<Update>(),
    [drafts, setDrafts] = useState<Record<string, string>>({}),
    [error, setError] = useState(""),
    [pending, setPending] = useState(false),
    [loading, setLoading] = useState(true),
    [unknown, setUnknown] = useState(false),
    [palette, setPalette] = useState("blue"),
    [layers, setLayers] = useState<string>(),
    [layer, setLayer] = useState<string>();
  const [row, setRow] = useState("0"),
    [col, setCol] = useState("0"),
    [weight, setWeight] = useState(""),
    [files, setFiles] = useState(""),
    [path, setPath] = useState("crossbar.json"),
    [project, setProject] = useState(""),
    [projectID, setProjectID] = useState(""),
    [root, setRoot] = useState("");
  const transport = useRef<Transport | undefined>(undefined),
    owner = useRef({}),
    alive = useRef(true),
    generation = useRef(0),
    tail = useRef<Promise<unknown>>(Promise.resolve()),
    poll = useRef<ReturnType<typeof setTimeout> | undefined>(undefined),
    lastAnnouncement = useRef(0),
    alert = useRef<HTMLDivElement>(null);
  const [announcement, setAnnouncement] = useState("");
  const active = !!data && ["running", "cancelling"].includes(data.job.state),
    busy = pending || active || loading || unknown;
  const value = (id: string) =>
    drafts[id] ?? data?.snapshot.fields?.find((f) => f.id === id)?.input ?? "";
  const report = (message: string) => {
    if (alive.current) setError(message);
  };
  const publish = (next: Update) => {
    if (!alive.current) return;
    setData(next);
    setUnknown(false);
    const active = ["running", "cancelling"].includes(next.job.state);
    if (!active) releaseScience(owner.current);
    if (!active || Date.now() - lastAnnouncement.current >= 5000) {
      setAnnouncement(next.job.message || next.snapshot.status);
      lastAnnouncement.current = Date.now();
    }
  };
  const enqueue = <T,>(fn: () => Promise<T>) => {
    const work = tail.current.then(fn);
    tail.current = work.catch(() => {});
    return work;
  };
  const watch = () => {
    clearTimeout(poll.current);
    const token = generation.current;
    poll.current = setTimeout(() => {
      void enqueue(async () => {
        if (token !== generation.current || !transport.current) return;
        try {
          const next = await transport.current.read();
          publish(next);
          if (["running", "cancelling"].includes(next.job.state)) watch();
        } catch (e) {
          setUnknown(true);
          report(
            "Execution outcome unknown; previous results retained. " +
              (e as Error).message,
          );
        }
      });
    }, 50);
  };
  const act = async (
    action: string,
    input: Record<string, string> = {},
    extra: Partial<Command> = {},
  ) => {
    setPending(true);
    setError("");
    try {
      await enqueue(async () => {
        if (!transport.current || !data)
          throw new Error("Crossbar transport is not ready");
        const command: Command = {
          action,
          input,
          revision: data.snapshot.revision,
          ...extra,
        };
        if (action === "run") {
          claimScience(owner.current);
          try {
            const next = await transport.current.start(command);
            publish(next);
            if (["running", "cancelling"].includes(next.job.state)) watch();
          } catch (e) {
            // A lost native response is not confirmation that Go stopped.
            setUnknown(true);
            throw new Error(
              "Execution outcome unknown; reconnect authoritative state. " +
                (e as Error).message,
            );
          }
        } else publish(await transport.current.action(command));
      });
      return true;
    } catch (e) {
      report((e as Error).message);
      return false;
    } finally {
      setPending(false);
    }
  };
  useEffect(() => {
    alive.current = true;
    const token = ++generation.current;
    void connect()
      .then(async (t) => {
        if (!alive.current || token !== generation.current) {
          await t.close();
          return;
        }
        transport.current = t;
        setRoot(t.roots?.[0]?.id || "");
        publish(await t.read());
      })
      .catch((e) => report((e as Error).message))
      .finally(() => {
        if (alive.current) setLoading(false);
      });
    return () => {
      alive.current = false;
      generation.current++;
      clearTimeout(poll.current);
      const t = transport.current;
      transport.current = undefined;
      // Release only after native draining or Worker termination confirms close.
      void tail.current
        .then(() => t?.close())
        .then(() => releaseScience(owner.current))
        .catch(() => {});
    };
  }, []);
  useEffect(() => {
    if (error && location.pathname === "/modules/crossbar")
      alert.current?.focus();
  }, [error]);
  const apply = async () => {
    const submitted = { ...drafts };
    if (await act("configure", submitted))
      setDrafts((old) =>
        Object.fromEntries(
          Object.entries(old).filter(([key, v]) => submitted[key] !== v),
        ),
      );
  };
  const mode =
      data?.snapshot.fields?.find((f) => f.id === "mode")?.input || "matrix",
    matrix = data?.snapshot.tables?.find(
      (t) =>
        t.id ===
        (mode === "ir"
          ? "ir"
          : mode === "solver"
            ? "solver-cells"
            : mode === "stack"
              ? "stack-cells"
              : "matrix"),
    );
  const shown = (id: string) =>
    id === "selection" ||
    (mode === "matrix" && id === "matrix") ||
    (mode === "vectors" && ["input", "output"].includes(id)) ||
    (mode === "accuracy" && ["output", "accuracy"].includes(id)) ||
    (mode === "ir" && ["ir", "ir_solver"].includes(id)) ||
    (mode === "sneak" && ["sneak", "sneak-trace"].includes(id)) ||
    (mode === "fecap" && ["matrix", "output"].includes(id)) ||
    (mode === "stack" && id === "stack-cells") ||
    (mode === "solver" && ["solver", "solver-cells"].includes(id));
  return (
    <section className={"section crossbar-lab palette-" + palette}>
      <p className="eyebrow">M2 / CPU ARRAY LAB</p>
      <h1>Crossbar</h1>
      <p>
        Program a cell, choose a model, inspect the Go-owned matrix and vectors.
        Heatmaps are not required to select or read any cell.
      </p>
      <p className="boundary">
        Simulation only.{" "}
        {__FECIM_PROFILE__ === "hosted"
          ? "Go runs in a dedicated browser Worker; science is ephemeral and is not uploaded or persisted. File writes require fecim-lattice-tools serve. Ordinary same-origin static asset requests still occur."
          : "Authenticated same-origin Go service; file actions require an explicitly opened, rooted project. No automatic downloads."}
      </p>
      <p data-testid="crossbar-status">
        {loading
          ? "Loading Go crossbar…"
          : data?.snapshot.status || "Crossbar unavailable"}
      </p>
      <span
        className="sr-only"
        role="status"
        aria-live="polite"
        aria-atomic="true"
      >
        {announcement}
      </span>
      {error && (
        <div
          className="error error-summary"
          role="alert"
          tabIndex={-1}
          ref={alert}
        >
          <h2>Crossbar action failed</h2>
          <p>{error}</p>
          <p>Previous scientific results and unapplied drafts are retained.</p>
          <button onClick={() => setError("")}>Dismiss error</button>
        </div>
      )}
      <div className="actions">
        <button
          disabled={pending || loading}
          onClick={() => {
            setPending(true);
            setError("");
            void enqueue(async () => {
              try {
                if (!transport.current)
                  throw new Error("No session to reconnect");
                const next = await transport.current.read();
                publish(next);
                if (["running", "cancelling"].includes(next.job.state)) watch();
              } catch (e) {
                report((e as Error).message);
              } finally {
                setPending(false);
              }
            });
          }}
        >
          Reconnect crossbar state
        </button>
        <button
          disabled={pending || loading}
          onClick={() => {
            setPending(true);
            setError("");
            clearTimeout(poll.current);
            void enqueue(async () => {
              try {
                await transport.current?.close();
                releaseScience(owner.current);
                transport.current = await connect();
                setDrafts({});
                publish(await transport.current.read());
              } catch (e) {
                setUnknown(true);
                report((e as Error).message);
              } finally {
                setPending(false);
              }
            });
          }}
        >
          Discard and restart crossbar
        </button>
      </div>
      {unknown && (
        <p>
          Execution outcome unknown. Scientific ownership is retained until
          authoritative state or confirmed session closure.
        </p>
      )}
      <div className="crossbar-layout">
        <div>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              void apply();
            }}
          >
            <fieldset disabled={busy}>
              <legend>Execution profile</legend>
              {data?.snapshot.fields
                ?.filter(
                  (f) =>
                    ![
                      "selected_row",
                      "selected_col",
                      "mode",
                      "job_id",
                      "job_state",
                      "retained_profile",
                      "layer_count",
                      "selected_layer",
                    ].includes(f.id),
                )
                .map((f) => (
                  <div className="field" key={f.id}>
                    <label htmlFor={"crossbar-field-" + f.id}>{f.label}</label>
                    {choices[f.id] ? (
                      <select
                        id={"crossbar-field-" + f.id}
                        value={value(f.id)}
                        onChange={(e) =>
                          setDrafts((d) => ({ ...d, [f.id]: e.target.value }))
                        }
                      >
                        {choices[f.id].map((v) => (
                          <option key={v}>{v}</option>
                        ))}
                      </select>
                    ) : f.id === "input" ? (
                      <textarea
                        id={"crossbar-field-" + f.id}
                        value={value(f.id)}
                        onChange={(e) =>
                          setDrafts((d) => ({ ...d, [f.id]: e.target.value }))
                        }
                      />
                    ) : (
                      <input
                        id={"crossbar-field-" + f.id}
                        value={value(f.id)}
                        onChange={(e) =>
                          setDrafts((d) => ({ ...d, [f.id]: e.target.value }))
                        }
                      />
                    )}
                    <small>
                      Go value: <Value value={f.presentation} />
                      {f.unit && <> · {f.unit}</>}
                    </small>
                  </div>
                ))}
              <button>Apply configuration</button>
            </fieldset>
          </form>
          {Object.keys(drafts).length > 0 && (
            <p>
              Unapplied settings. Retained results do not describe these drafts.
            </p>
          )}
          <div className="actions">
            <button
              disabled={busy || Object.keys(drafts).length > 0}
              onClick={() => void act("run")}
            >
              Run MVM
            </button>
            <button
              disabled={!active || data?.job.cancellation_requested || pending}
              onClick={() => {
                setPending(true);
                void enqueue(async () => {
                  try {
                    publish(await transport.current!.cancel());
                    watch();
                  } catch (e) {
                    report((e as Error).message);
                  } finally {
                    setPending(false);
                  }
                });
              }}
            >
              Cancel MVM
            </button>
            <button disabled={busy} onClick={() => void act("reset")}>
              Reset seeded array
            </button>
          </div>
          {active && (
            <p>
              Go progress: {data?.job.progress.completed} of{" "}
              {data?.job.progress.total}; {data?.job.state}. Cancellation is not
              confirmed until Go settles.
            </p>
          )}
          <fieldset disabled={busy}>
            <legend>Cell addressing / zero-based</legend>
            <label>
              Cell row
              <input value={row} onChange={(e) => setRow(e.target.value)} />
            </label>
            <label>
              Cell column
              <input value={col} onChange={(e) => setCol(e.target.value)} />
            </label>
            <button onClick={() => void act("select", { row, col })}>
              Select cell
            </button>
            <label>
              Cell state
              <input
                value={weight}
                onChange={(e) => setWeight(e.target.value)}
                aria-describedby="cell-state-help"
              />
            </label>
            <small id="cell-state-help">
              Normalized state from 0 to 1. Go applies the selected model's
              programming rules.
            </small>
            <button onClick={() => void act("program", { row, col, weight })}>
              Program selected cell
            </button>
          </fieldset>
        </div>
        <div>
          <label htmlFor="crossbar-mode">Analysis mode</label>
          <select
            id="crossbar-mode"
            disabled={busy}
            value={mode}
            onChange={(e) => void act("view", { mode: e.target.value })}
          >
            {modes.map((m) => (
              <option key={m}>{m}</option>
            ))}
          </select>
          <label htmlFor="crossbar-palette">Matrix palette</label>
          <select
            id="crossbar-palette"
            value={palette}
            onChange={(e) => setPalette(e.target.value)}
          >
            <option value="blue">Blue</option>
            <option value="amber">Amber</option>
            <option value="plain">Plain</option>
          </select>
          {data?.snapshot.evidence?.map((text, i) => (
            <p className="crossbar-evidence" key={i}>
              {text}
            </p>
          ))}
          {data?.snapshot.tables
            ?.filter((t) => shown(t.id))
            .map((t) => (
              <Table key={t.id} table={t} />
            ))}
          {["ir", "sneak", "solver", "fecap"].includes(mode) &&
            !(mode === "fecap"
              ? value("retained_profile") === "fecap"
              : data?.snapshot.tables?.some((t) => t.id === mode)) && (
              <p>
                Not available in the retained result. Choose a compatible model
                profile and run explicitly; unavailable analysis is not a
                computed zero.
              </p>
            )}
          {(mode === "matrix" ||
            mode === "ir" ||
            mode === "solver" ||
            mode === "fecap" ||
            mode === "stack") &&
            matrix && (
              <div className="actions">
                <button
                  disabled={busy || matrix.page <= 1}
                  onClick={() =>
                    void act("view", { mode, page: String(matrix.page - 1) })
                  }
                >
                  Previous cells
                </button>
                <button
                  disabled={busy || matrix.page >= matrix.total}
                  onClick={() =>
                    void act("view", { mode, page: String(matrix.page + 1) })
                  }
                >
                  Next cells
                </button>
              </div>
            )}
          {mode === "stack" && (
            <fieldset disabled={busy}>
              <legend>Illustrative stack layers</legend>
              <label>
                Layer count
                <input
                  value={layers ?? value("layer_count")}
                  onChange={(e) => setLayers(e.target.value)}
                />
              </label>
              <label>
                Selected layer
                <input
                  value={layer ?? value("selected_layer")}
                  onChange={(e) => setLayer(e.target.value)}
                />
              </label>
              <button
                onClick={() =>
                  void act("stack", {
                    layers: layers ?? value("layer_count"),
                    layer: layer ?? value("selected_layer"),
                  })
                }
              >
                Apply stack selection
              </button>
              <p>
                Layer data is the Go-owned seeded gradient/noise visualization
                from the native stack workflow, not measured material or
                inter-layer electrical coupling. Every layer and cell is
                accessible without rotation, a heatmap or subsampling.
              </p>
            </fieldset>
          )}
          {mode === "literature" && (
            <div>
              <h2>Model and literature context</h2>
              <p>
                The 30-state programming baseline is educational, not measured
                hardware. IR uses wire-resistance relaxation; SOR exposes
                convergence and residuals. FeCAP accumulates charge, not DC
                current.
              </p>
              <p>
                <a href="https://github.com/sandialabs/cross-sim">
                  Sandia CrossSim — iterative parasitic modeling reference
                </a>
              </p>
              <p>
                Numerical error is not dataset classification accuracy. Energy,
                latency, GPU comparisons and measured-device benchmark overlays
                are not qualified by this workflow.
              </p>
            </div>
          )}
        </div>
      </div>
      <details open>
        <summary>Explicit crossbar files</summary>
        <label>
          Import crossbar JSON
          <input
            type="file"
            accept=".json,application/json"
            disabled={busy}
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (!file) return;
              setPending(true);
              void file
                .text()
                .then(async (document) => {
                  if (await act("import", { document })) {
                    setDrafts({});
                    setFiles("Imported JSON — unverified input");
                  }
                })
                .catch((e) => report((e as Error).message))
                .finally(() => setPending(false));
              e.target.value = "";
            }}
          />
        </label>
        {__FECIM_PROFILE__ === "local" && (
          <fieldset disabled={busy}>
            <legend>Open existing file project</legend>
            <label htmlFor="crossbar-root">Granted root for crossbar</label>
            <select
              id="crossbar-root"
              value={root}
              onChange={(e) => setRoot(e.target.value)}
            >
              {transport.current?.roots?.map((r) => (
                <option key={r.id} value={r.id}>
                  {r.name}
                </option>
              ))}
            </select>
            <label>
              Project name for crossbar files
              <input
                value={project}
                onChange={(e) => setProject(e.target.value)}
              />
            </label>
            <button
              onClick={() => {
                setPending(true);
                void transport.current
                  ?.openProject?.(root, project)
                  .then((id) => {
                    setProjectID(id);
                    setFiles("Rooted project opened");
                  })
                  .catch((e) => report((e as Error).message))
                  .finally(() => setPending(false));
              }}
            >
              Open file project
            </button>
          </fieldset>
        )}
        <label>
          Relative JSON path
          <input value={path} onChange={(e) => setPath(e.target.value)} />
        </label>
        {(["load", "save"] as const).map((action) => (
          <button
            key={action}
            disabled={busy}
            onClick={() => {
              if (__FECIM_PROFILE__ !== "local") {
                setFiles(
                  "Requires fecim-lattice-tools serve --root DIRECTORY; no localhost request was made.",
                );
                return;
              }
              void act(
                action,
                { path },
                { resources: [{ id: projectID, kind: "project" }] },
              ).then((ok) => {
                if (ok) {
                  setFiles(
                    action === "save"
                      ? "Saved crossbar JSON inside rooted project"
                      : "Loaded crossbar JSON — unverified input",
                  );
                  if (action === "load") setDrafts({});
                }
              });
            }}
          >
            {action === "load" ? "Load" : "Save"} crossbar JSON
          </button>
        ))}
        <p data-testid="crossbar-files">
          {files || "No automatic file writes. Hosted data remains in memory."}
        </p>
      </details>
    </section>
  );
}

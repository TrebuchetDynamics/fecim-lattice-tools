import { useEffect, useRef, useState } from "react";
import { Value } from "../../research/Shared";
import { claimScience, releaseScience } from "../../worker/ownership";
import type { Schema } from "../../api/types";
import {
  connect,
  type Transport,
  type Update,
  type Command,
} from "./transport";
import "./circuits.css";

function Table({ table }: { table: Schema<"TablePage"> }) {
  return (
    <div
      className="table-scroll"
      role="region"
      aria-label={table.caption}
      tabIndex={0}
    >
      <table id={"circuits-" + table.id}>
        <caption>{table.caption} · Simulation only</caption>
        <thead>
          <tr>
            {table.columns?.map((c, i) => (
              <th key={i} scope="col">
                {c}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {table.rows?.map((row, i) => (
            <tr key={i}>
              {row?.map((c, j) =>
                j === 0 ? (
                  <th key={j} scope="row">
                    <Value value={c.presentation} />
                  </th>
                ) : (
                  <td key={j}>
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
function Plot({ plot }: { plot: Schema<"Plot"> }) {
  const points = (plot.points || []).filter(
    (p): p is NonNullable<typeof p> => p !== null,
  );
  const min = Math.min(0, ...points.map((p) => p[1])),
    max = Math.max(0, ...points.map((p) => p[1]));
  return (
    <figure>
      <figcaption>
        {plot.title} — supplementary plot; horizontal axis {plot.x_unit},
        vertical axis {plot.y_unit}. All values and selection are in the tables.
      </figcaption>
      <svg viewBox="0 0 600 100" role="img" aria-label={plot.title}>
        {points.map((p, i) => (
          <circle
            key={i}
            cx={20 + (i * 560) / Math.max(1, points.length - 1)}
            cy={90 - (80 * (p[1] - min)) / (max - min || 1)}
            r={3}
          />
        ))}
      </svg>
    </figure>
  );
}
const configIDs = [
  "rows",
  "cols",
  "material",
  "node",
  "architecture",
  "coupling",
  "dac_bits",
  "adc_bits",
  "gain",
  "supply",
  "sense_min",
  "sense_max",
  "max_iterations",
  "max_pulses",
  "seed",
  "sense_preset",
  "input",
  "voltages",
  "active_rows",
  "read_mask",
  "write_mask",
];
export default function Screen() {
  const [data, setData] = useState<Update>(),
    [drafts, setDrafts] = useState<Record<string, string>>({}),
    [loading, setLoading] = useState(true),
    [pending, setPending] = useState(false),
    [unknown, setUnknown] = useState(false),
    [error, setError] = useState("");
  const [row, setRow] = useState<string>(),
    [col, setCol] = useState<string>(),
    [target, setTarget] = useState<string>(),
    [announcement, setAnnouncement] = useState("");
  const [filesOpen, setFilesOpen] = useState(false),
    [files, setFiles] = useState(""),
    [root, setRoot] = useState(""),
    [project, setProject] = useState(""),
    [projectID, setProjectID] = useState(""),
    [path, setPath] = useState("");
  const [playing, setPlaying] = useState(false);
  const transport = useRef<Transport | undefined>(undefined),
    alive = useRef(true),
    owner = useRef({}),
    generation = useRef(0),
    tail = useRef<Promise<unknown>>(Promise.resolve()),
    poll = useRef<ReturnType<typeof setTimeout> | undefined>(undefined),
    alert = useRef<HTMLDivElement>(null),
    lastAnnounce = useRef(0);
  const active = !!data && ["running", "cancelling"].includes(data.job.state),
    busy = loading || pending || active || unknown || !data;
  const value = (id: string) =>
    drafts[id] ?? data?.snapshot.fields?.find((f) => f.id === id)?.input ?? "";
  const choices: Record<string, string[]> = JSON.parse(
    data?.snapshot.fields?.find((f) => f.id === "choices")?.input || "{}",
  );
  const publish = (next: Update) => {
    if (!alive.current) return;
    setData(next);
    setUnknown(false);
    const running = ["running", "cancelling"].includes(next.job.state);
    if (!running) releaseScience(owner.current);
    if (!running || Date.now() - lastAnnounce.current >= 5000) {
      setAnnouncement(next.job.message || next.snapshot.status);
      lastAnnounce.current = Date.now();
    }
  };
  const enqueue = <T,>(fn: () => Promise<T>) => {
    const next = tail.current.then(fn);
    tail.current = next.catch(() => {});
    return next;
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
          setError(
            "Execution outcome unknown; retained results remain readable. " +
              (e as Error).message,
          );
        }
      });
    }, 75);
  };
  const act = async (
    action: string,
    input: Record<string, string> = {},
    extra: Partial<Command> = {},
  ) => {
    setPending(true);
    setError("");
    setPlaying(false);
    try {
      await enqueue(async () => {
        if (!transport.current || !data)
          throw new Error("Circuits transport is not ready");
        const command: Command = {
          action,
          input,
          revision: data.snapshot.revision,
          ...extra,
        };
        if (["write", "read", "compute"].includes(action)) {
          claimScience(owner.current);
          try {
            const next = await transport.current.start(command);
            publish(next);
            if (["running", "cancelling"].includes(next.job.state)) watch();
          } catch (e) {
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
      if (alive.current) setError((e as Error).message);
      return false;
    } finally {
      if (alive.current) setPending(false);
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
      .catch((e) => {
        if (alive.current) setError((e as Error).message);
      })
      .finally(() => {
        if (alive.current) setLoading(false);
      });
    return () => {
      alive.current = false;
      generation.current++;
      clearTimeout(poll.current);
      const t = transport.current;
      transport.current = undefined;
      void tail.current
        .then(() => t?.close())
        .then(() => releaseScience(owner.current))
        .catch(() => {});
    };
  }, []);
  useEffect(() => {
    if (error && location.pathname === "/modules/circuits")
      alert.current?.focus();
  }, [error]);
  useEffect(() => {
    if (!playing || busy) return;
    const timer = setTimeout(() => {
      void act("timing", { intent: "step" }).then((ok) => {
        if (ok && alive.current) setPlaying(true);
      });
    }, 750);
    return () => clearTimeout(timer);
  }, [playing, busy, data]);
  const apply = async () => {
    const submitted = { ...drafts };
    if (await act("configure", submitted))
      setDrafts((old) =>
        Object.fromEntries(
          Object.entries(old).filter(([k, v]) => submitted[k] !== v),
        ),
      );
  };
  return (
    <section className="section circuits-lab">
      <header className="circuits-header">
        <p className="eyebrow">M4 / DEVICE + PERIPHERAL LAB</p>
        <h1>Circuits</h1>
        <p>
          Program. Sense. Multiply. A persistent material-aware array, with the
          native Go device path.
        </p>
      </header>
      <p className="boundary">
        Simulation only.{" "}
        {__FECIM_PROFILE__ === "hosted"
          ? "Dedicated CPU Worker; ephemeral science, no upload, persistence or file writes. Ordinary same-origin static asset requests still occur."
          : "Authenticated same-origin Go computation; explicit exports require an opened, rooted project."}{" "}
        No GPU/WebGPU execution. Real Orca qualification is NOT RUN.
      </p>
      <p className="circuits-status" data-testid="circuits-status">
        {loading
          ? "Loading Go circuits…"
          : data?.snapshot.status || "Circuits unavailable"}
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
        <div role="alert" tabIndex={-1} ref={alert}>
          <p>{error}</p>
          <button onClick={() => setError("")}>Dismiss error</button>
        </div>
      )}
      {unknown && (
        <p>
          Execution state is unknown. Do not start competing science until
          authoritative recovery or explicit discard.
        </p>
      )}
      <div className="actions">
        <button
          disabled={loading || pending}
          onClick={() => {
            setPending(true);
            void enqueue(async () => {
              if (!transport.current) transport.current = await connect();
              const next = await transport.current.read();
              publish(next);
              if (["running", "cancelling"].includes(next.job.state)) watch();
            })
              .catch((e) => setError((e as Error).message))
              .finally(() => setPending(false));
          }}
        >
          Reconnect circuit state
        </button>
        <button
          disabled={!active || pending || data?.job.state === "cancelling"}
          onClick={() => {
            setPending(true);
            void enqueue(async () => {
              if (!transport.current) return;
              publish(await transport.current.cancel());
              watch();
            })
              .catch((e) => {
                setUnknown(true);
                setError((e as Error).message);
              })
              .finally(() => setPending(false));
          }}
        >
          Cancel circuit computation
        </button>
        {unknown && (
          <button
            onClick={() => {
              setPending(true);
              void enqueue(async () => {
                await transport.current?.close();
                releaseScience(owner.current);
                transport.current = await connect();
                publish(await transport.current.read());
                setError("");
              })
                .catch((e) => setError((e as Error).message))
                .finally(() => setPending(false));
            }}
          >
            Discard session and retry
          </button>
        )}
      </div>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          void apply();
        }}
      >
        <fieldset disabled={busy}>
          <legend>01 / Array and sense configuration</legend>
          <p id="circuits-config-help">
            Apply changes explicitly. Go validates the complete transaction.
            Resizing or changing material resets device levels. Signed inputs
            are digital codes; voltages and units are supplied by Go.
          </p>
          <div className="circuits-fields">
            {data?.snapshot.fields
              ?.filter((f) => configIDs.includes(f.id))
              .map((f) => (
                <div key={f.id}>
                  <label htmlFor={"circuit-" + f.id}>{f.label}</label>
                  {choices[f.id] ? (
                    <select
                      id={"circuit-" + f.id}
                      name={f.id}
                      value={value(f.id)}
                      aria-describedby="circuits-config-help"
                      onChange={(e) =>
                        setDrafts((old) => ({ ...old, [f.id]: e.target.value }))
                      }
                    >
                      {choices[f.id].map((v) => (
                        <option key={v} value={v}>
                          {v}
                        </option>
                      ))}
                    </select>
                  ) : (
                    <input
                      id={"circuit-" + f.id}
                      name={f.id}
                      value={value(f.id)}
                      autoComplete="off"
                      aria-describedby="circuits-config-help"
                      onChange={(e) =>
                        setDrafts((old) => ({ ...old, [f.id]: e.target.value }))
                      }
                    />
                  )}
                </div>
              ))}
          </div>
          {Object.keys(drafts).length > 0 && (
            <p>
              Unapplied settings — retained results have not been recomputed.
            </p>
          )}
          <button type="submit">Apply configuration</button>
        </fieldset>
      </form>
      <fieldset disabled={busy}>
        <legend>02 / Select and operate</legend>
        <div className="circuits-fields">
          <label>
            Cell row
            <input
              value={row ?? value("selected_row")}
              onChange={(e) => setRow(e.target.value)}
            />
          </label>
          <label>
            Cell column
            <input
              value={col ?? value("selected_col")}
              onChange={(e) => setCol(e.target.value)}
            />
          </label>
          <label>
            Write target level
            <input
              value={target ?? value("write_target")}
              onChange={(e) => setTarget(e.target.value)}
            />
          </label>
        </div>
        <div className="actions">
          <button
            onClick={() =>
              void act("select", {
                row: row ?? value("selected_row"),
                col: col ?? value("selected_col"),
              })
            }
          >
            Select cell
          </button>
          <button
            className="circuit-write"
            onClick={() =>
              void act("write", { target: target ?? value("write_target") })
            }
          >
            WRITE selected cell
          </button>
          <button onClick={() => void act("read")}>READ selected cell</button>
          <button onClick={() => void act("compute")}>COMPUTE vector</button>
        </div>
        <p>
          0T1R programming affects a whole column. 1T1R selects one row; 2T1R
          additionally supports read and write masks. A pulse limit is not
          verified convergence.
        </p>
        <div className="actions">
          <button
            disabled={
              !data?.snapshot.actions?.find((a) => a.id === "undo")?.available
            }
            onClick={() => void act("undo")}
          >
            Undo device change
          </button>
          <button onClick={() => void act("randomize")}>Randomize array</button>
          <button onClick={() => void act("reset")}>Reset array</button>
        </div>
      </fieldset>
      <section aria-labelledby="circuit-results">
        <h2 id="circuit-results">03 / State and evidence</h2>
        {data?.snapshot.tables?.map((t) => (
          <div key={t.id}>
            <Table table={t} />
            {["cells", "cell_results"].includes(t.id) && (
              <div className="actions">
                <button
                  disabled={busy || t.page <= 1}
                  onClick={() =>
                    void act("view", {
                      scope: t.id === "cells" ? "cells" : "result",
                      page: String(t.page - 1),
                    })
                  }
                >
                  Previous{" "}
                  {t.id === "cells"
                    ? "persistent cells"
                    : "retained cell results"}
                </button>
                <button
                  disabled={busy || t.page >= t.total}
                  onClick={() =>
                    void act("view", {
                      scope: t.id === "cells" ? "cells" : "result",
                      page: String(t.page + 1),
                    })
                  }
                >
                  Next{" "}
                  {t.id === "cells"
                    ? "persistent cells"
                    : "retained cell results"}
                </button>
              </div>
            )}
          </div>
        ))}
        {data?.snapshot.plots?.map((plot) => (
          <Plot key={plot.id} plot={plot} />
        ))}
        <div className="actions">
          <button
            disabled={busy}
            onClick={() => void act("timing", { intent: "step" })}
          >
            Step timing
          </button>
          <button
            disabled={busy}
            aria-pressed={playing}
            onClick={() => setPlaying((v) => !v)}
          >
            {playing ? "Pause timing" : "Play timing"}
          </button>
          <button
            disabled={busy}
            onClick={() => void act("timing", { intent: "reset" })}
          >
            Reset timing
          </button>
        </div>
        <h3>Model boundaries and operation log</h3>
        <ul>
          {data?.snapshot.evidence?.map((entry, i) => (
            <li key={i}>{entry}</li>
          ))}
        </ul>
      </section>
      <button
        aria-expanded={filesOpen}
        aria-controls="circuit-files"
        onClick={() => setFilesOpen((v) => !v)}
      >
        Files and capabilities
      </button>
      {filesOpen && (
        <section id="circuit-files">
          <h2>Explicit local exports</h2>
          <p>
            No automatic downloads or hosted scientific uploads. GPU execution
            and external instrument access are unavailable.
          </p>
          {__FECIM_PROFILE__ === "local" && (
            <fieldset disabled={busy}>
              <legend>Rooted export project</legend>
              <label>
                Granted root for circuits
                <select value={root} onChange={(e) => setRoot(e.target.value)}>
                  {transport.current?.roots?.map((r) => (
                    <option key={r.id} value={r.id}>
                      {r.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Project name for circuit exports
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
                    .catch((e) => setError((e as Error).message))
                    .finally(() => setPending(false));
                }}
              >
                Open export project
              </button>
            </fieldset>
          )}
          <label>
            Relative JSON export path
            <input value={path} onChange={(e) => setPath(e.target.value)} />
          </label>
          <button
            disabled={busy}
            onClick={() => {
              if (__FECIM_PROFILE__ !== "local") {
                setFiles(
                  "Requires fecim-lattice-tools serve --root DIRECTORY; no localhost request was made",
                );
                return;
              }
              void act(
                "save",
                { path },
                { resources: [{ kind: "project", id: projectID }] },
              ).then((ok) => {
                if (ok) setFiles("Saved circuit JSON inside rooted project");
              });
            }}
          >
            Export circuit JSON
          </button>
          <p data-testid="circuits-files">
            {files || "No file writes requested"}
          </p>
        </section>
      )}
    </section>
  );
}

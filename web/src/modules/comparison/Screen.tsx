import { useEffect, useRef, useSyncExternalStore } from "react";
import type { Schema } from "../../api/types";
import { Value, ErrorSummary } from "../../research/Shared";
import { comparisonOwner as owner } from "./owner";
import "./comparison.css";
const optionTables: Record<string, string> = {
  workload: "workloads",
  profile: "profiles",
  mode: "modes",
  output: "outputs",
};
function ResultsTable({ table }: { table: Schema<"TablePage"> }) {
  return (
    <div className="table-scroll">
      <table id={"comparison-" + table.id}>
        <caption tabIndex={-1}>{table.caption}</caption>
        <thead>
          <tr>
            {table.columns?.map((c) => (
              <th scope="col" key={c}>
                {c}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {table.rows?.map((row, i) => (
            <tr key={i}>
              {row?.map((cell, j) =>
                j === 0 ? (
                  <th scope="row" key={j}>
                    <Value value={cell.presentation} />
                  </th>
                ) : (
                  <td key={j}>
                    <Value value={cell.presentation} />
                    <small>Evidence: {cell.evidence}</small>
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
  useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const ref = useRef<HTMLElement>(null),
    focusToken = useRef(0);
  useEffect(() => {
    void owner.connect();
  }, []);
  const act = async (action: string, format?: string) => {
    const token = ++focusToken.current;
    try {
      await owner.action(action, format);
      if (action === "diff" || action === "sensitivity")
        requestAnimationFrame(() => {
          if (token === focusToken.current)
            ref.current
              ?.querySelector<HTMLElement>("#comparison-" + action + " caption")
              ?.focus();
        });
    } catch (error) {
      owner.error = error as Error;
      owner.changed();
    }
  };
  const s = owner.snapshot;
  const exportUnavailable =
    __FECIM_PROFILE__ === "local" &&
    !s?.actions?.some((a) => a.id === "export" && a.available);
  return (
    <section className="section comparison-lab" ref={ref}>
      <header className="lab-header">
        <p className="eyebrow">Module 05 / Assumption workbench</p>
        <h1>Comparison</h1>
        <p>
          Inspect the assumptions before the advantage. Scenario estimates and
          architecture demonstrations are separate Go protocols.
        </p>
      </header>
      {owner.error && (
        <ErrorSummary
          error={owner.error}
          clear={() => {
            owner.error = undefined;
            owner.changed();
          }}
        />
      )}
      <p role="status" data-testid="comparison-status">
        {owner.busy ? "Calculating bounded Go request…" : owner.status}
      </p>
      {!s ? (
        <button onClick={() => void owner.retry()}>
          Retry comparison connection
        </button>
      ) : (
        <>
          <form
            className="comparison-controls"
            onSubmit={(e) => {
              e.preventDefault();
              void act("calculate");
            }}
          >
            <fieldset>
              <legend>Workload and scenario</legend>
              {s.fields?.map((f) => {
                const id = "comparison-field-" + f.id,
                  options = s.tables?.find((t) => t.id === optionTables[f.id]);
                return (
                  <div className="field" key={f.id}>
                    <label htmlFor={id}>{f.label}</label>
                    {options ? (
                      <select
                        id={id}
                        value={owner.drafts[f.id] ?? f.input}
                        onChange={(e) => owner.draft(f.id, e.target.value)}
                      >
                        {options.rows?.map(
                          (r, i) =>
                            r?.[0] && (
                              <option key={i} value={r[0].presentation.Display}>
                                {r[0].presentation.Display}
                              </option>
                            ),
                        )}
                      </select>
                    ) : (
                      <input
                        id={id}
                        inputMode="decimal"
                        value={owner.drafts[f.id] ?? f.input}
                        onChange={(e) => owner.draft(f.id, e.target.value)}
                        aria-describedby={id + "-value"}
                      />
                    )}
                    <small id={id + "-value"}>
                      Go value: <Value value={f.presentation} />
                    </small>
                  </div>
                );
              })}
            </fieldset>
            <p className="honesty">
              Workload, rate, scenario and review-mode edits apply on Calculate.
              Sensitivity output applies on Rank sensitivity. Until then, tables
              and exports retain the last successful Go result.
            </p>
            <div className="actions">
              {s.actions
                ?.filter((a) => a.id !== "export")
                .map((a) => (
                  <button
                    key={a.id}
                    type={a.id === "calculate" ? "submit" : "button"}
                    disabled={owner.busy || !a.available}
                    onClick={
                      a.id === "calculate" ? undefined : () => void act(a.id)
                    }
                  >
                    {a.label}
                  </button>
                ))}
            </div>
          </form>
          <div className="comparison-results">
            {s.tables
              ?.filter((t) => !Object.values(optionTables).includes(t.id))
              .map((t) => (
                <ResultsTable key={t.id} table={t} />
              ))}
          </div>
          <details id="comparison-evidence">
            <summary>Source evidence and limitations</summary>
            {s.evidence?.map((e, i) => (
              <p className="comparison-evidence" key={i}>
                {e}
              </p>
            ))}
          </details>
          <section aria-labelledby="comparison-export-title">
            <h2 id="comparison-export-title">Local versioned exports</h2>
            <p>
              JSON/CSV preserve canonical Go values, units, protocol versions
              and simulation evidence. No scientific download or upload in
              hosted mode.
            </p>
            {__FECIM_PROFILE__ === "local" && (
              <>
                <label htmlFor="comparison-export-project">
                  Opened export project
                </label>
                <select
                  id="comparison-export-project"
                  value={owner.project}
                  onChange={(e) => {
                    owner.project = e.target.value;
                    owner.changed();
                  }}
                >
                  <option value="">Open a granted project first</option>
                  {owner.projects.map((p) => (
                    <option value={p.id} key={p.id}>
                      {p.name}
                    </option>
                  ))}
                </select>
                <button
                  disabled={owner.busy}
                  onClick={() => {
                    void owner.refreshProjects().catch((error) => {
                      owner.error = error as Error;
                      owner.changed();
                    });
                  }}
                >
                  Refresh opened projects
                </button>
              </>
            )}
            <div className="actions">
              <button
                disabled={owner.busy || exportUnavailable}
                onClick={() => void act("export", "json")}
              >
                Export JSON
              </button>
              <button
                disabled={owner.busy || exportUnavailable}
                onClick={() => void act("export", "csv")}
              >
                Export CSV
              </button>
            </div>
          </section>
        </>
      )}
    </section>
  );
}

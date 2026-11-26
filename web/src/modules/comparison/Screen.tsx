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
// The Go summary labels and sensitivity option IDs describe the same outputs.
const metricOutputs: Record<string, string> = {
  "Energy reduction": "energy_reduction_pct",
  "Latency reduction": "latency_reduction_pct",
  "TCO reduction": "tco_reduction_pct",
  "CO2 reduction": "co2_reduction_pct",
};
function ResultsTable({
  table,
  inspect,
}: {
  table: Schema<"TablePage">;
  inspect?: (output: string) => void;
}) {
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
                    {inspect && metricOutputs[cell.presentation.Display] && (
                      <button
                        type="button"
                        aria-controls="comparison-metric-detail"
                        aria-label={"Inspect " + cell.presentation.Speakable}
                        onClick={() =>
                          inspect(metricOutputs[cell.presentation.Display])
                        }
                      >
                        Inspect
                      </button>
                    )}
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
    detailHeading = useRef<HTMLHeadingElement>(null),
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
  const inspect = (output: string) => {
    const token = ++focusToken.current;
    owner.draft("output", output);
    requestAnimationFrame(() => {
      if (token === focusToken.current) detailHeading.current?.focus();
    });
  };
  const s = owner.snapshot;
  const summary = s?.tables?.find((t) => t.id === "summary");
  const sources = s?.tables?.find((t) => t.id === "sources");
  const sensitivity = s?.tables?.find((t) => t.id === "sensitivity");
  const scenario = s?.fields?.find((f) => f.id === "profile");
  const rankedOutput = s?.fields?.find((f) => f.id === "output")?.input;
  const selectedOutput = owner.drafts.output ?? rankedOutput;
  const selectedMetric = summary?.rows?.find(
    (row) =>
      row?.[0] && metricOutputs[row[0].presentation.Display] === selectedOutput,
  );
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
        <p>
          {__FECIM_PROFILE__ === "local"
            ? "Authenticated local app · Export to an opened project."
            : "Hosted browser · Calculate in memory; exports require the local app."}
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
              Calculate applies workload, rate, scenario and review-mode edits.
              Until then, results and exports use the last applied scenario.
              Rank sensitivity runs separately for the selected output.
            </p>
            <div className="actions">
              {s.actions
                ?.filter((a) => a.id !== "export")
                .map((a) => (
                  <button
                    key={a.id}
                    type={a.id === "calculate" ? "submit" : "button"}
                    className={a.id === "calculate" ? "primary" : undefined}
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
          <section
            className="comparison-results"
            aria-labelledby="comparison-results-title"
          >
            <h2 id="comparison-results-title">Scenario results</h2>
            <div className="comparison-matrix">
              {summary && <ResultsTable table={summary} inspect={inspect} />}
              <section
                className="comparison-metric-detail"
                id="comparison-metric-detail"
                aria-labelledby="comparison-metric-title"
              >
                <h3 id="comparison-metric-title" ref={detailHeading} tabIndex={-1}>
                  Metric detail
                </h3>
                {scenario && (
                  <p>Applied scenario: <Value value={scenario.presentation} /></p>
                )}
                {selectedMetric?.[0] && selectedMetric[1] && (
                  <>
                    <h4><Value value={selectedMetric[0].presentation} /></h4>
                    <p><Value value={selectedMetric[1].presentation} /></p>
                    <p>Unit: {selectedMetric[1].unit}</p>
                    <p>Evidence: {selectedMetric[1].evidence}</p>
                  </>
                )}
                {sources && (
                  <details>
                    <summary>Scenario assumptions</summary>
                    <ResultsTable table={sources} />
                  </details>
                )}
                <details id="comparison-evidence">
                  <summary>Source evidence and limitations</summary>
                  {s.evidence?.map((e, i) => (
                    <p className="comparison-evidence" key={i}>
                      {e}
                    </p>
                  ))}
                </details>
                <h4>Sensitivity</h4>
                {!sensitivity ? (
                  <p>Use Rank sensitivity to see the selected output for this scenario.</p>
                ) : (
                  <>
                    {selectedOutput !== rankedOutput && (
                      <p>
                        Showing the previous output. Use Rank sensitivity to update it.
                      </p>
                    )}
                    <ResultsTable table={sensitivity} />
                  </>
                )}
              </section>
            </div>
            {s.tables
              ?.filter((t) => ["energy", "architectures", "diff"].includes(t.id))
              .map((t) => (
                <ResultsTable key={t.id} table={t} />
              ))}
          </section>
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

import { useEffect, useState } from "react";
import type { Results } from "./types";
import type { Engine } from "./engine";
import type { LocalAdapter } from "../local/adapter";
import { GoData } from "./GoData";
export function LocalActions({
  data,
  engine,
}: {
  data: Results;
  engine: Engine | null;
}) {
  const [message, setMessage] = useState("");
  const [adapter, setAdapter] = useState<LocalAdapter>();
  const [root, setRoot] = useState("");
  const [path, setPath] = useState("preset.json");
  const [loopKind, setLoopKind] = useState("pe-csv");
  const [analysisKind, setAnalysisKind] = useState("analysis-json");
  const [loaded, setLoaded] = useState<unknown>();
  const [busy, setBusy] = useState(false);
  useEffect(() => {
    let alive = true;
    // Vite removes this dynamic import completely from the hosted profile.
    if (__FECIM_PROFILE__ === "local" && data.catalogue) {
      void import("../local/adapter")
        .then((m) => m.connectLocal(data.catalogue!.Identity))
        .then((value) => {
          if (alive) {
            setAdapter(value);
            setRoot(value.roots[0]?.id || "");
          }
        })
        .catch((error) => {
          if (alive) setMessage(String(error));
        });
    }
    return () => {
      alive = false;
    };
  }, [data.catalogue]);
  async function perform(action: string) {
    if (__FECIM_PROFILE__ !== "local") {
      setMessage(
        "Requires the local application: run fecim-lattice-tools serve --root DIRECTORY (design shorthand: fecim serve). No localhost request was made; nothing was saved or downloaded.",
      );
      return;
    }
    if (action === "Verify stored run" || action === "EDA tools") {
      setMessage(
        "Not available in this Hysteresis slice. Browser snapshots are not verified persisted studies; use the existing native CLI for stored-run verification or EDA.",
      );
      return;
    }
    if (!adapter || !engine) {
      setMessage(
        "Authenticated local session unavailable. Renew the launch link in the terminal and open it.",
      );
      return;
    }
    setBusy(true);
    try {
      if (action === "Log out local session") {
        await adapter.logout();
        setMessage("Logged out. Renew the launch link in the terminal.");
        return;
      }
      const operation = action.startsWith("Load") ? "load" : "save";
      const kind = action.includes("preset")
        ? "preset"
        : action.includes("calibration")
          ? "calibration"
          : action === "Export loop"
            ? loopKind
            : analysisKind;
      const snapshot =
        operation === "save" ? await engine.fileSnapshot() : undefined;
      const response = await adapter.file(
        root,
        path,
        kind,
        operation,
        snapshot,
      );
      setLoaded(response.value);
      if (kind === "preset" && operation === "load" && response.value?.Config)
        await engine.apply({
          Kind: "configure",
          Config: response.value.Config,
        });
      setMessage(
        response.status +
          (kind === "calibration" && operation === "load"
            ? ". Imported calibration is displayed read-only, not applied or verified. Recalibrate in memory for programming."
            : ""),
      );
    } catch (error) {
      setMessage("Local file operation failed: " + String(error));
    } finally {
      setBusy(false);
    }
  }
  return (
    <details>
      <summary>Local files and exports</summary>
      <p>
        {__FECIM_PROFILE__ === "local"
          ? "Explicit actions below use only launch-granted roots. Existing destinations are never overwritten. Exports are browser simulation data, not server-verified studies."
          : "Hosted studies and references stay in memory. File writes, projects, stored-run verification and exports require the local application."}
      </p>
      {__FECIM_PROFILE__ === "local" && (
        <div className="lab-grid">
          <label>
            Granted root
            <select
              value={root}
              onChange={(e) => setRoot(e.target.value)}
              disabled={!adapter}
            >
              {adapter?.roots.map((r) => (
                <option key={r.id} value={r.id}>
                  {r.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Relative file path
            <input
              value={path}
              onChange={(e) => setPath(e.target.value)}
              aria-describedby="file-path-help"
            />
          </label>
          <p id="file-path-help">
            Enter a relative file path beneath this grant. Existing parent
            directory required. Use a new path on conflict; no automatic saves.
          </p>
          <label>
            Loop export format
            <select
              value={loopKind}
              onChange={(e) => setLoopKind(e.target.value)}
            >
              {["pe-csv", "pe-json", "pe-svg", "log-csv"].map((k) => (
                <option key={k}>{k}</option>
              ))}
            </select>
          </label>
          <label>
            Diagnostic export format
            <select
              value={analysisKind}
              onChange={(e) => setAnalysisKind(e.target.value)}
            >
              {[
                "analysis-json",
                "pund-csv",
                "forc-sweep-csv",
                "forc-density-csv",
              ].map((k) => (
                <option key={k}>{k}</option>
              ))}
            </select>
          </label>
        </div>
      )}
      <div className="actions">
        {[
          "Save preset",
          "Load preset",
          "Save calibration",
          "Load calibration",
          "Export loop",
          "Export diagnostics",
          "Verify stored run",
          "EDA tools",
          ...(__FECIM_PROFILE__ === "local" ? ["Log out local session"] : []),
        ].map((action) => (
          <button
            key={action}
            disabled={busy}
            onClick={() => void perform(action)}
          >
            {action}
          </button>
        ))}
      </div>
      {message && (
        <p role="status" data-testid="file-status">
          {message}
        </p>
      )}
      {loaded !== undefined && (
        <details>
          <summary>Imported file — unverified</summary>
          <GoData value={loaded} />
        </details>
      )}
    </details>
  );
}

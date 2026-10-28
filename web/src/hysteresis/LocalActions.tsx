import { useState } from "react";
export function LocalActions() {
  const [message, setMessage] = useState("");
  async function perform(action: string) {
    if (__FECIM_PROFILE__ === "local" && window.fecimLocalAdapter) {
      try {
        await window.fecimLocalAdapter.perform(action);
        setMessage("Local adapter completed " + action);
      } catch (error) {
        setMessage("Local adapter failed: " + String(error));
      }
    } else
      setMessage(
        "Requires the local application: run fecim serve. The authenticated local adapter is not implemented in this build. No localhost request was made; nothing was saved or downloaded.",
      );
  }
  return (
    <details>
      <summary>Local files and exports</summary>
      <p>
        Hosted studies and references stay in memory. File writes, projects,
        stored-run verification and exports require a local capability adapter.
        A build profile is not a security grant.
      </p>
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
        ].map((action) => (
          <button key={action} onClick={() => void perform(action)}>
            {action}
          </button>
        ))}
      </div>
      {message && <p role="status">{message}</p>}
    </details>
  );
}

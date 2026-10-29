import "../../worker/hysteresis-worker.mjs";
import { manifest } from "../../api/hosted";
import type { Schema } from "../../api/types";
export async function hostedComparison() {
  const m = await manifest(),
    entry = m.modules.comparison;
  if (!entry) throw new Error("Comparison Worker unavailable in this build");
  const client = new HysteresisWorkerClient({
    identity: entry.identity,
    prefix: entry.prefix,
    session: crypto.randomUUID(),
    workerURL: "/engine/hysteresis-worker.mjs",
    workerSHA256: m.assets["hysteresis-worker.mjs"],
    wasmURL: "/engine/" + entry.wasm,
    wasmSHA256: m.assets[entry.wasm],
    runtimeURL: "/engine/wasm_exec.js",
    runtimeSHA256: m.assets["wasm_exec.js"],
  });
  try {
    await client.ready;
  } catch (error) {
    client.terminate();
    throw error;
  }
  const call = async (op: string, input: Schema<"Command">) => {
    const reply = JSON.parse(await client.request(op, input)) as {
      result: Schema<"Snapshot">;
      error?: string;
    };
    if (reply.error) throw new Error(reply.error);
    return reply.result;
  };
  return {
    snapshot: () => call("snapshot", { action: "", revision: 0 }),
    apply: (command: Schema<"Command">) => call("action", command),
    close: () => client.terminate(),
    projects: async () => [] as { id: string; name: string }[],
  };
}

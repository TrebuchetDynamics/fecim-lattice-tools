// This same-origin capability adapter is compiled out of the hosted profile.
import createClient from "openapi-fetch";
import type { paths } from "../../api/generated";
import type { Schema } from "../../api/types";
import { localResearch } from "../../local/research";
import type { Transport, Job } from "./transport";
const client = createClient<paths>({
  credentials: "same-origin",
  cache: "no-store",
  redirect: "error",
});
client.use({
  onRequest({ request }) {
    if (request.method !== "GET") {
      request.headers.set("Content-Type", "application/json");
      request.headers.set("X-FeCIM-Request", "1");
    }
    return request;
  },
});
function unwrap<T>(r: { data?: T; error?: unknown; response: Response }): T {
  if (!r.response.ok || r.error)
    throw new Error(
      (r.error as { detail?: string })?.detail ||
        "Local circuits request rejected; check inputs, revision and rooted project",
    );
  return r.data!;
}
export async function connect(): Promise<Transport> {
  const research = await localResearch(),
    roots = await research.roots();
  let snapshot = unwrap(
    await client.POST("/api/v1/modules/{module}/sessions", {
      params: { path: { module: "circuits" } },
      body: {},
    }),
  );
  const session = snapshot.session!,
    params = { path: { session } };
  let job: Job = {
    state: "idle",
    results_state: "unavailable",
    progress: { total: 0, completed: 0, reused: 0, failed: 0, cancelled: 0 },
    id: "",
    cancellation_requested: false,
    input_revision: "",
    analysis_available: false,
    counts: {},
    results_revision: "",
  };
  const read = async () => {
    snapshot = unwrap(
      await client.GET("/api/v1/module-sessions/{session}", { params }),
    );
    // Discover an admitted job even if its start response was lost.
    const admitted = snapshot.fields?.find((f) => f.id === "job_id")?.input;
    if (admitted) job.id = admitted;
    if (job.id)
      job = unwrap(
        await client.GET("/api/v1/module-sessions/{session}/jobs/{job}", {
          params: { path: { session, job: job.id } },
        }),
      );
    snapshot = unwrap(
      await client.GET("/api/v1/module-sessions/{session}", { params }),
    );
    return { snapshot, job };
  };
  return {
    roots,
    openProject: async (root, path) => (await research.open(root, path)).id,
    read,
    action: async (command) => {
      snapshot = unwrap(
        await client.POST("/api/v1/module-sessions/{session}/actions", {
          params,
          body: command as Schema<"Command">,
        }),
      );
      return { snapshot, job };
    },
    start: async (command) => {
      job = unwrap(
        await client.POST("/api/v1/module-sessions/{session}/jobs", {
          params,
          body: command,
        }),
      );
      return read();
    },
    cancel: async () => {
      job = unwrap(
        await client.POST(
          "/api/v1/module-sessions/{session}/jobs/{job}/cancel",
          { params: { path: { session, job: job.id } }, body: {} },
        ),
      );
      return { snapshot, job };
    },
    close: async () => {
      unwrap(
        await client.POST("/api/v1/module-sessions/{session}/close", {
          params,
          body: {},
        }),
      );
    },
  };
}

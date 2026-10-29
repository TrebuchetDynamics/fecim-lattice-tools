// Local-only chunk. Hosted Learn must never construct this authenticated transport.
import createClient from "openapi-fetch";
import type { paths } from "../../api/generated";
import type { Schema } from "../../api/types";
const api = createClient<paths>({
  credentials: "same-origin",
  cache: "no-store",
  redirect: "error",
});
api.use({
  onRequest({ request }) {
    if (request.method !== "GET") {
      request.headers.set("X-FeCIM-Request", "1");
      request.headers.set("Content-Type", "application/json");
    }
    return request;
  },
});
function value<T>(r: { data?: T; error?: unknown; response: Response }): T {
  if (!r.response.ok || r.error)
    throw new Error(
      "EDA request rejected: check project, revision, typed input, artifact integrity and local tool availability.",
    );
  return r.data!;
}
export async function connect() {
  const roots = await fetch("/api/v1/roots", {
    credentials: "same-origin",
    cache: "no-store",
    redirect: "error",
  });
  if (!roots.ok)
    throw new Error("Open the authenticated terminal launch link first.");
  const info = (await roots.json()) as {
    UI: string;
    roots: { id: string; name: string }[];
  };
  if (info.UI !== __FECIM_UI__)
    throw new Error("Local/UI fingerprint mismatch; rebuild and reload.");
  const snapshot = value(
    await api.POST("/api/v1/modules/{module}/sessions", {
      params: { path: { module: "eda" } },
      body: {},
    }),
  );
  const session = snapshot.session;
  return {
    snapshot,
    roots: info.roots,
    refresh: async () =>
      value(
        await api.GET("/api/v1/module-sessions/{session}", {
          params: { path: { session } },
        }),
      ),
    apply: async (body: Schema<"Command">) =>
      value(
        await api.POST("/api/v1/module-sessions/{session}/actions", {
          params: { path: { session } },
          body,
        }),
      ),
    start: async (body: Schema<"Command">) =>
      value(
        await api.POST("/api/v1/module-sessions/{session}/jobs", {
          params: { path: { session } },
          body,
        }),
      ),
    job: async (job: string) =>
      value(
        await api.GET("/api/v1/module-sessions/{session}/jobs/{job}", {
          params: { path: { session, job } },
        }),
      ),
    cancel: async (job: string) =>
      value(
        await api.POST("/api/v1/module-sessions/{session}/jobs/{job}/cancel", {
          params: { path: { session, job } },
          body: {},
        }),
      ),
    weights: async (root: string, path: string) =>
      value(
        await api.POST("/api/v1/resources", {
          body: { root, path, kind: "input" },
        }),
      ),
  };
}
export type Local = Awaited<ReturnType<typeof connect>>;

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
      "MNIST request rejected: check model/dataset availability, revision, numeric input and admission limits.",
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
  // Local-only opaque identity, never model/data/result bytes or authority.
  // The authenticated server still validates the recovered session on every read.
  const key = "fecim.mnist.session." + __FECIM_UI__;
  let saved = "",
    reloadRecovery = true;
  try {
    saved = sessionStorage.getItem(key) || "";
  } catch {
    reloadRecovery = false;
  }
  let snapshot: Schema<"Snapshot"> | undefined;
  if (saved) {
    const resumed = await api.GET("/api/v1/module-sessions/{session}", {
      params: { path: { session: saved } },
    });
    if (resumed.response.status !== 404) snapshot = value(resumed);
  }
  if (!snapshot) {
    snapshot = value(
      await api.POST("/api/v1/modules/{module}/sessions", {
        params: { path: { module: "mnist" } },
        body: {},
      }),
    );
    try {
      sessionStorage.setItem(key, snapshot.session);
    } catch {
      reloadRecovery = false;
    }
  }
  const session = snapshot.session;
  return {
    snapshot,
    reloadRecovery,
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
    bind: async (root: string, path: string, kind: "model" | "dataset") =>
      value(
        await api.POST("/api/v1/resources", {
          body: { root, path, kind },
        }),
      ),
  };
}
export type Local = Awaited<ReturnType<typeof connect>>;

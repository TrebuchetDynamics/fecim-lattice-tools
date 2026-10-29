import { manifest } from "../../api/hosted";
import type { Schema } from "../../api/types";

export type Node = {
  kind: string;
  text?: string;
  url?: string;
  level?: number;
  anchor?: string;
  ordered?: boolean;
  children?: Node[];
};
export type Heading = { text: string; level: number; anchor: string };
export type Document = {
  id: string;
  path: string;
  title: string;
  category: string;
  sha256: string;
  rights: string;
  attribution: string;
  bytes: number;
  terms: string[];
  headings: Heading[];
  nodes: Node[];
};
export type Entry = Document & { content: string; content_sha256: string };
export type Manifest = {
  version: string;
  digest: string;
  document_count: number;
  decoded_bytes: number;
  page_size: number;
  documents: Entry[];
};
export type Page = {
  query: string;
  total: number;
  page: number;
  page_size: number;
  results: {
    DocPath: string;
    Title: string;
    Category: string;
    Snippet: string;
  }[];
};
export type History = { favorites: string[]; recent: string[] };
export type Output = {
  document?: Document;
  page: Page;
  document_count: number;
  decoded_bytes: number;
  corpus_sha256: string;
  history: History;
  evidence_status: string;
  report?: {
    summary: string;
    binding: {
      revision: string;
      source_sha256: string;
      fixture_sha256: string;
      tool: string;
    };
  };
};
export const validID = (id: string) => /^[a-z][a-z0-9-]{0,79}$/.test(id);
const hash = async (bytes: ArrayBuffer) =>
  [...new Uint8Array(await crypto.subtle.digest("SHA-256", bytes))]
    .map((x) => x.toString(16).padStart(2, "0"))
    .join("");
export async function loadManifest(): Promise<Manifest> {
  const response = await fetch("/docs-manifest.json", {
    credentials: "omit",
    redirect: "error",
  });
  if (!response.ok) throw new Error("Documentation manifest unavailable");
  const value: Manifest = await response.json();
  if (
    value.version !== "fecim-docs/1" ||
    value.document_count !== value.documents.length ||
    value.documents.length > 2048 ||
    value.decoded_bytes > 10 * 1024 * 1024 ||
    value.documents.some((d) => !validID(d.id))
  )
    throw new Error("Invalid documentation manifest");
  return value;
}
export async function loadDocument(entry: Entry): Promise<Document> {
  // Construct the URL from the stable ID, not an imported path or manifest URL.
  if (!validID(entry.id)) throw new Error("Invalid document ID");
  const response = await fetch("/docs-content/" + entry.id + ".json", {
    credentials: "omit",
    redirect: "error",
  });
  if (!response.ok) throw new Error("Document unavailable");
  const bytes = await response.arrayBuffer();
  if (
    bytes.byteLength > 8 * 1024 * 1024 ||
    (await hash(bytes)) !== entry.content_sha256
  )
    throw new Error("Document content hash mismatch");
  const doc: Document = JSON.parse(new TextDecoder().decode(bytes));
  if (doc.id !== entry.id || doc.sha256 !== entry.sha256)
    throw new Error("Document identity mismatch");
  return doc;
}
export function docsWorker() {
  let client: InstanceType<typeof HysteresisWorkerClient> | undefined,
    starting: Promise<void> | undefined;
  let tail: Promise<unknown> = Promise.resolve();
  const ready = () =>
    (starting ??= (async () => {
      const m = await manifest(),
        entry = m.modules.docs;
      if (!entry)
        throw new Error("Go documentation search unavailable in this build");
      client = new HysteresisWorkerClient({
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
      await client.ready;
    })().catch((error) => {
      client?.terminate();
      client = undefined;
      starting = undefined;
      throw error;
    }));
  return {
    call(action: string, input: Record<string, string> = {}) {
      const result = tail.then(async () => {
        await ready();
        try {
          const reply = JSON.parse(await client!.request(action, input));
          if (reply.error) throw new Error(reply.error);
          return reply.result as Output;
        } catch (error) {
          if (client?.state === "failed") {
            client.terminate();
            client = undefined;
            starting = undefined;
          }
          throw error;
        }
      });
      tail = result.catch(() => {});
      return result;
    },
    close() {
      client?.terminate();
      client = undefined;
      starting = undefined;
    },
  };
}
export function snapshotOutput(s: Schema<"Snapshot">): Output {
  const field = s.fields?.find((f) => f.id === "docs");
  if (!field) throw new Error("Go documentation output unavailable");
  return JSON.parse(field.input);
}

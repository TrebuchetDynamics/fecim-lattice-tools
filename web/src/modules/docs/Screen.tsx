import { useEffect, useRef, useState } from "react";
import { Reader } from "./Reader";
import {
  docsWorker,
  loadDocument,
  loadManifest,
  validID,
  type Document,
  type Entry,
  type History,
  type Manifest,
  type Output,
  type Page,
} from "./data";
import type { LocalDocs } from "./local";
import "./style.css";

const emptyHistory: History = { favorites: [], recent: [] };
function route() {
  const match = location.hash.match(
    /^#doc-([a-z][a-z0-9-]{0,79}):([a-z0-9-]*)$/,
  );
  return {
    id: match && validID(match[1]) ? match[1] : "curriculum",
    anchor: match?.[2] || undefined,
  };
}
export default function Screen() {
  const [catalog, setCatalog] = useState<Manifest>(),
    [entries, setEntries] = useState<Entry[]>([]),
    [selected, setSelected] = useState(route().id),
    [doc, setDoc] = useState<Document>();
  const [query, setQuery] = useState(""),
    [category, setCategory] = useState(""),
    [page, setPage] = useState<Page>(),
    [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [announcement, setAnnouncement] = useState("");
  const [history, setHistory] = useState(emptyHistory),
    [output, setOutput] = useState<Output>();
  const [localReady, setLocalReady] = useState(false),
    [roots, setRoots] = useState<{ id: string; name: string }[]>([]),
    [root, setRoot] = useState(""),
    [file, setFile] = useState(""),
    [importID, setImportID] = useState("local-lesson"),
    [project, setProject] = useState(""),
    [optin, setOptin] = useState(false);
  const cache = useRef(new Map<string, Document>()),
    engine = useRef<ReturnType<typeof docsWorker> | undefined>(undefined),
    workerHistoryReady = useRef<Promise<void> | undefined>(undefined),
    native = useRef<LocalDocs | undefined>(undefined),
    intent = useRef<string | undefined>(
      location.hash.startsWith("#doc-") ? route().anchor || "" : undefined,
    ),
    mounted = useRef(true);
  const reader = useRef<HTMLDivElement>(null);
  useEffect(() => {
    mounted.current = true;
    loadManifest()
      .then((m) => {
        if (mounted.current) {
          setCatalog(m);
          setEntries(m.documents);
        }
      })
      .catch((e) => setError(e.message));
    return () => {
      mounted.current = false;
      engine.current?.close();
      void native.current?.close().catch(() => {});
    };
  }, []);
  useEffect(() => {
    if (!catalog) return;
    let active = true;
    // Do not expose the previous article with a new ID's actions while a
    // navigation is loading or has failed.
    setDoc(undefined);
    setError("");
    const entry = entries.find((e) => e.id === selected);
    if (!entry) {
      setError("Unknown document ID. Choose a page from the library.");
      return;
    }
    const work = cache.current.has(selected)
      ? Promise.resolve(cache.current.get(selected)!)
      : loadDocument(entry);
    work
      .then((value) => {
        if (active) {
          cache.current.set(value.id, value);
          setDoc(value);
          setError("");
        }
      })
      .catch((e) => {
        if (active) setError(e.message);
      });
    return () => {
      active = false;
    };
  }, [catalog, entries, selected]);
  useEffect(() => {
    if (doc?.id !== selected || intent.current === undefined) return;
    const target = intent.current
      ? document.getElementById("doc-" + selected + ":" + intent.current)
      : reader.current?.querySelector<HTMLElement>("h2,h3,h4,h5,h6");
    if (target) {
      target.focus();
      target.scrollIntoView({ block: "start" });
      intent.current = undefined;
    }
  }, [doc, selected]);
  useEffect(() => {
    const pop = () => {
      if (
        !location.pathname.endsWith("docs") &&
        !location.pathname.endsWith("/docs/")
      )
        return;
      const r = route();
      intent.current = r.anchor || "";
      setSelected(r.id);
    };
    window.addEventListener("popstate", pop);
    return () => window.removeEventListener("popstate", pop);
  }, []);
  function accept(out: Output) {
    setOutput(out);
    setHistory(out.history);
  }
  async function call(action: string, input: Record<string, string> = {}) {
    if (native.current) return native.current.call(action, input);
    if (!engine.current) {
      const worker = docsWorker();
      engine.current = worker;
      workerHistoryReady.current = (async () => {
        for (const id of history.favorites)
          await worker.call("favorite", { id });
        for (const id of [...history.recent].reverse())
          await worker.call("open", { id });
        if (output?.report)
          await worker.call("evidence", {
            report: JSON.stringify(output.report),
          });
      })();
    }
    try {
      await workerHistoryReady.current;
      return await engine.current.call(action, input);
    } catch (error) {
      engine.current?.close();
      engine.current = undefined;
      workerHistoryReady.current = undefined;
      throw error;
    }
  }
  async function perform(work: () => Promise<void>) {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      await work();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  function open(id: string, anchor?: string) {
    if (!entries.some((e) => e.id === id)) {
      setError("Document is not in this approved corpus.");
      return;
    }
    historyPush(id, anchor);
    intent.current = anchor || "";
    if (doc?.id === id) {
      const target = anchor
        ? document.getElementById("doc-" + id + ":" + anchor)
        : reader.current?.querySelector<HTMLElement>("h2,h3,h4,h5,h6");
      target?.focus();
      target?.scrollIntoView({ block: "start" });
      intent.current = undefined;
    }
    setSelected(id);
    setHistory((old) => ({
      ...old,
      recent: [id, ...old.recent.filter((x) => x !== id)].slice(0, 20),
    }));
    if (engine.current || native.current)
      void perform(async () => {
        const out = await call("open", { id });
        accept(out);
      });
  }
  function historyPush(id: string, anchor?: string) {
    const target =
      anchor ||
      cache.current.get(id)?.headings[0]?.anchor ||
      entries.find((e) => e.id === id)?.headings[0]?.anchor ||
      "";
    window.history.pushState(null, "", "#doc-" + id + ":" + target);
  }
  async function search(number = 1) {
    await perform(async () => {
      const out = await call("search", {
        query,
        category,
        page: String(number),
      });
      accept(out);
      setPage(out.page);
      setAnnouncement(
        `${out.page.total} matches. Page ${out.page.page}; ${out.page.results.length} shown.`,
      );
    });
  }
  async function connect() {
    await perform(async () => {
      const api = await import("./local");
      const connection = await api.localDocs();
      try {
        const rs = await connection.roots();
        for (const id of history.favorites)
          await connection.call("favorite", { id });
        for (const id of [...history.recent].reverse())
          await connection.call("open", { id });
        native.current = connection;
        engine.current?.close();
        engine.current = undefined;
        setRoots(rs);
        setRoot(rs[0]?.id || "");
        setLocalReady(true);
        setAnnouncement(
          "Local documentation capability connected. Session favorites retained; history remains off.",
        );
      } catch (error) {
        await connection.close().catch(() => {});
        throw error;
      }
    });
  }
  async function importGranted() {
    await perform(async () => {
      if (!native.current) throw new Error("Connect local grants first");
      const ref = await native.current.bind(root, file);
      const out = await native.current.call(
        "import_granted",
        { id: importID },
        [ref],
      );
      accept(out);
      if (out.document) {
        const d = out.document;
        cache.current.set(d.id, d);
        setEntries((old) => [
          ...old,
          { ...d, content: "", content_sha256: "" },
        ]);
        intent.current = "";
        setSelected(d.id);
        setDoc(d);
        historyPush(d.id);
        setAnnouncement(
          "Granted document indexed locally; imported/unverified. Nothing was published.",
        );
      }
    });
  }
  async function enableHistory() {
    await perform(async () => {
      try {
        if (!native.current) throw new Error("Connect local grants first");
        const p = await native.current.openProject(root, project);
        await native.current.call("history_optin", { enabled: "true" }, [
          { id: p.id, kind: "project" },
        ]);
        setOptin(true);
        setAnnouncement(
          "History enabled for the selected granted project. Save and load remain explicit.",
        );
      } catch (error) {
        setOptin(false);
        throw error;
      }
    });
  }
  async function importEvidence(file?: File) {
    if (!file) return;
    await perform(async () => {
      if (file.size > 65536) throw new Error("Evidence report exceeds 64 KiB");
      const out = await call("evidence", { report: await file.text() });
      accept(out);
      setAnnouncement(out.evidence_status);
    });
  }
  const favorite = history.favorites.includes(selected),
    categories = [...new Set(entries.map((e) => e.category))].sort();
  const termAnchor = (term: string) => {
    const aliases: Record<string, string> = { "Remnant Polarization": "Pr" };
    return entries
      .find((e) => e.id === "glossary")
      ?.headings.find((h) => h.text === (aliases[term] || term))?.anchor;
  };
  return (
    <section className="section docs-screen" aria-busy={busy}>
      <header className="docs-heading">
        <p className="eyebrow">LEARN / REFERENCE</p>
        <h1>Documentation</h1>
        <p>Read the models. Check the assumptions. Follow the evidence.</p>
        <p data-testid="corpus-count">
          {output?.document_count ?? catalog?.document_count ?? "…"} indexed
          documents · {output?.decoded_bytes ?? catalog?.decoded_bytes ?? "…"}{" "}
          decoded bytes · {catalog?.page_size ?? 20} results per page
        </p>
      </header>
      <p className="docs-notice">
        Educational simulation, not measured hardware. Legacy learning text
        retains its source wording; quantitative claims require the honesty
        audit and source conditions. Reading never runs checks or stops an
        experiment. Orca qualification is not run.
      </p>
      <form
        className="docs-search"
        onSubmit={(e) => {
          e.preventDefault();
          void search();
        }}
      >
        <label>
          Search documentation
          <input
            maxLength={256}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="A term, author or DOI"
          />
        </label>
        <label>
          Category
          <select
            value={category}
            onChange={(e) => setCategory(e.target.value)}
          >
            <option value="">All categories</option>
            {categories.map((c) => (
              <option key={c}>{c}</option>
            ))}
          </select>
        </label>
        <button disabled={busy} type="submit">
          Search
        </button>
      </form>
      <p role="status" aria-live="polite">
        {announcement}
      </p>
      {error && <p role="alert">{error}</p>}
      {page && (
        <section aria-label="Search results" className="docs-results">
          <h2>Search results</h2>
          <p>
            {page.query
              ? `Query “${page.query}”`
              : "Empty query: enter a term to search"}{" "}
            · {page.total} matches
          </p>
          <ol start={(page.page - 1) * page.page_size + 1}>
            {page.results.map((hit) => (
              <li key={hit.DocPath}>
                <button
                  onClick={() => open(hit.DocPath)}
                  aria-label={"Read " + hit.Title}
                >
                  {hit.Title}
                </button>
                <span> {hit.Category}</span>
                <p>{hit.Snippet}</p>
              </li>
            ))}
          </ol>
          <div className="docs-actions">
            <button
              disabled={busy || page.page <= 1}
              onClick={() => void search(page.page - 1)}
            >
              Previous results
            </button>
            <span>Page {page.page}</span>
            <button
              disabled={busy || page.page * page.page_size >= page.total}
              onClick={() => void search(page.page + 1)}
            >
              Next results
            </button>
          </div>
        </section>
      )}
      <div className="docs-layout">
        <aside className="docs-library">
          <nav aria-label="Documentation library">
            <h2>Library</h2>
            <div className="docs-quick">
              {[
                ["curriculum", "Learning path"],
                ["glossary", "Glossary"],
                ["citations", "Citations"],
                ["validation", "Validation and evidence"],
              ].map(([id, label]) => (
                <button disabled={busy} key={id} onClick={() => open(id)}>
                  {label}
                </button>
              ))}
            </div>
            {categories.map((c) => (
              <details open key={c}>
                <summary>{c}</summary>
                <ul>
                  {entries
                    .filter((e) => e.category === c)
                    .map((e) => (
                      <li key={e.id}>
                        <button
                          disabled={busy}
                          aria-label={"Library: " + e.title}
                          aria-current={e.id === selected ? "page" : undefined}
                          onClick={() => open(e.id)}
                        >
                          {e.title}
                        </button>
                      </li>
                    ))}
                </ul>
              </details>
            ))}
          </nav>
          <details>
            <summary>Favorites ({history.favorites.length})</summary>
            <ul>
              {history.favorites.map((id) => (
                <li key={id}>
                  <button onClick={() => open(id)}>
                    {entries.find((e) => e.id === id)?.title || id}
                  </button>
                </li>
              ))}
            </ul>
          </details>
          <details>
            <summary>Recent ({history.recent.length})</summary>
            <ul>
              {history.recent.map((id) => (
                <li key={id}>
                  <button onClick={() => open(id)}>
                    {entries.find((e) => e.id === id)?.title || id}
                  </button>
                </li>
              ))}
            </ul>
          </details>
          <p>
            Favorites and recent pages are ephemeral unless local history is
            explicitly enabled and saved.
          </p>
        </aside>
        <div className="docs-content" ref={reader}>
          {doc && (
            <>
              <nav aria-label="Document breadcrumb">
                <button onClick={() => open("curriculum")}>
                  Documentation
                </button>
                <span>
                  {" "}
                  / {doc.category} / {doc.title}
                </span>
              </nav>
              <div className="docs-actions">
                <button
                  disabled={busy}
                  onClick={() =>
                    void perform(async () =>
                      accept(await call("favorite", { id: selected })),
                    )
                  }
                >
                  {favorite ? "Remove favorite" : "Add favorite"}
                </button>
                <a
                  href="#doc-top"
                  onClick={(e) => {
                    e.preventDefault();
                    reader.current?.querySelector<HTMLElement>("h2")?.focus();
                  }}
                >
                  Reader start
                </a>
              </div>
              <nav aria-label="On this page" className="docs-toc">
                <details open>
                  <summary>
                    On this page · {doc.headings.length} headings
                  </summary>
                  <ol>
                    {doc.headings.map((h) => (
                      <li key={h.anchor}>
                        <a
                          href={"#doc-" + doc.id + ":" + h.anchor}
                          onClick={(e) => {
                            e.preventDefault();
                            const target = document.getElementById(
                              "doc-" + doc.id + ":" + h.anchor,
                            );
                            target?.focus();
                            target?.scrollIntoView({ block: "start" });
                            historyPush(doc.id, h.anchor);
                          }}
                        >
                          {h.text}
                        </a>
                      </li>
                    ))}
                  </ol>
                </details>
              </nav>
              {doc.terms.length > 0 && (
                <details className="docs-terms">
                  <summary>Glossary terms in this page</summary>
                  <ul>
                    {doc.terms.map((term) => (
                      <li key={term}>
                        {termAnchor(term) ? (
                          <button
                            onClick={() => open("glossary", termAnchor(term))}
                          >
                            {term}
                          </button>
                        ) : (
                          term
                        )}
                      </li>
                    ))}
                  </ul>
                </details>
              )}
              {selected === "validation" && (
                <section
                  className="docs-evidence"
                  aria-label="Recorded evidence"
                >
                  <h2>Candidate evidence status</h2>
                  <p data-testid="evidence-status">
                    {output?.evidence_status || "not run/unverified"}
                  </p>
                  <p>
                    No trusted candidate-bound report is bundled. Importing is
                    read-only and never establishes verification.
                  </p>
                  <label>
                    Import evidence report
                    <input
                      type="file"
                      accept="application/json,.json"
                      disabled={busy}
                      onChange={(e) => void importEvidence(e.target.files?.[0])}
                    />
                  </label>
                  {output?.report && (
                    <>
                      <p>{output.report.summary}</p>
                      <dl>
                        <dt>Revision</dt>
                        <dd>{output.report.binding.revision || "Absent"}</dd>
                        <dt>Source SHA-256</dt>
                        <dd>
                          {output.report.binding.source_sha256 || "Absent"}
                        </dd>
                        <dt>Fixture SHA-256</dt>
                        <dd>
                          {output.report.binding.fixture_sha256 || "Absent"}
                        </dd>
                        <dt>Tool</dt>
                        <dd>{output.report.binding.tool || "Absent"}</dd>
                      </dl>
                    </>
                  )}
                </section>
              )}
              <Reader doc={doc} onOpen={open} />
            </>
          )}
          {!doc && (
            <p>
              {error
                ? "Document unavailable. Choose a page from the library."
                : "Loading document…"}
            </p>
          )}
        </div>
      </div>
      {__FECIM_PROFILE__ === "local" && (
        <details className="docs-local">
          <summary>Local granted documents and opt-in history</summary>
          <p>
            Only explicit grants are read. Additional documents remain
            local/unverified and are never published. History uses
            docs-history.json inside an explicitly selected, opened project; it
            is off by default. No .omc or browser storage is used.
          </p>
          {!localReady ? (
            <button disabled={busy} onClick={() => void connect()}>
              Connect local document grants
            </button>
          ) : (
            <>
              <label>
                Granted root
                <select value={root} onChange={(e) => setRoot(e.target.value)}>
                  {roots.map((r) => (
                    <option key={r.id} value={r.id}>
                      {r.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Markdown path within granted root
                <input value={file} onChange={(e) => setFile(e.target.value)} />
              </label>
              <label>
                Stable local document ID
                <input
                  maxLength={80}
                  value={importID}
                  onChange={(e) => setImportID(e.target.value)}
                />
              </label>
              <button
                disabled={busy || !root || !file}
                onClick={() => void importGranted()}
              >
                Read granted Markdown
              </button>
              <label>
                Project path within granted root
                <input
                  value={project}
                  onChange={(e) => setProject(e.target.value)}
                />
              </label>
              <label>
                <input
                  type="checkbox"
                  checked={optin}
                  disabled={busy}
                  onChange={(e) => {
                    if (e.target.checked) {
                      setOptin(true);
                      void enableHistory();
                    } else {
                      setOptin(false);
                      void perform(async () => {
                        await native.current!.call("history_optin", {
                          enabled: "false",
                        });
                      });
                    }
                  }}
                />
                Opt in to history in this granted project
              </label>
              <div className="docs-actions">
                <button
                  disabled={busy || !optin}
                  onClick={() =>
                    void perform(async () => {
                      accept(await native.current!.call("history_load"));
                      setAnnouncement(
                        "History loaded from the granted project.",
                      );
                    })
                  }
                >
                  Load history
                </button>
                <button
                  disabled={busy || !optin}
                  onClick={() =>
                    void perform(async () => {
                      accept(await native.current!.call("history_save"));
                      setAnnouncement("History saved in the granted project.");
                    })
                  }
                >
                  Save history
                </button>
              </div>
            </>
          )}
        </details>
      )}
    </section>
  );
}

import { createElement, type ReactNode } from "react";
import type { Document, Node } from "./data";
import { validID } from "./data";

export function Reader({
  doc,
  onOpen,
}: {
  doc: Document;
  onOpen: (id: string, anchor?: string) => void;
}) {
  function render(n: Node, key: number, header = false): ReactNode {
    const children = n.children?.map((child, i) => render(child, i, header));
    switch (n.kind) {
      case "text":
        return n.text;
      case "heading":
        return createElement(
          "h" + Math.min(6, Math.max(2, (n.level || 1) + 1)),
          { id: "doc-" + doc.id + ":" + n.anchor, tabIndex: -1, key },
          children,
        );
      case "paragraph":
        return <p key={key}>{children}</p>;
      case "pre":
        return (
          <pre key={key} tabIndex={0}>
            <code>{n.text}</code>
          </pre>
        );
      case "code":
        return <code key={key}>{children}</code>;
      case "strong":
        return <strong key={key}>{children}</strong>;
      case "em":
        return <em key={key}>{children}</em>;
      case "del":
        return <del key={key}>{children}</del>;
      case "list":
        return n.ordered ? (
          <ol key={key}>{children}</ol>
        ) : (
          <ul key={key}>{children}</ul>
        );
      case "item":
        return <li key={key}>{children}</li>;
      case "quote":
        return <blockquote key={key}>{children}</blockquote>;
      case "rule":
        return <hr key={key} />;
      case "table":
        return (
          <div
            className="docs-table"
            key={key}
            tabIndex={0}
            role="region"
            aria-label="Scrollable document table"
          >
            <table>
              <caption className="sr-only">
                {doc.title} — document table
              </caption>
              <thead>
                {n.children
                  ?.filter((c) => c.kind === "table-head")
                  .map((c, i) => render(c, i, true))}
              </thead>
              <tbody>
                {n.children
                  ?.filter((c) => c.kind !== "table-head")
                  .map((c, i) => render(c, i))}
              </tbody>
            </table>
          </div>
        );
      case "table-head":
      case "row":
        return <tr key={key}>{children}</tr>;
      case "cell":
        return header ? (
          <th scope="col" key={key}>
            {children}
          </th>
        ) : (
          <td key={key}>{children}</td>
        );
      case "link": {
        if (n.url?.startsWith("doc:")) {
          const [id, anchor] = n.url.slice(4).split("#");
          if (validID(id) && (!anchor || /^[a-z0-9-]+$/.test(anchor)))
            return (
              <a
                key={key}
                href={"#doc-" + id + ":" + (anchor || "")}
                onClick={(e) => {
                  e.preventDefault();
                  onOpen(id, anchor);
                }}
              >
                {n.text}
                {children}
              </a>
            );
        }
        // Defense in depth: even a malicious imported AST cannot create executable
        // links, remote images, arbitrary elements or event attributes.
        try {
          const u = new URL(n.url || "");
          if (
            ["http:", "https:"].includes(u.protocol) &&
            !u.username &&
            !u.password
          )
            return (
              <a key={key} href={u.href} rel="noreferrer noopener">
                {n.text}
                {children}
              </a>
            );
        } catch {}
        return (
          <span key={key}>
            {n.text}
            {children}
          </span>
        );
      }
      default:
        return (
          <span key={key}>
            {n.text}
            {children}
          </span>
        );
    }
  }
  return (
    <article aria-label="Document reader" className="docs-reader">
      {doc.nodes?.map((n, i) => render(n, i))}
      <footer className="docs-source">
        <p>
          Source: <code>{doc.path}</code> · {doc.bytes} bytes
        </p>
        <p>
          SHA-256: <code>{doc.sha256}</code>
        </p>
        <p>
          {doc.rights} · {doc.attribution}
        </p>
      </footer>
    </article>
  );
}

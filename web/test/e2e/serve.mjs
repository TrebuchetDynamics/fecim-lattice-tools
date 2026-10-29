// Worker-private static server. Optional aggregate link shaping affects this
// process only, never host networking. Brotli sizes are actual response bytes.
import { createServer } from "node:http";
import { readFileSync, statSync } from "node:fs";
import { resolve, extname } from "node:path";
import { brotliCompressSync, constants } from "node:zlib";
const port = Number(process.argv[2]);
const root = resolve(
  process.env.FECIM_WEB_PROFILE === "local" ? "dist-local" : "dist",
);
const shaped = process.argv.includes("--shaped");
const cache = new Map();
let nextTransfer = 0;
function asset(path) {
  if (cache.has(path)) return cache.get(path);
  const raw = readFileSync(path),
    br = brotliCompressSync(raw, {
      params: { [constants.BROTLI_PARAM_QUALITY]: 11 },
    });
  const value = { raw, br };
  cache.set(path, value);
  return value;
}
const server = createServer((req, res) => {
  if (req.method !== "GET") {
    res.writeHead(405).end();
    return;
  }
  let path;
  try {
    const pathname = decodeURIComponent(
      new URL(req.url, "http://static.invalid").pathname,
    );
    path = resolve(
      root,
      "." +
        (/^\/(?:hysteresis|projects|configuration|experiment|results|docs)?\/?$/.test(pathname) || /^\/modules\/[a-z][a-z0-9-]*\/?$/.test(pathname)
          ? "/index.html"
          : pathname),
    );
    if (!path.startsWith(root + "/") || !statSync(path).isFile())
      throw Error("not an asset");
  } catch {
    res.writeHead(404).end();
    return;
  }
  const { raw, br } = asset(path);
  const compressed = /\bbr\b/.test(req.headers["accept-encoding"] || "");
  const data = compressed ? br : raw;
  res.setHeader(
    "Content-Type",
    {
      ".html": "text/html",
      ".js": "text/javascript",
      ".mjs": "text/javascript",
      ".css": "text/css",
      ".wasm": "application/wasm",
      ".json": "application/json",
    }[extname(path)] || "application/octet-stream",
  );
  res.setHeader("Cache-Control", "no-store");
  res.setHeader("Content-Length", data.length);
  res.setHeader("X-Fecim-Brotli-Bytes", br.length);
  if (compressed) res.setHeader("Content-Encoding", "br");
  res.setHeader("Vary", "Accept-Encoding");
  // Startup compression is precomputed below: not charged to browser CPU.
  if (!shaped) {
    res.end(data);
    return;
  }
  const chunkBytes = 16 * 1024;
  let offset = 0;
  const earliest = Date.now() + 100;
  function chunk() {
    if (res.destroyed) return;
    if (offset >= data.length) {
      res.end();
      return;
    }
    const end = Math.min(data.length, offset + chunkBytes);
    // Reserve the one shared link at 10,000,000 bits/s. Concurrent assets and
    // Worker requests cannot each receive their own nominal 10 Mb/s channel.
    const start = Math.max(earliest, Date.now(), nextTransfer);
    nextTransfer = start + ((end - offset) * 8) / 10000;
    setTimeout(
      () => {
        if (!res.destroyed) res.write(data.subarray(offset, end));
        offset = end;
        chunk();
      },
      Math.max(0, nextTransfer - Date.now()),
    );
  }
  chunk();
});
import { readdirSync } from "node:fs";
function preload(dir) {
  for (const name of readdirSync(dir)) {
    const path = resolve(dir, name);
    if (statSync(path).isDirectory()) preload(path);
    else asset(path);
  }
}
preload(root);
server.listen(port, "127.0.0.1", () =>
  console.log(
    "Static " +
      root +
      " on " +
      port +
      (shaped ? " (10 Mb/s aggregate +100 ms RTT)" : ""),
  ),
);
for (const signal of ["SIGTERM", "SIGINT"])
  process.on(signal, () => server.close(() => process.exit(0)));

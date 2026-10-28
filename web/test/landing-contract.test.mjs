import assert from "node:assert/strict";
import { test } from "node:test";
import { existsSync, readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const webRoot = dirname(fileURLToPath(import.meta.url)).replace(/\/test$/, "");
const repoRoot = join(webRoot, "..");

function readText(path) {
  return readFileSync(path, "utf8");
}

test("web deploy contract is one React application without the old Fyne demo", () => {
  const packageJSONPath = join(webRoot, "package.json");
  const homePath = join(webRoot, "src/Home.tsx");
  const wranglerPath = join(webRoot, "wrangler.toml");

  assert.equal(
    existsSync(packageJSONPath),
    true,
    "web/package.json defines the shared application",
  );
  assert.equal(
    existsSync(homePath),
    true,
    "home belongs to the shared React route tree",
  );
  assert.equal(
    existsSync(join(webRoot, "src/pages/index.astro")),
    false,
    "no second Astro home",
  );
  assert.equal(
    existsSync(wranglerPath),
    true,
    "web/wrangler.toml should define the Cloudflare Pages output",
  );

  assert.match(
    readText(join(webRoot, "index.html")),
    /src\/main\.tsx/,
    "one Vite entrypoint",
  );
  assert.equal(
    existsSync(join(webRoot, "wasm_exec.js")),
    false,
    "root Go WASM runtime shim must be removed",
  );
  assert.equal(
    existsSync(join(webRoot, "scripts/build-fyne-wasm.mjs")),
    false,
    "Fyne WASM build script must be removed",
  );
  assert.equal(
    existsSync(join(webRoot, "public/fecim")),
    false,
    "Fyne WASM public route must be removed",
  );
  assert.equal(
    existsSync(join(repoRoot, "cmd/fecim-web-fyne")),
    false,
    "Fyne WASM demo command must be removed",
  );

  const pkg = JSON.parse(readText(packageJSONPath));
  assert.equal(pkg.dependencies?.astro, undefined);
  assert.match(
    pkg.dependencies?.react ?? "",
    /^\d+\.\d+\.\d+$/,
    "React is pinned",
  );
  assert.match(
    pkg.devDependencies?.vite ?? "",
    /^\d+\.\d+\.\d+$/,
    "Vite is pinned",
  );
  assert.ok(
    pkg.scripts["build:hosted"] && pkg.scripts["build:local"],
    "both profiles build",
  );
  assert.doesNotMatch(
    JSON.stringify(pkg.scripts ?? {}),
    /fyne/i,
    "no Fyne build path",
  );
  assert.match(
    pkg.scripts?.deploy ?? "",
    /wrangler pages deploy dist/,
    "deploy should publish web/dist through Wrangler Pages",
  );

  const page = readText(homePath);
  assert.match(
    page,
    /FeCIM Lattice Tools/,
    "landing page should name the product",
  );
  assert.match(
    page,
    /Education-phase simulation/i,
    "landing page should include the honesty/accuracy banner",
  );
  assert.doesNotMatch(
    page,
    /\/fecim\/|fecim\.wasm|wasm_exec|Fyne browser demo|Fyne WebAssembly/i,
    "landing page should not reference the removed WASM/Fyne demo",
  );

  const wrangler = readText(wranglerPath);
  assert.match(
    wrangler,
    /pages_build_output_dir\s*=\s*"\.\/dist"/,
    "Wrangler should deploy the hosted dist directory",
  );
});

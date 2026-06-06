import assert from 'node:assert/strict';
import { test } from 'node:test';
import { existsSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const webRoot = dirname(fileURLToPath(import.meta.url)).replace(/\/test$/, '');
const repoRoot = join(webRoot, '..');

function readText(path) {
  return readFileSync(path, 'utf8');
}

test('web deploy contract is Astro-only with the old Go WASM/Fyne demo removed', () => {
  const packageJSONPath = join(webRoot, 'package.json');
  const astroPagePath = join(webRoot, 'src/pages/index.astro');
  const wranglerPath = join(webRoot, 'wrangler.toml');

  assert.equal(existsSync(packageJSONPath), true, 'web/package.json should define the Astro project');
  assert.equal(existsSync(astroPagePath), true, 'web/src/pages/index.astro should be the landing page entrypoint');
  assert.equal(existsSync(wranglerPath), true, 'web/wrangler.toml should define the Cloudflare Pages output');

  assert.equal(existsSync(join(webRoot, 'index.html')), false, 'old static WASM loader must be removed');
  assert.equal(existsSync(join(webRoot, 'wasm_exec.js')), false, 'root Go WASM runtime shim must be removed');
  assert.equal(existsSync(join(webRoot, 'scripts/build-fyne-wasm.mjs')), false, 'Fyne WASM build script must be removed');
  assert.equal(existsSync(join(webRoot, 'public/fecim')), false, 'Fyne WASM public route must be removed');
  assert.equal(existsSync(join(repoRoot, 'cmd/fecim-web-fyne')), false, 'Fyne WASM demo command must be removed');

  const pkg = JSON.parse(readText(packageJSONPath));
  assert.equal(pkg.dependencies?.astro !== undefined, true, 'Astro should be a runtime dependency');
  assert.match(pkg.scripts?.build ?? '', /^astro build$/, 'build script should only run astro build');
  assert.doesNotMatch(JSON.stringify(pkg.scripts ?? {}), /wasm|fyne/i, 'package scripts should not build or launch Fyne/WASM');
  assert.match(pkg.scripts?.deploy ?? '', /wrangler pages deploy dist/, 'deploy should publish web/dist through Wrangler Pages');

  const page = readText(astroPagePath);
  assert.match(page, /FeCIM Lattice Tools/, 'landing page should name the product');
  assert.match(page, /Education-phase simulation/i, 'landing page should include the honesty/accuracy banner');
  assert.doesNotMatch(page, /\/fecim\/|fecim\.wasm|wasm_exec|Fyne browser demo|Fyne WebAssembly/i, 'landing page should not reference the removed WASM/Fyne demo');

  const wrangler = readText(wranglerPath);
  assert.match(wrangler, /pages_build_output_dir\s*=\s*"\.\/dist"/, 'Wrangler should deploy the Astro dist directory');
});

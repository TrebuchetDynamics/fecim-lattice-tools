import { test } from "node:test";
import assert from "node:assert/strict";
import {
  mkdtempSync,
  readFileSync,
  writeFileSync,
  mkdirSync,
  rmSync,
} from "node:fs";
import { join } from "node:path";
import { pathToFileURL } from "node:url";
import ts from "typescript";

// Test lifecycle at the owner/transport seam; no scientific result is mocked.
const temp = mkdtempSync(join(process.env.TMPDIR, "comparison-owner-"));
test.after(() => rmSync(temp, { recursive: true, force: true }));
for (const name of ["modules/comparison/owner", "worker/ownership"]) {
  const path = join(temp, name + ".mjs");
  mkdirSync(join(path, ".."), { recursive: true });
  const js = ts
    .transpileModule(readFileSync("src/" + name + ".ts", "utf8"), {
      compilerOptions: {
        target: ts.ScriptTarget.ES2022,
        module: ts.ModuleKind.ESNext,
      },
    })
    .outputText.replace(
      /(["'])(\.\.?\/[^"']+)\1/g,
      (_, quote, path) => quote + path + ".mjs" + quote,
    );
  writeFileSync(path, js);
}
writeFileSync(
  join(temp, "modules/comparison/hosted.mjs"),
  "export const hostedComparison = () => globalThis.comparisonTransport();",
);
globalThis.__FECIM_PROFILE__ = "hosted";
const { comparisonOwner: owner } = await import(
  pathToFileURL(join(temp, "modules/comparison/owner.mjs"))
);

test("failed comparison initialization closes acquired transports before retry", async () => {
  let created = 0,
    closed = 0;
  globalThis.comparisonTransport = async () => {
    const attempt = ++created;
    return {
      snapshot: async () => {
        if (attempt === 1) throw new Error("snapshot failed");
        return { status: "ready", revision: 0 };
      },
      projects: async () => {
        if (attempt === 2) throw new Error("project discovery failed");
        return [];
      },
      close: () => {
        closed++;
      },
    };
  };
  await owner.connect();
  assert.equal(closed, 1);
  assert.equal(owner.snapshot, undefined);
  await owner.retry();
  assert.equal(closed, 2);
  assert.equal(owner.snapshot, undefined);
  await owner.retry();
  assert.equal(created, 3);
  assert.equal(closed, 2);
  assert.equal(owner.status, "ready");
  assert.equal(owner.error, undefined);
});

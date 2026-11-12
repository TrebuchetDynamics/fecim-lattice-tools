import { test } from "node:test";
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import {
  cpSync,
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  readdirSync,
  rmSync,
  symlinkSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";

const success = "Staged matching local profile for native embedding\n";
const engine = {
  UI: "fixture-ui",
  API: "fixture-api",
  identity: {
    build: "fixture-build",
    go: "go-fixture",
    model: "fixture-model",
    protocol: "fixture-protocol",
  },
};

function put(root, name, bytes) {
  const path = join(root, name);
  mkdirSync(join(path, ".."), { recursive: true });
  writeFileSync(path, bytes);
}
function snapshot(dir, prefix = "") {
  const files = {};
  if (!existsSync(dir)) return files;
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const name = prefix + entry.name;
    if (entry.isDirectory())
      Object.assign(files, snapshot(join(dir, entry.name), name + "/"));
    else files[name] = readFileSync(join(dir, entry.name));
  }
  return files;
}
function fixture(t, prior = true) {
  const root = mkdtempSync(join(process.env.TMPDIR || tmpdir(), "local-stage-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  const script = join(root, "web/src/local/stage.mjs");
  mkdirSync(join(script, ".."), { recursive: true });
  cpSync(new URL("../src/local/stage.mjs", import.meta.url), script);
  const source = join(root, "web/dist-local");
  const parent = join(root, "internal/browserui");
  const target = join(parent, "assets");
  put(parent, "unrelated/keep", "sibling bytes");
  if (prior) {
    for (const [name, bytes] of Object.entries({
      "README.txt": "sentinel\0bytes",
      "index.html": "old UI",
      "local-manifest.json": "old manifest",
      "engine/old.wasm": Buffer.from([0, 255, 7]),
    })) put(target, name, bytes);
  }
  const before = snapshot(target);
  const siblings = readdirSync(parent).sort();
  return {
    source, target, before,
    valid() {
      rmSync(source, { recursive: true, force: true });
      put(source, "index.html", "new UI");
      put(source, "engine/manifest.json", JSON.stringify(engine));
      put(source, "engine/nested/core.wasm", Buffer.from([0, 1, 255]));
    },
    failPublication() {
      const hook = join(root, "fail-publication.mjs");
      writeFileSync(hook, `
        import fs from "node:fs";
        import { syncBuiltinESMExports } from "node:module";
        const rename = fs.renameSync;
        let fail = true;
        fs.renameSync = (from, to) => {
          if (to === ${JSON.stringify(target)} && fail) {
            fail = false;
            throw new Error("fixture publication failure");
          }
          return rename(from, to);
        };
        syncBuiltinESMExports();
      `);
      return ["--import", hook];
    },
    run(args = []) {
      return spawnSync(process.execPath, [...args, script], { cwd: root, encoding: "utf8" });
    },
    clean() {
      assert.deepEqual(readdirSync(parent).sort(), siblings);
      assert.equal(readFileSync(join(parent, "unrelated/keep"), "utf8"), "sibling bytes");
    },
  };
}
function failed(result) {
  assert.equal(result.error, undefined);
  assert.notEqual(result.status, 0, result.stderr);
  assert.equal(result.stdout, "");
}
function staged(f) {
  const source = snapshot(f.source);
  const result = f.run();
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout, success);
  const manifest = JSON.parse(readFileSync(join(f.target, "local-manifest.json"), "utf8"));
  const assets = Object.fromEntries(
    Object.entries(source).map(([name, bytes]) => [
      name, createHash("sha256").update(bytes).digest("hex"),
    ]),
  );
  assert.deepEqual(manifest, { profile: "local", ...engine, assets });
  assert.deepEqual(snapshot(f.target), {
    ...source,
    "README.txt": f.before["README.txt"],
    "local-manifest.json": Buffer.from(JSON.stringify(manifest)),
  });
  assert.deepEqual(snapshot(f.source), source);
  f.clean();
}

test("preparation failures preserve the old stage, clean private files and allow retry", async (t) => {
  for (const kind of [
    "missing source", "malformed engine", "unreadable asset", "manifest write", "publication",
  ])
    for (const prior of [true, false])
      await t.test(`${kind}, ${prior ? "previous" : "fresh"} target`, (t) => {
        const f = fixture(t, prior);
        if (kind !== "missing source") f.valid();
        if (kind === "malformed engine") put(f.source, "engine/manifest.json", "{");
        if (kind === "unreadable asset")
          symlinkSync(join(f.source, "absent"), join(f.source, "broken.bin"));
        if (kind === "manifest write") mkdirSync(join(f.source, "local-manifest.json"));
        failed(f.run(kind === "publication" ? f.failPublication() : []));
        assert.deepEqual(snapshot(f.target), f.before);
        if (!prior) assert.equal(existsSync(f.target), false);
        f.clean();
        f.valid();
        // Fresh-target retries have no sentinel to retain.
        if (prior) staged(f);
        else {
          assert.equal(f.run().status, 0);
          assert.equal(existsSync(join(f.target, "local-manifest.json")), true);
          assert.deepEqual(readdirSync(join(f.target, "..")).sort(), [
            "assets", "unrelated",
          ]);
        }
      });
});

test("successful stage and restage retain sentinel, replace obsolete bytes and preserve manifest format", (t) => {
  const f = fixture(t);
  f.valid();
  staged(f);
  put(f.source, "index.html", "restaged UI");
  rmSync(join(f.source, "engine/nested"), { recursive: true });
  put(f.source, "engine/next.wasm", "next bytes");
  staged(f);
});

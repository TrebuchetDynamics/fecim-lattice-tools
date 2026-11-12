// Stage generated local assets, never user data or committed binaries.
import { createHash } from "node:crypto";
import {
  cpSync,
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  readdirSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { resolve, join } from "node:path";
const web = resolve(import.meta.dirname, "../..");
const source = join(web, "dist-local");
const target = resolve(web, "../internal/browserui/assets");
mkdirSync(resolve(target, ".."), { recursive: true });
const temporary = mkdtempSync(resolve(target, "../.assets-stage-"));
const replacement = join(temporary, "replacement");
const backup = join(temporary, "backup");
let backedUp = false;
try {
  cpSync(source, replacement, { recursive: true });
  const engine = JSON.parse(
    readFileSync(join(replacement, "engine/manifest.json"), "utf8"),
  );
  const assets = {};
  function visit(dir, prefix = "") {
    for (const item of readdirSync(dir, { withFileTypes: true })) {
      const name = prefix + item.name;
      if (item.isDirectory()) visit(join(dir, item.name), name + "/");
      else
        assets[name] = createHash("sha256")
          .update(readFileSync(join(dir, item.name)))
          .digest("hex");
    }
  }
  visit(replacement);
  writeFileSync(
    join(replacement, "local-manifest.json"),
    JSON.stringify({
      profile: "local",
      UI: engine.UI,
      API: engine.API,
      identity: engine.identity,
      assets,
    }),
  );
  if (existsSync(join(target, "README.txt")))
    cpSync(join(target, "README.txt"), join(replacement, "README.txt"));

  if (existsSync(target)) {
    renameSync(target, backup);
    backedUp = true;
  }
  try {
    renameSync(replacement, target);
  } catch (error) {
    if (backedUp) {
      try {
        renameSync(backup, target);
        backedUp = false;
      } catch (rollbackError) {
        throw new AggregateError(
          [error, rollbackError],
          `Stage rollback failed; previous assets retained at ${backup}`,
        );
      }
    }
    throw error;
  }
  backedUp = false;
} finally {
  // Never delete the previous bundle if rollback itself failed.
  if (!backedUp) rmSync(temporary, { recursive: true, force: true });
}
console.log("Staged matching local profile for native embedding");

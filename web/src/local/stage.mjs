// Stage generated local assets, never user data or committed binaries.
import { createHash } from "node:crypto";
import {
  cpSync,
  mkdirSync,
  readFileSync,
  readdirSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { resolve, join } from "node:path";
const web = resolve(import.meta.dirname, "../..");
const source = join(web, "dist-local");
const target = resolve(web, "../internal/browserui/assets");
mkdirSync(target, { recursive: true });
for (const name of readdirSync(target))
  if (name !== "README.txt")
    rmSync(join(target, name), { recursive: true, force: true });
cpSync(source, target, { recursive: true });
const engine = JSON.parse(
  readFileSync(join(source, "engine/manifest.json"), "utf8"),
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
visit(source);
writeFileSync(
  join(target, "local-manifest.json"),
  JSON.stringify({
    profile: "local",
    UI: engine.UI,
    API: engine.API,
    identity: engine.identity,
    assets,
  }),
);
console.log("Staged matching local profile for native embedding");

import { execFileSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
const web = resolve(import.meta.dirname, "../..");
const repo = resolve(web, "..");
const m = JSON.parse(
  readFileSync(
    resolve(repo, "internal/browserui/assets/local-manifest.json"),
    "utf8",
  ),
);
const fingerprint = JSON.parse(
  execFileSync("node", ["src/build/build.mjs", "--fingerprint"], {
    cwd: web,
    encoding: "utf8",
  }),
);
if (fingerprint.build !== m.identity.build || fingerprint.UI !== m.UI || fingerprint.API !== m.API)
  throw new Error(
    "browser-assets-unavailable: source drift; rerun npm run build:local before building native",
  );
const output = process.argv[2] || resolve(repo, "fecim-lattice-tools");
const flags =
  "-s -w -X fecim-lattice-tools/workbench/hysteresis.BuildID=" +
  m.identity.build +
  " -X fecim-lattice-tools/internal/browserui.UIID=" +
  m.UI;
execFileSync(
  "go",
  [
    "build",
    "-buildvcs=false",
    "-trimpath",
    "-ldflags=" + flags,
    "-o",
    output,
    "./cmd/fecim-lattice-tools",
  ],
  {
    cwd: repo,
    env: { ...process.env, GOTOOLCHAIN: m.identity.go, CGO_ENABLED: "1" },
    stdio: "inherit",
  },
);

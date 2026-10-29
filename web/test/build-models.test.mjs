import { test } from "node:test";
import assert from "node:assert/strict";
import {
  mkdtempSync,
  mkdirSync,
  writeFileSync,
  rmSync,
  readFileSync,
} from "node:fs";
import { join, resolve } from "node:path";
import { execFileSync } from "node:child_process";
import {
  buildModels,
  mergeModelDescriptors,
} from "../scripts/build-models.mjs";

test("presentation-only feature metadata is discovered without inventing a scientific WASM engine", () => {
  const repo = mkdtempSync(join(process.env.TMPDIR, "presentation-entry-"));
  try {
    const dir = join(repo, "web/src/modules/eda/presentation");
    mkdirSync(dir, { recursive: true });
    writeFileSync(
      join(dir, "identity.go"),
      `package main\nimport("fmt")\nfunc main(){fmt.Print(\`{\"identity\":{\"build\":\"fixture\",\"go\":\"fixture\"},\"descriptor\":{\"id\":\"eda\"}}\`)}`,
    );
    const modules = buildModels({
      web: join(repo, "web"),
      repo,
      out: repo,
      env: process.env,
      flags: "",
      identity: { build: "fixture", go: "fixture" },
    });
    assert.equal(modules.eda.descriptor.id, "eda");
    assert.equal(modules.eda.wasm, undefined);
  } finally {
    rmSync(repo, { recursive: true, force: true });
  }
});

test("feature-owned Go descriptors merge only known, schema/profile-paired built bundles; identity-only research remains valid", () => {
  const temp = mkdtempSync(join(process.env.TMPDIR, "model-contract-"));
  try {
    const source = join(temp, "descriptor.go");
    writeFileSync(
      source,
      `package main
import("encoding/json";"os";"fecim-lattice-tools/workbench/moduleapi";"fecim-lattice-tools/workbench/hysteresis")
func main(){d:=moduleapi.ModuleDescriptor{ID:"crossbar",Title:"Feature-owned Crossbar",Available:true,SchemaVersion:moduleapi.SchemaVersion,ModelVersion:"descriptor-contract/1",Features:[]string{"descriptor-contract-only"},Capabilities:[]moduleapi.Capability{{Profile:"hosted",Available:true},{Profile:"local",Available:false,Reason:"not shipped"}}};json.NewEncoder(os.Stdout).Encode(map[string]any{"baseline":moduleapi.BrowserCatalogue(),"sidecar":map[string]any{"identity":hysteresis.BridgeIdentity(),"descriptor":d}})}`,
    );
    const { baseline, sidecar } = JSON.parse(
      execFileSync("go", ["run", source], { cwd: resolve("..") }),
    );
    assert.equal(
      baseline.find((d) => d.id === "hysteresis").model_version,
      sidecar.identity.model,
    );
    const modules = {
      crossbar: { identity: sidecar.identity, descriptor: sidecar.descriptor },
      research: { identity: sidecar.identity },
    };
    const merged = mergeModelDescriptors(baseline, modules);
    assert.equal(
      merged.find((d) => d.id === "crossbar").title,
      "Feature-owned Crossbar",
    );
    assert.equal(
      merged.find((d) => d.id === "crossbar").capabilities[0].available,
      true,
    );
    assert.equal(merged.find((d) => d.id === "mnist").available, false);
    assert.deepEqual(Object.keys(sidecar.identity).sort(), [
      "build",
      "go",
      "model",
      "protocol",
    ]);
    for (const descriptor of [
      { ...sidecar.descriptor, id: "unknown" },
      { ...sidecar.descriptor, schema_version: "wrong" },
      {
        ...sidecar.descriptor,
        capabilities: [{ profile: "unapproved", available: true }],
      },
    ])
      assert.throws(() =>
        mergeModelDescriptors(baseline, {
          crossbar: { ...modules.crossbar, descriptor },
        }),
      );
    const redirects = readFileSync("public/_redirects", "utf8");
    for (const path of [
      "/projects",
      "/configuration",
      "/experiment",
      "/results",
      "/docs",
      "/modules/*",
    ])
      assert.ok(redirects.includes(`${path} /index.html 200`), path);
  } finally {
    rmSync(temp, { recursive: true, force: true });
  }
});

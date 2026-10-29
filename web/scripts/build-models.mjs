// Module convention: src/modules/<feature>/wasm/main_js.go. Each feature is
// compiled independently and loaded only on demand, never a universal binary.
import { readdirSync, existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { execFileSync } from "node:child_process";
export function buildModels({ web, repo, out, env, flags, identity }) {
  const modules = {};
  const root = join(web, "src/modules");
  for (const entry of readdirSync(root, { withFileTypes: true }).sort((a, b) =>
    a.name.localeCompare(b.name),
  )) {
    if (!entry.isDirectory() || !/^[a-z][a-z0-9-]*$/.test(entry.name)) continue;
    const source = `web/src/modules/${entry.name}/wasm`;
    const presentation = `web/src/modules/${entry.name}/presentation/identity.go`;
    if (!existsSync(join(repo, source, "main_js.go"))) {
      if (!existsSync(join(repo, presentation))) continue;
      const value = JSON.parse(
        execFileSync(
          "go",
          [
            "run",
            "-buildvcs=false",
            "-trimpath",
            "-ldflags=" + flags,
            join(repo, presentation),
          ],
          { cwd: repo, env },
        ),
      );
      if (
        !value.descriptor ||
        value.identity?.build !== identity.build ||
        value.identity?.go !== identity.go
      )
        throw new Error("presentation compiler identity mismatch");
      modules[entry.name] = {
        identity: value.identity,
        descriptor: value.descriptor,
        presentation: true,
      };
      continue;
    }
    const wasm = `${entry.name}.wasm`;
    execFileSync(
      "go",
      [
        "build",
        "-buildvcs=false",
        "-trimpath",
        "-ldflags=" + flags,
        "-o",
        join(out, wasm),
        "./" + source,
      ],
      {
        cwd: repo,
        env: { ...env, GOOS: "js", GOARCH: "wasm", CGO_ENABLED: "0" },
        stdio: "inherit",
      },
    );
    // Each module supplies its Go identity alongside its WASM entry, using the
    // same package as main_js.go. No central feature registry is maintained.
    const catalogue = join(repo, source, "identity.go");
    if (!existsSync(catalogue))
      throw new Error(source + "/identity.go is required");
    const value = JSON.parse(
      execFileSync(
        "go",
        ["run", "-buildvcs=false", "-trimpath", "-ldflags=" + flags, catalogue],
        { cwd: repo, env },
      ),
    );
    const pairing = value.identity || value;
    if (pairing.build !== identity.build || pairing.go !== identity.go)
      throw new Error("module compiler identity mismatch");
    modules[entry.name] = {
      wasm,
      identity: pairing,
      ...(value.descriptor ? { descriptor: value.descriptor } : {}),
      prefix: "fecimModule",
    };
  }
  return modules;
}

export function mergeModelDescriptors(baseline, modules) {
  const merged = new Map(baseline.map((d) => [d.id, d]));
  for (const [id, bundle] of Object.entries(modules)) {
    if (!bundle.descriptor) continue; // identity-only research compatibility
    const d = bundle.descriptor,
      original = merged.get(id);
    if (
      !original ||
      d.id !== id ||
      d.schema_version !== original.schema_version ||
      !Array.isArray(d.capabilities) ||
      !d.capabilities.length ||
      d.capabilities.some(
        (c) => !original.capabilities.some((p) => p.profile === c.profile),
      )
    )
      throw new Error("Invalid feature-owned module descriptor: " + id);
    merged.set(id, d);
  }
  return [...merged.values()];
}

import { defineConfig } from "vite";
import { readFileSync } from "node:fs";
import { join } from "node:path";
const dir = join(import.meta.dirname, "node_modules/.fecim");
const manifest = JSON.parse(readFileSync(join(dir, "manifest.json"), "utf8"));
export default defineConfig(({ mode }) => ({
  define: {
    __FECIM_PROFILE__: JSON.stringify(mode === "local-app" ? "local" : "hosted"),
    __FECIM_UI__: JSON.stringify(manifest.UI),
    __FECIM_API__: JSON.stringify(manifest.API),
  },
  build: {
    outDir: mode === "local-app" ? "dist-local" : "dist",
    emptyOutDir: true,
  },
  plugins: [
    {
      name: "owned-go-assets",
      generateBundle() {
        for (const file of ["manifest.json", ...Object.keys(manifest.assets)])
          this.emitFile({
            type: "asset",
            fileName: "engine/" + file,
            source: readFileSync(join(dir, file)),
          });
      },
      configureServer(server) {
        server.middlewares.use("/engine", (req, res, next) => {
          const name = req.url?.slice(1);
          if (
            !name ||
            !["manifest.json", ...Object.keys(manifest.assets)].includes(name)
          )
            return next();
          res.setHeader(
            "Content-Type",
            name.endsWith(".json")
              ? "application/json"
              : name.endsWith(".wasm")
                ? "application/wasm"
                : "text/javascript",
          );
          res.end(readFileSync(join(dir, name)));
        });
      },
    },
  ],
}));

# FeCIM Lattice Tools Web

Static Astro landing page for FeCIM Lattice Tools.

## Develop

```bash
cd web
npm install
npm run dev
```

## Validate

```bash
cd web
npm test              # fast contract tests
npm run build         # Astro production build
npm run test:e2e      # optional Playwright layout tests against built dist
```

For a full local browser gate:

```bash
cd web
npm run test:e2e:build
```

The former Go WASM/Fyne browser demo was removed. The web folder is now a static marketing and project landing page. Add real native-app screenshots under `web/public/screenshots/` when they are ready, then reference them from Astro with `/screenshots/<name>.png`.

## Deploy with Cloudflare Wrangler

```bash
cd web
wrangler login
npm run deploy
```

The Wrangler Pages project name is `fecim`, the production branch is `development`, and the deploy output directory is `web/dist`.

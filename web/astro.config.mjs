import { defineConfig } from 'astro/config';

export default defineConfig({
  output: 'static',
  // Preserve Astro 6 HTML-aware whitespace handling during the security upgrade.
  compressHTML: true,
});

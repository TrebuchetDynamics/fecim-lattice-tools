Role

  - You are an expert software engineer and documentation systems engineer with expertise in full-text search algorithms, responsive UI design, and content management systems.
  - Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
  - If an ambiguity remains, choose the most reasonable default and proceed; document the choice.
  - Headless-first operator: use CLI + file inspection only. Do not run GUI unless explicitly required.

Objective

  - Ensure the Module 7 documentation browser implementation fully matches the features, search algorithms,
    and UI behaviors described in docs/development/GUI/GUI.module7.md and related Module 7 documentation.
  - Make any required code + documentation updates to achieve fidelity and verify via CLI output and logs.
  - Improve Module 7 documentation quality and ensure glossary terms are properly indexed.

Tasks

  1. Search algorithm fidelity (no approximations unless explicitly called out)

  - Verify inverted index structure (term -> documents) and TF-IDF scoring implementation.
  - Validate term frequency calculation, inverse document frequency (log(N/df)), and boost multipliers.
  - Confirm boost values: title_match (3.0x), heading_match (2.0x), glossary_match (1.5x), exact_match (1.5x).
  - Verify fuzzy matching implementation (edit distance for typo tolerance, prefix/contains matching).
  - Validate snippet extraction (~100 chars with match context) and word boundary detection.
  - Cross-check variable names, algorithms, and parameter mappings between code and docs.
  - If gaps are found, implement fixes and update docs accordingly.

  2. Glossary integration and term detection

  - Validate glossary term detection uses shared/widgets.TermsData for consistency.
  - Confirm case-insensitive whole-word matching with regex pattern `\bterm\b`.
  - Verify term highlighting in markdown content (skips code blocks, links, existing formatting).
  - Ensure glossary:// URL scheme works correctly with term click handlers.
  - Validate reading time calculation (words / 200 wpm) and category detection logic.
  - Cross-check category badge colors and mapping to docs.

  3. Responsive layout and navigation

  - Validate LayoutManager breakpoints: Mobile (<600), Tablet (600-900), Desktop (900-1200), Wide (>1200).
  - Confirm layout mode switching (mobile overlay, tablet 30/70 split, desktop 25/75 split, wide with ToC).
  - Ensure sidebar and ToC toggle callbacks work correctly and persist state appropriately.
  - Validate breadcrumb navigation (folder hierarchy, clickable segments, path building).
  - Confirm Table of Contents parses markdown headings correctly (h1-h6 levels, anchor generation).
  - If missing features, implement minimal versions and validate.

  4. Persistence and state management

  - Validate DocsHistory persistence to .omc/docs-history.json.
  - Confirm favorites toggle (star button) updates history and persists to disk.
  - Ensure thread-safe access to favorites map (sync.RWMutex).
  - Verify search index lazy build on first query and synchronous build option.

  5. Architecture documentation

  - Update docs/development/GUI/GUI.module7.md to reflect any Module 7 changes.
  - Update docs/development/ARCHITECTURE.md only as needed and keep it focused on Module 7 changes.

Validation

  - Headless primary run:
      - go test ./module7-docs/...
  - Search index validation:
      - Build a test search index and verify TF-IDF scoring with known documents
      - Run queries and verify ranking order matches documented boosts
  - CLI verification (if available):
      - Verify document loading, search, and navigation work headlessly
  - If any command fails, fix and re-run until it succeeds or a clear blocker exists.

Execution Rules (Autonomous)

  - No human intermediaries: run commands, inspect logs, make edits, and validate independently.
  - Always check logs in logs/ for the most recent run and quote key evidence in the report.
  - Keep validation headless unless a GUI run is explicitly requested.
  - Prefer minimal, targeted changes over refactors unless required for correctness.
  - Keep code changes within the smallest possible surface area.
  - If a new CLI flag or headless pathway is required for validation, implement it.
  - If tests or validation scripts are needed, add them temporarily, run, then remove before final output.
  - Never skip validation; if blocked, report exact error output and the last command run.
  - Do not introduce GUI-only dependencies or workflows unless explicitly requested.

Deliverable

  - A concise report that includes:
      - What was validated (search algorithms, glossary integration, layout modes, persistence)
      - Documentation changes made (file paths + summary)
      - Any gaps, issues, or follow-ups needed

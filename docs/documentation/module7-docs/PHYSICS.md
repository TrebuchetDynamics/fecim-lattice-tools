# Module 7: Docs - Physics

## Prerequisites

- Basic search concepts
- Markdown structure

## Core Model

- Search uses term frequency-inverse document frequency (TF-IDF).
- Breadcrumbs and ToC derive from file paths and headings.

## Key Equations (Simplified)

```
score(term, doc) = tf(term, doc) * log(N / df(term))
```

## Parameters and Units

| Symbol | Meaning | Units |
|---|---|---|
| tf | Term frequency | count |
| df | Document frequency | count |
| N | Total docs | count |

## Assumptions and Limits

- Markdown headings are the source of ToC structure.

## Where It Lives in Code

- `module7-docs/pkg/gui/embedded.go`
- `module7-docs/pkg/gui/search.go`
- `module7-docs/pkg/gui/navigation.go`

## Sources

- `docs/development/scriptReference.md#module-7-documentation-module7-docs`


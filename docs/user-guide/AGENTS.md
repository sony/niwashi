# AGENTS.md — user guide authoring guide (docs/user-guide/)

Follow [`docs/AGENTS.md`](../AGENTS.md) for prose style. Readers here are
niwashi CLI users and recipe authors. Don't explain *why* niwashi is built a
certain way; link to the relevant `docs/adr/NNNN-*.md` instead of inlining the
rationale.

## Section → Diátaxis mapping

- `getting-started/` — Tutorial. First run only; strip anything not needed to
  reach a working setup.
- `execution/`, most of `defining-recipes/` and `defining-desired-state/` —
  How-to. State the goal, then the steps.
- `recipes/` (per-technology pages) and field-listing pages (`recipe-spec.md`,
  `*-capability.md`) — Reference. List, don't narrate.
- `examples/` — worked tutorials. Link to the relevant How-to or Reference
  page instead of re-explaining a step documented elsewhere.

## Length targets

- Tutorial / How-to pages: 100–250 lines, 400–1000 words.
- Reference pages: up to ~300 lines / 1300 words. Split into multiple pages
  before going further, rather than growing one page indefinitely.

## Conventions

- Hugo front matter (`title`, `weight`) is required on every page.
- End procedural and reference pages with a `## Next Steps` section linking to
  related pages (see `defining-recipes/recipe-spec.md`).
- Use tables for field or flag listings rather than prose enumeration.
- `content/en/` and `content/ja/` are a structural mirror: the same page must
  exist with the same headings and the same `## Next Steps` targets in both
  trees, even though the prose is translated independently. A change to one
  tree's structure without the other is incomplete.

---
status: accepted
date: 2026-04-22
decision-makers: Takahiro Okayama
---

# Relative `stateChanges` Path Resolved Against a Per-`kind` Root

## Context and Problem Statement

A recipe task's `stateChanges` entries write into Niwashi's State via a `path` field. Before this decision, `path` still had to be written as the full absolute path into State. For `kind: infra`, `kind: node`, and `kind: cluster`, the writable subtree was already a target-specific location (e.g. `kind: infra` may only write under that generator's own subtree), so every entry had to repeat that target-specific root. `kind: host` was different: its writable subtree was a single fixed path shared by every host recipe, independent of which runtime resource it managed — a separate problem, addressed as part of a different design proposal and out of scope for this decision, but one that meant `host` recipes were not yet repeating a target-specific root the way the other three kinds were.

This is redundant, obscures which paths are actually writable without reading the constraint-checking code, and complicates substituting a different root at runtime (e.g. for dry-run). The question this decision answers is how `path` should be written so that it no longer needs to repeat the per-`kind` root explicitly.

A related but separate cleanup addressed in the same change: the `infra` root's `pool` key (holding the generator's created instances) is renamed to `instances`, since `pool` leaked an internal instance-allocation concept that is not meaningful to recipe authors.

## Decision Drivers

* `path` should not require repeating a prefix that is already implied by the task's `kind` and target.
* The mechanism should support substituting a different root at runtime (e.g. for dry-run) without rewriting every recipe's `path` values.

## Considered Options

* Keep `path` as an absolute path into State (status quo)
* Provide a `{{ .RootPath }}` template variable so recipes can reference the root explicitly
* Add a separate `pathPrefix` field to factor out the common prefix
* Resolve `path` as relative to a `kind`-specific root, determined implicitly from `kind` and target (chosen)

## Decision Outcome

Chosen option: "Resolve `path` as relative to a `kind`-specific root, determined implicitly from `kind` and target". This is the only option that removes the repeated prefix from recipe source rather than just renaming or relocating it, and it lets the root be swapped by the runtime without any recipe change, since the recipe never spells the root out itself. Each `kind`'s root is declared once, as a `DefaultPrefix`; which subpaths under that root a recipe may write varies by `kind` (`infra` allows two, `node` and `cluster` allow only one) — the exact per-`kind` table is documented in `docs/user-guide/content/{en,ja}/defining-recipes/state-changes.md`.

### Consequences

* Good, because a recipe no longer repeats an absolute prefix on every `stateChanges` entry; `path: "instances/my-instance"` replaces the previous fully-qualified path, and the root itself is now declared once per `kind` (as `DefaultPrefix`) rather than baked into every recipe string.
* Good, because the `infra` root's `pool` key is renamed to `instances`, which reads clearly in `stateChanges` entries and documentation without requiring readers to know the prior internal term.
* Neutral, because leading `/` in `path` is accepted and stripped rather than rejected, so both `path: "instances/x"` and `path: "/instances/x"` resolve identically; this was a deliberate leniency, not an oversight.
* Bad, because this is a breaking change with no compatibility path: every existing recipe's `stateChanges.*.path` must be rewritten from absolute to relative, and the State key rename (`pool` → `instances`) breaks compatibility with any previously persisted `state.json`.
* Bad, because a `path` containing `../` that walks outside the kind's allowed prefixes is rejected with a generic "invalid patch path" error, but a `../` that still resolves to a string within an allowed prefix (e.g. `store/../../foo` under a `.../store` prefix) is not rejected as path traversal at all. Whether it is caught afterwards, and with what error, depends on the kind and subpath involved, since `../` was never treated as a schema error in its own right; a recipe author who mistypes it that way may not learn anything is wrong until much later, if at all.

### Confirmation

- For a `kind: infra` recipe, write one `stateChanges` entry with `path: "instances/<name>"` and another with `path: "/instances/<name>"` (leading `/`); both should apply successfully and produce the same entry under that generator's `instances` map in `state.json`.
- For the same recipe, a `stateChanges` entry whose `path` names anything other than `instances/...` or `store/...` should fail at apply time with an invalid-patch-path error, confirming the declared allow-list for `kind: infra`.
- For a `kind: node` or `kind: cluster` recipe, only a `path` under `store/...` should apply successfully; any other `path`, including an empty one addressing the capability root itself, should be rejected the same way — confirming that only `store/` is writable for those two kinds.
- A `path` containing `../` that walks outside the kind's allowed prefixes (e.g. `path: "../../foo"`) should fail the same way. A `path` containing `../` that stays within an allowed prefix as a literal string (e.g. `store/../../foo` under a `.../store` prefix) is not rejected as path traversal; applying it either succeeds or fails with an unrelated error, confirming that `../` is not treated as a schema error in its own right.

## Pros and Cons of the Options

### Absolute path into State (status quo)

`path` was written as the full path from State's root, e.g. `path: "/infrastructure/generators/{{ .Target }}/pool/my-instance"` (the `infra` root's instance-collection key at the time was still `pool`; see Context).

* Good, because there is no implicit resolution step to reason about — the string in the recipe is exactly the path that gets written.
* Bad, because every entry repeats the same `kind`-specific prefix, which is pure boilerplate once a recipe has more than a couple of `stateChanges` entries.
* Bad, because nothing in the recipe schema itself signals which paths are actually writable; a recipe author has to read the constraint-checking code (or hit a validation error) to find out.
* Bad, because swapping in a different root (e.g. for dry-run) would require rewriting every `path` value, since the root is baked into each string.

### `{{ .RootPath }}` template variable

A template variable exposing the current `kind`'s root, so a recipe could write `path: "{{ .RootPath }}/instances/my-instance"`.

* Good, because it removes the need to spell out the root literally in each recipe.
* Neutral, because the root is still present in every `path` value, just as a variable reference instead of a literal string — the redundant repetition is unchanged.
* Bad, because it requires settling on a variable name and semantics (`RootPath`) that is not obviously self-explanatory next to the other template variables recipe authors already use.

### `pathPrefix` field

A new field on each `stateChanges` entry, e.g. `pathPrefix: generators`, expanding to the `kind`-specific root, with `path` then written relative to that expansion.

* Good, because it keeps `path` closer to its previous absolute form for the trailing portion.
* Bad, because it adds a second field whose relationship to `path` recipe authors must learn, increasing schema surface for a problem that a single relative `path` already solves.
* Bad, because the meaning of a `kind`'s root ends up encoded in the *value* of `pathPrefix` rather than being an intrinsic property of `kind` and target, which is a less direct expression of the same constraint the "Considered Options" above were trying to make explicit.

### Resolve `path` as relative to a `kind`-specific root, determined implicitly from `kind` and target (chosen)

`path` is written relative to a root that is implicit in the task's `kind` and target; the runtime strips any leading `/` from `path` and joins it to that root (an empty `path` resolves to the root itself).

```yaml
stateChanges:
  generate:
    op: add
    path: "instances/my-instance"          # kind: infra → /infrastructure/generators/<target>/instances/my-instance
    valueFromFile: "{{ .Outputs }}/instance.json"
```

* Good, because single and generator operations, and every `kind`, share one relative-path convention, with no per-recipe repetition of the root.
* Good, because the root is defined once per `kind` (`DefaultPrefix`), so it can be swapped by the runtime without touching recipe source.
* Neutral, because which paths are writable is now implicit in `kind` and target rather than spelled out in the recipe text — a recipe author has to know or look up the `kind`'s root/allowed-subpath table rather than reading it off the `path` string itself.

## More Information

This decision builds on the map-based `stateChanges` shape and `count`-driven loop generation established in [ADR 0001](0001-generate-multiple-state-entries-per-task-with-count-driven-statechanges-loops.md).

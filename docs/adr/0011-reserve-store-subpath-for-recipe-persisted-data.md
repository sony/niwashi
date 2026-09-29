---
status: accepted
date: 2026-04-22
decision-makers: Takahiro Okayama
---

# Reserve a `store/` Subpath per `kind` Root for Recipe-Persisted Data

## Context and Problem Statement

A recipe task's `stateChanges` write into State under a root resolved from the task's `kind` and target (see [ADR 0009](0009-relative-statechanges-path-per-kind-root.md)). Before this decision, that root had no reserved subtree for a recipe to persist its own data across runs or hand it to a later recipe. Two concrete needs motivated this: a recipe wanting to record idempotency metadata (e.g. "which version was applied last") to skip redundant re-execution, and a recipe wanting to publish data (e.g. a generated instance's connection details) that a later recipe or its own next run would read back. The only existing workaround was for a task to parse the raw State snapshot file (`NWS_STATE`) directly, with no naming convention governing where such data should live, which left every recipe author free to invent their own location and offered no path-level enforcement.

The question this decision answers is: what convention reserves a place for a recipe's own persisted data under each `kind`'s root, and how does a recipe read back data it persisted there in its own task templates?

## Decision Drivers

* The mechanism should reuse the per-`kind` relative-path root/prefix resolution already established for `stateChanges` (ADR 0009), rather than add a second, separately-declared field that could drift out of sync with it.
* The reserved name must read naturally for both a self-referential use (idempotency bookkeeping) and a hand-off-to-a-later-recipe use, ruling out a name whose connotation fits only one of the two.

## Considered Options

* Reserve a `store/` subpath under each `kind`'s existing `stateChanges` root, with no dedicated schema field (chosen)
* Reserve an `export/` subpath instead
* Reserve a `metadata/` subpath instead
* Add a `spec.store` field that declares each key explicitly, deriving the path constraint from the declaration

## Decision Outcome

Chosen option: "Reserve a `store/` subpath under each `kind`'s existing `stateChanges` root, with no dedicated schema field". This is the only option that adds no new schema surface — it is expressed entirely as one more entry in the `AllowedPrefixes` list each `kind`'s path constraint already carries (ADR 0009) — and `store` itself, unlike `export`, does not connote one-directional publication, which matters because the primary use case implemented first (idempotency bookkeeping) is self-referential rather than outward-facing.

As part of the same decision, a recipe reads back its own `store/` data through a `{{ .Store.<key> }}` template variable, populated from the calling recipe's own persisted data only; reading another recipe's `store/` from a template was intentionally left for future work.

### Consequences

* Good, because no new schema field was needed: for `infra`, `node`, and `cluster`, `store/` is just one more relative path resolved against each `kind`'s existing root, and each of those three `kind`s gained it as a small addition to its path-constraint definition. `kind: host` is an exception to that uniformity — its root is scoped by `spec.runtime` (ADR 0010) rather than by anything specific to this decision, and that root carries no subpath restriction at all, so `store/` became reachable there only as a side effect. `kind: adapter` has no `stateChanges` root at all, so `store/` does not apply to it.
* Good, because `{{ .Store.<key> }}` lets a recipe read its own prior run's data directly in a task template (e.g. inside `scriptTpl`) without parsing `NWS_STATE` itself, and a key that was never written does not make rendering fail — piping the access through the engine's `default` function (`{{ .Store.key | default "" }}`) yields an empty string for use in shell conditionals.
* Good, because `kind: node` and `kind: cluster` restrict their capability roots to *only* the `store` subpath, while `kind: infra` allows `store` alongside `instances`, and `kind: host` places no subpath restriction beyond the root itself — so `store/`'s write-scoping tightness differs by `kind`, inherited from each `kind`'s pre-existing root-constraint shape rather than from anything specific to this decision.
* Neutral, because for `kind: host`, `store/` is unreachable unless the recipe declares `spec.runtime` (ADR 0010); a `host` recipe without that declaration cannot write to `store/` or anywhere else under its root.

### Confirmation

* Write a recipe (any `kind` except `adapter`) with a task whose `scriptTpl` reads `{{ .Store.some_key | default "" }}` and a `stateChanges` entry that writes `path: "store/some_key"`. Plan and apply it twice: on the first apply, the template should render an empty string, since nothing has been written yet; after the first apply's `stateChanges` runs, a second apply's task should see the previously written value through `{{ .Store.some_key }}`.
* Attempt a `stateChanges` entry whose `path` falls outside `store/` (and, for `kind: infra`, outside `instances/` as well): applying it should fail, confirming only `store/` (plus `instances/` for `infra`) is writable for `node`, `cluster`, and `infra`.

## Pros and Cons of the Options

### Reserve a `store/` subpath under each `kind`'s existing `stateChanges` root, with no dedicated schema field (chosen)

```yaml
# kind: node
stateChanges:
  save-applied-version:
    op: add
    path: "store/applied_version"
    value: "{{ .Params.version }}"
```

* Good, because it is expressed entirely within the relative-path mechanism ADR 0009 already established — a `kind`'s path constraint simply gains `<root>/store` as one more allowed prefix.
* Good, because `store` reads naturally for the self-referential idempotency use case that motivated this decision, without implying the data is meant for outward publication.
* Neutral, because which paths under `store/` are writable is governed by the same free-form convention as the rest of `stateChanges` — recipe authors are still free to structure `store/` however they like, with no key-level validation.
* Bad, because a recipe author has no schema-level list of which keys a recipe actually writes to `store/`; that information exists only in the recipe's `stateChanges` entries and templates, not as a declared inventory.

### Reserve an `export/` subpath

* Good, because the name directly signals "this data is meant to be consumed by something else," which fits the hand-off-to-a-later-recipe use case.
* Bad, because it implies a one-directional, outward-facing semantic that doesn't fit the idempotency-bookkeeping use case, where a recipe is reading back its own data rather than publishing it for a consumer.

### Reserve a `metadata/` subpath

* Good, because "metadata" fits a recipe recording facts about its own execution, which suits idempotency bookkeeping.
* Bad, because "metadata" reads too narrowly for the hand-off-to-a-later-recipe use case, which is about substantive data (e.g. a generated resource's connection details), not incidental metadata about the recipe itself.
* Bad, because its scope overlaps with what `store` was already expected to cover, without a clear line dividing the two.

### `spec.store` field declaring keys explicitly

A dedicated field where a recipe would declare each key it intends to expose, with the path constraint derived from that declaration.

* Good, because it would give a schema-level inventory of which keys a recipe writes, checkable without reading `stateChanges` entries.
* Bad, because the existing relative-path convention (`stateChanges.*.path` under `store/`) already expresses the same constraint without a second field to keep in sync with it.
* Bad, because it adds declaration overhead — a recipe author would maintain both the field's key list and the actual `stateChanges` entries writing under those keys.

## More Information

This decision landed in the same commit as the per-`kind` relative-path resolution (ADR 0009) and the `spec.runtime`-scoped `kind: host` root (ADR 0010) — all three were part of a single change to the `stateChanges` mechanism.

Noted later (2026-08-31): for `kind: cluster`, a task decomposed into per-node executions (via a task's own `where` condition) still resolves its `store/` writes against the *cluster's* own capability path, never a per-node one, and the recipe's `{{ .Target }}` template variable stays fixed at the cluster ID in that situation too — so embedding it in a `store/` path as a per-node disambiguating subkey does not actually vary by node, and every node's write lands at the identical path. This is an inherent consequence of reusing the target resolution ADR 0009 already established, rather than a gap introduced later; a recipe author who needs one write per node has to derive a distinguishing key from data available through other means.

---
status: accepted
date: 2026-05-14
decision-makers: Takahiro Okayama
---

# Access a Dependency's `store/` Data by Extending `requires` with `as`

## Context and Problem Statement

[ADR 0011](0011-reserve-store-subpath-for-recipe-persisted-data.md) reserved a `store/` subpath under each `kind`'s capability root. It let a recipe read back its own persisted data via `{{ .Store.<key> }}`, but explicitly left reading *another* recipe's `store/` data out of scope.

Before this decision, a recipe needing data written by a recipe it depends on had only one option: parse the raw State snapshot file directly in a script. Examples include a generated instance's connection details, a control-plane join command, or a detected tool's install path. This was impractical because the path a piece of `store/` data lives under depends on factors the consuming recipe cannot know in advance — recipe ID, node ID, cluster ID, instance ID — and because it requires understanding the full State structure just to read one value.

The question this decision answers: how does a recipe declare that it wants template-level read access to a dependency's `store/` data, and how is that access exposed in task templates?

## Decision Drivers

* The mechanism should reuse the existing `requires` declaration rather than add a second, separately-tracked one. `requires` already establishes execution ordering between a recipe and its dependency; a second declaration could drift out of sync with it.
* A recipe author should not need to know whether a dependency name refers to a capability or a host-scoped tool/service. Niwashi should resolve that from the dependency's own declaration.

## Considered Options

* Extend `requires` entries to optionally take an object form with `name` + `as`, letting Niwashi resolve whether `name` refers to a capability or a host-scoped tool/service (chosen)
* Add a separate `wantData` field alongside `requires` for declaring data access
* Extend `requires` entries with explicit `capability`/`host` keys so the recipe author states the dependency's kind directly

## Decision Outcome

Chosen option: "Extend `requires` entries to optionally take an object form with `name` + `as`", because it reuses the ordering guarantee `requires` already provides instead of introducing a second declaration that could drift out of sync with it, and because it keeps the recipe author from having to know or state whether a dependency is a capability or a host-scoped tool/service.

A `requires` entry keeps its original plain-string form for ordering-only dependencies. Written as a map, it takes `name` (the same alias used by `provides`, unchanged) and an optional `as`, the key `{{ .Stores.<as>.<key> }}` addresses; `Stores` sits alongside a recipe's own `{{ .Store.<key> }}` (ADR 0011) as a separate template field. See the recipe format reference (`docs/user-guide/content/{en,ja}/defining-recipes/recipe-spec.md`, `spec.requires`) for the full field syntax.

Independently of `store` access, this decision also closes a pre-existing gap in which `kind` may depend on which: before it, whether one `kind` could declare a `requires` dependency on another was governed purely by comparing each kind's numeric execution priority (`host`=10, `infra`=20, `node`=30, `cluster`=40), so any higher-priority kind depending on any lower-priority kind passed — including `kind: node` → `kind: infra` and `kind: cluster` → `kind: infra`. Niwashi's State design deliberately separates logical configuration (`node`/`cluster`) from the physical infrastructure that backs it (`infra`); a logical-layer recipe's behavior should not be conditioned on which infrastructure generator happens to back it. A `node`/`cluster` dependency on `infra` would cut against that separation, so this decision adds an explicit per-kind allow-list, alongside the existing priority comparison, that excludes `infra` from what `node` and `cluster` may depend on. This applies at plan time to both the plain-string and object `requires` forms alike, so it blocks the dependency regardless of whether `store` access is requested. The full per-kind allow-list is documented in the same recipe format reference referenced above.

For `kind: cluster` depending on `kind: node`, `{{ .Stores.<as> }}` cannot always resolve to one specific node's data: whether it does depends on whether the *consuming* task itself runs against a resolved node (via `where` or per-node decomposition) or at the cluster level (no node resolved). The chosen design leaves this to the task's own execution unit, rather than requiring the recipe author to pick a node explicitly.

### Consequences

* Good, because a recipe's dependency on another recipe's `store/` data reuses one declaration (`requires`) rather than two, and the ordering guarantee `requires` already provides continues to apply to the same entry that grants data access.
* Good, because a recipe author does not need to know or declare whether a dependency resolves to a capability or a host-scoped tool/service; `name` uses the same `provides` alias either way, and Niwashi resolves the kind from that declaration.
* Good, because disallowing `node`/`cluster` → `infra` dependencies at plan time keeps a logical-layer recipe's behavior from being conditioned on which infrastructure generator backs it, preserving the separation between logical and physical configuration that Niwashi's State design relies on.
* Neutral, because for `kind: cluster` depending on `kind: node`, whether `{{ .Stores.<as> }}` resolves to a specific node's data is decided per-task by that task's own execution unit, not by the `requires` declaration itself. A recipe author reading only the `requires` entry cannot tell which behavior a given task will get without also checking that task's `where`/execution-unit setting.

### Confirmation

* Which recipe's data an alias resolves to, per required kind, can be confirmed by running a recipe where `kind: cluster` requires a `kind: node` recipe with `as`, once with a task whose `where` resolves it to a specific node and once with a task that stays at the cluster level — e.g. `test/e2e/testdata/remote/recipe/cluster/nws-recipe.yaml` (exercised by `TestRemote`), which requires a node-scoped recipe `as: node_status` and renders `{{ .Stores.node_status }}` from a task decomposed via `where: node.os == "linux"`.
* The per-kind allow-list can be confirmed by attempting to plan a recipe for each `(declaring kind, required kind)` pair documented in the recipe format reference: `node`/`cluster` requiring `infra` should be rejected at plan time, and every other pair should plan successfully.

## Pros and Cons of the Options

### Extend `requires` entries with an object form (`name` + `as`), kind resolved by Niwashi (chosen)

```yaml
spec:
  requires:
    - some.capability              # existing plain form: ordering only, unchanged
    - name: some.capability         # object form: store data exposed as {{ .Stores.ctrl_plane }}
      as: ctrl_plane
    - name: tool.ansible            # a host-scoped dependency resolves the same way
      as: ansible
```

* Good, because it is additive to an existing, already-understood field rather than a new one recipe authors must learn.
* Good, because the recipe author writes the same `name` they would have written for ordering-only purposes; adding `as` does not require them to also state the dependency's kind.
* Neutral, because a recipe author cannot tell from the `requires` entry alone, without also checking `provides`, whether a given `name` resolves to a capability or a host-scoped tool/service — this is now Niwashi's responsibility rather than the author's, which is the point of the option, but it does mean the entry is not fully self-describing.

### `wantData` as a separate field alongside `requires`

```yaml
spec:
  requires:
    - some.capability
  wantData:
    - name: some.capability
      as: ctrl_plane
```

* Good, because it keeps ordering-only dependencies and data-access dependencies visually and structurally distinct.
* Bad, because a data-access dependency still needs the same ordering guarantee `requires` provides, so this design still requires listing the same dependency under both fields — doubling the declaration for the same relationship rather than avoiding it.

### `requires` entries with explicit `capability`/`host` keys

```yaml
spec:
  requires:
    - capability: some.capability
      as: ctrl_plane
    - host: tool.ansible
      as: ansible
```

* Good, because the entry is self-describing: a reader can tell whether the dependency is a capability or a host-scoped tool/service without cross-referencing `provides`.
* Bad, because it pushes a distinction (capability vs. host) onto the recipe author that Niwashi can already resolve from the `provides` declaration, adding a choice for the author to get right with no corresponding benefit to the resolution logic.

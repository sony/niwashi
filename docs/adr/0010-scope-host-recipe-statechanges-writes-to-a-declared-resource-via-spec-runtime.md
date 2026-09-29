---
status: accepted
date: 2026-04-22
decision-makers: Takahiro Okayama
---

# Scope `host` Recipe `stateChanges` Writes to a Declared Resource via `spec.runtime`

## Context and Problem Statement

A `kind: host` recipe's `stateChanges` write into a subtree of State that Niwashi's per-`kind` root/prefix mechanism (see [ADR 0009](0009-relative-statechanges-path-per-kind-root.md)) resolves `path` against. Unlike `infra`, `node`, and `cluster`, whose roots are keyed off a generator, node, or cluster target, `host` recipes had no target-derived key: every `host` recipe shared one fixed root covering all host-managed resources. This meant any `host` recipe could write anywhere under that shared root — for example, a recipe installing `git` could write into the subtree that a different recipe uses to record `ansible`'s installation, with nothing in the schema preventing it. There was also no place in the recipe schema for a `host` recipe to declare what kind of resource it manages (a `tool` versus a `service`), so this distinction existed only as an implicit naming convention.

The question this decision answers is: how does a `host` recipe declare, in its schema, the specific resource it manages, so that its `stateChanges` root can be automatically scoped to that resource rather than shared across all `host` recipes?

## Decision Drivers

* Each `host` recipe's writable `stateChanges` subtree should be scoped to the specific resource it manages, not shared with every other `host` recipe.
* The recipe schema should have an explicit place to declare what kind of resource (tool or service) a `host` recipe manages, rather than relying on naming convention.
* The set of recognized resource kinds should be able to grow without requiring a new top-level schema field for each one.

## Considered Options

* Status quo: one fixed, shared root for all `host` recipes
* A dedicated top-level field per resource kind (e.g. `spec.tool`, and a future `spec.service`)
* Reuse the existing `provides` declaration's name as the scoping key
* A single `spec.runtime` field with `type` and `name` subfields (chosen)

## Decision Outcome

Chosen option: "A single `spec.runtime` field with `type` and `name` subfields". A `kind: host` recipe declares `spec.runtime.type` and `spec.runtime.name`, and its `stateChanges` root is derived from those two values as `/runtime/<type>/<name>`, using the relative-path resolution already established for every `kind` (ADR 0009). Unlike a dedicated field per resource kind, this keeps the set of recognized `type` values extensible without adding a new top-level schema field each time, and it avoids overloading `provides` — a declaration whose purpose (advertising a capability for dependency resolution) is unrelated to path scoping.

As part of the same decision, the corresponding State subtree's key was renamed from `runtime.tools` (plural) to `runtime.tool` (singular), so that it reads consistently with the field name `spec.runtime.type: tool` and with the `service` key introduced alongside it as a second recognized `type` value.

### Consequences

* Good, because a `host` recipe's `stateChanges` writes are confined to the subtree for the specific resource it declares (`/runtime/<type>/<name>`), instead of sharing one root with every other `host` recipe.
* Good, because `spec.runtime.type` gives the recipe schema an explicit, checkable declaration of what kind of resource a `host` recipe manages, extensible to resource kinds beyond `tool` without a new top-level field.
* Neutral, because `spec.runtime` is optional in the schema; a `host` recipe that omits it is not rejected when the recipe is loaded, but any `stateChanges` entry it defines is rejected when the recipe runs, since there is no resource declared to derive a root from.
* Bad, because this is a breaking change with no compatibility path: every existing `host` recipe that wrote under the old shared root must add a `spec.runtime` declaration and adjust its `stateChanges.*.path` values, and the State key rename (`tools` → `tool`) breaks compatibility with any previously persisted state.
* Bad, because `spec.runtime.name` is restricted to characters valid as a Go template map key (letters, digits, underscore, not starting with a digit), which is a narrower character set than typical resource names (e.g. package names containing `.` or `-`) — a recipe author renaming a resource to satisfy this constraint may pick a name that no longer matches the resource's own name elsewhere.

### Confirmation

* Write a `kind: host` recipe declaring `spec.runtime: {type: tool, name: <name>}`, with a `stateChanges` entry whose `path` is empty and a `value`/`valueFromFile`/`valueFromJson` to write. Plan and apply it: the entry should apply successfully, and the resulting `state.json` should record it under `runtime.tool.<name>` (or `runtime.service.<name>` for `type: service`).
* Write the same recipe but omit `spec.runtime` entirely, keeping the `stateChanges` entry. Apply it: the entry should fail to apply, since there is no declared resource to derive a root from — confirming that a `host` recipe cannot write anywhere without declaring `spec.runtime`.
* Write a `host` recipe whose `spec.runtime.name` uses a character outside what's valid as a Go template map key (e.g. a hyphen), or whose `spec.runtime.type` is neither `tool` nor `service`: loading the recipe should fail with a validation error before any task runs.

## Pros and Cons of the Options

### A single `spec.runtime` field with `type` and `name` subfields (chosen)

```yaml
kind: host
spec:
  runtime:
    type: tool
    name: ansible
  tasks:
    - name: check-ansible
      stateChanges:
        record-version:
          op: add
          path: ""   # resolves to /runtime/tool/ansible
          valueFromFile: "{{ .Outputs }}/ansible_info.json"
```

* Good, because one field covers every resource kind; `service` was already accepted as a second `type` value from this decision onward, and any further kind needs no schema change, only a new accepted value for `type`.
* Good, because `type` and `name` are visibly distinct in the schema, so which resource a recipe manages is explicit rather than inferred.
* Neutral, because a recipe author must remember to declare `spec.runtime` before writing any `stateChanges`, since there is no separate "required" schema error pointing them at the omission — the only signal is a failure at the `stateChanges` step itself.

### Status quo: one fixed, shared root for all `host` recipes

Every `host` recipe wrote under the same fixed root, with no per-recipe scoping.

* Good, because there was nothing to declare — the root never varied, so there was no new schema surface.
* Bad, because nothing prevented one `host` recipe's `stateChanges` from writing into the subtree another recipe uses to record a different resource.
* Bad, because the schema had no place to state which resource kind (tool vs. service) a recipe manages; that distinction existed only by convention.

### A dedicated top-level field per resource kind (e.g. `spec.tool`, and a future `spec.service`)

A separate top-level field for each resource kind, e.g. `spec.tool: ansible` or `spec.service: registry`.

* Good, because a recipe author sees the resource kind directly as a field name, without an intermediate `type` value to look up.
* Bad, because each new resource kind requires a new top-level field, and recipe-processing code must check which of several mutually-exclusive fields was set rather than reading one explicit `type` value.
* Bad, because nothing in the schema enforces that at most one of these fields is set on a given recipe, unlike a single `type` value which is inherently one of a fixed set.

### Reuse the existing `provides` declaration's name as the scoping key

Use the existing `spec.provides[].name` field — used elsewhere to declare capabilities a recipe makes available for dependency resolution — as the key for the `stateChanges` root as well.

* Good, because it avoids adding a new field, reusing one already present in the schema.
* Bad, because `provides` accepts multiple entries, so a recipe with more than one `provides` entry has no single, unambiguous name to derive one root from.
* Bad, because `provides` exists to declare capabilities for dependency resolution, a different concern from scoping a `stateChanges` root; conflating the two makes each harder to reason about independently.

## More Information

This decision landed in the same change as the per-`kind` relative-path resolution documented in [ADR 0009](0009-relative-statechanges-path-per-kind-root.md), which that ADR explicitly notes as "a separate problem, addressed as part of a different design proposal and out of scope for this decision" — this ADR is that decision.

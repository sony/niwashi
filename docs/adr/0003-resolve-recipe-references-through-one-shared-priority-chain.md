---
status: accepted
date: 2026-02-02
decision-makers: Takahiro Okayama
---

# Resolve Recipe References Through One Shared Priority Chain

## Context and Problem Statement

Before this decision, a recipe's `provides` entries resolved differently depending on which kind of reference pointed at them: infrastructure-provisioning references (`provisioner`, see [ADR 0002](0002-unify-infrastructure-recipe-selection-under-a-single-provisioner-field.md)) matched by simple name, `requires` resolved through an alias/selector model, and adapter references (`toolRef`) resolved through a separate alias/selector model scoped to adapter recipes only. A recipe or state author had to learn a different resolution mechanism per reference kind, `provisioner` could not be overridden by a profile the way `requires`/`toolRef` could through `capabilityBinding`/`toolAlias`, and provisioner resolution needed its own special-cased lookup code, separate from the alias/selector machinery used everywhere else. This decision unifies all three reference kinds under one resolution mechanism and priority order.

## Decision Drivers

* Reduce cognitive load: recipe/state authors should only need to learn one resolution model, not one per reference kind.
* Maintainability: a single resolution code path is easier to reason about and test than several special-cased ones.
* Flexibility: `provisioner` should be overridable by a profile binding, the same way `requires` already benefits from `capabilityBinding` and `toolRef` from `toolAlias`.

## Considered Options

* Unify around a priority chain of FQID → profile binding → alias/selector query (chosen)
* Do nothing (keep three separate resolution mechanisms)
* Unify on simple name matching only

## Decision Outcome

Chosen option: "Unify around a priority chain of FQID → profile binding → alias/selector query", because it is the only option that both removes the inconsistency across the three reference kinds and keeps the attribute-selector and profile-binding power that simple name matching would have discarded.

A reference name is resolved in two stages, shared by `provisioner`, `requires`, and `toolRef` alike. First, the name is unconditionally substituted through a profile-supplied binding table (`capabilityBinding` for `provisioner`/`requires`, `toolAlias` for `toolRef`), repeatedly until no further substitution applies. Second, the resulting name is looked up as a fully-qualified ID with a version, then as a bare ID without one, and only if neither matches is it treated as an alias/selector query (`name.key=value`) against one of two separate provider tables, one populated from every non-adapter recipe's `provides` entries, the other from adapter recipes' `provides` entries. The FQID and bare-ID lookups run against a single recipe index shared by every reference kind, so only the alias/selector step is separated by table; `toolRef` resolution rejects a match whose recipe is not an adapter as an independent check on top of that, closing the gap the table split alone leaves open for the two earlier lookups. By convention, profile-binding keys are abstract capability names rather than FQIDs, so a literal FQID input usually passes stage one unchanged and resolves directly in stage two; nothing in the implementation enforces this as a strict rule, so "FQID takes highest priority" describes the common case, not a guarantee.

Because this decision governs how a reference resolves rather than which fields exist, infrastructure-provisioning recipes also changed how they declare themselves: a `provisioner` recipe now expresses a variant via `attrs` (e.g. `name: infra.vm, attrs: {driver: vagrant}`) instead of folding it into the flattened capability name (e.g. `infra.vm.vagrant`), matching the pattern `requires`/`toolRef` providers already used. The `attrs` map itself predates this decision.

### Consequences

* Good, because `provisioner`, `requires`, and `toolRef` now share one resolution mental model and one code path, instead of three.
* Good, because `provisioner` can now be bound abstractly and overridden per-profile via `capabilityBinding`, which a fixed name-matching scheme could not offer.
* Good, because keeping `toolAlias` and `capabilityBinding` as two separate tables (rather than merging them), together with the adapter-kind check on the resolved recipe, preserves the guarantee that `toolRef` cannot resolve to a non-adapter recipe.
* Bad, because this was a breaking change: existing infrastructure-provisioning recipes and any state file whose `provisioner` value assumed the old flattened-name style had to be rewritten to the `name` + `attrs` form (or an equivalent selector query) to keep resolving correctly.

### Confirmation

* That `provisioner`, `requires`, and `toolRef` share the same resolution mechanism can be confirmed by reading `internal/workflow/phase/infra/plan.go` and `internal/workflow/planner/recipe.go` (`GetRequireVertices`, `ResolveToolAlias`): all three call `FindRecipe`/`FindRecipeAsAdapter` on the same `*cmap.NameMapper` instance, constructed once in `internal/workflow/planner/planner.go`'s `NewPlanner`. The binding-substitution stage, including its circular-binding detection, is covered by `TestNameMapper_FindRecipe`/`TestNameMapper_FindRecipe2` (`internal/cmap/name_mapper_test.go`); the FQID/bare-ID/alias-selector lookup order and the adapter-kind restriction are covered by `TestCapabilityMap_FindRecipe`/`TestCapabilityMap_FindRecipeAsAdapter` (`internal/cmap/capability_map_test.go`).
* That `provisioner` now follows the same alias-name + `attrs` resolution recipe authors already use for `requires`/`toolRef` is documented in `docs/user-guide/content/en/defining-desired-state/capabilities.md` ("provisioner (Generator)"), which shows a Generator selecting its provisioner via an alias name plus an attribute selector (`infra.vm.driver=vagrant`) rather than a flattened name — the same form `examples/k8s-on-vagrant/target.yaml` uses.

## Pros and Cons of the Options

### Unify around FQID → profile binding → alias/selector query

* Good, because it aligns all three reference kinds on one resolution mechanism.
* Good, because it lets any of the three be overridden abstractly through a profile, including `provisioner`, which previously could not be.
* Good, because attribute selectors let multiple recipes share a capability name while remaining distinguishable (e.g. a Terraform recipe for VMs vs. one for VPCs).
* Bad, because it requires infrastructure-provisioning recipe authors specifically to switch to expressing variants as `attrs` rather than continuing to encode them directly into the flattened capability name.

### Do nothing (keep three separate resolution mechanisms)

* Bad, because recipe/state authors must learn a different resolution model for `provisioner` than for `requires`/`toolRef`.
* Bad, because `provisioner` remains impossible to override via a profile.
* Bad, because provisioner resolution keeps a special-cased code path separate from the alias/selector machinery used everywhere else.

### Unify on simple name matching only

* Good, because a single flat name string is the simplest possible mental model.
* Bad, because it discards the attribute-selector model, so two recipes that share a capability name but differ by an attribute (e.g. driver, variant) could no longer be distinguished.
* Bad, because it discards profile-level binding indirection, so an abstract capability name could not be bound to a concrete implementation without hard-coding it in the recipe or state file.

## More Information

This decision extends [ADR 0002](0002-unify-infrastructure-recipe-selection-under-a-single-provisioner-field.md)'s replacement of `driver`/`class` with a single `provisioner` field so that `provisioner`'s resolution, not just its shape, matches `requires` and `toolRef`.

This decision predates this repository's own git history: it was already implemented, materially as described above, in the earliest tracked state of niwashi's capability-resolution code, and the decision itself was carried forward into a later rewrite from Python to Go, even though details of the substitution and lookup logic changed along the way. Treat the current codebase (`internal/`) and `docs/user-guide/` as authoritative if terminology or field names have since diverged from this document.

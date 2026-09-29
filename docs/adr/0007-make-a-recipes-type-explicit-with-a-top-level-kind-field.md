---
status: accepted
date: 2026-03-19
decision-makers: Takahiro Okayama
---

# Make a Recipe's Type Explicit with a Top-Level `kind` Field

## Context and Problem Statement

Before this decision, a recipe's `spec.scope` field named the execution phase a normal recipe ran in (`host`, `infra`, `node`, or `cluster`), and validation required every recipe, including an adapter recipe (one invoked from a `tool.run` action rather than run directly in a phase), to declare one of those same four values. For an adapter, though, that value was a phase name in form only: it did not determine which phase the adapter's own tasks actually ran in, since that was governed by whichever phase called the adapter via `tool.run`, together with a second, differently-scoped field — `allowedScope`, nested under `spec.adapter`, constraining which caller kind may invoke the adapter. A reader therefore had to reconcile two similarly named fields, one of which was partly vestigial for adapters. An adapter recipe was also identifiable only by the presence of that nested `adapter:` block, not by any explicit marker, and once a recipe's type was made explicit, nesting the adapter's own fields (`allowedScope`, `executionUnit`, `commands`, `tasks`) under `spec.adapter:` would add indentation without adding meaning. The question this decision answers is how a recipe should declare its type and, where applicable, its execution phase, and how an adapter recipe's fields should be laid out once that declaration exists.

## Decision Drivers

* A field name should carry one consistent meaning across a recipe; the same field should not mean "execution phase" for one recipe type and something else for another.

## Considered Options

* Add a top-level `kind` field, remove `spec.scope`, and flatten adapter-only fields directly under `spec` (chosen)
* Add `adapter` as a valid value of `spec.scope`
* Add the top-level `kind` field, but keep `spec.scope`

## Decision Outcome

Chosen option: "Add a top-level `kind` field, remove `spec.scope`, and flatten adapter-only fields directly under `spec`", because it is the only option in which each field carries exactly one meaning throughout the schema: `kind` states what a recipe is (and, for non-adapter kinds, which phase it runs in), while `allowedScope` — now unambiguous once `spec.scope` no longer exists — solely constrains which caller kind may invoke an adapter. It also makes the adapter/non-adapter distinction visible directly from `kind` rather than from the presence of a nested field, and removes a level of nesting that no longer served a purpose.

### Consequences

* Good, because a recipe's type and phase are readable from a single top-level field, without inspecting whether a nested block is present.
* Good, because `allowedScope` no longer shares a name-like relationship with a field (`spec.scope`) that meant something unrelated for adapters.
* Good, because an adapter recipe's `commands`/`tasks` sit at the same indentation level as a normal recipe's, instead of one level deeper.
* Good, because the new validation rules make several previously-implicit constraints explicit and enforced at load time (see Confirmation).
* Bad, because it is a breaking schema change: every existing recipe needed `spec.scope` replaced with a top-level `kind`, and every adapter recipe needed its `adapter:` nesting removed.

### Confirmation

* Write a normal recipe with a top-level `kind` (`host`, `infra`, `node`, or `cluster`) and no `spec.scope`, then plan/apply it: it should load and run against that phase. A recipe that declares `spec.scope` instead of `kind`, or omits `kind`, should fail to load.
* Write an adapter recipe with `kind: adapter` and `allowedScope`, `executionUnit`, `commands`, and `tasks` declared directly under `spec` — not nested under an `adapter:` block — and invoke one of its commands from another recipe's `tool.run` action: it should plan and run without an `adapter:` indentation level anywhere in the file.
* Setting `allowedScope`, `executionUnit`, or `commands` on a recipe whose `kind` is not `adapter` should fail validation, as should an adapter recipe that omits `allowedScope`, or one that sets `executionUnit: node` together with an `allowedScope` other than `cluster`.

## Pros and Cons of the Options

### Add a top-level `kind` field, remove `spec.scope`, and flatten adapter-only fields directly under `spec`

* Good, because `kind` alone tells a reader, and the loader, what a recipe is before any nested field needs to be inspected.
* Good, because `allowedScope` keeps one unambiguous meaning: the caller-kind constraint on an adapter.
* Good, because an adapter recipe's fields read at the same indentation level as a normal recipe's.
* Bad, because it is a breaking change requiring every existing recipe to be rewritten.

### Add `adapter` as a valid value of `spec.scope`

* Good, because it would have reused an existing field instead of introducing a new top-level key.
* Bad, because `scope` denotes a position in the `host` → `infra` → `node` → `cluster` execution-phase ordering, and an adapter recipe does not participate in that ordering — folding a non-phase concept into a phase-ordered field is semantically inconsistent.

### Add the top-level `kind` field, but keep `spec.scope`

* Good, because it would have been an incremental step: introducing `kind` without simultaneously removing an existing field.
* Neutral, because it still gives every recipe an explicit type marker via `kind`.
* Bad, because it leaves two fields doing work that increasingly only one of them needs to do, and lets `spec.scope` carry a different meaning depending on `kind`'s value — confusing for a reader who doesn't already know that special case exists.

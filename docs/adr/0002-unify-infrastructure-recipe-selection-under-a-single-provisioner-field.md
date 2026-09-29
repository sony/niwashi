---
status: accepted
date: 2026-02-02
decision-makers: Takahiro Okayama
---

# Unify Infrastructure Recipe Selection Under a Single `provisioner` Field

## Context and Problem Statement

Before this decision, a Generator in the state model selected its provisioning recipe via two separate fields: a `driver` (e.g. `vagrant`) and a `class` (e.g. `vm`). This was an outlier. Recipe dependencies (`requires`) and adapter references (`tool.run.toolRef`) already referenced a recipe by one capability name (a recipe's `provides.name`), not by a pair of ad hoc fields. Users had to know which `driver`/`class` combinations were valid, since that was defined only implicitly by which recipes happened to exist. `driver` alone also could not distinguish between two recipes that used the same underlying tool for different purposes (e.g. Terraform for VMs vs. Terraform for VPCs). The question this decision answers is how infrastructure provisioning recipes should be selected, and whether that reference should take the same single-name form used elsewhere.

## Decision Drivers

* Consistency: an infrastructure recipe should be referenced the same way as `requires` and `toolRef` — by one capability name — not via fields specific to infrastructure.
* Discoverability: users should not have to learn which `driver`/`class` combinations are valid. A descriptive capability name is more self-explanatory.

## Considered Options

* Replace `driver`/`class` with a single `provisioner` field resolved by capability name (chosen)
* Keep `driver`/`class`, and add an optional recipe FQID field as an override

## Decision Outcome

Chosen option: "Replace `driver`/`class` with a single `provisioner` field resolved by capability name". This removes the reference-form inconsistency between infrastructure recipe selection and every other recipe reference in the system. The alternative would still have needed an FQID override to distinguish two recipes sharing a driver, on top of the `driver`/`class` pair it kept for the common case.

An infrastructure-provisioning recipe declares the value referenced by `provisioner` as a `provides` entry, the same declaration used to advertise any other capability. A Generator is provisioned by whichever recipe that value identifies, instead of by a `driver`/`class` pair. See `docs/user-guide/content/en/defining-desired-state/generators.md` for the field's current usage.

### Consequences

* Good, because a recipe author now expresses intent as one capability name, matching how `requires` and `toolRef` already reference recipes.
* Good, because that one name is self-explanatory, unlike a `driver`/`class` pair whose valid combinations were implicit in whichever recipes existed.
* Bad, because it was a breaking change. Every state file's Generator entries that used `driver`/`class` had to be rewritten to use `provisioner`. Every infrastructure-provisioning recipe had to declare a matching `provides` entry.

### Confirmation

- Whether `provisioner` is the only field used to select an infrastructure recipe can be confirmed two ways. Check the `generator` and `generatorTemplate` schema definitions (`docs/schema/nws-state-v1.json`) for a `driver` or `class` property. Cross-check against `README.md`'s "Reference Recipes" table and `docs/user-guide/content/en/defining-desired-state/generators.md`.
- That a Generator selects a recipe using a single `provisioner` value, not a `driver`/`class` pair, can be confirmed by writing a generator with only a `provisioner` value in a state file, naming a recipe's `provides` entry. Run `plan` and `apply` against it and check that both succeed.

## Pros and Cons of the Options

### Replace `driver`/`class` with a single `provisioner` field

* Good, because it lets a recipe author identify the same underlying tool for two different purposes by name, rather than relying on `class` to carry that distinction alone.
* Bad, because it breaks every existing state file that specifies `driver`/`class`, and every infrastructure recipe's `provides` entry.

### Keep `driver`/`class`, add an optional FQID override

* Good, because it would have been backward-compatible. Existing state files would keep working unchanged.
* Bad, because it keeps two different ways of referring to the same recipe: `driver`/`class` for the common case, an FQID override for the rest. This increases the state model's complexity and risks inconsistent usage within a project.

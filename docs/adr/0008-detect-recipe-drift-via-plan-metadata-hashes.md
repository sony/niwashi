---
status: accepted
date: 2026-06-11
decision-makers: Takahiro Okayama
---

# Detect Recipe Drift Between Plan and Apply via Recorded Hashes

## Context and Problem Statement

`nwsctl plan` resolves a desired state into an execution plan (`plan.json`) by reading whatever recipes are currently reachable under `--recipe-dir`. `nwsctl apply` later reads that same `plan.json` and re-resolves recipes from `--recipe-dir` again, which may by then contain edited, added, or removed files. Before this decision, `plan.json` carried no record of which recipes (and which versions of their content) were used to produce it, so `apply` had no way to notice that a recipe changed between the two commands and could silently execute against recipes that no longer matched what was planned. Separately, `plan.json` also had no record of when it was generated or which `nwsctl` build produced it, which made it harder to debug an unexpected plan after the fact. The question this decision answers is how to record enough information in `plan.json` to detect recipe drift at apply time, and how strictly to react to it.

## Decision Drivers

* Collect the referenced-recipe set in a way that automatically covers every path by which a recipe can be reached (`requires`, `tool.run`'s `toolRef`, and jobs already placed in the execution plan), rather than relying on each call site to remember to record what it touches.
* Record plan-generation context (when it was generated, which `nwsctl` build produced it) to help debug an unexpected plan.

## Considered Options

* Collect referenced recipes at the single recipe-resolution chokepoint shared by capability (`requires`) and adapter (`toolRef`) lookups, compute each recipe's content hash there, and gate the apply-time comparison behind an opt-in `--strict-check` flag (chosen)
* After the execution plan is built, scan the whole capability map and reverse-map the jobs that ended up in the plan back to the recipes they came from
* Collect referenced recipes during the planner's DAG-construction phase (dependency-edge building and adapter-alias resolution)
* Additionally record the generating host's OS and architecture in plan metadata, for debugging

## Decision Outcome

Chosen option: "Collect referenced recipes at the single recipe-resolution chokepoint ... gate the apply-time comparison behind an opt-in `--strict-check` flag", because every way of reaching a recipe already passes through that one chokepoint, so collection stays correct automatically if a new reference path is added later, without needing every caller to separately remember to record what it resolved. `Plan.Metadata` is a nilable pointer, so a `plan.json` generated before this decision, with no `metadata` object, does not fail to decode. Recording OS/architecture was dropped because it cannot be used for the apply-time equality check that motivated this decision (applying from a different environment than the one that generated the plan is an intended use case, not drift), and its remaining value as debugging context was judged too narrow to justify the extra field.

### Consequences

* Good, because a single collection point automatically covers every current recipe-reference path (`requires`, `tool.run`'s `toolRef`, and execution-plan jobs) and any new one added later, without each caller needing its own bookkeeping.
* Good, because `--strict-check` is opt-in: a content-hash mismatch on a still-resolvable recipe is a non-blocking warning by default, and only escalates to an aborting error when explicitly requested (e.g. in CI or a production pipeline), so an existing workflow whose recipes are merely edited between plan and apply is not broken by the new check.

### Confirmation

* Run `nwsctl plan` for a desired state that resolves recipes through both a `requires` dependency and an adapter's `toolRef`, then inspect the resulting `plan.json`'s `metadata` object: it should contain `createdAt`, `nwsctlVersion`, and one `recipes` entry (`fqid` + `hash`) for each recipe reached either way.
* Edit the content of a recipe referenced by an existing plan, then run `nwsctl apply` against that plan twice: once without `--strict-check`, expecting the apply to proceed with a logged warning about the mismatched recipe, and once with `--strict-check`, expecting the apply to abort with an error before executing.
* To confirm collection stays centralized as new reference paths are added, review that any such path resolves recipe fqids through the shared helper in `internal/cmap/name_mapper.go` rather than adding its own tracking.

## Pros and Cons of the Options

### Collect at the shared recipe-resolution chokepoint

* Good, because it needs only one collection site to cover `requires`, `tool.run`'s `toolRef`, and any future reference path that goes through the same resolver.
* Neutral, because it couples plan metadata to a resolver-internal data structure (`NameMapper.ReferenceRecipes`), which the planner must know to read at the end of `MakePlan()`.

### Scan the capability map after the plan is built and reverse-map jobs to recipes

* Neutral, because it would keep collection logic out of the resolver, at the cost of a separate post-processing pass over the whole capability map.
* Bad, because filtering out recipes that exist in the capability map but were never actually reached by the plan (e.g. unused catalog entries) requires extra logic that the chokepoint approach gets for free, since only recipes that are actually resolved are ever recorded.

### Collect during DAG construction (dependency-edge building, adapter-alias resolution)

* Neutral, because it would still capture most recipe references, since most of them are discovered while building the DAG.
* Bad, because `requires`-driven and `toolRef`-driven resolution are handled by separate code paths during DAG construction, so collection logic would have to be added in at least two places instead of one, with a higher risk of one path being missed as the planner evolves.

### Additionally record OS/architecture in plan metadata

* Good, because it could have helped diagnose apply-time failures that are specific to a particular OS or architecture.
* Bad, because it has no role in the drift check that motivated this decision: applying a plan on a different machine than the one that generated it is expected to work, so OS/architecture mismatches are not something `apply` should ever treat as an error or even a meaningful warning.

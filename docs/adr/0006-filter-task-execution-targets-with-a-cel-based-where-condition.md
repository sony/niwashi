---
status: accepted
date: 2026-03-18
decision-makers: Takahiro Okayama
---

# Filter Task Execution Targets with a CEL-Based `where` Condition

## Context and Problem Statement

Recipe tasks had no way to restrict which target they applied to based on the target's own properties. This blocked two related use cases: OS- or architecture-conditional steps inside a single recipe (instead of one recipe per platform), and per-node role branching inside a `scope: cluster` recipe (e.g. treating a Kubernetes control-plane node differently from a worker node) without delegating the whole cluster to an external tool such as Ansible. The cluster case adds a wrinkle: filtering a cluster recipe's task by node means evaluating the condition once per node the cluster spans, and each node can carry cluster-specific role information (e.g. `role: master`) that only makes sense within that one cluster, not globally. The question this decision answers: how should a task declare "only run against targets matching this condition," what expression language should the condition use, and how should cluster-level node labels combine with global node labels when evaluating it.

## Decision Drivers

* The condition language needs real boolean expressiveness (AND, OR, NOT, `in`), not just single-field equality.
* Parse/evaluation errors should be clear to a recipe author, not templating-engine error noise.
* Labels scoped to one cluster must not leak into, or collide with, another cluster's view of the same node.
* Existing cluster definitions in already-persisted state files must stay valid without a migration.

## Considered Options

* CEL (Common Expression Language) expression string in a task-level `where` field (chosen)
* Reuse the existing Go template engine as the condition language
* Declarative key-value matching (implicit AND across fields, no boolean operators)
* Name the field `condition` instead of `where`
* Merge cluster-scoped node labels permanently into the global label set
* Add a separate `members` field to clusters instead of extending the existing `nodes` field

## Decision Outcome

Chosen option: "CEL (Common Expression Language) expression string in a task-level `where` field", because it composes with the existing Go-template-based recipe engine instead of overloading it with condition syntax it wasn't designed for — see Pros and Cons of the Options below for the full comparison.

To support the cluster per-node-role use case, this decision also adds an object form for a cluster's `nodes` entries — each node can carry its own `labels` — alongside the pre-existing plain string-list form, with a merge rule under which a cluster's own node labels take precedence over that node's global labels, scoped to that cluster's own recipe execution only.

### Consequences

* Good, because a single recipe can now branch by OS/architecture or by cluster-node role, instead of needing a separate recipe per platform or per role.
* Good, because a cluster recipe can express per-role logic natively, reducing (though not eliminating) delegation to an external orchestration tool.
* Good, because cluster-scoped labels cannot collide across clusters sharing a node, since the merge is rebuilt fresh from only the current target cluster on every run.
* Good, because omitting `where`, and the plain string-list form of `nodes`, both preserve prior behavior unchanged — not a breaking change for existing recipes or state files.
* Bad, because recipe authors now need to learn CEL syntax alongside the Go-template syntax already used elsewhere in recipes — two expression languages in one file format.
* Bad, because `google/cel-go` is now a required runtime dependency.

### Confirmation

* Give two tasks in a `scope: node` recipe `where` conditions combining `node.os`/`node.arch` with `&&`, `||`, or `in`, and apply it against nodes with differing OS/architecture: only the task whose condition matches a given node should execute against it.
* Give a task in a `scope: cluster` recipe a `where` referencing `node.labels.<key>`, on a cluster whose `nodes` use the object form with per-node `labels`: it should run only against nodes whose merged labels (global plus that cluster's own) satisfy the condition, and a second cluster giving a shared node a different value for the same key should be unaffected.
* Set an adapter's `executionUnit` to `node` (requiring `allowedScope: cluster`) and invoke it from a `scope: cluster` task carrying `where`: the plan should list one execution entry per matching node, instead of a single entry for the cluster as a whole.
* Apply a recipe with an invalid `where` (malformed syntax, or a nonexistent field): it should fail with a CEL-specific parse/evaluation error, not the usual template-rendering error.

## Pros and Cons of the Options

### CEL (Common Expression Language) expression string in a task-level `where` field

* Good, because it supports AND/OR/NOT/`in` and other boolean composition out of the box, with a mature Go implementation (adopted by Kubernetes admission control, among others), so niwashi did not need to design or maintain its own parser.
* Good, because it produces evaluation errors distinct from, and clearer than, Go-template execution errors.
* Good, because it leaves room to add a more concise declarative shorthand later without replacing the underlying evaluation engine.

### Reuse the existing Go template engine as the condition language

* Good, because it would have added no new dependency and no new expression language for authors to learn.
* Bad, because expressing a boolean condition through template syntax (e.g. `{{ eq .Node.Os "linux" }}`) requires awkward escaping and produces template-execution error messages that do not read as "condition failed to parse."

### Declarative key-value matching (implicit AND across fields)

* Good, because the syntax is minimal and reads naturally for the common case of matching a small number of fields.
* Bad, because it cannot express OR, NOT, or other compound conditions, which the motivating use cases (e.g. "Linux AND worker role") required.
* Neutral, because this shape was kept in mind as a possible future sugar layered on top of CEL, rather than discarded outright.

### Name the field `condition` instead of `where`

* Neutral, because the choice is naming-only; it has no effect on expressiveness or implementation.
* Bad, because "condition" reads as "run when true" without conveying that it filters a set of targets, whereas "where" reads more naturally as a filter over the list of tasks/targets being processed, consistent with how the term is used in query-like contexts.

### Merge cluster-scoped node labels permanently into the global label set

* Good, because it would have been a simpler data model with a single label map per node.
* Bad, because two different clusters sharing the same node could define the same label key with different values (e.g. `role: master` in one cluster, `role: replica` in another), and a permanent merge would let one cluster's recipe silently see the other's label value.

### Add a separate `members` field to clusters instead of extending `nodes`

* Bad, because it would have required renaming an existing, already-persisted state field, breaking backward compatibility for every existing cluster definition; extending `nodes` to accept either shape kept old state files valid unchanged.

## More Information

This ADR describes the `where`/CEL mechanism, the cluster label-merge rule, and the adapter's `allowedScope`/`executionUnit` fields using the `scope` and nested adapter schema in effect at decision time. Treat `internal/` and `docs/user-guide/` as authoritative if terminology has since diverged — including how a recipe declares its type and execution phase, which changed after this decision.

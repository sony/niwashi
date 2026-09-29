---
status: accepted
date: 2025-12-03
decision-makers: Takahiro Okayama
---

# Bind Nodes to Instances via an Explicit `instanceSelector` Field

## Context and Problem Statement

Before this decision, node-to-instance binding assigned instances by iterating over all instance generators and their pools and binding the first available instance to each node. This had three problems: the binding order depended on iteration order over an unordered collection, so it was not deterministic between runs; users had no way to say which instance should go to which node (e.g. reserving a specific VM for a control-plane role, or for debugging); and there was no way to route a node with higher resource requirements toward an instance that could satisfy them. The question this decision answers is how a user should be able to control which instance a logical node in the inventory is bound to, and how that binding should be prioritized when several nodes compete for instances at once.

## Decision Drivers

* A node's binding preference should live with the rest of the node's own definition, not in a separate file, so a single node's configuration stays readable as a whole.
* The binding relationship should be declared from the node (consumer) side, so infrastructure definitions don't need to know node identities.

## Considered Options

* Add an `instanceSelector` field on the node, with `generator` and optional `instance` (chosen)
* Define instance-to-node bindings in a separate profile-level mapping
* Define bindings from the instance/generator side (a generator lists which nodes it feeds)
* Bind by numeric pool index instead of instance ID

## Decision Outcome

Chosen option: "Add an `instanceSelector` field on the node, with `generator` and optional `instance`", because it keeps a node's binding preference in the same place as the rest of its definition and lets a node request an instance without requiring the infrastructure definition to reference node names.

A node's `instanceSelector` names a `generator`, optionally narrowed further to one `instance` within it. Binding processes nodes with both specified first, `generator` alone next, and no selector last, so a more specific request always wins over a less specific or absent one; a node with no selector is bound from whatever instances the more specific nodes left behind.

The original design also planned a separate `access` field, mutually exclusive with `instanceSelector`, for nodes with pre-existing, statically-defined connection info. This was not carried into the implementation as a second field: pre-existing/static machines are instead represented as instances from a built-in `external-instance` generator provisioner, so every node is bound through the same `instanceSelector`/`instanceRef` path and there is no second field left to guard against.

### Consequences

* Good, because a node with a more specific selector is bound before a less specific one, so an unrelated node's automatic pick is less likely to consume the pool a more specific node was drawing from.
* Bad, because binding is not fully deterministic even with a selector: which node is considered first within the same tier, and which instance is picked among several candidates in the same pool, are not otherwise ordered and can vary between runs.
* Neutral, because there is no second field left to conflict with `instanceSelector` (see Decision Outcome).
* Neutral, because condition-based binding (matching a node's resource requirements against instance metadata) was scoped as a later phase and remains unimplemented; the selector expresses instance identity, not resource requirements.

### Confirmation

* The three-tier priority order can be confirmed by writing a desired state where one node's `instanceSelector` names both `generator` and `instance`, a second names only `generator` (same generator, fewer remaining instances than competing nodes), and a third has no selector, then planning and applying it and checking that the first node is bound to the named instance, the second is still bound from that generator, and the third is bound only from whatever is left.
* That a `generator` selector narrows the search to that generator's pool can be confirmed by writing a desired state with two generators and a node selecting one of them, then planning and applying it and checking that the resulting `instanceRef` names an instance from the selected generator.

## Pros and Cons of the Options

### Add an `instanceSelector` field on the node, with `generator` and optional `instance`

* Good, because a node's binding preference stays in the same place as the rest of its definition.
* Good, because it requires no changes to existing configurations that don't use it — a node without `instanceSelector` keeps the previous "first available" behavior.

### Define instance-to-node bindings in a separate profile-level mapping

* Good, because it separates environment-specific configuration from the desired-state definition.
* Bad, because it splits a node's definition across two files, making a single node's configuration harder to read as a whole.

### Define bindings from the instance/generator side

* Good, because it gives a clear view, from the infrastructure side, of which nodes a generator's instances feed.
* Bad, because it requires the infrastructure definition to know node names, coupling infrastructure to inventory and inverting the more natural "node requests an instance" dependency direction.

### Bind by numeric pool index instead of instance ID

* Good, because a plain index is simpler to write and doesn't require knowing an instance's generated ID.
* Bad, because instance IDs are not guaranteed to map predictably onto a stable index, and an index-based reference is less self-descriptive than a named instance.
* Neutral, because the original design left this open as a possible additional selector form for a later phase rather than rejecting it outright; it was not implemented — the current schema's `instanceSelector` accepts only `generator` and `instance`.

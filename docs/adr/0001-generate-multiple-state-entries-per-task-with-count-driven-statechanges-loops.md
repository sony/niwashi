---
status: accepted
date: 2026-02-02
decision-makers: Takahiro Okayama
---

# Generate Multiple State Entries per Task with Count-Driven `stateChanges` Loops

## Context and Problem Statement

A recipe task uses `stateChanges` to write entries into Niwashi's State after the task runs. The original design represented `stateChanges` as an array of operation objects, each describing a single write (or removal) at one path.

This is sufficient for tasks that add exactly one entry to State, but it breaks down for `scope: infra` provisioner recipes that create a variable number of instances in one task (e.g. a Vagrant-based provisioner that brings up `params.count` VMs). Such a recipe needs to receive the instance count through `params`, write one State entry per created instance at a distinct path, and do so without a completely separate, parallel mechanism for "generate N entries" tasks that would coexist awkwardly with ordinary single-value operations.

The question this decision answers is how to change the `stateChanges` schema to support this.

## Decision Drivers

* A single task must be able to generate a variable, params-driven number of distinctly-addressed State entries from one declared operation, not just a single value.
* That generator capability must coexist with plain single-value operations under one schema, without a second, parallel array type dedicated to generators.

## Considered Options

* Keep `stateChanges` as an array of operation objects (status quo)
* Change `stateChanges` to a map keyed by operation name, with the presence of a `count` field marking an entry as a loop-generator

## Decision Outcome

Chosen option: "Change `stateChanges` to a map keyed by operation name, with the presence of a `count` field marking an entry as a loop-generator". This is the only option of the two that satisfies both decision drivers without introducing a second, parallel array type: a single schema lets `count`-bearing entries and plain entries live side by side in the same map. `count` can hold either a literal number or a `{{ .Params.count }}` template expression, which is resolved to a static integer before any patch is generated. Each operation also gets a name that is meaningful when reading the recipe, following the same "name each declared unit" convention used by Terraform `resource` blocks and Kubernetes `metadata.name`.

### Consequences

* Good, because one task can register an arbitrarily-sized set of State entries — e.g. one entry per provisioned VM — from a single `stateChanges` block, addressed via `{{ .Loop.index }}` in `path`, `valueFromFile`, and `valueFromJson`.
* Good, because single operations and generator operations use the same map schema and can be listed together under one task's `stateChanges`, exactly as the decision drivers required.
* Bad, because a YAML map does not guarantee key ordering the way an array preserves position; if operations in one task ever needed to run in a specific order, the schema itself gives no such guarantee.

### Confirmation

A decision at the recipe-schema level is confirmed by constructing a recipe against the shape decided here and observing it run, not by reading its Go representation.

Construct a recipe whose single task's `stateChanges` combines a plain operation (`set-provisioner-info`, writing `store/provisioner`) with a `count`-driven generator operation (`generate-instances`, addressing `instances/{{ .Params.prefix }}-{{ .Loop.index }}` via `{{ .Params.count }}`) — the same map+`count` shape shown in the chosen option's example below, written against today's schema and path conventions rather than this decision's original ones. Run it with `nwsctl plan`/`apply` against a target declaring `params: { count: 2, prefix: dummy-vm }`. The resulting `state.json` should contain exactly one `store.provisioner` entry (`{"name": "vagrant"}`) and two distinctly-addressed `instances` entries (`dummy-vm-0`, `dummy-vm-1`), confirming that the params-driven `count` resolves to that many generated entries and that a plain operation and a generator operation execute together from one `stateChanges` block.

## Pros and Cons of the Options

### Array of operation objects (status quo)

The pre-existing shape: `stateChanges` was a plain list, each element a single operation object with no name of its own.

* Good, because it is the simpler of the two shapes — no key/value distinction, no implicit mode-switch on a field's presence.
* Bad, because an operation has no identity beyond its position in the array, unlike a named map key, which is harder to identify when reading the recipe source.
* Bad, because adding "generate N entries from one operation" on top of a plain list requires either a second, incompatible list type dedicated to generators, or overloading list elements with ad hoc repeat-count fields — either way, coexistence of single and generated entries is not natural in this shape.

### Map keyed by operation name, `count` as loop trigger (chosen)

`stateChanges` becomes a map; each value is an operation object which is treated as a loop-generator if and only if it has a `count` field.

```yaml
spec:
  defaults:
    params:
      count: 2
      prefix: "nws-vm"
  tasks:
    - name: update-state
      dependsOn: [extract-ssh-configs]
      stateChanges:
        # single operation: no 'count' key
        set-provisioner-info:
          op: set
          path: "/infrastructure/generators/{{ .Target }}/provisioner"
          value: "vagrant"
        # generator operation: 'count' key present
        generate-instances:
          count: "{{ .Params.count }}"
          op: set
          path: "/infrastructure/generators/{{ .Target }}/pool/{{ .Params.prefix }}-{{ .Loop.index }}"
          valueFromFile: "{{ .Outputs.instances }}/{{ .Params.prefix }}-{{ .Loop.index }}.json"
```

`path` is written as a full absolute path here, and the generator's instance collection is keyed `pool` — both the convention in effect at the time of this decision. This ADR concerns only the `stateChanges` container shape and the `count` loop trigger.

* Good, because single and generator operations share one schema and can be declared side by side, as shown above.
* Good, because `count` accepts a template expression (e.g. `{{ .Params.count }}`), so the number of generated entries is driven by recipe params, per the decision drivers.
* Good, because the operation name in a map key is meaningful when reading the recipe file, unlike a bare array index.
* Neutral, because whether an entry is a single operation or a generator is inferred implicitly from the presence of `count`, rather than declared through an explicit discriminator field; the original proposal did not consider an explicit-discriminator alternative, so this trade-off was not weighed against alternatives at decision time.
* Bad, because YAML map key ordering is not guaranteed by the schema itself; if multiple operations in one task ever needed to run in a specific order, the map shape alone does not express that (no ordering guarantee was part of the requirements this proposal addressed, so this was not a factor in the decision, but is a latent limitation of the chosen shape).

## More Information

This ADR reflects the schema and terminology as implemented at the time of this decision. Where this document and the current codebase or user guide disagree, treat `internal/` and `docs/user-guide/` as authoritative.

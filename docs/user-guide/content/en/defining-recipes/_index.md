---
title: "Defining Recipes"
weight: 5
---

# Defining Recipes

This section describes how to define your own recipes. It is intended for those who want to create or extend recipes.

If you are writing a recipe for the first time, start with [Tutorial: Your First Recipe]({{< relref "first-recipe" >}}). In about 15 minutes, it walks you through the entire flow from writing a recipe to applying it to a node.

---

## Recipes in Niwashi

A recipe is an executable module that packages provisioning and configuration procedures. While a State file declares **what** you want to achieve, a recipe defines **how** to achieve it.

{{< mermaid >}}
flowchart LR
    subgraph input["Input"]
        state["State<br/>(what to achieve)"]
        recipe["Recipes<br/>(how to achieve it)"]
    end
    plan["nwsctl plan<br/>diff calculation and<br/>execution planning"]
    apply["nwsctl apply<br/>task execution"]
    subgraph target["Targets"]
        host["Host"]
        infra["Infrastructure<br/>(VMs / instances)"]
        node["Nodes"]
        cluster["Clusters"]
    end
    state --> plan
    recipe --> plan
    plan --> apply
    apply --> host
    apply --> infra
    apply --> node
    apply --> cluster
{{< /mermaid >}}

`nwsctl plan` loads the State and recipes, calculates the diff, and generates an execution plan that orders the recipe tasks according to their dependencies. `nwsctl apply` executes that plan. For details on how this works, see [Architecture]({{< relref "/getting-started/architecture" >}}).

---

## Design Principles

### One Recipe, One Feature

Niwashi recipes follow the principle that **one recipe provides one feature**.

- Do not pack multiple features into a single recipe
- Combine features by listing multiple recipes under `capabilities`
- Keeping recipes small improves reusability and maintainability

---

## Which kind of Recipe to Write

A recipe declares its type with the top-level `kind` field. The `kind` you should write is determined by what you want to do.

| What you want to do | `kind` | Where it runs | Guide |
|---------------------|--------|---------------|-------|
| Install and configure software on a node | `node` | On the target node (via remote connection) | [Defining Node Capabilities]({{< relref "node-capability" >}}) |
| Build and configure a cluster spanning multiple nodes | `cluster` | On the cluster as a whole (decomposed per node when needed) | [Defining Cluster Capabilities]({{< relref "cluster-capability" >}}) |
| Create and destroy VMs or instances | `infra` | On the host running nwsctl | [Defining Infrastructure Provisioners]({{< relref "infrastructure-provisioning" >}}) |
| Make a tool on the host running nwsctl available to other recipes | `host` | On the host running nwsctl | [Defining Host Configuration]({{< relref "host-configuration" >}}) |
| Provide commands that other recipes can call (e.g. running Ansible) | `adapter` | Follows the calling recipe | [Defining Adapters]({{< relref "defining-adapters" >}}) |

---

## How Recipes Are Referenced from State

A recipe you create is referenced from a different place depending on its `kind`.

| `kind` | Referenced from | Example |
|--------|-----------------|---------|
| `node` | `inventory.nodes.<node>.capabilities` in State | `- my-org.nginx` |
| `cluster` | `inventory.clusters.<cluster>.capabilities` in State | `- cluster.kubernetes` |
| `infra` | `infrastructure.generators.<generator>.provisioner` in State | `provisioner: infra.vm.driver=vagrant` |
| `host` | Cannot be specified directly in State. Executed when another recipe declares it as a dependency in `spec.requires` | `requires: [host.tool.ansible]` |
| `adapter` | Cannot be specified directly in State. Declared as a dependency in another recipe's `spec.requires` and invoked via `toolRef` in a `tool.run` action | `toolRef: adapter.tool.ansible` |

For how recipe names are resolved (direct specification by `metadata.id`, aliases via `spec.provides`, and narrowing with `attrs`), see [Capabilities and Recipe Specification]({{< relref "/defining-desired-state/capabilities" >}}) and [Recipe Loading Specification]({{< relref "recipe-loading" >}}).

---

## Placing and Loading Recipes

Recipes are placed as `nws-recipe.yaml` (one recipe per file) or `nws-catalog.yaml` (bundling multiple recipes). The directory specified with `nwsctl plan --recipe-dir` is searched recursively, and these files are loaded as recipes when found. If the same combination of ID and version (FQID) appears more than once, an error is raised.

For details, see [Bundling Multiple Recipes]({{< relref "catalog" >}}) and [Recipe Loading Specification]({{< relref "recipe-loading" >}}).

---

## Contents of This Section

### Getting Started

| Page | Description |
|------|-------------|
| [Tutorial: Your First Recipe]({{< relref "first-recipe" >}}) | Walk through writing and applying a recipe end to end |

### Common Specifications

| Page | Description |
|------|-------------|
| [Recipe Format Reference]({{< relref "recipe-spec" >}}) | Reference for all `spec` fields with per-kind annotations |
| [Defining Tasks]({{< relref "defining-tasks" >}}) | Reference for task fields, action types, dependsOn, and operation |
| [Updating State with stateChanges]({{< relref "state-changes" >}}) | How to update State after task execution |
| [Task Filtering Condition (where)]({{< relref "task-where" >}}) | Filtering task execution targets with the `where` field |
| [Variables Available in Recipes]({{< relref "recipe-variables" >}}) | Reference for template variables and environment variables |

### Per-Kind Guides

| Page | Description |
|------|-------------|
| [Defining Node Capabilities]({{< relref "node-capability" >}}) | How to write recipes applied to nodes (`kind: node`) |
| [Defining Cluster Capabilities]({{< relref "cluster-capability" >}}) | How to write recipes applied to clusters (`kind: cluster`) |
| [Defining Infrastructure Provisioners]({{< relref "infrastructure-provisioning" >}}) | How to write infrastructure generation recipes (`kind: infra`) |
| [Defining Host Configuration]({{< relref "host-configuration" >}}) | How to write installer recipes that run on the host (`kind: host`) |
| [Defining Adapters]({{< relref "defining-adapters" >}}) | How to implement adapter recipes (`kind: adapter`) |
| [Using Adapters]({{< relref "using-adapters" >}}) | How to call adapters from a recipe |

### Other

| Page | Description |
|------|-------------|
| [Recipe Naming Guidelines]({{< relref "recipe-naming" >}}) | Naming conventions for recipe IDs, capability aliases, and commands |
| [Bundling Multiple Recipes]({{< relref "catalog" >}}) | How to use nws-catalog.yaml |
| [Recipe Loading Specification]({{< relref "recipe-loading" >}}) | Details on FQID, version management, and alias resolution |

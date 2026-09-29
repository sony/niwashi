---
title: "Architecture"
weight: 3
---

# Architecture

This section describes Niwashi's internal structure and operating principles.

---

## Overview

Niwashi calculates the difference between the **current state** and the **desired state**, and automatically generates an execution plan to close that gap. The execution plan is built as a DAG (Directed Acyclic Graph) that accounts for task dependencies.

---

## Basic Operating Principles

### 1. State Management

Niwashi manages two states:

- **Current State**: The current state of the infrastructure and Nodes managed by Niwashi
- **Desired State**: The target state defined by the user in YAML files

### 2. Difference Calculation

Niwashi compares the current state with the desired state and calculates differences for the following elements:

- **Node Capabilities**: Capabilities to add or remove for each Node, and existing Capabilities to update (their `params` or recipe version changed)
- **Cluster Capabilities**: Capabilities to add or remove for each Cluster, and existing Capabilities to update (their `params` or recipe version changed)
- **Generators**: Generators to provision or remove

### 3. Execution Plan Generation

An execution plan is generated based on the calculated differences and the Recipes. The execution plan is built as a DAG (Directed Acyclic Graph) that accounts for task dependencies.

---

## Difference Calculation in Detail

### Node Capabilities

For each Node, the Capabilities defined in the desired state are compared against the current state:

```yaml
# Desired state
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter

# Current state (example: only web.nginx exists)
# → monitoring.prometheus-exporter needs to be added
```

**Difference**:
- Capabilities to add: `monitoring.prometheus-exporter`
- Capabilities to remove: none

When a Capability exists in both states, its `params` and the resolved recipe version are also compared:

```yaml
# Current state: web.nginx applied with params {port: 80}
# Desired state: web.nginx with params {port: 8080}
# → web.nginx needs to be updated (its operation: update tasks run)
```

See [Update (operation: update)]({{< relref "defining-recipes/defining-tasks#update-operation-update" >}}) for details.

### Cluster Capabilities

For each Cluster, the Capabilities defined in the desired state are compared against the current state:

```yaml
# Desired state
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes

# Current state (example: cluster does not exist)
# → cluster.kubernetes needs to be added
```

**Difference**:
- Capabilities to add: `cluster.kubernetes`
- Node changes: none (cp, worker1, worker2 already exist)

### Generator (Instance Creation)

Generators are compared by name. A Generator that exists only in the desired state is provisioned, and its provisioner creates the instances:

```yaml
# Desired state
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3

# Current state (example: Generator vms does not exist)
# → vms needs to be provisioned (the provisioner creates 3 instances)
```

**Difference**:
- Generators to provision: `vms`

Changes to the `params` of an existing Generator (for example, `count`) are not detected.

---

## Building the DAG (Directed Acyclic Graph)

### What is a DAG

A DAG (Directed Acyclic Graph) is a graph structure representing the dependency relationships between tasks. Niwashi builds a DAG based on the differences and the Recipes.

### Task Dependencies

The execution order of tasks is determined by the following elements:

#### 1. Recipe Dependencies (spec.requires)

Each Recipe declares its dependencies on other Recipes via `spec.requires`. For example, the Kubernetes Recipe depends on the Ansible Recipe:

```yaml
spec:
  requires:
    - host.tool.ansible
    - host.tool.git
```

In this case, the Ansible Recipe must be executed first.

#### 2. Task Dependencies (dependsOn)

Each task within a Recipe declares its dependencies on other tasks via `dependsOn`:

```yaml
tasks:
  - name: gen-inventory
    dependsOn: [ensure-dirs]
  - name: run-kubespray
    dependsOn: [gen-inventory]
```

In this case, the execution order is `ensure-dirs` → `gen-inventory` → `run-kubespray`.

#### 3. Phase

Each Recipe has a defined phase in which it executes:

1. **host**: Host-level configuration (tool installation, etc.)
2. **infra**: Infrastructure provisioning (Generator execution)
3. **node**: Applying Capabilities at the Node level
4. **cluster**: Applying Capabilities at the Cluster level

Each phase is executed only after the previous phase has fully completed.

Note that while the execution target of the cluster phase is normally the cluster as a whole (a single target), tasks are decomposed and executed per node within the cluster when a task specifies a `where` condition or when an adapter declares `executionUnit: node`. For the details of the decomposition conditions, see [Defining Cluster Capabilities]({{< relref "/defining-recipes/cluster-capability" >}}).

These dependencies are used to build the DAG.

### Example DAG

```
[Phase: host]
  ↓
[Host configuration: tool installation]
  ↓
[Phase: infra]
  ↓
[Instance creation: vms]
  ↓
[Phase: node]
  ↓
[Node: web-server - web.nginx]
  ↓
[Node: web-server - monitoring.prometheus-exporter]
  ↓
[Phase: cluster]
  ↓
[Cluster: k8s-cluster - cluster.kubernetes]
```

---

## Executing the Plan

### Plan File

The generated DAG is saved as a JSON-format plan file (`plan.json`).

### Execution

The `nwsctl apply` command reads the plan file and executes tasks sequentially according to the DAG:

1. **Dependency resolution**: Verify that each task's dependent tasks have completed
2. **Task execution**: Call the Recipe to execute the task
3. **State update**: After a task completes, update the current state
4. **Next task**: Execute the next task according to the dependency graph

### Error Handling

If an error occurs during task execution, subsequent tasks that depend on that task will not be executed.

---

## Benefits of the Architecture

### Declarative Definition

Users only define "what" they want to achieve; Niwashi and the Recipes determine "how" to achieve it.

### Idempotency

Because Niwashi executes nothing when there are no differences, idempotency is ideally maintained. However, actual idempotency depends on the Recipe implementation.

This matters in particular when a task fails: a failed construct is not recorded in State, so the next run executes it again from a partially completed state (see [When a Task Fails]({{< relref "defining-recipes/defining-tasks#when-a-task-fails" >}})).

The reference Recipes provided by Niwashi (Vagrant, Ansible, Kubernetes, etc.) have been confirmed to be idempotent. When creating custom Recipes, implementations that consider idempotency are recommended.

### Flexibility

The Recipe system makes it easy to add new Provisioners and configuration management tools.

---

## Next Steps

- [Defining State]({{< relref "defining-desired-state" >}}) - How to write State files
- [Defining Recipes]({{< relref "defining-recipes" >}}) - How to create your own Recipes

---
title: "Defining Cluster Capabilities"
weight: 6
---

# Defining Cluster Capabilities

This page explains how to write Capability recipes with `kind: cluster` that are applied to clusters.

---

## What Is a `kind: cluster` Recipe?

A `kind: cluster` recipe **executes operations against an entire cluster**. When a user specifies a Capability on a Cluster in a State file, Niwashi finds the matching recipe and executes its tasks against the target cluster.

{{< mermaid >}}
flowchart LR
    subgraph host["Host (nwsctl)"]
        recipe["kind: cluster recipe"]
    end
    subgraph cluster["Cluster (a single target)"]
        cp["cp"]
        w1["worker-1"]
        w2["worker-2"]
    end
    recipe -->|"executed against the whole cluster"| cluster
    recipe -.->|"decomposed per node with<br/>where / executionUnit: node"| cp
{{< /mermaid >}}

There are two implementation patterns for cluster recipes.

| Pattern | Approach | Best suited for |
|---------|----------|-----------------|
| **Adapter pattern** | Delegate processing to an adapter (Ansible, etc.) via `tool.run` | Processing the whole cluster with a single tool (Kubespray, etc.) |
| **`where` pattern** | Decompose per-node using `where` | Describing per-role logic with niwashi alone, without an adapter |

---

## Field Reference

### Common Fields

See [Recipe Format Reference]({{< relref "recipe-spec" >}}) for details on common `spec` fields. For `kind: cluster`, `spec.requires` can declare dependencies on `host`, `node`, and `cluster` recipes.

See [Defining Tasks]({{< relref "defining-tasks" >}}) for how to write tasks, and [Updating State with stateChanges]({{< relref "state-changes" >}}) for updating State after a task runs.

### Template Variables

In `kind: cluster` tasks, `{{ .Target }}` expands to the **ID of the target cluster**. Even when tasks are decomposed per-node via `where`, `{{ .Target }}` remains the cluster ID. To identify the node being executed against in per-node tasks, use the `$NWS_TARGET_ID` environment variable in your script.

For the full list of template variables and runtime environment variables, see [Variables Available in Recipes]({{< relref "recipe-variables" >}}).

### `where` Specification

When `where` is used, Niwashi decomposes the cluster into individual nodes and executes tasks against each matching node.

#### Without an Adapter

Without an adapter, decomposition depends only on whether `where` is present:

| `where` | Behavior |
|---------|----------|
| absent | Execute once against the whole cluster (one unit), with no per-node decomposition |
| present | Decompose into nodes and execute only on those matching the condition |

There is no way to omit `where` and still get per-node decomposition without an
adapter — `executionUnit` is a field on the adapter's own recipe, not something a
plain `where`-pattern task can set. To run a task on **every** node without an
adapter, make the intent explicit with `where: "true"` rather than omitting `where`:

```yaml
- name: prepare
  where: "true"
  action:
    exec.remote:
      scriptTpl: |
        #!/bin/bash
        apt-get install -y curl
```

This also matters because `exec.remote` requires a node-level target: a task with
no `where` runs as a single whole-cluster unit and has no node to connect to, so
`exec.remote` fails for it. `where: "true"` decomposes into all nodes, giving each
one a node-level execution context.

#### With an Adapter (`executionUnit`)

When a task uses an adapter, behavior also depends on the adapter's own
`executionUnit`:

| `where` | adapter's `executionUnit` | Behavior |
|---------|---------------------------|----------|
| absent | `default` | Execute once against the whole cluster (one unit); the adapter manages distribution |
| absent | `node` | Decompose into all nodes in the cluster and execute |
| present | `node` | Decompose and execute only on matching nodes |
| present | `default` | **Error** (the adapter does not support node-level execution) |

To filter nodes with `where`, define labels on `nodes` in the State file.

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        cp:
          labels:
            role: master
        worker1:
          labels:
            role: worker
```

Cluster node labels are only active when executing that cluster's recipes. Even if the same node belongs to multiple clusters, labels are evaluated independently per cluster.

See [Task Filtering Condition (where)]({{< relref "task-where" >}}) for the full variable list and CEL expression syntax.

---

## Minimal Example

```yaml
version: nws.recipe/v1
kind: cluster
metadata:
  id: my-org/my-cluster-feature
  version: "1.0.0"
  description: "My cluster feature recipe"
spec:
  provides:
    - name: my-org.my-cluster-feature
  tasks:
    - name: setup
      where: "true"
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            echo "Setting up node $NWS_TARGET_ID in cluster {{ .Target }}"
```

---

## Pattern 1: Cluster Setup with an Adapter

This pattern delegates all processing to an adapter (Ansible, etc.) via `tool.run`. The task's execution target is treated as "the entire cluster (one unit)". Distribution to nodes is managed by the adapter (e.g., via an Ansible Inventory file).

```yaml
spec:
  requires:
    - adapter.tool.ansible

  workspace:
    mode: persistent

  tasks:
    - name: gen-inventory
      action:
        tool.run:
          toolRef: adapter.tool.ansible
          command: gen-inventory
          workdirTpl: "{{ .Paths.work_dir }}"

    - name: run-playbook
      dependsOn: [gen-inventory]
      action:
        tool.run:
          toolRef: adapter.tool.ansible
          command: ansible-playbook
          workdirTpl: "{{ .Paths.work_dir }}"
          argvTpl:
            - -i
            - hosts.ini
            - playbook.yml
```

See [Ansible Adapter]({{< relref "../recipes/ansible" >}}) for details.

---

## Pattern 2: Cluster Setup with `where`

This pattern dispatches tasks based on node roles (labels) or attributes, allowing cluster configuration to be described with niwashi alone, without an adapter.

```yaml
spec:
  tasks:
    # Common processing for all nodes
    - name: prepare
      where: "true"
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            apt-get install -y curl

    # Only nodes with master role
    - name: init-master
      where: "node.labels.role == 'master'"
      dependsOn: [prepare]
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            kubeadm init

    # Only nodes with worker role
    - name: join-workers
      where: "node.labels.role == 'worker'"
      dependsOn: [init-master]
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            bash "{{ .Outputs }}/join_command.sh"
```

Compound conditions using CEL expressions are also supported.

```yaml
# Linux AND master role
- name: init-linux-master
  where: "node.os == 'linux' && node.labels.role == 'master'"
```

---

## Example: Kubernetes Cluster with master/worker Roles

A complete example using the `where` pattern to configure control plane and worker nodes.

```yaml
version: nws.recipe/v1
kind: cluster
metadata:
  id: my-org/my-k8s
  version: "1.0.0"
  description: "Kubernetes cluster setup"
spec:
  provides:
    - name: my-org.my-k8s

  tasks:
    # Common preprocessing for all nodes
    - name: prepare
      where: "true"
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            sudo apt-get update
            sudo apt-get install -y curl

    # Only master role nodes
    - name: init-master
      where: "node.labels.role == 'master'"
      dependsOn: [prepare]
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            kubeadm init --pod-network-cidr=10.244.0.0/16
            mkdir -p $HOME/.kube
            sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config
            kubeadm token create --print-join-command \
              > "{{ .Outputs }}/join_command.sh"
      stateChanges:
        save-join-command:
          op: set
          path: "store/join_command"
          valueFromFile: "{{ .Outputs }}/join_command.sh"

    # Only worker role nodes
    - name: join-workers
      where: "node.labels.role == 'worker'"
      dependsOn: [init-master]
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            bash "{{ .Outputs }}/join_command.sh"

    # Cluster-level postprocessing
    - name: verify
      dependsOn: [init-master, join-workers]
      action:
        exec.local:
          scriptTpl: |
            echo "Cluster {{ .Target }} setup completed"
```

---

## Next Steps

- [Task Filtering Condition (where)]({{< relref "task-where" >}}) — Details on the `where` field
- [Defining Tasks]({{< relref "defining-tasks" >}}) — Action types, dependsOn, and operation details
- [Using Adapters]({{< relref "using-adapters" >}}) — Details on calling adapters with `tool.run`
- [Ansible Adapter]({{< relref "../recipes/ansible" >}}) — How to use the reference Ansible adapter

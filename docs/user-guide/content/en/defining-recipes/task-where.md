---
title: "Task Filtering Condition (where)"
weight: 8
---

# Task Filtering Condition (`where`)

Specifying the optional `where` field on a task causes the task to execute only against targets that match the condition. The value is a [CEL (Common Expression Language)](https://github.com/google/cel-spec) expression string; the task runs only when the expression evaluates to `true`. When omitted, the task applies unconditionally to the target.

---

## Behavior by `kind`

The evaluation context of `where` differs by `kind`.

| `kind` | Execution target | `where` evaluation context |
|--------|-----------------|---------------------------|
| `node` | The specified node (one) | Target node's `os`, `arch`, `labels` |
| `infra` | The specified infra target (one) | Host running nwsctl: `os`, `arch` |
| `host` | The host running nwsctl (fixed, one) | Host itself: `os`, `arch` |

For `infra` and `host`, `where` acts more as an execution eligibility check ("should this task run in this host environment?") than as a target filter.

---

## Evaluation Context Variables

| Variable | Type | Available in `kind` | Description |
|----------|------|---------------------|-------------|
| `node.os` | string | `node`, `infra`, `host` | OS identifier |
| `node.arch` | string | `node`, `infra`, `host` | CPU architecture |
| `node.labels` | map(string, string) | `node`, `cluster` | User-defined labels |

### Values of node.os / node.arch

`node.os` and `node.arch` are set either by a `kind: infra` recipe writing them via `stateChanges`, or by nwsctl probing the target node and detecting the actual values. Writing `os` or `arch` keys in State's `labels` does **not** populate these variables.

The string notation follows [Go - Optional environment variables](https://golang.org/doc/install/source#environment) (e.g. `linux`, `windows`, `darwin`, `amd64`, `arm64`).

### Values of node.labels

`node.labels` is the merged result of the following:

- Global labels defined in `inventory.nodes.<node-name>.labels`
- When executing a `kind: cluster` recipe, the cluster-specific labels from `inventory.clusters.<cluster-name>.nodes.<node-name>.labels` are additionally merged in (cluster-specific labels take precedence)

```yaml
inventory:
  nodes:
    cp:
      labels:
        tier: control      # global label → node.labels.tier = "control"

  clusters:
    k8s-cluster:
      nodes:
        cp:
          labels:
            role: master   # cluster-specific label → node.labels.role = "master"
```

When executing `k8s-cluster` recipes, `node.labels` becomes `{tier: control, role: master}`. Cluster-specific labels are only active during that cluster's recipe execution and do not affect other clusters.

---

## Examples

### Branching by OS

```yaml
tasks:
  - name: install-linux
    where: "node.os == 'linux'"
    action:
      exec.remote:
        scriptTpl: |
          sudo apt-get install -y nginx

  - name: install-darwin
    where: "node.os == 'darwin'"
    action:
      exec.local:
        scriptTpl: |
          brew install nginx
```

### Filtering by label (`kind: node`)

```yaml
tasks:
  - name: configure-master
    where: "node.labels.role == 'master'"
    action:
      exec.remote:
        scriptTpl: |
          echo "Configuring master node"
```

### Compound condition

```yaml
tasks:
  - name: configure-worker-linux
    where: "node.os == 'linux' && node.labels.role == 'worker'"
    action:
      exec.remote:
        scriptTpl: |
          echo "Configuring linux worker"
```

### `in` operator

```yaml
tasks:
  - name: configure-control-plane
    where: "node.labels.role in ['master', 'etcd']"
    action:
      exec.remote:
        scriptTpl: |
          echo "Configuring control plane node"
```

---

## Related Pages

- [Defining Node Capabilities]({{< relref "node-capability" >}}) — Recipes with `kind: node`
- [Defining Infrastructure Provisioners]({{< relref "infrastructure-provisioning" >}}) — Recipes with `kind: infra`
- [Defining Host Configuration]({{< relref "host-configuration" >}}) — Recipes with `kind: host`

---
title: "State Merging in Detail"
weight: 2
---

# State Merging in Detail

This page explains the detailed behavior when merging multiple State files.

## Merge Basics

When multiple `-t` options are specified with the `nwsctl plan` command, State files are merged in the order specified.

```bash
nwsctl plan -t file1.yaml -t file2.yaml -t file3.yaml
```

Merge order:
1. Load `file1.yaml`
2. Load `file2.yaml` and merge with `file1.yaml`
3. Load `file3.yaml` and merge with the previous merge result

**Important**: Files specified later override the settings of earlier files.

---

## Merging Top-Level Elements

### version

The `version` field must match across all files. An error will occur if different versions are specified.

```yaml
# file1.yaml
version: nws.state/v1

# file2.yaml
version: nws.state/v1  # OK: same version
```

### metadata

For `metadata`, later files override earlier files.

```yaml
# file1.yaml
metadata:
  project: my-project

# file2.yaml
metadata:
  project: my-project-prod  # Overrides file1's setting

# Merge result
metadata:
  project: my-project-prod
```

### inventory

The `inventory` section is merged per Node and Cluster.

### infrastructure

The `infrastructure` section is merged per Generator.

### template

The `template` section is merged per template.

---

## Merging inventory

### nodes

Nodes are merged per identifier (key). When Nodes with the same identifier exist, their attributes are merged.

#### Example 1: Adding Nodes

```yaml
# file1.yaml
inventory:
  nodes:
    node1: {}

# file2.yaml
inventory:
  nodes:
    node2: {}

# Merge result
inventory:
  nodes:
    node1: {}
    node2: {}
```

#### Example 2: Adding Node Attributes

```yaml
# file1.yaml
inventory:
  nodes:
    web-server: {}

# file2.yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx

# Merge result
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
```

#### Example 3: Merging labels

labels are merged per key. When the same key exists, the later file overrides the earlier one.

```yaml
# file1.yaml
inventory:
  nodes:
    web-server:
      labels:
        env: dev
        region: us-east

# file2.yaml
inventory:
  nodes:
    web-server:
      labels:
        env: prod        # Override
        tier: frontend   # Add

# Merge result
inventory:
  nodes:
    web-server:
      labels:
        env: prod        # Overridden by file2
        region: us-east  # From file1
        tier: frontend   # Added by file2
```

#### Example 4: Merging capabilities

capabilities are concatenated as an array.

```yaml
# file1.yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx

# file2.yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - monitoring.prometheus-exporter

# Merge result
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter
```

#### Example 5: Overwriting instanceSelector

`instanceSelector` is completely overwritten by the later file.

```yaml
# file1.yaml
inventory:
  nodes:
    web-server:
      instanceSelector:
        generator: gen1

# file2.yaml
inventory:
  nodes:
    web-server:
      instanceSelector:
        generator: gen2  # Complete overwrite

# Merge result
inventory:
  nodes:
    web-server:
      instanceSelector:
        generator: gen2
```

### clusters

Clusters are also merged per identifier, the same as Nodes.

#### Example 1: Adding Clusters

```yaml
# file1.yaml
inventory:
  clusters:
    cluster1:
      nodes: [node1, node2]

# file2.yaml
inventory:
  clusters:
    cluster2:
      nodes: [node3, node4]

# Merge result
inventory:
  clusters:
    cluster1:
      nodes: [node1, node2]
    cluster2:
      nodes: [node3, node4]
```

#### Example 2: Overwriting nodes

The `nodes` of a Cluster is completely overwritten by the later file.

```yaml
# file1.yaml
inventory:
  clusters:
    my-cluster:
      nodes: [node1, node2]

# file2.yaml
inventory:
  clusters:
    my-cluster:
      nodes: [node1, node2, node3]  # Complete overwrite

# Merge result
inventory:
  clusters:
    my-cluster:
      nodes: [node1, node2, node3]
```

#### Example 3: Concatenating capabilities

```yaml
# file1.yaml
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes

# file2.yaml
inventory:
  clusters:
    k8s-cluster:
      capabilities:
        - monitoring.cluster-monitoring

# Merge result
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes
        - monitoring.cluster-monitoring
```

---

## Merging infrastructure

### generators

Generators are merged per identifier.

#### Example 1: Adding Generators

```yaml
# file1.yaml
infrastructure:
  generators:
    gen1:
      provisioner: external-instance

# file2.yaml
infrastructure:
  generators:
    gen2:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3

# Merge result
infrastructure:
  generators:
    gen1:
      provisioner: external-instance
    gen2:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
```

#### Example 2: Merging params

`params` is merged per key.

```yaml
# file1.yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        box: ubuntu/jammy64
        cpus: 2

# file2.yaml
infrastructure:
  generators:
    vms:
      params:
        count: 3
        memory: 2048

# Merge result
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        box: ubuntu/jammy64
        cpus: 2
        count: 3
        memory: 2048
```

#### Example 3: Overwriting params

When the same key exists, the later file overrides the earlier one.

```yaml
# file1.yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        memory: 1024

# file2.yaml
infrastructure:
  generators:
    vms:
      params:
        memory: 4096  # Override

# Merge result
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        memory: 4096  # Overridden by file2
```

---

## Merging template

Templates are merged per identifier.

### Example 1: Adding Templates

```yaml
# file1.yaml
template:
  node:
    base:
      labels:
        managed_by: niwashi

# file2.yaml
template:
  node:
    production:
      labels:
        environment: production

# Merge result
template:
  node:
    base:
      labels:
        managed_by: niwashi
    production:
      labels:
        environment: production
```

### Example 2: Overwriting Templates

```yaml
# file1.yaml
template:
  node:
    base:
      labels:
        env: dev

# file2.yaml
template:
  node:
    base:
      labels:
        env: prod  # Override

# Merge result
template:
  node:
    base:
      labels:
        env: prod
```

---

## Summary of Merge Behavior

| Element | Merge Behavior |
|---------|----------------|
| **Top-level** | |
| version | Must match |
| metadata | Later file overrides |
| **inventory.nodes** | |
| Node itself | Merged per identifier |
| labels | Merged per key (same key is overridden) |
| capabilities | Concatenated as array |
| instanceSelector | Later file completely overrides |
| templates | Later file completely overrides |
| **inventory.clusters** | |
| Cluster itself | Merged per identifier |
| nodes | Later file completely overrides |
| labels | Merged per key (same key is overridden) |
| capabilities | Concatenated as array |
| params | Merged per key (same key is overridden) |
| templates | Later file completely overrides |
| **infrastructure.generators** | |
| Generator itself | Merged per identifier |
| provisioner | Later file overrides |
| params | Merged per key (same key is overridden) |
| templates | Later file completely overrides |
| **template** | |
| Template itself | Merged per identifier |
| Attributes | Same rules as Node/Cluster/instance |

---

## Practical Examples

### Example 1: Separating Logical Configuration and Infrastructure

```yaml
# app.yaml (logical configuration)
version: nws.state/v1

inventory:
  nodes:
    web: {}
    db: {}

# infra-dev.yaml (infrastructure)
version: nws.state/v1

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
```

Merge command:
```bash
nwsctl plan -t app.yaml -t infra-dev.yaml
```

Merge result:
```yaml
version: nws.state/v1

inventory:
  nodes:
    web: {}
    db: {}

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
```

### Example 2: Incremental Definition

```yaml
# 01-base.yaml (basic structure)
version: nws.state/v1

inventory:
  nodes:
    web: {}
    db: {}

# 02-capabilities.yaml (function definition)
version: nws.state/v1

inventory:
  nodes:
    web:
      capabilities:
        - web.nginx
    db:
      capabilities:
        - database.postgresql

# 03-infra.yaml (infrastructure)
version: nws.state/v1

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
```

Merge command:
```bash
nwsctl plan -t 01-base.yaml -t 02-capabilities.yaml -t 03-infra.yaml
```

Merge result:
```yaml
version: nws.state/v1

inventory:
  nodes:
    web:
      capabilities:
        - web.nginx
    db:
      capabilities:
        - database.postgresql

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
```

---

## Notes

### 1. Merge Order

The merge order is critically important. Files specified later override settings from earlier files.

```bash
# Correct order
nwsctl plan -t base.yaml -t override.yaml

# Wrong order (override settings get overridden by base)
nwsctl plan -t override.yaml -t base.yaml
```

### 2. Array Override vs Concatenation

- **capabilities**: Concatenated
- **nodes (in Clusters)**: Overridden
- **templates**: Overridden

Understanding this difference is important.

### 3. Unintentional Overrides

Using the same identifier can unintentionally override settings.

```yaml
# file1.yaml
inventory:
  nodes:
    web-server:
      labels:
        env: dev

# file2.yaml
inventory:
  nodes:
    web-server:  # Same identifier
      labels:
        region: us-east  # Will be merged
```

Make sure identifiers are consistent as intended.

---

## Next Steps

- [Managing Multiple Environments]({{< relref "multi-environment" >}}) - Practical environment management using merge

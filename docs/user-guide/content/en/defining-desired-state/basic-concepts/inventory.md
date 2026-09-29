---
title: "Inventory Basics"
weight: 2
---

# Inventory Basics

The `inventory` section defines the logical configuration of the system. It is where you describe "what kind of system you want to build".

## Inventory Structure

```yaml
inventory:
  nodes:
    # Definition of individual logical Nodes
  clusters:
    # Definition of Clusters (optional)
```

- **nodes** (required): Definitions of individual logical Nodes
- **clusters** (optional): Definitions of Clusters composed of multiple Nodes

---

## Nodes

A Node is an individual logical unit of computation that makes up the system.

### Basic Definition

The simplest Node definition:

```yaml
inventory:
  nodes:
    node1: {}
    node2: {}
    node3: {}
```

This example defines three empty Nodes. Only identifiers (`node1`, `node2`, `node3`) are specified; attributes are empty.

### Node Identifiers

Each Node is given a unique identifier (key). Identifiers are used when referencing Nodes within the State file.

```yaml
inventory:
  nodes:
    web-server: {}
    db-primary: {}
    cache-server: {}
```

For identifier naming rules, see [Identifier Naming Rules]({{< relref "naming-rules" >}}).

### Attributes Available for Nodes

The following attributes can be specified for Nodes:

#### labels

Attaches labels (metadata) to a Node:

```yaml
inventory:
  nodes:
    web-server:
      labels:
        role: web
        environment: production
        region: us-east-1
```

#### capabilities

Specifies the Capabilities (functions) the Node should have:

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter
```

A Capability is a unit of functionality realized by a Recipe. See [Capabilities and Recipes in Detail]({{< relref "../capabilities" >}}) for more information.

#### instanceSelector

Controls which instance is assigned to the Node:

```yaml
inventory:
  nodes:
    web-server:
      instanceSelector:
        generator: web-vms
```

See [Nodes in Detail]({{< relref "../nodes" >}}) for more information.

#### templates

Inherits from templates:

```yaml
template:
  node:
    base:
      labels:
        environment: production

inventory:
  nodes:
    web-server:
      templates: [base]
```

See [Templates in Detail]({{< relref "../templates" >}}) for more information.

### Node Definition Examples

```yaml
inventory:
  nodes:
    # Simple Node
    worker-01: {}

    # Node with labels
    web-server:
      labels:
        role: web
        tier: frontend

    # Node with Capabilities
    db-server:
      capabilities:
        - database.postgresql

    # Node with detailed settings
    app-server:
      labels:
        role: application
        tier: backend
      capabilities:
        - id: runtime.python
          params:
            version: "3.11"
      instanceSelector:
        generator: app-vms
```

---

## Clusters

A Cluster is a group composed of multiple Nodes. By defining Cluster-level Capabilities, processing that spans multiple Nodes (such as building a Kubernetes Cluster) can be realized.

### Basic Definition

A Cluster requires **at least one Node**:

```yaml
inventory:
  nodes:
    node1: {}
    node2: {}
    node3: {}

  clusters:
    my-cluster:
      nodes:
        - node1
        - node2
        - node3
```

### Cluster Identifiers

Like Nodes, each Cluster is given a unique identifier:

```yaml
inventory:
  clusters:
    web-cluster: {}
    db-cluster: {}
    k8s-cluster: {}
```

### Attributes Available for Clusters

The following attributes can be specified for Clusters:

#### nodes (required)

Specifies the list of Nodes participating in the Cluster as an array:

```yaml
inventory:
  clusters:
    web-cluster:
      nodes:
        - web-01
        - web-02
        - web-03
```

Node names must match identifiers defined in `inventory.nodes`.

#### capabilities

Specifies Cluster-level Capabilities:

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        - control-plane
        - worker-01
        - worker-02
      capabilities:
        - cluster.kubernetes
```

#### params

Specifies parameters for the entire Cluster:

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        - control-plane
        - worker-01
        - worker-02
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [control-plane]
          kube_node: [worker-01, worker-02]
          etcd: [control-plane]
```

In this example, when building the Kubernetes Cluster, the role of each Node is defined as a group.

#### templates

Inherits from templates:

```yaml
template:
  cluster:
    base-cluster:
      labels:
        managed_by: niwashi

inventory:
  clusters:
    prod-cluster:
      templates: [base-cluster]
      nodes:
        - node1
        - node2
```

### Cluster Definition Examples

```yaml
inventory:
  nodes:
    control-plane: {}
    worker-01: {}
    worker-02: {}
    db-primary: {}
    db-replica: {}

  clusters:
    # Kubernetes Cluster
    k8s-cluster:
      labels:
        environment: production
        platform: kubernetes
      nodes:
        - control-plane
        - worker-01
        - worker-02
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [control-plane]
          kube_node: [worker-01, worker-02]
          etcd: [control-plane]

    # Database Cluster
    db-cluster:
      labels:
        environment: production
        database: postgresql
      nodes:
        - db-primary
        - db-replica
      capabilities:
        - id: database.postgresql-replication
          params:
            replication_mode: streaming
```

---

## Relationship Between Nodes and Clusters

### Standalone Node

A Node can be used standalone without belonging to any Cluster:

```yaml
inventory:
  nodes:
    standalone-server:
      capabilities:
        - web.nginx
```

### Cluster Member Node

By making a Node belong to a Cluster, processing that spans multiple Nodes can be realized:

```yaml
inventory:
  nodes:
    node1: {}
    node2: {}

  clusters:
    my-cluster:
      nodes:
        - node1
        - node2
      capabilities:
        - cluster-level-feature
```

### Belonging to Multiple Clusters

A single Node can belong to multiple Clusters:

```yaml
inventory:
  nodes:
    node1: {}
    node2: {}
    node3: {}

  clusters:
    cluster-a:
      nodes:
        - node1
        - node2

    cluster-b:
      nodes:
        - node2
        - node3
```

In this example, `node2` belongs to both `cluster-a` and `cluster-b`.

---

## Practical Examples

### Web Application

```yaml
inventory:
  nodes:
    web-01:
      capabilities:
        - web.nginx
    web-02:
      capabilities:
        - web.nginx
    db-01:
      capabilities:
        - database.postgresql

  clusters:
    web-tier:
      nodes:
        - web-01
        - web-02
```

### Kubernetes Cluster

```yaml
inventory:
  nodes:
    cp:
      instanceSelector:
        generator: vms
    worker1:
      instanceSelector:
        generator: vms
    worker2:
      instanceSelector:
        generator: vms

  clusters:
    k8s-cluster:
      nodes:
        - cp
        - worker1
        - worker2
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [cp]
          kube_node: [worker1, worker2]
          etcd: [cp]
```

---

## Next Steps

- [Infrastructure Basics]({{< relref "infrastructure" >}}) - How to generate instances to assign to Nodes
- [Identifier Naming Rules]({{< relref "naming-rules" >}}) - Naming rules for Nodes and Clusters
- [Nodes in Detail]({{< relref "../nodes" >}}) - All Node attributes and options
- [Clusters in Detail]({{< relref "../clusters" >}}) - All Cluster attributes and options

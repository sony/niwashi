---
title: "Clusters in Detail"
weight: 7
---

# Clusters in Detail

This page provides a detailed explanation of all Cluster attributes and how to use them.

## What Is a Cluster?

A Cluster is a group composed of multiple Nodes. By defining Cluster-level Capabilities, you can realize processing that spans multiple Nodes (such as building a Kubernetes Cluster).

---

## Basic Cluster Definition

A Cluster requires **at least one Node**.

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

---

## Cluster Attributes

The following attributes can be specified for a Cluster.

### nodes (required)

Specifies the Nodes participating in the Cluster. At least one Node is required. Can be written in two forms: **string list form** and **object form**.

#### String List Form

```yaml
inventory:
  clusters:
    web-cluster:
      nodes:
        - web-01
        - web-02
```

#### Object Form (with labels)

Specify the node name as a key and define per-node roles within the cluster using `labels`. Used for task filtering with the `where` field.

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
        worker2:
          labels:
            role: worker
```

Nodes that don't need `labels` can be omitted with `{}`.

```yaml
nodes:
  cp:
    labels:
      role: master
  worker1: {}   # no labels
```

#### Relationship Between Cluster Node Labels and Global Node Labels

Nodes can have global labels defined in `inventory.nodes`. Cluster node labels are only active when executing that cluster's recipes and are merged with the global labels (cluster node labels take precedence). These labels are accessible as `node.labels` in recipe `where` fields.

```yaml
inventory:
  nodes:
    cp:
      labels:
        tier: control      # global label

  clusters:
    k8s-cluster:
      nodes:
        cp:
          labels:
            role: master   # label within k8s-cluster (merged with tier: control)

    db-cluster:
      nodes:
        cp:
          labels:
            role: replica  # label within db-cluster (independent from k8s labels)
```

When executing `k8s-cluster` recipes, `cp`'s `node.labels` is evaluated as `{tier: control, role: master}`. See [Task Filtering Condition (where)]({{< relref "../defining-recipes/task-where" >}}) for details.

#### Notes

- If a Node does not exist in `inventory.nodes`, an error will occur
- An empty array or empty map cannot be specified (at least one Node is required)
- The same Node cannot be specified more than once

---

### capabilities

Specifies Cluster-level Capabilities. This is used when defining processing that spans multiple Nodes.

#### Short form (string only)

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        - cp
        - worker1
        - worker2
      capabilities:
        - cluster.kubernetes
```

#### Long form (with parameters)

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        - cp
        - worker1
        - worker2
      capabilities:
        - id: cluster.kubernetes
          params:
            version: "1.28"
            network_plugin: calico
```

#### Use Cases for Cluster Capabilities

Cluster Capabilities are used for multi-Node processing such as:

- **Building Kubernetes Clusters**: Configuration of control-plane and worker Nodes
- **Database replication**: Configuration of primary/replica
- **Load balancer configuration**: Load balancing across multiple web servers
- **Distributed storage**: Storage Clusters across multiple Nodes

For Capability details, see [Capabilities and Recipes in Detail]({{< relref "capabilities" >}}).

---

### params

Specifies parameters for the entire Cluster. This is used when defining parameters for Cluster Capabilities.

```yaml
inventory:
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

#### Difference Between capabilities params and Cluster params

- **capabilities params**: Parameters passed to individual Capabilities
- **Cluster params**: Parameters for the entire Cluster

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]

      # Parameters passed to Capability
      capabilities:
        - id: cluster.kubernetes
          params:
            version: "1.28"
            network_plugin: calico

      # Parameters for the entire Cluster
      params:
        groups:
          kube_control_plane: [cp]
          kube_node: [worker1, worker2]
```

Which to use depends on the Recipe's specification. Refer to the Recipe documentation.

---

### templates

Clusters can be defined by inheriting from templates.

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

#### Applying Multiple Templates

```yaml
template:
  cluster:
    base:
      labels:
        managed_by: niwashi

    production:
      labels:
        environment: production

inventory:
  clusters:
    prod-cluster:
      templates: [base, production]
      nodes:
        - node1
        - node2
```

See [Templates in Detail]({{< relref "templates" >}}) for more information.

---

## Practical Cluster Definition Examples

### Kubernetes Cluster

When using the Ansible adapter (string list form + `params.groups` to specify node roles):

```yaml
inventory:
  nodes:
    control-plane: {}
    worker-01: {}
    worker-02: {}

  clusters:
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
```

When branching processing by role using `where` in the recipe (object form defining node labels):

```yaml
inventory:
  nodes:
    cp: {}
    worker-01: {}
    worker-02: {}

  clusters:
    k8s-cluster:
      nodes:
        cp:
          labels:
            role: master
        worker-01:
          labels:
            role: worker
        worker-02:
          labels:
            role: worker
      capabilities:
        - my-org.my-k8s
```

### Database Replication

```yaml
inventory:
  nodes:
    db-primary: {}
    db-replica-01: {}
    db-replica-02: {}

  clusters:
    db-cluster:
      labels:
        environment: production
        database: postgresql
      nodes:
        - db-primary
        - db-replica-01
        - db-replica-02
      capabilities:
        - id: database.postgresql-replication
          params:
            replication_mode: streaming
            primary_node: db-primary
```

### Web Server Cluster

```yaml
inventory:
  nodes:
    web-01: {}
    web-02: {}
    web-03: {}
    lb: {}

  clusters:
    web-tier:
      labels:
        tier: frontend
      nodes:
        - web-01
        - web-02
        - web-03
      capabilities:
        - web.load-balancer
      params:
        backend_nodes:
          - web-01
          - web-02
          - web-03
        load_balancer_node: lb
```

### Multiple Capabilities Example

```yaml
inventory:
  clusters:
    app-cluster:
      nodes:
        - app-01
        - app-02
      capabilities:
        - cluster.monitoring
        - cluster.logging
        - cluster.backup
      params:
        monitoring_endpoint: "http://prometheus:9090"
        logging_endpoint: "http://elasticsearch:9200"
```

---

## Relationship Between Nodes and Clusters

### Standalone Node vs Cluster Member Node

#### Standalone Node

A Node can be used standalone without belonging to any Cluster:

```yaml
inventory:
  nodes:
    standalone-server:
      capabilities:
        - web.nginx
```

This Node does not belong to any Cluster.

#### Cluster Member Node

By making a Node belong to a Cluster, processing that spans multiple Nodes can be realized:

```yaml
inventory:
  nodes:
    node1:
      capabilities:
        - web.nginx  # Node-level Capability

    node2:
      capabilities:
        - web.nginx

  clusters:
    web-cluster:
      nodes:
        - node1
        - node2
      capabilities:
        - cluster.load-balancer  # Cluster-level Capability
```

In this example:
- Each Node individually has `web.nginx`
- `cluster.load-balancer` is configured at the Cluster level to group multiple Nodes

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
      capabilities:
        - feature-a

    cluster-b:
      nodes:
        - node2
        - node3
      capabilities:
        - feature-b
```

In this example, `node2` belongs to both `cluster-a` and `cluster-b`.

---

## Common Patterns

### Pattern 1: Kubernetes Cluster (Ansible adapter)

```yaml
inventory:
  nodes:
    cp: {}
    worker1: {}
    worker2: {}

  clusters:
    k8s:
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [cp]
          kube_node: [worker1, worker2]
          etcd: [cp]
```

### Pattern 1b: Kubernetes Cluster (`where`-based node role branching)

```yaml
inventory:
  nodes:
    cp: {}
    worker1: {}
    worker2: {}

  clusters:
    k8s:
      nodes:
        cp:
          labels:
            role: master
        worker1:
          labels:
            role: worker
        worker2:
          labels:
            role: worker
      capabilities:
        - my-org.my-k8s
```

### Pattern 2: Database Cluster

```yaml
inventory:
  nodes:
    db-primary:
      capabilities:
        - database.postgresql
    db-replica:
      capabilities:
        - database.postgresql

  clusters:
    db-cluster:
      nodes: [db-primary, db-replica]
      capabilities:
        - database.replication
      params:
        primary: db-primary
        replicas: [db-replica]
```

### Pattern 3: Microservice Configuration

```yaml
inventory:
  nodes:
    api-01: {}
    api-02: {}
    frontend-01: {}
    frontend-02: {}

  clusters:
    api-tier:
      nodes: [api-01, api-02]
      capabilities:
        - cluster.service-mesh

    frontend-tier:
      nodes: [frontend-01, frontend-02]
      capabilities:
        - cluster.service-mesh
```

---

## Troubleshooting

### Node Not Found

**Problem**: A Node specified in the Cluster's `nodes` cannot be found.

**Causes**:
- The Node name is misspelled
- The Node is not defined in `inventory.nodes`

**Solutions**:
1. Define the Node in `inventory.nodes`
2. Verify the Node name spelling

```yaml
# Bad: node1 is not defined
inventory:
  clusters:
    my-cluster:
      nodes:
        - node1  # Error

# Good: define node1
inventory:
  nodes:
    node1: {}

  clusters:
    my-cluster:
      nodes:
        - node1
```

### Cluster Capability Not Applied

**Problem**: A Capability was specified for a Cluster but it does not work as expected.

**Causes**:
- The Recipe ID is incorrect
- Parameters are missing
- A Node-level Capability is specified instead of a Cluster-level one

**Solutions**:
1. Verify the Recipe ID spelling
2. Check the Recipe documentation to confirm it can be used at the Cluster level
3. Verify that all required parameters are present

---

## Next Steps

- [Nodes in Detail]({{< relref "nodes" >}}) - All Node attributes and options
- [Generators in Detail]({{< relref "generators" >}}) - How to generate instances
- [Capabilities and Recipes in Detail]({{< relref "capabilities" >}}) - How to specify Recipes
- [Templates in Detail]({{< relref "templates" >}}) - Using templates

---
title: "Capabilities and Recipe Specification"
weight: 5
---

# Capabilities and Recipe Specification

The Capabilities and provisioners specified on Nodes, Clusters, and Generators are realized through **Recipes**. A Recipe is an executable definition that provides a specific function.

This page explains how to specify Recipes.

## What Is a Capability?

A **Capability** represents a function that a Node or Cluster should have:

- Web server (Nginx, Apache)
- Database (PostgreSQL, MySQL)
- Container orchestrator (Kubernetes)
- Monitoring tool (Prometheus)

Capabilities are realized by Recipes. A Recipe defines the steps required to set up that function.

---

## How to Specify a Recipe ID

There are two main ways to specify a Recipe.

### 1. Specifying by metadata.id

This method specifies the Recipe's `metadata.id` directly. It is the most explicit approach and lets you reliably target a specific Recipe.

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - com.example.nginx-server
```

#### Specifying a Version

You can explicitly specify a version using `@`:

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - com.example.nginx-server@1.2.0
```

When no version is specified, the latest available version is used.

Changing the version of an already applied Capability works as follows:

- Raising the version, or removing the pin so that a newer version is selected, updates the Capability: the `operation: update` tasks of the new version run (see [Update (operation: update)]({{< relref "defining-recipes/defining-tasks#update-operation-update" >}})).
- Pinning a version lower than the applied one makes `nwsctl plan` fail with an error.

#### Uses

- When you want to reliably target a specific Recipe
- When you want to pin a version
- When using a custom Recipe

---

### 2. Specifying by the spec.provide Alias Name

A Recipe can define an alias name in `spec.provide`. Using this alias name allows more abstract specification by hiding implementation details.

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.server  # Specified by alias name
```

#### Narrowing Down with attrs

When multiple Recipes match an alias name alone, you can narrow them down using `attrs` (attributes):

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.server.engine=nginx  # Select the one with engine attribute = nginx

    app-server:
      capabilities:
        - web.server.engine=apache  # Select the one with engine attribute = apache
```

#### Specifying Multiple Attributes

Multiple attributes can be chained with `.`:

```yaml
inventory:
  nodes:
    db-server:
      capabilities:
        - database.type=relational.engine=postgresql
```

In this example, the Recipe that satisfies the following conditions is selected:
- `type` attribute is `relational`
- `engine` attribute is `postgresql`

#### Uses

- When you want to specify a function abstractly without worrying about implementation details
- When you want to flexibly choose from multiple implementations
- When you want to make it easy to swap Recipes

---

## Short Form and Long Form

Capabilities have two notations: a short form (string only) and a long form (object).

### Short Form

When no parameters are needed, you can specify with a string only:

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter
```

### Long Form

When passing parameters, write in object form:

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - id: web.nginx
          params:
            port: 8080
            worker_processes: 4
            ssl_enabled: true
```

#### Required Fields

- **id**: Recipe identifier (metadata.id or alias name)

#### Optional Fields

- **params**: Parameters to pass to the Recipe

---

## Parameters (params)

Parameters can be passed to a Recipe to customize its behavior. The structure and meaning of parameters differ depending on each Recipe's specification.

### Parameters in Node Capabilities

```yaml
inventory:
  nodes:
    db-server:
      capabilities:
        - id: database.postgresql
          params:
            version: "15"
            port: 5432
            max_connections: 200
            shared_buffers: "256MB"
```

### Parameters in Cluster Capabilities

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - id: cluster.kubernetes
          params:
            version: "1.28"
            network_plugin: calico
            pod_network_cidr: "10.244.0.0/16"
```

### The Cluster-level params Field

Clusters also have a top-level `params` field separate from `capabilities`:

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [cp]
          kube_node: [worker1, worker2]
          etcd: [cp]
```

This top-level `params` defines parameters for the entire Cluster.

### Parameters for Generator provisioners

```yaml
infrastructure:
  generators:
    vagrant-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
        disk_size: "40GB"
```

### How to Check Parameters

Refer to each Recipe's documentation to find out what parameters it accepts. Recipes define required parameters, optional parameters, and default values.

---

## provisioner (Generator)

The Generator's `provisioner` follows the same rules as Recipe ID specification:

### Specifying by metadata.id

```yaml
infrastructure:
  generators:
    vms:
      provisioner: com.example.vagrant-provider
```

### Specifying by Alias Name + attrs

```yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
```

In this example, a Recipe with the alias name `infra.vm` and `driver` attribute equal to `vagrant` is selected.

---

## Practical Examples

### Example 1: metadata.id + Version Specification

```yaml
inventory:
  nodes:
    web-01:
      capabilities:
        - com.example.nginx-server@1.2.0

    db-01:
      capabilities:
        - com.example.postgresql@2.1.0
          params:
            version: "15"
            max_connections: 100
```

**Use case**: When you want to pin a specific Recipe and version.

### Example 2: Alias Name Only

```yaml
inventory:
  nodes:
    web-01:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter

    db-01:
      capabilities:
        - database.postgresql
```

**Use case**: When you want to specify a function simply.

### Example 3: Alias Name + attrs Narrowing

```yaml
inventory:
  nodes:
    nginx-server:
      capabilities:
        - web.server.engine=nginx

    apache-server:
      capabilities:
        - web.server.engine=apache

    postgres-db:
      capabilities:
        - database.type=relational.engine=postgresql

    mysql-db:
      capabilities:
        - database.type=relational.engine=mysql
```

**Use case**: When you want to select different implementations of the same type of function.

### Example 4: Long Form + Parameters

```yaml
inventory:
  nodes:
    cache-01:
      capabilities:
        - id: cache.store.type=in-memory
          params:
            max_memory: "1GB"
            eviction_policy: "lru"

    db-01:
      capabilities:
        - id: database.postgresql
          params:
            version: "15"
            port: 5432
            max_connections: 200
```

**Use case**: When you want to customize behavior with parameters.

### Example 5: Cluster Capability

```yaml
inventory:
  nodes:
    cp: {}
    worker1: {}
    worker2: {}

  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - id: cluster.kubernetes
          params:
            version: "1.28"
            network_plugin: calico
      params:
        groups:
          kube_control_plane: [cp]
          kube_node: [worker1, worker2]
          etcd: [cp]
```

**Use case**: When defining processing that spans multiple Nodes.

### Example 6: Generator provisioner

```yaml
infrastructure:
  generators:
    # Alias + attrs
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64

    # Built-in provisioner
    existing-servers:
      provisioner: external-instance
      params:
        instances:
          server-01:
            connection:
              ssh:
                address:
                  host: 192.168.1.10
                  port: 22
                  user: ubuntu
```

**Use case**: When specifying how instances are generated.

---

## Recipe Selection Guide

### metadata.id vs Alias Name

| Specification Method | Advantages | Disadvantages | Use Case |
|----------------------|------------|---------------|----------|
| **metadata.id** | - Reliably targets a specific Recipe<br>- Version can be pinned | - Difficult to swap Recipes<br>- Depends on implementation | - Production environments<br>- When version management is important |
| **Alias name** | - Hides implementation details<br>- Easy to swap Recipes | - Narrowing down required if multiple match | - Development environments<br>- When flexibility is important |

### With or Without Parameters

| Notation | Use Case |
|----------|----------|
| **Short form** (string) | When default settings are sufficient |
| **Long form** (object) | When you want to customize with parameters |

### Using attrs

- Useful when you want to select different implementations of the same type of function
- Example: `web.server.engine=nginx` vs `web.server.engine=apache`

---

## Summary

- **Recipe ID specification**: By metadata.id or alias name + attrs
- **Notation**: Short form (string) or long form (object + params)
- **Uses**:
  - Nodes: Individual functions
  - Clusters: Processing that spans multiple Nodes
  - Generators: How instances are generated

Understanding how to specify Recipes lets you create flexible, maintainable State files.

---

## Next Steps

- [Nodes in Detail]({{< relref "nodes" >}}) - Node Capability details
- [Clusters in Detail]({{< relref "clusters" >}}) - Cluster Capability details
- [Generators in Detail]({{< relref "generators" >}}) - provisioner details
- [Using Reference Recipes]({{< relref "/recipes" >}}) - List of reference Recipes

---
title: "Overview of State Definition"
weight: 1
---

# Overview of State Definition

## What Is State?

In Niwashi, the **desired form** of your infrastructure and application environment is defined as a "State". There are two kinds of State:

- **Target State (Desired State)**: The state that the user defines as the final goal the system should reach.
- **Current State**: The state that the system currently holds at this point in time.

Niwashi computes the difference between these two states and generates a Plan to realize the Target State.

### Writing in YAML Format

The Target State is described in YAML-format files. YAML is a format that is easy for humans to read and write, and by managing it with a version control system (such as Git), you can track the change history of your infrastructure.

### Schema Version

The current version of the State format is **`nws.state/v1`**. Every State file must include the `version` field at the top as follows:

```yaml
version: nws.state/v1
```

---

## Minimal Example

Let's start by looking at the simplest possible State file:

```yaml
version: nws.state/v1

inventory:
  nodes:
    node1: {}
    node2: {}
    node3: {}

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
```

In this example:
- Three logical Nodes (`node1`, `node2`, `node3`) are defined.
- A Generator (`vms`) that uses Vagrant to create three virtual machines is defined.

Niwashi automatically assigns virtual machines to Nodes from this definition.

---

## Basic Structure of State

A State file is composed of the following top-level elements:

```yaml
version: nws.state/v1

inventory:
  nodes:    # Definition of logical Nodes
  clusters: # Definition of Clusters

infrastructure:
  generators: # Definition of instance Generators

template:
  node:     # Node templates
  cluster:  # Cluster templates
  instance: # Instance templates
```

### version (required)

Specifies the schema version. Currently `nws.state/v1` is used.

### inventory (required)

The section that defines the **logical configuration** of the system:

- **nodes** (required): Definitions of individual logical Nodes
- **clusters**: Definitions of Clusters composed of multiple Nodes

#### Nodes

A Node is an individual logical unit of computation that makes up the system.

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
    db-server:
      capabilities:
        - database.postgresql
```

Each Node can have Capabilities specifying the functions it should have.

See [Nodes in Detail]({{< relref "nodes" >}}) for more information.

#### Clusters

A Cluster is a group composed of multiple Nodes. By defining Cluster-level Capabilities, you can realize processing that spans multiple Nodes.

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        - control-plane
        - worker1
        - worker2
      capabilities:
        - cluster.kubernetes
```

See [Clusters in Detail]({{< relref "clusters" >}}) for more information.

### infrastructure

The section that defines the physical/virtual Infrastructure.

- **generators**: Definitions of Generators that create instances (virtual machines, etc.)

```yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

A Generator creates instances based on the Recipe specified by `provisioner`.

See [Generators in Detail]({{< relref "generators" >}}) for more information.

### template

Common settings for Nodes, Clusters, and instances can be defined as templates.

```yaml
template:
  node:
    base-node:
      labels:
        environment: production
        managed_by: niwashi

inventory:
  nodes:
    web-01:
      templates: [base-node]
      capabilities:
        - web.nginx
```

Using templates lets you efficiently define multiple elements that share the same configuration.

See [Templates in Detail]({{< relref "templates" >}}) for more information.

---

## Separating Logical and Physical Configuration

An important characteristic of Niwashi is that it lets you separate **logical configuration** (what kind of system you want to build) from **physical configuration** (where to deploy it).

### Example: Environment-Specific Settings

#### base.yaml (logical configuration - common)

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
```

#### dev.yaml (physical configuration - development environment)

```yaml
version: nws.state/v1

infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
        box: ubuntu/jammy64
        cpus: 1
        memory: 1024
```

#### prod.yaml (physical configuration - production environment)

```yaml
version: nws.state/v1

infrastructure:
  generators:
    prod-servers:
      provisioner: external-instance
      params:
        instances:
          web-prod:
            connection:
              ssh:
                address:
                  host: web.example.com
                  port: 22
                  user: deploy
          db-prod:
            connection:
              ssh:
                address:
                  host: db.example.com
                  port: 22
                  user: deploy
```

#### Switching Environments

```bash
# Development environment
nwsctl plan -t base.yaml -t dev.yaml

# Production environment
nwsctl plan -t base.yaml -t prod.yaml
```

Multiple State files are merged in the order specified. This lets you manage the differences between environments in separate files.

---

## Capabilities and Recipes

In Niwashi, the functions that Nodes and Clusters should have are specified as "Capabilities". Capabilities are realized through **Recipes**.

### Basic Specification

#### Short form (string only)

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter
```

#### Long form (with parameters)

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
```

### Ways to Specify a Recipe

Recipes can be specified in the following ways:

1. **Alias name**: `web.nginx`
2. **Alias + attributes**: `database.type=relational.engine=postgresql`
3. **Recipe ID**: `com.example.nginx-server`
4. **Recipe ID + version**: `com.example.nginx-server@1.2.0`

See [Capabilities and Recipes in Detail]({{< relref "capabilities" >}}) for more information.

---

## Using State Files with Commands

Defined State files are used with the `nwsctl` command.

### Creating a Plan

```bash
nwsctl plan -t target.yaml
```

This command generates a Plan to realize the Target State.

You can also specify multiple files to merge them:

```bash
nwsctl plan -t base.yaml -t environment.yaml -t app.yaml
```

### Exporting State

```bash
# Output to standard output
nwsctl export

# Output to a file
nwsctl export -o current-state.yaml
```

Outputs the current State in YAML format.

---

## Next Steps

### Learning the Basics

To learn the basic concepts of State definition step by step, refer to the following pages:

- [State Structure]({{< relref "basic-concepts/state-structure" >}}) - Details on top-level elements
- [Inventory Basics]({{< relref "basic-concepts/inventory" >}}) - Basics of Nodes and Clusters
- [Infrastructure Basics]({{< relref "basic-concepts/infrastructure" >}}) - Basics of Generators
- [Identifier Naming Rules]({{< relref "basic-concepts/naming-rules" >}}) - Naming rules

### Learning More

For the complete reference of each element, refer to the following pages:

- [Nodes in Detail]({{< relref "nodes" >}}) - All Node attributes and usage
- [Clusters in Detail]({{< relref "clusters" >}}) - All Cluster attributes and usage
- [Generators in Detail]({{< relref "generators" >}}) - All Generator attributes and usage
- [Capabilities and Recipes in Detail]({{< relref "capabilities" >}}) - All ways to specify recipes
- [Templates in Detail]({{< relref "templates" >}}) - Everything about the template feature

### Advanced Usage

To learn practical patterns, refer to the following pages:

- [Managing Multiple Environments]({{< relref "advanced/multi-environment" >}}) - Switching between dev/prod environments
- [State Merging in Detail]({{< relref "advanced/state-merging" >}}) - Merge specifications

### Practical Examples

To see complete working examples, refer to the following pages:

- [Building a Kubernetes Cluster]({{< relref "../examples/kubernetes-cluster" >}})
- Configuring a Web Application

---

This concludes the overview of State definition in Niwashi. We recommend starting with the minimal example and gradually adding features from there.

---
title: "State Structure"
weight: 1
---

# State Structure

This page explains the top-level structure of a State file and the role of each element.

## Overall Structure

A State file is composed of the following top-level elements:

```yaml
version: nws.state/v1

metadata:
  project: my-project

inventory:
  nodes:
    # Definition of logical Nodes
  clusters:
    # Definition of Clusters

infrastructure:
  generators:
    # Definition of instance Generators

template:
  node:
    # Node templates
  cluster:
    # Cluster templates
  instance:
    # Instance templates
```

---

## Top-Level Elements

### version (required)

Specifies the schema version. Currently `nws.state/v1` is used.

```yaml
version: nws.state/v1
```

This version identifier allows Niwashi to interpret the file with the appropriate schema. Even if the format is extended in the future, specifying the version explicitly preserves backward compatibility.

**Required**: Needed in every State file.

---

### metadata (optional)

Defines metadata about the entire State file.

```yaml
metadata:
  project: my-k8s-cluster
```

#### Available Fields

- **project**: Project name (any string)

The metadata section is optional.

---

### inventory (required)

The section that defines the **logical configuration** of the system. Describes "what kind of system you want to build".

```yaml
inventory:
  nodes:
    web-server: {}
    db-server: {}
  clusters:
    my-cluster:
      nodes:
        - web-server
        - db-server
```

#### Elements Included

- **nodes** (required): Definitions of individual logical Nodes
- **clusters** (optional): Definitions of Clusters composed of multiple Nodes

**Required**: The `inventory` section and `nodes` within it are required.

See [Inventory Basics]({{< relref "inventory" >}}) for more information.

---

### infrastructure (required)

The section that defines physical/virtual Infrastructure. Describes "where to deploy".

```yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
```

#### Elements Included

- **generators**: Definitions of Generators that create and manage instances (virtual machines, containers, existing servers, etc.)

**Required**: Instances are needed to deploy Nodes to actual machines. Define Generators in the `infrastructure` section.

See [Infrastructure Basics]({{< relref "infrastructure" >}}) for more information.

---

### template (optional)

Common settings for Nodes, Clusters, and instances can be defined as templates.

```yaml
template:
  node:
    base-node:
      labels:
        environment: production
        managed_by: niwashi
  cluster:
    base-cluster:
      labels:
        managed_by: niwashi
  instance:
    standard-vm:
      provisioner: infra.vm.driver=vagrant
      params:
        cpus: 2
        memory: 2048

inventory:
  nodes:
    web-01:
      templates: [base-node]
      capabilities:
        - web.nginx
```

Using templates allows you to efficiently define multiple elements with the same settings, avoiding configuration duplication and improving maintainability.

See [Templates in Detail]({{< relref "../templates" >}}) for more information.

---

## Logical and Physical Configuration

An important characteristic of Niwashi is that it can separate **logical configuration** from **physical configuration**.

### Logical Configuration (inventory)

Defines "what you want to build":

- What roles do the Nodes need to play
- What functions do the Nodes need
- How to group the Nodes

```yaml
inventory:
  nodes:
    web:
      capabilities:
        - web.nginx
    db:
      capabilities:
        - database.postgresql
```

### Physical Configuration (infrastructure)

Defines "where to build it":

- Local virtual machines, cloud, or existing servers
- How many instances are needed
- How much resources (CPU, memory) are needed

```yaml
infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
        box: ubuntu/jammy64
```

### Benefits of Separation

This separation allows the same logical configuration to be deployed to different environments:

```bash
# Development environment: local virtual machines
nwsctl plan -t app.yaml -t infra-dev.yaml

# Production environment: existing cloud servers
nwsctl plan -t app.yaml -t infra-prod.yaml
```

Write the logical configuration in `app.yaml`, and write the physical configuration for each environment in `infra-dev.yaml` and `infra-prod.yaml`.

---

## Minimal Configuration

The minimum required elements are as follows:

```yaml
version: nws.state/v1

inventory:
  nodes:
    node1: {}
```

In this example:
- Version specification (required)
- One logical Node definition (inventory.nodes is required)

In practice, however, you need to define `infrastructure.generators` to provide instances.

---

## Complete Example

A complete example including all elements:

```yaml
version: nws.state/v1

metadata:
  project: my-web-app

template:
  node:
    base:
      labels:
        environment: production

inventory:
  nodes:
    web:
      templates: [base]
      capabilities:
        - web.nginx
    db:
      templates: [base]
      capabilities:
        - database.postgresql

  clusters:
    app-cluster:
      nodes:
        - web
        - db

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

---

## Next Steps

Let's learn more about each element:

- [Inventory Basics]({{< relref "inventory" >}}) - Basics of Nodes and Clusters
- [Infrastructure Basics]({{< relref "infrastructure" >}}) - Basics of Generators
- [Identifier Naming Rules]({{< relref "naming-rules" >}}) - Naming rules and best practices

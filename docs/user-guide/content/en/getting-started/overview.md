---
title: "Overview"
weight: 1
---

# Overview

## What is Niwashi

Niwashi is a tool for unified infrastructure provisioning and configuration management.
It lets you declaratively define a desired state and then plans and executes the changes needed to transition from the current state to that desired state.

---

## How to Use It

Niwashi is operated through the command-line tool **`nwsctl`**.

Basic workflow:
1. **Define the desired state**: Describe it in a YAML file
2. **Create a Plan**: `nwsctl plan -t state.yaml`
3. **Execute the Plan**: `nwsctl apply`

For details, see [Workflow]({{< relref "workflow" >}}).

---

## Features of Niwashi

### Recipe System

The greatest strength of Niwashi is its flexible **Recipe system**.

A Recipe is a module that packages provisioning and configuration management procedures. By declaratively defining "what to do" and delegating execution to the Recipe, input parameters and execution results are handled through a standardized interface.

- Reuse the same Recipe across multiple environments and projects
- Combine Recipes to achieve complex configurations
- Hide implementation details and use a simple interface

The Niwashi project provides reference Recipes for popular open-source tools:

| Recipe | Capability ID | Description |
|--------|--------------|-------------|
| **Vagrant** | `infra.vm.driver=vagrant` | Automated VM creation — ideal for local development environments |
| **Ansible** | `adapter.tool.ansible` | Adapter for running Ansible playbooks — enables coordination between Recipes |
| **Kubernetes** | `cluster.kubernetes` | Building and managing Kubernetes clusters |

For details, see [Reference Recipes]({{< relref "recipes" >}}).

### Integrated Workflow

The following tasks, which were previously performed by separate tools, can now be executed in a single unified workflow:

1. Infrastructure provisioning (e.g., Vagrant Recipe)
2. Running configuration management tools (e.g., Ansible Recipe)
3. Building clusters (e.g., Kubernetes Recipe)

### Declarative State Definition

In Niwashi, the desired state is described declaratively in YAML files (State). State is structured to separate **logical configuration (Inventory)** from **physical Infrastructure**, allowing the same logical configuration to run on different Infrastructure.

- Reuse common settings with templates
- Manage differences across development, staging, and production environments using State file merge functionality

---

## Use Case Examples

### 1. Building a Kubernetes Cluster

An example using reference Recipes (Vagrant, Kubernetes):

```yaml
# Define the logical configuration
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes  # reference Recipe

# Define the Infrastructure
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant  # reference Recipe
      params:
        count: 3
```

→ The Vagrant Recipe creates 3 VMs, and the Kubernetes Recipe automatically builds the Cluster

**Recipe coordination**:
This Kubernetes Recipe internally calls the Ansible adapter (`adapter.tool.ansible`) to run the Kubespray playbook. This is how coordination between Recipes is achieved.

### 2. Switching Between Development and Production Environments

An example with a shared logical configuration, switching only the Infrastructure:

```bash
# Development environment (Vagrant Recipe)
nwsctl plan -t infra-dev.yaml -t app.yaml

# Production environment (existing servers)
nwsctl plan -t infra-prod.yaml -t app.yaml
```

→ The Infrastructure definition is loaded first, then merged with the logical configuration (app.yaml)

---

## Next Steps

- [Workflow]({{< relref "workflow" >}}) - The basic working flow with Niwashi
- [Architecture]({{< relref "architecture" >}}) - Niwashi's internal structure
- [Reference Recipes]({{< relref "recipes" >}}) - List of available reference Recipes
- [Defining Recipes]({{< relref "defining-recipes" >}}) - How to create your own Recipes

---
title: "Niwashi User Guide"
weight: 1
---

# Niwashi User Guide

Niwashi is a tool for unified infrastructure provisioning and configuration management. It creates and executes an execution plan to transition from the current state to the desired state.

## Table of Contents

- [Getting Started](getting-started/)
  - [What You Can Do with Niwashi](getting-started/overview/)
  - [Workflow](getting-started/workflow/)
  - [Architecture](getting-started/architecture/)
- [Defining Desired State](defining-desired-state/)
  - [Overview of State Definition](defining-desired-state/overview/)
  - [Basic Concepts](defining-desired-state/basic-concepts/)
    - [Basic Structure of State](defining-desired-state/basic-concepts/state-structure/)
    - [Inventory Basics](defining-desired-state/basic-concepts/inventory/)
    - [Infrastructure Basics](defining-desired-state/basic-concepts/infrastructure/)
    - [Identifier Naming Rules](defining-desired-state/basic-concepts/naming-rules/)
  - [Capability and Recipe Details](defining-desired-state/capabilities/)
  - [Node Details](defining-desired-state/nodes/)
  - [Cluster Details](defining-desired-state/clusters/)
  - [Generator Details](defining-desired-state/generators/)
  - [Template Details](defining-desired-state/templates/)
  - [Advanced Usage](defining-desired-state/advanced/)
    - [Managing Multiple Environments](defining-desired-state/advanced/multi-environment/)
    - [State Merging Details](defining-desired-state/advanced/state-merging/)
- [nwsctl Command Reference](execution/)
  - [nwsctl init](execution/init/)
  - [nwsctl plan](execution/plan/)
  - [nwsctl apply](execution/apply/)
  - [nwsctl ssh](execution/ssh/)
  - [nwsctl export](execution/export/)
- [Defining Recipes](defining-recipes/)
  - [Defining Node Capabilities](defining-recipes/node-capability/)
  - [Defining Infrastructure Provisioners](defining-recipes/infrastructure-provisioning/)
  - [Using Adapters](defining-recipes/using-adapters/)
  - [Defining Adapters](defining-recipes/defining-adapters/)
  - [Updating State with stateChanges](defining-recipes/state-changes/)
  - [Combining Multiple Recipes](defining-recipes/catalog/)
  - [Recipe Loading Specification](defining-recipes/recipe-loading/)
  - [Defining Cluster Capabilities](defining-recipes/cluster-capability/) 🚧
  - [Defining Host Configuration](defining-recipes/host-configuration/) 🚧
- [Examples](examples/)
  - [Building a Kubernetes Cluster](examples/kubernetes-cluster/)
- [Recipes](recipes/)
  - [Ansible](recipes/ansible/)
  - [Kubernetes](recipes/kubernetes/)
  - [Vagrant](recipes/vagrant/)

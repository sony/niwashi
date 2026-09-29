# Niwashi (NWS)

*"Niwashi" (庭師) means "gardener" in Japanese — just as a gardener carefully cultivates a garden, Niwashi cultivates your infrastructure through composable automation workflows.*

Niwashi is a unified infrastructure provisioning and configuration management tool.
It lets you declaratively define a desired state and then plans and executes the changes needed to transition from the current state to that desired state.

## Features

**Recipe System** — The central concept of Niwashi. A Recipe is a module that packages provisioning and configuration management procedures. Recipes can be reused across projects, composed to achieve complex configurations, and shared within your team.

- **Reference Recipes**: Provided by the Niwashi project for popular open-source tools
- **Custom Recipes**: Define your own for internal tools, custom scripts, etc.

**Integrated Workflow** — Tasks that previously required separate tools can now be executed in a single unified workflow:

- Infrastructure provisioning (e.g., Vagrant Recipe)
- Configuration management (e.g., Ansible Recipe)
- Cluster orchestration (e.g., Kubernetes Recipe)

**Declarative State Definition** — Describe the desired state in YAML (State). State separates **logical configuration (Inventory)** from **physical Infrastructure**, allowing the same logical configuration to run on different infrastructure. State file merging enables multi-environment management across development, staging, and production.

## Installation

Download the archive for your architecture from the [releases page](https://github.com/sony/niwashi/releases).

```bash
# 1. Extract the archive
tar -xzf nwsctl_Linux_x86_64.tar.gz

# 2. Place the binary
sudo mv nwsctl /usr/local/bin/

# 3. Verify
nwsctl version
```

For Windows/macOS or alternative install locations, see the [Installation guide](docs/user-guide/content/en/getting-started/installation.md).

## Quick Start

Clone the [Reference Recipes](https://github.com/sony/niwashi-recipe) into `./recipe`:

```bash
git clone https://github.com/sony/niwashi-recipe recipe
```

Then:

```bash
# 1. Initialize the workspace
nwsctl init

# 2. Define the desired state
vi state.yaml

# 3. Create a Plan
nwsctl plan --recipe-dir ./recipe -t state.yaml

# 4. Execute the Plan
nwsctl apply --plan plan.json --recipe-dir ./recipe
```

## Reference Recipes

| Recipe | Capability ID | Description |
|--------|--------------|-------------|
| **Vagrant** | `infra.vm.driver=vagrant` | Automated VM creation — ideal for local development environments |
| **Ansible** | `adapter.tool.ansible` | Adapter for running Ansible playbooks — enables coordination between Recipes |
| **Kubernetes** | `cluster.kubernetes` | Container orchestration — building and managing Kubernetes clusters |

## Example: Building a Kubernetes Cluster

Using reference Recipes (Vagrant + Kubernetes):

```yaml
# Define the logical configuration
inventory:
  nodes:
    cp:
    worker1:
    worker2:

  clusters:
    k8s-cluster:
      nodes:
        - cp
        - worker1
        - worker2
      capabilities:
        - cluster.kubernetes  # reference Recipe
      params:
        groups:
          kube_control_plane: [cp]
          kube_node: [worker1, worker2]
          etcd: [cp]

# Define the Infrastructure
infrastructure:
  generators:
    vm:
      provisioner: infra.vm.driver=vagrant  # reference Recipe
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

The Vagrant Recipe creates 3 VMs; the Kubernetes Recipe automatically builds the cluster.
Internally, the Kubernetes Recipe calls the Ansible adapter (`adapter.tool.ansible`) to run the Kubespray playbook — this is how Recipe coordination works.

## Documentation

The full [User Guide](docs/user-guide/content/en/_index.md) is available as Markdown source under `docs/user-guide/`. (This will be replaced with a link to the published HTML site once it's available.)

Good starting points:

- [What You Can Do with Niwashi](docs/user-guide/content/en/getting-started/overview.md) — Overview of features and use cases
- [Workflow](docs/user-guide/content/en/getting-started/workflow.md) — The basic step-by-step usage
- [Tutorial: Your First Recipe](docs/user-guide/content/en/defining-recipes/first-recipe.md) — Write and apply your first Recipe
- [Defining Recipes](docs/user-guide/content/en/defining-recipes/_index.md) — How to create your own Recipes

## Contributing

Niwashi does not currently accept external contributions; see [CONTRIBUTING.md](CONTRIBUTING.md) for details and the current policy. Bug reports and feature requests via [Issues](https://github.com/sony/niwashi/issues) are welcome.

This project follows the [Code of Conduct](CODE_OF_CONDUCT.md).

## License

Licensed under the [Apache License 2.0](LICENSE).

Copyright 2026 Sony Group Corporation.

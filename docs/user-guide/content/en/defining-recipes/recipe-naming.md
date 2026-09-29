---
title: "Recipe Naming Guidelines"
weight: 9
---

# Recipe Naming Guidelines

This page describes naming conventions for recipe identifiers and capability aliases. Following these guidelines keeps recipes easy to discover, understand, and compose.

---

## Recipe ID (`metadata.id`)

A recipe ID has the form `<org>/<name>`:

```
my-org/ansible.installer
my-org/kubernetes.by.kubespray
```

### `org` — Organization or namespace

The `org` segment identifies who owns the recipe. Use your GitHub organization name, company name, or project name.

| Example | When to use |
|---------|-------------|
| `nws` | Reserved for Niwashi's reference recipes (built-in and distributed) |
| `my-org` | Your GitHub organization or company name |
| `acme` | Short project or team name |

**Reserved namespace**: `nws` is reserved for Niwashi. Do not use it for third-party recipes.

**Discouraged names**: Avoid org names that imply official endorsement or verification, such as `core`, `official`, `system`, `builtin`, `verified`, `trusted`, or `standard`. Using such names may mislead users into treating your recipes as officially sanctioned by Niwashi.

### `name` — Recipe name

The `name` segment identifies what the recipe does within the organization. Use lowercase, dot-separated segments to express hierarchy:

```
ansible.installer       # Ansible installer
kubernetes.by.kubespray # Kubernetes setup using Kubespray
vagrant.vm-cluster      # Vagrant-based VM cluster
nginx                   # Simple single-segment name is also valid
```

- Use dots (`.`) to group related recipes (e.g., `ansible.installer`, `ansible.adapter`)
- The rightmost segment is the most specific descriptor
- Hyphens (`-`) are allowed within a segment; underscores (`_`) are also allowed

### FQID

The **FQID (Fully Qualified ID)** combines `metadata.id` and `metadata.version` with `@`:

```
my-org/ansible.installer@1.0.0
nws/kubernetes.by.kubespray@0.1.0
```

FQIDs are used in `spec.requires` for exact or version-constrained references. See [Recipe Loading Specification]({{< relref "recipe-loading" >}}) for version resolution rules.

---

## Capability Alias (`spec.provides[].name`)

A capability alias is a **semantic name** for what a recipe provides — independent of who wrote it or which version it is. The alias is what consumers write in `capabilities:` or `spec.requires:`.

### Naming structure

Aliases use a dot-separated hierarchy. The **first segment is the `kind`** of the recipe that provides it:

> **Note**: In fields where the `kind` is determined by context — such as `capabilities`, `provisioner`, and `toolRef` — the kind prefix will be made optional in a future release. In `spec.requires`, the kind prefix will remain required because multiple kinds can be mixed there.

| First segment | Corresponding `kind` | Description |
|---------------|----------------------|-------------|
| `host.tool`   | `host`               | Tool installed on the host (e.g. Ansible, Git) |
| `host.service`| `host`               | Service running on the host (e.g. Docker daemon) |
| `adapter.tool`| `adapter`            | Adapter for a tool-type execution environment |
| `adapter.runtime` | `adapter`        | Adapter for a runtime-type execution environment (e.g. Python venv) |
| `node.runtime`| `node`               | Runtime capability installed on a node |
| `cluster`     | `cluster`            | Cluster-level capability (e.g. Kubernetes) |
| `infra.vm`    | `infra`              | VM infrastructure |
| `infra.network` | `infra`            | Network infrastructure |
| `infra.compute` | `infra`            | Compute infrastructure (cloud instances, bare metal) |

> **Note**: Single-segment names (e.g. `external-instance`) are reserved for Niwashi built-in system recipes.

### Examples

```yaml
# kind: host — tool installer
spec:
  provides:
    - name: host.tool.ansible

# kind: host — service installer
spec:
  provides:
    - name: host.service.docker

# kind: adapter — tool adapter
spec:
  provides:
    - name: adapter.tool.ansible

# kind: adapter — runtime adapter
spec:
  provides:
    - name: adapter.runtime.python-venv

# kind: node — runtime capability
spec:
  provides:
    - name: node.runtime.container

# kind: cluster — cluster capability
spec:
  provides:
    - name: cluster.kubernetes
      attrs: { by: kubespray }

# kind: infra — VM provisioner
spec:
  provides:
    - name: infra.vm
      attrs: { driver: vagrant }
```

### Using `attrs` to distinguish implementations

When multiple recipes provide the same alias name, use `attrs` to distinguish them:

```yaml
# Recipe A
spec:
  provides:
    - name: cluster.kubernetes
      attrs: { by: kubespray }

# Recipe B
spec:
  provides:
    - name: cluster.kubernetes
      attrs: { by: rke2 }
```

Consumers can then select a specific implementation:

```yaml
capabilities:
  - cluster.kubernetes              # resolves automatically if only one candidate
  - cluster.kubernetes.by=kubespray # selects Recipe A explicitly
```

See [Recipe Loading Specification]({{< relref "recipe-loading" >}}) for alias resolution rules.

### `attrs` key naming

- Use lowercase letters and underscores only (`[a-z][a-z0-9_]*`)
- To pin a specific version, use FQID rather than an attrs value — version strings like `2.16.3` cannot be used as attrs values because dots are reserved as separators

**Criteria for defining an attrs key**: An attrs key must serve to distinguish between multiple recipes that share the same `provides.name`. Do not use attrs for values that are determined at runtime (e.g. `os`, `arch`) — those belong in the `where` condition on tasks.

The following keys are defined for common cases:

| Key | When to use | Example |
|-----|-------------|---------|
| `engine` | Selecting the runtime engine or implementation of a software stack | `cluster.kubernetes.engine=k3s`, `node.runtime.proxy.engine=nginx` |
| `by` | Selecting the tool responsible for setup or configuration | `cluster.kubernetes.by=kubespray` |
| `driver` | Selecting the driver or provider at the infrastructure layer | `infra.vm.driver=vagrant` |
| `from` | Selecting the input source or format | `external-instance.from=file`, `external-instance.from=ssh_config` |

For cases not covered by the above, define a key with a clear noun that describes the axis of variation.

```yaml
# Wrong: version string in attrs value
attrs: { version: 2.16.3 }

# Correct: use FQID in spec.requires for version pinning
spec:
  requires:
    - my-org/my-recipe@2.16.3
```

---

## Task Names (`tasks[].name`)

Task names are human-readable labels shown in logs and plan output:

- Start with an alphanumeric character
- Subsequent characters may be alphanumeric, spaces, hyphens, underscores, or periods
- Uppercase letters are allowed

```yaml
tasks:
  - name: Install Ansible
  - name: check-version
  - name: Apply k8s manifests
```

> **Note**: When used as filenames internally, task names are lowercased and spaces/periods are replaced. Names that differ only in case, spaces, or periods are considered duplicates (e.g. `My Task`, `my task`, and `my.task` all collide).

---

## Command Names (`spec.commands` in adapter recipes)

Command names follow POSIX utility naming conventions:

- Start with a lowercase letter
- Subsequent characters may be lowercase letters, digits, hyphens, or underscores
- Uppercase letters are not permitted

```yaml
spec:
  commands:
    apply:
      ...
    dry-run:
      ...
    install_package:
      ...
```

---

## Summary

| Identifier | Pattern | Example |
|------------|---------|---------|
| `metadata.id` | `<org>/<name>` (lowercase, dots/hyphens allowed) | `my-org/ansible.installer` |
| `metadata.version` | Semantic version `MAJOR.MINOR.PATCH` | `1.2.0` |
| FQID | `<id>@<version>` | `my-org/ansible.installer@1.2.0` |
| `spec.provides[].name` | Dot-separated, starts with kind prefix | `host.tool.ansible` |
| `spec.provides[].attrs` key | `[a-z][a-z0-9_]*` | `by`, `driver` |
| `spec.provides[].attrs` value | `[a-z0-9][a-z0-9_-]*` | `kubespray`, `vagrant` |
| `tasks[].name` | Human-readable, alphanumeric start | `Install Ansible` |
| `spec.commands` key | POSIX-style, starts with lowercase letter | `apply`, `dry-run` |

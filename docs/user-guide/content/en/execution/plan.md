---
title: "nwsctl plan"
weight: 2
---

# nwsctl plan

Compares the current state with the desired state and creates an execution Plan.

```
nwsctl plan [flags]
```

---

## Overview

`nwsctl plan` compares the current state stored in the Workspace (`state.json`) with the desired state specified by `--target`, calculates a task DAG to close the gap, and writes it to `plan.json`.

---

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--target` | `-t` | | Path to the desired State file (can be specified multiple times) |
| `--recipe-dir` | | | Path to a Recipe directory (can be specified multiple times) |
| `--out` | `-o` | `plan.json` | Output path for the generated Plan file |
| `--destroy` | | `false` | Create a Plan to delete all resources |
| `--prune` | | `false` | Create a Plan containing only deletion operations |
| `--profile` | | | Path to a profile file |
| `--with-init` | | `false` | Initialize with an empty state before creating the Plan |
| `--work-dir` | | `.niwashi` | Path to the Workspace |

---

## Usage

### Basic Usage

```bash
nwsctl plan \
  -t infra.yaml \
  -t target.yaml \
  --recipe-dir ./recipes
```

When multiple `--target` values are specified, they are merged from left to right and used as the desired state. For details, see [State Merging Details]({{< relref "defining-desired-state/advanced/state-merging" >}}).

### Updating Applied Capabilities

A normal `nwsctl plan` also detects Capabilities whose `params` or recipe version changed since the last apply, and includes their `operation: update` tasks in the Plan. No extra flag is required.

If a changed Capability's recipe has no `operation: update` task, the change is not applied and is listed in the plan output instead:

```
Pending Updates (recipe has no update tasks):
- node:my-org/myapp@1.0.0:web-1
```

See [Update (operation: update)]({{< relref "defining-recipes/defining-tasks#update-operation-update" >}}) for details.

### Changing the Output Path

```bash
nwsctl plan -t target.yaml --recipe-dir ./recipes -o my-plan.json
```

Pass the generated `plan.json` to `nwsctl apply`.

### Plan to Delete All Resources (--destroy)

```bash
nwsctl plan --recipe-dir ./recipes --destroy
```

Calculates with an empty desired state. A Plan to delete all resources in the current state is generated. To execute it, you must also specify `--destroy` on `nwsctl apply` (see [nwsctl apply]({{< relref "apply" >}})).

### Plan to Delete Only Unused Resources (--prune)

```bash
nwsctl plan -t target.yaml --recipe-dir ./recipes --prune
```

Plans only the deletion of resources not included in the desired state. No addition operations are included. To execute it, you must also specify `--prune` on `nwsctl apply` (see [nwsctl apply]({{< relref "apply" >}})).

### Performing Initial Setup in One Step (--with-init)

```bash
nwsctl plan -t target.yaml --recipe-dir ./recipes --with-init
```

Runs `nwsctl init` and then creates the Plan. Useful when the Workspace does not yet exist.

---

## Profile (--profile)

Using a profile file lets you customize the behavior of the Plan.

```yaml
version: nws.profile/v1
name: dev

# Tool aliases (Capability name → mapping to Recipe FQID)
toolAlias:
  infra.vm: infra.vm.driver=vagrant   # Resolves infra.vm to Vagrant

# Capability bindings (arbitrary capability name → mapping to Recipe FQID)
capabilityBinding:
  my-org.nginx: my-org/nginx@1.0.0

# Parameter overrides
params:
  capability:
    my-org.nginx:
      nginx_port: 8080
```

```bash
nwsctl plan -t target.yaml --recipe-dir ./recipes --profile profile-dev.yaml
```

---

## Recipe Integrity

When a plan is created, nwsctl records a fingerprint of each recipe used in the plan file. `nwsctl apply` compares these fingerprints with the recipes in `--recipe-dir`. See [Recipe Integrity Check]({{< relref "apply#recipe-integrity-check" >}}) for what happens on a mismatch.

---

## Next Steps

- [nwsctl apply]({{< relref "apply" >}}) — Execute the created Plan

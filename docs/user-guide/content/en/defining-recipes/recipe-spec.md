---
title: "Recipe Format Reference"
weight: 0
---

# Recipe Format Reference

All recipes share a common top-level format. This page describes each field under `spec`. Fields that are only available in specific `kind` values are marked as **\[kind: xxx only\]**.

---

## Top-level fields

```yaml
version: nws.recipe/v1
kind: node          # node / infra / host / cluster / adapter
metadata:
  id: my-org/my-recipe
  version: "1.0.0"
  description: "Description"
spec:
  ...
```

| Field | Required | Description |
|-------|----------|-------------|
| `version` | Yes | Schema version. Specify `nws.recipe/v1` |
| `kind` | Yes | Recipe kind. See below |
| `metadata.id` | Yes | Recipe ID. See [Recipe Naming]({{< relref "recipe-naming" >}}) for naming conventions |
| `metadata.version` | Yes | Recipe version |
| `metadata.description` | | Description text |

---

## kind

Specifies the type of recipe.

| Value | Description |
|-------|-------------|
| `node` | Defines a node-level Capability |
| `cluster` | Defines a cluster-level Capability |
| `infra` | Defines infrastructure provisioning |
| `host` | Defines configuration on the host running nwsctl |
| `adapter` | Defines an adapter called by other recipes |

---

## spec.provides

Defines aliases for the Capability this recipe provides. These correspond to names specified in `capabilities` in the State file.

> **\[All kinds except kind: adapter\]**

```yaml
spec:
  provides:
    - name: my-org.my-feature        # simple alias
    - name: my-org.my-feature
      attrs:
        variant: lite                # alias with attributes
```

Multiple aliases can be defined. Attributes (`attrs`) allow multiple implementations to coexist under the same Capability name. See [Recipe Loading]({{< relref "recipe-loading" >}}) for details.

---

## spec.requires

Declares dependencies on other recipes that must run before this recipe. The planner resolves these dependencies to determine execution order.

> **\[All kinds except kind: adapter\]**

Each entry can be specified as a plain string or as a map with an optional `as:` field:

```yaml
spec:
  requires:
    - host.tool.ansible           # plain string
    - name: some.capability       # map form — required when using 'as:'
      as: upstream                # alias for {{ .Stores.upstream }} access
```

The `as:` field creates an alias for the required recipe's `store/` data, accessible via `{{ .Stores.<alias>.<key> }}` in task templates. See [Using the store/ Subpath]({{< relref "state-changes#using-the-store-subpath" >}}) for details.

The kinds that can be declared as dependencies depend on the declaring recipe's `kind`:

| Declaring kind | Allowed dependency kinds |
|----------------|--------------------------|
| `host` | `host` |
| `infra` | `host`, `infra` |
| `node` | `host`, `node` |
| `cluster` | `host`, `node`, `cluster` |

Declaring a dependency from `kind: node` or `kind: cluster` to `kind: infra` is a plan-time error. Information such as IP addresses, OS, and architecture of bound instances is propagated to the logical layer through separate mechanisms provided by Niwashi, so direct access to infra recipe data is not needed.

---

## spec.defaults

Defines default parameter values and environment variables that the recipe accepts. Users writing State can customize recipe behavior by specifying `params`.

> **\[All kinds\]**

```yaml
spec:
  defaults:
    params:
      version: "1.0.0"
      config_file: "/etc/myapp/config.yaml"
    env:
      MY_ENV_VAR: "default_value"
```

For `kind: adapter`, these defaults are merged with the calling recipe's `spec.defaults`.

---

## spec.assets

Specifies files to bundle with the recipe. For `exec.remote`, files are automatically copied to the target node.

> **\[All kinds except kind: adapter\]**

```yaml
spec:
  assets:
    - "scripts/setup.sh"
    - "config/default.conf"
```

Paths are relative to the directory containing `nws-recipe.yaml`.

---

## spec.workspace

Specifies how the task working directory (`{{ .Paths.work_dir }}`) is managed.

> **\[All kinds except kind: adapter\]**

```yaml
spec:
  workspace:
    mode: ephemeral    # create a new working directory for each run (default)
    # mode: persistent  # retain the working directory across runs
```

| Value | Description |
|-------|-------------|
| `ephemeral` (default) | A new working directory is created for each run |
| `persistent` | The same working directory is reused across runs, keeping files from previous runs. Use this for recipes that manage files that must survive between runs, such as VM images (e.g. the Vagrant provisioner) |

### Working Directory Structure

The workspace (`--work-dir`, default `.niwashi`) is laid out as follows.

```
<work-dir>/                       # workspace ({{ .Paths.workspace }})
├── state/
│   └── state.json                # current State
├── runs/
│   └── <run-id>/                 # created per run
│       ├── plan.json             # plan file for this run
│       └── <phase>-<target>/     # per phase and execution target (e.g. node-web-server)
│           └── <recipe-fqid>/    # per recipe ← the ephemeral work_dir
│               ├── inputs/       # parameters and State snapshot ({{ .Paths.input_dir }})
│               ├── outputs/      # data shared between tasks ({{ .Outputs }} / {{ .Paths.output_dir }})
│               └── logs/         # rendered scripts and logs ({{ .Paths.log_dir }})
└── store/
    └── <phase>-<target>/
        └── <recipe-fqid>/        # ← the persistent work_dir (kept across runs)
```

- Working directories are created per "**phase + execution target** (e.g. `node-web-server`) × **recipe FQID**". Even for the same recipe, a different target node results in a different directory. Phases without an execution target (such as host) use the phase name alone
- With `mode: ephemeral`, `{{ .Paths.work_dir }}` points under `runs/<run-id>/`, so each run starts from an empty directory
- With `mode: persistent`, `{{ .Paths.work_dir }}` points under `store/`, and the same directory is reused across runs
- `inputs/`, `outputs/`, and `logs/` are always created under `runs/<run-id>/` (per run) regardless of the `mode` setting. Only `work_dir` is switched by `persistent`

For how to reference these directories from tasks, see [Variables Available in Recipes]({{< relref "recipe-variables" >}}).

---

## spec.runtime

Declares the resource managed on the host. Required when using `stateChanges`.

> **\[kind: host only\]**

```yaml
spec:
  runtime:
    type: tool      # resource type (tool / service)
    name: ansible   # resource name
```

When `spec.runtime` is declared, `stateChanges` writes are automatically scoped to `/runtime/<type>/<name>`. See [Defining Host Configuration]({{< relref "host-configuration" >}}) for details.

---

## spec.allowedScope

Constrains which recipe `kind` can call this adapter.

> **\[kind: adapter only\] \[Required\]**

```yaml
spec:
  allowedScope: node   # node / cluster / infra / host / *
```

| Value | Description |
|-------|-------------|
| `node` / `cluster` / `infra` / `host` | Only callable from the specified `kind` |
| `*` | Callable from any `kind` |

`*` is appropriate for adapters that provide scope-agnostic operations such as tool installation or environment setup. If the adapter needs different behavior per scope, define a separate adapter for each scope.

---

## spec.executionUnit

Specifies the execution unit for the adapter.

> **\[kind: adapter only\]**

```yaml
spec:
  executionUnit: default   # or node
```

| Value | Description |
|-------|-------------|
| `default` (omitted) | The calling recipe's `kind` is used as the execution unit as-is |
| `node` | Only valid with `allowedScope: cluster`. Niwashi decomposes execution per node within the cluster |

---

## spec.commands

Defines commands exposed by the adapter. These correspond to the name specified in the `command` field of `tool.run`.

> **\[kind: adapter only\]**

```yaml
spec:
  commands:
    default:              # used when command is omitted
      task: run
    ansible-playbook:
      task: run-playbook
      description: "Run ansible-playbook"
```

| Field | Required | Description |
|-------|----------|-------------|
| `task` | Yes | Task name to invoke when this command is called |
| `description` | | Description of the command |

---

## spec.tasks

The list of tasks.

> **\[All kinds\]**

See [Defining Tasks]({{< relref "defining-tasks" >}}) for how to write tasks. For `kind: adapter`, only the `exec.local` action is available.

---

## Next Steps

- [Defining Tasks]({{< relref "defining-tasks" >}}) — Action types, dependsOn, and operation details
- [Updating State with stateChanges]({{< relref "state-changes" >}}) — Details on `stateChanges`
- [Defining Node Capabilities]({{< relref "node-capability" >}}) — How to write `kind: node` recipes
- [Defining Cluster Capabilities]({{< relref "cluster-capability" >}}) — How to write `kind: cluster` recipes
- [Defining Infrastructure Provisioners]({{< relref "infrastructure-provisioning" >}}) — How to write `kind: infra` recipes
- [Defining Host Configuration]({{< relref "host-configuration" >}}) — How to write `kind: host` recipes
- [Defining Adapters]({{< relref "defining-adapters" >}}) — How to write `kind: adapter` recipes

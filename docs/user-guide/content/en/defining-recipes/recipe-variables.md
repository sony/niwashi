---
title: "Variables Available in Recipes"
weight: 20
---

# Variables Available in Recipes

Inside recipe scripts (`scriptTpl`) and template fields (`argvTpl`, `envTpl`, etc.), you can reference information from the execution context as variables. There are two types of variables.

- **Template variables** — `{{ .Xxx }}` format. Used inside template fields such as `scriptTpl` and `argvTpl`.
- **Environment variables** — `$NWS_XXX` format. Automatically set when a script is executed.

---

## Template Variables

Available inside template fields (`scriptTpl`, `argvTpl`, `envTpl`, `workdirTpl`).

| Variable | Description | Example |
|----------|-------------|---------|
| `{{ .Target }}` | Name of the target node or cluster | `web-server-01` |
| `{{ .Params.xxx }}` | Recipe parameters. Default values from `spec.defaults.params`, overridden by values specified in State. For key naming rules, see [params Key Naming]({{< relref "/defining-desired-state/basic-concepts/naming-rules#params-key-naming" >}}) | `{{ .Params.version }}` |
| `{{ .Assets }}` | Recipe assets directory. Expanded to the node-side path for `exec.remote` | `/home/user/.niwashi/recipe/...` |
| `{{ .Paths.work_dir }}` | Task working directory. Node-side path for `exec.remote` | `/home/user/.niwashi/...` |
| `{{ .Paths.input_dir }}` | Task input directory | `/home/user/.niwashi/.../inputs/` |
| `{{ .Paths.output_dir }}` | Task output directory | `/home/user/.niwashi/.../outputs/` |
| `{{ .Paths.log_dir }}` | Task log directory | `/home/user/.niwashi/.../logs/` |
| `{{ .Paths.workspace }}` | Workspace root directory | `/home/user/.niwashi/` |
| `{{ .Outputs }}` | Shared data directory across tasks (see below) | `/path/to/outputs/` |
| `{{ .Store.xxx }}` | Value written to `store/<key>` by this recipe's own `stateChanges`. Returns an empty string if the key does not exist. See [Using the store/ Subpath]({{< relref "state-changes#using-the-store-subpath" >}}) | `{{ .Store.applied_version }}` |
| `{{ .Stores.<alias>.xxx }}` | Value stored in a required recipe's `store/` subpath, accessed via the alias defined by `requires.as`. See [spec.requires]({{< relref "recipe-spec#specrequires" >}}) | `{{ .Stores.upstream.endpoint }}` |
| `{{ .Runtime.tool.<name>.xxx }}` | Execution path of a tool recorded in State by an installer recipe | `{{ .Runtime.tool.ansible.path }}` |

### {{ .Assets }}

`{{ .Assets }}` is the directory where files listed in the recipe's `assets` are placed. For `exec.remote`, assets are copied to the node, so the variable is expanded to the node-side path.

```yaml
spec:
  assets:
    - bin/setup.sh
  tasks:
    - name: setup
      action:
        exec.remote:
          scriptTpl: |
            {{ .Assets }}/bin/setup.sh
```

For `kind: adapter` recipes, `{{ .Assets }}` refers to the adapter's own assets directory.

### {{ .Outputs }}

`.Outputs` is a shared directory for passing data between tasks within a recipe. Files placed here by one task can be referenced by subsequent tasks.

```yaml
tasks:
  - name: generate
    action:
      exec.local:
        scriptTpl: |
          echo "value=123" > {{ .Outputs }}/result.env

  - name: consume
    dependsOn: [generate]
    action:
      exec.local:
        scriptTpl: |
          source {{ .Outputs }}/result.env
          echo "Got: $value"
```

You can specify a subpath using dot notation. `{{ .Outputs.foo.bar }}` expands to `{{ .Outputs }}/foo/bar`.

For `exec.remote`, the contents of `{{ .Outputs }}` on the node side are automatically copied to the host-side `{{ .Outputs }}` after task execution.

---

## Environment Variables

Environment variables automatically set by nwsctl when a script is executed. They can be referenced as `$NWS_XXX` inside the script body of `scriptTpl`.

### Execution Control

| Variable | Description |
|----------|-------------|
| `NWS_DRY_RUN` | Dry-run mode (`off` / `simulate`) |
| `NWS_DRY_RUN_ENABLED` | `1` when dry-run is enabled, `0` otherwise |
| `NWS_LOG_LEVEL` | Log level |

### Execution Context

| Variable | Description |
|----------|-------------|
| `NWS_SCOPE` | Execution phase (`node` / `cluster` / `host`, etc.) |
| `NWS_TARGET_ID` | ID of the execution target. Normally corresponds to the phase (node ID for the node phase, cluster ID for the cluster phase, generator ID for the infra phase). However, even in the cluster phase, it becomes a node ID when tasks are broken down per node via `where` conditions or when an adapter sets `executionUnit: node` |

### Paths and Files

| Variable | Description |
|----------|-------------|
| `NWS_WORK_DIR` | Task working directory (same value as `{{ .Paths.work_dir }}`) |
| `NWS_LOG_DIR` | Task log directory (same value as `{{ .Paths.log_dir }}`) |
| `NWS_INPUT_DIR` | Task input directory (same value as `{{ .Paths.input_dir }}`) |
| `NWS_OUTPUT_DIR` | Task output directory (same value as `{{ .Paths.output_dir }}`) |
| `NWS_PARAMS` | Path to the parameters file (JSON) |
| `NWS_STATE` | Path to the State snapshot file (JSON) |

For where each directory is created inside the workspace, see [spec.workspace (Working Directory Structure)]({{< relref "recipe-spec#specworkspace" >}}).

### NWS_PARAMS

`NWS_PARAMS` is a file that records the parameters passed to the task in JSON format. You can read it from within a script using `jq` or similar tools.

```bash
VERSION=$(jq -r '.version' "${NWS_PARAMS}")
echo "Installing version ${VERSION}"
```

### NWS_STATE

`NWS_STATE` is the path to a file that records a snapshot of the State related to the execution target, in JSON format. Through this file, a recipe can look up the target nodes' connection details, OS, labels, and so on.

The file is structured as follows.

```json
{
  "host": {
    "os": "linux",
    "arch": "amd64"
  },
  "inventory": {
    "nodes": {
      "<node-name>": {
        "labels": { "role": "master" },
        "os": "linux",
        "arch": "amd64",
        "addresses": ["192.168.1.10"],
        "connection": { ... }
      }
    },
    "clusters": {
      "<cluster-name>": { ... }
    }
  },
  "runtime": { ... }
}
```

| Field | Description |
|-------|-------------|
| `host` | `os` / `arch` of the host running nwsctl |
| `inventory.nodes` | Information about the target nodes (labels, OS, architecture, addresses, connection details) |
| `inventory.clusters` | Information about the target clusters |
| `runtime` | Information about tools and services on the host (recorded by `kind: host` recipes) |

**What the snapshot contains depends on the execution target.**

| Recipe execution phase | Contents of `inventory` |
|-----------------------|-------------------------|
| node | The target node only |
| cluster | The target cluster and all of its member nodes (with cluster-specific labels merged) |
| host / infra | Empty (only `host` and `runtime` are available) |

In `kind: cluster` recipes, you can read the member node list and their addresses from `NWS_STATE`, for example to generate an inventory file.

```bash
# Example: list the addresses of the cluster's member nodes
jq -r '.inventory.nodes[].addresses[0]' "${NWS_STATE}"
```

**`NWS_PARAMS` vs `NWS_STATE`**: `NWS_PARAMS` holds "the values passed to this recipe via `params` in State", while `NWS_STATE` holds "a snapshot of the environment around the execution target". Receive recipe configuration through `params` (or `{{ .Params.* }}`), and read environment-derived information such as node connection details and labels from `NWS_STATE`.

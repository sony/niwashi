---
title: "Updating State with stateChanges"
weight: 5
---

# Updating State with stateChanges

`stateChanges` is the mechanism for updating Niwashi's State after a task runs. It is primarily used in the following recipes:

- `kind: infra` — Record connection information for provisioned instances
- `kind: host` — Record tool information on the host
- `kind: node` / `kind: cluster` — Capability state management

---

## Basic Syntax

`stateChanges` is written directly under a task as a key map. The keys are operation names (arbitrary), and multiple operations can be defined.

```yaml
tasks:
  - name: provision
    action:
      exec.local:
        scriptTpl: |
          #!/bin/bash
          # ... start instance and write connection info as JSON ...
    stateChanges:
      register-instance:          # operation name (arbitrary)
        op: set
        path: "instances/my-vm-0"
        valueFromFile: "{{ .Outputs }}/my-vm-0.json"
```

---

## Field Reference

| Field | Required | Description |
|-------|----------|-------------|
| `op` | Yes | Operation type. `set` (replace value) or `remove` (delete) |
| `path` | Yes | Path to write to (see below). Template variables supported |
| `value` | | Value to write (inline YAML) |
| `valueFromFile` | | Read value from a file. `.json`/`.yaml` files are loaded as structured data |
| `valueFromJson` | | Write value as a JSON string. Template variables supported |
| `count` | | Number of loop iterations (see below). Used with `{{ .Loop.index }}` |

`value`, `valueFromFile`, and `valueFromJson` are mutually exclusive. If none is specified, `null` is written.

`op: set` replaces the value at the specified path entirely. Any existing value is overwritten.

### valueFromFile Loading Format

The loading format depends on the file extension specified in `valueFromFile`.

| Extension | Loading Format |
|-----------|---------------|
| `.json`, `.jsonc` | Loaded as structured JSON data |
| `.yaml`, `.yml` | Loaded as structured YAML data |
| Other | Loaded as plain text |

---

## Writing paths by kind

The base of `path` depends on the recipe's `kind`. `path: "/"` or `path: ""` refers to the base itself.

### `kind: infra`

Provisioner recipes record connection information for created instances under `instances/<id>`.

```yaml
stateChanges:
  register-instance:
    op: set
    path: "instances/my-vm-0"
    valueFromFile: "{{ .Outputs }}/my-vm-0.json"
```

See [Defining Infrastructure Provisioners]({{< relref "infrastructure-provisioning" >}}) for the expected value format.

To register multiple instances at once, use `count` (see below).

The `store/` subpath can be used to pass information to node recipes (see [Using the store/ Subpath](#using-the-store-subpath)).

### `kind: host`

Installer recipes record tool information on the host. The base of `path` is determined by the `type` and `name` declared in `spec.runtime`. `path: "/"` refers to that base itself, so use it to write the entire tool information at once.

```yaml
# When spec.runtime: { type: tool, name: ansible } is declared
# → the base of path: "/" becomes /runtime/tool/ansible
stateChanges:
  record-ansible:
    op: set
    path: "/"
    valueFromFile: "{{ .Outputs }}/ansible_info.json"
```

The recorded value can be referenced by adapter recipes as `{{ .Runtime.tool.<name>.<key> }}`. See [Defining Host Configuration]({{< relref "host-configuration" >}}) for details.

### `kind: node` / `kind: cluster`

The main Capability state is managed automatically by Niwashi, so `stateChanges` can only write to the `store/` subpath (see [Using the store/ Subpath](#using-the-store-subpath)).

```yaml
stateChanges:
  save-endpoint:
    op: set
    path: "store/endpoint"
    valueFromFile: "{{ .Outputs }}/endpoint.json"
```

---

## Deletion (op: remove)

Combining `stateChanges` with an `operation: destruct` task allows the corresponding State entries to be removed when infrastructure is deleted.

```yaml
tasks:
  - name: teardown
    operation: destruct
    action:
      exec.local:
        scriptTpl: |
          #!/bin/bash
          vagrant destroy -f
    stateChanges:
      remove-from-state:
        op: remove
        path: "instances"
```

---

## Looping (count)

Specifying `count` repeats the same operation the given number of times. `{{ .Loop.index }}` expands to an index starting from 0. Used by infra Provisioners to register multiple instances at once.

```yaml
stateChanges:
  generate-instances:
    count: "{{ .Params.count }}"    # can be specified with a template variable
    op: set
    path: "instances/{{ .Params.prefix }}-{{ .Loop.index }}"
    valueFromFile: "{{ .Outputs.instances }}/{{ .Params.prefix }}-{{ .Loop.index }}.json"
```

---

## Using the store/ Subpath

The `store/` subpath is a free-use data storage area for recipes.

`store/` is kept when the capability is updated, and removed when a construct task fails. See [Update (operation: update)]({{< relref "defining-tasks#update-operation-update" >}}) and [When a Task Fails]({{< relref "defining-tasks#when-a-task-fails" >}}).

### Use case 1: Idempotency management

Save metadata such as "which version was applied" so the recipe can skip re-execution on subsequent runs. This pattern is available in all recipe kinds except `kind: adapter`.

Use `{{ .Store.<key> }}` in `scriptTpl` to read values saved in your recipe's own `store/`. When a key does not exist (e.g., on the first run), an empty string is returned. Use the `default` function to handle this case in shell conditionals.

```yaml
tasks:
  - name: apply
    action:
      exec.local:
        scriptTpl: |
          #!/bin/bash
          set -euo pipefail
          APPLIED="{{ .Store.applied_version | default "" }}"
          if [ "$APPLIED" = "{{ .Params.version }}" ]; then
            echo "already applied (version={{ .Params.version }}), skipping"
            exit 0
          fi
          # ... actual processing ...
    stateChanges:
      save-applied-version:
        op: set
        path: "store/applied_version"
        value: "{{ .Params.version }}"
```

> **Note**: Use underscores in `store/` key names. Keys containing hyphens (e.g., `applied-version`) cannot be accessed via the dot notation `{{ .Store.applied-version }}`.

### Use case 2: Passing data to downstream recipes

Save information collected or generated by one recipe to `store/` so downstream recipes can access it directly via `{{ .Stores.<alias>.<key> }}` in task templates.

The downstream recipe adds an `as:` field to the `requires` entry to create an alias. Niwashi makes the required recipe's `store/` data available under that alias name.

```yaml
# saving recipe
stateChanges:
  save-endpoint:
    op: set
    path: "store/endpoint"
    valueFromFile: "{{ .Outputs }}/endpoint.json"
```

```yaml
# downstream recipe — access via {{ .Stores.upstream.<key> }}
spec:
  requires:
    - name: some.capability
      as: upstream            # alias used in {{ .Stores.upstream }}
```

```yaml
tasks:
  - name: use-endpoint
    action:
      exec.local:
        scriptTpl: |
          ENDPOINT="{{ .Stores.upstream.endpoint }}"
          echo "Connecting to ${ENDPOINT}"
```

Execution order is guaranteed: the saving recipe always runs before the downstream recipe. This works across execution phases as well (e.g., from an `infra` recipe to a `node` recipe).

### Note for `kind: cluster`

Even when a `kind: cluster` recipe runs tasks per node via a `where` condition, `store/` writes are always aggregated into the cluster's capability path. If multiple nodes write to the same path, the last write wins. Use a node-identifying subkey to avoid collisions.

```yaml
stateChanges:
  save-node-info:
    op: set
    path: "store/nodes/{{ .Target }}"    # use node ID as key to avoid collisions
    valueFromFile: "{{ .Outputs }}/info.json"
```

---

## Example: The Ansible Adapter Pattern

The recipes in `recipe/ansible/` are a canonical example of installer + adapter + stateChanges.

1. **ansible-installer.yaml** (`kind: host`) — Detects Ansible and records its information via `stateChanges`
2. **ansible-adapter.yaml** (`kind: adapter`) — References `{{ .Runtime.tool.ansible.playbook.path }}` to execute commands

See [Defining Adapters]({{< relref "defining-adapters" >}}) for details.

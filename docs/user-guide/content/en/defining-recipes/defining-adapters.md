---
title: "Defining Adapters"
weight: 4
---

# Defining Adapters

This page explains how to write `kind: adapter` recipes (adapters). An adapter exposes commands that other recipes can invoke via the `tool.run` action.

For information on how to use adapters from a recipe, see [Using Adapters]({{< relref "using-adapters" >}}).

---

## What Is a `kind: adapter` Recipe?

A `kind: adapter` recipe **exposes commands that other recipes can invoke via `tool.run`**.

{{< mermaid >}}
flowchart LR
    subgraph host["Host (nwsctl)"]
        caller["Calling recipe<br/>(kind: node / cluster / infra / host)"]
        adapter["kind: adapter recipe<br/>(exposes commands)"]
        tool["Tool on the host<br/>(e.g. ansible-playbook)"]
    end
    caller -->|"tool.run (toolRef / command)"| adapter
    adapter -->|"runs the tool via the recorded path"| tool
{{< /mermaid >}}

An adapter is typically defined as a pair of two recipes.

| Recipe | `kind` | Role |
|--------|--------|------|
| **installer recipe** | `host` | Checks whether the tool exists on the host and records its execution path in State |
| **adapter recipe** | `adapter` | Exposes commands that invoke the tool using the recorded execution path |

The installer provides the dependency target referenced in `spec.requires` (e.g. `host.tool.ansible`), and the adapter is called via `tool.run` after that dependency is satisfied.

---

## Field Reference

### Common Fields

See [Recipe Format Reference]({{< relref "recipe-spec" >}}) for details on all `spec` fields, including `kind: adapter`-specific fields (`spec.allowedScope`, `spec.executionUnit`, `spec.commands`).

See [Defining Tasks]({{< relref "defining-tasks" >}}) for how to write tasks. For `kind: adapter`, tasks can only use `exec.local`.

### Template Variables

In `kind: adapter` tasks, `{{ .Target }}` expands to the execution target ID of the calling recipe (e.g. nodeID). You can reference the tool path recorded by the installer recipe via `{{ .Runtime.tool.<name>.<field> }}`.

For the full list of template variables and runtime environment variables, see [Variables Available in Recipes]({{< relref "recipe-variables" >}}).

---

## Minimal Example

```yaml
version: nws.recipe/v1
kind: adapter
metadata:
  id: my-org/my-tool.adapter
  version: "1.0.0"
  description: "Adapter for my-tool"
spec:
  provides:
    - name: adapter.tool.my-tool   # name used in tool.run's toolRef

  allowedScope: node                # constrains the kind of calling recipe (required)

  commands:
    default:
      task: run
    run:
      task: run
      description: "Run my-tool"

  tasks:
    - name: run
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            exec {{ .Runtime.tool.my_tool.path }} "$@"
```

---

## Pattern: Using Installer and Adapter as a Pair

For the adapter to reference the tool path via `{{ .Runtime.tool.<name>.path }}`, the installer recipe must have recorded that path in State.

**Installer recipe (`kind: host`)**

```yaml
spec:
  provides:
    - name: host.tool.my-tool

  runtime:
    type: tool
    name: my_tool   # used as a template map key, so hyphens/dots are not allowed here

  tasks:
    - name: check
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            BIN="$(command -v my-tool || { echo 'my-tool not found' >&2; exit 1; })"
            jq -n --arg path "$BIN" '{ path: $path }' \
              > "{{ .Outputs }}/tool_info.json"
      stateChanges:
        record-path:
          op: set
          path: "/"
          valueFromFile: "{{ .Outputs }}/tool_info.json"
```

When the installer recipe runs, `{ "path": "/usr/bin/my-tool" }` is written to `/runtime/tool/my_tool` in State.

**Referencing from the adapter recipe**

```yaml
spec:
  requires:
    - host.tool.my-tool   # declare installer recipe as a dependency

  tasks:
    - name: run
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            exec {{ .Runtime.tool.my_tool.path }} "$@"
```

> **Important**: Declaring `requires` on the adapter recipe alone does not schedule the installer job. The adapter's tasks run inline inside the *caller's* job rather than as a separately scheduled job, so the adapter's own `requires` is never evaluated. The recipe that actually invokes `tool.run` — here, the caller — must also declare `requires: [host.tool.<name>]` itself, or `{{ .Runtime.tool.<name>.path }}` will be unresolved when the adapter task runs. See `examples/recipe/adapter/` in the repository for a working example with both declarations.

---

## Example: The Ansible Adapter

See the `recipe/ansible/` directory for a real-world adapter implementation example.

- `ansible-installer.yaml` — `kind: host` recipe. Detects Ansible and records its path under `/runtime/tool/ansible` via `stateChanges`
- `ansible-adapter.yaml` — `kind: adapter` recipe. References `{{ .Runtime.tool.ansible.playbook.path }}` to execute commands

---

## Next Steps

- [Using Adapters]({{< relref "using-adapters" >}}) — How to call a defined adapter from a recipe
- [Recipe Format Reference]({{< relref "recipe-spec" >}}) — Details on `allowedScope`, `executionUnit`, and `commands`

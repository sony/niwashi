---
title: "Defining Host Configuration"
weight: 7
---

# Defining Host Configuration

This page explains how to write recipes with `kind: host` that run on the host machine executing nwsctl.

---

## What Is a `kind: host` Recipe?

A `kind: host` recipe **executes operations on the host machine running nwsctl**.

{{< mermaid >}}
flowchart LR
    subgraph host["Host (nwsctl)"]
        other["Other recipes<br/>(declared via spec.requires)"]
        recipe["kind: host recipe (installer)"]
        tool["Tool on the host (e.g. Ansible)"]
        state["State (/runtime/tool/...)"]
    end
    other -->|"triggers execution as a dependency"| recipe
    recipe -->|"detect"| tool
    recipe -->|"record path etc."| state
{{< /mermaid >}}

A `kind: host` recipe cannot be specified directly in the `capabilities` of a State file. It is executed **only when declared as a dependency in another recipe's `spec.requires`**.

```yaml
spec:
  requires:
    - host.tool.ansible   # this causes the kind: host installer recipe to be called
```

The planner resolves this dependency and executes the installer recipe before the recipe that depends on it.

The primary current use case is **installer recipes**. These record the execution path of tools installed on the host (Ansible, git, etc.) in State, so that adapter recipes can reference those paths.

```
kind: host installer recipe
  → records tool path in /runtime/tool/<name> via stateChanges
      → kind: adapter recipe references it as {{ .Runtime.tool.<name>.path }}
```

---

## Field Reference

### Common Fields

See [Recipe Format Reference]({{< relref "recipe-spec" >}}) for details on common `spec` fields. For `kind: host`, `spec.requires` can only declare dependencies on `host` recipes.

See [Defining Tasks]({{< relref "defining-tasks" >}}) for how to write tasks, and [Updating State with stateChanges]({{< relref "state-changes" >}}) for updating State after a task runs.

### spec.runtime

When using `stateChanges`, declare the resource managed on the host via `spec.runtime`.

```yaml
spec:
  runtime:
    type: tool      # resource type (tool / service)
    name: ansible   # resource name (alphanumeric and underscores; must start with a letter or underscore)
```

| Field | Required | Description |
|-------|----------|-------------|
| `type` | Yes | `tool` (CLI tool on the host) or `service` (local service) |
| `name` | Yes | Resource name. Only alphanumeric characters and underscores are allowed (must start with a letter or underscore). This name is used as a Go template map key (`{{ .Runtime.tool.<name>... }}`), so hyphens and dots are not supported. |

Declaring `spec.runtime` automatically restricts `stateChanges` writes to `/runtime/<type>/<name>` and below. Can be omitted if `stateChanges` is not used.

### Template Variables and `where` Specification

In `kind: host` tasks, `{{ .Target }}` is an empty string. When using `where`, the evaluation context is the host's own `os` and `arch` (not the node's information).

For the full list of template variables and runtime environment variables, see [Variables Available in Recipes]({{< relref "recipe-variables" >}}). For details on the `where` field, see [Task Filtering Condition (where)]({{< relref "task-where" >}}).

### Data Structure Written to State

Values written to State via `stateChanges` are stored under `/runtime/<type>/<name>`. The structure is free-form, but the following conventions are used by the reference installer recipes.

**For `type: tool`**

A JSON object with at least a `path` field. Tools with multiple executables use nested objects.

```json
{ "path": "/usr/local/bin/my-tool", "version": "1.2.3" }
```

```json
{
  "playbook": { "path": "/usr/local/bin/ansible-playbook", "version": "2.16.3" },
  "galaxy":   { "path": "/usr/local/bin/ansible-galaxy" }
}
```

The recorded value can be referenced in adapter recipes as `{{ .Runtime.tool.<name>.<field> }}`.

**For `type: service`**

There is no fixed convention. Store whatever connection or configuration information the consuming recipe needs.

---

## Minimal Example

```yaml
version: nws.recipe/v1
kind: host
metadata:
  id: my-org/my-tool.installer
  version: "1.0.0"
  description: "Check my-tool installation"
spec:
  provides:
    - name: host.tool.my-tool

  runtime:
    type: tool
    name: my-tool

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

---

## Pattern: Installer Recipes

### Detecting a Tool and Recording Its Path

Detect a tool installed on the host and write its execution path and version information to State.

```yaml
tasks:
  - name: check
    action:
      exec.local:
        scriptTpl: |
          #!/bin/bash
          set -euo pipefail
          BIN="$(command -v my-tool || { echo 'my-tool not found in PATH' >&2; exit 1; })"
          VER="$(my-tool --version | head -n1)"
          jq -n \
            --arg version "$VER" \
            --arg path "$BIN" \
            '{ version: $version, path: $path }' \
            > "{{ .Outputs }}/tool_info.json"
    stateChanges:
      record-path:
        op: set
        path: "/"
        valueFromFile: "{{ .Outputs }}/tool_info.json"
```

### Version Requirement Check

An example that accepts a version requirement via `spec.defaults.params` and checks it.

```yaml
defaults:
  params:
    min_version: ""   # no version check when omitted

tasks:
  - name: check
    action:
      exec.local:
        scriptTpl: |
          #!/bin/bash
          set -euo pipefail
          BIN="$(command -v my-tool || { echo 'my-tool not found' >&2; exit 1; })"
          VER="$(my-tool --version | grep -oP '[0-9]+\.[0-9]+\.[0-9]+')"
          if [ -n "{{ .Params.min_version }}" ]; then
            case "$VER" in
              {{ .Params.min_version }}* ) : ;;
              * ) echo "my-tool version $VER does not meet minimum {{ .Params.min_version }}" >&2; exit 1 ;;
            esac
          fi
          jq -n --arg path "$BIN" --arg version "$VER" \
            '{ path: $path, version: $version }' \
            > "{{ .Outputs }}/tool_info.json"
    stateChanges:
      record-path:
        op: set
        path: "/"
        valueFromFile: "{{ .Outputs }}/tool_info.json"
```

---

## Example: Reference Installer Recipes

Refer to the reference `kind: host` recipes under `recipe/`.

| File | `provides` | Description |
|------|-----------|-------------|
| `recipe/ansible/ansible-installer.yaml` | `host.tool.ansible` | Detects ansible-playbook / ansible-galaxy and records their paths |
| `recipe/nws/tools/git.yaml` | `host.tool.git` | Detects git and records its path |
| `recipe/nws/tools/jq.yaml` | `host.tool.jq` | Detects jq and records its path |

---

## Next Steps

- [Updating State with stateChanges]({{< relref "state-changes" >}}) — Details on `stateChanges`
- [Task Filtering Condition (where)]({{< relref "task-where" >}}) — Details on the `where` field

---
title: "Defining Node Capabilities"
weight: 1
---

# Defining Node Capabilities

This page explains how to write Capability recipes with `kind: node` that are applied to nodes.

---

## What Is a `kind: node` Recipe?

A `kind: node` recipe **executes operations against a single node**. Because it runs independently per node, it is well suited for node-specific configuration and software installation.

{{< mermaid >}}
flowchart LR
    subgraph host["Host (nwsctl)"]
        recipe["kind: node recipe"]
    end
    subgraph targets["Targets"]
        n1["node-1"]
        n2["node-2"]
    end
    recipe -->|"executed independently per node"| n1
    recipe --> n2
{{< /mermaid >}}

When a user specifies a Capability in a State file, Niwashi finds the matching recipe and executes its tasks against the target node.

---

## Field Reference

### Common Fields

See [Recipe Format Reference]({{< relref "recipe-spec" >}}) for details on common `spec` fields. For `kind: node`, `spec.requires` can declare dependencies on `host` and `node` recipes.

See [Defining Tasks]({{< relref "defining-tasks" >}}) for how to write tasks, and [Updating State with stateChanges]({{< relref "state-changes" >}}) for updating State after a task runs.

### Template Variables

In `kind: node` tasks, `{{ .Target }}` expands to the **ID of the target node**. This is useful for identifying the node in scripts or as a key in `store/` paths.

For the full list of template variables and runtime environment variables, see [Variables Available in Recipes]({{< relref "recipe-variables" >}}).

---

## Minimal Example

```yaml
version: nws.recipe/v1
kind: node
metadata:
  id: my-org/my-feature
  version: "1.0.0"
  description: "My feature recipe for nodes"
spec:
  provides:
    - name: my-org.my-feature
  tasks:
    - name: install
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            echo "Installing my-feature on {{ .Target }}"
```

---

## Example: Installing and Removing a Package

An example that connects to a node via SSH using `exec.remote` and installs or removes an apt package.

```yaml
version: nws.recipe/v1
kind: node
metadata:
  id: my-org/nginx
  version: "1.0.0"
  description: "Install nginx"
spec:
  provides:
    - name: my-org.nginx

  defaults:
    params:
      version: ""   # empty means install the latest version

  tasks:
    - name: install
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            PKG="nginx{{ if .Params.version }}={{ .Params.version }}{{ end }}"
            apt-get install -y "$PKG"

    - name: uninstall
      operation: destruct
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            apt-get purge -y nginx
            apt-get autoremove -y
```

---

## Next Steps

- [Defining Tasks]({{< relref "defining-tasks" >}}) — Action types, dependsOn, and operation details
- [Updating State with stateChanges]({{< relref "state-changes" >}}) — Writing to store/ and idempotency management
- [Using Adapters]({{< relref "using-adapters" >}}) — Details on calling adapters with `tool.run`

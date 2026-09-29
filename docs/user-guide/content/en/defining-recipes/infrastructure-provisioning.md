---
title: "Defining Infrastructure Provisioners"
weight: 2
---

# Defining Infrastructure Provisioners

This page explains how to write `kind: infra` recipes (Provisioners). A Provisioner creates and destroys VMs or instances and records their connection information in State.

---

## What Is a `kind: infra` Recipe?

A `kind: infra` recipe **generates infrastructure (VMs / instances)**. It operates as the `provisioner` referenced by `infrastructure.generators` in State.

{{< mermaid >}}
flowchart LR
    subgraph host["Host (nwsctl)"]
        recipe["kind: infra recipe<br/>(exec.local only)"]
        state["State"]
    end
    subgraph infra["Infrastructure"]
        i1["instance-1"]
        i2["instance-2"]
    end
    recipe -->|"create / destroy"| infra
    recipe -->|"record connection info (stateChanges)"| state
{{< /mermaid >}}

Primary responsibilities of a Provisioner:

1. Start VMs or instances
2. Record connection information for created instances (IP address, SSH private key, etc.) in State via `stateChanges`

Using the connection information recorded in State, Niwashi automatically establishes `exec.remote` connections to nodes.

> **Note**: Provisioners use only `exec.local`. Because nodes are not yet reachable at this stage, `exec.remote` cannot be used.

---

## Field Reference

### Common Fields

See [Recipe Format Reference]({{< relref "recipe-spec" >}}) for details on common `spec` fields. For `kind: infra`, `spec.requires` can declare dependencies on `host` and `infra` recipes.

See [Defining Tasks]({{< relref "defining-tasks" >}}) for how to write tasks, and [Updating State with stateChanges]({{< relref "state-changes" >}}) for updating State after a task runs.

Fields specific to `kind: infra`:

| Field | Description |
|-------|-------------|
| `spec.provides` | Name referenced by the generator's `provisioner` field (e.g. `infra.vm`) |
| `spec.workspace.mode` | `persistent` is recommended to retain VM files and the workspace |

### Template Variables

In `kind: infra` tasks, `{{ .Target }}` expands to the **Generator name** (the key under `infrastructure.generators` in State).

```yaml
infrastructure:
  generators:
    my-vms:             # <- this becomes {{ .Target }}
      provisioner: infra.vm.driver=vagrant
```

| Variable | Description |
|----------|-------------|
| `{{ .Target }}` | Generator name |
| `{{ .Params.xxx }}` | Parameters specified under the generator's `params` |
| `{{ .Loop.index }}` | Loop index (used with `count`) |

For the full list of template variables and runtime environment variables, see [Variables Available in Recipes]({{< relref "recipe-variables" >}}).

### Instance Information Schema

The JSON structure written to State via `stateChanges` for instance information is as follows.

```json
{
  "connection": {
    "ssh": {
      "address": {
        "host": "192.168.56.10",
        "port": 22,
        "user": "vagrant"
      },
      "auth": {
        "method": "privateKey",
        "privateKeyPath": "/path/to/private_key"
      },
      "hostKey": {
        "knownHostsPath": "/path/to/known_hosts"
      }
    }
  },
  "addresses": ["192.168.56.10"]
}
```

Write this JSON to a file for each instance, then register it under `instances/<instance-name>` via `stateChanges`.

### stateChanges Patterns

**Single instance**

```yaml
stateChanges:
  register-instance:
    op: set
    path: "instances/my-instance"
    valueFromFile: "{{ .Outputs }}/instance.json"
```

**Multiple instances (`count` + `{{ .Loop.index }}`)**

```yaml
stateChanges:
  register-instances:
    count: "{{ .Params.count }}"
    op: set
    path: "instances/{{ .Params.prefix }}-{{ .Loop.index }}"
    valueFromFile: "{{ .Outputs.instances }}/{{ .Params.prefix }}-{{ .Loop.index }}.json"
```

**On deletion (`operation: destruct`)**

```yaml
stateChanges:
  remove-from-state:
    op: remove
    path: "instances"
```

### When Provisioning Fails

If a construct task fails, the Generator entry is removed from State, including the instances already registered by `stateChanges`, and the next `nwsctl plan` / `nwsctl apply` runs the provisioning again. Write the construct tasks so that they converge when they start from partially created resources (for example, reuse a VM that already exists instead of failing).

If you decide not to retry, remove the partially created resources manually: they are no longer recorded in State, so `nwsctl plan --destroy` does not include them.

If a destruct task fails, the Generator entry is kept in State, and the next `--destroy` / `--prune` run destructs it again.

---

## Minimal Example

```yaml
version: nws.recipe/v1
kind: infra
metadata:
  id: my-org/my-provisioner
  version: "1.0.0"
  description: "Provision VMs"
spec:
  provides:
    - name: infra.my-provisioner

  workspace:
    mode: persistent

  defaults:
    params:
      count: 1
      prefix: "vm"

  tasks:
    - name: provision
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            mkdir -p "{{ .Outputs.instances }}"

            for i in $(seq 0 $(({{ .Params.count }} - 1))); do
              VM_NAME="{{ .Params.prefix }}-${i}"
              # ... start VM and obtain connection information ...
              jq -n \
                --arg host "192.168.56.$((10 + i))" \
                --arg user "ubuntu" \
                --arg key "/path/to/key" \
                '{
                  connection: {
                    ssh: {
                      address: { host: $host, port: 22, user: $user },
                      auth: { method: "privateKey", privateKeyPath: $key }
                    }
                  },
                  addresses: [$host]
                }' > "{{ .Outputs.instances }}/${VM_NAME}.json"
            done
      stateChanges:
        register:
          count: "{{ .Params.count }}"
          op: set
          path: "instances/{{ .Params.prefix }}-{{ .Loop.index }}"
          valueFromFile: "{{ .Outputs.instances }}/{{ .Params.prefix }}-{{ .Loop.index }}.json"

    - name: teardown
      operation: destruct
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            # ... destroy VMs ...
      stateChanges:
        remove:
          op: remove
          path: "instances"
```

---

## Example: Vagrant Provisioner

`recipe/vagrant/nws-recipe.yaml` is a real-world Provisioner implementation example.

```yaml
version: nws.recipe/v1
kind: infra
metadata:
  id: nws/vagrant.vm-cluster
  version: "1.0.0"
spec:
  provides:
    - name: "infra.vm"
      attrs:
        driver: "vagrant"

  workspace:
    mode: persistent

  defaults:
    params:
      count: 1
      prefix: "nws-vm"
      box: "bento/ubuntu-22.04"
      memory: "1024"
      cpus: "1"

  assets:
    - "bin/create-config.sh"      # parses vagrant ssh-config and generates JSON
    - "bin/cleanup-vagrant.sh"

  tasks:
    - name: prepare-vagrantfile
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            if [ -f Vagrantfile ]; then exit 0; fi
            # ... generate Vagrantfile ...

    - name: provision-and-extract
      dependsOn: [prepare-vagrantfile]
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            mkdir -p "{{ .Outputs.instances }}"
            vagrant up
            vagrant ssh-config > "{{ .Outputs.all_ssh_config }}"

            for i in $(seq 0 $(({{ .Params.count }} - 1))); do
              VM_NAME="{{ .Params.prefix }}-${i}"
              {{ .Assets }}/bin/create-config.sh \
                "{{ .Outputs.all_ssh_config }}" \
                "$VM_NAME" \
                "{{ .Outputs.instances }}/${VM_NAME}.json"
            done
      stateChanges:
        generate-instances:
          count: "{{ .Params.count }}"
          op: set
          path: "instances/{{ .Params.prefix }}-{{ .Loop.index }}"
          valueFromFile: "{{ .Outputs.instances }}/{{ .Params.prefix }}-{{ .Loop.index }}.json"

    - name: teardown-instance
      operation: destruct
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            vagrant destroy -f
      stateChanges:
        remove-instance-from-state:
          op: remove
          path: "instances"
```

---

## Next Steps

- [Updating State with stateChanges]({{< relref "state-changes" >}}) — Details on `stateChanges`
- [Defining Tasks]({{< relref "defining-tasks" >}}) — Action types, dependsOn, and operation details

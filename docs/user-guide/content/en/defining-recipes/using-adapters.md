---
title: "Using Adapters"
weight: 3
---

# Using Adapters

Adapters are a mechanism for calling external tools such as Ansible or Terraform from Niwashi recipes. This page explains how to use existing adapters from a recipe.

For information on how to define your own adapters, see [Defining Adapters]({{< relref "defining-adapters" >}}).

---

## What Is an Adapter?

An adapter is a `kind: adapter` special recipe that exposes commands wrapping an external tool. A recipe calls those commands using the `tool.run` action.

Adapters publish their Capability name via `provides` (e.g. `adapter.tool.ansible`). A recipe declares that name in `spec.requires` and references it in the `toolRef` field of `tool.run`.

---

## Declaring a Dependency on an Adapter

Adapters themselves are **not** listed in `spec.requires`. An adapter is resolved automatically when referenced via `tool.run.toolRef` — no explicit declaration is needed.

`spec.requires` is for host tools (`host.tool.*`) or other recipe capabilities that the recipe depends on.

```yaml
spec:
  requires:
    - host.tool.git   # host tool dependencies go in requires
    # adapter.tool.ansible is NOT listed here → reference it via tool.run.toolRef
```

---

## The tool.run Action

Use the `tool.run` action to invoke adapter functionality.

```yaml
action:
  tool.run:
    toolRef: adapter.tool.ansible   # Capability name of the adapter
    command: ansible-playbook        # Command name defined by the adapter
    argvTpl:                         # Arguments to pass (template variables supported)
      - "-i"
      - "hosts.ini"
      - "playbook.yml"
    workdirTpl: "{{ .Paths.work_dir }}"  # Working directory (defaults to the default working directory)
    envTpl:                              # Additional environment variables (optional)
      MY_VAR: "{{ .Params.my_var }}"
    inheritEnv:                          # Host environment variables to inherit (optional)
      - SSH_AUTH_SOCK
      - ANSIBLE_*
```

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `toolRef` | Yes | Capability name of the adapter |
| `command` | Yes | Command name to invoke (as defined by the adapter) |
| `argvTpl` | | List of arguments to pass. Template variables supported |
| `workdirTpl` | | Working directory. Defaults to the default working directory |
| `envTpl` | | Additional environment variables. Template variables supported |
| `inheritEnv` | | List of host environment variable names to inherit. Glob patterns supported |

### Environment Variable Inheritance (`inheritEnv`)

`inheritEnv` allows you to pass environment variables from the host running nwsctl into the adapter's task. This is useful when calling tools that depend on host-side variables such as `SSH_AUTH_SOCK` or `ANSIBLE_*`.

```yaml
action:
  tool.run:
    toolRef: adapter.tool.python.venv
    command: run
    argvTpl:
      - "{{ .Paths.work_dir }}/venv"
      - ansible-playbook
      - site.yml
    inheritEnv:
      - SSH_AUTH_SOCK   # SSH agent forwarding
      - ANSIBLE_*       # Ansible configuration variables
```

**Behavior details:**

- `PATH` is always inherited regardless of `inheritEnv`.
- Glob patterns such as `ANSIBLE_*` are supported.
- If the adapter's `exec.local` task also defines `inheritEnv`, both lists are merged.
- When the same key appears in both `envTpl` and `inheritEnv`, the `envTpl` value takes precedence.

> **Note:** `inheritEnv` only takes effect when the adapter is implemented with `exec.local`. Adapters implemented with `exec.remote` are not affected.

---

## Reference Adapters

| Adapter | Capability Name | Description |
|---------|----------------|-------------|
| Ansible | `adapter.tool.ansible` | Runs ansible-playbook / ansible-galaxy |

For details on each adapter, see the [Recipe Reference]({{< relref "../recipes" >}}).

- [Ansible Adapter]({{< relref "../recipes/ansible" >}})

---

## Next Steps

- [Defining Adapters]({{< relref "defining-adapters" >}}) - How to create your own adapter
- [Defining Node Capabilities]({{< relref "node-capability" >}}) - Example recipes that use `tool.run`

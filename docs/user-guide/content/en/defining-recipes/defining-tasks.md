---
title: "Defining Tasks"
weight: 9
---

# Defining Tasks

This page describes the common task definition format shared across all recipe `kind` values.

---

## Task Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Task name (unique within the recipe) |
| `action` | Yes | Action to execute (see below) |
| `operation` | | `construct` (default) / `update` / `destruct` |
| `dependsOn` | | List of task names this task depends on |
| `where` | | Filter condition for the target node |
| `stateChanges` | | State change operations to apply after the task runs |

---

## Action Types

### exec.remote — Run a script on a node

Connects to a node via remote connection (e.g., SSH) and executes a script. Connection information is configured automatically by nwsctl. Files specified in `assets` are also copied to the node automatically.

```yaml
action:
  exec.remote:
    scriptTpl: |
      #!/bin/bash
      set -euo pipefail
      VERSION="{{ .Params.version }}"
      echo "Installing version ${VERSION} on {{ .Target }}"
      cp "{{ .Assets }}/config/default.conf" /etc/myapp/
    envTpl:
      MY_VAR: "{{ .Params.my_var }}"   # additional environment variables
```

### exec.local — Run a script on the host

Executes a script on the host running nwsctl. Does not connect directly to the node.

```yaml
action:
  exec.local:
    scriptTpl: |
      #!/bin/bash
      set -euo pipefail
      echo "Preparing for {{ .Target }}"
    envTpl:
      MY_VAR: "{{ .Params.my_var }}"   # additional environment variables
    inheritEnv:
      - KUBECONFIG        # inherit a specific variable from the host environment
      - ANSIBLE_*         # glob patterns are supported
```

**Environment variable inheritance (`inheritEnv`)**

By default, `exec.local` inherits only `PATH` from the host environment. Use `inheritEnv` to explicitly specify additional variables to inherit.

| Field | Description |
|-------|-------------|
| `inheritEnv` | List of host environment variable names (or glob patterns) to inherit. `PATH` is always inherited. |

- Glob patterns such as `ANSIBLE_*` are supported.
- Variables set in `envTpl` take precedence over inherited values when the same key exists in both.

### tool.run — Execute via an adapter

Calls an adapter declared as a dependency in `spec.requires`. Used, for example, to run an Ansible playbook through the Ansible adapter. See [Using Adapters]({{< relref "using-adapters" >}}) for details.

```yaml
action:
  tool.run:
    toolRef: adapter.tool.ansible
    command: ansible-playbook
    argvTpl:
      - "-i"
      - "hosts.ini"
      - "playbook.yml"
    workdirTpl: "{{ .Paths.work_dir }}"
    inheritEnv:           # inherit host environment variables (exec.local-based adapters only)
      - SSH_AUTH_SOCK
      - ANSIBLE_*
```

> **\[kind: adapter only\]** In `kind: adapter` tasks, only `exec.local` is available. `exec.remote` and `tool.run` cannot be used.

---

## Writing Scripts (scriptTpl)

The script you write in `scriptTpl` is rendered as a template, saved to a file, and then executed.

### Controlling the Interpreter with a Shebang

If the first line of the script is a shebang (a line starting with `#!`), the script is executed with that interpreter. The shebang is **interpreted by niwashi itself**, not by the OS kernel: niwashi launches `<interpreter> [args] <script-file>`. The script file therefore does not need execute permission.

```yaml
action:
  exec.remote:
    scriptTpl: |
      #!/bin/bash
      set -euo pipefail
      echo "runs with bash"
```

If the shebang is omitted, a default interpreter is chosen based on the OS of the execution target.

| Target OS | Default interpreter |
|-----------|--------------------|
| Linux / macOS | `/bin/sh` |
| Windows | `powershell` (Windows PowerShell) |

The "execution target" is the host running nwsctl for `exec.local`, and the connected node for `exec.remote`.

The default interpreter on Windows is Windows PowerShell (`powershell`), which ships with every supported Windows version. To use PowerShell 7+ (`pwsh`), which requires a separate install, specify it explicitly via a shebang (e.g., `#!pwsh`).

> **Note**: Running a PowerShell script on Windows requires the execution policy on the machine that actually runs it (the host running nwsctl for `exec.local`, or the node connected over WinRM for `exec.remote`) to permit script execution. Under the default execution policy (`Restricted`), running an unsigned local script is blocked and fails with an `UnauthorizedAccess` error. Change the execution policy on that machine beforehand (e.g., `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`); see [about_Execution_Policies](https://go.microsoft.com/fwlink/?LinkID=135170) for details. niwashi itself never changes the execution policy.

### Recommendations

- **Write an explicit shebang** rather than relying on the implicit default interpreter
- For Linux / macOS, use `sh` or `bash`. If you rely on bash-specific features, state `#!/bin/bash` explicitly
- For Windows, write the script as a PowerShell script. When the shebang is omitted, it runs with `powershell` (Windows PowerShell) automatically
- Start bash scripts with `set -euo pipefail` so that failures of intermediate commands are detected
- To vary behavior per OS, split the work into separate tasks filtered with `where` instead of branching inside one script ([Task Filtering Condition (where)]({{< relref "task-where" >}}))

> **Note**: The shebang line is truncated at 127 characters, matching the Linux kernel limit.

The rendered script is saved in the task's log directory (`{{ .Paths.log_dir }}`), so you can inspect the actual script content after execution.

---

## Template Variables and Environment Variables

For a list of template variables available in `scriptTpl`, `argvTpl`, `workdirTpl`, and `envTpl`, and environment variables automatically set at script execution, see [Variables Available in Recipes]({{< relref "recipe-variables" >}}).

---

## Task Dependencies

Declaring a dependency with `dependsOn` causes the task to run only after the specified tasks have completed.

```yaml
tasks:
  - name: prepare
    action:
      exec.remote:
        scriptTpl: |
          mkdir -p /opt/myapp

  - name: install
    dependsOn: [prepare]
    action:
      exec.remote:
        scriptTpl: |
          cp "{{ .Assets }}/scripts/install.sh" /opt/myapp/
          /opt/myapp/install.sh

  - name: configure
    dependsOn: [install]
    action:
      exec.remote:
        scriptTpl: |
          /opt/myapp/configure.sh --version={{ .Params.version }}
```

---

## Update (operation: update)

When a capability has already been applied and its `params` change in the desired state, or the recipe version resolved for it moves forward, `nwsctl plan` detects the change and schedules the tasks with `operation: update`. No extra flag is needed: a normal `nwsctl plan` / `nwsctl apply` picks up updates together with additions.

```yaml
tasks:
  - name: install
    action:
      exec.remote:
        scriptTpl: |
          apt-get install -y myapp

  - name: reconfigure
    operation: update
    action:
      exec.remote:
        scriptTpl: |
          #!/bin/bash
          set -euo pipefail
          /opt/myapp/configure.sh --port={{ .Params.port }}
          systemctl restart myapp
```

### When update tasks run

| Change since the last apply | Result |
|---|---|
| `params` changed (same recipe version) | The `operation: update` tasks run |
| Recipe version upgraded (e.g. `1.0.0` → `1.1.0`) | The `operation: update` tasks of the new version run, even if `params` did not change |
| Recipe version downgraded | `nwsctl plan` fails with an error |
| No change | Nothing runs |

After the update tasks succeed, the new `params` and recipe version are recorded in State. The capability's `store/` is kept across updates, so update tasks can read values written by the construct tasks through `{{ .Store }}`.

### Recipes without update tasks

If the recipe has no `operation: update` task, the change cannot be applied. `nwsctl plan` does not schedule any task for it and lists it under `Pending Updates` instead:

```
Pending Updates (recipe has no update tasks):
- node:my-org/myapp@1.0.0:web-1
```

State keeps the previously applied `params` and version, so the change stays pending. Once the recipe provides an `operation: update` task, the next `nwsctl plan` picks the change up. Niwashi does not fall back to destroying and reconstructing the capability.

### Notes

- Updates are detected only for `kind: node` and `kind: cluster` capabilities.
- `params` are compared as a whole. `{{ .Params }}` in an update task holds the complete new params, merged with `spec.defaults.params`.
- Only the `params` written in the desired state (including profile overrides) are compared. Changing `spec.defaults.params` in a recipe does not trigger an update unless you also raise `metadata.version`.
- Capabilities that depend on the updated capability through `requires` are not re-run.

---

## Removal (operation: destruct)

When `plan --destroy` is run, tasks with `operation: destruct` are executed. Write the reverse of installation and configuration here.

```yaml
tasks:
  - name: install
    operation: construct   # default, can be omitted
    action:
      exec.remote:
        scriptTpl: |
          apt-get install -y myapp

  - name: uninstall
    operation: destruct
    action:
      exec.remote:
        scriptTpl: |
          apt-get remove -y myapp
          rm -rf /etc/myapp
```

`construct` and `destruct` are defined as separate tasks.

---

## When a Task Fails

When a task fails, `nwsctl apply` stops starting new jobs, waits for the tasks that are already running, and exits with an error. What happens to State depends on the operation:

| Operation | State after the failure | Next run |
|---|---|---|
| `construct` | The capability is removed from State, including its `store/` | The next `nwsctl plan` / `nwsctl apply` constructs it again |
| `update` | The previously applied `params` and version are kept | The next `nwsctl plan` / `nwsctl apply` runs the update again |
| `destruct` | The capability is kept in State | The next `--destroy` / `--prune` run destructs it again |

This table describes `kind: node` and `kind: cluster` capabilities. For `kind: infra`, see [Defining Infrastructure Provisioners]({{< relref "infrastructure-provisioning#when-provisioning-fails" >}}).

Because a failed construct is retried from scratch on the next run, write construct tasks so that they converge even when they start from a partially completed state. For example, check whether a package is already installed or a file already exists before creating it.

---

## Filtering Condition (where)

Specifying the optional `where` field on a task causes the task to execute only against targets matching the condition. See [Task Filtering Condition (where)]({{< relref "task-where" >}}) for details.

---

## stateChanges

Defines operations to update State after a task runs. See [Updating State with stateChanges]({{< relref "state-changes" >}}) for details.

---

## Next Steps

- [Updating State with stateChanges]({{< relref "state-changes" >}}) — Details on `stateChanges`
- [Task Filtering Condition (where)]({{< relref "task-where" >}}) — Details on the `where` field
- [Variables Available in Recipes]({{< relref "recipe-variables" >}}) — Reference for template variables and environment variables

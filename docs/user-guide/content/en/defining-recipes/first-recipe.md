---
title: "Tutorial: Your First Recipe"
weight: -1
---

# Tutorial: Your First Recipe

In this tutorial, you will write a minimal recipe from scratch and walk through the entire flow of applying it to a node. It takes about 15 minutes.

You will build a `kind: node` recipe that simply writes a greeting file on a node. The subject is trivial, but by the end of this tutorial you will understand:

- The basic structure of a recipe file (`nws-recipe.yaml`)
- How to reference a recipe from State
- The `nwsctl plan` → `nwsctl apply` execution cycle
- How to write deletion handling (destruct)

---

## Prerequisites

- `nwsctl` is installed ([Installation]({{< relref "/getting-started/installation" >}}))
- One Linux machine reachable over SSH (a VM, cloud instance, or physical machine)
  - Have its hostname (or IP address), user name, and private key path at hand
  - You have connected to it with the `ssh` command at least once, so its host key is registered in `~/.ssh/known_hosts` (Niwashi verifies the host key when connecting)

---

## Step 1: Prepare a Working Directory

Create a working directory with the following layout.

```
my-first-recipe/
├── recipes/
│   └── hello/
│       └── nws-recipe.yaml   ← the recipe you are about to write
└── state.yaml                ← the State you are about to write
```

```bash
mkdir -p my-first-recipe/recipes/hello
cd my-first-recipe
```

---

## Step 2: Write the Recipe

Create `recipes/hello/nws-recipe.yaml` with the following content.

```yaml
version: nws.recipe/v1
kind: node
metadata:
  id: my-org/hello
  version: "1.0.0"
  description: "My first recipe"
spec:
  provides:
    - name: my-org.hello
  tasks:
    - name: greet
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            echo "Hello, Niwashi! (node: {{ .Target }})" > /tmp/hello-niwashi.txt
            cat /tmp/hello-niwashi.txt

    - name: cleanup
      operation: destruct
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            rm -f /tmp/hello-niwashi.txt
```

Each block means the following.

| Block | Meaning |
|-------|---------|
| `kind: node` | Declares that this recipe targets a single node. For how to choose a recipe kind, see [Defining Recipes]({{< relref "_index" >}}) |
| `metadata.id` / `metadata.version` | The recipe identifier, written in `<org>/<name>` format |
| `spec.provides` | The capability name this recipe provides. State references the recipe by this name |
| `spec.tasks` | The list of operations to run. `exec.remote` is an action that connects to the node remotely and runs a script |
| `{{ .Target }}` | A template variable expanded to the name of the target node |
| `operation: destruct` | A task that runs only on deletion (`plan --destroy`). Write the inverse of the construction steps here |

---

## Step 3: Write the State

Next, write a State file that declares which node the recipe applies to. Create `state.yaml` with the following content, replacing everything under `connection` with the connection details of your machine.

```yaml
version: nws.state/v1

inventory:
  nodes:
    hello-node:
      instanceSelector:
        generator: my-servers
      capabilities:
        - my-org.hello

infrastructure:
  generators:
    my-servers:
      provisioner: external-instance
      params:
        instances:
          server-01:
            connection:
              ssh:
                address:
                  host: 192.168.1.10    # ← change to your machine
                  port: 22
                  user: ubuntu          # ← change to your machine
                auth:
                  method: privateKey
                  privateKeyPath: ~/.ssh/id_rsa   # ← change to your machine
                hostKey:
                  knownHostsPath: ~/.ssh/known_hosts   # ← file used for host key verification
```

This State declares two things.

- `infrastructure.generators`: registers an existing machine as an instance. `external-instance` is a built-in provisioner that uses an already running machine as-is (see [Generators]({{< relref "/defining-desired-state/generators" >}}))
- `inventory.nodes`: defines a logical node `hello-node` and lists the recipe's `provides` name (`my-org.hello`) under `capabilities`

---

## Step 4: Create a Plan

Initialize the workspace and create an execution plan.

```bash
nwsctl init
nwsctl plan --recipe-dir ./recipes -t state.yaml
```

`plan` loads the State and recipes, calculates the diff against the current state, and writes the execution plan to `plan.json`. Check that the output includes registering the instance and running the `greet` task.

---

## Step 5: Apply and Verify

Execute the plan.

```bash
nwsctl apply --plan plan.json --recipe-dir ./recipes
```

The `greet` task runs on the node, and the script output (`Hello, Niwashi! ...`) appears in the log.

Let's confirm the file was actually created on the node. With `nwsctl ssh`, you can log in to the node using the connection details registered in State.

```bash
nwsctl ssh hello-node
```

```bash
# on the node
cat /tmp/hello-niwashi.txt
# => Hello, Niwashi! (node: hello-node)
exit
```

If you run `nwsctl plan` again, the resulting plan contains no tasks to run because there is no diff. Niwashi executes only the difference between the current state and the desired state.

---

## Step 6: Clean Up

Create a destroy plan with the `--destroy` flag and apply it. When applying a destroy plan, the `--destroy` flag is also required for `apply`.

```bash
nwsctl plan --recipe-dir ./recipes --destroy
nwsctl apply --plan plan.json --recipe-dir ./recipes --destroy
```

When `apply` runs, a confirmation prompt asks whether it is OK to destroy all resources; approve it to continue (pass `--yes` to skip the prompt).

This time the `cleanup` task with `operation: destruct` runs and removes `/tmp/hello-niwashi.txt` from the node. If you like, log in to the node with plain `ssh` and confirm that the file is gone.

---

## What You Have Learned

- A recipe consists of `kind`, `metadata`, `spec.provides`, and `spec.tasks`
- Listing a recipe's `provides` name under `capabilities` in State applies the recipe to a node
- `nwsctl plan` builds an execution plan from the diff, and `nwsctl apply` executes it
- Deletion handling is written as tasks with `operation: destruct`

## Next Steps

Continue to the per-kind guide that matches the kind of recipe you want to build.

- [Defining Node Capabilities]({{< relref "node-capability" >}}) — more on `kind: node` used in this tutorial (parameters, real examples)
- [Defining Cluster Capabilities]({{< relref "cluster-capability" >}}) — operations spanning multiple nodes
- [Defining Infrastructure Provisioners]({{< relref "infrastructure-provisioning" >}}) — creating VMs and instances
- [Defining Host Configuration]({{< relref "host-configuration" >}}) — providing tools on the host
- [Defining Adapters]({{< relref "defining-adapters" >}}) — commands callable from other recipes

For the full specification, see [Recipe Format Reference]({{< relref "recipe-spec" >}}) and [Defining Tasks]({{< relref "defining-tasks" >}}).

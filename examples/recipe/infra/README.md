# `kind: infra` sample

A minimal `kind: infra` recipe that shows the core mechanics of an
infrastructure Provisioner: `exec.local`-only tasks, `count`/`{{ .Loop.index }}`
based looping, and recording instance connection info into State via
`stateChanges`.

This sample **fabricates** instance data instead of provisioning real
infrastructure (no Docker, no VM). That keeps it focused purely on the
`kind: infra` mechanics; see [`../node/`](../node/) for a recipe that binds to
a real, reachable instance.

Reference: [Defining Infrastructure Provisioners](../../../docs/user-guide/content/en/defining-recipes/infrastructure-provisioning.md).
Also see [Updating State with stateChanges](../../../docs/user-guide/content/en/defining-recipes/state-changes.md)
and [Generators in Detail](../../../docs/user-guide/content/en/defining-desired-state/generators.md).

## Prerequisites

- `nwsctl` resolvable on `PATH`
- `jq` installed (the recipe's `provision` task uses it to build the fabricated instance JSON)
- Run every command below from inside this directory (`examples/recipe/infra/`)

## Files

- `recipe/nws-recipe.yaml` — the `kind: infra` recipe (`examples/infra.dummy-vms`)
- `target.yaml` — declares a single generator, `demo-vms`, using the recipe as
  its provisioner with `count: 2`. No `inventory.nodes` are needed — a `kind: infra`
  recipe runs purely from `infrastructure.generators`.

## Run it

```bash
nwsctl plan --target ./target.yaml --recipe-dir ./recipe --out plan.json --with-init
nwsctl apply --plan plan.json --recipe-dir ./recipe
```

## Verify

Inspect the two fabricated instances that were registered in State:

```bash
jq '.infrastructure.generators."demo-vms".instances' .niwashi/state/state.json
```

You should see `dummy-vm-0` and `dummy-vm-1`, each with a `connection.ssh` block —
this is exactly the shape a real provisioner (Vagrant, cloud API, etc.) would
populate so that `kind: node` recipes can later connect over SSH.

## Destroy

```bash
nwsctl plan --target ./target.yaml --recipe-dir ./recipe --out plan.json --destroy
nwsctl apply --plan plan.json --recipe-dir ./recipe --destroy --yes
```

The recipe's `teardown` task (`operation: destruct`) runs, and
`jq '.infrastructure' .niwashi/state/state.json` shows the `instances` entries
removed.

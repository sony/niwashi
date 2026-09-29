# `kind: host` sample

A minimal `kind: host` recipe: an "installer" that detects a tool on the
host running `nwsctl` (here, `git`) and records its path/version in State.

A `kind: host` recipe **cannot** be listed directly in a State file's
`capabilities` — it only runs when pulled in via another recipe's
`spec.requires`. So this sample pairs the installer with a tiny `kind: infra`
"trigger" recipe whose only job is to declare `requires: [examples.host.tool.git]`
and print what the installer recorded. (`kind: infra` is used here purely as
the simplest way to trigger a host dependency and needs no real
infrastructure of its own — see [`../infra/`](../infra/) for more on
`kind: infra` itself.)

Reference: [Defining Host Configuration](../../../docs/user-guide/content/en/defining-recipes/host-configuration.md).
Also see [Recipe Format Reference](../../../docs/user-guide/content/en/defining-recipes/recipe-spec.md)
(`spec.runtime`, the `spec.requires` dependency table) and
[Updating State with stateChanges](../../../docs/user-guide/content/en/defining-recipes/state-changes.md).

## Prerequisites

- `nwsctl` resolvable on `PATH`
- `git` and `jq` installed (the installer's `check` task uses both)
- Run every command below from inside this directory (`examples/recipe/host/`)

## Files

- `recipe/nws-catalog.yaml` — bundles the two recipes below (multiple recipes
  in one directory require `nws-catalog.yaml`; see
  [Bundling Multiple Recipes](../../../docs/user-guide/content/en/defining-recipes/catalog.md))
- `recipe/installer.yaml` — `kind: host`, `examples/host.git-installer`, provides `examples.host.tool.git`
- `recipe/trigger.yaml` — `kind: infra`, `examples/infra.host-trigger`, requires `examples.host.tool.git`
- `target.yaml` — declares a single generator, `trigger`, using the trigger recipe

## Run it

```bash
nwsctl plan --target ./target.yaml --recipe-dir ./recipe --out plan.json --with-init
nwsctl apply --plan plan.json --recipe-dir ./recipe
```

## Verify

The apply log for the `report` task shows the path/version the installer
detected:

```
git detected at /usr/bin/git (version 2.34.1)
```

(find it under `.niwashi/runs/<run-id>/infra-trigger/.../logs/report_trigger.log`)

It's also recorded in State:

```bash
jq '.runtime.tool.examples_git' .niwashi/state/state.json
# => { "path": "/usr/bin/git", "version": "2.34.1" }
```

## Destroy

```bash
nwsctl plan --target ./target.yaml --recipe-dir ./recipe --out plan.json --destroy
nwsctl apply --plan plan.json --recipe-dir ./recipe --destroy --yes
```

This removes the `trigger` generator entry. `runtime.tool.examples_git` stays
in State — the installer only records a *detected fact* about the host
("git is present at this path"), so unlike `kind: infra` instances it has no
`operation: destruct` counterpart in the real reference installers either
(e.g. `recipe/core/tools/git.yaml` in this repo, `nws/tools.git`).

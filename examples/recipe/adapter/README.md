# `kind: adapter` sample

A minimal `kind: adapter` recipe that wraps `git` and exposes a `version`
command, invoked from another recipe via the `tool.run` action — the pattern
described in [Defining Adapters](../../../docs/user-guide/content/en/defining-recipes/defining-adapters.md)
and [Using Adapters](../../../docs/user-guide/content/en/defining-recipes/using-adapters.md).

Three recipes work together here:

- `recipe/installer.yaml` (`kind: host`) — detects `git` and records its path
  under `/runtime/tool/examples_git` (same pattern as [`../host/`](../host/),
  duplicated here so this sample is self-contained)
- `recipe/adapter.yaml` (`kind: adapter`) — references
  `{{ .Runtime.tool.examples_git.path }}` and exposes it as a `version` command
- `recipe/caller.yaml` (`kind: infra`) — invokes `tool.run` with
  `toolRef: examples.adapter.tool.git`, `command: version`

**Note on `spec.requires`**: both the adapter *and* the caller declare
`requires: [examples.host.tool.git]`. Declaring it only on the adapter is not
enough — the planner only schedules the installer job when something in the
executed phase graph asks for it, and the adapter's own tasks run inline
inside the caller's job rather than as a separate scheduled job. So the
recipe that actually needs `{{ .Runtime.* }}` populated before it runs — here,
transitively, the caller — must declare the dependency itself.

Also see [Recipe Format Reference](../../../docs/user-guide/content/en/defining-recipes/recipe-spec.md)
for `allowedScope`/`executionUnit`/`commands`, [Defining Host Configuration](../../../docs/user-guide/content/en/defining-recipes/host-configuration.md)
for the installer half of the pattern, and
[Ansible Adapter](../../../docs/user-guide/content/en/recipes/ansible.md) for a
real-world reference adapter.

## Prerequisites

- `nwsctl` resolvable on `PATH`
- `git` and `jq` installed
- Run every command below from inside this directory (`examples/recipe/adapter/`)

## Files

- `recipe/nws-catalog.yaml` — bundles the three recipes (multiple recipes in
  one directory require `nws-catalog.yaml`)
- `recipe/installer.yaml`, `recipe/adapter.yaml`, `recipe/caller.yaml`
- `target.yaml` — declares a single generator, `trigger`, using the caller recipe

## Run it

```bash
nwsctl plan --target ./target.yaml --recipe-dir ./recipe --out plan.json --with-init
nwsctl apply --plan plan.json --recipe-dir ./recipe
```

## Verify

The plan should include a `host:examples/host.git-installer` job before the
`infra:examples/infra.adapter-caller` job. After apply, the `call-adapter`
task's log shows the adapter's output — the real `git --version` output,
executed through the adapter rather than called directly:

```
git version 2.34.1
```

(find it under `.niwashi/runs/<run-id>/infra-trigger/.../logs/call-adapter_trigger.log`)

## Destroy

```bash
nwsctl plan --target ./target.yaml --recipe-dir ./recipe --out plan.json --destroy
nwsctl apply --plan plan.json --recipe-dir ./recipe --destroy --yes
```

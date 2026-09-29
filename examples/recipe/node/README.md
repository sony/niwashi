# `kind: node` sample

A minimal `kind: node` recipe — connects to a target node over SSH
(`exec.remote`), writes a greeting file that includes the node's name
(`{{ .Target }}`), and removes it again on destroy (`operation: destruct`).
This is essentially the recipe from the
[Tutorial: Your First Recipe](../../../docs/user-guide/content/en/defining-recipes/first-recipe.md),
wired up to run standalone via a disposable Docker SSH server instead of a
machine you own.

Reference: [Defining Node Capabilities](../../../docs/user-guide/content/en/defining-recipes/node-capability.md).
Also see [Recipe Format Reference](../../../docs/user-guide/content/en/defining-recipes/recipe-spec.md),
and [`nwsctl plan`](../../../docs/user-guide/content/en/execution/plan.md) /
[`nwsctl apply`](../../../docs/user-guide/content/en/execution/apply.md) /
[`nwsctl ssh`](../../../docs/user-guide/content/en/execution/ssh.md).

## Prerequisites

- `nwsctl` resolvable on `PATH`
- Docker with `docker compose`
- Host port `2222` free (used to reach the sample's SSH server; change
  `SSH_PORT` in `.env` and the `port` in `target.yaml` together if it's taken)
- Run every command below from inside this directory (`examples/recipe/node/`)

## Files

- `recipe/nws-recipe.yaml` — the `kind: node` recipe (`examples/node.hello`)
- `compose.yml`, `.env`, `gen-key.sh`, `gen-known_hosts.sh` — a disposable
  `linuxserver/openssh-server` container standing in for a real, reachable
  node (same pattern as [`../../external-instance/`](../../external-instance/))
- `target.yaml` — registers the container as an instance via the built-in
  `external-instance` provisioner, and binds a logical node `hello-node` to it
  with `capabilities: [examples.node.hello]`

## Set up the SSH target

```bash
./gen-key.sh
docker compose up -d
# wait a couple of seconds for sshd to start
./gen-known_hosts.sh
```

## Run it

```bash
nwsctl plan --target ./target.yaml --recipe-dir ./recipe --out plan.json --with-init
nwsctl apply --plan plan.json --recipe-dir ./recipe
```

## Verify

The `greet` task actually ran on the container, over SSH:

```bash
docker exec nws-node-sample cat /tmp/niwashi-node-hello.txt
# => Hello, Niwashi! (node: hello-node)
```

You can also log in with the connection info Niwashi tracked in State:

```bash
nwsctl ssh hello-node
# on the node: cat /tmp/niwashi-node-hello.txt
```

## Destroy

```bash
nwsctl plan --target ./target.yaml --recipe-dir ./recipe --out plan.json --destroy
nwsctl apply --plan plan.json --recipe-dir ./recipe --destroy --yes
```

The `cleanup` task (`operation: destruct`) runs and removes the file:

```bash
docker exec nws-node-sample sh -c 'test -f /tmp/niwashi-node-hello.txt && echo EXISTS || echo REMOVED'
# => REMOVED
```

## Clean up the SSH target

```bash
docker compose down
rm -f id_ed25519 id_ed25519.pub known_hosts
```

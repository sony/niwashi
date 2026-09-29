# `kind: cluster` sample

A minimal `kind: cluster` recipe demonstrating the **`where` pattern**
described in [Defining Cluster Capabilities](../../../docs/user-guide/content/en/defining-recipes/cluster-capability.md):
a step that runs on every member node, role-specific steps dispatched by
label (`node.labels.role`), and a final centralized step for the cluster as
a whole.

Two logical nodes (`cp`, `worker1`) belong to one cluster (`demo-cluster`),
each bound to its own disposable `linuxserver/openssh-server` container
(same pattern as [`../node/`](../node/), just with two containers instead
of one).

Reference: [Defining Cluster Capabilities](../../../docs/user-guide/content/en/defining-recipes/cluster-capability.md).
Also see [Task Filtering Condition (where)](../../../docs/user-guide/content/en/defining-recipes/task-where.md),
[Recipe Format Reference](../../../docs/user-guide/content/en/defining-recipes/recipe-spec.md),
and [Clusters in Detail](../../../docs/user-guide/content/en/defining-desired-state/clusters.md)
for the `inventory.clusters.<name>.nodes` object form used in `target.yaml`.

**Note on `prepare`'s `where`**: per-node decomposition for a `kind: cluster`
task is gated purely on whether `where` is non-empty (this is intended
behavior, not a bug) — a task with no `where` runs once against the cluster
as a single unit, which has no SSH connection of its own, so `exec.remote`
is rejected (`ExecRemoteAction can only be used for node-level tasks`).
`cluster-capability.md`'s own "runs on every node" example omits `where`
entirely and doesn't work as written for that reason. To explicitly select
every member node, either give the task a `where` that's unconditionally
true (`where: "true"`) or one that matches every node's labels — this sample
uses the latter (`where: "node.labels.role in ['master', 'worker']"`) since
it's more self-documenting about which roles the cluster is expected to have.

## Prerequisites

- `nwsctl` resolvable on `PATH`
- Docker with `docker compose`
- Host ports `2222` and `2223` free (used to reach the two containers;
  change `CP_SSH_PORT`/`WORKER_SSH_PORT` in `.env` and the matching `port`
  values in `target.yaml` together if either is taken)
- Run every command below from inside this directory (`examples/recipe/cluster/`)

## Files

- `recipe/nws-recipe.yaml` — the `kind: cluster` recipe (`examples/cluster.hello`)
- `compose.yml`, `.env`, `gen-key.sh`, `gen-known_hosts.sh` — two disposable
  SSH containers, `nws-cluster-cp` and `nws-cluster-worker1`
- `target.yaml` — registers both containers as instances, binds them to
  logical nodes `cp`/`worker1`, and groups them into cluster `demo-cluster`
  with `role: master` / `role: worker` labels (the object form of
  `inventory.clusters.<name>.nodes`)

## Set up the SSH targets

```bash
./gen-key.sh
docker compose up -d
# wait a couple of seconds for sshd to start on both containers
./gen-known_hosts.sh
```

## Run it

```bash
nwsctl plan --target ./target.yaml --recipe-dir ./recipe --out plan.json --with-init
nwsctl apply --plan plan.json --recipe-dir ./recipe
```

## Verify

`prepare` runs on both nodes; `init-master` only on `cp` (`role: master`);
`join-workers` only on `worker1` (`role: worker`, and only after
`init-master` completes via `dependsOn`); `verify` runs once, centrally:

```bash
docker exec nws-cluster-cp sh -c 'cat /tmp/niwashi-cluster-hello/*'
# => prepared: cp
# => master initialized: cp
# (no worker.txt on this node)

docker exec nws-cluster-worker1 sh -c 'cat /tmp/niwashi-cluster-hello/*'
# => prepared: worker1
# => worker joined: worker1
# (no master.txt on this node)
```

`verify`'s output is in the apply log
(`.niwashi/runs/<run-id>/cluster-demo-cluster/.../logs/verify.log`):

```
Cluster demo-cluster setup completed
```

## Destroy

```bash
nwsctl plan --target ./target.yaml --recipe-dir ./recipe --out plan.json --destroy
nwsctl apply --plan plan.json --recipe-dir ./recipe --destroy --yes
```

This recipe has no `operation: destruct` tasks of its own, so destroy just
tears down the two provisioned instances (their containers keep running
until you `docker compose down` — the destroy plan only affects Niwashi's
State, not the containers themselves).

## Clean up the SSH targets

```bash
docker compose down
rm -f id_ed25519 id_ed25519.pub known_hosts
```

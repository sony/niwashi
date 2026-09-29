---
status: accepted
date: 2025-12-01
decision-makers: Takahiro Okayama
---

# Execute Recipe Scripts on Remote Nodes via a Built-in SSH Action

## Context and Problem Statement

Before this decision, Niwashi could only execute recipe scripts on the machine running the CLI (`exec.local`). Configuring a remote target (a VM, a physical server) required delegating to an external tool such as Ansible through `tool.run`. For a simple script — install a package, write a config file, restart a service — pulling in a full external automation tool added a dependency and broke the interface consistency recipe authors already had with `exec.local`. The question this decision answers is whether Niwashi should provide a built-in mechanism to run a recipe's script directly on a remote node over SSH, and if so, how connection, file transfer, and execution should work.

## Decision Drivers

* Reduce dependencies: a simple remote configuration task should not require installing and learning a separate automation tool.
* Interface consistency: a remote execution action should look and behave like `exec.local` (same script/env template mechanism, same logging) so recipe authors do not learn two mental models.
* Security-sensitive wire-protocol handling should not be maintained in-house without a corresponding benefit over an existing, widely-used library.
* Remote file transfer must not depend on an external binary being present and behaving consistently on whatever host runs the CLI.
* Preserve each recipe asset's relative path when uploading, so scripts can reference files the same way they do on the host, and same-named files in different subdirectories do not collide.

## Considered Options

* Add a built-in `exec.ssh` action with its own SSH+SFTP client (chosen)
* Require all remote execution to go through Ansible via `tool.run`
* Implement a custom SSH protocol handler instead of using an existing SSH library
* Shell out to the system `ssh`/`scp` binaries via `subprocess`
* Flatten transferred recipe assets into a single remote directory instead of preserving their relative paths

## Decision Outcome

Chosen option: "Add a built-in `exec.ssh` action with its own SSH+SFTP client", because it is the only option that removes the external-tool dependency for simple cases while keeping the same script/env template interface `exec.local` already offers, because a mature SSH library avoids reimplementing a security-sensitive wire protocol in-house, and because it avoids depending on a platform-specific `ssh` binary being present on the host running the CLI. The action was additionally restricted to node-scoped recipes for its initial implementation: cluster support was deferred for a later phase, while infra/host support was ruled out for this action outright.

### Consequences

* Good, because a node-scoped recipe now expresses remote configuration with the same `scriptTpl`/`envTpl` fields as `exec.local`, so a recipe author who already knows one knows the other.
* Good, because recipe assets are transferred with their original relative directory structure preserved, so scripts can reference `./bin/setup.sh` or `config/app.conf` exactly as they appear in the recipe, avoiding the collisions a flattened layout would risk.
* Good, because the action reuses the same output/log mechanism as `exec.local`, rather than a parallel logging mechanism specific to remote execution.
* Bad, because the action is hard-restricted to node-scoped recipes: a recipe of any other scope cannot use it at all, even indirectly.
* Neutral, because host key verification uses a `known_hosts`-based check rather than accepting unknown hosts outright, a stricter default than what the original design proposed.

### Confirmation

* The node-scope restriction can be confirmed by writing a recipe of any scope other than node that uses `exec.ssh`, and observing that it is rejected at plan/execution time.
* That recipe assets preserve their relative paths on the remote side can be confirmed by writing a recipe with `exec.ssh` whose `spec.assets` include a nested path (e.g. `bin/setup.sh`), and observing that the script can reference it by that same relative path on the remote node.
* That an off-the-shelf SSH client library is used, rather than a custom protocol implementation or a shelled-out `ssh`/`scp` binary, is confirmed by `go.mod` listing a third-party SSH client library as a dependency for the SSH transport (`internal/transport/ssh/`).
* The `known_hosts`-based host key default, stricter than the original design's "accept unknown hosts" plan, is confirmed by `internal/transport/ssh/transport.go`'s `Validate`, which rejects a connection whose `hostKey.knownHostsPath` is empty.

## Pros and Cons of the Options

### Add a built-in `exec.ssh` action with its own SSH+SFTP client (chosen)

* Good, because it removes the external-tool dependency for simple remote tasks.
* Good, because it lets the action share `exec.local`'s script/env template interface and logging.
* Neutral, because it makes SSH connection management part of Niwashi's own maintenance surface, in exchange for not depending on an external tool being installed.

### Require all remote execution to go through Ansible via `tool.run`

* Good, because it reuses an existing, battle-tested tool with a rich module ecosystem and built-in idempotency handling.
* Bad, because it requires Ansible to be installed even for a single-line remote script, and adds Ansible's own module/inventory model for recipe authors to learn on top of Niwashi's.

### Implement a custom SSH protocol handler instead of using an existing SSH library

* Bad, because it means reimplementing a security-sensitive wire protocol in-house, with the associated security risk and maintenance burden, for no benefit over an existing, widely-used library.

### Shell out to the system `ssh`/`scp` binaries via `subprocess`

* Good, because it avoids linking any SSH library at all.
* Bad, because it depends on an external `ssh` binary being present and behaving consistently on whatever machine runs the Niwashi CLI, and makes file transfer and structured error/exit-code handling considerably harder to get right than a library that exposes them as typed calls.

### Flatten transferred recipe assets into a single remote directory instead of preserving their relative paths

* Good, because extraction logic would be simpler if it never had to recreate subdirectories.
* Bad, because two assets with the same filename in different subdirectories would collide once flattened, and scripts could no longer reference assets by the relative paths they use on the host.

## More Information

The original design deferred two further considerations beyond the scope of the initial decision:

1. **Password-based SSH authentication** — deprioritized; `privateKey` remains the only supported method today (`internal/transport/ssh/auth.go`, `AuthMethodPrivateKey`).
2. **Support for `scope: cluster`** — deferred, because which task execution unit a cluster-wide remote script should run under was not resolved at the time; only node-scoped recipes can use this action today.

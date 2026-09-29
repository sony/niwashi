# Change Log

## v0.5.0 (2026-06-12)

### Added

#### Recipe Format

- Add `allowedScope: *` wildcard to `kind: adapter` to allow execution under any scope.
- Add `inheritEnv` field to `tool.run` action (same as `exec.local.inheritEnv`).
- Add `as` field to `spec.requires` entries. Enables cross-recipe `store/` access via `{{ .Stores.<alias>.<key> }}` template variable.

#### Plan Format

- Add `metadata` field to plan files. Contains `createdAt`, `nwsctlVersion`, and recipe fingerprints (`fqid` + `hash`) for traceability and tamper detection.

#### CLI

- Add `--strict-check` option to `apply` command. When specified, apply fails if any recipe has been modified since the plan was created. Without this flag, mismatches are logged as warnings and apply proceeds.

#### JSON Schema

- Add JSON Schema files for State (`nws-state-v1.json`), Recipe (`nws-recipe-v1.json`), Catalog (`nws-catalog-v1.json`), and Profile (`nws-profile-v1.json`) formats under `docs/schema/`.

### Changed

#### Recipe Format

- Change `stateChanges.op: add` to `op: set`. The old `add` value is deprecated and prints a warning; it will be removed in a future release.

#### Connection

- Add SSH connection retry, configurable via environment variables `NWS_SSH_RETRY_MAX_COUNT` (default: 3) and `NWS_SSH_RETRY_INTERVAL_SEC` (default: 5s). Also add `NWS_SSH_CONNECT_TIMEOUT_SEC` and `NWS_SSH_HANDSHAKE_TIMEOUT_SEC` for timeout tuning.

### Fixed

- Fix crash when an adapter recipe references an undefined command.
- Fix state validation: state schema is now validated after `stateChanges` patches are applied per task.

### Internal

- Refactor workflow engine internals (`applier` → `recipe_exec` package).
- Add unit tests for recipe spec, filter, and identity-related logic.


## v0.4.0 (2026-04-30)

### Added

#### Recipe Format

- Add `inheritEnv` field to `exec.local`.
- Add `runtime` field for `kind: host` recipe.


#### State Format

- Add `store` field for storing or sharing recipe specific data.

#### CLI

- Add `--prune` and `--yes` option to `plan` and `apply` command
- Add new `external-instance.from=<source>` recipe for loading instance information from external source. source=ssh_config|file.
- Add `--concurrency` option to `apply` command.

### Changed

#### Recipe Format

- Change `exec.ssh` -> `exec.remote`.
- Change `stateChanges.path` to **relative path** from absolute path.

#### State Format

- Change `pool` field name to `instances`.

#### Plan format

- Add `mode` field to protect unexpected destroy/prune execution.

#### CLI

- Change destroy/prune operation. `plan` and `apply` requires same option(`--destroy` or `--prune`).
- Change dry-run `simulate`. Remote operation, command execution and probing system information will return dummy implementation.
- Change console format of `plan` command.

### Fixed

- Fix tty issue with `ssh` command.
- Fix envTpl was not available.
- Documentation Link
- Update README.md

### Deprecated

- Deprecated `exec.ssh`.

### Removed

- Remove default inherit environments except PATH for `exec.local`

### Internal

- Add mid layer structure for state and recipe.
- Refactor CmdFactory and TransportFactory.
- Strict validation against state format.


## v0.3.0 (2026-03-27)

### Added

- Enable probing system information.
- Conccurent task execution.
- Add `labels` and `where` fields for condition.
- Add TUI option for `apply` command.
- Add specifying instance option for `instanceSelector`.

### Changed

- Sync `{{ .Outputs }}` data when `exec.ssh`.
- Add path constraints for `stateChanges`.

### Fixed

- Fix an issue that `nodeRef` is cleard unexpectedly when dup applying infra recipe.
- Fix an issue that mix log format.
- Fixed an issue with cluster recipe concurrency.

### Deprecated
### Removed
### Internal

- Refactor recipe structure.
- Moved all packages to internal directory.


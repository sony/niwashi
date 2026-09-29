# Change Log

## v0.5.3 (2026-09-28)

### Added

#### Recipe Format

- Add `operation: update` to tasks. When the `params` of an applied `kind: node` / `kind: cluster` capability change, or its resolved recipe version is upgraded, `nwsctl plan` schedules the recipe's `update` tasks. No extra CLI flag is required.

#### Plan Format

- Add `pendingUpdates` field to plan files. It lists capabilities whose `params` or version changed but whose recipe has no `update` task. These are also shown under `Pending Updates` in the `plan` output.

### Changed

#### State Format

- The applied `params` of a capability are now recorded in State (previously stored as empty), and are updated together with `version` after `update` tasks succeed.

#### CLI

- `plan` fails with an error when the recipe version pinned for an applied capability is lower than the applied version.
- When a task fails, `apply` stops starting new jobs, waits for the running tasks, and exits with an error.
- When a construct fails, the capability (including its `store/`) or the generator entry is removed from State, so the next run constructs it again.

### Fixed

- Fix the capability or generator entry being removed from State even when a destruct task failed.
- Fix errors from job setup and job completion (e.g. State patch failures) not being reported as `apply` failures.

### Internal

- Add e2e tests for update, prune, and failure handling.


## v0.5.2 (2026-09-11)

### Added

#### CLI

- Add Windows builds to release artifacts.
- Support terminal resizing in the `ssh` command on Windows.
- Warn about `.yaml` / `.yml` files in recipe directories that are neither `nws-recipe.yaml`, `nws-catalog.yaml`, nor listed under `spec.assets`.

#### Documentation

- Add `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, and Architecture Decision Records under `docs/adr/`.
- Add sample recipes for each kind under `examples/recipe/` and remove outdated examples.

### Changed

#### Recipe Format

- Rendered scripts are saved with an extension derived from the shebang (`.sh` for `sh` / `bash`, `.ps1` for `powershell` / `pwsh`) instead of `.action`. PowerShell scripts are executed with `-File`.

#### State Format

- Validate the format of `instanceRef` in nodes and `nodeRef` in instances.

#### Recipe Loading

- Skip files and directories whose names start with `.` when loading recipe directories.
- Improve the error message when no recipe provides a requested capability (`no recipe provides "<name>"`).

#### Distribution

- Include `LICENSE`, `README`, `CHANGELOG`, and third-party license texts (`LICENSES/`) in release archives.

### Fixed

- Fix capability `params` in a profile not taking precedence over those in State.
- Fix crashes on State files with a node that references an instance without a connection, with no `infrastructure`, or with an empty patch path.

### Internal

- Remove the remote-side script rename (`Move-Item`) in WinRM execution; the `.ps1` extension and `-File` are now handled when the script is generated.
- Add SPDX license headers to source files.
- Add `licenses` and `docs-lint` Makefile targets.


## v0.5.1 (2026-07-30)

### Added

#### Connection

- Add WinRM transport for Windows nodes (`connection.winrm`). Supports NTLM authentication, HTTPS (port 5986) by default, and server certificate verification settings (`serverCert.insecureSkipVerify`, `serverCert.caCertPath`).

#### JSON Schema

- Add `winrm` connection to the State schema.

#### Documentation

- Add a first recipe tutorial, connection references for SSH and WinRM, script writing guidelines (shebang and default interpreter), the workspace directory layout, and the structure of `NWS_STATE`.
- Add an example for WinRM external instances (`examples/external-instance-winrm/`).

### Changed

#### Recipe Format

- Change the default interpreter for scripts without a shebang on Windows nodes from `pwsh` to `powershell` (Windows PowerShell 5.1).
- `exec.remote` uses the path separator of the remote OS for the remote workspace.

#### Documentation

- Remove `update` from the documented values of `operation` (it was not implemented).
- Rename "official" provisioners and recipes to "reference" provisioners and recipes.

### Fixed

- Fix the caller recipe directory not being set for `exec.remote` tasks.
- Fix probing being attempted for instances with empty connection data.
- Fix prober creation errors being ignored.
- Fix errors on closing files and sessions being ignored.

### Internal

- Add the Apache License 2.0 (`LICENSE`).
- Introduce `golangci-lint` and fix reported issues (errcheck, staticcheck, ineffassign, unused).
- Add Makefile targets for setup, lint, and coverage, and a `check` target that runs lint, unit tests, and e2e tests.
- Replace the shell-based integration tests with Go e2e tests that start an SSH server container.


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


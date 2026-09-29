---
title: "nwsctl init"
weight: 1
---

# nwsctl init

Initializes a Workspace. Run this once before executing `nwsctl plan` or `nwsctl apply`.

```
nwsctl init [flags]
```

---

## Overview

`nwsctl init` creates the Workspace directory and the current state file (`state.json`) used by Niwashi.

Directory structure created:

```
.niwashi/          # Workspace (can be changed with --work-dir)
├── state/
│   └── state.json   # Current state (managed by Niwashi)
├── store/
└── runs/
```

`state.json` is a file managed internally by Niwashi and does not need to be edited directly. It is separate from the desired State file defined by the user.

---

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--work-dir` | | `.niwashi` | Path to the Workspace |
| `--state` | `-s` | (none) | Path to a State file to load as the initial state |

---

## Usage

### New Setup (starting from an empty state)

```bash
nwsctl init
```

An empty state is created at `.niwashi/state/state.json`. Use this when starting with Niwashi for the first time.

### Starting from an Existing State

```bash
nwsctl init -s current.yaml
```

Loads the specified State file as the current state. Use this when taking over a state exported with `nwsctl export`.

### Changing the Workspace Location

```bash
nwsctl init --work-dir /path/to/workspace
```

Creates the Workspace at a location other than the default `.niwashi`. You will also need to specify the same `--work-dir` when running `nwsctl plan` or `nwsctl apply`.

---

## Notes

- Running this on an already-initialized Workspace will result in an error.
- Supported file formats: `.yaml`, `.yml`, `.json`, `.jsonc`

### Constraints When Loading a State with `-s`

It is possible to load a State exported with `nwsctl export` using `-s`, but depending on the Recipe, it may not work correctly.

**Example**: The Vagrant Recipe creates VM images and other files under the Workspace directory. Since these files are not included in the `export`, re-provisioning may be required after importing.

---

## Next Steps

- [nwsctl plan]({{< relref "plan" >}}) — Create an execution Plan

---
title: "nwsctl export"
weight: 5
---

# nwsctl export

Exports the current state of the Workspace to a YAML file.

```
nwsctl export [flags]
```

---

## Overview

`nwsctl export` converts the Workspace's `state.json` (a JSON format file managed internally by Niwashi) to YAML format and outputs it.

The exported YAML can be passed to `nwsctl init -s` to use as the initial state of another Workspace.

---

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--out` | `-o` | (stdout) | Path to the output YAML file |
| `--include-runtime` | | `false` | Include the runtime section in the output |
| `--work-dir` | | `.niwashi` | Path to the Workspace |

---

## Usage

### Export to Standard Output

```bash
nwsctl export
```

### Export to a File

```bash
nwsctl export -o current.yaml
```

### Transferring the Environment

When starting from the same state on a different machine or in a different Workspace:

```bash
# Export the current state to a file
nwsctl export -o current.yaml

# Transfer that state to another Workspace
nwsctl init -s current.yaml --work-dir /path/to/new-workspace
```

### Including the Runtime Section

```bash
nwsctl export --include-runtime -o current-full.yaml
```

By default, the runtime section (runtime information collected by Niwashi, such as tool paths) is excluded. Adding `--include-runtime` includes it. Normally, pass the export without it to `nwsctl init -s`.

---

## Notes

It is possible to use `nwsctl init -s` with an exported State in another Workspace, but depending on the Recipe, it may not work correctly.

**Example**: The Vagrant Recipe creates VM images and other files under the Workspace directory. Since these files are not included in the `export`, re-provisioning may be required after importing.

---

## Next Steps

- [nwsctl init]({{< relref "init" >}}) — Initialize a new Workspace using the exported State

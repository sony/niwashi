---
title: "nwsctl ssh"
weight: 4
---

# nwsctl ssh

SSH into a managed Node.

```
nwsctl ssh NODE_NAME [flags]
```

---

## Overview

`nwsctl ssh` automatically retrieves the connection information (hostname, port, user, private key) from the Workspace's `state.json` and establishes an SSH connection. There is no need to look up IP addresses or key paths manually.

It starts an interactive shell. To exit, run `exit` or close the terminal.

---

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--work-dir` | `.niwashi` | Path to the Workspace |

---

## Usage

```bash
nwsctl ssh web-server-01
```

The Node name must match the key name in `inventory.nodes` of the State file.

```bash
# Specifying the Workspace location
nwsctl ssh web-server-01 --work-dir /path/to/workspace
```

---

## Notes

- If the Node's connection information does not exist in the State (e.g., the infrastructure has not yet been applied), an error will occur.
- Nodes whose connection type is not SSH cannot be connected to.

---

## Next Steps

- [nwsctl export]({{< relref "export" >}}) — Export the current State to a file

---
title: "Installation"
weight: 3
---

# Installation

## Supported Environments

| OS | Architecture | Status |
|----|--------------|--------|
| Linux | x86_64 | Verified |
| Windows | x86_64 | Unverified (reference recipes not supported) |
| macOS | x86_64 | Unverified (reference recipes not supported) |

> **Note**: nwsctl itself can run on Windows and macOS, but reference recipes assume a bash environment and may not work correctly. See [Limitations]({{< relref "limitations" >}}) for details.

---

## Installing nwsctl

### 1. Obtain the Archive

Download the archive for your architecture from the [releases page](https://github.com/sony/niwashi/releases).

| OS | File |
|----|------|
| Linux | `nwsctl_Linux_x86_64.tar.gz` |
| Windows | `nwsctl_Windows_x86_64.zip` |
| macOS | `nwsctl_Darwin_x86_64.tar.gz` |

### 2. Extract the Archive

**Linux / macOS**

```bash
tar -xzf nwsctl_Linux_x86_64.tar.gz
```

This produces the `nwsctl` binary.

**Windows**

Extract `nwsctl_Windows_x86_64.zip` using File Explorer or any ZIP tool. This produces `nwsctl.exe`.

### 3. Place the Binary

**Linux / macOS**

Move the binary to a directory in your `PATH` (e.g., `/usr/local/bin`):

```bash
sudo mv nwsctl /usr/local/bin/
```

Alternatively, place it in a user-local directory such as `~/.local/bin/`, as long as it is included in your `$PATH`.

**Windows**

Place `nwsctl.exe` in any directory, then add that directory to the `PATH` environment variable.

### 4. Verify the Installation

```bash
nwsctl version
```

If the version number is displayed, the installation is complete.

---

## Next Steps

- [Workflow]({{< relref "workflow" >}}) — Basic workflow using nwsctl

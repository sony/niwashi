---
title: "nwsctl Command Reference"
weight: 4
---

# nwsctl Command Reference

`nwsctl` is the main interface for Niwashi. It lets you plan, execute, and manage your infrastructure from the command line.

---

## Basic Usage

For a typical Niwashi workflow, see [Workflow]({{< relref "workflow" >}}).

---

## Defining the Desired State

For information on how to write the desired State passed to `nwsctl plan`, see [Defining the Desired State]({{< relref "defining-desired-state" >}}).

---

## Command List

| Command | Description |
|---------|-------------|
| [`nwsctl init`](init/) | Initialize a Workspace |
| [`nwsctl plan`](plan/) | Create an execution Plan |
| [`nwsctl apply`](apply/) | Apply an execution Plan |
| [`nwsctl ssh`](ssh/) | SSH into a managed Node |
| [`nwsctl export`](export/) | Export the current State to a file |

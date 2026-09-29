---
title: "Connections in Detail"
weight: 9
---

# Connections in Detail

This page is for anyone registering an already-running machine with the `external-instance` provisioner (see [Generators in Detail]({{< relref "generators" >}})) and writing its `connection` block by hand.

```yaml
infrastructure:
  generators:
    existing-servers:
      provisioner: external-instance
      params:
        instances:
          server-01:
            connection:
              ssh:      # or winrm
                ...
```

Each instance's `connection` tells Niwashi how to reach that specific machine: the address to dial, how to authenticate as the client, and (for `winrm`) how to verify the server. Niwashi supports two connection types:

- [SSH Connection Reference]({{< relref "connection-ssh" >}}) -- Linux/Unix machines
- [WinRM Connection Reference]({{< relref "connection-winrm" >}}) -- Windows machines (non-domain-joined)

See [Infrastructure Basics]({{< relref "basic-concepts/infrastructure" >}}) for a minimal example of each.

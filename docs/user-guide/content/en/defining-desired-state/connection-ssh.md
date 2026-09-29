---
title: "SSH Connection Reference"
weight: 10
---

# SSH Connection Reference

`connection.ssh` describes how to reach a Linux/Unix machine registered via the `external-instance` provisioner. See [Connections in Detail]({{< relref "connections" >}}) for where this fits.

```yaml
connection:
  ssh:
    address:
      host: 192.168.1.10
      port: 22
      user: ubuntu
    auth:
      method: privateKey
      privateKeyPath: ~/.ssh/id_rsa
    hostKey:
      knownHostsPath: ~/.ssh/known_hosts
```

## address

| Field | Required | Default | Description |
|-------|----------|---------|--------------|
| `host` | Yes | | Hostname or IP address |
| `port` | | `22` | SSH port |
| `user` | Yes | | Username to connect as |

## auth

The client's own credential -- how Niwashi authenticates itself to the machine.

| Field | Required | Default | Description |
|-------|----------|---------|--------------|
| `method` | | `privateKey` | Only `privateKey` is supported |
| `privateKeyPath` | Yes | | Path to the private key file |
| `passphraseRef.fromEnv` / `passphraseRef.fromFile` | | | Passphrase for an encrypted private key, if any. No plaintext passphrase field exists by design |

## hostKey

How the machine's identity is verified -- kept separate from `auth` because verifying the party on the other end of the connection is a different concern from the client's own credentials.

| Field | Required | Default | Description |
|-------|----------|---------|--------------|
| `method` | | `knownHostsFile` | Only `knownHostsFile` is supported |
| `knownHostsPath` | | | Path to a standard OpenSSH `known_hosts` file that already contains this host's key (e.g. from a prior manual connection or `ssh-keyscan`). Unknown hosts are rejected, not auto-trusted |

## Connection mechanics (timeouts, retries)

Unlike `address`/`auth`/`hostKey`, these aren't set per-instance in `connection.ssh` -- they're process-wide environment variables:

| Variable | Default | Description |
|----------|---------|--------------|
| `NWS_SSH_CONNECT_TIMEOUT_SEC` | `10` | TCP connect timeout |
| `NWS_SSH_HANDSHAKE_TIMEOUT_SEC` | `10` | SSH handshake timeout |
| `NWS_SSH_RETRY_MAX_COUNT` | `3` | Max connection retry attempts |
| `NWS_SSH_RETRY_INTERVAL_SEC` | `5` | Interval between retries |

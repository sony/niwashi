---
title: "WinRM Connection Reference"
weight: 11
---

# WinRM Connection Reference

`connection.winrm` describes how to reach a Windows machine (standalone/non-domain-joined -- an AWS EC2 Windows instance or physical machine) registered via the `external-instance` provisioner. See [Connections in Detail]({{< relref "connections" >}}) for where this fits.

```yaml
connection:
  winrm:
    address:
      host: 192.168.1.20
      user: Administrator
    auth:
      method: ntlm
      passwordRef:
        fromEnv: WINDOWS_ADMIN_PASSWORD
    serverCert:
      insecureSkipVerify: true
    options:
      allowInsecureHttp: false
```

For a hands-on walkthrough (preparing the Windows target, generating a self-signed certificate, running the sample), see `examples/external-instance-winrm/README.md` in the repo.

## address

| Field | Required | Default | Description |
|-------|----------|---------|--------------|
| `host` | Yes | | Hostname or IP address |
| `port` | | `5986`, or `5985` if `options.allowInsecureHttp` is true | WinRM port |
| `user` | Yes | | Username to connect as |

## auth

The client's own credential -- how Niwashi authenticates itself to the machine.

| Field | Required | Default | Description |
|-------|----------|---------|--------------|
| `method` | | `ntlm` | Only `ntlm` is supported for now (Kerberos and CredSSP are not) |
| `passwordRef.fromEnv` / `passwordRef.fromFile` | Yes | | Password source. No plaintext password field exists by design |
| `clientCertPath` / `clientKeyPath` | | | Reserved for a future `certificate` auth method. Not usable yet |

## serverCert

How the server's TLS certificate is verified -- kept separate from `auth` because verifying the party on the other end of the connection is a different concern from the client's own credentials (mirrors `hostKey` on the [SSH side]({{< relref "connection-ssh" >}})).

| Field | Default | Description |
|-------|---------|--------------|
| `insecureSkipVerify` | `false` | Skip validating the server's certificate against a trusted CA. Needed for a self-signed certificate on a standalone host; the session is still TLS-encrypted either way |
| `caCertPath` | | Path to a specific CA certificate to trust, as an alternative to installing it in the client machine's system trust store |

## options

Connection mechanics -- not credentials or verification.

| Field | Default | Description |
|-------|---------|--------------|
| `allowInsecureHttp` | `false` | Use plain HTTP (port 5985) instead of HTTPS. Not recommended outside a lab -- see the sample README for why |
| `connectTimeoutSec` | `30` | Connection timeout |

## Constraints

- Only `ntlm` authentication is supported for now (Kerberos and CredSSP are not).
- Files larger than ~100KB can't be uploaded to the node in a single command (no chunked transfer yet).

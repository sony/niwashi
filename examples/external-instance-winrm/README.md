# WinRM (Windows node) `external-instance` sample

A minimal example of registering an already-running Windows machine (a standalone AWS EC2 Windows instance or physical machine, not domain-joined) as a node and running a PowerShell task on it over WinRM.

Unlike the [`external-instance`](../external-instance/) SSH sample, this one has no Docker-based stand-in: Windows WinRM targets can't be spun up locally the way `linuxserver/openssh-server` stands in for SSH, so you need a real Windows machine to try this against.

## 1. Prepare the Windows target

Enable WinRM over HTTPS on the target Windows machine (as Administrator):

```powershell
winrm quickconfig
winrm set winrm/config/service/Auth '@{Negotiate="true"}'

$cert = New-SelfSignedCertificate -DnsName "<the host/IP you'll put in target.yaml>" -CertStoreLocation Cert:\LocalMachine\My
New-Item -Path WSMan:\LocalHost\Listener -Transport HTTPS -Address * -CertificateThumbPrint $cert.Thumbprint -Force
New-NetFirewallRule -DisplayName "WinRM HTTPS" -Direction Inbound -LocalPort 5986 -Protocol TCP -Action Allow
```

`target.yaml` is already set up to match this (HTTPS on port `5986`, `serverCert.insecureSkipVerify: true` for the self-signed certificate). See [Notes](#notes) below for why these particular steps are needed and for alternatives (a real CA chain, plain HTTP).

## 2. Point target.yaml at your machine

Edit `address.host` in `target.yaml` -- it ships set to `192.0.2.10`, a documentation-only address (RFC 5737) that will never connect. Set it to the same host/IP you used for `-DnsName` above, and change `address.user` if you're not using the built-in `Administrator` account.

## 3. Set the password

```bash
export WINDOWS_ADMIN_PASSWORD='...'
```

`target.yaml` reads it via `auth.passwordRef.fromEnv` rather than embedding it directly, so it doesn't end up committed in the state file.

## 4. Plan and apply

```bash
nwsctl plan --target ./target.yaml --recipe-dir ./recipe --out plan.json --with-init

nwsctl apply --plan plan.json --recipe-dir ./recipe
```

This runs `recipe/nws-recipe.yaml`'s `hello-windows` task (gated by `where: node.os == "windows"`), which writes `hello.txt` to the task's output directory via PowerShell.

## Constraints

These are `nwsctl`'s constraints for WinRM/Windows nodes in general, not specific to this sample -- see [WinRM Connection Reference](../../docs/user-guide/content/en/defining-desired-state/connection-winrm.md) for the full, current list.

## Notes

### Why the HTTPS setup in step 1 looks like this

`winrm quickconfig` on its own only sets up the HTTP/5985 listener; it does not create an HTTPS/5986 listener or a certificate for you (`winrm quickconfig -transport:https` fails unless a suitable certificate already exists). The commands in step 1 create a self-signed certificate and bind it to a new HTTPS listener instead. Since that certificate is self-signed, the client can't validate it against a trusted CA -- `target.yaml` sets `serverCert.insecureSkipVerify: true` to skip that check. The session itself is still TLS-encrypted; this setting only means the client won't verify the server's identity via the certificate chain, which is fine for a lab/test target but not for one you don't control.

If you have a real CA chain instead (e.g. an internal intermediate CA), install it in the client machine's system trust store and drop `serverCert.insecureSkipVerify` -- there's also `serverCert.caCertPath` for a specific CA cert file, though the server certificate's SAN must include the exact host/IP used in `address.host` for hostname verification to pass.

If you only have HTTP/5985 available, set `options.allowInsecureHttp: true` in `target.yaml` instead of the HTTPS listener in step 1 -- not recommended outside of a lab. WinRM over HTTP additionally requires `AllowUnencrypted="true"` on the Windows side (`winrm set winrm/config/service '@{AllowUnencrypted="true"}'`) because the client library doesn't negotiate NTLM message-level encryption, so this path sends everything -- including the password exchange -- unencrypted.

### Output file encoding (BOM)

`hello.txt` is written with `Out-File`, which defaults to **UTF-16LE with a BOM** on Windows PowerShell 5.1 -- the version this sample runs under, since niwashi's default shell for Windows nodes is `powershell` (5.1, always present), not `pwsh` (PowerShell 7+, which would need a separate install and an explicit `#!pwsh` shebang in the task). Reading `hello.txt` with a non-Windows/non-PowerShell tool (`cat`, a text editor expecting UTF-8, etc.) will show the BOM as garbled leading bytes. This isn't a niwashi bug -- it's `Out-File`'s own default encoding. Use `-Encoding utf8` (still BOM-prefixed on Windows PowerShell 5.1) or `[System.IO.File]::WriteAllText(...)` (BOM-less UTF-8 on both) if you need a clean UTF-8 file.

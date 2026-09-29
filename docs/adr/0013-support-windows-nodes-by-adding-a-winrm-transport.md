---
status: accepted
date: 2026-07-30
decision-makers: Takahiro Okayama
---

# Support Windows Nodes by Adding a WinRM Transport

## Context and Problem Statement

Niwashi's node targets were reachable only over SSH. Some target machines are Windows, not Linux. Some are standalone — a cloud VM or a physical machine — not joined to an Active Directory domain.

Remote execution already goes through a protocol-independent `Transport`/`Operator`/`RemoteSession` abstraction in `internal/transport` ([ADR 0004](0004-execute-recipe-scripts-on-remote-nodes-via-a-built-in-ssh-action.md)). Recipes already branch on a target's OS via the CEL-based `where` condition ([ADR 0006](0006-filter-task-execution-targets-with-a-cel-based-where-condition.md)) and the `kind` field ([ADR 0007](0007-recipe-kind-field-replaces-spec-scope.md)).

This decision covers three questions:

* How a new transport for non-domain-joined Windows targets should relate to the existing SSH one.
* Whether it requires changes to the action layer or the recipe schema.
* Which authentication method it uses.

## Decision Drivers

* The existing `Transport`/`Operator`/`RemoteSession`/`Prober` abstraction should be reused as-is if it already generalizes to WinRM. `exec.remote` should not need to know which protocol it is talking to.
* The target environment — standalone, non-domain-joined Windows machines — treats WinRM, not SSH, as its standard remote-management protocol.

## Considered Options

* Add a WinRM `Transport` implementation under `internal/transport/winrm/`, reusing the existing `exec.remote` action unchanged (chosen)
* Stage the rollout through Ansible first (v1 delegates to Ansible, v2 replaces it with a native implementation)
* Implement SSH support for Windows targets (via OpenSSH for Windows) before WinRM
* Add a separate `exec.winrm` action distinct from `exec.remote`

## Decision Outcome

Chosen option: "Add a WinRM `Transport` implementation under `internal/transport/winrm/`, reusing the existing `exec.remote` action unchanged", because it is the only option that adds Windows support as a new implementation of the existing abstraction rather than working around or duplicating it.

`exec.remote` does not depend on which protocol is behind it; it already selects the shell/script convention for the target's OS. Reusing it means Windows support needs no new action type and no new recipe field, and no change to how a recipe author writes OS-conditional tasks — the existing `where` condition is enough. The only new work is a `Transport`/`Operator`/`RemoteSession`/`Prober` implementation, plus fixing `exec_remote.go` call sites that used the host's own path separator instead of the workspace's path abstraction.

Authentication uses NTLM only. Kerberos needs domain membership, which the target environment lacks by design; CredSSP is not implemented by the WinRM client library adopted here; Basic authentication is disabled by default on WinRM servers and sends credentials unencrypted when enabled. A password is accepted only as a reference (an environment variable name or a file path), never embedded in plaintext in a recipe or a state file. See the WinRM Connection Reference (`docs/user-guide/content/{en,ja}/defining-desired-state/connection-winrm.md`) for the full connection schema and current constraints.

### Consequences

* Good, because a recipe written against `exec.remote` for a Linux node works unchanged against a Windows node once the target's transport is `winrm` and `where` selects the right tasks.
* Good, because NTLM keeps the non-domain-joined environment usable without Kerberos or CredSSP infrastructure.
* Bad, because WinRM has no SFTP-like streaming transfer: an upload is base64-encoded into a single command, and both the client's SOAP envelope and the server's `MaxEnvelopeSizeKb` setting cap how large a single upload can be.

### Confirmation

* That `exec.remote` stays protocol-agnostic can be confirmed by reading `internal/action/exec_remote/exec_remote.go`: it reaches a target only via `Subject().GetTransport()` and `Operator.NewRemoteSession()`, branching on OS but never on transport/protocol type.
* Its tests confirm the same boundary by substituting a fake `Operator`/`RemoteSession` pair, the same pattern the SSH tests use.
* That `internal/transport/winrm/` implements the same contract as SSH can be confirmed by reading it alongside `internal/transport/ssh/`: both implement `Transport`/`Operator`/`RemoteSession`/`Prober` and register under their own type name (`ssh`/`winrm`).
* Its own tests confirm niwashi's command construction, escaping, and exit-code handling by substituting a fake behind the WinRM client library's interface, not by exercising the wire-level SOAP/WS-Management protocol.

## Pros and Cons of the Options

### Add a WinRM `Transport` implementation, reusing `exec.remote` unchanged

See Consequences above.

* Good, because it keeps the SSH and WinRM implementations structurally parallel (same directory shape, interfaces, test style), easing comparison and maintenance.
* Neutral, because it is not purely additive: it also requires fixing pre-existing `exec_remote.go` plumbing that had bypassed the workspace path abstraction.

### Stage the rollout through Ansible first, then replace with a native implementation

* Good, because Ansible's Windows support is mature and could have de-risked an initial rollout.
* Bad, because the `Transport`/`Operator`/`RemoteSession` abstraction already made a native implementation no more costly than integrating Ansible, so a two-phase rollout would mean maintaining, then discarding, an Ansible integration path for no benefit.

### Implement SSH support for Windows targets first

* Bad, because the target environment (standalone Windows machines) treats WinRM as the standard remote-management protocol. An OpenSSH server cannot be assumed present or maintained on such a target.
* Neutral, because this option is deferred, not rejected outright: it stays available as a follow-up if some target environment needs it.

### Add a separate `exec.winrm` action

* Good, because it would keep all WinRM-specific behavior in one dedicated action, arguably easier to reason about in isolation.
* Bad, because `exec.remote` already does not depend on the transport's protocol type. A separate action would duplicate shared machinery (workspace setup, file transfer, state-change application) that `exec.remote` already provides generically.
* Bad, because it would require a recipe author to know whether a target will be reached over SSH or WinRM. That information should live in the transport configuration, not the task's action.

## More Information

`serverCert` (server-certificate verification) was split out of `options` during review, matching the same split on the SSH side rather than folding certificate verification into general connection mechanics. Certificate-based client authentication is deferred, not rejected; its fields are reserved but unused. See the WinRM Connection Reference cited above for the full field list.

Whether NTLM remains the right default as the protocol landscape evolves is worth watching. This decision does not resolve that question.

Streaming stdout/stderr in real time works via the WinRM client library's own run loop. Niwashi needs no extra mechanism for it.

Background and further detail on the protocol and library choices referenced above:

* [Authentication for Remote Connections — Microsoft Learn](https://learn.microsoft.com/en-us/windows/win32/winrm/authentication-for-remote-connections)
* [Security considerations for PowerShell Remoting using WinRM — Microsoft Learn](https://learn.microsoft.com/en-us/powershell/scripting/security/remoting/winrm-security)
* [masterzen/winrm](https://github.com/masterzen/winrm)
* [Windows Remote Management — Ansible Community Documentation](https://docs.ansible.com/ansible/latest/os_guide/windows_winrm.html)

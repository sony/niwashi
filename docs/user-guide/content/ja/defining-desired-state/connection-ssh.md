---
title: "SSH接続リファレンス"
weight: 10
---

# SSH接続リファレンス

`connection.ssh`は、`external-instance` provisionerで登録するLinux/Unixマシンへの到達方法を指定します。このページの位置づけは[接続の詳細]({{< relref "connections" >}})を参照してください。

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

| フィールド | 必須 | 既定値 | 説明 |
|-----------|------|--------|------|
| `host` | 要 | | ホスト名またはIPアドレス |
| `port` | | `22` | SSHのポート番号 |
| `user` | 要 | | 接続するユーザー名 |

## auth

クライアント自身の資格情報 -- niwashiがそのマシンに対して自分自身を証明する方法。

| フィールド | 必須 | 既定値 | 説明 |
|-----------|------|--------|------|
| `method` | | `privateKey` | `privateKey`のみサポート |
| `privateKeyPath` | 要 | | 秘密鍵ファイルのパス |
| `passphraseRef.fromEnv` / `passphraseRef.fromFile` | | | 秘密鍵が暗号化されている場合のパスフレーズ。設計上、平文のパスフレーズフィールドは存在しない |

## hostKey

接続先マシンの身元をどう検証するか -- `auth`とは分離されている。接続相手を検証することと、クライアント自身の資格情報は別の関心事のため。

| フィールド | 必須 | 既定値 | 説明 |
|-----------|------|--------|------|
| `method` | | `knownHostsFile` | `knownHostsFile`のみサポート |
| `knownHostsPath` | | | 対象ホストの鍵が既に登録済みの、標準的なOpenSSHの`known_hosts`ファイルへのパス（事前の手動接続や`ssh-keyscan`等で用意しておく）。未知のホストは自動的に信頼されず、拒否される |

## 接続の機構（タイムアウト・リトライ）

`address`/`auth`/`hostKey`と異なり、これらは`connection.ssh`ごとに設定するものではなく、プロセス全体に効く環境変数です:

| 環境変数 | 既定値 | 説明 |
|---------|--------|------|
| `NWS_SSH_CONNECT_TIMEOUT_SEC` | `10` | TCP接続のタイムアウト |
| `NWS_SSH_HANDSHAKE_TIMEOUT_SEC` | `10` | SSHハンドシェイクのタイムアウト |
| `NWS_SSH_RETRY_MAX_COUNT` | `3` | 接続リトライの最大回数 |
| `NWS_SSH_RETRY_INTERVAL_SEC` | `5` | リトライ間隔 |

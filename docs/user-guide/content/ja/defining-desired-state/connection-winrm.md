---
title: "WinRM接続リファレンス"
weight: 11
---

# WinRM接続リファレンス

`connection.winrm`は、`external-instance` provisionerで登録するWindowsマシン（スタンドアロン・ドメイン非参加。AWS EC2 Windowsインスタンスや物理マシンを想定）への到達方法を指定します。このページの位置づけは[接続の詳細]({{< relref "connections" >}})を参照してください。

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

Windowsターゲットの準備・自己署名証明書の作成・サンプルの実行という一連の手順は、リポジトリの`examples/external-instance-winrm/README.md`を参照してください。

## address

| フィールド | 必須 | 既定値 | 説明 |
|-----------|------|--------|------|
| `host` | 要 | | ホスト名またはIPアドレス |
| `port` | | `5986`（`options.allowInsecureHttp`が`true`の場合は`5985`） | WinRMのポート番号 |
| `user` | 要 | | 接続するユーザー名 |

## auth

クライアント自身の資格情報 -- niwashiがそのマシンに対して自分自身を証明する方法。

| フィールド | 必須 | 既定値 | 説明 |
|-----------|------|--------|------|
| `method` | | `ntlm` | 現時点では`ntlm`のみサポート（KerberosとCredSSPは非対応） |
| `passwordRef.fromEnv` / `passwordRef.fromFile` | 要 | | パスワードの取得元。設計上、平文のパスワードフィールドは存在しない |
| `clientCertPath` / `clientKeyPath` | | | 将来の`certificate`認証方式向けに予約済み。まだ使用不可 |

## serverCert

サーバーのTLS証明書をどう検証するか -- `auth`とは分離されている。接続相手を検証することと、クライアント自身の資格情報は別の関心事のため（[SSH側]({{< relref "connection-ssh" >}})の`hostKey`と同じ位置づけ）。

| フィールド | 既定値 | 説明 |
|-----------|--------|------|
| `insecureSkipVerify` | `false` | サーバー証明書を信頼済みCAに対して検証する処理をスキップする。スタンドアロンホストで自己署名証明書を使う場合に必要（通信自体はどちらの場合もTLSで暗号化されたまま） |
| `caCertPath` | | 信頼させる特定のCA証明書へのパス。クライアントマシンのシステム証明書ストアにインストールする代わりに使える |

## options

接続の機構 -- 資格情報や検証とは別の関心事。

| フィールド | 既定値 | 説明 |
|-----------|--------|------|
| `allowInsecureHttp` | `false` | HTTPS の代わりに平文HTTP（ポート5985）を使う。lab環境以外では非推奨（理由はサンプルのREADME.mdを参照） |
| `connectTimeoutSec` | `30` | 接続タイムアウト |

## 制約

- 現時点では`ntlm`認証のみサポート（KerberosとCredSSPは非対応）。
- 一括アップロードでは約100KBを超えるファイルを転送できない（チャンク転送は未実装）。

---
title: "接続の詳細"
weight: 9
---

# 接続の詳細

このページは、`external-instance` provisioner（[ジェネレーターの詳細]({{< relref "generators" >}})を参照）を使って既に起動しているマシンを登録し、`connection`ブロックを自分で書く人向けです。

```yaml
infrastructure:
  generators:
    existing-servers:
      provisioner: external-instance
      params:
        instances:
          server-01:
            connection:
              ssh:      # または winrm
                ...
```

各インスタンスの`connection`は、niwashiがそのマシンにどうやって到達するか（接続先アドレス、クライアントとしての認証方法、winrmの場合はサーバーの検証方法）を指定します。niwashiは2つの接続タイプをサポートしています:

- [SSH接続リファレンス]({{< relref "connection-ssh" >}}) -- Linux/Unixマシン向け
- [WinRM接続リファレンス]({{< relref "connection-winrm" >}}) -- Windowsマシン（ドメイン非参加）向け

それぞれの最小限の例は[Infrastructure の基本]({{< relref "basic-concepts/infrastructure" >}})を参照してください。

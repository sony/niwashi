---
title: "Infrastructure の基本"
weight: 3
---

# Infrastructure の基本

`infrastructure` セクションでは、物理的/仮想的なインフラストラクチャの生成方法を定義します。「どこにデプロイするか」を記述する場所です。

## Infrastructure の構造

```yaml
infrastructure:
  generators:
    # インスタンス生成器の定義
```

- **generators**: インスタンス（仮想マシン、コンテナ、既存サーバーなど）を生成・管理するための生成器の定義

---

## Generators（生成器）

ジェネレーターは、インスタンスを生成・管理するための定義です。各ジェネレーターには、どのようにインスタンスを生成するかを指定する `provisioner` を必ず指定します。

### 基本的な定義

最小限のジェネレーター定義：

```yaml
infrastructure:
  generators:
    my-generator:
      provisioner: external-instance
```

この例では、`external-instance` という組み込みprovisionerを使用しています。

### ジェネレーターの識別子

各ジェネレーターには、一意な識別子（キー）を指定します：

```yaml
infrastructure:
  generators:
    web-vms: {}
    db-vms: {}
    cache-vms: {}
```

識別子の命名規則については、[識別子の命名規則]({{< relref "naming-rules" >}}) を参照してください。

---

## ジェネレーターの属性

### provisioner（必須）

インスタンスをどのように生成するかを指定するレシピIDです。provisionerの種類によって、生成されるインスタンスの性質が決まります。

```yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
```

provisionerのレシピID指定方法については、[Capabilityとレシピの詳細]({{< relref "../capabilities" >}}) を参照してください。

### params

provisioner固有のパラメータを指定します。パラメータの内容は、使用するprovisionerのレシピによって異なります。

```yaml
infrastructure:
  generators:
    vagrant-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

---

## インスタンスとは

**インスタンス**は、ノードに割り当てられる実際の計算リソースです：

- 仮想マシン（VM）
- コンテナ
- 既存の物理サーバー
- クラウドインスタンス

ジェネレーターは、これらのインスタンスを生成・管理します。

### インスタンスとノードの関係

```
┌─────────────────┐
│  Inventory      │
│  ┌───────────┐  │
│  │ Node      │  │  論理的な計算単位
│  └─────┬─────┘  │
└────────┼────────┘
         │ 割り当て
         ↓
┌────────┼────────┐
│  Infrastructure│
│  ┌─────┴─────┐ │
│  │ Instance  │ │  実際の計算リソース
│  └───────────┘ │
└────────────────┘
```

- **Node**: 論理的な計算単位（「Webサーバーが必要」という概念）
- **Instance**: 実際の計算リソース（「192.168.1.10の仮想マシン」という実体）

Niwashiは、ジェネレーターが生成したインスタンスをノードに自動的に割り当てます。

---

## 組み込みprovisioner

Niwashiには、特別な組み込みprovisionerが用意されています。

### external-instance

既に起動している外部のマシンやサーバーをインスタンスとして利用する場合に使用します。

```yaml
infrastructure:
  generators:
    existing-servers:
      provisioner: external-instance
      params:
        instances:
          server-01:
            connection:
              ssh:
                address:
                  host: 192.168.1.10
                  port: 22
                  user: ubuntu
                auth:
                  method: privateKey
                  privateKeyPath: ~/.ssh/id_rsa
          server-02:
            connection:
              ssh:
                address:
                  host: 192.168.1.11
                  port: 22
                  user: ubuntu
                auth:
                  method: privateKey
                  privateKeyPath: ~/.ssh/id_rsa
```

この例では、既に起動している2台のサーバー（`server-01` と `server-02`）をインスタンスとして登録しています。詳細なフィールドリファレンスは[接続の詳細]({{< relref "../connections" >}}) → [SSH接続リファレンス]({{< relref "../connection-ssh" >}})を参照してください。

Windowsノードは`ssh`の代わりに`winrm`接続タイプを使用します：

```yaml
infrastructure:
  generators:
    existing-servers:
      provisioner: external-instance
      params:
        instances:
          windows-server-01:
            connection:
              winrm:
                address:
                  host: 192.168.1.20
                  user: Administrator
                auth:
                  method: ntlm
                  passwordRef:
                    fromEnv: WINDOWS_ADMIN_PASSWORD
```

この例では、Windowsマシン（スタンドアロン・ドメイン非参加のホスト）をインスタンスとして登録しています。詳細なフィールドリファレンスと制約は[接続の詳細]({{< relref "../connections" >}}) → [WinRM接続リファレンス]({{< relref "../connection-winrm" >}})を参照してください。

### external-instance の用途

- 既存のクラウドインスタンスを利用
- 社内の既存サーバーを利用
- 手動でセットアップした環境を管理対象に追加

---

## リファレンスprovisioner

Niwashiは、組み込み以外にも、リファレンスとなるprovisionerレシピを提供しています。

### Vagrant

ローカルで仮想マシンを生成します：

```yaml
infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

#### Vagrant provisioner のパラメータ例

- **count**: 生成するインスタンスの数
- **box**: 使用するVagrant box
- **cpus**: CPU数
- **memory**: メモリ（MB）
- **disk_size**: ディスクサイズ
- **network**: ネットワーク設定

詳細は、Vagrantレシピのドキュメントを参照してください。

---

## ジェネレーターの使い方

### 基本的な使い方

1. ジェネレーターを定義
2. ノードから参照（オプション）

```yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3

inventory:
  nodes:
    node1:
      instanceSelector:
        generator: vms  # ジェネレーターを明示的に指定
    node2: {}           # 自動選択
    node3: {}           # 自動選択
```

### 複数のジェネレーター

異なる目的で複数のジェネレーターを定義できます：

```yaml
infrastructure:
  generators:
    # Webサーバー用（スペック低め）
    web-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
        box: ubuntu/jammy64
        cpus: 1
        memory: 1024

    # DBサーバー用（スペック高め）
    db-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 1
        box: ubuntu/jammy64
        cpus: 4
        memory: 4096

inventory:
  nodes:
    web-01:
      instanceSelector:
        generator: web-vms
    web-02:
      instanceSelector:
        generator: web-vms
    db-01:
      instanceSelector:
        generator: db-vms
```

---

## 実践例

### ローカル開発環境

```yaml
infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048

inventory:
  nodes:
    web: {}
    app: {}
    db: {}
```

### 既存サーバーの管理

```yaml
infrastructure:
  generators:
    prod-servers:
      provisioner: external-instance
      params:
        instances:
          web-prod:
            connection:
              ssh:
                address:
                  host: web.example.com
                  port: 22
                  user: deploy
                auth:
                  method: privateKey
                  privateKeyPath: ~/.ssh/deploy_key
          db-prod:
            connection:
              ssh:
                address:
                  host: db.example.com
                  port: 22
                  user: deploy
                auth:
                  method: privateKey
                  privateKeyPath: ~/.ssh/deploy_key

inventory:
  nodes:
    web:
      instanceSelector:
        generator: prod-servers
    db:
      instanceSelector:
        generator: prod-servers
```

### 混在環境

開発環境と本番環境で異なるprovisionerを使用：

#### base.yaml（共通の論理構成）

```yaml
version: nws.state/v1

inventory:
  nodes:
    web:
      capabilities:
        - web.nginx
    db:
      capabilities:
        - database.postgresql
```

#### infra-dev.yaml（開発環境）

```yaml
version: nws.state/v1

infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
        box: ubuntu/jammy64
```

#### infra-prod.yaml（本番環境）

```yaml
version: nws.state/v1

infrastructure:
  generators:
    prod-servers:
      provisioner: external-instance
      params:
        instances:
          web-prod:
            connection:
              ssh:
                address:
                  host: web.example.com
                  port: 22
                  user: deploy
          db-prod:
            connection:
              ssh:
                address:
                  host: db.example.com
                  port: 22
                  user: deploy
```

#### 使い分け

```bash
# 開発環境
nwsctl plan -t base.yaml -t infra-dev.yaml

# 本番環境
nwsctl plan -t base.yaml -t infra-prod.yaml
```

---

## インスタンスの割り当て

ノードにインスタンスを割り当てる方法は複数あります。

### 自動割り当て

`instanceSelector` を指定しない場合、利用可能なインスタンスを自動的に選択します：

```yaml
inventory:
  nodes:
    node1: {}  # 自動割り当て
    node2: {}  # 自動割り当て
```

### ジェネレーター指定

特定のジェネレーターが生成したインスタンスから選択します：

```yaml
inventory:
  nodes:
    web-node:
      instanceSelector:
        generator: web-vms
```

詳細は [ノードの詳細]({{< relref "../nodes" >}}) を参照してください。

---

## 次のステップ

- [識別子の命名規則]({{< relref "naming-rules" >}}) - ジェネレーターの命名ルール
- [ジェネレーターの詳細]({{< relref "../generators" >}}) - ジェネレーターの全属性とオプション
- [複数環境の管理]({{< relref "../advanced/multi-environment" >}}) - 環境ごとのインフラ切り替え

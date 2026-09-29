---
title: "ジェネレーターの詳細"
weight: 8
---

# ジェネレーターの詳細

このページでは、ジェネレーター（インスタンス生成器）の全属性と使い方について詳しく説明します。

## ジェネレーターとは

ジェネレーターは、インスタンス（仮想マシン、コンテナ、既存サーバーなど）を生成・管理するための定義です。各ジェネレーターには、どのようにインスタンスを生成するかを指定する `provisioner` を必ず指定します。

---

## 基本的なジェネレーターの定義

ジェネレーターには、識別子（キー）と、最低限 `provisioner` を指定する必要があります。

```yaml
infrastructure:
  generators:
    my-generator:
      provisioner: external-instance
```

---

## ジェネレーターの属性

ジェネレーターには以下の属性を指定できます。

### provisioner（必須）

インスタンスをどのように生成するかを指定するレシピIDです。provisionerの種類によって、生成されるインスタンスの性質が決まります。

```yaml
infrastructure:
  generators:
    vm-generator:
      provisioner: infra.vm.driver=vagrant
```

provisionerのレシピID指定方法については、[Capabilityとレシピの詳細]({{< relref "capabilities" >}}) を参照してください。

#### provisionerの種類

- **組み込みprovisioner**: Niwashiに組み込まれている特別なprovisioner
- **リファレンスprovisioner**: Niwashiが提供するリファレンスとなるprovisionerレシピ
- **カスタムprovisioner**: ユーザーが作成したprovisionerレシピ

---

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

#### パラメータの確認方法

各provisionerがどのようなパラメータを受け取るかは、provisionerレシピのドキュメントを参照してください。

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
                hostKey:
                  knownHostsPath: ~/.ssh/known_hosts
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
                hostKey:
                  knownHostsPath: ~/.ssh/known_hosts
```

#### params の構造

- **instances**: インスタンスの定義（オブジェクト）
  - キーはインスタンスID（任意の識別子）
  - 値は接続情報を含むオブジェクト

#### connection の構造

- **type**: 接続タイプ（現在は `ssh` のみ）
- **address**: 接続先情報
  - **host**: ホスト名またはIPアドレス
  - **port**: ポート番号（デフォルト: 22）
  - **user**: ユーザー名
- **auth**: 認証情報
  - **method**: 認証方式（現在は `privateKey` のみ）
  - **privateKeyPath**: 秘密鍵のパス

#### 用途

- 既存のクラウドインスタンスを利用
- 社内の既存サーバーを利用
- 手動でセットアップした環境を管理対象に追加

---

## リファレンスprovisioner

Niwashiは、組み込み以外にも、リファレンスとなるprovisionerレシピを提供しています。

### Vagrant

ローカルで仮想マシンを生成します。

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

#### 主なパラメータ

- **count**: 生成するインスタンスの数
- **box**: 使用するVagrant box
- **cpus**: CPU数
- **memory**: メモリ（MB）
- **disk_size**: ディスクサイズ（オプション）
- **network**: ネットワーク設定（オプション）

#### 用途

- ローカル開発環境の構築
- テスト環境の構築
- デモ環境の構築

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

    node2: {}  # 自動選択
    node3: {}  # 自動選択
```

### インスタンスの自動割り当て

Niwashiは、ジェネレーターが生成したインスタンスを、ノードに自動的に割り当てます。

```
┌─────────────────────┐
│  Generator          │
│  provisioner: ...   │
│  params:            │
│    count: 3         │
└──────────┬──────────┘
           │ 生成
           ↓
┌──────────┴──────────┐
│  Instances (3個)    │
│  - instance-1       │
│  - instance-2       │
│  - instance-3       │
└──────────┬──────────┘
           │ 割り当て
           ↓
┌──────────┴──────────┐
│  Nodes (3個)        │
│  - node1            │
│  - node2            │
│  - node3            │
└─────────────────────┘
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

## ジェネレーター定義の実践例

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
        network:
          type: private_network
          ip_prefix: 192.168.56
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
                hostKey:
                  knownHostsPath: ~/.ssh/known_hosts

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
                hostKey:
                  knownHostsPath: ~/.ssh/known_hosts
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

### テンプレートの使用

```yaml
template:
  instance:
    standard-vm:
      provisioner: infra.vm.driver=vagrant
      params:
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048

infrastructure:
  generators:
    web-vms:
      templates: [standard-vm]
      params:
        count: 2

    db-vms:
      templates: [standard-vm]
      params:
        count: 1
        memory: 4096  # メモリだけ上書き
```

---

## よくあるパターン

### パターン1: 均一なVM群

すべて同じスペックのVMを生成：

```yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 5
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

### パターン2: 役割別のVM群

役割ごとに異なるスペックのVMを生成：

```yaml
infrastructure:
  generators:
    control-plane-vm:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 1
        box: ubuntu/jammy64
        cpus: 4
        memory: 4096

    worker-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

### パターン3: 既存サーバーとVMの混在

```yaml
infrastructure:
  generators:
    # 既存の本番サーバー
    prod-servers:
      provisioner: external-instance
      params:
        instances:
          prod-db:
            connection:
              ssh:
                address:
                  host: db.prod.example.com
                  port: 22
                  user: deploy

    # 開発用のVM
    dev-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
        box: ubuntu/jammy64
```

---

## トラブルシューティング

### インスタンスが生成されない

**問題**: ジェネレーターを定義したが、インスタンスが生成されない。

**原因**:
- provisionerのレシピIDが間違っている
- 必須パラメータが不足している
- provisionerが正しくインストールされていない

**解決策**:
1. provisionerのレシピIDを確認する
2. provisionerのドキュメントで必須パラメータを確認する
3. レシピが正しくインストールされているか確認する

### インスタンスの数が足りない

**問題**: ノードの数に対してインスタンスが足りず、割り当てに失敗する。

**原因**:
- `params.count` が不足している
- 複数のジェネレーターを使用しているが、合計数が足りない

**解決策**:
1. `params.count` を増やす
2. ノードの数と生成されるインスタンスの数を確認する

```yaml
# Bad: ノード3つに対してインスタンス2つ
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2  # 不足

inventory:
  nodes:
    node1: {}
    node2: {}
    node3: {}  # 割り当てられない

# Good: ノード3つに対してインスタンス3つ
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3  # OK
```

### 接続できない（external-instance）

**問題**: external-instanceで既存サーバーを指定したが、接続できない。

**原因**:
- ホスト名/IPアドレスが間違っている
- ポート番号が間違っている
- 秘密鍵のパスが間違っている
- 秘密鍵の権限が不適切

**解決策**:
1. 接続情報を確認する
2. 手動でSSH接続を試してみる
3. 秘密鍵のパーミッションを確認する（`chmod 600`）

---

## 次のステップ

- [ノードの詳細]({{< relref "nodes" >}}) - ノードの全属性とインスタンス割り当て
- [Capabilityとレシピの詳細]({{< relref "capabilities" >}}) - provisionerの指定方法
- [テンプレートの詳細]({{< relref "templates" >}}) - ジェネレーターのテンプレート活用
- [複数環境の管理]({{< relref "advanced/multi-environment" >}}) - 環境ごとのインフラ切り替え

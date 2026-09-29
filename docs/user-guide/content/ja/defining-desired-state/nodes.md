---
title: "ノードの詳細"
weight: 6
---

# ノードの詳細

このページでは、ノードの全属性と使い方について詳しく説明します。

## ノードとは

ノードは、システムを構成する個々の論理的な計算単位です。各ノードには識別子（キー）と、そのノードが持つべき属性を指定します。

---

## 基本的なノードの定義

最もシンプルなノード定義は、識別子のみを指定する形式です：

```yaml
inventory:
  nodes:
    node1: {}
    node2: {}
    node3: {}
```

---

## ノードの属性

ノードには以下の属性を指定できます。

### labels

ノードにラベル（メタデータ）を付与します。ラベルは文字列のキー・バリューペアで指定します。

```yaml
inventory:
  nodes:
    web-server:
      labels:
        role: web
        environment: production
        region: us-east-1
```

#### 用途

- ノードの分類や識別（`where` フィールドによるタスクの絞り込み）
- メタデータの付与

ここで定義したラベルはレシピのタスクの `where` フィールドで参照できます。

```yaml
# レシピ側で node.labels.role を参照する例
tasks:
  - name: configure-web
    where: "node.labels.role == 'web'"
    action:
      exec.remote:
        scriptTpl: |
          echo "Configuring web server"
```

詳細は [タスクのフィルタリング条件 (where)]({{< relref "../../../defining-recipes/task-where" >}}) を参照してください。

---

### capabilities

ノードが持つべき機能（Capability）を指定します。Capabilityは、レシピによって実現される機能の単位です。

#### 簡略形（文字列のみ）

```yaml
inventory:
  nodes:
    db-server:
      capabilities:
        - database.postgresql
        - monitoring.prometheus-node-exporter
```

#### 詳細形（パラメータを指定）

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - id: web.nginx
          params:
            port: 8080
            worker_processes: 4
```

#### 複数のCapabilityを指定

ノードには、複数のCapabilityを指定できます：

```yaml
inventory:
  nodes:
    app-server:
      capabilities:
        - runtime.python
        - web.nginx
        - monitoring.prometheus-exporter
        - logging.fluentd
```

Capabilityは、指定された順に適用されます。

#### Capabilityの詳細

Capabilityとレシピの指定方法の詳細は、[Capabilityとレシピの詳細]({{< relref "capabilities" >}}) を参照してください。

---

### templates

テンプレートを継承してノードを定義できます。複数のテンプレートを配列で指定可能です。

```yaml
template:
  node:
    base:
      labels:
        environment: production

inventory:
  nodes:
    web-server:
      templates: [base]
      capabilities:
        - web.nginx
```

#### 複数テンプレートの適用

```yaml
template:
  node:
    base:
      labels:
        managed_by: niwashi

    web-base:
      capabilities:
        - monitoring.prometheus-exporter

inventory:
  nodes:
    web-01:
      templates: [base, web-base]
      capabilities:
        - web.nginx
```

後に指定したテンプレートが前のテンプレートの設定を上書きします。

テンプレートの詳細は、[テンプレートの詳細]({{< relref "templates" >}}) を参照してください。

---

### instanceSelector

ノードは、実際の物理マシンや仮想マシン（インスタンス）に紐付ける必要があります。この紐付けを制御するのが `instanceSelector` です。

#### 自動選択（指定なし）

`instanceSelector` を指定しない場合、Niwashiは利用可能な（まだ割り当てられていない）インスタンスを自動的に選択します。

```yaml
inventory:
  nodes:
    node1: {}
    node2: {}
```

#### generator指定

特定のジェネレーターが生成したインスタンスから自動選択します：

```yaml
inventory:
  nodes:
    vm-node:
      instanceSelector:
        generator: my-vm-generator
```

この例では、`my-vm-generator` が生成した未割り当てのインスタンスが自動的に選択されます。

#### generator + instance指定

特定のジェネレーターの特定のインスタンスを明示的に指定できます：

```yaml
inventory:
  nodes:
    specific-node:
      instanceSelector:
        generator: my-vm-generator
        instance: vm-01
```

#### 割り当て失敗時の動作

インスタンスの割り当てに失敗した場合（利用可能なインスタンスがない、指定されたジェネレーターが存在しないなど）、Niwashiはエラーを返します。計画の実行は中断されます。

---

## ノードとインスタンスの関係

### 概念

```
┌───────────────────┐
│  Node（論理）      │
│  web-server       │  論理的な計算単位
└─────────┬─────────┘
          │ 割り当て
          ↓
┌─────────┴─────────┐
│  Instance（実体）  │
│  192.168.1.10     │  実際の計算リソース
└───────────────────┘
```

- **Node**: 「Webサーバーが必要」という論理的な要求
- **Instance**: 「192.168.1.10の仮想マシン」という物理的な実体

### 自動割り当ての仕組み

Niwashiは以下の優先順位でノードへのインスタンス割り当てを処理します。より具体的な指定があるノードを先に処理することで、明示指定のないノードへの干渉を防ぎます。

1. `generator` と `instance` の両方を指定したノードを先に処理（特定インスタンスを確保）
2. `generator` のみを指定したノードを次に処理（そのジェネレーターの未割り当てインスタンスから選択）
3. `instanceSelector` を指定しないノードを最後に処理（すべての未割り当てインスタンスから選択）

利用可能なインスタンスがない場合はエラーになります。

### 例

```yaml
infrastructure:
  generators:
    web-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3  # 3つのインスタンスを生成

    db-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 1  # 1つのインスタンスを生成

inventory:
  nodes:
    web-primary:
      instanceSelector:
        generator: web-vms
        instance: nws-vm-0  # web-vmsの特定インスタンスを指定（先に確保される）

    web-01:
      instanceSelector:
        generator: web-vms  # web-vmsの未割り当てインスタンスから自動選択

    web-02:
      instanceSelector:
        generator: web-vms  # web-vmsの未割り当てインスタンスから自動選択

    db-01:
      instanceSelector:
        generator: db-vms   # db-vmsの未割り当てインスタンスから自動選択
```

---

## ノード定義の実践例

### シンプルなノード

```yaml
inventory:
  nodes:
    worker-01: {}
```

識別子のみ。属性はすべてデフォルト。

### ラベル付きノード

```yaml
inventory:
  nodes:
    web-server:
      labels:
        role: web
        tier: frontend
        environment: production
```

メタデータを付与。

### Capability付きノード

```yaml
inventory:
  nodes:
    db-server:
      capabilities:
        - database.postgresql
        - monitoring.prometheus-exporter
```

複数のCapabilityを指定。

### パラメータ付きCapability

```yaml
inventory:
  nodes:
    app-server:
      capabilities:
        - id: runtime.python
          params:
            version: "3.11"
        - id: web.nginx
          params:
            port: 8080
            worker_processes: 4
```

各Capabilityにパラメータを渡す。

### ジェネレーター指定

```yaml
inventory:
  nodes:
    high-spec-node:
      capabilities:
        - database.postgresql
      instanceSelector:
        generator: high-spec-vms
```

特定のジェネレーターのインスタンスを使用。

### テンプレート使用

```yaml
template:
  node:
    production-base:
      labels:
        environment: production
        managed_by: niwashi

inventory:
  nodes:
    web-01:
      templates: [production-base]
      capabilities:
        - web.nginx
```

テンプレートで共通設定を継承。

### 完全な例

すべての属性を使用した例：

```yaml
template:
  node:
    base:
      labels:
        environment: production

inventory:
  nodes:
    app-server:
      templates: [base]
      labels:
        role: application
        tier: backend
      capabilities:
        - id: runtime.python
          params:
            version: "3.11"
        - id: web.nginx
          params:
            port: 8080
        - monitoring.prometheus-exporter
      instanceSelector:
        generator: app-vms
```

---

## よくあるパターン

### Webサーバーノード

```yaml
inventory:
  nodes:
    web-01:
      labels:
        role: web
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter
```

### データベースノード

```yaml
inventory:
  nodes:
    db-primary:
      labels:
        role: database
        replication: primary
      capabilities:
        - id: database.postgresql
          params:
            version: "15"
            max_connections: 200
```

### アプリケーションノード

```yaml
inventory:
  nodes:
    app-server:
      labels:
        role: application
      capabilities:
        - runtime.python
        - web.gunicorn
        - monitoring.prometheus-exporter
```

### キャッシュノード

```yaml
inventory:
  nodes:
    cache-server:
      labels:
        role: cache
      capabilities:
        - id: cache.redis
          params:
            max_memory: "1GB"
```

---

## トラブルシューティング

### インスタンスが割り当てられない

**問題**: ノードにインスタンスが割り当てられず、エラーが発生する。

**原因**:
- 利用可能なインスタンスが不足している
- `instanceSelector.generator` で指定したジェネレーターが存在しない
- ジェネレーターが十分な数のインスタンスを生成していない

**解決策**:
1. ジェネレーターの `params.count` を増やす
2. `instanceSelector.generator` の指定を確認する
3. ジェネレーターの定義を確認する

### Capabilityが適用されない

**問題**: ノードにCapabilityを指定したが、期待通りに動作しない。

**原因**:
- レシピIDが間違っている
- パラメータが不足している
- レシピが見つからない

**解決策**:
1. レシピIDの綴りを確認する
2. レシピのドキュメントで必須パラメータを確認する
3. レシピが正しくインストールされているか確認する

---

## 次のステップ

- [クラスタの詳細]({{< relref "clusters" >}}) - 複数ノードで構成されるクラスタ
- [ジェネレーターの詳細]({{< relref "generators" >}}) - インスタンスの生成方法
- [Capabilityとレシピの詳細]({{< relref "capabilities" >}}) - レシピの指定方法
- [テンプレートの詳細]({{< relref "templates" >}}) - テンプレートの活用

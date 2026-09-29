---
title: "Capabilities とレシピの指定"
weight: 5
---

# Capabilities とレシピの指定

ノード、クラスタ、ジェネレーターで指定する機能（Capability）やprovisioner は、**レシピ**によって実現されます。レシピは、特定の機能を提供するための実行可能な定義です。

このページでは、レシピをどのように指定するかを説明します。

## Capabilityとは

**Capability**は、ノードやクラスタが持つべき機能を表します：

- Webサーバー（Nginx, Apache）
- データベース（PostgreSQL, MySQL）
- コンテナオーケストレーター（Kubernetes）
- 監視ツール（Prometheus）

Capabilityは、レシピによって実現されます。レシピは、その機能をセットアップするための手順を定義しています。

---

## レシピIDの指定方法

レシピを指定する方法は、主に2つあります。

### 1. metadata.id による指定

レシピの `metadata.id` を直接指定する方法です。これは最も明示的で、特定のレシピを確実に指定できます。

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - com.example.nginx-server
```

#### バージョン指定

`@` を使ってバージョンを明示的に指定できます：

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - com.example.nginx-server@1.2.0
```

バージョンを指定しない場合、利用可能な最新バージョンが使用されます。

適用済みのCapabilityのバージョンを変更した場合の動作は次のとおりです。

- バージョンを上げた場合、またはバージョン指定を外して新しいバージョンが選ばれた場合、Capabilityは更新されます。新しいバージョンの `operation: update` のタスクが実行されます（[更新処理（operation: update）]({{< relref "defining-recipes/defining-tasks#更新処理operation-update" >}}) 参照）。
- 適用済みのバージョンより低いバージョンを指定すると、`nwsctl plan` がエラーで失敗します。

#### 用途

- 特定のレシピを確実に指定したい場合
- バージョンを固定したい場合
- カスタムレシピを使用する場合

---

### 2. spec.provide のエイリアス名による指定

レシピは `spec.provide` でエイリアス名を定義できます。このエイリアス名を使うことで、実装の詳細を隠蔽し、より抽象的な指定が可能になります。

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.server  # エイリアス名で指定
```

#### attrsによる絞り込み

エイリアス名だけでは複数のレシピが該当する場合、`attrs`（属性）を使って絞り込めます：

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.server.engine=nginx  # engine属性がnginxのものを選択

    app-server:
      capabilities:
        - web.server.engine=apache  # engine属性がapacheのものを選択
```

#### 複数属性の指定

複数の属性を `.` で連結することもできます：

```yaml
inventory:
  nodes:
    db-server:
      capabilities:
        - database.type=relational.engine=postgresql
```

この例では、以下の条件を満たすレシピを選択します：
- `type` 属性が `relational`
- `engine` 属性が `postgresql`

#### 用途

- 実装の詳細を意識せず、抽象的に機能を指定したい場合
- 複数の実装から柔軟に選択したい場合
- レシピの入れ替えを容易にしたい場合

---

## 簡略形と詳細形

Capabilityは、簡略形（文字列のみ）と詳細形（オブジェクト）の2つの記法があります。

### 簡略形

パラメータが不要な場合は、文字列のみで指定できます：

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter
```

### 詳細形

パラメータを渡す場合は、オブジェクト形式で記述します：

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - id: web.nginx
          params:
            port: 8080
            worker_processes: 4
            ssl_enabled: true
```

#### 必須フィールド

- **id**: レシピの識別子（metadata.id またはエイリアス名）

#### オプションフィールド

- **params**: レシピに渡すパラメータ

---

## パラメータ（params）

レシピには、動作をカスタマイズするためのパラメータを渡すことができます。パラメータの構造や意味は、各レシピの仕様によって異なります。

### ノードのCapabilityでのパラメータ

```yaml
inventory:
  nodes:
    db-server:
      capabilities:
        - id: database.postgresql
          params:
            version: "15"
            port: 5432
            max_connections: 200
            shared_buffers: "256MB"
```

### クラスタのCapabilityでのパラメータ

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - id: cluster.kubernetes
          params:
            version: "1.28"
            network_plugin: calico
            pod_network_cidr: "10.244.0.0/16"
```

### クラスタのparamsフィールド

クラスタには、`capabilities` とは別に、トップレベルの `params` フィールドもあります：

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [cp]
          kube_node: [worker1, worker2]
          etcd: [cp]
```

このトップレベルの `params` は、クラスタ全体に対するパラメータを定義します。

### ジェネレーターのprovisionerでのパラメータ

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
        disk_size: "40GB"
```

### パラメータの確認方法

各レシピがどのようなパラメータを受け取るかは、レシピのドキュメントを参照してください。レシピは、必須パラメータとオプショナルパラメータ、デフォルト値を定義しています。

---

## provisioner（ジェネレーター）

ジェネレーターの `provisioner` も、レシピIDの指定方法と同じルールに従います：

### metadata.idによる指定

```yaml
infrastructure:
  generators:
    vms:
      provisioner: com.example.vagrant-provider
```

### エイリアス名 + attrsによる指定

```yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
```

この例では、`infra.vm` というエイリアス名のレシピで、`driver` 属性が `vagrant` のものを選択します。

---

## 実践例

### 例1: metadata.id + バージョン指定

```yaml
inventory:
  nodes:
    web-01:
      capabilities:
        - com.example.nginx-server@1.2.0

    db-01:
      capabilities:
        - com.example.postgresql@2.1.0
          params:
            version: "15"
            max_connections: 100
```

**用途**: 特定のレシピとバージョンを固定したい場合。

### 例2: エイリアス名のみ

```yaml
inventory:
  nodes:
    web-01:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter

    db-01:
      capabilities:
        - database.postgresql
```

**用途**: シンプルに機能を指定したい場合。

### 例3: エイリアス名 + attrs絞り込み

```yaml
inventory:
  nodes:
    nginx-server:
      capabilities:
        - web.server.engine=nginx

    apache-server:
      capabilities:
        - web.server.engine=apache

    postgres-db:
      capabilities:
        - database.type=relational.engine=postgresql

    mysql-db:
      capabilities:
        - database.type=relational.engine=mysql
```

**用途**: 同じ種類の機能で、異なる実装を選択したい場合。

### 例4: 詳細形 + パラメータ

```yaml
inventory:
  nodes:
    cache-01:
      capabilities:
        - id: cache.store.type=in-memory
          params:
            max_memory: "1GB"
            eviction_policy: "lru"

    db-01:
      capabilities:
        - id: database.postgresql
          params:
            version: "15"
            port: 5432
            max_connections: 200
```

**用途**: パラメータで動作をカスタマイズしたい場合。

### 例5: クラスタCapability

```yaml
inventory:
  nodes:
    cp: {}
    worker1: {}
    worker2: {}

  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - id: cluster.kubernetes
          params:
            version: "1.28"
            network_plugin: calico
      params:
        groups:
          kube_control_plane: [cp]
          kube_node: [worker1, worker2]
          etcd: [cp]
```

**用途**: 複数ノードにまたがる処理を定義する場合。

### 例6: ジェネレーターのprovisioner

```yaml
infrastructure:
  generators:
    # エイリアス + attrs
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64

    # 組み込みprovisioner
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
```

**用途**: インスタンスの生成方法を指定する場合。

---

## レシピの選択ガイド

### metadata.id vs エイリアス名

| 指定方法 | メリット | デメリット | 用途 |
|---------|---------|-----------|------|
| **metadata.id** | ・特定のレシピを確実に指定<br>・バージョン固定可能 | ・レシピの入れ替えが困難<br>・実装に依存 | ・本番環境<br>・バージョン管理が重要 |
| **エイリアス名** | ・実装の詳細を隠蔽<br>・レシピの入れ替えが容易 | ・複数該当する場合は絞り込みが必要 | ・開発環境<br>・柔軟性が重要 |

### パラメータの有無

| 記法 | 用途 |
|-----|------|
| **簡略形**（文字列） | デフォルト設定で十分な場合 |
| **詳細形**（オブジェクト） | パラメータでカスタマイズしたい場合 |

### attrsの活用

- 同じ種類の機能で、異なる実装を選択したい場合に便利
- 例: `web.server.engine=nginx` vs `web.server.engine=apache`

---

## まとめ

- **レシピID指定方法**: metadata.id または エイリアス名 + attrs
- **記法**: 簡略形（文字列）または 詳細形（オブジェクト + params）
- **用途**:
  - ノード: 個別の機能
  - クラスタ: 複数ノードにまたがる処理
  - ジェネレーター: インスタンスの生成方法

レシピの指定方法を理解することで、柔軟で保守性の高いStateファイルを作成できます。

---

## 次のステップ

- [ノードの詳細]({{< relref "nodes" >}}) - ノードのCapabilityの詳細
- [クラスタの詳細]({{< relref "clusters" >}}) - クラスタのCapabilityの詳細
- [ジェネレーターの詳細]({{< relref "generators" >}}) - provisionerの詳細
- [リファレンスレシピの利用]({{< relref "/recipes" >}}) - リファレンスレシピ一覧

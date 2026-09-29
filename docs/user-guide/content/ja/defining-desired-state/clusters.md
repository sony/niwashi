---
title: "クラスタの詳細"
weight: 7
---

# クラスタの詳細

このページでは、クラスタの全属性と使い方について詳しく説明します。

## クラスタとは

クラスタは、複数のノードで構成されるグループです。クラスタレベルの機能（Capability）を定義することで、複数ノードにまたがる処理（Kubernetesクラスタの構築など）を実現できます。

---

## 基本的なクラスタの定義

クラスタには、**最低1つ以上のノード**を指定する必要があります。

```yaml
inventory:
  nodes:
    node1: {}
    node2: {}
    node3: {}

  clusters:
    my-cluster:
      nodes:
        - node1
        - node2
        - node3
```

---

## クラスタの属性

クラスタには以下の属性を指定できます。

### nodes（必須）

クラスタに参加するノードを指定します。少なくとも1つのノードが必要です。**文字列リスト形式**と**オブジェクト形式**の2通りで書けます。

#### 文字列リスト形式

```yaml
inventory:
  clusters:
    web-cluster:
      nodes:
        - web-01
        - web-02
```

#### オブジェクト形式（ラベル付き）

ノード名をキー、`labels` でそのクラスタ内での役割などを定義します。`where` フィールドを使ったタスクの絞り込みに利用します。

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        cp:
          labels:
            role: master
        worker1:
          labels:
            role: worker
        worker2:
          labels:
            role: worker
```

`labels` が不要なノードは `{}` で省略できます。

```yaml
nodes:
  cp:
    labels:
      role: master
  worker1: {}   # labels なし
```

#### クラスタノードラベルとグローバルノードラベルの関係

ノードには `inventory.nodes` でグローバルなラベルを定義できます。クラスタノードのラベルはそのクラスタのレシピ実行時のみ有効で、グローバルラベルにマージされます（クラスタノードラベルが優先）。これらのラベルはレシピの `where` フィールドで `node.labels` としてアクセスできます。

```yaml
inventory:
  nodes:
    cp:
      labels:
        tier: control      # グローバルラベル

  clusters:
    k8s-cluster:
      nodes:
        cp:
          labels:
            role: master   # k8sクラスタ内でのラベル（tier: control とマージされる）

    db-cluster:
      nodes:
        cp:
          labels:
            role: replica  # dbクラスタ内でのラベル（k8sのラベルとは独立）
```

`k8s-cluster` のレシピ実行時、`cp` の `node.labels` は `{tier: control, role: master}` として評価されます。詳細は [タスクのフィルタリング条件 (where)]({{< relref "../defining-recipes/task-where" >}}) を参照してください。

#### 注意事項

- ノードが `inventory.nodes` に存在しない場合、エラーになります
- 空の配列・空のマップは指定できません（最低1つのノードが必要）
- 同じノードを複数回指定することはできません

---

### capabilities

クラスタレベルの機能（Capability）を指定します。これは、複数ノードにまたがる処理を定義する際に使用します。

#### 簡略形（文字列のみ）

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        - cp
        - worker1
        - worker2
      capabilities:
        - cluster.kubernetes
```

#### 詳細形（パラメータを指定）

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        - cp
        - worker1
        - worker2
      capabilities:
        - id: cluster.kubernetes
          params:
            version: "1.28"
            network_plugin: calico
```

#### クラスタCapabilityの用途

クラスタCapabilityは、以下のような複数ノードにまたがる処理に使用します：

- **Kubernetesクラスタの構築**: control-planeとworkerノードの設定
- **データベースレプリケーション**: primary/replicaの設定
- **ロードバランサー構成**: 複数のWebサーバーへの負荷分散
- **分散ストレージ**: 複数ノードでのストレージクラスタ

Capabilityの詳細は、[Capabilityとレシピの詳細]({{< relref "capabilities" >}}) を参照してください。

---

### params

クラスタ全体に対するパラメータを指定します。これは、クラスタのCapabilityで使用されるパラメータを定義する際に利用します。

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        - cp
        - worker1
        - worker2
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [cp]
          kube_node: [worker1, worker2]
          etcd: [cp]
```

#### capabilities の params との違い

- **capabilities の params**: 個々のCapabilityに渡すパラメータ
- **クラスタの params**: クラスタ全体に対するパラメータ

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]

      # Capabilityに渡すパラメータ
      capabilities:
        - id: cluster.kubernetes
          params:
            version: "1.28"
            network_plugin: calico

      # クラスタ全体のパラメータ
      params:
        groups:
          kube_control_plane: [cp]
          kube_node: [worker1, worker2]
```

どちらを使うかは、レシピの仕様によって異なります。レシピのドキュメントを参照してください。

---

### templates

テンプレートを継承してクラスタを定義できます。

```yaml
template:
  cluster:
    base-cluster:
      labels:
        managed_by: niwashi

inventory:
  clusters:
    prod-cluster:
      templates: [base-cluster]
      nodes:
        - node1
        - node2
```

#### 複数テンプレートの適用

```yaml
template:
  cluster:
    base:
      labels:
        managed_by: niwashi

    production:
      labels:
        environment: production

inventory:
  clusters:
    prod-cluster:
      templates: [base, production]
      nodes:
        - node1
        - node2
```

テンプレートの詳細は、[テンプレートの詳細]({{< relref "templates" >}}) を参照してください。

---

## クラスタ定義の実践例

### Kubernetesクラスタ

Ansibleアダプターを使う場合（文字列リスト形式＋`params.groups` でノード役割を指定）：

```yaml
inventory:
  nodes:
    control-plane: {}
    worker-01: {}
    worker-02: {}

  clusters:
    k8s-cluster:
      labels:
        environment: production
        platform: kubernetes
      nodes:
        - control-plane
        - worker-01
        - worker-02
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [control-plane]
          kube_node: [worker-01, worker-02]
          etcd: [control-plane]
```

`where` を使ったレシピで役割ごとに処理を分岐する場合（オブジェクト形式でノードラベルを定義）：

```yaml
inventory:
  nodes:
    cp: {}
    worker-01: {}
    worker-02: {}

  clusters:
    k8s-cluster:
      nodes:
        cp:
          labels:
            role: master
        worker-01:
          labels:
            role: worker
        worker-02:
          labels:
            role: worker
      capabilities:
        - my-org.my-k8s
```

### データベースレプリケーション

```yaml
inventory:
  nodes:
    db-primary: {}
    db-replica-01: {}
    db-replica-02: {}

  clusters:
    db-cluster:
      labels:
        environment: production
        database: postgresql
      nodes:
        - db-primary
        - db-replica-01
        - db-replica-02
      capabilities:
        - id: database.postgresql-replication
          params:
            replication_mode: streaming
            primary_node: db-primary
```

### Webサーバークラスタ

```yaml
inventory:
  nodes:
    web-01: {}
    web-02: {}
    web-03: {}
    lb: {}

  clusters:
    web-tier:
      labels:
        tier: frontend
      nodes:
        - web-01
        - web-02
        - web-03
      capabilities:
        - web.load-balancer
      params:
        backend_nodes:
          - web-01
          - web-02
          - web-03
        load_balancer_node: lb
```

### 複数Capabilityの例

```yaml
inventory:
  clusters:
    app-cluster:
      nodes:
        - app-01
        - app-02
      capabilities:
        - cluster.monitoring
        - cluster.logging
        - cluster.backup
      params:
        monitoring_endpoint: "http://prometheus:9090"
        logging_endpoint: "http://elasticsearch:9200"
```

---

## ノードとクラスタの関係

### 単独ノード vs クラスタ所属ノード

#### 単独ノード

ノードは、クラスタに所属させずに単独で使用できます：

```yaml
inventory:
  nodes:
    standalone-server:
      capabilities:
        - web.nginx
```

このノードは、どのクラスタにも所属していません。

#### クラスタ所属ノード

ノードをクラスタに所属させることで、複数ノードにまたがる処理を実現できます：

```yaml
inventory:
  nodes:
    node1:
      capabilities:
        - web.nginx  # ノード個別のCapability

    node2:
      capabilities:
        - web.nginx

  clusters:
    web-cluster:
      nodes:
        - node1
        - node2
      capabilities:
        - cluster.load-balancer  # クラスタレベルのCapability
```

この例では：
- 各ノードは個別に `web.nginx` を持つ
- クラスタレベルで `cluster.load-balancer` を設定し、複数ノードを束ねる

### 複数クラスタへの所属

1つのノードを複数のクラスタに所属させることも可能です：

```yaml
inventory:
  nodes:
    node1: {}
    node2: {}
    node3: {}

  clusters:
    cluster-a:
      nodes:
        - node1
        - node2
      capabilities:
        - feature-a

    cluster-b:
      nodes:
        - node2
        - node3
      capabilities:
        - feature-b
```

この例では、`node2` が `cluster-a` と `cluster-b` の両方に所属しています。

---

## よくあるパターン

### パターン1: Kubernetesクラスタ（Ansibleアダプター使用）

```yaml
inventory:
  nodes:
    cp: {}
    worker1: {}
    worker2: {}

  clusters:
    k8s:
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [cp]
          kube_node: [worker1, worker2]
          etcd: [cp]
```

### パターン1b: Kubernetesクラスタ（`where` によるノード役割分岐）

```yaml
inventory:
  nodes:
    cp: {}
    worker1: {}
    worker2: {}

  clusters:
    k8s:
      nodes:
        cp:
          labels:
            role: master
        worker1:
          labels:
            role: worker
        worker2:
          labels:
            role: worker
      capabilities:
        - my-org.my-k8s
```

### パターン2: データベースクラスタ

```yaml
inventory:
  nodes:
    db-primary:
      capabilities:
        - database.postgresql
    db-replica:
      capabilities:
        - database.postgresql

  clusters:
    db-cluster:
      nodes: [db-primary, db-replica]
      capabilities:
        - database.replication
      params:
        primary: db-primary
        replicas: [db-replica]
```

### パターン3: マイクロサービス構成

```yaml
inventory:
  nodes:
    api-01: {}
    api-02: {}
    frontend-01: {}
    frontend-02: {}

  clusters:
    api-tier:
      nodes: [api-01, api-02]
      capabilities:
        - cluster.service-mesh

    frontend-tier:
      nodes: [frontend-01, frontend-02]
      capabilities:
        - cluster.service-mesh
```

---

## トラブルシューティング

### ノードが見つからない

**問題**: クラスタの `nodes` で指定したノードが見つからない。

**原因**:
- ノード名の綴りが間違っている
- ノードが `inventory.nodes` で定義されていない

**解決策**:
1. `inventory.nodes` でノードを定義する
2. ノード名の綴りを確認する

```yaml
# Bad: node1 が定義されていない
inventory:
  clusters:
    my-cluster:
      nodes:
        - node1  # エラー

# Good: node1 を定義
inventory:
  nodes:
    node1: {}

  clusters:
    my-cluster:
      nodes:
        - node1
```

### クラスタCapabilityが適用されない

**問題**: クラスタにCapabilityを指定したが、期待通りに動作しない。

**原因**:
- レシピIDが間違っている
- パラメータが不足している
- クラスタレベルのCapabilityではなく、ノードレベルのCapabilityを指定している

**解決策**:
1. レシピIDの綴りを確認する
2. レシピのドキュメントで、クラスタレベルで使用できるか確認する
3. 必須パラメータが揃っているか確認する

---

## 次のステップ

- [ノードの詳細]({{< relref "nodes" >}}) - ノードの全属性とオプション
- [ジェネレーターの詳細]({{< relref "generators" >}}) - インスタンスの生成方法
- [Capabilityとレシピの詳細]({{< relref "capabilities" >}}) - レシピの指定方法
- [テンプレートの詳細]({{< relref "templates" >}}) - テンプレートの活用

---
title: "Inventory の基本"
weight: 2
---

# Inventory の基本

`inventory` セクションでは、システムの論理的な構成を定義します。「どのようなシステムを作りたいか」を記述する場所です。

## Inventory の構造

```yaml
inventory:
  nodes:
    # 個々の論理ノードの定義
  clusters:
    # クラスタの定義（オプション）
```

- **nodes**（必須）: 個々の論理ノードの定義
- **clusters**（オプション）: 複数のノードで構成されるクラスタの定義

---

## Nodes（ノード）

ノードは、システムを構成する個々の論理的な計算単位です。

### 基本的な定義

最もシンプルなノード定義：

```yaml
inventory:
  nodes:
    node1: {}
    node2: {}
    node3: {}
```

この例では、3つの空のノードを定義しています。識別子（`node1`, `node2`, `node3`）のみを指定し、属性は空です。

### ノードの識別子

各ノードには、一意な識別子（キー）を指定します。識別子は、Stateファイル内でノードを参照する際に使用されます。

```yaml
inventory:
  nodes:
    web-server: {}
    db-primary: {}
    cache-server: {}
```

識別子の命名規則については、[識別子の命名規則]({{< relref "naming-rules" >}}) を参照してください。

### ノードに指定できる属性

ノードには、以下の属性を指定できます：

#### labels

ノードにラベル（メタデータ）を付与します：

```yaml
inventory:
  nodes:
    web-server:
      labels:
        role: web
        environment: production
        region: us-east-1
```

#### capabilities

ノードが持つべき機能（Capability）を指定します：

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter
```

Capabilityは、レシピによって実現される機能の単位です。詳細は [Capabilityとレシピの詳細]({{< relref "../capabilities" >}}) を参照してください。

#### instanceSelector

ノードに割り当てるインスタンスを制御します：

```yaml
inventory:
  nodes:
    web-server:
      instanceSelector:
        generator: web-vms
```

詳細は [ノードの詳細]({{< relref "../nodes" >}}) を参照してください。

#### templates

テンプレートを継承します：

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
```

詳細は [テンプレートの詳細]({{< relref "../templates" >}}) を参照してください。

### ノード定義の例

```yaml
inventory:
  nodes:
    # シンプルなノード
    worker-01: {}

    # ラベル付きノード
    web-server:
      labels:
        role: web
        tier: frontend

    # Capability付きノード
    db-server:
      capabilities:
        - database.postgresql

    # 詳細な設定を持つノード
    app-server:
      labels:
        role: application
        tier: backend
      capabilities:
        - id: runtime.python
          params:
            version: "3.11"
      instanceSelector:
        generator: app-vms
```

---

## Clusters（クラスタ）

クラスタは、複数のノードで構成されるグループです。クラスタレベルの機能を定義することで、複数ノードにまたがる処理（Kubernetesクラスタの構築など）を実現できます。

### 基本的な定義

クラスタには、**最低1つ以上のノード**を指定する必要があります：

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

### クラスタの識別子

ノードと同様に、各クラスタには一意な識別子を指定します：

```yaml
inventory:
  clusters:
    web-cluster: {}
    db-cluster: {}
    k8s-cluster: {}
```

### クラスタに指定できる属性

クラスタには、以下の属性を指定できます：

#### nodes（必須）

クラスタに参加するノードのリストを配列で指定します：

```yaml
inventory:
  clusters:
    web-cluster:
      nodes:
        - web-01
        - web-02
        - web-03
```

ノード名は、`inventory.nodes` で定義された識別子と一致している必要があります。

#### capabilities

クラスタレベルの機能（Capability）を指定します：

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        - control-plane
        - worker-01
        - worker-02
      capabilities:
        - cluster.kubernetes
```

#### params

クラスタ全体に対するパラメータを指定します：

```yaml
inventory:
  clusters:
    k8s-cluster:
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

この例では、Kubernetesクラスタを構築する際に、各ノードの役割をグループとして定義しています。

#### templates

テンプレートを継承します：

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

### クラスタ定義の例

```yaml
inventory:
  nodes:
    control-plane: {}
    worker-01: {}
    worker-02: {}
    db-primary: {}
    db-replica: {}

  clusters:
    # Kubernetesクラスタ
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

    # データベースクラスタ
    db-cluster:
      labels:
        environment: production
        database: postgresql
      nodes:
        - db-primary
        - db-replica
      capabilities:
        - id: database.postgresql-replication
          params:
            replication_mode: streaming
```

---

## ノードとクラスタの関係

### 単独ノード

ノードは、クラスタに所属させずに単独で使用できます：

```yaml
inventory:
  nodes:
    standalone-server:
      capabilities:
        - web.nginx
```

### クラスタ所属ノード

ノードをクラスタに所属させることで、複数ノードにまたがる処理を実現できます：

```yaml
inventory:
  nodes:
    node1: {}
    node2: {}

  clusters:
    my-cluster:
      nodes:
        - node1
        - node2
      capabilities:
        - cluster-level-feature
```

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

    cluster-b:
      nodes:
        - node2
        - node3
```

この例では、`node2` が `cluster-a` と `cluster-b` の両方に所属しています。

---

## 実践例

### Webアプリケーション

```yaml
inventory:
  nodes:
    web-01:
      capabilities:
        - web.nginx
    web-02:
      capabilities:
        - web.nginx
    db-01:
      capabilities:
        - database.postgresql

  clusters:
    web-tier:
      nodes:
        - web-01
        - web-02
```

### Kubernetesクラスタ

```yaml
inventory:
  nodes:
    cp:
      instanceSelector:
        generator: vms
    worker1:
      instanceSelector:
        generator: vms
    worker2:
      instanceSelector:
        generator: vms

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

---

## 次のステップ

- [Infrastructure の基本]({{< relref "infrastructure" >}}) - ノードに割り当てるインスタンスの生成方法
- [識別子の命名規則]({{< relref "naming-rules" >}}) - ノードとクラスタの命名ルール
- [ノードの詳細]({{< relref "../nodes" >}}) - ノードの全属性とオプション
- [クラスタの詳細]({{< relref "../clusters" >}}) - クラスタの全属性とオプション

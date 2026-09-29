---
title: "Stateマージの詳細"
weight: 2
---

# Stateマージの詳細

このページでは、複数のStateファイルをマージする際の詳細な動作について説明します。

## マージの基本

`nwsctl plan` コマンドで複数の `-t` オプションを指定すると、Stateファイルは指定された順にマージされます。

```bash
nwsctl plan -t file1.yaml -t file2.yaml -t file3.yaml
```

マージの順序:
1. `file1.yaml` を読み込む
2. `file2.yaml` を読み込み、`file1.yaml` とマージする
3. `file3.yaml` を読み込み、マージ結果とさらにマージする

**重要**: 後に指定したファイルが前のファイルの設定を上書きします。

---

## トップレベル要素のマージ

### version

`version` フィールドは、すべてのファイルで一致している必要があります。異なるバージョンが指定されている場合、エラーになります。

```yaml
# file1.yaml
version: nws.state/v1

# file2.yaml
version: nws.state/v1  # OK: 同じバージョン
```

### metadata

`metadata` は、後のファイルが前のファイルを上書きします。

```yaml
# file1.yaml
metadata:
  project: my-project

# file2.yaml
metadata:
  project: my-project-prod  # file1の設定を上書き

# マージ結果
metadata:
  project: my-project-prod
```

### inventory

`inventory` セクションは、ノードやクラスタごとにマージされます。

### infrastructure

`infrastructure` セクションは、ジェネレーターごとにマージされます。

### template

`template` セクションは、テンプレートごとにマージされます。

---

## inventory のマージ

### nodes

ノードは、識別子（キー）ごとにマージされます。同じ識別子のノードがある場合、属性がマージされます。

#### 例1: ノードの追加

```yaml
# file1.yaml
inventory:
  nodes:
    node1: {}

# file2.yaml
inventory:
  nodes:
    node2: {}

# マージ結果
inventory:
  nodes:
    node1: {}
    node2: {}
```

#### 例2: ノードの属性追加

```yaml
# file1.yaml
inventory:
  nodes:
    web-server: {}

# file2.yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx

# マージ結果
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
```

#### 例3: labels のマージ

labels は、キーごとにマージされます。同じキーがある場合、後のファイルが前のファイルを上書きします。

```yaml
# file1.yaml
inventory:
  nodes:
    web-server:
      labels:
        env: dev
        region: us-east

# file2.yaml
inventory:
  nodes:
    web-server:
      labels:
        env: prod        # 上書き
        tier: frontend   # 追加

# マージ結果
inventory:
  nodes:
    web-server:
      labels:
        env: prod        # file2で上書きされた
        region: us-east  # file1から
        tier: frontend   # file2で追加
```

#### 例4: capabilities のマージ

capabilities は、配列として結合されます。

```yaml
# file1.yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx

# file2.yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - monitoring.prometheus-exporter

# マージ結果
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter
```

#### 例5: instanceSelector の上書き

`instanceSelector` は、後のファイルが前のファイルを完全に上書きします。

```yaml
# file1.yaml
inventory:
  nodes:
    web-server:
      instanceSelector:
        generator: gen1

# file2.yaml
inventory:
  nodes:
    web-server:
      instanceSelector:
        generator: gen2  # 完全に上書き

# マージ結果
inventory:
  nodes:
    web-server:
      instanceSelector:
        generator: gen2
```

### clusters

クラスタも、ノードと同様に識別子ごとにマージされます。

#### 例1: クラスタの追加

```yaml
# file1.yaml
inventory:
  clusters:
    cluster1:
      nodes: [node1, node2]

# file2.yaml
inventory:
  clusters:
    cluster2:
      nodes: [node3, node4]

# マージ結果
inventory:
  clusters:
    cluster1:
      nodes: [node1, node2]
    cluster2:
      nodes: [node3, node4]
```

#### 例2: nodes の上書き

クラスタの `nodes` は、後のファイルが前のファイルを完全に上書きします。

```yaml
# file1.yaml
inventory:
  clusters:
    my-cluster:
      nodes: [node1, node2]

# file2.yaml
inventory:
  clusters:
    my-cluster:
      nodes: [node1, node2, node3]  # 完全に上書き

# マージ結果
inventory:
  clusters:
    my-cluster:
      nodes: [node1, node2, node3]
```

#### 例3: capabilities の結合

```yaml
# file1.yaml
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes

# file2.yaml
inventory:
  clusters:
    k8s-cluster:
      capabilities:
        - monitoring.cluster-monitoring

# マージ結果
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes
        - monitoring.cluster-monitoring
```

---

## infrastructure のマージ

### generators

ジェネレーターは、識別子ごとにマージされます。

#### 例1: ジェネレーターの追加

```yaml
# file1.yaml
infrastructure:
  generators:
    gen1:
      provisioner: external-instance

# file2.yaml
infrastructure:
  generators:
    gen2:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3

# マージ結果
infrastructure:
  generators:
    gen1:
      provisioner: external-instance
    gen2:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
```

#### 例2: params のマージ

`params` は、キーごとにマージされます。

```yaml
# file1.yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        box: ubuntu/jammy64
        cpus: 2

# file2.yaml
infrastructure:
  generators:
    vms:
      params:
        count: 3
        memory: 2048

# マージ結果
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        box: ubuntu/jammy64
        cpus: 2
        count: 3
        memory: 2048
```

#### 例3: params の上書き

同じキーがある場合、後のファイルが前のファイルを上書きします。

```yaml
# file1.yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        memory: 1024

# file2.yaml
infrastructure:
  generators:
    vms:
      params:
        memory: 4096  # 上書き

# マージ結果
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        memory: 4096  # file2で上書きされた
```

---

## template のマージ

テンプレートは、識別子ごとにマージされます。

### 例1: テンプレートの追加

```yaml
# file1.yaml
template:
  node:
    base:
      labels:
        managed_by: niwashi

# file2.yaml
template:
  node:
    production:
      labels:
        environment: production

# マージ結果
template:
  node:
    base:
      labels:
        managed_by: niwashi
    production:
      labels:
        environment: production
```

### 例2: テンプレートの上書き

```yaml
# file1.yaml
template:
  node:
    base:
      labels:
        env: dev

# file2.yaml
template:
  node:
    base:
      labels:
        env: prod  # 上書き

# マージ結果
template:
  node:
    base:
      labels:
        env: prod
```

---

## マージ動作のまとめ

| 要素 | マージ動作 |
|------|----------|
| **トップレベル** | |
| version | 一致している必要がある |
| metadata | 後のファイルが上書き |
| **inventory.nodes** | |
| ノード自体 | 識別子ごとにマージ |
| labels | キーごとにマージ（同じキーは上書き） |
| capabilities | 配列として結合 |
| instanceSelector | 後のファイルが完全に上書き |
| templates | 後のファイルが完全に上書き |
| **inventory.clusters** | |
| クラスタ自体 | 識別子ごとにマージ |
| nodes | 後のファイルが完全に上書き |
| labels | キーごとにマージ（同じキーは上書き） |
| capabilities | 配列として結合 |
| params | キーごとにマージ（同じキーは上書き） |
| templates | 後のファイルが完全に上書き |
| **infrastructure.generators** | |
| ジェネレーター自体 | 識別子ごとにマージ |
| provisioner | 後のファイルが上書き |
| params | キーごとにマージ（同じキーは上書き） |
| templates | 後のファイルが完全に上書き |
| **template** | |
| テンプレート自体 | 識別子ごとにマージ |
| 属性 | ノード/クラスタ/インスタンスと同じルール |

---

## 実践例

### 例1: 論理構成とインフラの分離

```yaml
# app.yaml（論理構成）
version: nws.state/v1

inventory:
  nodes:
    web: {}
    db: {}

# infra-dev.yaml（インフラ）
version: nws.state/v1

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
```

マージコマンド:
```bash
nwsctl plan -t app.yaml -t infra-dev.yaml
```

マージ結果:
```yaml
version: nws.state/v1

inventory:
  nodes:
    web: {}
    db: {}

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
```

### 例2: 段階的な定義

```yaml
# 01-base.yaml（基本構造）
version: nws.state/v1

inventory:
  nodes:
    web: {}
    db: {}

# 02-capabilities.yaml（機能定義）
version: nws.state/v1

inventory:
  nodes:
    web:
      capabilities:
        - web.nginx
    db:
      capabilities:
        - database.postgresql

# 03-infra.yaml（インフラ）
version: nws.state/v1

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
```

マージコマンド:
```bash
nwsctl plan -t 01-base.yaml -t 02-capabilities.yaml -t 03-infra.yaml
```

マージ結果:
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

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
```

---

## 注意点

### 1. マージの順序

マージの順序は非常に重要です。後のファイルが前のファイルを上書きします。

```bash
# 正しい順序
nwsctl plan -t base.yaml -t override.yaml

# 間違った順序（overrideの設定がbaseに上書きされる）
nwsctl plan -t override.yaml -t base.yaml
```

### 2. 配列の上書き vs 結合

- **capabilities**: 結合される
- **nodes（クラスタの）**: 上書きされる
- **templates**: 上書きされる

この違いを理解しておくことが重要です。

### 3. 意図しない上書き

同じ識別子を使うと、意図せず設定が上書きされる可能性があります。

```yaml
# file1.yaml
inventory:
  nodes:
    web-server:
      labels:
        env: dev

# file2.yaml
inventory:
  nodes:
    web-server:  # 同じ識別子
      labels:
        region: us-east  # マージされる
```

識別子が一致していることを確認しましょう。

---

## 次のステップ

- [複数環境の管理]({{< relref "multi-environment" >}}) - マージを活用した環境管理の実践

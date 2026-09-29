---
title: "状態定義の概要"
weight: 1
---

# 状態定義の概要

## Stateとは

Niwashiでは、インフラストラクチャとアプリケーション環境の**あるべき姿**を「State（状態）」として定義します。Stateには以下の2種類があります：

- **目標の状態（Target State / Desired State）**: ユーザーが定義する、システムが最終的に到達すべき状態
- **現在の状態（Current State）**: システムが実際に保持している現時点での状態

Niwashiは、この2つの状態の差分を計算し、目標の状態を実現するための計画（Plan）を生成します。

### YAMLフォーマットでの記述

目標の状態は、YAMLフォーマットのファイルで記述します。YAMLは人間が読み書きしやすい形式であり、バージョン管理システム（Git等）で管理することで、インフラストラクチャの変更履歴を追跡できます。

### スキーマバージョン

現在のStateフォーマットのバージョンは **`nws.state/v1`** です。すべてのStateファイルには、先頭に以下のように `version` フィールドを記述する必要があります：

```yaml
version: nws.state/v1
```

---

## 最小限の例

まず、最もシンプルなStateファイルを見てみましょう：

```yaml
version: nws.state/v1

inventory:
  nodes:
    node1: {}
    node2: {}
    node3: {}

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
```

この例では：
- 3つの論理ノード（`node1`, `node2`, `node3`）を定義
- Vagrantを使って3台の仮想マシンを生成するジェネレーター（`vms`）を定義

Niwashiはこの定義から、ノードに仮想マシンを自動的に割り当てます。

---

## State の基本構造

Stateファイルは、以下のトップレベル要素で構成されます：

```yaml
version: nws.state/v1

inventory:
  nodes:    # 論理ノードの定義
  clusters: # クラスタの定義

infrastructure:
  generators: # インスタンス生成器の定義

template:
  node:     # ノードのテンプレート
  cluster:  # クラスタのテンプレート
  instance: # インスタンスのテンプレート
```

### version（必須）

スキーマのバージョンを指定します。現在は `nws.state/v1` を使用します。

### inventory（必須）

システムの**論理構成**を定義するセクションです：

- **nodes**（必須）: 個々の論理ノードの定義
- **clusters**: 複数のノードで構成されるクラスタの定義

#### Nodes（ノード）

ノードは、システムを構成する個々の論理的な計算単位です。

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
    db-server:
      capabilities:
        - database.postgresql
```

各ノードには、そのノードが持つべき機能（Capability）を指定できます。

詳細は [ノードの詳細]({{< relref "nodes" >}}) を参照してください。

#### Clusters（クラスタ）

クラスタは、複数のノードで構成されるグループです。クラスタレベルの機能を定義することで、複数ノードにまたがる処理を実現できます。

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        - control-plane
        - worker1
        - worker2
      capabilities:
        - cluster.kubernetes
```

詳細は [クラスタの詳細]({{< relref "clusters" >}}) を参照してください。

### infrastructure

物理的/仮想的なインフラストラクチャの定義を行うセクションです。

- **generators**: インスタンス（仮想マシンなど）を生成するための生成器の定義

```yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

ジェネレーターは、`provisioner` で指定されたレシピに基づいてインスタンスを生成します。

詳細は [ジェネレーターの詳細]({{< relref "generators" >}}) を参照してください。

### template

ノード、クラスタ、インスタンスの共通設定をテンプレートとして定義できます。

```yaml
template:
  node:
    base-node:
      labels:
        environment: production
        managed_by: niwashi

inventory:
  nodes:
    web-01:
      templates: [base-node]
      capabilities:
        - web.nginx
```

テンプレートを使用することで、同じ設定を持つ複数の要素を効率的に定義できます。

詳細は [テンプレートの詳細]({{< relref "templates" >}}) を参照してください。

---

## 論理構成と物理構成の分離

Niwashiの重要な特徴は、**論理構成**（どんなシステムを作りたいか）と**物理構成**（どこにデプロイするか）を分離できることです。

### 例: 環境別の設定

#### base.yaml（論理構成 - 共通）

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

#### dev.yaml（物理構成 - 開発環境）

```yaml
version: nws.state/v1

infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
        box: ubuntu/jammy64
        cpus: 1
        memory: 1024
```

#### prod.yaml（物理構成 - 本番環境）

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

#### 環境の切り替え

```bash
# 開発環境
nwsctl plan -t base.yaml -t dev.yaml

# 本番環境
nwsctl plan -t base.yaml -t prod.yaml
```

複数のStateファイルは、指定された順にマージされます。これにより、環境ごとの違いを別ファイルで管理できます。

---

## Capabilities とレシピ

Niwashiでは、ノードやクラスタが持つべき機能を「Capability」として指定します。Capabilityは、**レシピ**によって実現されます。

### 基本的な指定方法

#### 簡略形（文字列のみ）

```yaml
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter
```

#### 詳細形（パラメータ付き）

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
```

### レシピの指定方法

レシピは以下の方法で指定できます：

1. **エイリアス名**: `web.nginx`
2. **エイリアス + 属性**: `database.type=relational.engine=postgresql`
3. **レシピID**: `com.example.nginx-server`
4. **レシピID + バージョン**: `com.example.nginx-server@1.2.0`

詳細は [Capabilityとレシピの詳細]({{< relref "capabilities" >}}) を参照してください。

---

## コマンドでの利用

定義したStateファイルは、`nwsctl` コマンドで利用します。

### 計画の作成

```bash
nwsctl plan -t target.yaml
```

このコマンドは、目標の状態を実現するための計画を生成します。

複数ファイルを指定してマージすることもできます：

```bash
nwsctl plan -t base.yaml -t environment.yaml -t app.yaml
```

### 状態のエクスポート

```bash
# 標準出力に出力
nwsctl export

# ファイルに出力
nwsctl export -o current-state.yaml
```

現在の状態をYAML形式で出力します。

---

## 次のステップ

### 基本を学ぶ

状態定義の基本概念を順を追って学びたい場合は、以下のページを参照してください：

- [State の基本構造]({{< relref "basic-concepts/state-structure" >}}) - トップレベル要素の詳細
- [Inventory の基本]({{< relref "basic-concepts/inventory" >}}) - ノードとクラスタの基本
- [Infrastructure の基本]({{< relref "basic-concepts/infrastructure" >}}) - ジェネレーターの基本
- [識別子の命名規則]({{< relref "basic-concepts/naming-rules" >}}) - 命名ルール

### 詳細を学ぶ

各要素の完全なリファレンスは以下のページを参照してください：

- [ノードの詳細]({{< relref "nodes" >}}) - ノードの全属性と使い方
- [クラスタの詳細]({{< relref "clusters" >}}) - クラスタの全属性と使い方
- [ジェネレーターの詳細]({{< relref "generators" >}}) - ジェネレーターの全属性と使い方
- [Capabilityとレシピの詳細]({{< relref "capabilities" >}}) - レシピ指定の全方法
- [テンプレートの詳細]({{< relref "templates" >}}) - テンプレート機能の全て

### 高度な使い方

実践的なパターンを学びたい場合は、以下のページを参照してください：

- [複数環境の管理]({{< relref "advanced/multi-environment" >}}) - dev/prod環境の切り替え
- [Stateマージの詳細]({{< relref "advanced/state-merging" >}}) - マージの仕様

### 実践例

動作する完全な例を見たい場合は、以下のページを参照してください：

- [Kubernetesクラスタの構築]({{< relref "../examples/kubernetes-cluster" >}})
- Webアプリケーションの構成

---

以上が、Niwashiの状態定義の概要です。まずは最小限の例から始めて、徐々に機能を追加していくことをお勧めします。

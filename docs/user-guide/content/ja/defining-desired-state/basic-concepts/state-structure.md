---
title: "State の基本構造"
weight: 1
---

# State の基本構造

このページでは、Stateファイルのトップレベル構造と各要素の役割について説明します。

## 全体構造

Stateファイルは、以下のトップレベル要素で構成されます：

```yaml
version: nws.state/v1

metadata:
  project: my-project

inventory:
  nodes:
    # 論理ノードの定義
  clusters:
    # クラスタの定義

infrastructure:
  generators:
    # インスタンス生成器の定義

template:
  node:
    # ノードのテンプレート
  cluster:
    # クラスタのテンプレート
  instance:
    # インスタンスのテンプレート
```

---

## トップレベル要素

### version（必須）

スキーマのバージョンを指定します。現在は `nws.state/v1` を使用します。

```yaml
version: nws.state/v1
```

このバージョン識別子により、Niwashiは適切なスキーマでファイルを解釈します。将来的にフォーマットが拡張された場合でも、バージョンを明示することで後方互換性が保たれます。

**必須**: すべてのStateファイルに必要です。

---

### metadata（オプション）

Stateファイル全体に関するメタデータを定義します。

```yaml
metadata:
  project: my-k8s-cluster
```

#### 使用可能なフィールド

- **project**: プロジェクト名（任意の文字列）

metadataセクションは省略可能です。

---

### inventory（必須）

システムの**論理構成**を定義するセクションです。「どのようなシステムを作りたいか」を記述します。

```yaml
inventory:
  nodes:
    web-server: {}
    db-server: {}
  clusters:
    my-cluster:
      nodes:
        - web-server
        - db-server
```

#### 含まれる要素

- **nodes**（必須）: 個々の論理ノードの定義
- **clusters**（オプション）: 複数のノードで構成されるクラスタの定義

**必須**: `inventory` セクションと、その中の `nodes` は必須です。

詳細は [Inventory の基本]({{< relref "inventory" >}}) を参照してください。

---

### infrastructure（必須）

物理的/仮想的なインフラストラクチャの定義を行うセクションです。「どこにデプロイするか」を記述します。

```yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
```

#### 含まれる要素

- **generators**: インスタンス（仮想マシン、コンテナ、既存サーバーなど）を生成・管理するための生成器の定義

**必須**: ノードを実際のマシンにデプロイするためにはインスタンスが必要です。`infrastructure` セクションでジェネレーターを定義してください。

詳細は [Infrastructure の基本]({{< relref "infrastructure" >}}) を参照してください。

---

### template（オプション）

ノード、クラスタ、インスタンスの共通設定をテンプレートとして定義できます。

```yaml
template:
  node:
    base-node:
      labels:
        environment: production
        managed_by: niwashi
  cluster:
    base-cluster:
      labels:
        managed_by: niwashi
  instance:
    standard-vm:
      provisioner: infra.vm.driver=vagrant
      params:
        cpus: 2
        memory: 2048

inventory:
  nodes:
    web-01:
      templates: [base-node]
      capabilities:
        - web.nginx
```

テンプレートを使用することで、同じ設定を持つ複数の要素を効率的に定義できます。設定の重複を避け、メンテナンス性が向上します。

詳細は [テンプレートの詳細]({{< relref "../templates" >}}) を参照してください。

---

## 論理構成と物理構成

Niwashiの重要な特徴は、**論理構成**と**物理構成**を分離できることです。

### 論理構成（inventory）

「何を作りたいか」を定義します：

- どんな役割のノードが必要か
- ノードにはどんな機能が必要か
- ノードをどうグループ化するか

```yaml
inventory:
  nodes:
    web:
      capabilities:
        - web.nginx
    db:
      capabilities:
        - database.postgresql
```

### 物理構成（infrastructure）

「どこに作るか」を定義します：

- ローカルの仮想マシンか、クラウドか、既存サーバーか
- 何台のインスタンスが必要か
- リソース（CPU、メモリ）はどれくらいか

```yaml
infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
        box: ubuntu/jammy64
```

### 分離のメリット

この分離により、同じ論理構成を異なる環境にデプロイできます：

```bash
# 開発環境: ローカルの仮想マシン
nwsctl plan -t app.yaml -t infra-dev.yaml

# 本番環境: 既存のクラウドサーバー
nwsctl plan -t app.yaml -t infra-prod.yaml
```

`app.yaml` には論理構成を記述し、`infra-dev.yaml` と `infra-prod.yaml` には環境ごとの物理構成を記述します。

---

## 最小限の構成

最小限必要な要素は以下の通りです：

```yaml
version: nws.state/v1

inventory:
  nodes:
    node1: {}
```

この例では：
- バージョン指定（必須）
- 1つの論理ノード定義（inventory.nodes は必須）

ただし、実際には `infrastructure.generators` を定義してインスタンスを用意する必要があります。

---

## 完全な例

すべての要素を含む完全な例：

```yaml
version: nws.state/v1

metadata:
  project: my-web-app

template:
  node:
    base:
      labels:
        environment: production

inventory:
  nodes:
    web:
      templates: [base]
      capabilities:
        - web.nginx
    db:
      templates: [base]
      capabilities:
        - database.postgresql

  clusters:
    app-cluster:
      nodes:
        - web
        - db

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 2
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

---

## 次のステップ

各要素の詳細について学びましょう：

- [Inventory の基本]({{< relref "inventory" >}}) - ノードとクラスタの基本
- [Infrastructure の基本]({{< relref "infrastructure" >}}) - ジェネレーターの基本
- [識別子の命名規則]({{< relref "naming-rules" >}}) - 命名ルールとベストプラクティス

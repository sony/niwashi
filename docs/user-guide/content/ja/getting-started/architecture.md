---
title: "アーキテクチャ"
weight: 3
---

# アーキテクチャ

Niwashiの内部構造と動作原理について説明します。

---

## 概要

Niwashiは、**現在の状態**と**目標の状態**の差分を計算し、その差分を埋めるための実行計画を自動生成します。実行計画は、タスクの依存関係を考慮したDAG（有向非巡回グラフ）として構築されます。

---

## 基本的な動作原理

### 1. 状態の管理

Niwashiは2つの状態を管理します：

- **現在の状態（Current State）**: Niwashiが管理している現在のインフラとノードの状態
- **目標の状態（Desired State）**: ユーザーがYAMLファイルで定義した目標とする状態

### 2. 差分の計算

Niwashiは、現在の状態と目標の状態を比較し、以下の要素について差分を計算します：

- **ノードのCapability**: 各ノードに必要なCapabilityの追加・削除と、既存Capabilityの更新（`params` またはレシピのバージョンが変わったもの）
- **クラスタのCapability**: 各クラスタに必要なCapabilityの追加・削除と、既存Capabilityの更新（`params` またはレシピのバージョンが変わったもの）
- **Generator**: プロビジョニング・削除が必要なGenerator

### 3. 実行計画の生成

計算された差分とレシピを元に、実行計画を生成します。実行計画は、タスクの依存関係を考慮したDAG（有向非巡回グラフ）として構築されます。

---

## 差分計算の詳細

### ノードのCapability

各ノードについて、目標状態で定義されたCapabilityと現在の状態を比較します：

```yaml
# 目標状態
inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter

# 現在の状態（例: web.nginxのみ存在）
# → monitoring.prometheus-exporter を追加する必要がある
```

**差分**:
- 追加が必要なCapability: `monitoring.prometheus-exporter`
- 削除が必要なCapability: なし

両方の状態に存在するCapabilityについては、`params` と解決されたレシピのバージョンも比較します：

```yaml
# 現在の状態: web.nginx が params {port: 80} で適用済み
# 目標状態: web.nginx の params は {port: 8080}
# → web.nginx を更新する必要がある（operation: update のタスクが実行される）
```

詳細は [更新処理（operation: update）]({{< relref "defining-recipes/defining-tasks#更新処理operation-update" >}}) を参照してください。

### クラスタのCapability

各クラスタについて、目標状態で定義されたCapabilityと現在の状態を比較します：

```yaml
# 目標状態
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes

# 現在の状態（例: クラスタが存在しない）
# → cluster.kubernetes を追加する必要がある
```

**差分**:
- 追加が必要なCapability: `cluster.kubernetes`
- ノードの変更: なし（既にcp, worker1, worker2が存在）

### Generator（インスタンス生成）

Generatorは名前で比較します。目標状態にだけ存在するGeneratorはプロビジョニングされ、そのprovisionerがインスタンスを生成します：

```yaml
# 目標状態
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3

# 現在の状態（例: Generator vms が存在しない）
# → vms をプロビジョニングする必要がある（provisionerが3個のインスタンスを生成する）
```

**差分**:
- プロビジョニングが必要なGenerator: `vms`

既存のGeneratorの `params`（例: `count`）の変更は検出されません。

---

## DAG（有向非巡環グラフ）の構築

### DAGとは

DAG（Directed Acyclic Graph）は、タスク間の依存関係を表すグラフ構造です。Niwashiは、差分とレシピを元にDAGを構築します。

### タスクの依存関係

タスクの実行順序は、以下の要素から決定されます：

#### 1. レシピの依存関係（spec.requires）

各レシピは、`spec.requires` で他のレシピへの依存を宣言します。例えば、Kubernetesレシピは Ansible レシピに依存します：

```yaml
spec:
  requires:
    - host.tool.ansible
    - host.tool.git
```

この場合、Ansibleレシピが先に実行される必要があります。

#### 2. タスクの依存関係（dependsOn）

各レシピ内のタスクは、`dependsOn` で他のタスクへの依存を宣言します：

```yaml
tasks:
  - name: gen-inventory
    dependsOn: [ensure-dirs]
  - name: run-kubespray
    dependsOn: [gen-inventory]
```

この場合、`ensure-dirs` → `gen-inventory` → `run-kubespray` の順で実行されます。

#### 3. フェーズ（Phase）

レシピは、実行されるフェーズが決まっています：

1. **host**: ホスト単位の設定（ツールのインストールなど）
2. **infra**: インフラのプロビジョニング（Generatorの実行）
3. **node**: ノード単位のCapability適用
4. **cluster**: クラスタ単位のCapability適用

各フェーズは、前フェーズが完全に終了したことを確認してから実行されます。

なお、clusterフェーズの実行対象は基本的にクラスタ全体（1つ）ですが、タスクに `where` 条件が指定されている場合や、アダプターが `executionUnit: node` を宣言している場合は、タスクがクラスタ内のノード単位に分解されて実行されます。分解の条件の詳細は [Clusterのcapabilityを定義する]({{< relref "/defining-recipes/cluster-capability" >}}) を参照してください。

これらの依存関係を元に、DAGが構築されます。

### DAGの例

```
[フェーズ: host]
  ↓
[ホスト設定: ツールのインストール]
  ↓
[フェーズ: infra]
  ↓
[インスタンス生成: vms]
  ↓
[フェーズ: node]
  ↓
[ノード: web-server - web.nginx]
  ↓
[ノード: web-server - monitoring.prometheus-exporter]
  ↓
[フェーズ: cluster]
  ↓
[クラスタ: k8s-cluster - cluster.kubernetes]
```

---

## 実行計画の実行

### 計画ファイル

生成されたDAGは、JSON形式の計画ファイル（`plan.json`）として保存されます。

### 実行

`nwsctl apply` コマンドは、計画ファイルを読み込み、DAGに従ってタスクを順次実行します：

1. **依存関係の解決**: 各タスクの依存タスクが完了しているか確認
2. **タスクの実行**: レシピを呼び出してタスクを実行
3. **状態の更新**: タスクが完了したら、現在の状態を更新
4. **次のタスクへ**: 依存関係に従って次のタスクを実行

### エラーハンドリング

タスクの実行中にエラーが発生した場合、そのタスクに依存する後続のタスクは実行されません。

---

## アーキテクチャの利点

### 宣言的な定義

ユーザーは「何を」実現したいかだけを定義し、「どうやって」実現するかはNiwashiとレシピが決定します。

### 冪等性

Niwashiは、差分がない場合は何も実行しないため、理想的には冪等性が保たれます。ただし、実際の冪等性はレシピの実装に依存します。

これは特にタスクが失敗した場合に重要です。失敗したconstructはStateに記録されないため、次回の実行では途中まで完了した状態から再度実行されます（[タスクが失敗した場合]({{< relref "defining-recipes/defining-tasks#タスクが失敗した場合" >}}) 参照）。

Niwashiが提供しているリファレンスレシピ（Vagrant、Ansible、Kubernetesなど）は、冪等であることが確認されています。独自のレシピを作成する場合は、冪等性を考慮した実装が推奨されます。

### 柔軟性

レシピシステムにより、新しいプロビジョナーや設定管理ツールを簡単に追加できます。

---

## 次のステップ

- [状態を定義する]({{< relref "defining-desired-state" >}}) - Stateファイルの書き方
- [レシピを定義する]({{< relref "defining-recipes" >}}) - 独自のレシピを作成する方法

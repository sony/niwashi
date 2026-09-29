---
title: "Niwashi概要"
weight: 1
---

# Niwashi概要

## Niwashiとは

Niwashiは、インフラストラクチャのプロビジョニングと設定管理を統合的に行うツールです。
目標とする状態を宣言的に定義し、現在の状態から目標の状態への変更を計画・実行します。

---

## 使い方

Niwashiは、コマンドラインツール **`nwsctl`** を通して操作します。

基本的なワークフロー：
1. **目標の状態を定義**: YAMLファイルで記述
2. **計画を作成**: `nwsctl plan -t state.yaml`
3. **計画を実行**: `nwsctl apply`

詳細は [ワークフロー]({{< relref "workflow" >}}) を参照してください。

---

## Niwashiの特徴

### レシピシステム

Niwashiの最大の特徴は、柔軟な**レシピシステム**です。

レシピは、プロビジョニングや設定管理の手順をパッケージ化したモジュールです。宣言的に「何をするか」を定義し、実行をレシピに委ねることで、入力パラメータと実行結果を標準化されたインターフェースで扱えます。

- 同じレシピを複数の環境・プロジェクトで再利用
- レシピの組み合わせで複雑な構成を実現
- 実装の詳細を隠蔽し、シンプルなインターフェースで利用

Niwashiプロジェクトは、ポピュラーなOSSツール向けのリファレンスレシピを提供しています：

| レシピ | Capability ID | 説明 |
|--------|--------------|------|
| **Vagrant** | `infra.vm.driver=vagrant` | 仮想マシンの自動生成。ローカル開発環境の構築に最適 |
| **Ansible** | `adapter.tool.ansible` | Ansible実行のためのアダプター。レシピ間の連携を実現 |
| **Kubernetes** | `cluster.kubernetes` | Kubernetesクラスタの構築と管理 |

詳細は [リファレンスレシピ]({{< relref "recipes" >}}) を参照してください。

### 統合されたワークフロー

従来は別々のツールで行っていた以下の作業を、一つのワークフローで実行できます：

1. インフラのプロビジョニング（例: Vagrantレシピ）
2. 構成管理ツールの実行（例: Ansibleレシピ）
3. クラスタの構築（例: Kubernetesレシピ）

### 宣言的な状態定義

Niwashiでは、目標とする状態をYAMLファイル（State）で宣言的に記述します。Stateは**論理構成（Inventory）**と**物理インフラ（Infrastructure）**を分離して定義する構造になっており、同じ論理構成を異なるインフラで実行できます。

- テンプレートで共通設定を再利用
- Stateファイルのマージ機能により、開発・ステージング・本番環境の差分を管理

---

## ユースケース例

### 1. Kubernetesクラスタの構築

リファレンスレシピ（Vagrant、Kubernetes）を使った例：

```yaml
# 論理構成を定義
inventory:
  clusters:
    k8s-cluster:
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes  # リファレンスレシピ

# インフラを定義
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant  # リファレンスレシピ
      params:
        count: 3
```

→ Vagrantレシピで3台のVMを作成し、Kubernetesレシピでクラスタを自動構築

**レシピの連携**:
このKubernetesレシピは、内部的にAnsibleアダプター（`adapter.tool.ansible`）を呼び出してKubespray playbookを実行しています。これにより、レシピ間の連携が実現されています。

### 2. 開発環境と本番環境の切り替え

論理構成は共通、インフラだけ切り替える例：

```bash
# 開発環境（Vagrantレシピ）
nwsctl plan -t infra-dev.yaml -t app.yaml

# 本番環境（既存サーバー）
nwsctl plan -t infra-prod.yaml -t app.yaml
```

→ インフラ定義を先に読み込み、論理構成（app.yaml）でマージ

---

## 次のステップ

- [ワークフロー]({{< relref "workflow" >}}) - Niwashiの基本的な作業の流れ
- [アーキテクチャ]({{< relref "architecture" >}}) - Niwashiの内部構造
- [リファレンスレシピ]({{< relref "recipes" >}}) - 利用可能なリファレンスレシピの一覧
- [レシピを定義する]({{< relref "defining-recipes" >}}) - 独自のレシピを作成する方法

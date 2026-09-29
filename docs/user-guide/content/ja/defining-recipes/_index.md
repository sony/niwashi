---
title: "レシピを定義する"
weight: 5
---

# レシピを定義する

独自のレシピを定義する方法について説明します。このセクションは、レシピを作成・拡張したい方を対象としています。

初めてレシピを書く方は、[チュートリアル: はじめてのレシピ]({{< relref "first-recipe" >}}) から始めてください。レシピを書いてノードに適用するまでの一連の流れを15分程度で体験できます。

---

## Niwashiにおけるレシピ

レシピは、プロビジョニングや設定の手順をパッケージ化した実行可能なモジュールです。Stateファイルが「**何を**実現したいか」を宣言するのに対し、レシピは「**どうやって**実現するか」を定義します。

{{< mermaid >}}
flowchart LR
    subgraph input["入力"]
        state["State<br/>（何を実現したいか）"]
        recipe["レシピ<br/>（どうやって実現するか）"]
    end
    plan["nwsctl plan<br/>差分計算・実行計画の生成"]
    apply["nwsctl apply<br/>タスクの実行"]
    subgraph target["構築対象"]
        host["ホスト"]
        infra["インフラ<br/>（VM・インスタンス）"]
        node["ノード"]
        cluster["クラスタ"]
    end
    state --> plan
    recipe --> plan
    plan --> apply
    apply --> host
    apply --> infra
    apply --> node
    apply --> cluster
{{< /mermaid >}}

`nwsctl plan` は、Stateとレシピを読み込んで差分を計算し、レシピのタスクを依存関係に従って並べた実行計画を生成します。`nwsctl apply` がその計画を実行します。この仕組みの詳細は [アーキテクチャ]({{< relref "/getting-started/architecture" >}}) を参照してください。

---

## 設計の原則

### 1レシピ1機能

Niwashiのレシピは、**1つのレシピが1つの機能を提供する**ことを原則としています。

- 1つのレシピに複数の機能を詰め込まない
- 機能の組み合わせは、複数のレシピを `capabilities` に列挙することで実現する
- レシピを小さく保つことで、再利用性と保守性が高まる

---

## どの kind のレシピを書くか

レシピはトップレベルの `kind` フィールドで種別を宣言します。書くべき `kind` は「何をしたいか」で決まります。

| やりたいこと | `kind` | 実行場所 | ガイド |
|-------------|--------|----------|--------|
| ノードにソフトウェアをインストール・設定したい | `node` | 対象ノード上（リモート接続） | [Nodeのcapabilityを定義する]({{< relref "node-capability" >}}) |
| 複数ノードにまたがるクラスタを構築・設定したい | `cluster` | クラスタ全体（必要に応じてノード単位に分解） | [Clusterのcapabilityを定義する]({{< relref "cluster-capability" >}}) |
| VMやインスタンスを生成・破棄したい | `infra` | nwsctlを実行するホスト上 | [Infraのプロビジョナーを定義する]({{< relref "infrastructure-provisioning" >}}) |
| nwsctlを実行するホスト上のツールを、他のレシピから使えるようにしたい | `host` | nwsctlを実行するホスト上 | [Hostの設定を定義する]({{< relref "host-configuration" >}}) |
| 他のレシピから呼び出せるコマンド（Ansible実行など）を提供したい | `adapter` | 呼び出し元のレシピに従う | [アダプターを定義する]({{< relref "defining-adapters" >}}) |

---

## レシピはStateからどう参照されるか

作成したレシピは、`kind` に応じて異なる場所から参照されます。

| `kind` | 参照元 | 記述例 |
|--------|--------|--------|
| `node` | Stateの `inventory.nodes.<ノード名>.capabilities` | `- my-org.nginx` |
| `cluster` | Stateの `inventory.clusters.<クラスタ名>.capabilities` | `- cluster.kubernetes` |
| `infra` | Stateの `infrastructure.generators.<ジェネレーター名>.provisioner` | `provisioner: infra.vm.driver=vagrant` |
| `host` | Stateからは直接指定できない。他のレシピの `spec.requires` で依存宣言されたときに実行される | `requires: [host.tool.ansible]` |
| `adapter` | Stateからは直接指定できない。他のレシピの `spec.requires` で依存宣言し、タスクの `tool.run` の `toolRef` で呼び出す | `toolRef: adapter.tool.ansible` |

レシピ名の解決方法（`metadata.id` による直接指定、`spec.provides` のエイリアス、`attrs` による絞り込み）は [Capabilities とレシピの指定]({{< relref "/defining-desired-state/capabilities" >}}) と [レシピの読み込み仕様]({{< relref "recipe-loading" >}}) を参照してください。

---

## レシピの配置と読み込み

レシピは `nws-recipe.yaml`（1ファイル1レシピ）または `nws-catalog.yaml`（複数レシピをまとめて提供）として配置します。`nwsctl plan --recipe-dir` で指定されたディレクトリは再帰的に探索され、これらのファイルが見つかるとレシピとしてロードされます。同じIDとバージョンの組み合わせ（FQID）が重複した場合はエラーになります。

詳細は [複数のレシピをまとめる]({{< relref "catalog" >}}) と [レシピの読み込み仕様]({{< relref "recipe-loading" >}}) を参照してください。

---

## このセクションの内容

### 入門

| ページ | 内容 |
|--------|------|
| [チュートリアル: はじめてのレシピ]({{< relref "first-recipe" >}}) | レシピを書いて適用するまでの流れを体験する |

### 共通仕様

| ページ | 内容 |
|--------|------|
| [レシピのフォーマット仕様]({{< relref "recipe-spec" >}}) | `spec` 以下の全フィールドのリファレンス（kind別の注記つき） |
| [タスクの定義]({{< relref "defining-tasks" >}}) | タスクフィールド・アクション種別・dependsOn・operation のリファレンス |
| [stateChanges で State を更新する]({{< relref "state-changes" >}}) | タスク実行後にStateを更新する方法 |
| [タスクのフィルタリング条件 (where)]({{< relref "task-where" >}}) | `where` フィールドによるタスクの実行対象の絞り込み |
| [レシピで使える変数]({{< relref "recipe-variables" >}}) | テンプレート変数・環境変数のリファレンス |

### kind 別ガイド

| ページ | 内容 |
|--------|------|
| [Nodeのcapabilityを定義する]({{< relref "node-capability" >}}) | ノードに適用するレシピ（`kind: node`）の書き方 |
| [Clusterのcapabilityを定義する]({{< relref "cluster-capability" >}}) | クラスタに適用するレシピ（`kind: cluster`）の書き方 |
| [Infraのプロビジョナーを定義する]({{< relref "infrastructure-provisioning" >}}) | インフラ生成レシピ（`kind: infra`）の書き方 |
| [Hostの設定を定義する]({{< relref "host-configuration" >}}) | ホスト上で動作するinstallerレシピ（`kind: host`）の書き方 |
| [アダプターを定義する]({{< relref "defining-adapters" >}}) | アダプターレシピ（`kind: adapter`）の実装方法 |
| [アダプターを利用する]({{< relref "using-adapters" >}}) | レシピからアダプターを呼び出す方法 |

### その他

| ページ | 内容 |
|--------|------|
| [レシピ命名ガイドライン]({{< relref "recipe-naming" >}}) | レシピID・ケーパビリティエイリアス・コマンド名の命名規則 |
| [複数のレシピをまとめる]({{< relref "catalog" >}}) | nws-catalog.yaml の使い方 |
| [レシピの読み込み仕様]({{< relref "recipe-loading" >}}) | FQID・バージョン管理・別名解決の詳細 |

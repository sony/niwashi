---
title: "nwsctl plan"
weight: 2
---

# nwsctl plan

現在の状態と目標の状態を比較し、実行計画を作成します。

```
nwsctl plan [flags]
```

---

## 概要

`nwsctl plan` は、ワークスペースに保存された現在の状態（`state.json`）と `--target` で指定した目標の状態を比較し、差分を埋めるためのタスクDAGを計算して `plan.json` に書き出します。

---

## フラグ

| フラグ | 省略形 | デフォルト | 説明 |
|-------|-------|----------|------|
| `--target` | `-t` | | 目標のStateファイルのパス（複数指定可） |
| `--recipe-dir` | | | レシピディレクトリのパス（複数指定可） |
| `--out` | `-o` | `plan.json` | 生成する計画ファイルの保存先 |
| `--destroy` | | `false` | 全リソースを削除する計画を作成する |
| `--prune` | | `false` | 削除操作のみを含む計画を作成する |
| `--profile` | | | プロファイルファイルのパス |
| `--with-init` | | `false` | 空の状態で初期化してから計画を作成する |
| `--work-dir` | | `.niwashi` | ワークスペースのパス |

---

## 使い方

### 基本的な使い方

```bash
nwsctl plan \
  -t infra.yaml \
  -t target.yaml \
  --recipe-dir ./recipes
```

複数の `--target` を指定すると、左から順にマージされた状態が目標の状態として使われます。詳細は [Stateマージの詳細]({{< relref "defining-desired-state/advanced/state-merging" >}}) を参照してください。

### 適用済みのCapabilityを更新する

通常の `nwsctl plan` は、前回の適用以降に `params` またはレシピのバージョンが変わったCapabilityも検出し、その `operation: update` のタスクを計画に含めます。追加のフラグは不要です。

変更されたCapabilityのレシピに `operation: update` のタスクがない場合、変更は適用されず、代わりにplanの出力に一覧表示されます。

```
Pending Updates (recipe has no update tasks):
- node:my-org/myapp@1.0.0:web-1
```

詳細は [更新処理（operation: update）]({{< relref "defining-recipes/defining-tasks#更新処理operation-update" >}}) を参照してください。

### 出力先を変更する

```bash
nwsctl plan -t target.yaml --recipe-dir ./recipes -o my-plan.json
```

生成された `plan.json` を `nwsctl apply` に渡します。

### 全リソースを削除する計画（--destroy）

```bash
nwsctl plan --recipe-dir ./recipes --destroy
```

目標の状態を空として計算します。現在の状態にあるすべてのリソースを削除する計画が生成されます。実行には `nwsctl apply` にも `--destroy` の指定が必要です（[nwsctl apply]({{< relref "apply" >}}) を参照）。

### 不要なリソースだけを削除する計画（--prune）

```bash
nwsctl plan -t target.yaml --recipe-dir ./recipes --prune
```

目標の状態に含まれないリソースの削除のみを計画します。新規追加の操作は含まれません。実行には `nwsctl apply` にも `--prune` の指定が必要です（[nwsctl apply]({{< relref "apply" >}}) を参照）。

### 初回セットアップをまとめて行う（--with-init）

```bash
nwsctl plan -t target.yaml --recipe-dir ./recipes --with-init
```

`nwsctl init` を実行してから計画を作成します。ワークスペースがまだ存在しない場合に便利です。

---

## プロファイル（--profile）

プロファイルファイルを使うと、計画の挙動をカスタマイズできます。

```yaml
version: nws.profile/v1
name: dev

# ツールの別名（capability名 → レシピのFQIDへのマッピング）
toolAlias:
  infra.vm: infra.vm.driver=vagrant   # infra.vm を Vagrant に解決する

# Capabilityの紐付け（任意のcapability名 → レシピのFQIDへのマッピング）
capabilityBinding:
  my-org.nginx: my-org/nginx@1.0.0

# パラメータの上書き
params:
  capability:
    my-org.nginx:
      nginx_port: 8080
```

```bash
nwsctl plan -t target.yaml --recipe-dir ./recipes --profile profile-dev.yaml
```

---

## レシピの整合性チェック

計画を作成すると、nwsctlは使用したレシピのフィンガープリント（ハッシュ値）を計画ファイルに記録します。`nwsctl apply` は、このフィンガープリントを `--recipe-dir` のレシピと照合します。不一致の場合の動作は [レシピの整合性チェック]({{< relref "apply#レシピの整合性チェック" >}}) を参照してください。

---

## 次のステップ

- [nwsctl apply]({{< relref "apply" >}}) — 作成した計画を実行する

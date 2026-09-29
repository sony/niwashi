---
title: "ワークフロー"
weight: 2
---

# ワークフロー

Niwashiを使った基本的な作業の流れを説明します。

---

## 基本的なワークフロー

Niwashiの作業は、以下のステップで構成されます：

1. **準備**: レシピのセットアップ
2. **初期化**: ワークスペースを作成（`nwsctl init`）
3. **目標状態の定義**: YAMLファイルで記述
4. **計画の作成**: 実行計画を生成（`nwsctl plan`）
5. **計画の実行**: 変更を適用（`nwsctl apply`）

---

## 準備: レシピのセットアップ

Niwashiを使用する前に、必要なレシピをローカルにセットアップします。

### レシピの取得

```bash
# 作業ディレクトリを作成
mkdir -p ~/niwashi-project
cd ~/niwashi-project

# レシピをクローン
git clone <レシピリポジトリのURL> recipes
```

リファレンスレシピを使用する場合は、`${REFERENCE_RECIPE_REPO}` を指定します。独自のレシピリポジトリがある場合は、そのURLを指定してください。

### レシピディレクトリの構成例

```
~/niwashi-project/
├── recipes/          # レシピディレクトリ
│   ├── ansible/
│   ├── kubernetes/
│   └── vagrant/
└── state.yaml        # 目標状態ファイル（後で作成）
```

---

## ステップ1: 初期化

### 新規プロジェクトの初期化

新しいプロジェクトを始める場合：

```bash
nwsctl init
```

これにより、Niwashiのワークスペース（デフォルト: `.niwashi`）が作成されます。

### 既存状態のインポート

既存の状態をインポートする場合：

```bash
nwsctl init -s existing-state.yaml
# または
nwsctl init --state existing-state.yaml
```

---

## ステップ2: 目標状態の定義

YAMLファイルで目標の状態を記述します。

例: `state.yaml`

```yaml
version: nws.state/v1

inventory:
  nodes:
    web-server:
      capabilities:
        - web.nginx

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 1
```

詳細は [状態を定義する]({{< relref "defining-desired-state" >}}) を参照してください。

---

## ステップ3: 計画の作成

目標状態から実行計画を作成します：

```bash
nwsctl plan \
  --recipe-dir ./recipes \
  -t state.yaml
```

**オプション**:
- `--recipe-dir`: レシピが含まれるディレクトリ（複数指定可能）
- `-t, --target`: 目標状態ファイル（複数指定可能）

Niwashiは、現在の状態と目標の状態を比較し、必要な変更を計画します。

### 複数のStateファイルをマージ

複数のファイルをマージして計画を作成できます：

```bash
nwsctl plan \
  --recipe-dir ./recipes \
  -t infra-dev.yaml \
  -t app.yaml
```

ファイルは左から右の順でマージされます。

### 複数のレシピディレクトリを指定

```bash
nwsctl plan \
  --recipe-dir ./recipes \
  --recipe-dir ./custom-recipes \
  -t state.yaml
```

---

## ステップ4: 計画の実行

計画を実行して、目標の状態を実現します：

```bash
nwsctl apply \
  --plan plan.json \
  --recipe-dir ./recipes
```

**オプション**:
- `--recipe-dir`: レシピが含まれるディレクトリ（複数指定可能）
- `--plan`: 実行する計画ファイル（必須）

Niwashiは、計画に従ってインフラのプロビジョニングと設定管理を実行します。

---

## リソースの削除

### 削除計画の作成

すべてのリソースを削除する計画を作成します：

```bash
nwsctl plan \
  --recipe-dir ./recipes \
  --destroy
```

`--destroy` フラグを指定すると、目標状態を空として扱い、現在のすべてのリソースを削除する計画が生成されます。

### 削除計画の実行

生成された削除計画を実行します。削除計画の適用には `apply` にも `--destroy` フラグが必要です：

```bash
nwsctl apply \
  --plan plan.json \
  --recipe-dir ./recipes \
  --destroy
```

実行時に、リソースを削除してよいかの確認プロンプトが表示されます（`--yes` を付けるとスキップできます）。Niwashiは、計画に従ってリソースを削除します。

### 注意事項

- **レシピの対応**: すべてのレシピが削除処理に対応しているわけではありません。レシピによっては、正しく削除されない場合があります。
- **確認**: 削除前に `plan.json` を確認し、意図しないリソースが削除されないか確認してください。
- **バックアップ**: 重要なデータは事前にバックアップを取ることをお勧めします。

---

## ワークスペースと状態管理

### ワークスペースの場所

Niwashiは、ワークスペース（`--work-dir`で指定）に以下を保存します：

- **現在の状態**: Niwashiが管理している現在の状態
- **レシピのデータ**: レシピが生成したファイルやデータ

デフォルトのワークスペース: `.niwashi`（カレントディレクトリ配下）

### 現在の状態の管理

現在の状態は、Niwashiの管理下にあり、直接編集することはできません。状態を変更するには、目標状態を定義して `plan` → `apply` を実行します。

### export/import の制約

`nwsctl export` で現在の状態をエクスポートし、別のワークスペースで `init -s` することは可能ですが、レシピによっては正しく動作しない場合があります。

**例**: Vagrantレシピは、VMのイメージなどをワークスペース配下に作成します。これらのファイルは `export` に含まれないため、インポート後に再度プロビジョニングが必要になることがあります。

---

## 典型的な使い方

### 開発サイクル

```bash
# 1. 目標状態を修正（state.yamlを編集）

# 2. 計画を確認
nwsctl plan --recipe-dir ./recipes -t state.yaml

# 3. 適用
nwsctl apply --plan plan.json --recipe-dir ./recipes

# 4. 動作を確認

# 5. 1に戻る
```

### 環境の切り替え

```bash
# 開発環境
nwsctl plan --recipe-dir ./recipes -t infra-dev.yaml -t app.yaml
nwsctl apply --plan plan.json --recipe-dir ./recipes

# 本番環境
nwsctl plan --recipe-dir ./recipes -t infra-prod.yaml -t app.yaml
nwsctl apply --plan plan.json --recipe-dir ./recipes
```

---

## 次のステップ

- [アーキテクチャ]({{< relref "architecture" >}}) - Niwashiの内部構造
- [状態を定義する]({{< relref "defining-desired-state" >}}) - Stateファイルの書き方
- [計画の実行]({{< relref "execution" >}}) - plan/applyの詳細

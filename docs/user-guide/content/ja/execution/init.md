---
title: "nwsctl init"
weight: 1
---

# nwsctl init

ワークスペースを初期化します。`nwsctl plan` や `nwsctl apply` を実行する前に一度だけ実行します。

```
nwsctl init [flags]
```

---

## 概要

`nwsctl init` は、Niwashiが使用するワークスペースディレクトリと現在の状態ファイル（`state.json`）を作成します。

作成されるディレクトリ構造：

```
.niwashi/          # ワークスペース（--work-dir で変更可能）
├── state/
│   └── state.json   # 現在の状態（Niwashiが管理）
├── store/
└── runs/
```

`state.json` はNiwashiが内部的に管理するファイルで、直接編集する必要はありません。ユーザーが定義する目標の状態（Stateファイル）とは別物です。

---

## フラグ

| フラグ | 省略形 | デフォルト | 説明 |
|-------|-------|----------|------|
| `--work-dir` | | `.niwashi` | ワークスペースのパス |
| `--state` | `-s` | （なし） | 初期状態として読み込むStateファイルのパス |

---

## 使い方

### 新規セットアップ（空の状態から始める）

```bash
nwsctl init
```

空の状態が `.niwashi/state/state.json` に作成されます。初めてNiwashiを使う場合はこれで始めます。

### 既存のStateから始める

```bash
nwsctl init -s current.yaml
```

指定したStateファイルを現在の状態として読み込みます。`nwsctl export` で書き出した状態を引き継ぐ場合などに使います。

### ワークスペースの場所を変更する

```bash
nwsctl init --work-dir /path/to/workspace
```

デフォルトの `.niwashi` 以外の場所にワークスペースを作成します。`nwsctl plan` や `nwsctl apply` でも同じ `--work-dir` を指定する必要があります。

---

## 注意事項

- すでに初期化済みのワークスペースに対して実行するとエラーになります
- 対応するファイル形式: `.yaml`、`.yml`、`.json`、`.jsonc`

### `-s` で引き継ぐ場合の制約

`nwsctl export` でエクスポートしたStateを `-s` で読み込むことは可能ですが、レシピによっては正しく動作しない場合があります。

**例**: Vagrantレシピは、VMのイメージなどをワークスペース配下に作成します。これらのファイルは `export` に含まれないため、インポート後に再度プロビジョニングが必要になることがあります。

---

## 次のステップ

- [nwsctl plan]({{< relref "plan" >}}) — 実行計画を作成する

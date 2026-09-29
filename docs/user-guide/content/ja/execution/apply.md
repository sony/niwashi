---
title: "nwsctl apply"
weight: 3
---

# nwsctl apply

`nwsctl plan` で生成した計画を実行し、インフラを目標の状態に遷移させます。

```
nwsctl apply --plan <plan-file> [flags]
```

---

## 概要

`nwsctl apply` は `plan.json` に記述されたタスクDAGを順番に実行します。タスクが成功するとStateが更新され、次回の `nwsctl plan` では適用済みとして扱われます。

---

## フラグ

| フラグ | 省略形 | デフォルト | 説明 |
|-------|-------|----------|------|
| `--plan` | | （必須） | 実行する計画ファイルのパス |
| `--recipe-dir` | | | レシピディレクトリのパス（複数指定可） |
| `--work-dir` | | `.niwashi` | ワークスペースのパス |
| `--dry-run` | | `off` | ドライランモード（`off` / `simulate`） |
| `--destroy` | | `false` | 削除計画（`plan --destroy` で生成）を実行する |
| `--prune` | | `false` | 不要リソースを削除する計画（`plan --prune` で生成）を実行する |
| `--yes` | | `false` | 削除系計画（`--destroy` / `--prune`）の確認プロンプトをスキップする |
| `--strict-check` | | `false` | 計画作成後にレシピが変更されていた場合にエラーで停止する（後述） |

---

## 使い方

### 基本的な使い方

```bash
nwsctl apply --plan plan.json --recipe-dir ./recipes
```

### ドライラン（--dry-run simulate）

```bash
nwsctl apply --plan plan.json --recipe-dir ./recipes --dry-run simulate
```

実際の処理を実行せずに計画を流します。レシピスクリプト内では `NWS_DRY_RUN_ENABLED=1` が設定されるため、レシピ側でこの変数を参照してドライラン時の動作（構文チェックのみ実行するなど）を実装できます。

### 削除系計画の実行（--destroy / --prune）

`nwsctl plan --destroy` で生成した削除計画を実行するには、`apply` にも `--destroy` を指定します。同様に、`plan --prune` で生成した不要リソースを削除する計画には `--prune` を指定します。フラグなしでこれらの計画を適用しようとするとエラーになります。

```bash
# 削除計画の実行
nwsctl apply --plan plan.json --recipe-dir ./recipes --destroy

# 不要リソースを削除する計画の実行
nwsctl apply --plan plan.json --recipe-dir ./recipes --prune
```

いずれの場合も、リソースを削除してよいかの確認プロンプトが表示されます。CIなどの非対話環境では `--yes` を付けてスキップできます（ドライラン時はプロンプトは表示されません）。

```bash
nwsctl apply --plan plan.json --recipe-dir ./recipes --destroy --yes
```

---

## レシピの整合性チェック

`nwsctl apply` は `--recipe-dir` のレシピが計画ファイルに記録されたフィンガープリントと一致するかを確認します。計画作成後にレシピが変更されていた場合（CIパイプラインで plan と apply の間にレシピが更新された場合など）を検出できます。

デフォルトでは、不一致があっても警告をログに記録して処理を続行します。`--strict-check` を付けると、不一致をエラーとして扱い処理を停止します。

```bash
nwsctl apply --plan plan.json --recipe-dir ./recipes --strict-check
```

更新したレシピで apply するには、`nwsctl plan` を再実行して新しい計画を作成してください。

---

## SSH接続設定

`nwsctl apply` がSSH経由でインスタンスに接続する際のタイムアウト・リトライ動作は、以下の環境変数で調整できます。ネットワークが遅い環境や、クラウドVMが作成後に起動するまで時間がかかる場合のチューニングに使います。

| 環境変数 | デフォルト | 説明 |
|---------|-----------|------|
| `NWS_SSH_CONNECT_TIMEOUT_SEC` | `10` | TCP接続タイムアウト（秒） |
| `NWS_SSH_HANDSHAKE_TIMEOUT_SEC` | `10` | SSHハンドシェイクタイムアウト（秒） |
| `NWS_SSH_RETRY_MAX_COUNT` | `3` | 接続試行回数の上限 |
| `NWS_SSH_RETRY_INTERVAL_SEC` | `5` | リトライ間隔（秒） |

```bash
NWS_SSH_RETRY_MAX_COUNT=10 NWS_SSH_RETRY_INTERVAL_SEC=10 \
  nwsctl apply --plan plan.json --recipe-dir ./recipes
```

---

## 次のステップ

- [nwsctl ssh]({{< relref "ssh" >}}) — 適用後にノードへSSH接続する
- [nwsctl export]({{< relref "export" >}}) — 現在のStateをファイルに書き出す

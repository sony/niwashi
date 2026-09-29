---
title: "アダプターを定義する"
weight: 4
---

# アダプターを定義する

この ページでは `kind: adapter` のレシピ（アダプター）の書き方を説明します。アダプターは他のレシピから `tool.run` アクションで呼び出せるコマンドを公開します。

アダプターの利用方法については [アダプターを利用する]({{< relref "using-adapters" >}}) を参照してください。

---

## `kind: adapter` のレシピとは

`kind: adapter` のレシピは、**他のレシピから `tool.run` で呼び出せるコマンドを公開する**レシピです。

{{< mermaid >}}
flowchart LR
    subgraph host["ホスト（nwsctl）"]
        caller["呼び出し元レシピ<br/>（kind: node / cluster / infra / host）"]
        adapter["kind: adapter のレシピ<br/>（コマンドを公開）"]
        tool["ホスト上のツール<br/>（例: ansible-playbook）"]
    end
    caller -->|"tool.run（toolRef / command）"| adapter
    adapter -->|"記録されたパスでツールを実行"| tool
{{< /mermaid >}}

アダプターは通常、以下の2つのレシピをセットで定義します。

| レシピ | `kind` | 役割 |
|-------|--------|------|
| **installerレシピ** | `host` | ホストにツールが存在するか確認し、実行パスをStateに記録する |
| **adapterレシピ** | `adapter` | 記録された実行パスを使ってツールを呼び出すコマンドを公開する |

installerレシピが `spec.requires` の依存対象（例: `host.tool.ansible`）を提供し、adapterレシピがその依存が満たされた後に `tool.run` から呼び出されます。

---

## フィールドリファレンス

### 共通フィールド

`spec` 以下の共通フィールドの詳細は [レシピのフォーマット仕様]({{< relref "recipe-spec" >}}) を参照してください。`kind: adapter` 固有のフィールド（`spec.allowedScope`、`spec.executionUnit`、`spec.commands`）についても同ページに記載しています。

タスクの書き方は [タスクの定義]({{< relref "defining-tasks" >}}) を参照してください。`kind: adapter` では `exec.local` のみ使用可能です。

### テンプレート変数

`kind: adapter` のタスクでは `{{ .Target }}` は呼び出し元の実行対象ID（nodeID 等）に展開されます。`{{ .Runtime.tool.<name>.<field> }}` で installerレシピが記録したツールのパスを参照できます。

その他のテンプレート変数と実行時環境変数の一覧は [レシピで使える変数]({{< relref "recipe-variables" >}}) を参照してください。

---

## 最小構成のサンプル

```yaml
version: nws.recipe/v1
kind: adapter
metadata:
  id: my-org/my-tool.adapter
  version: "1.0.0"
  description: "Adapter for my-tool"
spec:
  provides:
    - name: adapter.tool.my-tool   # tool.run の toolRef で指定する名前

  allowedScope: node                # 呼び出し元レシピの kind を制約（必須）

  commands:
    default:
      task: run
    run:
      task: run
      description: "Run my-tool"

  tasks:
    - name: run
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            exec {{ .Runtime.tool.my_tool.path }} "$@"
```

---

## パターン：installerレシピとセットで使う

adapterレシピが `{{ .Runtime.tool.<name>.path }}` を参照するためには、installerレシピがそのパスをStateに記録しておく必要があります。

**installerレシピ（`kind: host`）**

```yaml
spec:
  provides:
    - name: host.tool.my-tool

  runtime:
    type: tool
    name: my_tool   # テンプレートのマップキーとして使われるため、ハイフン・ドットは使用不可

  tasks:
    - name: check
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            BIN="$(command -v my-tool || { echo 'my-tool not found' >&2; exit 1; })"
            jq -n --arg path "$BIN" '{ path: $path }' \
              > "{{ .Outputs }}/tool_info.json"
      stateChanges:
        record-path:
          op: set
          path: "/"
          valueFromFile: "{{ .Outputs }}/tool_info.json"
```

installerレシピが実行されると、Stateの `/runtime/tool/my_tool` に `{ "path": "/usr/bin/my-tool" }` が書き込まれます。

**adapterレシピからの参照**

```yaml
spec:
  requires:
    - host.tool.my-tool   # installerレシピを依存として宣言

  tasks:
    - name: run
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            exec {{ .Runtime.tool.my_tool.path }} "$@"
```

> **重要**: `requires` をadapterレシピにのみ宣言しても、installerジョブはスケジュールされません。adapterのタスクは別ジョブとしてスケジュールされるのではなく、**呼び出し元（caller）のジョブの中にインラインで展開される**ため、adapter自身の `requires` は評価されません。実際に `tool.run` を呼び出すレシピ——ここではcaller——が自分自身で `requires: [host.tool.<name>]` を宣言する必要があります。宣言しないと、adapterのタスク実行時に `{{ .Runtime.tool.<name>.path }}` が解決されません。両方の宣言を含む動作例はリポジトリ内の `examples/recipe/adapter/` を参照してください。

---

## 実例：Ansibleアダプター

`recipe/ansible/` ディレクトリに実際のアダプター実装例があります。

- `ansible-installer.yaml` — `kind: host` のレシピ。Ansibleの存在確認と実行パスの記録（`stateChanges` で `/runtime/tool/ansible` に書き込み）
- `ansible-adapter.yaml` — `kind: adapter` のレシピ。`{{ .Runtime.tool.ansible.playbook.path }}` を参照してコマンドを実行

---

## 次のステップ

- [アダプターを利用する]({{< relref "using-adapters" >}}) — 定義したアダプターをレシピから呼び出す方法
- [レシピのフォーマット仕様]({{< relref "recipe-spec" >}}) — `allowedScope`・`executionUnit`・`commands` の詳細

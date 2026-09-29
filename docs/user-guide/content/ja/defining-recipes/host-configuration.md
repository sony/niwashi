---
title: "Hostの設定を定義する"
weight: 7
---

# Hostの設定を定義する

nwsctl を実行するホスト上で動作するレシピ（`kind: host`）の書き方を説明します。

---

## `kind: host` のレシピとは

`kind: host` のレシピは、**nwsctlを実行するホストマシン上で処理を実行する**レシピです。

{{< mermaid >}}
flowchart LR
    subgraph host["ホスト（nwsctl）"]
        other["他のレシピ<br/>（spec.requires で依存宣言）"]
        recipe["kind: host のレシピ（installer）"]
        tool["ホスト上のツール（例: Ansible）"]
        state["State（/runtime/tool/...）"]
    end
    other -->|"依存として実行をトリガー"| recipe
    recipe -->|"検出"| tool
    recipe -->|"パス等を記録"| state
{{< /mermaid >}}

`kind: host` のレシピは、Stateファイルの `capabilities` に直接指定することはできません。**他のレシピの `spec.requires` で依存宣言されたときのみ**実行されます。

```yaml
spec:
  requires:
    - host.tool.ansible   # このレシピが kind: host のinstallerレシピを呼び出す
```

プランナーはこの依存を解決し、installerレシピを依存元のレシピより先に実行します。

現在の主な用途は **installerレシピ** です。ホストにインストール済みのツール（Ansible、git 等）の実行パスをStateに記録し、アダプターレシピがそのパスを参照できるようにします。

```
kind: host の installerレシピ
  → ツールのパスを stateChanges で /runtime/tool/<name> に記録
      → kind: adapter のアダプターレシピが {{ .Runtime.tool.<name>.path }} で参照
```

---

## フィールドリファレンス

### 共通フィールド

`spec` 以下の共通フィールドの詳細は [レシピのフォーマット仕様]({{< relref "recipe-spec" >}}) を参照してください。`kind: host` では `spec.requires` に `host` の依存のみ宣言できます。

タスクの書き方は [タスクの定義]({{< relref "defining-tasks" >}})、State の更新は [stateChanges で State を更新する]({{< relref "state-changes" >}}) を参照してください。

### spec.runtime

`stateChanges` を使う場合、`spec.runtime` でホスト上で管理するリソースを宣言します。

```yaml
spec:
  runtime:
    type: tool      # リソース種別（tool / service）
    name: ansible   # リソース名（英数字・アンダースコアのみ。先頭は英字かアンダースコア）
```

| フィールド | 必須 | 説明 |
|-----------|------|------|
| `type` | Yes | `tool`（ホスト上のCLIツール）または `service`（ローカルサービス） |
| `name` | Yes | リソース名。英数字とアンダースコアのみ使用可（先頭は英字かアンダースコア）。この名前はGoテンプレートのマップキー（`{{ .Runtime.tool.<name>... }}`）として使われるため、ハイフンやドットは使用できません。 |

`spec.runtime` を宣言すると、`stateChanges` の書き込み先は `/runtime/<type>/<name>` 以下に自動的に制限されます。`stateChanges` を使わない場合は省略できます。

### テンプレート変数と `where` の仕様

`kind: host` のタスクでは `{{ .Target }}` は空文字列です。`where` を使う場合、評価コンテキストはホスト自身の `os`・`arch` です（ノードの情報ではありません）。

その他のテンプレート変数と実行時環境変数の一覧は [レシピで使える変数]({{< relref "recipe-variables" >}}) を参照してください。`where` フィールドの詳細は [タスクのフィルタリング条件 (where)]({{< relref "task-where" >}}) を参照してください。

### Stateに書き込むデータ構造

`stateChanges` でStateに書き込む値は `/runtime/<type>/<name>` 以下に格納されます。値の構造は自由形式ですが、リファレンスとなるinstallerレシピでは以下の規約が使われています。

**`type: tool` の場合**

少なくとも `path` フィールドを持つJSONオブジェクトです。複数の実行ファイルを持つツールはネストしたオブジェクトを使います。

```json
{ "path": "/usr/local/bin/my-tool", "version": "1.2.3" }
```

```json
{
  "playbook": { "path": "/usr/local/bin/ansible-playbook", "version": "2.16.3" },
  "galaxy":   { "path": "/usr/local/bin/ansible-galaxy" }
}
```

記録した値はアダプターレシピから `{{ .Runtime.tool.<name>.<field> }}` で参照できます。

**`type: service` の場合**

固定の規約はありません。消費側のレシピが必要とする接続情報や設定情報を自由に格納してください。

---

## 最小構成のサンプル

```yaml
version: nws.recipe/v1
kind: host
metadata:
  id: my-org/my-tool.installer
  version: "1.0.0"
  description: "Check my-tool installation"
spec:
  provides:
    - name: host.tool.my-tool

  runtime:
    type: tool
    name: my-tool

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

---

## パターン：installerレシピ

### ツールの検出と実行パスの記録

ホストにインストール済みのツールを検出し、実行パスとバージョン情報をStateに書き込みます。

```yaml
tasks:
  - name: check
    action:
      exec.local:
        scriptTpl: |
          #!/bin/bash
          set -euo pipefail
          BIN="$(command -v my-tool || { echo 'my-tool not found in PATH' >&2; exit 1; })"
          VER="$(my-tool --version | head -n1)"
          jq -n \
            --arg version "$VER" \
            --arg path "$BIN" \
            '{ version: $version, path: $path }' \
            > "{{ .Outputs }}/tool_info.json"
    stateChanges:
      record-path:
        op: set
        path: "/"
        valueFromFile: "{{ .Outputs }}/tool_info.json"
```

### バージョン要件のチェック

`spec.defaults.params` でバージョン要件を受け取り、チェックする例です。

```yaml
defaults:
  params:
    min_version: ""   # 省略時はバージョンチェックしない

tasks:
  - name: check
    action:
      exec.local:
        scriptTpl: |
          #!/bin/bash
          set -euo pipefail
          BIN="$(command -v my-tool || { echo 'my-tool not found' >&2; exit 1; })"
          VER="$(my-tool --version | grep -oP '[0-9]+\.[0-9]+\.[0-9]+')"
          if [ -n "{{ .Params.min_version }}" ]; then
            case "$VER" in
              {{ .Params.min_version }}* ) : ;;
              * ) echo "my-tool version $VER does not meet minimum {{ .Params.min_version }}" >&2; exit 1 ;;
            esac
          fi
          jq -n --arg path "$BIN" --arg version "$VER" \
            '{ path: $path, version: $version }' \
            > "{{ .Outputs }}/tool_info.json"
    stateChanges:
      record-path:
        op: set
        path: "/"
        valueFromFile: "{{ .Outputs }}/tool_info.json"
```

---

## 実例：リファレンスinstallerレシピ

`recipe/` 配下のリファレンス `kind: host` レシピを参考にしてください。

| ファイル | `provides` | 概要 |
|---------|-----------|------|
| `recipe/ansible/ansible-installer.yaml` | `host.tool.ansible` | ansible-playbook / ansible-galaxy の検出とパス記録 |
| `recipe/nws/tools/git.yaml` | `host.tool.git` | git の検出とパス記録 |
| `recipe/nws/tools/jq.yaml` | `host.tool.jq` | jq の検出とパス記録 |

---

## 次のステップ

- [stateChanges で State を更新する]({{< relref "state-changes" >}}) — `stateChanges` の詳細
- [タスクのフィルタリング条件 (where)]({{< relref "task-where" >}}) — `where` フィールドの詳細

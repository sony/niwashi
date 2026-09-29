---
title: "Nodeのcapabilityを定義する"
weight: 1
---

# Nodeのcapabilityを定義する

ノードに適用するCapabilityのレシピ（`kind: node`）の書き方を説明します。

---

## `kind: node` のレシピとは

`kind: node` のレシピは、**単一のノードに対して処理を実行する**レシピです。ノードごとに独立して実行されるため、ノード固有の設定やソフトウェアのインストールに適しています。

{{< mermaid >}}
flowchart LR
    subgraph host["ホスト（nwsctl）"]
        recipe["kind: node のレシピ"]
    end
    subgraph targets["実行対象"]
        n1["node-1"]
        n2["node-2"]
    end
    recipe -->|"ノードごとに独立して実行"| n1
    recipe --> n2
{{< /mermaid >}}

ユーザーがStateファイルでCapabilityを指定すると、Niwashiは該当するレシピを見つけ、対象ノードに対してレシピのタスクを実行します。

---

## フィールドリファレンス

### 共通フィールド

`spec` 以下の共通フィールドの詳細は [レシピのフォーマット仕様]({{< relref "recipe-spec" >}}) を参照してください。`kind: node` では `spec.requires` に `host` と `node` の依存を宣言できます。

タスクの書き方は [タスクの定義]({{< relref "defining-tasks" >}})、State の更新は [stateChanges で State を更新する]({{< relref "state-changes" >}}) を参照してください。

### テンプレート変数

`kind: node` のタスクでは、`{{ .Target }}` は**実行対象のノードID**に展開されます。スクリプト内でノードを識別したり、store のキーに使うときに便利です。

その他のテンプレート変数と実行時環境変数の一覧は [レシピで使える変数]({{< relref "recipe-variables" >}}) を参照してください。

---

## 最小構成のサンプル

```yaml
version: nws.recipe/v1
kind: node
metadata:
  id: my-org/my-feature
  version: "1.0.0"
  description: "My feature recipe for nodes"
spec:
  provides:
    - name: my-org.my-feature
  tasks:
    - name: install
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            echo "Installing my-feature on {{ .Target }}"
```

---

## 実例：パッケージのインストールと削除

`exec.remote` でノードに直接 SSH して apt パッケージをインストール・削除する例です。

```yaml
version: nws.recipe/v1
kind: node
metadata:
  id: my-org/nginx
  version: "1.0.0"
  description: "Install nginx"
spec:
  provides:
    - name: my-org.nginx

  defaults:
    params:
      version: ""   # 空文字の場合は最新版をインストール

  tasks:
    - name: install
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            PKG="nginx{{ if .Params.version }}={{ .Params.version }}{{ end }}"
            apt-get install -y "$PKG"

    - name: uninstall
      operation: destruct
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            apt-get purge -y nginx
            apt-get autoremove -y
```

---

## 次のステップ

- [タスクの定義]({{< relref "defining-tasks" >}}) — アクション種別・dependsOn・operation の詳細
- [stateChanges で State を更新する]({{< relref "state-changes" >}}) — store への書き込みと冪等性管理
- [アダプターを利用する]({{< relref "using-adapters" >}}) — `tool.run` でアダプターを呼び出す詳細

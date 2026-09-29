---
title: "チュートリアル: はじめてのレシピ"
weight: -1
---

# チュートリアル: はじめてのレシピ

このチュートリアルでは、最小構成のレシピをゼロから書き、実際にノードへ適用するまでの一連の流れを体験します。所要時間は15分程度です。

作るのは、ノード上にあいさつ文を書き込むだけの `kind: node` のレシピです。単純な題材ですが、このチュートリアルを終えると以下がひととおり分かります。

- レシピファイル（`nws-recipe.yaml`）の基本構造
- レシピをStateから参照する方法
- `nwsctl plan` → `nwsctl apply` の実行サイクル
- 削除処理（destruct）の書き方

---

## 前提条件

- `nwsctl` がインストールされていること（[インストール]({{< relref "/getting-started/installation" >}})）
- SSH接続できるLinuxマシンが1台あること（VM・クラウドインスタンス・物理マシンのいずれでも可）
  - ホスト名（またはIPアドレス）、ユーザー名、秘密鍵のパスを手元に用意してください
  - 一度 `ssh` コマンドで接続したことがあり、ホスト鍵が `~/.ssh/known_hosts` に登録されていること（Niwashiは接続時にホスト鍵を検証します）

---

## ステップ1: 作業ディレクトリを準備する

以下の構成で作業ディレクトリを作成します。

```
my-first-recipe/
├── recipes/
│   └── hello/
│       └── nws-recipe.yaml   ← これから書くレシピ
└── state.yaml                ← これから書くState
```

```bash
mkdir -p my-first-recipe/recipes/hello
cd my-first-recipe
```

---

## ステップ2: レシピを書く

`recipes/hello/nws-recipe.yaml` を以下の内容で作成します。

```yaml
version: nws.recipe/v1
kind: node
metadata:
  id: my-org/hello
  version: "1.0.0"
  description: "My first recipe"
spec:
  provides:
    - name: my-org.hello
  tasks:
    - name: greet
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            echo "Hello, Niwashi! (node: {{ .Target }})" > /tmp/hello-niwashi.txt
            cat /tmp/hello-niwashi.txt

    - name: cleanup
      operation: destruct
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            rm -f /tmp/hello-niwashi.txt
```

各ブロックの意味は次の通りです。

| ブロック | 意味 |
|---------|------|
| `kind: node` | このレシピが単一ノードを対象にすることを宣言します。レシピの種別の選び方は [レシピを定義する]({{< relref "_index" >}}) を参照 |
| `metadata.id` / `metadata.version` | レシピの識別子。`<org>/<name>` 形式で記述します |
| `spec.provides` | このレシピが提供するCapabilityの名前。Stateからはこの名前で参照します |
| `spec.tasks` | 実行する処理の一覧。`exec.remote` はノードにリモート接続してスクリプトを実行するアクションです |
| `{{ .Target }}` | 実行対象のノード名に展開されるテンプレート変数 |
| `operation: destruct` | 削除時（`plan --destroy`）にのみ実行されるタスク。構築処理の逆操作を書きます |

---

## ステップ3: Stateを書く

次に、このレシピを「どのノードに適用するか」を宣言するStateファイルを書きます。`state.yaml` を以下の内容で作成し、`connection` 以下をお使いのマシンの接続情報に書き換えてください。

```yaml
version: nws.state/v1

inventory:
  nodes:
    hello-node:
      instanceSelector:
        generator: my-servers
      capabilities:
        - my-org.hello

infrastructure:
  generators:
    my-servers:
      provisioner: external-instance
      params:
        instances:
          server-01:
            connection:
              ssh:
                address:
                  host: 192.168.1.10    # ← 接続先に合わせて変更
                  port: 22
                  user: ubuntu          # ← 接続先に合わせて変更
                auth:
                  method: privateKey
                  privateKeyPath: ~/.ssh/id_rsa   # ← 接続先に合わせて変更
                hostKey:
                  knownHostsPath: ~/.ssh/known_hosts   # ← ホスト鍵検証に使うファイル
```

このStateは2つのことを宣言しています。

- `infrastructure.generators`: 既存のマシンをインスタンスとして登録します。`external-instance` は、すでに起動しているマシンをそのまま利用する組み込みprovisionerです（詳細は [ジェネレーター]({{< relref "/defining-desired-state/generators" >}})）
- `inventory.nodes`: 論理ノード `hello-node` を定義し、`capabilities` に先ほどのレシピの `provides` 名（`my-org.hello`）を指定します

---

## ステップ4: 計画を作成する

ワークスペースを初期化し、実行計画を作成します。

```bash
nwsctl init
nwsctl plan --recipe-dir ./recipes -t state.yaml
```

`plan` は、Stateとレシピを読み込んで現在の状態との差分を計算し、実行計画を `plan.json` に書き出します。出力に、インスタンスの登録と `greet` タスクの実行が含まれていることを確認してください。

---

## ステップ5: 適用して確認する

計画を実行します。

```bash
nwsctl apply --plan plan.json --recipe-dir ./recipes
```

`greet` タスクがノード上で実行され、スクリプトの出力（`Hello, Niwashi! ...`）がログに表示されます。

実際にノード上にファイルが作られたことを確認してみましょう。`nwsctl ssh` を使うと、Stateに登録した接続情報でそのままノードに入れます。

```bash
nwsctl ssh hello-node
```

```bash
# ノード上で
cat /tmp/hello-niwashi.txt
# => Hello, Niwashi! (node: hello-node)
exit
```

もう一度 `nwsctl plan` を実行すると、今度は差分がないため実行すべきタスクがない計画になります。Niwashiは「現在の状態」と「目標の状態」の差分だけを実行します。

---

## ステップ6: 片付ける

`--destroy` フラグで削除計画を作成し、適用します。削除計画を適用するときは、`apply` にも `--destroy` フラグが必要です。

```bash
nwsctl plan --recipe-dir ./recipes --destroy
nwsctl apply --plan plan.json --recipe-dir ./recipes --destroy
```

`apply` の実行時に、すべてのリソースを削除してよいかの確認プロンプトが表示されるので、承認して続行してください（`--yes` を付けるとプロンプトをスキップできます）。

今度は `operation: destruct` の `cleanup` タスクが実行され、ノード上の `/tmp/hello-niwashi.txt` が削除されます。必要であれば、通常の `ssh` でノードに入ってファイルが消えていることを確認してください。

---

## ここまでで学んだこと

- レシピは `kind`・`metadata`・`spec.provides`・`spec.tasks` で構成される
- Stateの `capabilities` に `provides` の名前を書くと、そのレシピがノードに適用される
- `nwsctl plan` が差分から実行計画を作り、`nwsctl apply` が実行する
- 削除処理は `operation: destruct` のタスクとして書く

## 次のステップ

作りたいレシピの種類に応じて、kind別ガイドに進んでください。

- [Nodeのcapabilityを定義する]({{< relref "node-capability" >}}) — 今回の `kind: node` の詳細（パラメータ・実例）
- [Clusterのcapabilityを定義する]({{< relref "cluster-capability" >}}) — 複数ノードにまたがる処理
- [Infraのプロビジョナーを定義する]({{< relref "infrastructure-provisioning" >}}) — VM・インスタンスの生成
- [Hostの設定を定義する]({{< relref "host-configuration" >}}) — ホスト上のツールの提供
- [アダプターを定義する]({{< relref "defining-adapters" >}}) — 他のレシピから呼び出せるコマンド

仕様の全体像は [レシピのフォーマット仕様]({{< relref "recipe-spec" >}}) と [タスクの定義]({{< relref "defining-tasks" >}}) を参照してください。

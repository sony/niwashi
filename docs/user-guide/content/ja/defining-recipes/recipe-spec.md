---
title: "レシピのフォーマット仕様"
weight: 0
---

# レシピのフォーマット仕様

すべてのレシピは共通のトップレベルフォーマットを持ちます。このページでは `spec` 以下の各フィールドを説明します。フィールドによっては特定の `kind` でのみ使用できます。その場合は **\[kind: xxx のみ\]** のように明示します。

---

## トップレベルフィールド

```yaml
version: nws.recipe/v1
kind: node          # node / infra / host / cluster / adapter
metadata:
  id: my-org/my-recipe
  version: "1.0.0"
  description: "説明"
spec:
  ...
```

| フィールド | 必須 | 説明 |
|-----------|------|------|
| `version` | Yes | スキーマバージョン。`nws.recipe/v1` を指定 |
| `kind` | Yes | レシピの種別。後述 |
| `metadata.id` | Yes | レシピID。命名規則は [レシピの命名規則]({{< relref "recipe-naming" >}}) を参照 |
| `metadata.version` | Yes | レシピのバージョン |
| `metadata.description` | | 説明文 |

---

## kind

レシピの種別を指定します。

| 値 | 説明 |
|----|------|
| `node` | ノード単位のCapabilityを定義する |
| `cluster` | クラスタ単位のCapabilityを定義する |
| `infra` | インフラのプロビジョニングを定義する |
| `host` | nwsctlを実行するホスト上の設定を定義する |
| `adapter` | 他のレシピから呼び出されるアダプターを定義する |

---

## spec.provides

このレシピが提供するCapabilityの別名（alias）を定義します。Stateファイルの `capabilities` で指定される名前に対応します。

> **\[kind: adapter を除くすべての kind\]**

```yaml
spec:
  provides:
    - name: my-org.my-feature        # 単純な別名
    - name: my-org.my-feature
      attrs:
        variant: lite                # 属性付きの別名
```

複数の別名を定義することも可能です。属性（`attrs`）を使うと、同じCapability名で異なる実装を共存させられます。詳細は [レシピの読み込み仕様]({{< relref "recipe-loading" >}}) を参照してください。

---

## spec.requires

このレシピが実行前に必要とする他のレシピへの依存を宣言します。プランナーが依存関係を解決して実行順序を決定します。

> **\[kind: adapter を除くすべての kind\]**

各エントリは文字列またはマップ形式で指定できます。`as:` フィールドを使う場合はマップ形式が必要です。

```yaml
spec:
  requires:
    - host.tool.ansible           # 文字列形式
    - name: some.capability       # マップ形式 — 'as:' を使う場合に必要
      as: upstream                # {{ .Stores.upstream }} アクセス用のエイリアス
```

`as:` フィールドを指定すると、依存先レシピの `store/` データにタスクのテンプレート内から `{{ .Stores.<エイリアス>.<キー> }}` でアクセスできます。詳細は [store/ サブパスの活用]({{< relref "state-changes#store-サブパスの活用" >}}) を参照してください。

依存先として宣言できるレシピの `kind` は、宣言元の `kind` によって異なります。

| 宣言元 | 依存できる kind |
|--------|----------------|
| `host` | `host` |
| `infra` | `host`、`infra` |
| `node` | `host`、`node` |
| `cluster` | `host`、`node`、`cluster` |

`kind: node` / `kind: cluster` から `kind: infra` への依存はplan時エラーになります。nodeがbindしたinstanceのIPアドレス・OS・アーキテクチャ等の情報はNiwashiが別の手段で論理層に伝播するため、infraレシピのデータへの直接アクセスは不要です。

---

## spec.defaults

レシピが受け取れるパラメータのデフォルト値と環境変数を定義します。Stateを記述するユーザーが `params` でパラメータを指定することで、レシピの動作をカスタマイズできます。

> **\[すべての kind\]**

```yaml
spec:
  defaults:
    params:
      version: "1.0.0"
      config_file: "/etc/myapp/config.yaml"
    env:
      MY_ENV_VAR: "default_value"
```

`kind: adapter` の場合、呼び出し元レシピの `spec.defaults` と合成されます。

---

## spec.assets

レシピと一緒に同梱するファイルを指定します。`exec.remote` の場合は接続先ノードに自動的にコピーされます。

> **\[kind: adapter を除くすべての kind\]**

```yaml
spec:
  assets:
    - "scripts/setup.sh"
    - "config/default.conf"
```

パスは `nws-recipe.yaml` が置かれているディレクトリからの相対パスで指定します。

---

## spec.workspace

タスクの作業ディレクトリ（`{{ .Paths.work_dir }}`）の管理方法を指定します。

> **\[kind: adapter を除くすべての kind\]**

```yaml
spec:
  workspace:
    mode: ephemeral    # 実行（run）ごとに新しい作業ディレクトリを作成（デフォルト）
    # mode: persistent  # 作業ディレクトリを保持し続ける（前回の状態を引き継ぐ）
```

| 値 | 説明 |
|----|------|
| `ephemeral`（デフォルト） | 実行（run）ごとに新しい作業ディレクトリが作られる |
| `persistent` | 実行をまたいで同じ作業ディレクトリが使われ、前回のファイルが残る。VMイメージなど実行間で引き継ぎたいファイルを扱うレシピ（Vagrantプロビジョナーなど）で使う |

### 作業ディレクトリの構造

ワークスペース（`--work-dir`、デフォルト `.niwashi`）の中は以下の構造になっています。

```
<work-dir>/                       # ワークスペース（{{ .Paths.workspace }}）
├── state/
│   └── state.json                # 現在のState
├── runs/
│   └── <run-id>/                 # 実行（run）ごとに作られる
│       ├── plan.json             # この実行の計画ファイル
│       └── <phase>-<target>/     # フェーズと実行対象の組ごと（例: node-web-server）
│           └── <recipe-fqid>/    # レシピごと ← ephemeral の work_dir
│               ├── inputs/       # パラメータ・Stateスナップショット（{{ .Paths.input_dir }}）
│               ├── outputs/      # タスク間のデータ共有（{{ .Outputs }} / {{ .Paths.output_dir }}）
│               └── logs/         # レンダリング済みスクリプト・ログ（{{ .Paths.log_dir }}）
└── store/
    └── <phase>-<target>/
        └── <recipe-fqid>/        # ← persistent の work_dir（実行をまたいで保持）
```

- 作業ディレクトリは「**フェーズ＋実行対象**（例: `node-web-server`）× **レシピのFQID**」の単位で作られます。同じレシピでも実行対象のノードが異なれば別のディレクトリになります。実行対象を持たないフェーズ（hostなど）ではフェーズ名のみが使われます
- `mode: ephemeral` の場合、`{{ .Paths.work_dir }}` は `runs/<run-id>/` 配下を指すため、実行のたびに空の状態から始まります
- `mode: persistent` の場合、`{{ .Paths.work_dir }}` は `store/` 配下を指し、実行をまたいで同じディレクトリが使われます
- `inputs/` `outputs/` `logs/` は、`mode` の指定にかかわらず常に `runs/<run-id>/` 配下（実行ごと）に作られます。`persistent` で切り替わるのは `work_dir` のみです

各ディレクトリをタスクから参照する方法は [レシピで使える変数]({{< relref "recipe-variables" >}}) を参照してください。

---

## spec.runtime

ホスト上で管理するリソースを宣言します。`stateChanges` を使う場合に必要です。

> **\[kind: host のみ\]**

```yaml
spec:
  runtime:
    type: tool      # リソース種別（tool / service）
    name: ansible   # リソース名
```

`spec.runtime` を宣言すると、`stateChanges` の書き込み先は `/runtime/<type>/<name>` 以下に自動的に制限されます。詳細は [Hostの設定を定義する]({{< relref "host-configuration" >}}) を参照してください。

---

## spec.allowedScope

このアダプターを呼び出せるレシピの `kind` を制約します。

> **\[kind: adapter のみ\] \[必須\]**

```yaml
spec:
  allowedScope: node   # node / cluster / infra / host / *
```

| 値 | 説明 |
|----|------|
| `node` / `cluster` / `infra` / `host` | 指定したスコープの `kind` からのみ呼び出せる |
| `*` | すべての `kind` から呼び出せる |

`*` は、ツールのインストールや環境構築など、スコープに依存しない処理を提供するアダプターに適しています。スコープによって動作を変える必要がある場合は、スコープごとに別のアダプターを定義してください。

---

## spec.executionUnit

アダプターの実行単位を指定します。

> **\[kind: adapter のみ\]**

```yaml
spec:
  executionUnit: default   # または node
```

| 値 | 説明 |
|----|------|
| `default`（省略時） | 呼び出し元の `kind` がそのまま実行単位 |
| `node` | `allowedScope: cluster` との組み合わせのみ有効。クラスタ内のノードをNiwashi側で分解して実行する |

---

## spec.commands

アダプターが公開するコマンドの定義です。`tool.run` の `command` フィールドに指定した名前に対応します。

> **\[kind: adapter のみ\]**

```yaml
spec:
  commands:
    default:              # command を省略した場合に使われるデフォルト
      task: run
    ansible-playbook:
      task: run-playbook
      description: "Run ansible-playbook"
```

| フィールド | 必須 | 説明 |
|-----------|------|------|
| `task` | Yes | このコマンドを実行したときに呼び出すタスク名 |
| `description` | | コマンドの説明 |

---

## spec.tasks

タスクのリストです。

> **\[すべての kind\]**

タスクの書き方は [タスクの定義]({{< relref "defining-tasks" >}}) を参照してください。`kind: adapter` の場合は `exec.local` アクションのみ使用できます。

---

## 次のステップ

- [タスクの定義]({{< relref "defining-tasks" >}}) — アクション種別・dependsOn・operation の詳細
- [stateChanges で State を更新する]({{< relref "state-changes" >}}) — `stateChanges` の詳細
- [Nodeのcapabilityを定義する]({{< relref "node-capability" >}}) — `kind: node` の書き方
- [Clusterのcapabilityを定義する]({{< relref "cluster-capability" >}}) — `kind: cluster` の書き方
- [Infraのプロビジョナーを定義する]({{< relref "infrastructure-provisioning" >}}) — `kind: infra` の書き方
- [Hostの設定を定義する]({{< relref "host-configuration" >}}) — `kind: host` の書き方
- [アダプターを定義する]({{< relref "defining-adapters" >}}) — `kind: adapter` の書き方

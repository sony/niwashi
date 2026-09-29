---
title: "レシピで使える変数"
weight: 20
---

# レシピで使える変数

レシピのスクリプト（`scriptTpl`）やテンプレートフィールド（`argvTpl`、`envTpl`など）の中では、実行コンテキストの情報を変数として参照できます。変数には2種類あります。

- **テンプレート変数** — `{{ .Xxx }}` 形式。`scriptTpl` や `argvTpl` などのテンプレートフィールドの中で使用します。
- **環境変数** — `$NWS_XXX` 形式。スクリプト実行時に自動的にセットされます。

---

## テンプレート変数

テンプレートフィールド（`scriptTpl`、`argvTpl`、`envTpl`、`workdirTpl`）の中で使用できます。

| 変数 | 説明 | 例 |
|------|------|-----|
| `{{ .Target }}` | 実行対象のノード名またはクラスター名 | `web-server-01` |
| `{{ .Params.xxx }}` | レシピのパラメータ。`spec.defaults.params` のデフォルト値を、Stateで指定した値で上書きしたもの。キーの命名規則は [paramsキーの命名]({{< relref "/defining-desired-state/basic-concepts/naming-rules#params-キーの命名" >}}) を参照 | `{{ .Params.version }}` |
| `{{ .Assets }}` | レシピのアセットディレクトリ。`exec.remote` ではノード側のパスに展開される | `/home/user/.niwashi/recipe/...` |
| `{{ .Paths.work_dir }}` | タスクの作業ディレクトリ。`exec.remote` ではノード側のパス | `/home/user/.niwashi/...` |
| `{{ .Paths.input_dir }}` | タスクのインプットディレクトリ | `/home/user/.niwashi/.../inputs/` |
| `{{ .Paths.output_dir }}` | タスクのアウトプットディレクトリ | `/home/user/.niwashi/.../outputs/` |
| `{{ .Paths.log_dir }}` | タスクのログディレクトリ | `/home/user/.niwashi/.../logs/` |
| `{{ .Paths.workspace }}` | ワークスペースのルートディレクトリ | `/home/user/.niwashi/` |
| `{{ .Outputs }}` | タスク間共通のデータ置き場（後述） | `/path/to/outputs/` |
| `{{ .Store.xxx }}` | 自レシピの `stateChanges` によって `store/<key>` に書き込まれた値。キーが存在しない場合は空文字列を返す。詳細は [store/ サブパスの活用]({{< relref "state-changes#store-サブパスの活用" >}}) を参照 | `{{ .Store.applied_version }}` |
| `{{ .Stores.<エイリアス>.xxx }}` | `requires.as` で定義したエイリアスを通じて、依存先レシピの `store/` サブパスに保存された値にアクセスする。詳細は [spec.requires]({{< relref "recipe-spec#specrequires" >}}) を参照 | `{{ .Stores.upstream.endpoint }}` |

### {{ .Assets }}

`{{ .Assets }}` は、レシピの `assets` に列挙したファイルが配置されるディレクトリです。`exec.remote` ではアセットはノード側にコピーされるため、ノード側のパスに展開されます。

```yaml
spec:
  assets:
    - bin/setup.sh
  tasks:
    - name: setup
      action:
        exec.remote:
          scriptTpl: |
            {{ .Assets }}/bin/setup.sh
```

`kind: adapter` のレシピでは、`{{ .Assets }}` はアダプター自身のアセットディレクトリを指します。

### {{ .Outputs }}

`.Outputs` はレシピ内のタスク間でデータを共有するための共通ディレクトリです。タスクが生成したファイルをここに置くと、後続のタスクから参照できます。

```yaml
tasks:
  - name: generate
    action:
      exec.local:
        scriptTpl: |
          echo "value=123" > {{ .Outputs }}/result.env

  - name: consume
    dependsOn: [generate]
    action:
      exec.local:
        scriptTpl: |
          source {{ .Outputs }}/result.env
          echo "Got: $value"
```

ドット区切りでサブパスを指定できます。`{{ .Outputs.foo.bar }}` は `{{ .Outputs }}/foo/bar` に展開されます。

`exec.remote` の場合、タスク実行後にノード側の `{{ .Outputs }}` の内容がホスト側の `{{ .Outputs }}` に自動的にコピーされます。

---

## 環境変数

スクリプト実行時にnwsctlが自動的にセットする環境変数です。`scriptTpl` のスクリプト本体の中で `$NWS_XXX` の形式で参照できます。

### 実行制御

| 変数 | 説明 |
|------|------|
| `NWS_DRY_RUN` | ドライランモード（`off` / `simulate`） |
| `NWS_DRY_RUN_ENABLED` | ドライラン有効時は `1`、無効時は `0` |
| `NWS_LOG_LEVEL` | ログレベル |

### 実行コンテキスト

| 変数 | 説明 |
|------|------|
| `NWS_SCOPE` | 実行フェーズ（`node` / `cluster` / `host` など） |
| `NWS_TARGET_ID` | 実行対象のID。通常はフェーズに対応するID（nodeフェーズならノードID、clusterフェーズならクラスターID、infraフェーズならジェネレータID）を表す。ただし、clusterフェーズでも `where` 条件によるノード分解やアダプターの `executionUnit: node` 指定の場合はノードIDになる |

### パスとファイル

| 変数 | 説明 |
|------|------|
| `NWS_WORK_DIR` | タスクの作業ディレクトリ（`{{ .Paths.work_dir }}` と同じ値） |
| `NWS_LOG_DIR` | タスクのログディレクトリ（`{{ .Paths.log_dir }}` と同じ値） |
| `NWS_INPUT_DIR` | タスクのインプットディレクトリ（`{{ .Paths.input_dir }}` と同じ値） |
| `NWS_OUTPUT_DIR` | タスクのアウトプットディレクトリ（`{{ .Paths.output_dir }}` と同じ値） |
| `NWS_PARAMS` | パラメータファイル（JSON）のパス |
| `NWS_STATE` | Stateスナップショットファイル（JSON）のパス |

各ディレクトリがワークスペースのどこに作られるかは [spec.workspace（作業ディレクトリの構造）]({{< relref "recipe-spec#specworkspace" >}}) を参照してください。

### NWS_PARAMS

`NWS_PARAMS` はタスクに渡されたパラメータをJSON形式で記録したファイルです。スクリプト内から `jq` などで読み込んで利用できます。

```bash
VERSION=$(jq -r '.version' "${NWS_PARAMS}")
echo "Installing version ${VERSION}"
```

### NWS_STATE

`NWS_STATE` は、実行対象に関するStateのスナップショットをJSON形式で記録したファイルのパスです。レシピはこのファイルを通じて、実行対象のノードの接続情報やOS・ラベルなどを参照できます。

ファイルの構造は以下の通りです。

```json
{
  "host": {
    "os": "linux",
    "arch": "amd64"
  },
  "inventory": {
    "nodes": {
      "<ノード名>": {
        "labels": { "role": "master" },
        "os": "linux",
        "arch": "amd64",
        "addresses": ["192.168.1.10"],
        "connection": { ... }
      }
    },
    "clusters": {
      "<クラスタ名>": { ... }
    }
  },
  "runtime": { ... }
}
```

| フィールド | 説明 |
|-----------|------|
| `host` | nwsctlを実行しているホストの `os` / `arch` |
| `inventory.nodes` | 実行対象のノードの情報（ラベル、OS、アーキテクチャ、アドレス、接続情報） |
| `inventory.clusters` | 実行対象のクラスタの情報 |
| `runtime` | ホスト上のツール・サービスの情報（`kind: host` のレシピが記録したもの） |

**スナップショットに含まれる範囲は実行対象によって異なります。**

| レシピの実行フェーズ | `inventory` に含まれる内容 |
|--------------------|--------------------------|
| node | 実行対象のノードのみ |
| cluster | 実行対象のクラスタと、そのメンバーノードすべて（クラスタ固有ラベルをマージ済み） |
| host / infra | 空（`host` と `runtime` のみ利用可能） |

`kind: cluster` のレシピでは、メンバーノードの一覧や接続先アドレスを `NWS_STATE` から取得して、インベントリファイルの生成などに利用できます。

```bash
# クラスタのメンバーノードのアドレス一覧を取得する例
jq -r '.inventory.nodes[].addresses[0]' "${NWS_STATE}"
```

**`NWS_PARAMS` との使い分け**: `NWS_PARAMS` は「Stateの `params` でこのレシピに渡された値」、`NWS_STATE` は「実行対象の周辺情報のスナップショット」です。レシピの動作設定は `params`（または `{{ .Params.* }}`）で受け取り、ノードの接続情報やラベルなど環境由来の情報は `NWS_STATE` から読み取ってください。

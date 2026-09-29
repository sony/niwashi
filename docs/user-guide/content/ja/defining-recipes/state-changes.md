---
title: "stateChanges で State を更新する"
weight: 5
---

# stateChanges で State を更新する

`stateChanges` は、タスクの実行後にNiwashiのStateを更新する仕組みです。主に以下のレシピで使用します。

- `kind: infra` — 生成したインスタンスの接続情報を記録する
- `kind: host` — ホスト上のツール情報を記録する
- `kind: node` / `kind: cluster` — Capabilityの状態管理

---

## 基本的な書き方

`stateChanges` はタスクの直下にキーマップとして記述します。キーは操作の名前（任意）で、複数の操作を定義できます。

```yaml
tasks:
  - name: provision
    action:
      exec.local:
        scriptTpl: |
          #!/bin/bash
          # ... インスタンスを起動して接続情報をJSONで書き出す ...
    stateChanges:
      register-instance:          # 操作名（任意）
        op: set
        path: "instances/my-vm-0"
        valueFromFile: "{{ .Outputs }}/my-vm-0.json"
```

---

## フィールドリファレンス

| フィールド | 必須 | 説明 |
|-----------|------|------|
| `op` | Yes | 操作種別。`set`（値を置き換える）または `remove`（削除） |
| `path` | Yes | 書き込み先のパス（後述）。テンプレート変数使用可 |
| `value` | | 書き込む値（YAML形式でインラインに記述） |
| `valueFromFile` | | 書き込む値をファイルから読み込む。`.json`/`.yaml` は構造化データとして読み込まれる |
| `valueFromJson` | | 書き込む値をJSON文字列として記述。テンプレート変数使用可 |
| `count` | | ループ回数（後述）。`{{ .Loop.index }}` と組み合わせて使用 |

`value`、`valueFromFile`、`valueFromJson` は排他的に使用します。いずれも指定しない場合は `null` が書き込まれます。

`op: set` は指定したパスの値を丸ごと置き換えます。既存の値があっても上書きされます。

### valueFromFile の読み込み形式

`valueFromFile` で指定したファイルの拡張子に応じて読み込み形式が変わります。

| 拡張子 | 読み込み形式 |
|--------|------------|
| `.json`, `.jsonc` | JSON として構造化データを読み込む |
| `.yaml`, `.yml` | YAML として構造化データを読み込む |
| その他 | テキストとして読み込む |

---

## kind 別の書き方

`path` の基点はレシピの `kind` によって決まります。`path: "/"` または `path: ""` は基点そのものを指します。

### `kind: infra`

プロビジョナーレシピは、生成したインスタンスの接続情報を `instances/<id>` に記録します。

```yaml
stateChanges:
  register-instance:
    op: set
    path: "instances/my-vm-0"
    valueFromFile: "{{ .Outputs }}/my-vm-0.json"
```

`instances/<id>` に書き込む値の構造は [Infraのプロビジョナーを定義する]({{< relref "infrastructure-provisioning" >}}) を参照してください。

複数インスタンスを一括登録する場合は `count` を使います（後述）。

`store/` サブパスは、nodeレシピへの情報引き継ぎなどに使えます（[store/ サブパスの活用](#store-サブパスの活用) 参照）。

### `kind: host`

installerレシピは、ホスト上のツール情報を記録します。`path` の基点は `spec.runtime` で宣言した `type` と `name` によって決まります。`path: "/"` はその基点そのものを指すため、ツール情報全体をまとめて書き込む場合に使います。

```yaml
# spec.runtime: { type: tool, name: ansible } を宣言した場合
# → path: "/" の基点が /runtime/tool/ansible になる
stateChanges:
  record-ansible:
    op: set
    path: "/"
    valueFromFile: "{{ .Outputs }}/ansible_info.json"
```

記録された値は、アダプターレシピのテンプレート変数 `{{ .Runtime.tool.<name>.<key> }}` として参照できます。詳細は [Hostの設定を定義する]({{< relref "host-configuration" >}}) を参照してください。

### `kind: node` / `kind: cluster`

Capabilityのメイン状態はNiwashiが自動管理するため、`stateChanges` では `store/` サブパスのみ書き込めます（[store/ サブパスの活用](#store-サブパスの活用) 参照）。

```yaml
stateChanges:
  save-endpoint:
    op: set
    path: "store/endpoint"
    valueFromFile: "{{ .Outputs }}/endpoint.json"
```

---

## 削除（op: remove）

`operation: destruct` のタスクに `stateChanges` を組み合わせると、インフラを削除したときにStateからも対応するエントリを削除できます。

```yaml
tasks:
  - name: teardown
    operation: destruct
    action:
      exec.local:
        scriptTpl: |
          #!/bin/bash
          vagrant destroy -f
    stateChanges:
      remove-from-state:
        op: remove
        path: "instances"
```

---

## ループ（count）

`count` を指定すると、同じ操作を指定回数繰り返します。`{{ .Loop.index }}` が 0 から始まるインデックスに展開されます。複数のインスタンスをまとめて登録するinfraプロビジョナーで使います。

```yaml
stateChanges:
  generate-instances:
    count: "{{ .Params.count }}"    # テンプレート変数で指定可能
    op: set
    path: "instances/{{ .Params.prefix }}-{{ .Loop.index }}"
    valueFromFile: "{{ .Outputs.instances }}/{{ .Params.prefix }}-{{ .Loop.index }}.json"
```

---

## store/ サブパスの活用

各 `kind` の `store/` サブパスは、レシピが自由に使えるデータ保存領域です。

`store/` はCapabilityの更新時には保持され、constructタスクが失敗した場合は削除されます。[更新処理（operation: update）]({{< relref "defining-tasks#更新処理operation-update" >}}) と [タスクが失敗した場合]({{< relref "defining-tasks#タスクが失敗した場合" >}}) を参照してください。

### 用途1: 冪等性管理

「どのバージョンで適用済みか」などのメタ情報を保存し、再実行時のスキップ判定に使います。`kind: adapter` を除くすべての kind で利用できます。

`{{ .Store.<key> }}` でタスクの scriptTpl から自レシピの `store/` に保存した値を参照できます。キーが存在しない場合（初回実行時など）は空文字列が返るため、`default` 関数と組み合わせてシェルの条件判定に使えます。

```yaml
tasks:
  - name: apply
    action:
      exec.local:
        scriptTpl: |
          #!/bin/bash
          set -euo pipefail
          APPLIED="{{ .Store.applied_version | default "" }}"
          if [ "$APPLIED" = "{{ .Params.version }}" ]; then
            echo "already applied (version={{ .Params.version }}), skipping"
            exit 0
          fi
          # ... 実際の処理 ...
    stateChanges:
      save-applied-version:
        op: set
        path: "store/applied_version"
        value: "{{ .Params.version }}"
```

> **注意**: `store/` に保存するキー名はアンダースコアを使ってください。ハイフンを含むキー（例: `applied-version`）はテンプレートのドット記法（`{{ .Store.applied-version }}`）では参照できません。

### 用途2: 後段レシピへのデータ受け渡し

あるレシピが収集・生成した情報を `store/` に保存し、後段のレシピがタスクのテンプレート内で `{{ .Stores.<エイリアス>.<キー> }}` として直接参照できます。

参照側のレシピは `requires` エントリに `as:` フィールドを追加してエイリアスを設定します。Niwashiは依存先レシピの `store/` データをそのエイリアス名で利用可能にします。

```yaml
# 保存側レシピ
stateChanges:
  save-endpoint:
    op: set
    path: "store/endpoint"
    valueFromFile: "{{ .Outputs }}/endpoint.json"
```

```yaml
# 参照側レシピ — {{ .Stores.upstream.<key> }} でアクセス
spec:
  requires:
    - name: some.capability
      as: upstream            # {{ .Stores.upstream }} で使うエイリアス
```

```yaml
tasks:
  - name: use-endpoint
    action:
      exec.local:
        scriptTpl: |
          ENDPOINT="{{ .Stores.upstream.endpoint }}"
          echo "Connecting to ${ENDPOINT}"
```

実行順序は保証されます（保存側が先に実行されます）。実行フェーズをまたいだ受け渡し（例: `infra` レシピ → `node` レシピ）も同様に動作します。

### `kind: cluster` での注意点

`kind: cluster` のレシピで `where` 条件によりノード単位でタスクが実行される場合も、`store/` の書き込み先はクラスタのcapabilityパスに集約されます。複数ノードが同じパスに書き込むと後から書いた値で上書きされるため、ノードを識別するサブキーを含めてください。

```yaml
stateChanges:
  save-node-info:
    op: set
    path: "store/nodes/{{ .Target }}"    # ノードIDをキーにして衝突を避ける
    valueFromFile: "{{ .Outputs }}/info.json"
```

---

## 実例：Ansibleアダプターのパターン

`recipe/ansible/` のレシピが installer + adapter + stateChanges の典型例です。

1. **ansible-installer.yaml**（`kind: host`）— Ansibleを検出し、パスを `stateChanges` で記録する
2. **ansible-adapter.yaml**（`kind: adapter`）— `{{ .Runtime.tool.ansible.playbook.path }}` を参照してコマンドを実行する

詳細は [アダプターを定義する]({{< relref "defining-adapters" >}}) を参照してください。

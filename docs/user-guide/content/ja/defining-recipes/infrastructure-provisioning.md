---
title: "Infraのプロビジョナーを定義する"
weight: 2
---

# Infraのプロビジョナーを定義する

`kind: infra` のレシピ（プロビジョナー）の書き方を説明します。プロビジョナーはVMやインスタンスを生成・破棄し、その接続情報をStateに記録する役割を担います。

---

## `kind: infra` のレシピとは

`kind: infra` のレシピは、**インフラ（VM・インスタンス）を生成する**レシピです。Stateの `infrastructure.generators` で参照された `provisioner` として動作します。

{{< mermaid >}}
flowchart LR
    subgraph host["ホスト（nwsctl）"]
        recipe["kind: infra のレシピ<br/>（exec.local のみ）"]
        state["State"]
    end
    subgraph infra["インフラ"]
        i1["instance-1"]
        i2["instance-2"]
    end
    recipe -->|"生成・破棄"| infra
    recipe -->|"接続情報を記録（stateChanges）"| state
{{< /mermaid >}}

プロビジョナーの主な役割：

1. VMやインスタンスを起動する
2. 生成したインスタンスの接続情報（IPアドレス、SSH秘密鍵など）を `stateChanges` でStateに記録する

Stateに記録された接続情報を元に、Niwashiはノードへの `exec.remote` 接続を自動的に確立します。

> **注意**: プロビジョナーは `exec.local` のみを使用します。ノードにはまだ接続できない段階のため `exec.remote` は使えません。

---

## フィールドリファレンス

### 共通フィールド

`spec` 以下の共通フィールドの詳細は [レシピのフォーマット仕様]({{< relref "recipe-spec" >}}) を参照してください。`kind: infra` では `spec.requires` に `host` と `infra` の依存を宣言できます。

タスクの書き方は [タスクの定義]({{< relref "defining-tasks" >}})、State の更新は [stateChanges で State を更新する]({{< relref "state-changes" >}}) を参照してください。

`kind: infra` 固有のフィールドの注意点：

| フィールド | 説明 |
|-----------|------|
| `spec.provides` | ジェネレーターの `provisioner` で指定される名前（例: `infra.vm`） |
| `spec.workspace.mode` | VMファイルやワークスペースを保持するため `persistent` を推奨 |

### テンプレート変数

`kind: infra` のタスクでは、`{{ .Target }}` は**ジェネレーター名**（Stateの `infrastructure.generators` のキー名）に展開されます。

```yaml
infrastructure:
  generators:
    my-vms:             # ← これが {{ .Target }}
      provisioner: infra.vm.driver=vagrant
```

| 変数 | 説明 |
|------|------|
| `{{ .Target }}` | ジェネレーター名 |
| `{{ .Params.xxx }}` | ジェネレーターの `params` で指定したパラメータ |
| `{{ .Loop.index }}` | ループインデックス（`count` と組み合わせて使用） |

その他のテンプレート変数と実行時環境変数の一覧は [レシピで使える変数]({{< relref "recipe-variables" >}}) を参照してください。

### インスタンス情報のスキーマ

`stateChanges` でStateに書き込むインスタンス情報のJSON構造は以下のとおりです。

```json
{
  "connection": {
    "ssh": {
      "address": {
        "host": "192.168.56.10",
        "port": 22,
        "user": "vagrant"
      },
      "auth": {
        "method": "privateKey",
        "privateKeyPath": "/path/to/private_key"
      },
      "hostKey": {
        "knownHostsPath": "/path/to/known_hosts"
      }
    }
  },
  "addresses": ["192.168.56.10"]
}
```

このJSONをインスタンスごとにファイルに書き出し、`stateChanges` で `instances/<instance-name>` に登録します。

### stateChanges の書き方

インスタンスの登録・削除には以下のパターンを使います。

**単一インスタンスの登録**

```yaml
stateChanges:
  register-instance:
    op: set
    path: "instances/my-instance"
    valueFromFile: "{{ .Outputs }}/instance.json"
```

**複数インスタンスの一括登録（`count` + `{{ .Loop.index }}`）**

```yaml
stateChanges:
  register-instances:
    count: "{{ .Params.count }}"
    op: set
    path: "instances/{{ .Params.prefix }}-{{ .Loop.index }}"
    valueFromFile: "{{ .Outputs.instances }}/{{ .Params.prefix }}-{{ .Loop.index }}.json"
```

**削除時（`operation: destruct`）**

```yaml
stateChanges:
  remove-from-state:
    op: remove
    path: "instances"
```

### プロビジョニングが失敗した場合

constructタスクが失敗すると、Generatorのエントリは、`stateChanges` で登録済みのインスタンスも含めてStateから削除されます。次回の `nwsctl plan` / `nwsctl apply` では、プロビジョニングが再度実行されます。途中まで作成されたリソースから始まっても収束するようにconstructタスクを書いてください（例: 失敗させずに、既に存在するVMを再利用する）。

再実行しない場合は、途中まで作成されたリソースを手動で削除してください。これらのリソースはStateに記録されていないため、`nwsctl plan --destroy` の対象になりません。

destructタスクが失敗した場合、GeneratorのエントリはStateに残ります。次回の `--destroy` / `--prune` で再度destructされます。

---

## 最小構成のサンプル

```yaml
version: nws.recipe/v1
kind: infra
metadata:
  id: my-org/my-provisioner
  version: "1.0.0"
  description: "Provision VMs"
spec:
  provides:
    - name: infra.my-provisioner

  workspace:
    mode: persistent

  defaults:
    params:
      count: 1
      prefix: "vm"

  tasks:
    - name: provision
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            mkdir -p "{{ .Outputs.instances }}"

            for i in $(seq 0 $(({{ .Params.count }} - 1))); do
              VM_NAME="{{ .Params.prefix }}-${i}"
              # ... VMを起動して接続情報を取得 ...
              jq -n \
                --arg host "192.168.56.$((10 + i))" \
                --arg user "ubuntu" \
                --arg key "/path/to/key" \
                '{
                  connection: {
                    ssh: {
                      address: { host: $host, port: 22, user: $user },
                      auth: { method: "privateKey", privateKeyPath: $key }
                    }
                  },
                  addresses: [$host]
                }' > "{{ .Outputs.instances }}/${VM_NAME}.json"
            done
      stateChanges:
        register:
          count: "{{ .Params.count }}"
          op: set
          path: "instances/{{ .Params.prefix }}-{{ .Loop.index }}"
          valueFromFile: "{{ .Outputs.instances }}/{{ .Params.prefix }}-{{ .Loop.index }}.json"

    - name: teardown
      operation: destruct
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            # ... VMを破棄 ...
      stateChanges:
        remove:
          op: remove
          path: "instances"
```

---

## 実例：Vagrant プロビジョナー

`recipe/vagrant/nws-recipe.yaml` が実際のプロビジョナー実装例です。

```yaml
version: nws.recipe/v1
kind: infra
metadata:
  id: nws/vagrant.vm-cluster
  version: "1.0.0"
spec:
  provides:
    - name: "infra.vm"
      attrs:
        driver: "vagrant"

  workspace:
    mode: persistent

  defaults:
    params:
      count: 1
      prefix: "nws-vm"
      box: "bento/ubuntu-22.04"
      memory: "1024"
      cpus: "1"

  assets:
    - "bin/create-config.sh"      # vagrant ssh-config を解析してJSONを生成
    - "bin/cleanup-vagrant.sh"

  tasks:
    - name: prepare-vagrantfile
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            if [ -f Vagrantfile ]; then exit 0; fi
            # ... Vagrantfile を生成 ...

    - name: provision-and-extract
      dependsOn: [prepare-vagrantfile]
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            mkdir -p "{{ .Outputs.instances }}"
            vagrant up
            vagrant ssh-config > "{{ .Outputs.all_ssh_config }}"

            for i in $(seq 0 $(({{ .Params.count }} - 1))); do
              VM_NAME="{{ .Params.prefix }}-${i}"
              {{ .Assets }}/bin/create-config.sh \
                "{{ .Outputs.all_ssh_config }}" \
                "$VM_NAME" \
                "{{ .Outputs.instances }}/${VM_NAME}.json"
            done
      stateChanges:
        generate-instances:
          count: "{{ .Params.count }}"
          op: set
          path: "instances/{{ .Params.prefix }}-{{ .Loop.index }}"
          valueFromFile: "{{ .Outputs.instances }}/{{ .Params.prefix }}-{{ .Loop.index }}.json"

    - name: teardown-instance
      operation: destruct
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            vagrant destroy -f
      stateChanges:
        remove-instance-from-state:
          op: remove
          path: "instances"
```

---

## 次のステップ

- [stateChanges で State を更新する]({{< relref "state-changes" >}}) — `stateChanges` の詳細
- [タスクの定義]({{< relref "defining-tasks" >}}) — アクション種別・dependsOn・operation の詳細

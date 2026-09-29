---
title: "Clusterのcapabilityを定義する"
weight: 6
---

# Clusterのcapabilityを定義する

クラスタに適用するCapabilityのレシピ（`kind: cluster`）の書き方を説明します。

---

## `kind: cluster` のレシピとは

`kind: cluster` のレシピは、**クラスタ全体に対して処理を実行する**レシピです。ユーザーがStateファイルのClusterにCapabilityを指定すると、Niwashiは該当するレシピを見つけ、対象クラスタに対してレシピのタスクを実行します。

{{< mermaid >}}
flowchart LR
    subgraph host["ホスト（nwsctl）"]
        recipe["kind: cluster のレシピ"]
    end
    subgraph cluster["クラスタ（実行対象は1つ）"]
        cp["cp"]
        w1["worker-1"]
        w2["worker-2"]
    end
    recipe -->|"クラスタ全体として実行"| cluster
    recipe -.->|"where / executionUnit: node で<br/>ノード単位に分解"| cp
{{< /mermaid >}}

クラスタレシピには2つの実装パターンがあります。

| パターン | 使い方 | 向いているケース |
|---------|--------|----------------|
| **Adapterパターン** | `tool.run` でアダプター（Ansible等）に処理を委譲する | クラスタ全体に対して1つのツールで処理する場合（Kubespray等） |
| **`where` パターン** | `where` でノード単位に分解して実行する | niwashi単独でノード役割ごとに異なる処理を記述する場合 |

---

## フィールドリファレンス

### 共通フィールド

`spec` 以下の共通フィールドの詳細は [レシピのフォーマット仕様]({{< relref "recipe-spec" >}}) を参照してください。`kind: cluster` では `spec.requires` に `host`、`node`、`cluster` の依存を宣言できます。

タスクの書き方は [タスクの定義]({{< relref "defining-tasks" >}})、State の更新は [stateChanges で State を更新する]({{< relref "state-changes" >}}) を参照してください。

### テンプレート変数

`kind: cluster` のタスクでは、`{{ .Target }}` は**実行対象のクラスタID**に展開されます。`where` によってノード単位に分解されて実行される場合も、`{{ .Target }}` はクラスタIDのままです。ノード単位実行時に実行対象のノードIDを知りたい場合は、スクリプト内で環境変数 `$NWS_TARGET_ID` を参照してください。

その他のテンプレート変数と実行時環境変数の一覧は [レシピで使える変数]({{< relref "recipe-variables" >}}) を参照してください。

### `where` の仕様

`where` を使うとNiwashiがクラスタ内のノードを分解して各ノードにタスクを実行します。

#### Adapterを使わない場合

Adapterを使わない場合、分解されるかどうかは `where` の有無だけで決まります。

| `where` の有無 | 動作 |
|--------------|------|
| なし | クラスタ全体を1つのユニットとして1回だけ実行する（ノード単位への分解はしない） |
| あり | 条件にマッチしたノードのみに分解して実行する |

`where` を省略したままノード単位の分解を得る方法はありません——`executionUnit` は
アダプター側のレシピが持つフィールドであり、素の `where` パターンのタスクが指定できる
ものではないためです。アダプターを使わずに**すべてのノード**にタスクを実行したい
場合は、`where` を省略するのではなく `where: "true"` と明示してください。

```yaml
- name: prepare
  where: "true"
  action:
    exec.remote:
      scriptTpl: |
        #!/bin/bash
        apt-get install -y curl
```

これは `exec.remote` がノード単位の実行対象を必須とすることとも関係します。`where`
がないタスクはクラスタ全体を1つのユニットとして実行するため接続先ノードがなく、
`exec.remote` は失敗します。`where: "true"` は全ノードに分解するため、各ノードに
ノード単位の実行コンテキストを与えます。

#### Adapterを使う場合（`executionUnit`）

タスクがアダプターを使う場合、動作はアダプター自身の `executionUnit` にも依存します。

| `where` の有無 | アダプターの`executionUnit` | 動作 |
|--------------|------------------------------|------|
| なし | `default` | クラスタ全体を1つのユニットとして1回だけ実行する（分配はアダプターが管理する） |
| なし | `node` | クラスタ内の全ノードに分解して実行する |
| あり | `node` | 条件にマッチしたノードのみに分解して実行する |
| あり | `default` | **エラー**（アダプターがノード単位の実行をサポートしていない） |

`where` でノードを絞り込むには、Stateファイルの `nodes` にラベルを定義します。

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        cp:
          labels:
            role: master
        worker1:
          labels:
            role: worker
```

クラスタノードのラベルはそのクラスタのレシピ実行時のみ有効です。同じノードが別のクラスタに属している場合でも、ラベルは独立して評価されます。

`where` フィールドの詳細（変数一覧・CEL式の書き方）は [タスクのフィルタリング条件 (where)]({{< relref "task-where" >}}) を参照してください。

---

## 最小構成のサンプル

```yaml
version: nws.recipe/v1
kind: cluster
metadata:
  id: my-org/my-cluster-feature
  version: "1.0.0"
  description: "My cluster feature recipe"
spec:
  provides:
    - name: my-org.my-cluster-feature
  tasks:
    - name: setup
      where: "true"
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            echo "Setting up node $NWS_TARGET_ID in cluster {{ .Target }}"
```

---

## パターン1: Adapterを使ったクラスタ構築

クラスタ全体を対象としたツール（Ansible等）を使って処理を実行するパターンです。`tool.run` でアダプターにすべての処理を委譲します。タスクの実行対象は「クラスタ全体（1つ）」として扱われ、ノードへの分配はアダプター側が管理します（AnsibleであればInventoryファイルで制御）。

```yaml
spec:
  requires:
    - adapter.tool.ansible

  workspace:
    mode: persistent

  tasks:
    - name: gen-inventory
      action:
        tool.run:
          toolRef: adapter.tool.ansible
          command: gen-inventory
          workdirTpl: "{{ .Paths.work_dir }}"

    - name: run-playbook
      dependsOn: [gen-inventory]
      action:
        tool.run:
          toolRef: adapter.tool.ansible
          command: ansible-playbook
          workdirTpl: "{{ .Paths.work_dir }}"
          argvTpl:
            - -i
            - hosts.ini
            - playbook.yml
```

詳細は [Ansibleアダプター]({{< relref "../recipes/ansible" >}}) を参照してください。

---

## パターン2: `where` を使ったクラスタ構築

ノードの役割（ラベル）や属性に応じてタスクを振り分けるパターンです。アダプターを使わず、niwashi単独でクラスタ構成を記述できます。

```yaml
spec:
  tasks:
    # 全ノードに共通の処理
    - name: prepare
      where: "true"
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            apt-get install -y curl

    # master ロールのノードのみ
    - name: init-master
      where: "node.labels.role == 'master'"
      dependsOn: [prepare]
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            kubeadm init

    # worker ロールのノードのみ
    - name: join-workers
      where: "node.labels.role == 'worker'"
      dependsOn: [init-master]
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            bash "{{ .Outputs }}/join_command.sh"
```

CEL式を使った複合条件も記述できます。

```yaml
# Linux かつ master ロールのノードのみ
- name: init-linux-master
  where: "node.os == 'linux' && node.labels.role == 'master'"
```

---

## 実例：master/workerで構成するKubernetesクラスタ

`where` パターンを使って、control planeとworkerノードの役割を分けてKubernetesクラスタを構成する例です。

```yaml
version: nws.recipe/v1
kind: cluster
metadata:
  id: my-org/my-k8s
  version: "1.0.0"
  description: "Kubernetes cluster setup"
spec:
  provides:
    - name: my-org.my-k8s

  tasks:
    # 全ノードに共通の前処理
    - name: prepare
      where: "true"
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            sudo apt-get update
            sudo apt-get install -y curl

    # master ロールのノードのみ
    - name: init-master
      where: "node.labels.role == 'master'"
      dependsOn: [prepare]
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            kubeadm init --pod-network-cidr=10.244.0.0/16
            mkdir -p $HOME/.kube
            sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config
            kubeadm token create --print-join-command \
              > "{{ .Outputs }}/join_command.sh"
      stateChanges:
        save-join-command:
          op: set
          path: "store/join_command"
          valueFromFile: "{{ .Outputs }}/join_command.sh"

    # worker ロールのノードのみ
    - name: join-workers
      where: "node.labels.role == 'worker'"
      dependsOn: [init-master]
      action:
        exec.remote:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail
            bash "{{ .Outputs }}/join_command.sh"

    # クラスタレベルの後処理
    - name: verify
      dependsOn: [init-master, join-workers]
      action:
        exec.local:
          scriptTpl: |
            echo "Cluster {{ .Target }} setup completed"
```

---

## 次のステップ

- [タスクのフィルタリング条件 (where)]({{< relref "task-where" >}}) — `where` フィールドの詳細
- [タスクの定義]({{< relref "defining-tasks" >}}) — アクション種別・dependsOn・operation の詳細
- [アダプターを利用する]({{< relref "using-adapters" >}}) — `tool.run` でアダプターを呼び出す詳細
- [Ansibleアダプター]({{< relref "../recipes/ansible" >}}) — Ansible用のリファレンスアダプターの使い方

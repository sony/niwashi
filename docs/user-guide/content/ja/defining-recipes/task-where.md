---
title: "タスクのフィルタリング条件 (where)"
weight: 8
---

# タスクのフィルタリング条件 (`where`)

タスクに省略可能な `where` フィールドを指定すると、条件に一致した実行対象に対してのみタスクを実行します。値は [CEL (Common Expression Language)](https://github.com/google/cel-spec) の文字列で、`true` に評価された場合のみ実行されます。省略時は対象に無条件で適用されます。

---

## `kind` ごとの動作の違い

`where` の評価コンテキストは `kind` によって異なります。

| `kind` | 実行対象 | `where` の評価コンテキスト |
|--------|----------|--------------------------|
| `node` | 指定されたノード（1つ） | 対象ノードの `os`, `arch`, `labels` |
| `infra` | 指定された infra ターゲット（1つ） | nwsctl を実行するホストの `os`, `arch` |
| `host` | nwsctl を実行するホスト（1つ固定） | ホスト自身の `os`, `arch` |

`infra` と `host` の `where` は、実行対象を絞り込むというより「このホスト環境でこのタスクを実行するか」の実行可否判定として機能します。

---

## 評価コンテキスト変数

| 変数 | 型 | 利用可能な `kind` | 説明 |
|------|----|-------------------|------|
| `node.os` | string | `node`, `infra`, `host` | OSの識別子 |
| `node.arch` | string | `node`, `infra`, `host` | CPUアーキテクチャ |
| `node.labels` | map(string, string) | `node`, `cluster` | ユーザー定義ラベル |

### node.os / node.arch の値

`node.os` と `node.arch` は、`kind: infra` のレシピが `stateChanges` で書き込むか、nwsctlが対象ノードをprobeして実際に検出した値が設定されます。Stateの `labels` に `os` や `arch` を記述しても、この変数には反映されません。

文字列の表記は [Go - Optional environment variables](https://golang.org/doc/install/source#environment) に従います（例: `linux`, `windows`, `darwin`, `amd64`, `arm64`）。

### node.labels の値

`node.labels` は以下のラベルをマージした値です。

- `inventory.nodes.<node-name>.labels` で定義したグローバルラベル
- `kind: cluster` のレシピ実行時は、さらに `inventory.clusters.<cluster-name>.nodes.<node-name>.labels` で定義したクラスタ固有ラベルがマージされます（クラスタ固有ラベルが優先）

```yaml
inventory:
  nodes:
    cp:
      labels:
        tier: control      # グローバルラベル → node.labels.tier = "control"

  clusters:
    k8s-cluster:
      nodes:
        cp:
          labels:
            role: master   # クラスタ固有ラベル → node.labels.role = "master"
```

`k8s-cluster` のレシピ実行時、`node.labels` は `{tier: control, role: master}` になります。クラスタ固有ラベルはそのクラスタのレシピ実行時のみ有効で、他のクラスタには影響しません。

---

## 使用例

### OS による分岐

```yaml
tasks:
  - name: install-linux
    where: "node.os == 'linux'"
    action:
      exec.remote:
        scriptTpl: |
          sudo apt-get install -y nginx

  - name: install-darwin
    where: "node.os == 'darwin'"
    action:
      exec.local:
        scriptTpl: |
          brew install nginx
```

### ラベルによる絞り込み（`kind: node`）

```yaml
tasks:
  - name: configure-master
    where: "node.labels.role == 'master'"
    action:
      exec.remote:
        scriptTpl: |
          echo "Configuring master node"
```

### 複合条件

```yaml
tasks:
  - name: configure-worker-linux
    where: "node.os == 'linux' && node.labels.role == 'worker'"
    action:
      exec.remote:
        scriptTpl: |
          echo "Configuring linux worker"
```

### `in` 演算子

```yaml
tasks:
  - name: configure-control-plane
    where: "node.labels.role in ['master', 'etcd']"
    action:
      exec.remote:
        scriptTpl: |
          echo "Configuring control plane node"
```

---

## 関連ページ

- [Nodeのcapabilityを定義する]({{< relref "node-capability" >}}) — `kind: node` のレシピ全般
- [Infraのプロビジョナーを定義する]({{< relref "infrastructure-provisioning" >}}) — `kind: infra` のレシピ全般
- [Hostの設定を定義する]({{< relref "host-configuration" >}}) — `kind: host` のレシピ全般

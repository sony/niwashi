---
title: "Kubernetesクラスタの構築"
weight: 1
---

# Kubernetesクラスタの構築

このページでは、Niwashiを使ってKubernetesクラスタを構築する完全な例を示します。

## 概要

この例では、以下の構成のKubernetesクラスタを構築します：

- **Control Plane**: 1ノード
- **Worker**: 2ノード
- **インフラ**: Vagrantで仮想マシンを生成

---

## 前提

リファレンスレシピ（Vagrant / Ansible / Kubernetes）を `./recipes` に取得済みであることを前提とします。レシピの取得方法は [ワークフロー]({{< relref "/getting-started/workflow" >}}) を参照してください。

---

## 完全な State ファイル

```yaml
version: nws.state/v1

metadata:
  project: my-k8s-cluster

# 論理構成
inventory:
  nodes:
    control-plane:
      instanceSelector:
        generator: vms

    worker-01:
      instanceSelector:
        generator: vms

    worker-02:
      instanceSelector:
        generator: vms

  clusters:
    k8s-cluster:
      nodes:
        - control-plane
        - worker-01
        - worker-02
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [control-plane]
          kube_node: [worker-01, worker-02]
          etcd: [control-plane]

# インフラ定義
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

---

## ファイルの説明

### metadata

```yaml
metadata:
  project: my-k8s-cluster
```

プロジェクト名を設定します。これは、状態ファイルを識別するためのメタデータです。

### inventory.nodes

```yaml
inventory:
  nodes:
    control-plane:
      instanceSelector:
        generator: vms

    worker-01:
      instanceSelector:
        generator: vms

    worker-02:
      instanceSelector:
        generator: vms
```

3つのノードを定義します：
- `control-plane`: Kubernetesのコントロールプレーンノード
- `worker-01`, `worker-02`: Kubernetesのワーカーノード

すべてのノードは、`vms` ジェネレーターが生成するインスタンスを使用します。

### inventory.clusters

```yaml
inventory:
  clusters:
    k8s-cluster:
      nodes:
        - control-plane
        - worker-01
        - worker-02
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [control-plane]
          kube_node: [worker-01, worker-02]
          etcd: [control-plane]
```

Kubernetesクラスタを定義します：
- **nodes**: クラスタに参加する3つのノード
- **capabilities**: `cluster.kubernetes` でKubernetesクラスタを構築
- **params.groups**: 各ノードの役割を定義
  - `kube_control_plane`: コントロールプレーンノード
  - `kube_node`: ワーカーノード
  - `etcd`: etcdが動作するノード

### infrastructure.generators

```yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

Vagrantを使って3台の仮想マシンを生成します：
- **count**: 3台のVMを生成
- **box**: Ubuntu 22.04 LTSを使用
- **cpus**: 各VM に2コア
- **memory**: 各VMに2GB のメモリ

---

## 実行手順

### 1. Stateファイルを作成

上記の内容を `k8s-cluster.yaml` として保存します。

```bash
cat > k8s-cluster.yaml <<'EOF'
version: nws.state/v1

metadata:
  project: my-k8s-cluster

inventory:
  nodes:
    control-plane:
      instanceSelector:
        generator: vms
    worker-01:
      instanceSelector:
        generator: vms
    worker-02:
      instanceSelector:
        generator: vms

  clusters:
    k8s-cluster:
      nodes:
        - control-plane
        - worker-01
        - worker-02
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [control-plane]
          kube_node: [worker-01, worker-02]
          etcd: [control-plane]

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
EOF
```

### 2. 計画を作成

```bash
nwsctl plan --recipe-dir ./recipes -t k8s-cluster.yaml
```

Niwashiは、以下の処理を計画します：
1. Vagrantで3台のVMを生成
2. 各VMをノードに割り当て
3. Kubernetesクラスタを構築

### 3. 計画を実行

```bash
nwsctl apply --plan plan.json --recipe-dir ./recipes
```

Niwashiは計画を実行し、Kubernetesクラスタを構築します。

### 4. 確認

クラスタが正しく構築されたか確認します：

```bash
# ノードの状態を確認
nwsctl ssh control-plane -- kubectl get nodes

# Pod の状態を確認
nwsctl ssh control-plane -- kubectl get pods --all-namespaces
```

---

## 環境別の構成

本番環境と開発環境で異なる構成を使い分ける例です。

### ファイル分割

```
├── base-k8s.yaml       # 論理構成（共通）
├── infra-dev.yaml      # 開発環境のインフラ
└── infra-prod.yaml     # 本番環境のインフラ
```

### base-k8s.yaml（論理構成）

```yaml
version: nws.state/v1

metadata:
  project: k8s-cluster

inventory:
  nodes:
    control-plane: {}
    worker-01: {}
    worker-02: {}

  clusters:
    k8s-cluster:
      nodes:
        - control-plane
        - worker-01
        - worker-02
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [control-plane]
          kube_node: [worker-01, worker-02]
          etcd: [control-plane]
```

### infra-dev.yaml（開発環境）

```yaml
version: nws.state/v1

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

### infra-prod.yaml（本番環境）

```yaml
version: nws.state/v1

infrastructure:
  generators:
    prod-servers:
      provisioner: external-instance
      params:
        instances:
          k8s-cp:
            connection:
              ssh:
                address:
                  host: k8s-cp.example.com
                  port: 22
                  user: ubuntu
          k8s-worker-01:
            connection:
              ssh:
                address:
                  host: k8s-worker-01.example.com
                  port: 22
                  user: ubuntu
          k8s-worker-02:
            connection:
              ssh:
                address:
                  host: k8s-worker-02.example.com
                  port: 22
                  user: ubuntu
```

### 使い分け

```bash
# 開発環境
nwsctl plan --recipe-dir ./recipes -t base-k8s.yaml -t infra-dev.yaml

# 本番環境
nwsctl plan --recipe-dir ./recipes -t base-k8s.yaml -t infra-prod.yaml
```

---

## カスタマイズ

### ワーカーノードを増やす

ワーカーノードを3台に増やす例：

```yaml
inventory:
  nodes:
    control-plane: {}
    worker-01: {}
    worker-02: {}
    worker-03: {}  # 追加

  clusters:
    k8s-cluster:
      nodes:
        - control-plane
        - worker-01
        - worker-02
        - worker-03  # 追加
      params:
        groups:
          kube_control_plane: [control-plane]
          kube_node: [worker-01, worker-02, worker-03]  # 追加
          etcd: [control-plane]

infrastructure:
  generators:
    vms:
      params:
        count: 4  # 3→4 に変更
```

### リソースを増やす

各VMのリソースを増やす例：

```yaml
infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 4      # 2→4 に変更
        memory: 4096  # 2048→4096 に変更
```

### HAクラスタ構成

Control Planeを3台にする高可用性構成：

```yaml
inventory:
  nodes:
    cp-01: {}
    cp-02: {}
    cp-03: {}
    worker-01: {}
    worker-02: {}
    worker-03: {}

  clusters:
    k8s-cluster:
      nodes:
        - cp-01
        - cp-02
        - cp-03
        - worker-01
        - worker-02
        - worker-03
      capabilities:
        - cluster.kubernetes
      params:
        groups:
          kube_control_plane: [cp-01, cp-02, cp-03]
          kube_node: [worker-01, worker-02, worker-03]
          etcd: [cp-01, cp-02, cp-03]

infrastructure:
  generators:
    vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 6
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

---

## トラブルシューティング

### VMが起動しない

**問題**: Vagrantでのイン スタンス生成に失敗する。

**解決策**:
1. Vagrantが正しくインストールされているか確認
2. 仮想化が有効になっているか確認（BIOS設定）
3. 十分なメモリがあるか確認

### Kubernetesクラスタ構築に失敗

**問題**: クラスタの構築処理でエラーが発生する。

**解決策**:
1. ノードのリソースが十分か確認（特にメモリ）
2. レシピのログを確認
3. ネットワーク接続を確認

---

## 次のステップ

- Webアプリケーションの構成 - Webアプリケーション構成の例
- [複数環境の管理]({{< relref "../defining-desired-state/advanced/multi-environment" >}}) - 環境別の構成管理

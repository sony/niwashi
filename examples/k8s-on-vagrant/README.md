
# Sample

このサンプルがやっていること。

1. 3つの仮想マシンの構築(Vagrantレシピ)
2. Kubernetesクラスタの構築(Kubesprayレシピ)

以下、目標状態(`target.yaml`)の抜粋です:

```yaml
# 1. インベントリ
inventory:
  # 1.1. ノードの定義
  # - 3つのノードの作成
  # - 詳細がなければ infrastructure.instances から自動的に割り当てられる
  nodes:
    cp:
    worker1:
    worker2:

  # 1.2. クラスタの定義
  clusters:
    k8s-cluster:
      # クラスタのノードは3つ
      nodes:
        - cp
        - worker1
        - worker2

      # orchestrator.kubernetes のケーパビリティを持つ
      capabilities:
        - orchestrator.kubernetes

      # Kubernetesレシピに渡すパラメータ
      params:
        groups:
          kube_control_plane: [ cp ]
          kube_node: [ worker1, worker2 ]
          etcd: [ cp ]

# 2. インフラの定義
infrastructure:
  # 2.1. インスタンスの定義
  instances:
    # Vagrantで3つのインスタンスを作成する
    vm:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3   # インスタンスの数
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

## 事前セットアップ

コマンドはpythonスクリプトです。面倒な場合、以下のようにショートカット作っておくと楽です。

```bash
nwsctl="python3 ../../bin/nwsctl.py" 
```

### ツールのセットアップ

```note
今後、ホストへのツールインストールは不要にする予定です。
```

以下のツールをあらかじめインストールしてください。Ansibleは最新バージョンを推奨します。

- git
- jq
- Ansible
- Vagrant
- VirtualBox

## クラスタ構築

- 実行計画の作成(`plan`)
  - コマンドの引数
    - `plan`: サブコマンド
    - `--initial ./initial.yaml`: 初期状態
    - `--target ./target.yaml`: 目標状態
    - `--recipe-dir ../../recipe`: レシピのあるディレクトリ。このこのディレクトリから再帰的に検索されます。
    - `--profile ./profile.yaml`: プロファイル。(レシピの固定化など)
    - `--out ./plan.json`: 実行計画の保存先

```bash
$nwsctl plan --initial ./initial.yaml --target ./target.yaml --recipe-dir ../../recipe --profile ./profile.yaml --out ./plan.json
```

* 計画の実行

`plan`で作成した実行計画を実行します。

- 計画の実行(`apply`)
  - コマンドの引数
  - `apply`: サブコマンド
  - `--recipe-dir ../../recipe`: レシピのあるディレクトリ。このこのディレクトリから再帰的に検索されます。
  - `--plan ./plan.json`: 実行計画
- 作業用ディレクトリ
  - デフォルトでは`.niwashi`サブディレクトリ以下で処理が実行されます
  - `--work-dir`オプションで変更できます

```bash
$nwsctl apply --recipe-dir ../../recipe --plan ./plan.json
```

## クラスタの削除

設定を初期状態に戻すには、現在の状態から**空の目標の状態**へのセットアップを実行します。

現在の状態モデルは、`export`サブコマンドを使ってエクスポートできます。

```bash
$nwsctl export --out current.yaml
```

* 実行計画の作成

`plan`で目標の状態を空(`initial.yaml`)にして、現在の状態(`current.yaml`)からの実行計画を作成します。

```bash
$nwsctl plan --initial ./current.yaml  --target ./initial.yaml  --recipe-dir ../../recipe  --profile ./profile.yaml  --out ./destroy-plan.json
```

* 計画の実行

`apply`で計画を実行します。

```bash
$nwsctl apply --recipe-dir ../../recipe --plan ./destroy-plan.json
```

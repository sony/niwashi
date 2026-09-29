---
title: "レシピ命名ガイドライン"
weight: 9
---

# レシピ命名ガイドライン

このページでは、レシピの識別子とケーパビリティエイリアスの命名規則を説明します。このガイドラインに従うことで、レシピの発見・理解・組み合わせがスムーズになります。

---

## レシピID（`metadata.id`）

レシピIDは `<org>/<name>` の形式で記述します：

```
my-org/ansible.installer
my-org/kubernetes.by.kubespray
```

### `org` — 組織またはネームスペース

`org` セグメントはレシピの所有者を示します。GitHubの組織名、会社名、またはプロジェクト名を使用してください。

| 例 | 用途 |
|----|------|
| `nws` | Niwashiのリファレンスレシピ用（組み込み・配布レシピを含む）（予約済み） |
| `my-org` | GitHubの組織名または会社名 |
| `acme` | プロジェクトやチームの短縮名 |

**予約済みネームスペース**: `nws` はNiwashi専用です。サードパーティのレシピには使用しないでください。

**推奨しない名前**: `core`、`official`、`system`、`builtin`、`verified`、`trusted`、`standard` など、公式の保証や検証済みであることを示唆するorg名は使用しないでください。このような名前を使用すると、レシピがNiwashiによって公式に承認されたものであるとユーザーが誤解する可能性があります。

### `name` — レシピ名

`name` セグメントはそのレシピが何をするかを表します。小文字のドット区切りで階層を表現します：

```
ansible.installer       # Ansibleのインストーラ
kubernetes.by.kubespray # Kubesprayを使ったKubernetesセットアップ
vagrant.vm-cluster      # VagrantベースのVMクラスタ
nginx                   # 単一セグメントの名前も有効
```

- ドット（`.`）で関連するレシピをグループ化する（例: `ansible.installer`, `ansible.adapter`）
- 右端のセグメントが最も具体的な説明
- ハイフン（`-`）はセグメント内で使用可能。アンダースコア（`_`）も使用可能

### FQID

**FQID（Fully Qualified ID）** は `metadata.id` と `metadata.version` を `@` で結合したものです：

```
my-org/ansible.installer@1.0.0
nws/kubernetes.by.kubespray@0.1.0
```

FQIDは `spec.requires` で完全一致またはバージョン制約付きの参照に使用します。バージョン解決のルールは [レシピのロード仕様]({{< relref "recipe-loading" >}}) を参照してください。

---

## ケーパビリティエイリアス（`spec.provides[].name`）

ケーパビリティエイリアスは、レシピが提供する能力の**意味的な名前**です。誰が書いたか・どのバージョンかに依存せず、コンシューマが `capabilities:` や `spec.requires:` に記述する名前です。

### 命名構造

エイリアスはドット区切りの階層で表現します。**第1セグメントはそのレシピの `kind`** を示します：

> **注意**: `capabilities`、`provisioner`、`toolRef` など、書く位置で `kind` が確定するフィールドでは、将来のリリースでkindプレフィックスを省略可能にする予定です。`spec.requires` では異なるkindが混在するため、kindプレフィックスは引き続き必須となります。

| 第1セグメント | 対応する `kind` | 説明 |
|--------------|----------------|------|
| `host.tool`  | `host`         | ホストにインストールされたツール（Ansible、Gitなど） |
| `host.service` | `host`       | ホストで動作するサービス（Dockerデーモンなど） |
| `adapter.tool` | `adapter`    | ツール型実行環境のアダプタ |
| `adapter.runtime` | `adapter` | ランタイム型実行環境のアダプタ（Python venvなど） |
| `node.runtime` | `node`       | ノードにインストールされたランタイム能力 |
| `cluster`    | `cluster`      | クラスタレベルの能力（Kubernetesなど） |
| `infra.vm`   | `infra`        | VMインフラ |
| `infra.network` | `infra`     | ネットワークインフラ |
| `infra.compute` | `infra`     | コンピュートインフラ（クラウドインスタンス、ベアメタルなど） |

> **注意**: 単一セグメント名（例: `external-instance`）はNiwashi組み込みのシステムレシピ用に予約されています。

### 例

```yaml
# kind: host — ツールインストーラ
spec:
  provides:
    - name: host.tool.ansible

# kind: host — サービスインストーラ
spec:
  provides:
    - name: host.service.docker

# kind: adapter — ツールアダプタ
spec:
  provides:
    - name: adapter.tool.ansible

# kind: adapter — ランタイムアダプタ
spec:
  provides:
    - name: adapter.runtime.python-venv

# kind: node — ランタイム能力
spec:
  provides:
    - name: node.runtime.container

# kind: cluster — クラスタ能力
spec:
  provides:
    - name: cluster.kubernetes
      attrs: { by: kubespray }

# kind: infra — VMプロビジョナ
spec:
  provides:
    - name: infra.vm
      attrs: { driver: vagrant }
```

### `attrs` で実装を区別する

同じエイリアス名を複数のレシピが提供する場合、`attrs` で区別します：

```yaml
# レシピA
spec:
  provides:
    - name: cluster.kubernetes
      attrs: { by: kubespray }

# レシピB
spec:
  provides:
    - name: cluster.kubernetes
      attrs: { by: rke2 }
```

コンシューマは特定の実装を選択できます：

```yaml
capabilities:
  - cluster.kubernetes              # 候補が1つなら自動解決
  - cluster.kubernetes.by=kubespray # レシピAを明示的に選択
```

エイリアス解決のルールは [レシピのロード仕様]({{< relref "recipe-loading" >}}) を参照してください。

### `attrs` キーの命名

- 小文字とアンダースコアのみ使用可（`[a-z][a-z0-9_]*`）
- バージョンで絞り込みたい場合はFQIDを使用してください。`2.16.3` のようなバージョン文字列はドットがセパレータとして予約されているためattsの値に使用できません

**attrsキーを定義する基準**: attrsキーは、同じ `provides.name` を持つ複数のレシピを区別するために使用します。実行時に動的に決まる値（`os`, `arch` など）はattrsに入れず、タスクの `where` 条件で処理してください。

以下のキーはよくあるケース向けに定義されています：

| キー | 使うケース | 例 |
|-----|-----------|-----|
| `engine` | ソフトウェアスタックの実行エンジン・実装を選ぶ | `cluster.kubernetes.engine=k3s`, `node.runtime.proxy.engine=nginx` |
| `by` | セットアップ・構成を担うツールを選ぶ | `cluster.kubernetes.by=kubespray` |
| `driver` | インフラ層のドライバ・プロバイダを選ぶ | `infra.vm.driver=vagrant` |
| `from` | 入力ソースや形式を選ぶ | `external-instance.from=file`, `external-instance.from=ssh_config` |

上記に該当しないケースでは、変化の軸を明確に表す名詞でキーを定義してください。

```yaml
# NG: バージョン文字列をattsの値に使用
attrs: { version: 2.16.3 }

# OK: バージョン固定にはspec.requiresでFQIDを使用
spec:
  requires:
    - my-org/my-recipe@2.16.3
```

---

## タスク名（`tasks[].name`）

タスク名はログやplan出力に表示される人間可読なラベルです：

- 英数字で始める
- 2文字目以降は英数字・スペース・ハイフン・アンダースコア・ピリオドが使用可能
- 大文字も使用可能

```yaml
tasks:
  - name: Install Ansible
  - name: check-version
  - name: Apply k8s manifests
```

> **注意**: タスク名は内部的にファイル名として使用される際、小文字に変換され、スペースとピリオドはファイルシステムセーフな文字に置換されます。大文字・小文字・スペース・ピリオドのみが異なる名前は重複とみなされます（例: `My Task`、`my task`、`my.task` はすべて衝突します）。

---

## コマンド名（adapterレシピの `spec.commands`）

コマンド名はPOSIXユーティリティの命名規則に準じます：

- 小文字のアルファベットで始める
- 2文字目以降は小文字・数字・ハイフン・アンダースコアが使用可能
- 大文字は使用不可

```yaml
spec:
  commands:
    apply:
      ...
    dry-run:
      ...
    install_package:
      ...
```

---

## まとめ

| 識別子 | パターン | 例 |
|--------|---------|-----|
| `metadata.id` | `<org>/<name>`（小文字、ドット・ハイフン可） | `my-org/ansible.installer` |
| `metadata.version` | セマンティックバージョン `MAJOR.MINOR.PATCH` | `1.2.0` |
| FQID | `<id>@<version>` | `my-org/ansible.installer@1.2.0` |
| `spec.provides[].name` | ドット区切り、kindプレフィックスで始める | `host.tool.ansible` |
| `spec.provides[].attrs` キー | `[a-z][a-z0-9_]*` | `by`, `driver` |
| `spec.provides[].attrs` 値 | `[a-z0-9][a-z0-9_-]*` | `kubespray`, `vagrant` |
| `tasks[].name` | 人間可読、英数字始め | `Install Ansible` |
| `spec.commands` キー | POSIXスタイル、小文字始め | `apply`, `dry-run` |

---
title: "レシピの読み込み仕様"
weight: 6
---

# レシピの読み込み仕様

レシピがどのように探索・識別・解決されるかの詳細を説明します。

---

## レシピの探索と読み込み

`nwsctl plan --recipe-dir <dir>` で指定されたディレクトリは**再帰的に**探索されます。見つかったファイル名に応じて、以下の処理が行われます：

| ファイル名 | 処理 |
|-----------|------|
| `nws-recipe.yaml` | そのファイルを直接レシピとしてロード |
| `nws-catalog.yaml` | カタログを読み込み、そこで参照されている各レシピファイルをロード |
| その他 | 無視 |

> **補足**: ファイル名・ディレクトリ名がドット（`.`）で始まるもの（例: `.git/`、`.github/`、`.gitlab-ci.yml`）は、再帰探索中であっても常にスキップされます。

> **補足**: 「その他」に該当する`.yaml`/`.yml`ファイルのうち、いずれのレシピの`spec.assets`にも列挙されていないものは、警告としてログに出力されます。

複数の `--recipe-dir` を指定した場合は、左から順に処理されます。

```bash
# recipes/ → custom-recipes/ の順でロード
nwsctl plan \
  --recipe-dir ./recipes \
  --recipe-dir ./custom-recipes \
  -t state.yaml
```

---

## FQID とバージョン管理

### FQID とは

レシピは **FQID（Fully Qualified ID）** で一意に識別されます。FQIDは `metadata.id` と `metadata.version` の組み合わせです：

```
<id>@<version>
```

例：
```
nws/kubernetes.by.kubespray@0.0.1
nws/ansible.adapter@0.1.0
```

### バージョンの共存

同じ `id` でもバージョンが異なれば、複数のレシピを共存させることができます：

```
# 共存可能
my-org/my-recipe@1.0.0
my-org/my-recipe@1.2.0
```

```
# エラー（同じFQIDが重複）
my-org/my-recipe@1.0.0  ← 1つ目
my-org/my-recipe@1.0.0  ← 2つ目（エラー）
```

### バージョン指定

`spec.requires` やCapabilityの指定でバージョンを省略した場合、セマンティックバージョニングに従って最新版が選ばれます。バージョン制約を指定することも可能です：

```yaml
spec:
  requires:
    - my-org/my-recipe          # 最新版
    - my-org/my-recipe@1.0.0   # 完全一致
    - my-org/my-recipe@^1.0.0  # 1.x.x の最新版
```

---

## provides による別名（Alias）

### 別名とは

レシピは `spec.provides` で**別名（alias）** を定義できます。別名を使うことで、ユーザーはFQIDを意識せずにレシピを参照できます：

```yaml
# レシピ定義
metadata:
  id: nws/kubernetes.by.kubespray
  version: 0.0.1
spec:
  provides:
    - name: cluster.kubernetes
      attrs: { by: kubespray }
```

```yaml
# 利用側（別名で参照）
capabilities:
  - cluster.kubernetes
```

### 属性による絞り込み

別名に属性（`attrs`）を付与することで、同じ別名でも複数の実装を共存させられます。属性はドット区切りで指定します：

```yaml
# 別名: cluster.kubernetes
# 属性: by=kubespray
capabilities:
  - cluster.kubernetes          # 候補が1つなら解決、複数ならエラー
  - cluster.kubernetes.by=kubespray  # 属性を指定して完全一致
```

### 別名解決のルール

1. **完全一致**: 指定した属性がエントリの属性と完全に一致する場合、そのレシピを返す
2. **部分一致**: 指定した属性がエントリの属性のサブセットである候補が1つの場合、そのレシピを返す
3. **曖昧**: 部分一致の候補が複数ある場合、エラーになる
4. **未発見**: 一致する候補がない場合、エラーになる

**例**:

```yaml
# 2つのレシピが同じ別名を提供しているケース
# レシピA: name=cluster.kubernetes, attrs={by: kubespray}
# レシピB: name=cluster.kubernetes, attrs={by: rke2}

# 属性なしで指定 → 候補が2つで曖昧エラー
- cluster.kubernetes

# 属性を指定 → 完全一致でレシピAが選ばれる
- cluster.kubernetes.by=kubespray
```

---

## アダプターの別名

アダプターレシピ（`spec.adapter` を持つレシピ）の別名は、通常のCapabilityとは別のテーブルで管理されます。`tool.run` の `toolRef` で指定する際に使用されます：

```yaml
# アダプターレシピの定義
kind: adapter
spec:
  provides:
    - name: adapter.tool.ansible

# 利用側（tool.run で参照）
action:
  tool.run:
    toolRef: adapter.tool.ansible
```

詳細は [アダプターを利用する]({{< relref "using-adapters" >}}) を参照してください。

---
title: "識別子の命名規則"
weight: 4
---

# 識別子の命名規則

ノード、クラスタ、ジェネレーターを定義する際、それぞれに**識別子**（キー）を指定します。この識別子は、すべての要素で共通の命名規則に従う必要があります。

## 使用可能な文字

識別子には以下の文字を使用できます：

- **英字**（大文字・小文字）: `a-z`, `A-Z`
- **数字**: `0-9`
- **アンダースコア**: `_`
- **ハイフン**: `-`

## 命名パターン

識別子は正規表現パターン **`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`** に一致する必要があります。

つまり：
- 先頭は英数字
- 2文字目以降は英数字、アンダースコア、ハイフンのみ使用可能
- スペース、ドット、特殊文字は使用不可

---

## 有効な識別子の例

```yaml
inventory:
  nodes:
    node1: {}          # 英数字
    web-server: {}     # ハイフン
    db_primary: {}     # アンダースコア
    Worker-01: {}      # 大文字小文字の混在
    api-server-v2: {}  # 組み合わせ

  clusters:
    k8s-cluster: {}
    web-tier-01: {}

infrastructure:
  generators:
    vm-gen-01: {}
    local-vms: {}
```

すべて有効な識別子です。

---

## 無効な識別子の例

以下の識別子は使用できません：

```yaml
# スペースは不可
inventory:
  nodes:
    my node: {}         # ❌

# @ 記号は不可
inventory:
  nodes:
    node@01: {}         # ❌

# # 記号は不可
inventory:
  nodes:
    node#1: {}          # ❌

# 非ASCII文字は不可
inventory:
  nodes:
    ノード1: {}          # ❌

# スラッシュは不可
infrastructure:
  generators:
    vm/prod: {}         # ❌

# カッコは不可
inventory:
  clusters:
    cluster(prod): {}   # ❌
```

---

## 命名スタイル

識別子には、いくつかの命名スタイルがあります。プロジェクト内で統一したスタイルを使用することをお勧めします。

### ケバブケース（推奨）

ハイフン `-` で単語を区切ります：

```yaml
inventory:
  nodes:
    web-server: {}
    db-primary: {}
    cache-server: {}

  clusters:
    k8s-cluster: {}
    web-tier: {}

infrastructure:
  generators:
    local-vms: {}
    prod-servers: {}
```

**推奨理由**: 読みやすく、多くのツールで一般的に使用されます。

### スネークケース

アンダースコア `_` で単語を区切ります：

```yaml
inventory:
  nodes:
    web_server: {}
    db_primary: {}
    cache_server: {}
```

### パスカルケース風

大文字で単語を区切ります（ただし、最初は小文字が一般的）：

```yaml
inventory:
  nodes:
    WebServer: {}
    DbPrimary: {}
```

**注意**: スネークケースやケバブケースの方が一般的です。

---

## 命名のベストプラクティス

### 1. 一貫性を保つ

プロジェクト内で統一した命名スタイルを使用します：

```yaml
# Good: ケバブケースで統一
inventory:
  nodes:
    web-server: {}
    db-primary: {}
    cache-server: {}

# Bad: スタイルが混在
inventory:
  nodes:
    web-server: {}
    db_primary: {}
    CacheServer: {}
```

### 2. わかりやすい名前を付ける

識別子の役割が明確に分かる名前を付けます：

```yaml
# Good: 役割が明確
inventory:
  nodes:
    web-frontend: {}
    api-backend: {}
    postgres-primary: {}

# Bad: 役割が不明確
inventory:
  nodes:
    node1: {}
    server2: {}
    thing3: {}
```

### 3. 簡潔さを心がける

必要以上に長い名前は避けます：

```yaml
# Good: 簡潔で明確
inventory:
  nodes:
    web-01: {}
    web-02: {}

# Bad: 冗長
inventory:
  nodes:
    web-server-instance-number-01: {}
    web-server-instance-number-02: {}
```

### 4. 環境や番号を含める

複数環境や複数インスタンスを扱う場合は、識別子に含めます：

```yaml
inventory:
  nodes:
    # 番号で区別
    web-01: {}
    web-02: {}
    web-03: {}

    # 環境で区別（ファイルを分ける場合は不要）
    db-prod: {}
    db-dev: {}

infrastructure:
  generators:
    # 環境を含める
    vms-dev: {}
    vms-prod: {}
```

### 5. 予約語を避ける

一般的な予約語や特殊な意味を持つ単語は避けます：

```yaml
# Avoid（避けるべき）
inventory:
  nodes:
    default: {}
    system: {}
    root: {}
```

---

## 実践例

### シンプルな命名

```yaml
inventory:
  nodes:
    web: {}
    app: {}
    db: {}
```

小規模なプロジェクトや開発環境では、シンプルな名前で十分です。

### 役割ベースの命名

```yaml
inventory:
  nodes:
    web-frontend: {}
    api-backend: {}
    postgres-db: {}
    redis-cache: {}

  clusters:
    web-tier: {}
    data-tier: {}
```

役割を明示することで、複雑なシステムでも理解しやすくなります。

### 番号付き命名

```yaml
inventory:
  nodes:
    control-plane: {}
    worker-01: {}
    worker-02: {}
    worker-03: {}
```

同じ役割のノードを複数定義する場合に便利です。

## `params` キーの命名

`params` フィールドはユーザーが自由に定義するデータ構造です。任意の深さのキーは識別子と同じパターンに従います：

- **パターン**: `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`
- 英数字、アンダースコア、ハイフンが使用可能
- 先頭は英数字

これにより、インスタンス名やホスト名などの実際の識別子をキーとして使用できます。

```yaml
infrastructure:
  generators:
    my-vms:
      params:
        count: 3           # シンプルなキー
        prefix: worker     # シンプルなキー
        instances:
          worker-01: {}    # ハイフン入りキー — 有効
          worker-02: {}
```

> **注意**: テンプレート変数（`{{ .Params.xxx }}`）で `params` の値にアクセスする場合、ハイフンを含むキーはドット記法でアクセスできません。テンプレートから参照するキーにはハイフンではなくアンダースコアを使用してください。
>
> ```yaml
> params:
>   applied_version: "1.0.0"   # {{ .Params.applied_version }} でアクセス可能
>   applied-version: "1.0.0"   # ドット記法ではアクセス不可
> ```

---

## まとめ

- **使用可能**: 英数字、アンダースコア、ハイフン
- **パターン**: `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`
- **推奨スタイル**: ケバブケース（例: `web-server`）
- **ベストプラクティス**: 一貫性、わかりやすさ、簡潔さ

識別子の命名規則を守ることで、Stateファイルの可読性と保守性が向上します。

---

## 次のステップ

基本概念を学び終えたら、各要素の詳細について学びましょう：

- [ノードの詳細]({{< relref "../nodes" >}}) - ノードの全属性とオプション
- [クラスタの詳細]({{< relref "../clusters" >}}) - クラスタの全属性とオプション
- [ジェネレーターの詳細]({{< relref "../generators" >}}) - ジェネレーターの全属性とオプション

---
title: "テンプレートの詳細"
weight: 12
---

# テンプレートの詳細

テンプレートを使用することで、共通の設定を持つ複数のノード、クラスタ、インスタンスを効率的に定義できます。テンプレートは、設定の重複を避け、メンテナンス性を向上させます。

## テンプレートとは

テンプレートは、ノード、クラスタ、ジェネレーターの共通設定を定義するための仕組みです。テンプレートを定義しておくことで、同じ設定を複数の要素で再利用できます。

### テンプレートのメリット

- **重複の削減**: 同じ設定を何度も書く必要がない
- **一貫性**: 共通設定を一箇所で管理
- **保守性**: 変更が必要な場合、テンプレートを変更するだけ

---

## テンプレートの定義

テンプレートは、`template` セクションの下に定義します。

```yaml
template:
  node:
    # ノードテンプレートの定義
  cluster:
    # クラスタテンプレートの定義
  instance:
    # インスタンステンプレートの定義
```

### 定義できるテンプレート

- **template.node**: ノードのテンプレート
- **template.cluster**: クラスタのテンプレート
- **template.instance**: ジェネレーターのテンプレート

---

## ノードテンプレート

ノードの共通設定をテンプレートとして定義できます。

### 基本的な定義

```yaml
template:
  node:
    base-server:
      labels:
        environment: production
        managed_by: niwashi

inventory:
  nodes:
    web-01:
      templates: [base-server]
      capabilities:
        - web.nginx
```

この例では、`web-01` は `base-server` テンプレートのlabelsを継承します。

### 複数のテンプレート

```yaml
template:
  node:
    base-server:
      labels:
        environment: production
        managed_by: niwashi

    web-base:
      labels:
        tier: frontend
      capabilities:
        - monitoring.prometheus-exporter

inventory:
  nodes:
    web-01:
      templates: [base-server, web-base]
      capabilities:
        - web.nginx

    web-02:
      templates: [base-server, web-base]
      capabilities:
        - web.nginx
```

この例では、`web-01` と `web-02` は、`base-server` と `web-base` の設定を継承し、両方とも同じlabelsとmonitoringのcapabilityを持ちます。

---

## クラスタテンプレート

クラスタの共通設定もテンプレート化できます。

```yaml
template:
  cluster:
    production-cluster:
      labels:
        environment: production
        backup_enabled: "true"

inventory:
  clusters:
    k8s-prod:
      templates: [production-cluster]
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes

    db-prod:
      templates: [production-cluster]
      nodes: [db-primary, db-replica]
```

---

## インスタンステンプレート

インフラストラクチャのジェネレーターでも、テンプレートが利用できます。

```yaml
template:
  instance:
    standard-vm:
      provisioner: infra.vm.driver=vagrant
      params:
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048

infrastructure:
  generators:
    web-vms:
      templates: [standard-vm]
      params:
        count: 2

    db-vms:
      templates: [standard-vm]
      params:
        count: 1
        memory: 4096  # メモリだけ上書き
```

この例では、`web-vms` と `db-vms` は `standard-vm` テンプレートの設定を継承し、`db-vms` はメモリだけを上書きしています。

---

## テンプレートの継承（from）

テンプレート自体も、他のテンプレートを継承できます（単一継承）。

```yaml
template:
  node:
    base:
      labels:
        managed_by: niwashi

    production:
      from: base
      labels:
        environment: production

    production-web:
      from: production
      labels:
        tier: frontend
      capabilities:
        - monitoring.prometheus-exporter

inventory:
  nodes:
    web-server:
      templates: [production-web]
      capabilities:
        - web.nginx
```

この例では：
1. `base` テンプレート: `managed_by: niwashi`
2. `production` テンプレート: `base` を継承 + `environment: production`
3. `production-web` テンプレート: `production` を継承 + `tier: frontend` + monitoring capability
4. `web-server` ノード: `production-web` を使用

結果として、`web-server` は以下の設定を持ちます：
- `managed_by: niwashi`（baseから）
- `environment: production`（productionから）
- `tier: frontend`（production-webから）
- `monitoring.prometheus-exporter`（production-webから）
- `web.nginx`（ノード自身で定義）

---

## 複数テンプレートの適用

ノード、クラスタ、インスタンスには、複数のテンプレートを配列で指定できます。後に指定したテンプレートが前のテンプレートの設定を上書きします。

```yaml
template:
  node:
    base:
      labels:
        environment: development

    monitoring:
      capabilities:
        - monitoring.prometheus-exporter

inventory:
  nodes:
    web-server:
      templates: [base, monitoring]
      labels:
        environment: production  # baseの設定を上書き
```

この例では、`web-server` は：
1. `base` テンプレートから `environment: development` を継承
2. `monitoring` テンプレートから `monitoring.prometheus-exporter` を継承
3. ノード自身で `environment: production` を定義し、baseの設定を上書き

---

## 属性のマージ動作

テンプレートと実際の定義をマージする際の動作は、属性によって異なります。

### labels

マージされます（同じキーは上書き）：

```yaml
template:
  node:
    template-a:
      labels:
        env: dev
        region: us-east

inventory:
  nodes:
    my-node:
      templates: [template-a]
      labels:
        env: prod  # 上書き
        tier: web  # 追加

# 結果:
# labels:
#   env: prod         (上書きされた)
#   region: us-east   (テンプレートから)
#   tier: web         (追加された)
```

### capabilities

配列として結合されます：

```yaml
template:
  node:
    template-a:
      capabilities:
        - capability-a

    template-b:
      capabilities:
        - capability-b

inventory:
  nodes:
    my-node:
      templates: [template-a, template-b]
      capabilities:
        - capability-c

# 結果:
# capabilities:
#   - capability-a  (template-aから)
#   - capability-b  (template-bから)
#   - capability-c  (ノード自身で定義)
```

### 複数テンプレート適用時のマージ

```yaml
template:
  node:
    template-a:
      labels:
        env: dev
        region: us-east
      capabilities:
        - capability-a

    template-b:
      labels:
        env: prod  # template-aを上書き
        tier: web  # 追加
      capabilities:
        - capability-b  # 追加

inventory:
  nodes:
    my-node:
      templates: [template-a, template-b]
      labels:
        app: myapp  # 追加
      capabilities:
        - capability-c  # 追加

# 結果:
# labels:
#   env: prod         (template-bがtemplate-aを上書き)
#   region: us-east   (template-aから)
#   tier: web         (template-bから)
#   app: myapp        (ノード自身で定義)
# capabilities:
#   - capability-a    (template-aから)
#   - capability-b    (template-bから)
#   - capability-c    (ノード自身で定義)
```

---

## 実践例

### 例1: 環境別テンプレート

```yaml
template:
  node:
    base:
      labels:
        managed_by: niwashi

    development:
      from: base
      labels:
        environment: development

    staging:
      from: base
      labels:
        environment: staging

    production:
      from: base
      labels:
        environment: production

inventory:
  nodes:
    dev-web-01:
      templates: [development]
      capabilities:
        - web.nginx

    staging-web-01:
      templates: [staging]
      capabilities:
        - web.nginx

    prod-web-01:
      templates: [production]
      capabilities:
        - web.nginx
```

### 例2: 役割別テンプレート

```yaml
template:
  node:
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter

    db-server:
      capabilities:
        - database.postgresql
        - backup.automated

    cache-server:
      capabilities:
        - cache.redis
        - monitoring.prometheus-exporter

inventory:
  nodes:
    web-01:
      templates: [web-server]

    web-02:
      templates: [web-server]

    db-01:
      templates: [db-server]

    cache-01:
      templates: [cache-server]
```

### 例3: 組み合わせ

環境と役割を組み合わせる：

```yaml
template:
  node:
    # 環境別
    production:
      labels:
        environment: production

    staging:
      labels:
        environment: staging

    # 役割別
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter

    db-server:
      capabilities:
        - database.postgresql
        - backup.automated

inventory:
  nodes:
    prod-web-01:
      templates: [production, web-server]

    prod-web-02:
      templates: [production, web-server]

    prod-db-01:
      templates: [production, db-server]

    staging-web-01:
      templates: [staging, web-server]

    staging-db-01:
      templates: [staging, db-server]
```

### 例4: 大規模環境

```yaml
template:
  node:
    # 共通設定
    base:
      labels:
        managed_by: niwashi

    # 環境別
    production:
      from: base
      labels:
        environment: production

    # 役割別
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter
        - logging.fluentd

    api-server:
      capabilities:
        - runtime.python
        - web.gunicorn
        - monitoring.prometheus-exporter

    db-server:
      capabilities:
        - database.postgresql
        - backup.automated
        - monitoring.postgres-exporter

inventory:
  nodes:
    # Webサーバー
    prod-web-01:
      templates: [production, web-server]
    prod-web-02:
      templates: [production, web-server]

    # APIサーバー
    prod-api-01:
      templates: [production, api-server]
    prod-api-02:
      templates: [production, api-server]

    # データベース
    prod-db-primary:
      templates: [production, db-server]
      capabilities:
        - id: database.postgresql
          params:
            replication_role: primary

    prod-db-replica:
      templates: [production, db-server]
      capabilities:
        - id: database.postgresql
          params:
            replication_role: replica
```

---

## ベストプラクティス

### 1. 階層的な構造を作る

```yaml
template:
  node:
    base:
      # すべてのノードに共通の設定

    environment-specific:
      from: base
      # 環境固有の設定

    role-specific:
      from: environment-specific
      # 役割固有の設定
```

### 2. 適度な粒度で分割する

テンプレートを細かく分割しすぎると、かえって複雑になります。適度な粒度で分割しましょう。

```yaml
# Good: 適度な粒度
template:
  node:
    web-server:
      # Webサーバーとして必要な設定をまとめる
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter

# Bad: 細かすぎる
template:
  node:
    nginx-only:
      capabilities:
        - web.nginx

    monitoring-only:
      capabilities:
        - monitoring.prometheus-exporter

# 使う側が複雑になる
inventory:
  nodes:
    web-01:
      templates: [nginx-only, monitoring-only]
```

### 3. 命名規則を統一する

```yaml
template:
  node:
    base-*:        # 基本テンプレート
    env-*:         # 環境別テンプレート
    role-*:        # 役割別テンプレート
```

---

## トラブルシューティング

### テンプレートが見つからない

**問題**: `templates: [my-template]` を指定したが、テンプレートが見つからない。

**原因**:
- テンプレート名の綴りが間違っている
- `template.node` セクションにテンプレートが定義されていない

**解決策**:
1. テンプレート名の綴りを確認する
2. `template.node` セクションにテンプレートを定義する

### 期待通りの設定にならない

**問題**: テンプレートを使用したが、期待通りの設定にならない。

**原因**:
- マージの順序を理解していない
- 上書きされている

**解決策**:
1. テンプレートの適用順序を確認する（配列の順）
2. マージ動作を理解する（labels は上書き、capabilities は結合）

---

## 次のステップ

- [ノードの詳細]({{< relref "nodes" >}}) - ノードでのテンプレート使用
- [クラスタの詳細]({{< relref "clusters" >}}) - クラスタでのテンプレート使用
- [ジェネレーターの詳細]({{< relref "generators" >}}) - ジェネレーターでのテンプレート使用

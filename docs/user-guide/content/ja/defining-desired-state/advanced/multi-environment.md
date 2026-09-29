---
title: "複数環境の管理"
weight: 1
---

# 複数環境の管理

Niwashiでは、複数の環境（開発環境、ステージング環境、本番環境など）を効率的に管理できます。Stateファイルのマージ機能を活用することで、環境ごとの違いを別ファイルで管理し、柔軟に切り替えられます。

## 基本的な考え方

複数環境を管理する際の基本的な戦略は、**共通部分と環境固有部分を分離する**ことです。

### 分離の方針

- **共通部分**: すべての環境で同じ設定（論理構成、アプリケーション構成など）
- **環境固有部分**: 環境ごとに異なる設定（インフラ構成、リソースサイズなど）

---

## パターン1: 論理構成とインフラ構成の分離

### ファイル構成

```
├── app.yaml          # 論理構成（共通）
├── infra-dev.yaml    # 開発環境のインフラ
└── infra-prod.yaml   # 本番環境のインフラ
```

### app.yaml（論理構成 - 共通）

アプリケーションの論理的な構成を定義します。環境に依存しない部分です。

```yaml
version: nws.state/v1

metadata:
  project: my-web-app

inventory:
  nodes:
    web:
      capabilities:
        - web.nginx
    app:
      capabilities:
        - runtime.python
        - web.gunicorn
    db:
      capabilities:
        - database.postgresql
```

### infra-dev.yaml（開発環境）

開発環境用のインフラ構成。ローカルの仮想マシンを使用します。

```yaml
version: nws.state/v1

infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 1
        memory: 1024
```

### infra-prod.yaml（本番環境）

本番環境用のインフラ構成。既存のクラウドサーバーを使用します。

```yaml
version: nws.state/v1

infrastructure:
  generators:
    prod-servers:
      provisioner: external-instance
      params:
        instances:
          web-prod:
            connection:
              ssh:
                address:
                  host: web.example.com
                  port: 22
                  user: deploy
                auth:
                  method: privateKey
                  privateKeyPath: ~/.ssh/deploy_key

          app-prod:
            connection:
              ssh:
                address:
                  host: app.example.com
                  port: 22
                  user: deploy
                auth:
                  method: privateKey
                  privateKeyPath: ~/.ssh/deploy_key

          db-prod:
            connection:
              ssh:
                address:
                  host: db.example.com
                  port: 22
                  user: deploy
                auth:
                  method: privateKey
                  privateKeyPath: ~/.ssh/deploy_key
```

### 使い分け

```bash
# 開発環境
nwsctl plan -t app.yaml -t infra-dev.yaml

# 本番環境
nwsctl plan -t app.yaml -t infra-prod.yaml
```

---

## パターン2: ベース + 環境固有の差分

### ファイル構成

```
├── base.yaml         # 共通の基本設定
├── dev.yaml          # 開発環境の差分
├── staging.yaml      # ステージング環境の差分
└── prod.yaml         # 本番環境の差分
```

### base.yaml（共通設定）

```yaml
version: nws.state/v1

metadata:
  project: my-service

inventory:
  nodes:
    web-01: {}
    web-02: {}
    db-01: {}

  clusters:
    web-tier:
      nodes:
        - web-01
        - web-02
```

### dev.yaml（開発環境の差分）

```yaml
version: nws.state/v1

inventory:
  nodes:
    web-01:
      labels:
        environment: development
    web-02:
      labels:
        environment: development
    db-01:
      labels:
        environment: development

infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 1
        memory: 1024
```

### prod.yaml（本番環境の差分）

```yaml
version: nws.state/v1

inventory:
  nodes:
    web-01:
      labels:
        environment: production
    web-02:
      labels:
        environment: production
    db-01:
      labels:
        environment: production

infrastructure:
  generators:
    prod-servers:
      provisioner: external-instance
      params:
        instances:
          # 既存サーバーの定義
```

### 使い分け

```bash
# 開発環境
nwsctl plan -t base.yaml -t dev.yaml

# 本番環境
nwsctl plan -t base.yaml -t prod.yaml
```

---

## パターン3: 多層構造

より複雑な環境では、複数のレイヤーに分割できます。

### ファイル構成

```
├── 01-logical.yaml       # 論理構成
├── 02-capabilities.yaml  # Capability定義
├── 03-infra-dev.yaml     # 開発環境インフラ
├── 03-infra-staging.yaml # ステージング環境インフラ
└── 03-infra-prod.yaml    # 本番環境インフラ
```

### 01-logical.yaml（論理構成）

```yaml
version: nws.state/v1

metadata:
  project: microservices-app

inventory:
  nodes:
    api-01: {}
    api-02: {}
    frontend-01: {}
    frontend-02: {}
    db-primary: {}
    db-replica: {}

  clusters:
    api-tier:
      nodes:
        - api-01
        - api-02

    frontend-tier:
      nodes:
        - frontend-01
        - frontend-02

    db-cluster:
      nodes:
        - db-primary
        - db-replica
```

### 02-capabilities.yaml（Capability定義）

```yaml
version: nws.state/v1

inventory:
  nodes:
    api-01:
      capabilities:
        - runtime.python
        - web.gunicorn

    api-02:
      capabilities:
        - runtime.python
        - web.gunicorn

    frontend-01:
      capabilities:
        - web.nginx

    frontend-02:
      capabilities:
        - web.nginx

    db-primary:
      capabilities:
        - database.postgresql

    db-replica:
      capabilities:
        - database.postgresql

  clusters:
    db-cluster:
      capabilities:
        - database.replication
      params:
        primary: db-primary
        replicas: [db-replica]
```

### 03-infra-dev.yaml（開発環境）

```yaml
version: nws.state/v1

infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 6
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

### 使い分け

```bash
# 開発環境
nwsctl plan -t 01-logical.yaml -t 02-capabilities.yaml -t 03-infra-dev.yaml

# ステージング環境
nwsctl plan -t 01-logical.yaml -t 02-capabilities.yaml -t 03-infra-staging.yaml

# 本番環境
nwsctl plan -t 01-logical.yaml -t 02-capabilities.yaml -t 03-infra-prod.yaml
```

---

## ファイル命名のベストプラクティス

### 番号プレフィックス

マージの順序を明確にするため、ファイル名に番号を付けることをお勧めします：

```
01-base.yaml
02-app.yaml
03-infra-dev.yaml
```

この方法なら、ファイル名を見るだけでマージ順序が分かります。

### 環境名の統一

環境名を統一することで、混乱を避けられます：

```
# Good
infra-dev.yaml
infra-staging.yaml
infra-prod.yaml

# Bad（一貫性がない）
dev.yaml
stg-infra.yaml
production-infrastructure.yaml
```

---

## 実践例: 3環境の管理

### ディレクトリ構造

```
my-project/
├── README.md
├── base.yaml               # 共通設定
├── app.yaml                # アプリケーション定義
├── environments/
│   ├── dev.yaml            # 開発環境
│   ├── staging.yaml        # ステージング環境
│   └── prod.yaml           # 本番環境
└── Makefile                # 便利スクリプト
```

### Makefile

```makefile
.PHONY: plan-dev plan-staging plan-prod

plan-dev:
	nwsctl plan -t base.yaml -t app.yaml -t environments/dev.yaml

plan-staging:
	nwsctl plan -t base.yaml -t app.yaml -t environments/staging.yaml

plan-prod:
	nwsctl plan -t base.yaml -t app.yaml -t environments/prod.yaml
```

### 使用方法

```bash
# 開発環境の計画を作成
make plan-dev

# 本番環境の計画を作成
make plan-prod
```

---

## 環境変数の活用

環境変数を使って動的に切り替えることもできます：

### スクリプト例

```bash
#!/bin/bash

ENV=${1:-dev}

case $ENV in
  dev)
    nwsctl plan -t base.yaml -t infra-dev.yaml
    ;;
  staging)
    nwsctl plan -t base.yaml -t infra-staging.yaml
    ;;
  prod)
    nwsctl plan -t base.yaml -t infra-prod.yaml
    ;;
  *)
    echo "Unknown environment: $ENV"
    exit 1
    ;;
esac
```

### 使用方法

```bash
# 開発環境
./deploy.sh dev

# 本番環境
./deploy.sh prod
```

---

## 注意点

### 1. マージの順序に注意

Stateファイルは指定された順にマージされます。後に指定したファイルが前のファイルの設定を上書きします。

```bash
# 正しい順序
nwsctl plan -t base.yaml -t env-specific.yaml

# 間違った順序（env-specificの設定がbaseに上書きされる）
nwsctl plan -t env-specific.yaml -t base.yaml
```

### 2. 識別子の一致

ノード、クラスタ、ジェネレーターの識別子は、すべてのファイルで一致している必要があります。

```yaml
# base.yaml
inventory:
  nodes:
    web-server: {}

# env-dev.yaml
inventory:
  nodes:
    web-server:  # 識別子が一致
      instanceSelector:
        generator: local-vms
```

### 3. インスタンス数の管理

環境ごとにノード数が異なる場合、インスタンス数も調整が必要です。

---

## 次のステップ

- [Stateマージの詳細]({{< relref "state-merging" >}}) - マージの仕様と動作を詳しく理解する

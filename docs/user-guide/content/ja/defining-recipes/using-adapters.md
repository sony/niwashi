---
title: "アダプターを利用する"
weight: 3
---

# アダプターを利用する

アダプターは、AnsibleやTerraformなどの外部ツールをNiwashiのレシピから呼び出すための仕組みです。このページでは、既存のアダプターをレシピから利用する方法を説明します。

アダプターの定義方法については [アダプターを定義する]({{< relref "defining-adapters" >}}) を参照してください。

---

## アダプターとは

アダプターは `kind: adapter` の特殊なレシピで、外部ツールのラッパーとなるコマンドを公開します。レシピ側はこのコマンドを `tool.run` アクションで呼び出します。

アダプターは `provides` でCapability名を公開します（例: `adapter.tool.ansible`）。レシピ側はこの名前を `spec.requires` で宣言し、`tool.run` の `toolRef` で参照します。

---

## アダプターへの依存を宣言する

アダプター自体は `spec.requires` に書きません。アダプターは `tool.run` の `toolRef` で参照するだけで自動的に解決されます。

`spec.requires` に書くのは、アダプターが依存するホストツール（`host.tool.*`）や他のレシピの Capability 名です。

```yaml
spec:
  requires:
    - host.tool.git   # ホストツールの依存は requires に書く
    # adapter.tool.ansible は書かない → tool.run.toolRef で参照する
```

---

## tool.run アクション

アダプターの機能を呼び出すには `tool.run` アクションを使います。

```yaml
action:
  tool.run:
    toolRef: adapter.tool.ansible   # アダプターのCapability名
    command: ansible-playbook        # アダプターが定義するコマンド名
    argvTpl:                         # コマンドに渡す引数（テンプレート変数使用可）
      - "-i"
      - "hosts.ini"
      - "playbook.yml"
    workdirTpl: "{{ .Paths.work_dir }}"  # 作業ディレクトリ（省略時はデフォルトの作業ディレクトリ）
    envTpl:                              # 追加の環境変数（省略可）
      MY_VAR: "{{ .Params.my_var }}"
    inheritEnv:                          # ホスト環境から継承する環境変数（省略可）
      - SSH_AUTH_SOCK
      - ANSIBLE_*
```

### フィールド

| フィールド | 必須 | 説明 |
|-----------|------|------|
| `toolRef` | Yes | アダプターのCapability名 |
| `command` | Yes | 呼び出すコマンド名（アダプターが定義するもの） |
| `argvTpl` | | コマンドへの引数リスト。テンプレート変数使用可 |
| `workdirTpl` | | 作業ディレクトリ。省略時はデフォルトの作業ディレクトリ |
| `envTpl` | | 追加で渡す環境変数。テンプレート変数使用可 |
| `inheritEnv` | | ホスト環境から継承する環境変数名のリスト。グロブパターン使用可 |

### 環境変数の継承 (`inheritEnv`)

`inheritEnv` を使うと、nwsctl を実行しているホストの環境変数をアダプターのタスクに引き渡せます。`SSH_AUTH_SOCK` や `ANSIBLE_*` のように、ホスト環境に依存するツールを呼び出す際に使用します。

```yaml
action:
  tool.run:
    toolRef: adapter.tool.python.venv
    command: run
    argvTpl:
      - "{{ .Paths.work_dir }}/venv"
      - ansible-playbook
      - site.yml
    inheritEnv:
      - SSH_AUTH_SOCK   # SSHエージェント転送
      - ANSIBLE_*       # Ansible関連の設定変数
```

**動作の詳細：**

- `PATH` は `inheritEnv` の指定に関わらず常に継承されます。
- `ANSIBLE_*` のようなグロブパターンが使用できます。
- アダプターの `exec.local` タスク自身が `inheritEnv` を持つ場合、両方がマージされます。
- `envTpl` で同名のキーが指定された場合は `envTpl` の値が優先されます。

> **注意：** `inheritEnv` はアダプターが `exec.local` で実装されている場合のみ有効です。`exec.remote` で実装されたアダプターには影響しません。

---

## リファレンスアダプター

| アダプター | Capability名 | 説明 |
|-----------|-------------|------|
| Ansible | `adapter.tool.ansible` | ansible-playbook / ansible-galaxy を実行する |

各アダプターの詳細は [レシピリファレンス]({{< relref "../recipes" >}}) を参照してください。

- [Ansibleアダプター]({{< relref "../recipes/ansible" >}})

---

## 次のステップ

- [アダプターを定義する]({{< relref "defining-adapters" >}}) - 独自のアダプターを作成する方法
- [Nodeのcapabilityを定義する]({{< relref "node-capability" >}}) - `tool.run` を使ったレシピの書き方の例

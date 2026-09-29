---
title: "タスクの定義"
weight: 9
---

# タスクの定義

すべての `kind` のレシピで共通のタスク定義フォーマットを説明します。

---

## task フィールド

| フィールド | 必須 | 説明 |
|-----------|------|------|
| `name` | Yes | タスク名（レシピ内で一意） |
| `action` | Yes | 実行するアクション（後述） |
| `operation` | | `construct`（デフォルト）/ `update` / `destruct` |
| `dependsOn` | | 依存タスク名のリスト |
| `where` | | ノードのフィルタ条件 |
| `stateChanges` | | タスク実行後の状態変更操作 |

---

## アクションの種類

### exec.remote — ノードでスクリプトを実行

ノードにリモート接続（SSH等）してスクリプトを実行します。接続情報はnwsctlが自動的に設定します。`assets` で指定したファイルも自動的にノードにコピーされます。

```yaml
action:
  exec.remote:
    scriptTpl: |
      #!/bin/bash
      set -euo pipefail
      VERSION="{{ .Params.version }}"
      echo "Installing version ${VERSION} on {{ .Target }}"
      cp "{{ .Assets }}/config/default.conf" /etc/myapp/
    envTpl:
      MY_VAR: "{{ .Params.my_var }}"   # 環境変数を追加で渡せる
```

### exec.local — ホストでスクリプトを実行

nwsctlを実行しているホスト上でスクリプトを実行します。ノードへの直接接続は行いません。

```yaml
action:
  exec.local:
    scriptTpl: |
      #!/bin/bash
      set -euo pipefail
      echo "Preparing for {{ .Target }}"
    envTpl:
      MY_VAR: "{{ .Params.my_var }}"   # 環境変数を追加で渡せる
    inheritEnv:
      - KUBECONFIG        # ホスト環境から特定の変数を継承
      - ANSIBLE_*         # グロブパターンも使用可能
```

**環境変数の継承 (`inheritEnv`)**

`exec.local` はデフォルトでは `PATH` のみをホスト環境から継承します。追加で継承したい変数は `inheritEnv` に明示的に列挙してください。

| フィールド | 説明 |
|-----------|------|
| `inheritEnv` | ホスト環境から継承する環境変数名（またはグロブパターン）のリスト。`PATH` は常に継承される。 |

- `ANSIBLE_*` のようなグロブパターンが使用できます。
- `envTpl` と `inheritEnv` の両方に同じキーが存在する場合、`envTpl` の値が優先されます。

### tool.run — アダプターを経由して実行

`spec.requires` で依存を宣言したアダプターを呼び出します。Ansibleアダプターを使ってPlaybookを実行する場合などに使用します。詳細は [アダプターを利用する]({{< relref "using-adapters" >}}) を参照してください。

```yaml
action:
  tool.run:
    toolRef: adapter.tool.ansible
    command: ansible-playbook
    argvTpl:
      - "-i"
      - "hosts.ini"
      - "playbook.yml"
    workdirTpl: "{{ .Paths.work_dir }}"
    inheritEnv:           # ホスト環境変数を継承（exec.localベースのアダプターのみ有効）
      - SSH_AUTH_SOCK
      - ANSIBLE_*
```

> **\[kind: adapter のみ\]** `kind: adapter` のタスクでは `exec.local` のみ使用できます。`exec.remote` と `tool.run` は使用できません。

---

## スクリプトの書き方（scriptTpl）

`scriptTpl` に書いたスクリプトは、テンプレートとしてレンダリングされたあとファイルに保存され、実行されます。

### shebangによるインタープリタの制御

スクリプトの1行目に shebang（`#!` で始まる行）を書くと、そのインタープリタでスクリプトが実行されます。shebangはOSのカーネルではなく**niwashi自身が解釈**し、`<インタープリタ> [引数] <スクリプトファイル>` の形式で起動します。そのため、スクリプトファイルに実行権限は不要です。

```yaml
action:
  exec.remote:
    scriptTpl: |
      #!/bin/bash
      set -euo pipefail
      echo "runs with bash"
```

shebangを省略した場合は、実行対象のOSに応じたデフォルトのインタープリタが使われます。

| 実行対象のOS | デフォルトのインタープリタ |
|-------------|--------------------------|
| Linux / macOS | `/bin/sh` |
| Windows | `powershell`（Windows PowerShell） |

「実行対象のOS」は、`exec.local` ではnwsctlを実行しているホストのOS、`exec.remote` では接続先ノードのOSです。

Windowsの既定インタープリタは、標準搭載のWindows PowerShell（`powershell`）です。別インストールが必要なPowerShell 7+（`pwsh`）を使う場合は、shebangで明示的に指定してください（例: `#!pwsh`）。

> **注意**: Windows上でPowerShellスクリプトを実行するには、実行するマシン（`exec.local` ではnwsctlを実行しているホスト、`exec.remote` ではWinRM接続先のノード）側で、スクリプトの実行が許可された実行ポリシーになっている必要があります。既定の実行ポリシー（`Restricted`）のままでは、署名されていないローカルスクリプトの実行がブロックされ、`UnauthorizedAccess` エラーになります。対象マシン上で事前に実行ポリシーを変更してください（例: `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`）。詳細は [about_Execution_Policies](https://go.microsoft.com/fwlink/?LinkID=135170) を参照してください。niwashi自身が実行ポリシーを変更することはありません。

### 推奨する書き方

- インタープリタを暗黙のデフォルトに任せず、**shebangを明示する**ことを推奨します
- Linux / macOS向けには `sh` または `bash` を推奨します。bash固有の機能を使う場合は `#!/bin/bash` を明示してください
- Windows向けにはPowerShellスクリプトとして書くことを推奨します。shebangを省略すると自動的に `powershell`（Windows PowerShell）で実行されます
- 途中のコマンドの失敗を検出するため、スクリプト冒頭に `set -euo pipefail`（bash）を入れることを推奨します
- OSごとに処理を分けたい場合は、1つのスクリプト内で分岐せず、`where` でタスクごとに実行対象を分けてください（[タスクのフィルタリング条件 (where)]({{< relref "task-where" >}})）

> **注意**: shebang行はLinuxカーネルの制限に合わせて127文字で切り詰められます。

レンダリング済みのスクリプトはタスクのログディレクトリ（`{{ .Paths.log_dir }}`）に保存されるため、実行後に実際のスクリプト内容を確認できます。

---

## テンプレート変数・環境変数

`scriptTpl`、`argvTpl`、`workdirTpl`、`envTpl` で使用できるテンプレート変数と、スクリプト実行時にセットされる環境変数の一覧は [レシピで使える変数]({{< relref "recipe-variables" >}}) を参照してください。

---

## タスクの依存関係

`dependsOn` で他のタスクへの依存を宣言すると、依存タスクが完了してから実行されます。

```yaml
tasks:
  - name: prepare
    action:
      exec.remote:
        scriptTpl: |
          mkdir -p /opt/myapp

  - name: install
    dependsOn: [prepare]
    action:
      exec.remote:
        scriptTpl: |
          cp "{{ .Assets }}/scripts/install.sh" /opt/myapp/
          /opt/myapp/install.sh

  - name: configure
    dependsOn: [install]
    action:
      exec.remote:
        scriptTpl: |
          /opt/myapp/configure.sh --version={{ .Params.version }}
```

---

## 更新処理（operation: update）

適用済みのCapabilityについて、目標状態の `params` が変わった場合や、解決されるレシピのバージョンが上がった場合、`nwsctl plan` はその変更を検出し、`operation: update` のタスクを実行対象にします。追加のフラグは不要です。通常の `nwsctl plan` / `nwsctl apply` で、追加と一緒に更新も扱われます。

```yaml
tasks:
  - name: install
    action:
      exec.remote:
        scriptTpl: |
          apt-get install -y myapp

  - name: reconfigure
    operation: update
    action:
      exec.remote:
        scriptTpl: |
          #!/bin/bash
          set -euo pipefail
          /opt/myapp/configure.sh --port={{ .Params.port }}
          systemctl restart myapp
```

### updateタスクが実行される条件

| 前回の適用からの変化 | 結果 |
|---|---|
| `params` が変わった（レシピのバージョンは同じ） | `operation: update` のタスクが実行される |
| レシピのバージョンが上がった（例: `1.0.0` → `1.1.0`） | `params` が同じでも、新しいバージョンの `operation: update` のタスクが実行される |
| レシピのバージョンが下がった | `nwsctl plan` がエラーで失敗する |
| 変化なし | 何も実行されない |

updateタスクが成功すると、新しい `params` とレシピのバージョンがStateに記録されます。Capabilityの `store/` は更新後も保持されます。そのため、updateタスクはconstructタスクが書き込んだ値を `{{ .Store }}` で参照できます。

### updateタスクを持たないレシピ

レシピに `operation: update` のタスクがない場合、変更は適用できません。`nwsctl plan` はそのレシピのタスクを計画に含めず、代わりに `Pending Updates` として一覧表示します。

```
Pending Updates (recipe has no update tasks):
- node:my-org/myapp@1.0.0:web-1
```

Stateには前回適用した `params` とバージョンが残るため、変更は保留のままになります。レシピに `operation: update` のタスクが追加されると、次回の `nwsctl plan` でその変更が計画に含まれます。Niwashiは、削除して再構築する方法へのフォールバックは行いません。

### 注意点

- 更新の検出対象は `kind: node` と `kind: cluster` のCapabilityだけです。
- `params` は全体として比較されます。updateタスクの `{{ .Params }}` には、`spec.defaults.params` とマージされた新しいparams全体が入ります。
- 比較されるのは、目標状態に書かれた `params`（プロファイルによる上書きを含む）だけです。レシピの `spec.defaults.params` を変更しても、`metadata.version` を上げない限り更新は発生しません。
- `requires` で更新対象のCapabilityに依存しているCapabilityは、再実行されません。

---

## 削除処理（operation: destruct）

`plan --destroy` 実行時には、`operation: destruct` のタスクが実行されます。インストールや設定の逆操作を記述します。

```yaml
tasks:
  - name: install
    operation: construct   # デフォルトなので省略可
    action:
      exec.remote:
        scriptTpl: |
          apt-get install -y myapp

  - name: uninstall
    operation: destruct
    action:
      exec.remote:
        scriptTpl: |
          apt-get remove -y myapp
          rm -rf /etc/myapp
```

`construct` と `destruct` は別々のタスクとして定義します。

---

## タスクが失敗した場合

タスクが失敗すると、`nwsctl apply` は新しいジョブの開始を止め、実行中のタスクの完了を待ってからエラーで終了します。Stateがどうなるかはoperationによって異なります。

| operation | 失敗後のState | 次回の実行 |
|---|---|---|
| `construct` | Capabilityは `store/` も含めてStateから削除される | 次回の `nwsctl plan` / `nwsctl apply` で再度constructされる |
| `update` | 前回適用した `params` とバージョンが残る | 次回の `nwsctl plan` / `nwsctl apply` で再度updateが実行される |
| `destruct` | CapabilityはStateに残る | 次回の `--destroy` / `--prune` で再度destructされる |

この表は `kind: node` と `kind: cluster` のCapabilityについての説明です。`kind: infra` については [Infraのプロビジョナーを定義する]({{< relref "infrastructure-provisioning#プロビジョニングが失敗した場合" >}}) を参照してください。

失敗したconstructは次回の実行で最初からやり直されます。そのため、途中まで完了した状態から始まっても収束するようにconstructタスクを書いてください。たとえば、パッケージがインストール済みか、ファイルが既に存在するかを確認してから作成します。

---

## フィルタリング条件（where）

タスクに `where` フィールドを指定すると、条件に一致したノードに対してのみタスクを実行します。詳細は [タスクのフィルタリング条件 (where)]({{< relref "task-where" >}}) を参照してください。

---

## stateChanges

タスク実行後にStateを更新する操作を定義します。詳細は [stateChanges で State を更新する]({{< relref "state-changes" >}}) を参照してください。

---

## 次のステップ

- [stateChanges で State を更新する]({{< relref "state-changes" >}}) — `stateChanges` の詳細
- [タスクのフィルタリング条件 (where)]({{< relref "task-where" >}}) — `where` フィールドの詳細
- [レシピで使える変数]({{< relref "recipe-variables" >}}) — テンプレート変数・環境変数のリファレンス

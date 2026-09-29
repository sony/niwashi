---
title: "Ansibleアダプター"
weight: 1
---

# Ansibleアダプター（`nws/ansible.adapter`）

`nws/ansible.adapter` はAnsible用のリファレンスアダプターです。Ansibleをホストにインストール済みであることを前提とし、レシピから `tool.run` アクションで呼び出せます。

---

## 提供するコマンド

| コマンド | 説明 |
|---------|------|
| `gen-inventory` | ノードの接続情報からAnsibleインベントリを生成する |
| `ansible-playbook` | `ansible-playbook` を実行する |
| `ansible-galaxy` | `ansible-galaxy` を実行する |

---

## gen-inventory コマンド

`gen-inventory` は、Niwashiが管理するノードの接続情報をAnsibleが読み込めるインベントリファイルに変換するコマンドです。`ansible-playbook` を実行する前に必ずこのコマンドを実行してください。

`workdirTpl` で指定したディレクトリに以下のファイルを生成します：

- `hosts.ini` — Ansibleインベントリファイル（`[all]` セクションと、`params.groups` で定義したグループセクションを含む）
- `ssh_config` — SSH接続設定ファイル（ノードのホスト名・ポート・ユーザー・秘密鍵のパスを含む）

生成されたインベントリでは `ansible_ssh_common_args='-F ./ssh_config'` が設定されるため、`ansible-playbook` に `-i hosts.ini` を指定するだけでNiwashiが管理するノードに接続できます。

```yaml
- name: gen-inventory
  action:
    tool.run:
      toolRef: adapter.tool.ansible
      command: gen-inventory
      workdirTpl: "{{ .Paths.work_dir }}"
```

### groupsパラメータ

`gen-inventory` が生成するグループセクションは `params.groups` で制御します。

```yaml
spec:
  defaults:
    params:
      groups:
        kube_control_plane:
          - master-node-01
        kube_node:
          - worker-node-01
          - worker-node-02
```

この設定により `hosts.ini` に `[kube_control_plane]` と `[kube_node]` セクションが生成されます。

---

## ansible-playbook コマンド

`ansible-playbook` コマンドを実行します。`argvTpl` で引数を渡します。

```yaml
- name: run-playbook
  dependsOn: [gen-inventory]
  action:
    tool.run:
      toolRef: adapter.tool.ansible
      command: ansible-playbook
      workdirTpl: "{{ .Paths.work_dir }}"
      argvTpl:
        - -i
        - hosts.ini
        - playbook.yml
        - -e
        - "nginx_port={{ .Params.nginx_port }}"
```

`gen-inventory` と `workdirTpl` を同じディレクトリに設定することで、生成された `hosts.ini` を参照できます。

---

## ansible-galaxy コマンド

`ansible-galaxy` コマンドを実行します。Roleのインストールなどに使用します。

```yaml
- name: install-roles
  action:
    tool.run:
      toolRef: adapter.tool.ansible
      command: ansible-galaxy
      workdirTpl: "{{ .Paths.work_dir }}"
      argvTpl:
        - install
        - -r
        - requirements.yml
```

---

## 実例：Kubesprayを使ったKubernetes構築

以下はKubesprayを使ってKubernetesクラスターを構築するレシピの例です（`recipe/kubernetes/kubespray/nws-recipe.yaml` を参照）。

```yaml
version: nws.recipe/v1
kind: cluster
metadata:
  id: nws/kubernetes.by.kubespray
  version: "0.0.1"
  description: "Install Kubernetes using Kubespray"
spec:
  provides:
    - name: cluster.kubernetes
      attrs: { by: kubespray }

  requires:
    - adapter.tool.ansible
    - host.tool.git

  workspace:
    mode: persistent     # Kubesprayのクローンを保持する

  defaults:
    params:
      kubespray_url: https://github.com/kubernetes-sigs/kubespray.git
      version: v2.27.1
      become_user: root
      groups:
        kube_control_plane: []
        kube_node: []
        etcd: []

  tasks:
    - name: ensure-dirs
      action:
        exec.local:
          scriptTpl: |
            #!/bin/bash
            set -euo pipefail

            if [ ! -d kubespray ]; then
              git clone {{ .Params.kubespray_url }} kubespray
            fi

            pushd kubespray
            git fetch --all
            git checkout refs/tags/{{ .Params.version }}
            popd

    - name: gen-inventory
      dependsOn: [ensure-dirs]
      action:
        tool.run:
          toolRef: adapter.tool.ansible
          command: gen-inventory
          workdirTpl: "{{ .Paths.work_dir }}/kubespray"

    - name: run-kubespray
      dependsOn: [gen-inventory]
      action:
        tool.run:
          toolRef: adapter.tool.ansible
          command: ansible-playbook
          workdirTpl: "{{ .Paths.work_dir }}/kubespray"
          argvTpl:
            - -i
            - hosts.ini
            - --become-user={{ .Params.become_user }}
            - cluster.yml
```

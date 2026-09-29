---
title: "制約"
weight: 5
---

# 制約

## nwsctlの制約

- 動作環境
  - Linux(x86_64)
  - <s>Windows(x86_64)</s>
  - <s>Mac(x86_64)</s>
- Windows, Macは動作未確認
  - レシピがWindows, Macでの動作に対応しておらず、正しく動作しない可能性があります

## リファレンスレシピの制約

### Ansible

- ホストにAnsibleをインストールしておく必要があります(*1)
  - 最新バージョンを推奨

### Vagrant

- ホストにVagrantをインストールしておく必要があります(*1)

### Kubernetes

- Ansibleのバージョンに依存するが、nwsctlでの自動解決には対応していません

## 将来の計画

- (*1)
  - 必要なツール（AnsibleやVagrant）は、レシピの処理でインストールすることを検討中です
  - AnsibleはPythonのvenvかコンテナを検討しています
  - Vagrantはコンテナを検討しています

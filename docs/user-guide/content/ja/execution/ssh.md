---
title: "nwsctl ssh"
weight: 4
---

# nwsctl ssh

管理対象のノードにSSH接続します。

```
nwsctl ssh NODE_NAME [flags]
```

---

## 概要

`nwsctl ssh` は、ワークスペースの `state.json` から接続情報（ホスト名、ポート、ユーザー、秘密鍵）を自動的に取得してSSH接続します。IPアドレスや鍵のパスを手動で調べる必要がありません。

対話的なシェルを起動します。終了するには `exit` を実行するかターミナルを閉じてください。

---

## フラグ

| フラグ | デフォルト | 説明 |
|-------|----------|------|
| `--work-dir` | `.niwashi` | ワークスペースのパス |

---

## 使い方

```bash
nwsctl ssh web-server-01
```

ノード名はStateファイルの `inventory.nodes` のキー名と一致している必要があります。

```bash
# ワークスペースの場所を指定する場合
nwsctl ssh web-server-01 --work-dir /path/to/workspace
```

---

## 注意事項

- ノードの接続情報がStateに存在しない場合（インフラがまだ適用されていない場合など）はエラーになります
- 接続タイプが SSH でないノードには接続できません

---

## 次のステップ

- [nwsctl export]({{< relref "export" >}}) — 現在のStateをファイルに書き出す

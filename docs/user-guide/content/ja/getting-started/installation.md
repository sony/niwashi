---
title: "インストール"
weight: 3
---

# インストール

## 動作環境

| OS | アーキテクチャ | 状態 |
|----|----------------|------|
| Linux | x86_64 | 動作確認済み |
| Windows | x86_64 | 未確認（リファレンスレシピ非対応） |
| macOS | x86_64 | 未確認（リファレンスレシピ非対応） |

> **注意**: Windows・macOS でも nwsctl 自体は起動できますが、リファレンスレシピが bash を前提としているため、正しく動作しない場合があります。詳細は [制約]({{< relref "limitations" >}}) を参照してください。

---

## nwsctl のインストール

### 1. アーカイブを入手する

[リリースページ](https://github.com/sony/niwashi/releases) から対象アーキテクチャのアーカイブをダウンロードしてください。

| OS | ファイル名 |
|----|-----------|
| Linux | `nwsctl_Linux_x86_64.tar.gz` |
| Windows | `nwsctl_Windows_x86_64.zip` |
| macOS | `nwsctl_Darwin_x86_64.tar.gz` |

### 2. アーカイブを展開する

**Linux / macOS**

```bash
tar -xzf nwsctl_Linux_x86_64.tar.gz
```

展開すると `nwsctl` バイナリが取り出せます。

**Windows**

エクスプローラーなどで `nwsctl_Windows_x86_64.zip` を展開してください。`nwsctl.exe` が取り出せます。

### 3. バイナリを配置する

**Linux / macOS**

パスが通っているディレクトリ（例: `/usr/local/bin`）にバイナリを配置します。

```bash
sudo mv nwsctl /usr/local/bin/
```

ユーザーのホームディレクトリ配下に置く場合は `~/.local/bin/` など、`$PATH` に含まれるディレクトリを選んでください。

**Windows**

任意のディレクトリに `nwsctl.exe` を配置し、そのディレクトリを環境変数 `PATH` に追加してください。

### 4. インストールを確認する

```bash
nwsctl version
```

バージョン番号が表示されればインストール完了です。

---

## 次のステップ

- [ワークフロー]({{< relref "workflow" >}}) - nwsctl を使った基本的な作業の流れ

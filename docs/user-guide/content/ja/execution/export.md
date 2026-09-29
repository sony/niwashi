---
title: "nwsctl export"
weight: 5
---

# nwsctl export

ワークスペースの現在の状態をYAMLファイルに書き出します。

```
nwsctl export [flags]
```

---

## 概要

`nwsctl export` は、ワークスペースの `state.json`（Niwashiが内部管理するJSON形式）をYAML形式に変換して出力します。

出力したYAMLは `nwsctl init -s` に渡すことで、別のワークスペースの初期状態として利用できます。

---

## フラグ

| フラグ | 省略形 | デフォルト | 説明 |
|-------|-------|----------|------|
| `--out` | `-o` | （標準出力） | 書き出し先のYAMLファイルパス |
| `--include-runtime` | | `false` | runtimeセクションを含めて書き出す |
| `--work-dir` | | `.niwashi` | ワークスペースのパス |

---

## 使い方

### 標準出力に書き出す

```bash
nwsctl export
```

### ファイルに書き出す

```bash
nwsctl export -o current.yaml
```

### 環境の引き継ぎ

別のマシンや別のワークスペースで同じ状態から始める場合：

```bash
# 現在の状態をファイルに書き出す
nwsctl export -o current.yaml

# 別のワークスペースにその状態を引き継ぐ
nwsctl init -s current.yaml --work-dir /path/to/new-workspace
```

### runtimeセクションを含める

```bash
nwsctl export --include-runtime -o current-full.yaml
```

デフォルトではruntimeセクション（ツールのパスなどNiwashiが収集した実行時情報）は除外されます。`--include-runtime` を付けると含めることができます。通常は除外したものを `nwsctl init -s` に渡してください。

---

## 注意事項

エクスポートしたStateを別のワークスペースで `nwsctl init -s` することは可能ですが、レシピによっては正しく動作しない場合があります。

**例**: Vagrantレシピは、VMのイメージなどをワークスペース配下に作成します。これらのファイルは `export` に含まれないため、インポート後に再度プロビジョニングが必要になることがあります。

---

## 次のステップ

- [nwsctl init]({{< relref "init" >}}) — 書き出したStateを使って新しいワークスペースを初期化する

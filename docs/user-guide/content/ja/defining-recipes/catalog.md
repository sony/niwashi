---
title: "複数のレシピをまとめる（nws-catalog.yaml）"
weight: 5
---

# 複数のレシピをまとめる（nws-catalog.yaml）

1つのディレクトリで複数のレシピをまとめて提供したい場合に、`nws-catalog.yaml` を使います。

---

## nws-catalog.yaml とは

通常、レシピは1ファイル1レシピ（`nws-recipe.yaml`）として配置します。しかし、関連するレシピをまとめてディレクトリ単位で管理したい場合は、`nws-catalog.yaml` を使って複数のレシピをまとめて提供できます。

```
my-recipes/
├── nws-catalog.yaml   ← カタログファイル
├── feature-a.yaml     ← レシピ本体
└── feature-b.yaml     ← レシピ本体
```

`nwsctl` は `nws-catalog.yaml` を見つけると、そこに列挙されたレシピファイルをまとめてロードします。

---

## フォーマット

```yaml
version: nws.catalog/v1
recipes:
  <ラベル>:
    path: <レシピファイルへの相対パス>
  <ラベル>:
    path: <レシピファイルへの相対パス>
```

| フィールド | 説明 |
|-----------|------|
| `version` | `nws.catalog/v1` 固定 |
| `recipes` | レシピのマップ |
| `recipes.<ラベル>` | 任意の識別ラベル（ドキュメント用途。読み込み時には使用されない） |
| `recipes.<ラベル>.path` | `nws-catalog.yaml` からの相対パスでレシピファイルを指定 |

---

## 実例：ansible レシピ

ansible のリファレンスレシピは `nws-catalog.yaml` を使って2つのレシピを提供しています：

```
recipe/ansible/
├── nws-catalog.yaml
├── ansible-adapter.yaml   ← アダプターレシピ
└── ansible-installer.yaml ← ツールインストーラレシピ
```

```yaml
# recipe/ansible/nws-catalog.yaml
version: nws.catalog/v1
recipes:
  adapter:
    path: ./ansible-adapter.yaml
  tool:
    path: ./ansible-installer.yaml
```

`adapter` や `tool` はラベルとして任意に付けた名前です。読み込み時は `path` で指定されたファイルが実際に使用されます。

---

## nws-recipe.yaml との使い分け

| | `nws-recipe.yaml` | `nws-catalog.yaml` |
|--|---|---|
| 用途 | 1ディレクトリに1つのレシピを提供 | 1ディレクトリに複数のレシピをまとめて提供 |
| 構成 | ファイル単体で完結 | カタログ + 複数のレシピファイル |
| 向いているケース | 独立したシンプルなレシピ | 関連する複数レシピをセットで配布したい場合 |

---

## 注意事項

- `nws-catalog.yaml` と `nws-recipe.yaml` を同じディレクトリに共存させることも可能ですが、混乱を避けるためどちらかに統一することを推奨します
- `path` は `nws-catalog.yaml` が置かれているディレクトリからの相対パスで指定します
- カタログから参照するファイルは `nws-recipe.yaml` という名前でなくても構いません（任意の名前が使えます）

---
title: "nwsctl コマンドリファレンス"
weight: 4
---

# nwsctl コマンドリファレンス

`nwsctl` はNiwashiのメインインターフェースです。インフラの計画・実行・管理をコマンドラインから行います。

---

## 基本的な使い方

Niwashiの典型的なワークフローについては [ワークフロー]({{< relref "workflow" >}}) を参照してください。

---

## 目標の状態を定義する

`nwsctl plan` に渡す目標状態（State）の書き方は [目標の状態を定義する]({{< relref "defining-desired-state" >}}) を参照してください。

---

## コマンド一覧

| コマンド | 説明 |
|---------|------|
| [`nwsctl init`](init/) | ワークスペースを初期化する |
| [`nwsctl plan`](plan/) | 実行計画を作成する |
| [`nwsctl apply`](apply/) | 実行計画を適用する |
| [`nwsctl ssh`](ssh/) | 管理ノードにSSH接続する |
| [`nwsctl export`](export/) | 現在のStateをファイルに書き出す |

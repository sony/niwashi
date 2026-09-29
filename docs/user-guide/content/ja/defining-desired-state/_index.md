---
title: "状態を定義する"
weight: 2
---

# 状態を定義する

Niwashiで目標の状態（Desired State）を定義する方法を学びます。

## 学習パス

### 初めての方

まずは概要から読み始めることをお勧めします：

1. **[状態定義の概要]({{< relref "overview" >}})** - Stateの基本概念と最小限の例
2. **[基本概念](basic-concepts/)** - State の基本構造と各要素の役割
   - [State の基本構造]({{< relref "basic-concepts/state-structure" >}})
   - [Inventory の基本]({{< relref "basic-concepts/inventory" >}})
   - [Infrastructure の基本]({{< relref "basic-concepts/infrastructure" >}})
   - [識別子の命名規則]({{< relref "basic-concepts/naming-rules" >}})

### 詳しく学びたい方

各要素の完全なリファレンス：

- **[Capabilityとレシピの詳細]({{< relref "capabilities" >}})** - レシピの指定方法
- **[ノードの詳細]({{< relref "nodes" >}})** - ノードの全属性とオプション
- **[クラスタの詳細]({{< relref "clusters" >}})** - クラスタの全属性とオプション
- **[ジェネレーターの詳細]({{< relref "generators" >}})** - インスタンス生成器の詳細
- **[テンプレートの詳細]({{< relref "templates" >}})** - テンプレート機能の全て

### 高度な使い方

実践的なパターンと高度な機能：

- **[高度な使い方](advanced/)** - 高度なパターンと使い方
  - [複数環境の管理]({{< relref "advanced/multi-environment" >}}) - dev/prod環境の切り替え
  - [Stateマージの詳細]({{< relref "advanced/state-merging" >}}) - マージの仕様

### 実践例

動作する完全な例：

- **[Kubernetesクラスタの構築]({{< relref "../examples/kubernetes-cluster" >}})** - K8sクラスタ構築の完全な例

---

## ドキュメント一覧

### 概要と基本
- [状態定義の概要]({{< relref "overview" >}})
- [基本概念](basic-concepts/)

### 詳細リファレンス
- [Capabilityとレシピの詳細]({{< relref "capabilities" >}})
- [ノードの詳細]({{< relref "nodes" >}})
- [クラスタの詳細]({{< relref "clusters" >}})
- [ジェネレーターの詳細]({{< relref "generators" >}})
- [テンプレートの詳細]({{< relref "templates" >}})

### 高度な使い方
- [高度な使い方](advanced/)

### 実践例
- [Kubernetesクラスタの構築]({{< relref "../examples/kubernetes-cluster" >}})

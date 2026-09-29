---
title: "Niwashi ユーザーガイド"
weight: 1
---

# Niwashi ユーザーガイド

Niwashiは、インフラストラクチャのプロビジョニングと設定管理を統合的に行うためのツールです。現在の状態から目標の状態への実行計画の作成と実行を行います。

## 目次

- [はじめに](getting-started/)
  - [Niwashiでできること](getting-started/overview/)
  - [ワークフロー](getting-started/workflow/)
  - [アーキテクチャ](getting-started/architecture/)
- [目標の状態を定義する](defining-desired-state/)
  - [状態定義の概要](defining-desired-state/overview/)
  - [基本概念](defining-desired-state/basic-concepts/)
    - [State の基本構造](defining-desired-state/basic-concepts/state-structure/)
    - [Inventory の基本](defining-desired-state/basic-concepts/inventory/)
    - [Infrastructure の基本](defining-desired-state/basic-concepts/infrastructure/)
    - [識別子の命名規則](defining-desired-state/basic-concepts/naming-rules/)
  - [Capabilityとレシピの詳細](defining-desired-state/capabilities/)
  - [ノードの詳細](defining-desired-state/nodes/)
  - [クラスタの詳細](defining-desired-state/clusters/)
  - [ジェネレーターの詳細](defining-desired-state/generators/)
  - [テンプレートの詳細](defining-desired-state/templates/)
  - [高度な使い方](defining-desired-state/advanced/)
    - [複数環境の管理](defining-desired-state/advanced/multi-environment/)
    - [Stateマージの詳細](defining-desired-state/advanced/state-merging/)
- [nwsctl コマンドリファレンス](execution/)
  - [nwsctl init](execution/init/)
  - [nwsctl plan](execution/plan/)
  - [nwsctl apply](execution/apply/)
  - [nwsctl ssh](execution/ssh/)
  - [nwsctl export](execution/export/)
- [レシピを定義する](defining-recipes/)
  - [Nodeのcapabilityを定義する](defining-recipes/node-capability/)
  - [Infraのプロビジョナーを定義する](defining-recipes/infrastructure-provisioning/)
  - [アダプターを利用する](defining-recipes/using-adapters/)
  - [アダプターを定義する](defining-recipes/defining-adapters/)
  - [stateChanges で State を更新する](defining-recipes/state-changes/)
  - [複数のレシピをまとめる](defining-recipes/catalog/)
  - [レシピの読み込み仕様](defining-recipes/recipe-loading/)
  - [Clusterのcapabilityを定義する](defining-recipes/cluster-capability/) 🚧
  - [Hostの設定を定義する](defining-recipes/host-configuration/) 🚧
- [サンプル](examples/)
  - [Kubernetesクラスタの構築](examples/kubernetes-cluster/)
- [レシピ](recipes/)
  - [Ansible](recipes/ansible/)
  - [Kubernetes](recipes/kubernetes/)
  - [Vagrant](recipes/vagrant/)

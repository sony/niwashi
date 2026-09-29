// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"fmt"

	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/patch"
)

type Inventory struct {
	Nodes    map[string]*node.Node       `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	Clusters map[string]*cluster.Cluster `json:"clusters,omitempty" yaml:"clusters,omitempty"`
}

func NewInventory() *Inventory {
	return &Inventory{
		Nodes:    make(map[string]*node.Node),
		Clusters: make(map[string]*cluster.Cluster),
	}
}

func (i *Inventory) SetDefaults() {
	if i.Nodes == nil {
		i.Nodes = make(map[string]*node.Node)
	}
	for nodeName, n := range i.Nodes {
		if n == nil {
			i.Nodes[nodeName] = node.NewNode()
		} else {
			n.SetDefaults()
		}
	}
	if i.Clusters == nil {
		i.Clusters = make(map[string]*cluster.Cluster)
	}
	for clusterName, cls := range i.Clusters {
		if cls == nil {
			i.Clusters[clusterName] = cluster.NewCluster()
		} else {
			cls.SetDefaults()
		}
	}
}

func (i *Inventory) Clone() *Inventory {
	if i == nil {
		return nil
	}

	clone := &Inventory{}
	clone.Nodes = make(map[string]*node.Node, len(i.Nodes))
	for k, v := range i.Nodes {
		clone.Nodes[k] = v.Clone()
	}
	clone.Clusters = make(map[string]*cluster.Cluster, len(i.Clusters))
	for k, v := range i.Clusters {
		clone.Clusters[k] = v.Clone()
	}

	return clone
}

func (i *Inventory) Merge(other *Inventory) (*Inventory, error) {
	if i == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return i.Clone(), nil
	}

	clone := &Inventory{}
	var err error
	clone.Nodes, err = merge.StringMapT(i.Nodes, other.Nodes)
	if err != nil {
		return nil, err
	}
	clone.Clusters, err = merge.StringMapT(i.Clusters, other.Clusters)
	if err != nil {
		return nil, err
	}
	return clone, nil
}

func (i *Inventory) Add(path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "nodes":
		return patch.AddCompositeToMap(
			path[1:],
			value,
			i.Nodes)
	case "clusters":
		return patch.AddCompositeToMap(
			path[1:],
			value,
			i.Clusters)
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (i *Inventory) Remove(path []string) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "nodes":
		return patch.RemoveCompositeMap(
			path[1:],
			i.Nodes)
	case "clusters":
		return patch.RemoveCompositeMap(
			path[1:],
			i.Clusters)
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (*Inventory) Update(o patch.Patchable) error {
	return fmt.Errorf("not supported update for Inventory")
}

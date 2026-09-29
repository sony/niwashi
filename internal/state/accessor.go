// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/node"
)

type Accessor struct {
	*State
}

func NewAccessor(s *State) *Accessor {
	return &Accessor{
		State: s,
	}
}

func (s *Accessor) GetNode(nodeName string) *node.Accessor {
	if s == nil || s.State == nil || s.Inventory == nil {
		return nil
	}

	return node.NewAccessor(nodeName, s.Inventory.Nodes[nodeName])
}

func (s *Accessor) SetNode(nodeName string, n *node.Node) *node.Accessor {
	if s == nil || s.State == nil || s.Inventory == nil {
		return nil
	}

	s.Inventory.Nodes[nodeName] = n
	return node.NewAccessor(nodeName, n)
}

func (s *Accessor) GetBoundInstance(nodeName string) *instance.InstanceAccessor {
	if s == nil || s.State == nil || s.Inventory == nil {
		return nil
	}

	n := s.GetNode(nodeName)
	if n == nil || n.InstanceRef == "" {
		return nil
	}
	genName, instName := instance.ParseInstanceRef(n.InstanceRef)
	return s.GetGenerator(genName).GetInstance(instName)
}

func (s *Accessor) GetCluster(clusterName string) *cluster.Accessor {
	if s == nil || s.State == nil || s.Inventory == nil {
		return nil
	}

	return cluster.NewAccessor(clusterName, s.Inventory.Clusters[clusterName])
}

func (s *Accessor) SetCluster(clusterName string, c *cluster.Cluster) *cluster.Accessor {
	if s == nil || s.State == nil || s.Inventory == nil {
		return nil
	}

	s.Inventory.Clusters[clusterName] = c
	return cluster.NewAccessor(clusterName, c)
}

func (s *Accessor) GetGenerator(generatorName string) *instance.GeneratorAccessor {
	if s == nil || s.State == nil || s.Infrastructure == nil {
		return nil
	}

	return instance.NewGeneratorAccessor(generatorName, s.Infrastructure.Generators[generatorName])
}

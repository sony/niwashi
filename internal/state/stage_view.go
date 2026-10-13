// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"fmt"

	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/transport"
)

type StageView struct {
	Inventory      *Inventory
	Infrastructure *Infrastructure
	Runtime        *Runtime
}

func NewEmptyStageView(r *Runtime) *StageView {
	return &StageView{
		Inventory:      NewInventory(),
		Infrastructure: &Infrastructure{},
		Runtime:        r,
	}
}

func (s *StageView) SetDefaults() {
	if s.Infrastructure == nil {
		s.Infrastructure = &Infrastructure{}
	}
	s.Infrastructure.SetDefaults()

	if s.Inventory == nil {
		s.Inventory = NewInventory()
	}
	s.Inventory.SetDefaults()

	if s.Runtime == nil {
		s.Runtime = &Runtime{}
	}
	s.Runtime.SetDefaults()
}

func (s *StageView) Add(path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "inventory":
		return patch.AddValue(
			path[1:],
			value,
			func() (*Inventory, bool) {
				return s.Inventory, s.Inventory != nil
			},
			func(v *Inventory) {
				s.Inventory = v
			})
	case "infrastructure":
		return patch.AddValue(
			path[1:],
			value,
			func() (*Infrastructure, bool) {
				return s.Infrastructure, s.Infrastructure != nil
			},
			func(v *Infrastructure) {
				s.Infrastructure = v
			})
	case "runtime":
		return patch.AddValue(
			path[1:],
			value,
			func() (*Runtime, bool) {
				return s.Runtime, s.Runtime != nil
			},
			func(v *Runtime) {
				s.Runtime = v
			})
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (s *StageView) Remove(path []string) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "inventory":
		return patch.RemoveComposite(
			path[1:],
			func() (*Inventory, bool) {
				return s.Inventory, s.Inventory != nil
			},
			func() {
				s.Inventory = nil
			})
	case "infrastructure":
		return patch.RemoveComposite(
			path[1:],
			func() (*Infrastructure, bool) {
				return s.Infrastructure, s.Infrastructure != nil
			},
			func() {
				s.Infrastructure = nil
			})
	case "runtime":
		return patch.RemoveComposite(
			path[1:],
			func() (*Runtime, bool) {
				return s.Runtime, s.Runtime != nil
			},
			func() {
				s.Runtime = nil
			})
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (s *StageView) Update(o patch.Patchable) error {
	return fmt.Errorf("not supported update for StageView")
}

func (s *StageView) GetNode(nodeName string) *node.Accessor {
	if s == nil || s.Inventory == nil {
		return nil
	}

	return node.NewAccessor(nodeName, s.Inventory.Nodes[nodeName])
}

func (s *StageView) GetBoundInstance(nodeName string) *instance.InstanceAccessor {
	if s == nil || s.Inventory == nil {
		return nil
	}

	n := s.GetNode(nodeName)
	if n == nil || n.InstanceRef == "" {
		return nil
	}
	genName, instName := instance.ParseInstanceRef(n.InstanceRef)
	return s.GetGenerator(genName).GetInstance(instName)
}

func (s *StageView) GetCluster(clusterName string) *cluster.Accessor {
	if s == nil || s.Inventory == nil {
		return nil
	}

	return cluster.NewAccessor(clusterName, s.Inventory.Clusters[clusterName])
}

func (s *StageView) GetGenerator(generatorName string) *instance.GeneratorAccessor {
	if s == nil || s.Infrastructure == nil {
		return nil
	}

	return instance.NewGeneratorAccessor(generatorName, s.Infrastructure.Generators[generatorName])
}

func (s *StageView) GetConnection(nodeId string) *transport.Connection {
	if s == nil || s.Inventory == nil || s.Inventory.Nodes == nil {
		return nil
	}

	node := s.GetNode(nodeId)
	if node == nil {
		return nil
	}

	// look for instance connection
	if node.InstanceRef == "" {
		return nil
	}

	genName, instName := instance.ParseInstanceRef(node.InstanceRef)
	instance := s.GetGenerator(genName).GetInstance(instName)
	if instance == nil {
		return nil
	}
	c := instance.Connection.Clone()
	if c != nil {
		c.SetDefaults()
	}
	return c
}

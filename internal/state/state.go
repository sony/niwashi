// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"fmt"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/profile"
	"github.com/sony/niwashi/internal/transport"
	"github.com/sony/niwashi/internal/types"
)

type State struct {
	Version        string          `json:"version" yaml:"version"`
	Metadata       types.Dict      `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Template       *Template       `json:"template,omitempty" yaml:"template,omitempty"`
	Inventory      *Inventory      `json:"inventory,omitempty" yaml:"inventory,omitempty"`
	Infrastructure *Infrastructure `json:"infrastructure,omitempty" yaml:"infrastructure,omitempty"`
	Runtime        *Runtime        `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	Status         *Status         `json:"status,omitempty" yaml:"status,omitempty"`
	Hash           string          `json:"-" yaml:"-"`
}

func NewState() *State {
	state := &State{
		Version: "nws.state/v1",
	}

	state.SetDefaults()

	return state
}

func (s *State) Clone() *State {
	if s == nil {
		return nil
	}

	clone := &State{}
	clone.Version = s.Version
	clone.Metadata = s.Metadata.Clone()
	clone.Template = s.Template.Clone()
	clone.Inventory = s.Inventory.Clone()
	clone.Infrastructure = s.Infrastructure.Clone()
	clone.Runtime = s.Runtime.Clone()
	clone.Status = s.Status.Clone()

	return clone
}

func (s *State) SetDefaults() {

	if s.Template == nil {
		s.Template = s.Template.SetDefaults()
	}

	// fill minimal required fields
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

	if s.Status == nil {
		s.Status = &Status{}
	}
	s.Status.SetDefaults()
}

func (s *State) Merge(other *State) (*State, error) {

	if s == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return s.Clone(), nil
	}

	// Merge only the following fields:
	// - Template
	// - Inventory
	// - Infrastructure
	clone := &State{
		Version:  s.Version,
		Metadata: s.Metadata.Clone(),
		Runtime:  s.Runtime.Clone(),
		Status:   s.Status.Clone(),
	}

	var err error
	clone.Template, err = s.Template.Merge(other.Template)
	if err != nil {
		return nil, err
	}
	clone.Inventory, err = s.Inventory.Merge(other.Inventory)
	if err != nil {
		return nil, err
	}
	clone.Infrastructure, err = s.Infrastructure.Merge(other.Infrastructure)
	if err != nil {
		return nil, err
	}

	return clone, nil
}

func (s *State) ApplyTemplates() error {
	if s == nil {
		return nil
	}

	if len(s.Template.Node) != 0 {
		for _, n := range s.Inventory.Nodes {
			err := n.ApplyTemplate(s.Template.Node)
			if err != nil {
				return err
			}
		}
	}

	if len(s.Template.Cluster) != 0 {
		for _, cluster := range s.Inventory.Clusters {
			err := cluster.ApplyTemplate(s.Template.Cluster)
			if err != nil {
				return err
			}
		}
	}

	if len(s.Template.Generator) != 0 {
		for _, inst := range s.Infrastructure.Generators {
			err := inst.ApplyTemplate(s.Template.Generator)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *State) Normalize(f cmap.RecipeFinder) error {
	if s == nil {
		return nil
	}

	if f == nil {
		return fmt.Errorf("RecipeFinder is required for normalization")
	}

	// Normalize for each capabilities in nodes and clusters
	var err error
	var caps cap.CapabilityList
	if s.Inventory != nil {
		for _, node := range s.Inventory.Nodes {
			caps, err = node.Capabilities.Normalize(f)
			if err != nil {
				return err
			}
			node.Capabilities = caps
		}
		for _, cluster := range s.Inventory.Clusters {
			caps, err = cluster.Capabilities.Normalize(f)
			if err != nil {
				return err
			}
			cluster.Capabilities = caps
		}
	}

	// Normalize provisioner in instances
	for _, inst := range s.Infrastructure.Generators {
		if inst.Provisioner != "" {
			recipe, err := f.FindRecipe(inst.Provisioner)
			if err != nil {
				return err
			}
			if inst.Provisioner != recipe.Fqid() {
				logger.Info("UpdateRecipeAlias", "from", inst.Provisioner, "to", recipe.Fqid())
			}
			inst.Provisioner = recipe.Fqid()
		}
	}

	return nil
}

func (s *State) ResolveParams(prof *profile.Profile) error {

	// for node capability
	for _, node := range s.Inventory.Nodes {
		for id, cap := range node.Capabilities {
			id = cap.SearchId(id)
			pcap, exist := prof.Params.Capabilities[id]
			if exist {
				var mergedParam types.Dict

				// capability in profile has higher priority than that in state
				mergedParam, err := types.Merge(cap.Params, pcap)
				if err != nil {
					return err
				}
				cap.Params = mergedParam
			}
		}
	}

	// for cluster capability
	for _, cluster := range s.Inventory.Clusters {

		// params of cluster in state
		stateCluster := cluster.Params

		// params of cluster in profile
		profCluster := prof.Params.Clusters[cluster.Version]

		for id, cap := range cluster.Capabilities {
			// params of capability in profile
			id = cap.SearchId(id)
			profCap := prof.Params.Capabilities[id]

			// params of capability in state
			var stateCap types.Dict
			if len(cap.Params) > 0 {
				stateCap = cap.Params
			}

			// Priority: cluster params in state < capability in state < cluster params in profile < capability in profile
			var mergedParam types.Dict
			var err error
			mergedParam, err = types.Merge(stateCluster, stateCap)
			if err != nil {
				return err
			}
			mergedParam, err = types.Merge(mergedParam, profCluster)
			if err != nil {
				return err
			}
			mergedParam, err = types.Merge(mergedParam, profCap)
			if err != nil {
				return err
			}

			// Replace capability params
			cap.Params = mergedParam
		}

		// clear cluster's params after merged
		cluster.Params = make(types.Params)
	}
	return nil
}

func (s *State) GetConnection(nodeId string) *transport.Connection {
	if s == nil || s.Inventory == nil || s.Inventory.Nodes == nil {
		return nil
	}

	st := NewAccessor(s)

	node := st.GetNode(nodeId)
	if node == nil {
		return nil
	}

	// look for instance connection
	if node.InstanceRef == "" {
		return nil
	}

	genName, instName := instance.ParseInstanceRef(node.InstanceRef)
	instance := st.GetGenerator(genName).GetInstance(instName)
	if instance == nil {
		return nil
	}
	c := instance.Connection.Clone()
	if c != nil {
		c.SetDefaults()
	}
	return c
}

func (s *State) Add(path []string, value any) error {
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

func (s *State) Remove(path []string) error {

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

func (s *State) Update(o patch.Patchable) error {
	return fmt.Errorf("not supported update for State")
}

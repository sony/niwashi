// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/profile"
	"github.com/sony/niwashi/internal/types"
)

const InitStageName = "_init"
const MainStageName = "main"
const DefaultStageName = MainStageName

type Stage struct {
	Inventory      *Inventory      `json:"inventory,omitempty" yaml:"inventory,omitempty"`
	Infrastructure *Infrastructure `json:"infrastructure,omitempty" yaml:"infrastructure,omitempty"`
	Export         Exports         `json:"export,omitempty" yaml:"export,omitempty"`
}

func (s *Stage) setDefaults() {
	if s == nil {
		return
	}

	if s.Infrastructure == nil {
		s.Infrastructure = &Infrastructure{}
	}
	s.Infrastructure.SetDefaults()

	if s.Inventory == nil {
		s.Inventory = NewInventory()
	}
	s.Inventory.SetDefaults()

	// don't init Export field by default
}

func (s *Stage) isEmpty() bool {
	if s == nil {
		return true
	}

	if s.Inventory != nil {
		if len(s.Inventory.Nodes) > 0 || len(s.Inventory.Clusters) > 0 {
			return false
		}
	}
	if s.Infrastructure != nil {
		if len(s.Infrastructure.Generators) > 0 {
			return false
		}
	}
	return true
}

func (s *Stage) clone() *Stage {
	if s == nil {
		return nil
	}

	clone := &Stage{}
	clone.Inventory = s.Inventory.Clone()
	clone.Infrastructure = s.Infrastructure.Clone()
	clone.Export = s.Export.Clone()

	return clone
}

func (s *Stage) merge(other *Stage) (*Stage, error) {

	if s == nil {
		return other.clone(), nil
	}
	if other == nil {
		return s.clone(), nil
	}

	var err error
	clone := &Stage{}
	clone.Inventory, err = s.Inventory.Merge(other.Inventory)
	if err != nil {
		return nil, err
	}
	clone.Infrastructure, err = s.Infrastructure.Merge(other.Infrastructure)
	if err != nil {
		return nil, err
	}
	clone.Export, err = s.Export.Merge(other.Export)
	if err != nil {
		return nil, err
	}

	return clone, nil
}

func (s *Stage) validate() error {

	var err error
	if s.Infrastructure != nil {
		if e := s.Infrastructure.Validate(); e != nil {
			err = errors.Join(err, e)
		}
	}

	if s.Inventory != nil {
		if e := s.Inventory.Validate(); e != nil {
			err = errors.Join(err, e)
		}
	}

	if s.Export != nil {
		if e := s.Export.Validate(); e != nil {
			err = errors.Join(err, e)
		}
	}

	e := s.validateNodeBinding()
	if e != nil {
		err = errors.Join(err, e)
	}

	return err
}

func (s *Stage) validateNodeBinding() error {

	if s.Inventory != nil {
		if e := s.validateInventory(); e != nil {
			return e
		}
	}

	if s.Infrastructure != nil {
		if e := s.validateInfrastructure(); e != nil {
			return e
		}
	}

	return nil
}

func (s *Stage) validateInventory() error {

	// node -> instance
	for nodeName, v := range s.Inventory.Nodes {
		if v.InstanceRef == "" {
			continue
		}

		if err := instance.ValidateInstanceRef(v.InstanceRef); err != nil {
			return fmt.Errorf("invalid instance reference %q in node %q: %w", v.InstanceRef, nodeName, err)
		}

		genName, instName := instance.ParseInstanceRef(v.InstanceRef)
		_, inst, err := s.findInstance(genName, instName)
		if err != nil {
			return err
		}

		if inst.NodeRef != nodeName {
			return fmt.Errorf("node %q and instance %q binding mismatch: node's instanceRef=%q but instance's nodeRef=%q", nodeName, v.InstanceRef, v.InstanceRef, inst.NodeRef)
		}
	}
	return nil
}

func (s *Stage) validateInfrastructure() error {
	// instance -> node
	for genName, v := range s.Infrastructure.Generators {
		for instName, i := range v.Instances {
			if i.NodeRef == "" {
				continue
			}

			if s.Inventory == nil || s.Inventory.Nodes == nil {
				return fmt.Errorf("node %q not found for instance %q because inventory is nil or inventory nodes are nil", i.NodeRef, instName)
			}

			node, ok := s.Inventory.Nodes[i.NodeRef]
			if !ok {
				return fmt.Errorf("node %q not found for instance %q", i.NodeRef, instName)
			}

			instanceRef := instance.MakeInstanceRef(genName, instName)
			if node.InstanceRef != instanceRef {
				return fmt.Errorf("node %q and instance %q binding mismatch: node's instanceRef=%q but instance's nodeRef=%q", i.NodeRef, instanceRef, node.InstanceRef, i.NodeRef)
			}
		}
	}
	return nil
}

func (s *Stage) findInstance(genName, instName string) (*instance.Generator, *instance.Instance, error) {
	if s.Infrastructure == nil {
		return nil, nil, fmt.Errorf("infrastructure is nil")
	}
	if s.Infrastructure.Generators == nil {
		return nil, nil, fmt.Errorf("infrastructure generators is nil")
	}

	gen, ok := s.Infrastructure.Generators[genName]
	if !ok {
		return nil, nil, fmt.Errorf("generator %q not found", genName)
	}

	inst, ok := gen.Instances[instName]
	if !ok {
		return nil, nil, fmt.Errorf("instance %q not found in generator %q", instName, genName)
	}
	return gen, inst, nil
}

func (s *Stage) applyTemplate(t *Template) error {

	if s == nil || t == nil {
		return nil
	}

	if len(t.Node) != 0 && s.Inventory != nil {
		for _, n := range s.Inventory.Nodes {
			err := n.ApplyTemplate(t.Node)
			if err != nil {
				return err
			}
		}
	}

	if len(t.Cluster) != 0 && s.Inventory != nil {
		for _, cluster := range s.Inventory.Clusters {
			err := cluster.ApplyTemplate(t.Cluster)
			if err != nil {
				return err
			}
		}
	}

	if len(t.Generator) != 0 && s.Infrastructure != nil {
		for _, inst := range s.Infrastructure.Generators {
			err := inst.ApplyTemplate(t.Generator)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Stage) normalize(f cmap.RecipeFinder) error {
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
	if s.Infrastructure != nil {
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
	}

	// Normalize exports
	if s.Export != nil {
		if err := s.Export.Normalize(f); err != nil {
			return err
		}
	}

	return nil
}

func (s *Stage) resolveParams(prof *profile.Profile) error {
	if s == nil || prof == nil || s.Inventory == nil {
		return nil
	}

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

func (s *Stage) ValidateImports(exportData Exports) error {
	if s == nil {
		return nil
	}

	if err := validateNodeCapImports(exportData, s); err != nil {
		return err
	}
	if err := validateClusterCapImports(exportData, s); err != nil {
		return err
	}
	if err := validateGeneratorImports(exportData, s); err != nil {
		return err
	}

	return nil
}

func validateNodeCapImports(exportData Exports, stage *Stage) error {
	if stage == nil || stage.Inventory == nil || stage.Inventory.Nodes == nil {
		return nil
	}

	for name, node := range stage.Inventory.Nodes {
		if node.Capabilities == nil {
			continue
		}

		if err := validateCapImports(exportData, node.Capabilities); err != nil {
			return fmt.Errorf("node %q: %w", name, err)
		}
	}

	return nil
}

func validateClusterCapImports(exportData Exports, stage *Stage) error {
	if stage == nil || stage.Inventory == nil || stage.Inventory.Clusters == nil {
		return nil
	}

	for name, cluster := range stage.Inventory.Clusters {
		if cluster.Capabilities == nil {
			continue
		}

		if err := validateCapImports(exportData, cluster.Capabilities); err != nil {
			return fmt.Errorf("cluster %q: %w", name, err)
		}
	}
	return nil
}

func validateCapImports(exportData Exports, capList cap.CapabilityList) error {
	for capID, c := range capList {
		if err := exportData.ValidateImports(c.Params, c.Import); err != nil {
			return fmt.Errorf("capability %q: %w", capID, err)
		}
	}
	return nil
}

func validateGeneratorImports(exportData Exports, stage *Stage) error {
	if stage == nil || stage.Infrastructure == nil || stage.Infrastructure.Generators == nil {
		return nil
	}

	for name, gen := range stage.Infrastructure.Generators {
		if err := exportData.ValidateImports(gen.Params, gen.Import); err != nil {
			return fmt.Errorf("generator %q: %w", name, err)
		}
	}

	return nil
}

func (s *Stage) GetExportValue(exp *Export) (any, error) {

	caps, err := exp.Capabilities(s)
	if err != nil {
		return nil, err
	}

	c, exists := caps[exp.Capability]
	if !exists {
		return nil, fmt.Errorf("capability %q not found", exp.Capability)
	}

	paths, err := patch.ParsePath("/" + patch.TrimForPath(exp.Path))
	if err != nil {
		return nil, err
	}
	d := c.Store.Clone()
	data, ok := d.Get(paths...)
	if !ok {
		return nil, fmt.Errorf("export path %q not found in capability %q", exp.Path, exp.Capability)
	}

	return data, nil
}

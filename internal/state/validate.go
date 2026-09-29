// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/types"
)

func (s *State) Validate() error {
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

	if s.Template != nil {
		if e := s.Template.Validate(); e != nil {
			err = errors.Join(err, e)
		}
	}

	e := s.ValidateNodeBinding()
	if e != nil {
		err = errors.Join(err, e)
	}

	if s.Runtime != nil {
		if e := s.Runtime.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid runtime: %w", e))
		}
	}

	return err
}

func (s *State) findInstance(genName, instName string) (*instance.Generator, *instance.Instance, error) {
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

func (s *State) ValidateNodeBinding() error {

	if s.Inventory != nil {
		if e := s.validateInventory(); e != nil {
			return e
		}
	}

	if s.Infrastructure != nil {
		if e := s.ValidateInfrastructure(); e != nil {
			return e
		}
	}

	return nil
}

func (s *State) validateInventory() error {

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

func (s *State) ValidateInfrastructure() error {
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

func (i *Infrastructure) Validate() error {
	var err error
	for k, g := range i.Generators {
		if !types.GeneratorIdPattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid generator ID pattern: %s", k))
		}

		if e := g.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid generator=%q %w", k, e))
		}
	}
	return err
}

func (i *Inventory) Validate() error {
	var err error
	for k, n := range i.Nodes {
		if !types.NodeIdPattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid node ID pattern: %s", k))
		}

		if e := n.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("node=%q %w", k, e))
		}
	}
	for k, c := range i.Clusters {
		if !types.ClusterIdPattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid cluster ID pattern: %s", k))
		}

		if e := c.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("cluster=%q %w", k, e))
		}
	}
	return err
}

func (t *Template) Validate() error {
	var err error
	for k, v := range t.Node {
		if !types.NodeIdPattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid node template ID pattern: %s", k))
		}

		if e := v.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid node template %q: %w", k, e))
		}
	}

	for k, v := range t.Cluster {
		if !types.ClusterIdPattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid cluster template ID pattern: %s", k))
		}

		if e := v.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid cluster template %q: %w", k, e))
		}
	}

	for k, v := range t.Generator {
		if !types.GeneratorIdPattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid generator template ID pattern: %s", k))
		}

		if e := v.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid generator template %q: %w", k, e))
		}
	}

	return err
}

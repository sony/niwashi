// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/types"
)

func (s *State) Validate() error {
	var err error

	if e := s.Stage.validate(); e != nil { //nolint:staticcheck
		err = errors.Join(err, e)
	}

	// deny "export" field
	if s.Export != nil {
		err = errors.Join(err, fmt.Errorf("export field is not allowed in the main stage"))
	}

	if s.Stages != nil {
		if e := s.Stages.validate(); e != nil {
			err = errors.Join(err, e)
		}
	}

	if s.Template != nil {
		if e := s.Template.Validate(); e != nil {
			err = errors.Join(err, e)
		}
	}

	if s.Runtime != nil {
		if e := s.Runtime.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid runtime: %w", e))
		}
	}

	return err
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

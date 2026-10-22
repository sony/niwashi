// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/recipe"
)

type Exports map[string]*Export

func (e Exports) Clone() Exports {
	if e == nil {
		return nil
	}
	clone := make(Exports, len(e))
	for k, v := range e {
		clone[k] = v.Clone()
	}
	return clone
}

func (e Exports) Merge(other Exports) (Exports, error) {
	if e == nil {
		return other.Clone(), nil
	} else if other == nil {
		return e.Clone(), nil
	}

	merged := e.Clone()
	for k := range other {
		merged[k] = other[k].Clone()
	}
	return merged, nil
}

func (e Exports) Validate() error {
	if e == nil {
		return nil
	}

	var err error
	for k, v := range e {
		if er := v.Validate(); er != nil {
			err = errors.Join(err, fmt.Errorf("export %q: %w", k, er))
		}
	}
	return err
}

func (e Exports) ValidateWith(stage, base *Stage) error {
	// this should be called after normalization
	if e == nil {
		return nil
	}

	var err error
	for k, v := range e {
		if er := v.ValidateWith(stage, base); er != nil {
			err = errors.Join(err, fmt.Errorf("export %q: %w", k, er))
		}
	}
	return err
}

func (e Exports) Normalize(f cmap.RecipeFinder) error {
	if e == nil {
		return nil
	}

	for k, v := range e {
		if err := v.Normalize(f); err != nil {
			return fmt.Errorf("export %q: %w", k, err)
		}
	}
	return nil
}

type Export struct {
	Node       string `json:"node,omitempty" yaml:"node,omitempty"`
	Cluster    string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Capability string `json:"capability,omitempty" yaml:"capability,omitempty"`
	Path       string `json:"path,omitempty" yaml:"path,omitempty"`
}

func (e *Export) Clone() *Export {
	if e == nil {
		return nil
	}
	return &Export{
		Node:       e.Node,
		Cluster:    e.Cluster,
		Capability: e.Capability,
		Path:       e.Path,
	}
}

func (e *Export) Validate() error {
	if e == nil {
		return fmt.Errorf("export is nil")
	}

	var err error
	if e.Node == "" && e.Cluster == "" {
		err = errors.Join(err, fmt.Errorf("export has empty node or cluster"))
	} else if e.Node != "" && e.Cluster != "" {
		err = errors.Join(err, fmt.Errorf("export has both node and cluster set"))
	}
	if e.Capability == "" {
		err = errors.Join(err, fmt.Errorf("export has empty capability"))
	} else if strings.Contains(e.Capability, "@") {
		if _, _, er := recipe.ParseFqid(e.Capability); er != nil {
			err = errors.Join(err, fmt.Errorf("export has invalid capability: %w", er))
		} else {
			// Print warning
			logger.Warn("export will ignore capability version", "capability", e.Capability)
		}
	}
	if e.Path == "" {
		err = errors.Join(err, fmt.Errorf("export has empty path"))
	}
	return err
}

func (e *Export) Capabilities(s *Stage) (cap.CapabilityList, error) {
	if s == nil {
		return nil, fmt.Errorf("stage is nil")
	}

	if e.Node != "" {
		if s.Inventory == nil || s.Inventory.Nodes == nil {
			return nil, fmt.Errorf("no inventory for node: %q", e.Node)
		}
		if n, ok := s.Inventory.Nodes[e.Node]; ok {
			return n.Capabilities, nil
		} else {
			return nil, fmt.Errorf("node not found: %q", e.Node)
		}
	} else if e.Cluster != "" {
		if s.Inventory == nil || s.Inventory.Clusters == nil {
			return nil, fmt.Errorf("no inventory for cluster: %q", e.Cluster)
		}
		if c, ok := s.Inventory.Clusters[e.Cluster]; ok {
			return c.Capabilities, nil
		} else {
			return nil, fmt.Errorf("cluster not found: %q", e.Cluster)
		}
	} else {
		return nil, fmt.Errorf("export has empty node or cluster")
	}
}

func (e *Export) ValidateWith(stages ...*Stage) error {

	looked, subjectFound := false, false
	for _, s := range stages {
		if s == nil {
			continue
		}

		looked = true

		c, err := e.Capabilities(s)
		if err != nil {
			continue
		}
		subjectFound = true

		if _, exists := c[e.Capability]; exists {
			return nil
		}
	}

	switch {
	case !looked:
		return fmt.Errorf("no stage to look up the capability")
	case !subjectFound:
		return fmt.Errorf("%s not found", e.subject())
	default:
		return fmt.Errorf("capability %q not found on %s", e.Capability, e.subject())
	}
}

func (e *Export) subject() string {
	if e.Node != "" {
		return fmt.Sprintf("node %q", e.Node)
	} else if e.Cluster != "" {
		return fmt.Sprintf("cluster %q", e.Cluster)
	} else {
		return "unknown subject"
	}
}

func (e *Export) Normalize(f cmap.RecipeFinder) error {
	if e == nil {
		return nil
	}

	c := e.Capability
	if strings.Contains(c, "@") {
		id, _, err := recipe.ParseFqid(c)
		if err != nil {
			return fmt.Errorf("failed to parse capability fqid %q: %w", c, err)
		}
		c = id
	}

	r, err := f.FindRecipe(c)
	if err != nil {
		return fmt.Errorf("failed to find recipe for capability %q: %w", c, err)
	}

	e.Capability = r.Metadata.Id

	return nil
}

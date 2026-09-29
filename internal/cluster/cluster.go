// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cluster

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/tmpl"
	"github.com/sony/niwashi/internal/types"
)

type Cluster struct {
	Templates    []string           `json:"templates,omitempty" yaml:"templates,omitempty"`
	Labels       map[string]string  `json:"labels,omitempty" yaml:"labels,omitempty"`
	Nodes        NodeList           `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	Capabilities cap.CapabilityList `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
	Version      string             `json:"version,omitempty" yaml:"version,omitempty"`
	Params       types.Params       `json:"params,omitempty" yaml:"params,omitempty"`
}

func NewCluster() *Cluster {
	cluster := &Cluster{}
	cluster.SetDefaults()
	return cluster
}

func (cluster *Cluster) SetDefaults() {
	if cluster.Templates == nil {
		cluster.Templates = []string{}
	}
	if cluster.Nodes == nil {
		cluster.Nodes = NodeList{}
	}
	if cluster.Capabilities == nil {
		cluster.Capabilities = cap.CapabilityList{}
	}
	if cluster.Params == nil {
		cluster.Params = make(types.Params)
	}
}

func (c *Cluster) Validate() error {

	var err error
	if c.Capabilities != nil {
		if e := c.Capabilities.Validate(); e != nil {
			err = errors.Join(err, e)
		}
	}

	if e := c.Params.ValidateAsParams(); e != nil {
		err = errors.Join(err, fmt.Errorf("invalid cluster params: %w", e))
	}

	if c.Nodes != nil {
		if e := c.Nodes.Validate(); e != nil {
			err = errors.Join(err, e)
		}
	}

	if c.Labels != nil {
		err = errors.Join(err, fmt.Errorf("cluster labels are deprecated"))
	}

	return err
}

func (c *Cluster) Clone() *Cluster {
	if c == nil {
		return nil
	}
	cloned := &Cluster{}
	cloned.Templates = append([]string{}, c.Templates...)
	cloned.Nodes = c.Nodes.Clone()
	cloned.Capabilities = c.Capabilities.Clone()
	cloned.Version = c.Version
	cloned.Params = c.Params.Clone()
	return cloned
}

func (c *Cluster) Merge(other *Cluster) (*Cluster, error) {
	if c == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return c.Clone(), nil
	}

	clone := &Cluster{}
	var err error
	clone.Templates = merge.StringArray(c.Templates, other.Templates)
	clone.Nodes = c.Nodes.Merge(other.Nodes)
	clone.Capabilities, err = c.Capabilities.Merge(other.Capabilities)
	if err != nil {
		return nil, err
	}
	clone.Version = merge.String(c.Version, other.Version)
	clone.Params, err = types.Merge(c.Params, other.Params)
	if err != nil {
		return nil, err
	}
	return clone, nil
}

func (c *Cluster) ApplyTemplate(tpl map[string]*Template) error {
	if len(c.Templates) == 0 {
		// do nothing
		return nil
	}

	order := []*Template{}
	for _, tplName := range c.Templates {
		t, exist := tpl[tplName]
		if !exist {
			return fmt.Errorf("node template %q not found", tplName)
		}
		t2, err := tmpl.Resolve(tpl, t)
		if err != nil {
			return err
		}

		t2.From = "" // remove From field after merge
		order = append(order, t2)
	}

	for _, t := range order {
		c.apply(t)
	}

	// clear Templates after applied
	c.Templates = []string{}
	return nil
}

func (c *Cluster) apply(tpl *Template) {
	// merge fields from tpl into i
	c.Version = merge.String(c.Version, tpl.Version)
	var err error
	c.Params, err = types.Merge(c.Params, tpl.Params)
	if err != nil {
		// handle error if needed
		return
	}
	c.Capabilities, err = c.Capabilities.Merge(tpl.Capabilities)
	if err != nil {
		// handle error if needed
		return
	}
}

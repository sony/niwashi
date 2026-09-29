// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package node

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/clone"
	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/tmpl"
	"github.com/sony/niwashi/internal/types"
)

type Node struct {
	Labels           map[string]string  `json:"labels,omitempty" yaml:"labels,omitempty"`
	Templates        []string           `json:"templates,omitempty" yaml:"templates,omitempty"`
	InstanceRef      string             `json:"instanceRef,omitempty" yaml:"instanceRef,omitempty"` // "generatorName:instanceName"
	InstanceSelector *instance.Selector `json:"instanceSelector,omitempty" yaml:"instanceSelector,omitempty"`
	Capabilities     cap.CapabilityList `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
}

func NewNode() *Node {
	node := &Node{}
	node.SetDefaults()
	return node
}

func (n *Node) SetDefaults() {
	if n.Labels == nil {
		n.Labels = make(map[string]string)
	}
	if n.Templates == nil {
		n.Templates = []string{}
	}
	if n.Capabilities == nil {
		n.Capabilities = cap.CapabilityList{}
	}
}

func (n *Node) Validate() error {

	var err error

	if n.InstanceRef != "" {
		if e := instance.ValidateInstanceRef(n.InstanceRef); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid instance reference: %w", e))
		}
	}

	if n.Capabilities != nil {
		if e := n.Capabilities.Validate(); e != nil {
			err = errors.Join(err, e)
		}
	}

	for k := range n.Labels {
		if !types.LabelKeyPattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid label key: %q", k))
		}
	}

	return err
}

func (n *Node) Clone() *Node {
	if n == nil {
		return nil
	}

	c := &Node{}
	c.Labels = clone.StringMap(n.Labels)
	c.Templates = append([]string{}, n.Templates...)
	c.InstanceRef = n.InstanceRef
	c.InstanceSelector = n.InstanceSelector.Clone()
	c.Capabilities = n.Capabilities.Clone()
	return c
}

func (n *Node) Merge(other *Node) (*Node, error) {
	if n == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return n.Clone(), nil
	}

	clone := &Node{}
	var err error
	clone.Labels = merge.StringMap(n.Labels, other.Labels)
	clone.Templates = merge.StringArray(n.Templates, other.Templates)
	clone.InstanceRef = merge.String(n.InstanceRef, other.InstanceRef)
	clone.InstanceSelector = n.InstanceSelector.Merge(other.InstanceSelector)
	clone.Capabilities, err = n.Capabilities.Merge(other.Capabilities)
	if err != nil {
		return nil, err
	}
	return clone, nil
}

func (n *Node) ApplyTemplate(tpl map[string]*Template) error {
	if len(n.Templates) == 0 {
		// do nothing
		return nil
	}

	order := []*Template{}
	for _, tplName := range n.Templates {
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
		n.apply(t)
	}

	// clear Templates after applied
	n.Templates = []string{}

	return nil
}

func (n *Node) apply(tpl *Template) {
	// merge fields from tpl into n
	n.Labels = merge.StringMap(n.Labels, tpl.Labels)
	caps, _ := tpl.Capabilities.Merge(n.Capabilities)
	n.Capabilities = caps
}

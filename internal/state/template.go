// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/node"
)

type Template struct {
	Node      map[string]*node.Template     `json:"node,omitempty" yaml:"node,omitempty"`
	Cluster   map[string]*cluster.Template  `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Generator map[string]*instance.Template `json:"generator,omitempty" yaml:"generator,omitempty"`
}

func (t *Template) SetDefaults() *Template {
	if t == nil {
		t = &Template{}
	}

	if t.Node == nil {
		t.Node = map[string]*node.Template{}
	}
	if t.Cluster == nil {
		t.Cluster = map[string]*cluster.Template{}
	}
	if t.Generator == nil {
		t.Generator = map[string]*instance.Template{}
	}

	return t
}

func (t *Template) Clone() *Template {
	if t == nil {
		return nil
	}

	clone := &Template{}
	clone.Node = make(map[string]*node.Template, len(t.Node))
	for k, v := range t.Node {
		clone.Node[k] = v.Clone()
	}
	clone.Cluster = make(map[string]*cluster.Template, len(t.Cluster))
	for k, v := range t.Cluster {
		clone.Cluster[k] = v.Clone()
	}
	clone.Generator = make(map[string]*instance.Template, len(t.Generator))
	for k, v := range t.Generator {
		clone.Generator[k] = v.Clone()
	}

	return clone
}

func (t *Template) Merge(other *Template) (*Template, error) {
	if t == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return t.Clone(), nil
	}

	clone := &Template{}
	var err error
	clone.Node, err = merge.StringMapT(t.Node, other.Node)
	if err != nil {
		return nil, err
	}
	clone.Cluster, err = merge.StringMapT(t.Cluster, other.Cluster)
	if err != nil {
		return nil, err
	}
	clone.Generator, err = merge.StringMapT(t.Generator, other.Generator)
	if err != nil {
		return nil, err
	}
	return clone, nil
}

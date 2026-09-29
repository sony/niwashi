// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package instance

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/tmpl"
	"github.com/sony/niwashi/internal/types"
)

type Generator struct {
	Templates   []string             `json:"templates,omitempty" yaml:"templates,omitempty"`
	Provisioner string               `json:"provisioner,omitempty" yaml:"provisioner,omitempty"`
	Params      types.Params         `json:"params,omitempty" yaml:"params,omitempty"`
	Instances   map[string]*Instance `json:"instances,omitempty" yaml:"instances,omitempty"`
	Store       types.Dict           `json:"store,omitempty" yaml:"store,omitempty"` // for runtime use, not merged or applied
}

func NewGenerator() *Generator {
	g := &Generator{}
	g.SetDefaults()
	return g
}

func (g *Generator) SetDefaults() {
	if g.Params == nil {
		g.Params = make(types.Params)
	}
	if g.Instances == nil {
		g.Instances = make(map[string]*Instance)
	}
	for _, i := range g.Instances {
		i.SetDefaults()
	}
}

func (g *Generator) Validate() error {
	var err error

	if g.Provisioner == "" {
		err = errors.Join(err, fmt.Errorf("provisioner is required for generator"))
	}

	for k, i := range g.Instances {
		if !types.InstanceIdPattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid instance name: %s", k))
		}
		if e := i.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid instance %q in generator: %w", k, e))
		}
	}

	e := g.Params.ValidateAsParams()
	if e != nil {
		err = errors.Join(err, fmt.Errorf("invalid generator params: %w", e))
	}

	e = g.Store.ValidateAsTemplateVar()
	if e != nil {
		err = errors.Join(err, fmt.Errorf("invalid generator store: %w", e))
	}

	return err
}

func (g *Generator) Clone() *Generator {
	if g == nil {
		return nil
	}
	clone := &Generator{}
	clone.Templates = append([]string{}, g.Templates...)
	clone.Provisioner = g.Provisioner
	clone.Params = g.Params.Clone()
	clone.Instances = make(map[string]*Instance, len(g.Instances))
	for k, v := range g.Instances {
		clone.Instances[k] = v.Clone()
	}
	clone.Store = g.Store.Clone()

	return clone
}

func (g *Generator) Merge(other *Generator) (*Generator, error) {
	if g == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return g.Clone(), nil
	}

	clone := &Generator{}
	var err error
	clone.Templates = merge.StringArray(g.Templates, other.Templates)
	clone.Provisioner = merge.String(g.Provisioner, other.Provisioner)
	clone.Params, err = types.Merge(g.Params, other.Params)
	if err != nil {
		return nil, err
	}
	clone.Instances, err = merge.StringMapT(g.Instances, other.Instances)
	if err != nil {
		return nil, err
	}
	clone.Store, err = types.Merge(g.Store, other.Store)
	if err != nil {
		return nil, err
	}
	return clone, nil
}

func (g *Generator) ApplyTemplate(tpl map[string]*Template) error {
	if len(g.Templates) == 0 {
		// do nothing
		return nil
	}

	order := []*Template{}
	for _, tplName := range g.Templates {
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
		g.apply(t)
	}

	// clear Templates after applied
	g.Templates = []string{}
	return nil
}

func (g *Generator) apply(tpl *Template) {
	// merge fields from tpl into g
	g.Provisioner = merge.String(g.Provisioner, tpl.Provisioner)
	g.Params, _ = types.Merge(g.Params, tpl.Params)
}

func (g *Generator) Add(path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "instances":
		if len(path) == 1 {
			// add instances entirely
			p, err := patch.ConvertValue[map[string]*Instance](value)
			if err != nil {
				return err
			}
			for k, v := range p {
				if g.Instances[k] != nil {
					logger.Warn("overwriting existing instance entry:", "key", k)
				}
				g.Instances[k] = v
			}
			return nil
		}
		return patch.AddCompositeToMap(
			path[1:],
			value,
			g.Instances)
	case "store":
		if len(path) == 1 {
			v, ok := value.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid value type for generator store: %T", value)
			}
			g.Store = types.CloneMap(v)
			return nil
		}
		if g.Store == nil {
			g.Store = make(map[string]any)
		}
		return g.Store.SetPath(value, path[1:]...)
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (g *Generator) Remove(path []string) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "instances":
		if len(path) == 1 {
			// remove instances entirely
			g.Instances = make(map[string]*Instance)
			return nil
		}
		return patch.RemoveCompositeMap(
			path[1:],
			g.Instances)
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (*Generator) Update(o patch.Patchable) error {
	return fmt.Errorf("not supported update for Generator")
}

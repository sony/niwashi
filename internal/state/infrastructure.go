// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"fmt"

	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/patch"
)

type Infrastructure struct {
	Generators map[string]*instance.Generator `json:"generators,omitempty" yaml:"generators,omitempty"`
}

func (inf *Infrastructure) SetDefaults() {
	if inf.Generators == nil {
		inf.Generators = make(map[string]*instance.Generator)
	}
	for instName, inst := range inf.Generators {
		if inst == nil {
			inf.Generators[instName] = instance.NewGenerator()
		} else {
			inst.SetDefaults()
		}
	}
}

func (inf *Infrastructure) Clone() *Infrastructure {
	if inf == nil {
		return nil
	}

	clone := &Infrastructure{}
	clone.Generators = make(map[string]*instance.Generator, len(inf.Generators))
	for k, v := range inf.Generators {
		clone.Generators[k] = v.Clone()
	}

	return clone
}

func (i *Infrastructure) Merge(other *Infrastructure) (*Infrastructure, error) {
	if i == nil {
		return other.Clone(), nil
	}
	if other == nil {
		return i.Clone(), nil
	}

	clone := &Infrastructure{}
	var err error
	clone.Generators, err = merge.StringMapT(i.Generators, other.Generators)
	if err != nil {
		return nil, err
	}

	return clone, nil
}

func (i *Infrastructure) Add(path []string, value any) error {

	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "generators":
		return patch.AddCompositeToMap(
			path[1:],
			value,
			i.Generators)
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (i *Infrastructure) Remove(path []string) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "generators":
		return patch.RemoveCompositeMap(
			path[1:],
			i.Generators)
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (*Infrastructure) Update(o patch.Patchable) error {
	return fmt.Errorf("not supported update for Infrastructure")
}

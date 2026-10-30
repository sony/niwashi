// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cap

import (
	"errors"
	"fmt"
	"maps"

	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/types"
)

type Capability struct {
	Version  string            `json:"version,omitempty" yaml:"version,omitempty"`
	Params   types.Params      `json:"params,omitempty" yaml:"params,omitempty"`
	Store    types.Dict        `json:"store,omitempty" yaml:"store,omitempty"` // for runtime use, not merged or applied
	Import   types.Imports     `json:"import,omitempty" yaml:"import,omitempty"`
	Imported map[string]string `json:"imported,omitempty" yaml:"imported,omitempty"`
}

func (c *Capability) SearchId(id string) string {
	if c == nil {
		return id
	}
	if c.Version != "" {
		return id + "@" + c.Version
	}
	return id
}

func (c *Capability) Clone() *Capability {
	if c == nil {
		return nil
	}

	clone := &Capability{}
	clone.Version = c.Version
	clone.Params = c.Params.Clone()
	clone.Store = c.Store.Clone()
	clone.Import = c.Import.Clone()
	clone.Imported = maps.Clone(c.Imported)

	return clone
}

func (c *Capability) Validate() error {
	var err error
	if e := c.Params.ValidateAsParams(); e != nil {
		err = errors.Join(err, e)
	}

	if e := c.Store.ValidateAsTemplateVar(); e != nil {
		err = errors.Join(err, e)
	}

	if e := c.Import.Validate(); e != nil {
		err = errors.Join(err, e)
	}

	return err
}

func (c *Capability) Merge(other *Capability) (*Capability, error) {
	if other == nil {
		return c.Clone(), nil
	}
	if c == nil {
		return other.Clone(), nil
	}

	clone := &Capability{}
	clone.Version = func() string {
		if len(other.Version) == 0 {
			return c.Version
		}
		return other.Version
	}()
	var err error
	clone.Params, err = types.Merge(c.Params, other.Params)
	if err != nil {
		return nil, err
	}

	clone.Store, err = types.Merge(c.Store, other.Store)
	if err != nil {
		return nil, err
	}

	clone.Import = c.Import.Merge(other.Import)

	return clone, nil
}

// replaceWhole sets a field of a capability that is replaced as a whole.
func replaceWhole[T any](field string, rest []string, value any, set func(T)) error {
	if len(rest) != 0 {
		return fmt.Errorf("invalid path for capability %s: %v", field, rest)
	}
	v, ok := value.(T)
	if !ok {
		return fmt.Errorf("invalid value type for capability %s: %T", field, value)
	}
	set(v)
	return nil
}

func (c *Capability) Add(path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("path is required to add capability")
	}

	switch path[0] {
	case "version":
		return replaceWhole("version", path[1:], value, func(v string) {
			c.Version = v
		})
	case "params":
		return replaceWhole("params", path[1:], value, func(v map[string]any) {
			c.Params = types.CloneMap(v)
		})
	case "store":
		if len(path) == 1 {
			v, ok := value.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid value type for capability store: %T", value)
			}
			c.Store = types.CloneMap(v)
			return nil
		}
		if c.Store == nil {
			c.Store = make(map[string]any)
		}
		return c.Store.SetPath(value, path[1:]...)

	case "import":
		return replaceWhole("import", path[1:], value, func(v types.Imports) {
			c.Import = v.Clone()
		})

	case "imported":
		return replaceWhole("imported", path[1:], value, func(v map[string]string) {
			c.Imported = maps.Clone(v)
		})

	default:
		return fmt.Errorf("unsupported path for capability: %s", path[0])
	}
}

func (c *Capability) Remove(path []string) error {
	return fmt.Errorf("removing capability is not supported")
}

func (c *Capability) Update(p patch.Patchable) error {
	return fmt.Errorf("updating capability is not supported")
}

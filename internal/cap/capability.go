// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cap

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/types"
)

type Capability struct {
	Version string       `json:"version,omitempty" yaml:"version,omitempty"`
	Params  types.Params `json:"params,omitempty" yaml:"params,omitempty"`
	Store   types.Dict   `json:"store,omitempty" yaml:"store,omitempty"` // for runtime use, not merged or applied
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

	return clone, nil
}

func (c *Capability) Add(path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("path is required to add capability")
	}

	switch path[0] {
	case "version":
		if len(path) == 1 {
			v, ok := value.(string)
			if !ok {
				return fmt.Errorf("invalid value type for capability version: %T", value)
			}
			c.Version = v
			return nil
		}
		// deny adding nested paths under version
		return fmt.Errorf("invalid path for capability version: %v", path[1:])
	case "params":
		if len(path) == 1 {
			v, ok := value.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid value type for capability params: %T", value)
			}
			c.Params = types.CloneMap(v)
			return nil
		}

		// deny adding nested paths under params
		return fmt.Errorf("invalid path for capability params: %v", path[1:])

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

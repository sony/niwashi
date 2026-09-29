// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cap

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/types"
	"gopkg.in/yaml.v3"
)

type CapabilityList map[string]*Capability

type userCapability struct {
	Id      string       `json:"id,omitempty" yaml:"id,omitempty"`
	Version string       `json:"version,omitempty" yaml:"version,omitempty"`
	Params  types.Params `json:"params,omitempty" yaml:"params,omitempty"`
}

func (c *CapabilityList) Clone() CapabilityList {
	if c == nil || *c == nil {
		return nil
	}
	clone := CapabilityList{}
	for id, cap := range *c {
		clone[id] = cap.Clone()
	}
	return clone
}

func (c *CapabilityList) Validate() error {
	var err error
	for id, cap := range *c {
		if id == "" {
			err = errors.Join(err, fmt.Errorf("capability id cannot be empty"))
		}
		if cap == nil {
			err = errors.Join(err, fmt.Errorf("capability cannot be nil for id: %s", id))
		}
		if e := cap.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("capability=%q %w", id, e))
		}
	}
	return err
}

func (c CapabilityList) Merge(other CapabilityList) (CapabilityList, error) {
	if len(other) == 0 {
		return c.Clone(), nil
	}
	if len(c) == 0 {
		return other.Clone(), nil
	}

	capMap := c.Clone()
	var err error
	for i, cap := range other {
		if _, exist := capMap[i]; exist {
			capMap[i], err = capMap[i].Merge(cap)
			if err != nil {
				return nil, err
			}
		} else {
			capMap[i] = cap.Clone()
		}
	}

	return capMap, nil
}

func (c CapabilityList) Find(fqid string) *Capability {
	id, version, err := recipe.ParseFqid(fqid)
	if err != nil {
		// not FQID format
		id = fqid
		version = ""
	}

	cap, found := c[id]
	if !found {
		return nil
	}

	if version != "" && cap.Version != version {
		// version mismatch
		return nil
	}

	return cap
}

func (c *CapabilityList) UnmarshalYAML(node *yaml.Node) error {
	if *c == nil {
		*c = CapabilityList{}
	}

	switch node.Kind {
	case yaml.SequenceNode:
		// list of capabilities
		for _, item := range node.Content {
			switch item.Kind {
			case yaml.ScalarNode:

				// ID as string (not yet normalized)
				(*c)[item.Value] = &Capability{}

			case yaml.MappingNode:
				// Capability as map

				var cap userCapability
				if err := item.Decode(&cap); err != nil {
					return err
				}
				(*c)[cap.Id] = &Capability{
					Version: cap.Version,
					Params:  cap.Params,
				}
			default:
				return fmt.Errorf("invalid capability format")
			}
		}

	case yaml.MappingNode:
		// map of capabilities
		var capMap map[string]*Capability
		if err := node.Decode(&capMap); err != nil {
			return err
		}

		for id, cap := range capMap {
			(*c)[id] = cap
		}
	default:
		return fmt.Errorf("capabilities must be a sequence")
	}

	return nil
}

func (c *CapabilityList) Normalize(f cmap.RecipeFinder) (CapabilityList, error) {
	normalized := CapabilityList{}

	for orig, cap := range *c {
		id := cap.SearchId(orig)

		recipe, err := f.FindRecipe(id)
		if err != nil {
			return nil, err
		}

		nCap := Capability{
			Version: recipe.Metadata.Version,
			Params:  cap.Params,
		}

		normalized[recipe.Metadata.Id] = &nCap

		if orig != recipe.Fqid() {
			logger.Info("UpdateRecipeAlias", "from", orig, "to", recipe.Fqid())
		}
	}

	return normalized, nil
}

func (c CapabilityList) Add(path []string, value any) error {

	if len(path) == 1 {
		id := path[0]

		v, ok := value.(*Capability)
		if ok {
			c[id] = v
		} else {
			jsonData, err := json.Marshal(value)
			if err != nil {
				return nil
			}
			var cap Capability
			err = json.Unmarshal(jsonData, &cap)
			if err != nil {
				return err
			}
			c[id] = &cap
		}
	} else if len(path) > 1 {
		return patch.AddCompositeToMap(
			path,
			value,
			c)
	} else {
		return fmt.Errorf("invalid path for capabilities: %v", path)
	}

	return nil
}

func (c CapabilityList) Remove(path []string) error {
	if len(path) != 1 {
		return fmt.Errorf("invalid path for capabilities: %v", path)
	}

	id := path[0]
	_, exist := c[id]
	if !exist {
		return fmt.Errorf("no capability to remove exists: %v", id)
	}
	delete(c, id)

	return nil
}

func (*CapabilityList) Update(o patch.Patchable) error {
	return fmt.Errorf("not supported update for CapabilityList")
}

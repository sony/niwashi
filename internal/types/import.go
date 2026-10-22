// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package types

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Imports map[string]*Import

func (i Imports) Clone() Imports {
	if i == nil {
		return nil
	}
	clone := make(Imports, len(i))
	for k, v := range i {
		clone[k] = v.Clone()
	}
	return clone
}

func (i Imports) Merge(other Imports) Imports {
	if i == nil {
		return other.Clone()
	}
	if other == nil {
		return i.Clone()
	}

	merged := i.Clone()
	for k, v := range other {
		merged[k] = v.Clone()
	}
	return merged
}

func (i Imports) Validate() error {
	for k, v := range i {
		if v == nil {
			return fmt.Errorf("import %q: value is empty", k)
		}
		if e := v.Validate(); e != nil {
			return fmt.Errorf("import %q: %w", k, e)
		}
	}
	return nil
}

type Import struct {
	From     string `json:"from,omitempty" yaml:"from,omitempty"`
	Optional bool   `json:"optional,omitempty" yaml:"optional,omitempty"`
}

func (i *Import) Clone() *Import {
	if i == nil {
		return nil
	}
	return &Import{
		From:     i.From,
		Optional: i.Optional,
	}
}

func (i *Import) Validate() error {
	if i == nil {
		return nil
	}
	if i.From == "" {
		return fmt.Errorf("from is required")
	}
	return nil
}

func (i *Import) UnmarshalYAML(node *yaml.Node) error {

	switch node.Kind {
	case yaml.ScalarNode:
		var from string
		if err := node.Decode(&from); err != nil {
			return err
		}
		*i = Import{From: from}
		return nil

	case yaml.MappingNode:
		for k := 0; k < len(node.Content); k += 2 {
			switch key := node.Content[k].Value; key {
			case "from", "optional":
			default:
				return fmt.Errorf("unknown key %q in import", key)
			}
		}

		type plain Import
		var p plain
		if err := node.Decode(&p); err != nil {
			return err
		}
		*i = Import(p)
		return nil
	default:
		return fmt.Errorf("line %d: import must be an export name or a mapping", node.Line)
	}
}

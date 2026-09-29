// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type RequireList []*Require

type Require struct {
	Name string `json:"name" yaml:"name"`
	As   string `json:"as,omitempty" yaml:"as,omitempty"`
}

func (c *RequireList) UnmarshalYAML(node *yaml.Node) error {
	if *c == nil {
		*c = RequireList{}
	}

	switch node.Kind {
	case yaml.SequenceNode:
		// list of requires
		for _, item := range node.Content {
			switch item.Kind {
			case yaml.ScalarNode:

				// ID as string (not yet normalized)
				*c = append(*c, &Require{Name: item.Value})

			case yaml.MappingNode:
				// Require as map

				var cap Require
				if err := item.Decode(&cap); err != nil {
					return err
				}
				*c = append(*c, &cap)
			default:
				return fmt.Errorf("invalid require format")
			}
		}

	default:
		return fmt.Errorf("requires must be a sequence")
	}

	return nil
}

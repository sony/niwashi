// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cluster

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/clone"
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/types"
	"gopkg.in/yaml.v3"
)

type NodeList map[string]*Node

type Node struct {
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

func (nl *NodeList) Clone() NodeList {
	if nl == nil || *nl == nil {
		return nil
	}
	newList := NodeList{}
	for id, node := range *nl {
		newList[id] = &Node{
			Labels: clone.StringMap(node.Labels),
		}
	}
	return newList
}

func (nl *NodeList) Validate() error {
	var err error
	for _, node := range *nl {
		if node == nil {
			continue
		}
		if e := node.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid node: %w", e))
		}
	}
	return err
}

func (nl *NodeList) Merge(other NodeList) NodeList {
	if len(other) == 0 {
		return nl.Clone()
	}
	if len(*nl) == 0 {
		return other.Clone()
	}

	nodeMap := nl.Clone()
	for id, node := range other {
		if existingNode, exist := nodeMap[id]; exist {
			existingNode.Labels = merge.StringMap(existingNode.Labels, node.Labels)
		} else {
			nodeMap[id] = &Node{
				Labels: clone.StringMap(node.Labels),
			}
		}
	}

	return nodeMap
}

func (nl *NodeList) UnmarshalYAML(node *yaml.Node) error {
	if *nl == nil {
		*nl = NodeList{}
	}

	switch node.Kind {
	case yaml.SequenceNode:
		// list of node names
		for _, item := range node.Content {
			switch item.Kind {
			case yaml.ScalarNode:

				// ID as string
				(*nl)[item.Value] = &Node{}

			default:
				return fmt.Errorf("invalid nodes format")
			}
		}

	case yaml.MappingNode:
		// map of node
		var nodeMap map[string]*Node
		if err := node.Decode(&nodeMap); err != nil {
			return err
		}

		for id, n := range nodeMap {
			(*nl)[id] = n
		}

	default:
		return fmt.Errorf("nodes must be a sequence or mapping")
	}

	return nil
}

func (nl *Node) Validate() error {
	var err error
	for k := range nl.Labels {
		if !types.LabelKeyPattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid label key: %q", k))
		}
	}
	return err
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package runtime

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/merge"
	"github.com/sony/niwashi/internal/platform"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/transport"
)

// State represents the runtime state.
// - Reflects bound instance information onto nodes
// - Merges labels of nodes targeted for cluster target
// - Adds information about the execution host
// - Removes unnecessary information
// - Adds Runtime structure
//
// This structure is used for:
// - context.json for recipe execution

type State struct {
	Host      *Node          `json:"host,omitempty" yaml:"host,omitempty"`
	Inventory *Inventory     `json:"inventory,omitempty" yaml:"inventory,omitempty"`
	Runtime   *state.Runtime `json:"runtime,omitempty" yaml:"runtime,omitempty"`
}

type Inventory struct {
	Nodes    map[string]*Node    `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	Clusters map[string]*Cluster `json:"clusters,omitempty" yaml:"clusters,omitempty"`
}

// dynamic data only
type Node struct {
	Labels     map[string]string     `json:"labels,omitempty" yaml:"labels,omitempty"`
	Os         string                `json:"os,omitempty" yaml:"os,omitempty"`
	Arch       string                `json:"arch,omitempty" yaml:"arch,omitempty"`
	SystemInfo *platform.SystemInfo  `json:"systemInfo,omitempty" yaml:"systemInfo,omitempty"`
	Addresses  []string              `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Connection *transport.Connection `json:"connection,omitempty" yaml:"connection,omitempty"`
}

type Cluster cluster.Cluster

func NewStateFrom(s *state.Accessor, nodes []string, clusters []string) (*State, error) {
	context := &State{
		Host: &Node{
			Os:   platform.HostOs,
			Arch: platform.HostArch,
		},
		Inventory: &Inventory{
			Nodes:    make(map[string]*Node, len(nodes)),
			Clusters: make(map[string]*Cluster, len(clusters)),
		},
		Runtime: s.Runtime,
	}

	// copy node
	for _, nodeName := range nodes {
		if s.Inventory == nil {
			return nil, fmt.Errorf("inventory is required in state for node %s", nodeName)
		}

		inst := s.GetBoundInstance(nodeName)
		if inst == nil {
			return nil, fmt.Errorf("instance for node %s not found", nodeName)
		}

		n, ok := s.Inventory.Nodes[nodeName]
		if !ok {
			return nil, fmt.Errorf("node %s not found in inventory", nodeName)
		}
		node := &Node{
			Labels:     n.Labels,
			Addresses:  inst.Addresses,
			Connection: inst.Connection,
		}

		if inst.SystemInfo != nil {
			node.Os = inst.SystemInfo.Os
			node.Arch = inst.SystemInfo.Arch
			node.SystemInfo = inst.SystemInfo
		}

		if node.Os == "" {
			logger.Warn("no os information in instance for node, using unknown os", "node", nodeName)
			node.Os = platform.OsUnknown
		}
		if node.Arch == "" {
			logger.Warn("no arch information in instance for node, using unknown arch", "node", nodeName)
			node.Arch = platform.ArchUnknown
		}

		context.Inventory.Nodes[nodeName] = node
	}

	// copy cluster
	for _, clusterName := range clusters {
		if s.Inventory == nil {
			return nil, fmt.Errorf("inventory is required in state for cluster %s", clusterName)
		}

		c, ok := s.Inventory.Clusters[clusterName]
		if !ok {
			return nil, fmt.Errorf("cluster %s not found in inventory", clusterName)
		}
		context.Inventory.Clusters[clusterName] = (*Cluster)(c)

		// merge labels
		for nodeName, n := range c.Nodes {
			if n.Labels == nil {
				continue
			}

			node, ok := context.Inventory.Nodes[nodeName]
			if !ok {
				return nil, fmt.Errorf("node %s in cluster %s not found in inventory", nodeName, clusterName)
			}

			node.Labels = merge.StringMap(node.Labels, n.Labels)
		}
	}

	return context, nil
}

func (s *State) WriteToFile(p string, fs file.Creator) error {
	err := file.WriteWithEncoding(s, p, fs)
	if err != nil {
		return errors.Join(err, fmt.Errorf("fail to write state to: %q", p))
	}

	return nil
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"fmt"
	"maps"
	"slices"

	"github.com/Masterminds/semver/v3"
	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/node"
)

type Diff struct {
	Construct *Difference `json:"construct,omitempty" yaml:"construct,omitempty"`
	Update    *Difference `json:"update,omitempty" yaml:"update,omitempty"`
	Destruct  *Difference `json:"destruct,omitempty" yaml:"destruct,omitempty"`
}

type Difference struct {
	Instance       []string       `json:"instance,omitempty" yaml:"instance,omitempty"`
	Node           []string       `json:"node,omitempty" yaml:"node,omitempty"`
	Cluster        []string       `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Capability     CapabilityDiff `json:"capability,omitempty" yaml:"capability,omitempty"`
	BaseVertexData map[string]any `json:"-" yaml:"-"`
}

type CapabilityDiff struct {
	Node    map[string][]string `json:"node,omitempty" yaml:"node,omitempty"`
	Cluster map[string][]string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
}

func keys[T any](m map[string]T) []string {
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func capKeys(m cap.CapabilityList) []string {
	return slices.Collect(maps.Keys(m))
}

// compute difference: a - b
func difference(a, b []string) []string {
	result := []string{}
	bSet := make(map[string]struct{}, len(b))
	for _, v := range b {
		bSet[v] = struct{}{}
	}

	for _, v := range a {
		if _, exists := bSet[v]; !exists {
			result = append(result, v)
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// compute node capability difference: a - b
func differenceCap[T *node.Node | *cluster.Cluster](a, b map[string]T, getCaps func(T) cap.CapabilityList) map[string][]string {
	result := make(map[string][]string)
	for nodeName, nodeA := range a {
		capA := capKeys(getCaps(nodeA))
		nodeB, exist := b[nodeName]
		if !exist {
			// add all capabilities
			result[nodeName] = capA
			continue
		}
		capB := capKeys(getCaps(nodeB))
		diff := difference(capA, capB)
		if len(diff) > 0 {
			result[nodeName] = diff
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// compute node capability difference for changed params : a - b
func differenceCapChange[T *node.Node | *cluster.Cluster](a, b map[string]T, getCaps func(T) cap.CapabilityList) (map[string][]string, error) {
	// a->b
	// recipe version: check semver
	// params: compute hash and compare for changes

	result := make(map[string][]string)
	for name, targetA := range a {
		// b exists?
		targetB, exist := b[name]
		if !exist {
			// node does not exist in b
			continue
		}

		for capName, capA := range getCaps(targetA) {
			// capB exists?
			capB, exist := getCaps(targetB)[capName]
			if !exist {
				// capability does not exist in b
				continue
			}

			versionA, err := semver.NewVersion(capA.Version)
			if err != nil {
				return nil, err
			}
			versionB, err := semver.NewVersion(capB.Version)
			if err != nil {
				return nil, err
			}

			// compare semver of the capability
			// - newer version, then store the difference
			// - same version, check for parameter hash changes
			// - older version, then return error

			if versionB.GreaterThan(versionA) {
				// Version B is greater than version A. means "Upgrade"
				// e.g., versionA = 1.0.0, versionB = 1.1.0
				// newer version, then store the difference
				result[name] = append(result[name], capName)
			} else if versionB.Equal(versionA) {
				// same version, check for parameter changes
				if capA.Params.Hash() != capB.Params.Hash() {
					result[name] = append(result[name], capName)
				}
			} else {
				// This means "Downgrade"
				// older version, then return error
				return nil, fmt.Errorf("capability %s of %s has older version in target", capName, name)
			}
		}
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

func newAppearanceDifference(a, b *State) *Difference {

	// compute appearance difference between a and b

	return &Difference{
		Instance: difference(keys(a.Infrastructure.Generators), keys(b.Infrastructure.Generators)),
		Node:     difference(keys(a.Inventory.Nodes), keys(b.Inventory.Nodes)),
		Cluster:  difference(keys(a.Inventory.Clusters), keys(b.Inventory.Clusters)),
		Capability: CapabilityDiff{
			Node: differenceCap(a.Inventory.Nodes, b.Inventory.Nodes, func(n *node.Node) cap.CapabilityList {
				return n.Capabilities
			}),
			Cluster: differenceCap(a.Inventory.Clusters, b.Inventory.Clusters, func(n *cluster.Cluster) cap.CapabilityList {
				return n.Capabilities
			}),
		},
	}
}

func newUpdateDifference(initial, target *State) (*Difference, error) {
	// compute the following changes:
	// - recipe version changes
	// - params changes

	diff := &Difference{
		BaseVertexData: map[string]any{
			"update": true,
		},
	}

	// Check Capability params
	capDiff, err := differenceCapChange(initial.Inventory.Nodes, target.Inventory.Nodes, func(n *node.Node) cap.CapabilityList {
		return n.Capabilities
	})
	if err != nil {
		return nil, err
	}
	diff.Capability.Node = capDiff

	capDiff, err = differenceCapChange(initial.Inventory.Clusters, target.Inventory.Clusters, func(n *cluster.Cluster) cap.CapabilityList {
		return n.Capabilities
	})
	if err != nil {
		return nil, err
	}
	diff.Capability.Cluster = capDiff

	return diff, nil
}

func newEmptyState() *State {
	return &State{
		Inventory: &Inventory{
			Nodes:    nil,
			Clusters: nil,
		},
		Infrastructure: &Infrastructure{
			Generators: nil,
		},
	}
}

func normalizeForDiff(s *State) *State {
	ret := newEmptyState()
	if s == nil {
		return ret
	}

	// copy minimal reference
	if s.Inventory != nil {
		ret.Inventory = s.Inventory
	}
	if s.Infrastructure != nil {
		ret.Infrastructure = s.Infrastructure
	}

	return ret
}

func ComputeConstructDiff(initial, target *State) *Difference {

	i := normalizeForDiff(initial)
	t := normalizeForDiff(target)

	return newAppearanceDifference(t, i)
}

func ComputeUpdateDiff(initial, target *State) (*Difference, error) {
	i := normalizeForDiff(initial)
	t := normalizeForDiff(target)

	return newUpdateDifference(i, t)
}

func ComputeDestructDiff(initial, target *State) *Difference {

	i := normalizeForDiff(initial)
	t := normalizeForDiff(target)

	return newAppearanceDifference(i, t)
}

func (d *Diff) String() string {
	return fmt.Sprintf("Construct: %+v\nDestruct: %+v", d.Construct, d.Destruct)
}

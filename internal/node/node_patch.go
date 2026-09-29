// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package node

import (
	"fmt"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/patch"
)

func (n *Node) Add(path []string, value any) error {

	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "instanceRef":
		return patch.AddSimpleValue(path[1:], value, func(v string) {
			n.InstanceRef = v
		})
	case "capabilities":
		return patch.AddCompositeValue(
			path[1:],
			value,
			func() (*cap.CapabilityList, bool) {
				return &n.Capabilities, true
			},
			func(v *cap.CapabilityList) {
				n.Capabilities = *v
			})
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (n *Node) Remove(path []string) error {

	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "instanceRef":
		return patch.RemoveSimpleValue(path[1:], func() {
			n.InstanceRef = ""
		})
	case "capabilities":
		return patch.RemoveComposite(
			path[1:],
			func() (*cap.CapabilityList, bool) {
				return &n.Capabilities, true
			},
			func() {
				n.Capabilities = cap.CapabilityList{}
			})
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (*Node) Update(o patch.Patchable) error {
	return fmt.Errorf("not supported update for Node")
}

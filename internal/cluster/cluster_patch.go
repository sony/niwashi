// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cluster

import (
	"fmt"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/patch"
)

func (c *Cluster) Add(path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	switch path[0] {
	case "capabilities":
		return patch.AddCompositeValue(
			path[1:],
			value,
			func() (*cap.CapabilityList, bool) {
				return &c.Capabilities, true
			},
			func(v *cap.CapabilityList) {
				c.Capabilities = *v
			})
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (c *Cluster) Remove(path []string) error {
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}
	switch path[0] {
	case "capabilities":
		return c.Capabilities.Remove(path[1:])
	default:
		return fmt.Errorf("unsupported path: %s", path[0])
	}
}

func (*Cluster) Update(o patch.Patchable) error {
	return fmt.Errorf("not supported update for Cluster")
}

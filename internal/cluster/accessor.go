// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cluster

type Accessor struct {
	*Cluster
	Name string
}

func NewAccessor(name string, c *Cluster) *Accessor {
	if c == nil {
		return nil
	}
	return &Accessor{
		Name:    name,
		Cluster: c,
	}
}

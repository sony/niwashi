// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package node

type Accessor struct {
	*Node
	Name string
}

func NewAccessor(name string, n *Node) *Accessor {
	if n == nil {
		return nil
	}
	return &Accessor{
		Name: name,
		Node: n,
	}
}

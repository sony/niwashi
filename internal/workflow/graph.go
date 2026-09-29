// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

type Graph interface {
	Reset()
	GetVertices() []Vertex
	AddVertex(vertex Vertex)
	HasVertex(id string) bool
	AddEdge(from Vertex, to Vertex)
	RemoveVertex(id string)
	String() string
	TopologicalSort() ([]Vertex, error)
	SwapOrder()
}

type Vertex interface {
	GetId() string
	GetData() any
}

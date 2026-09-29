// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import (
	"fmt"
	"strings"
)

type DagGraph struct {
	Vertices map[string]Vertex
	Edges    []*Edge
}

type Edge struct {
	From Vertex
	To   Vertex
}

func NewDagGraph() *DagGraph {
	return &DagGraph{
		Vertices: make(map[string]Vertex),
		Edges:    []*Edge{},
	}
}

func (d *DagGraph) Reset() {
	d.Vertices = make(map[string]Vertex)
	d.Edges = []*Edge{}
}

func (d *DagGraph) AddVertex(vertex Vertex) {
	d.Vertices[vertex.GetId()] = vertex
}

func (d *DagGraph) GetVertices() []Vertex {
	vertices := make([]Vertex, 0, len(d.Vertices))
	for _, v := range d.Vertices {
		vertices = append(vertices, v)
	}
	return vertices
}

func (d *DagGraph) AddEdge(from Vertex, to Vertex) {
	d.Edges = append(d.Edges, &Edge{From: from, To: to})
}

func (d *DagGraph) HasVertex(id string) bool {
	_, exist := d.Vertices[id]
	return exist
}

func (d *DagGraph) RemoveVertex(id string) {
	delete(d.Vertices, id)

	// also remove edges which links to/from this vertex
	var newEdges []*Edge
	for _, edge := range d.Edges {
		if edge.From.GetId() != id && edge.To.GetId() != id {
			newEdges = append(newEdges, edge)
		}
	}
	d.Edges = newEdges
}

func (d *DagGraph) String() string {
	var sb strings.Builder
	sb.WriteString("DagGraph:\n")
	sb.WriteString("Vertices:\n")
	for _, v := range d.Vertices {
		sb.WriteString("  - ID: " + v.GetId() + "\n")
	}
	sb.WriteString("Edges:\n")
	for _, e := range d.Edges {
		sb.WriteString("  - From: " + e.From.GetId() + " -> To: " + e.To.GetId() + "\n")
	}
	return sb.String()
}

func (d *DagGraph) TopologicalSort() ([]Vertex, error) {
	inDegree := make(map[string]int)
	for id := range d.Vertices {
		inDegree[id] = 0
	}

	for _, edge := range d.Edges {
		inDegree[edge.To.GetId()]++
	}

	var queue []Vertex
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, d.Vertices[id])
		}
	}

	var sorted []Vertex
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		sorted = append(sorted, v)

		for _, edge := range d.Edges {
			if edge.From.GetId() == v.GetId() {
				inDegree[edge.To.GetId()]--
				if inDegree[edge.To.GetId()] == 0 {
					queue = append(queue, edge.To)
				}
			}
		}
	}

	if len(sorted) != len(d.Vertices) {
		return nil, fmt.Errorf("graph has at least one cycle")
	}

	return sorted, nil
}

func (d *DagGraph) SwapOrder() {
	for _, edge := range d.Edges {
		edge.From, edge.To = edge.To, edge.From
	}
}

func (d *DagGraph) VerticesSize() int {
	return len(d.Vertices)
}

func (d *DagGraph) FindVerticesWithoutIncomingEdges() []Vertex {
	var noIncoming []Vertex
	for _, v := range d.Vertices {
		hasIncoming := false
		for _, edge := range d.Edges {
			if edge.To.GetId() == v.GetId() {
				hasIncoming = true
				break
			}
		}
		if !hasIncoming {
			noIncoming = append(noIncoming, v)
		}
	}

	return noIncoming
}

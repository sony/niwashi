// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_barrier

import (
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
)

func init() {
	workflow.AddPhase(workflow.PhaseBarrier, workflow.PhaseBarrierPriority, nil)

	planner.RegisterPhasePlanner(workflow.PhaseBarrier, &BarrierPlanner{
		prePhases: make(map[string]*BarrierDagVertex),
	})
}

type BarrierDagVertex struct {
	id    string
	phase string
}

func NewBarrierDagVertex(phase string) *BarrierDagVertex {
	return &BarrierDagVertex{
		id:    workflow.MakeId(workflow.PhaseBarrier, phase),
		phase: phase,
	}
}

func (v *BarrierDagVertex) GetId() string {
	return v.id
}

func (v *BarrierDagVertex) GetData() any {
	return nil
}

type BarrierPlanner struct {
	prePhases map[string]*BarrierDagVertex
}

func (p *BarrierPlanner) BuildVertices(ctx planner.Context) error {
	// add barrier vertices for each phase

	var before *BarrierDagVertex
	for _, phase := range workflow.GetOrderedPhase() {
		if phase == workflow.PhaseBarrier {
			continue
		}
		// Add barrier vertex for the phase
		v := NewBarrierDagVertex(phase)
		ctx.AddVertex(v)
		p.prePhases[phase] = before
		before = v
	}

	return nil
}

func (p *BarrierPlanner) BuildEdges(ctx planner.Context, vertex workflow.Vertex) error {

	// example)
	// recipe:<host_recipe> -> barrier:host
	// recipe:<infra_recipe> -> barrier:infra

	vert := vertex.(*BarrierDagVertex)
	pre := p.prePhases[vert.phase]

	// make links:
	// - pre -> vert
	// - pre -> v -> vert

	somelink := false
	g := ctx.Graph()
	for _, v := range g.Vertices {
		vPhase, _, _ := workflow.ParseId(v.GetId())
		if vPhase != vert.phase {
			// skip vertices of other phases
			continue
		}

		// pre -> v if exists
		if pre != nil {
			ctx.AddEdge(pre, v)
		}

		somelink = true

		// v -> barrier
		ctx.AddEdge(v, vert)
	}

	if !somelink && pre != nil {
		ctx.AddEdge(pre, vert)
	}

	return nil
}

func (p *BarrierPlanner) CanOmit() bool {
	return true
}

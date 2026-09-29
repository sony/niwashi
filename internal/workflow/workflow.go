// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import (
	"fmt"
	"strings"

	"github.com/sony/niwashi/internal/plan"
)

type Job plan.Job

const (
	IdSeparator = ":"
)

func BuildWorkflow(dag *DagGraph) (*plan.Workflow, error) {

	vtx, err := dag.TopologicalSort()
	if err != nil {
		return nil, err
	}

	wf := &plan.Workflow{
		Jobs:  []*plan.Job{},
		Links: []*plan.Link{},
	}

	for _, v := range vtx {
		job := &plan.Job{
			Id:   v.GetId(),
			Data: v.GetData(),
		}
		wf.Jobs = append(wf.Jobs, job)
	}

	for _, e := range dag.Edges {
		link := &plan.Link{
			From: e.From.GetId(),
			To:   e.To.GetId(),
		}
		wf.Links = append(wf.Links, link)
	}

	return wf, nil
}

func MakeId(phase string, options ...string) string {
	return phase + IdSeparator + strings.Join(options, IdSeparator)
}

func ParseId(id string) (phase string, options []string, err error) {
	parts := strings.Split(id, IdSeparator)
	if len(parts) < 1 {
		return "", nil, fmt.Errorf("invalid id: %s", id)
	}
	return parts[0], parts[1:], nil
}

func (j *Job) GetPhase() string {
	phase, _, err := ParseId(j.Id)
	if err != nil {
		return ""
	}
	return phase
}

func (j *Job) GetOptions() []string {
	_, options, err := ParseId(j.Id)
	if err != nil {
		return nil
	}
	return options
}

func (j *Job) GetFromData(key string) (any, bool) {
	m, ok := j.Data.(map[string]any)
	if !ok {
		return nil, false
	}
	v, ok := m[key]
	return v, ok
}

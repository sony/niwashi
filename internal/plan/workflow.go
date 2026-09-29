// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package plan

type Workflow struct {
	Jobs  []*Job  `json:"jobs,omitempty" yaml:"jobs,omitempty"`
	Links []*Link `json:"links,omitempty" yaml:"links,omitempty"`
}

type Job struct {
	Id   string `json:"id,omitempty" yaml:"id,omitempty"`
	Data any    `json:"data,omitempty" yaml:"data,omitempty"`
}

type Link struct {
	From string `json:"from,omitempty" yaml:"from,omitempty"`
	To   string `json:"to,omitempty" yaml:"to,omitempty"`
}

func WithNoIncomingLinks(w *Workflow, j *Job) bool {
	for _, link := range w.Links {
		if link.To == j.Id {
			return false
		}
	}
	return true
}

func (w *Workflow) FindJob(fn func(*Workflow, *Job) bool) (*Job, bool) {
	for _, job := range w.Jobs {
		if fn(w, job) {
			return job, true
		}
	}
	return nil, false
}

func (w *Workflow) FindJobs(fn func(*Workflow, *Job) bool) []*Job {
	result := []*Job{}
	for _, job := range w.Jobs {
		if fn(w, job) {
			result = append(result, job)
		}
	}
	return result
}

func (w *Workflow) RemoveJob(id string) {
	newJobs := []*Job{}
	for _, job := range w.Jobs {
		if job.Id != id {
			newJobs = append(newJobs, job)
		}
	}
	w.Jobs = newJobs
}

func (w *Workflow) RemoveEdgesWithFrom(fromId string) {
	newLinks := []*Link{}
	for _, link := range w.Links {
		if link.From != fromId {
			newLinks = append(newLinks, link)
		}
	}
	w.Links = newLinks
}

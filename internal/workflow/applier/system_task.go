// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/workflow"
)

type SystemTask interface {
	GetName() string
	GetSubject() string
	GetAffectSubjects() []string
	GetJob() *workflow.Job
	GetDryRunMode() string
	GetFileSystem() file.FileSystem
}

type systemTask struct {
	Name           string
	Job            *workflow.Job
	Subject        string
	AffectSubjects []string
	DryRunMode     string
	FileSystem     file.FileSystem
}

func NewSystemTask(name string, job *workflow.Job, opts ...func(*systemTask)) *systemTask {
	t := &systemTask{
		Name: name,
		Job:  job,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func WithSubject(subject string, affectSubjects []string) func(*systemTask) {
	return func(t *systemTask) {
		t.Subject = subject
		t.AffectSubjects = affectSubjects
	}
}

func WithDryRunMode(mode string) func(*systemTask) {
	return func(t *systemTask) {
		t.DryRunMode = mode
	}
}

func WithFileSystem(fs file.FileSystem) func(*systemTask) {
	return func(t *systemTask) {
		t.FileSystem = fs
	}
}

func (t *systemTask) GetName() string {
	return t.Name
}

func (t *systemTask) GetSubject() string {
	if t.Subject == "" {
		return workflow.SubjectSystemPrefix + "system"
	}
	return t.Subject
}

func (t *systemTask) GetAffectSubjects() []string {
	return t.AffectSubjects
}

func (t *systemTask) GetJob() *workflow.Job {
	return t.Job
}

func (t *systemTask) GetDryRunMode() string {
	return t.DryRunMode
}

func (t *systemTask) GetFileSystem() file.FileSystem {
	return t.FileSystem
}

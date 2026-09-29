// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/file"
)

type Spec struct {
	Category  string      `json:"category" yaml:"category"`
	Provides  []*Provide  `json:"provides" yaml:"provides"`
	Requires  RequireList `json:"requires" yaml:"requires"`
	Defaults  *Defaults   `json:"defaults" yaml:"defaults"`
	Tasks     []*Task     `json:"tasks" yaml:"tasks"`
	Assets    []string    `json:"assets" yaml:"assets"`
	Workspace *Workspace  `json:"workspace" yaml:"workspace"`
	Runtime   *Runtime    `json:"runtime" yaml:"runtime"`

	// for adapter
	Commands      map[string]*Command `json:"commands" yaml:"commands"`
	ExecutionUnit string              `json:"executionUnit" yaml:"executionUnit"`
	AllowedScope  string              `json:"allowedScope" yaml:"allowedScope"`
}

func (s *Spec) SetDefaults() {
	for _, t := range s.Tasks {
		t.SetDefaults()
	}
	if s.Workspace == nil {
		s.Workspace = &Workspace{}
	}
	s.Workspace.SetDefaults()
}

func (s *Spec) validateBase() error {
	var err error
	if e := validateTasks(s.Tasks); e != nil {
		err = errors.Join(err, e)
	}

	if s.Defaults != nil {
		if e := s.Defaults.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("invalid defaults: %w", e))
		}
	}

	if s.Provides != nil {
		for i, p := range s.Provides {
			if e := p.Validate(); e != nil {
				err = errors.Join(err, fmt.Errorf("invalid provide at index %d: %w", i, e))
			}
		}
	}

	return err
}

func (s *Spec) Validate() error {

	err := s.validateBase()

	// kind: host
	if e := s.Runtime.Validate(); e != nil {
		err = errors.Join(err, fmt.Errorf("invalid runtime: %w", e))
	}

	if len(s.Commands) > 0 {
		err = errors.Join(err, fmt.Errorf("commands is not allowed for non-adapter recipe"))
	}
	if s.ExecutionUnit != "" {
		err = errors.Join(err, fmt.Errorf("executionUnit is not allowed for non-adapter recipe"))
	}
	if s.AllowedScope != "" {
		err = errors.Join(err, fmt.Errorf("allowedScope is not allowed for non-adapter recipe"))
	}
	return err
}

var validOperations = []string{OperationConstruct, OperationDestruct, OperationUpdate}

func validateTasks(tasks []*Task) error {
	if len(tasks) == 0 {
		return fmt.Errorf("at least one task is required")
	}

	var err error
	names := make(map[string]struct{})
	for _, t := range tasks {

		if e := t.Validate(); e != nil {
			err = errors.Join(err, e)
		}

		if t.Name != "" {
			n := file.SanitizeFilename(t.Name)
			if _, exists := names[n]; exists {
				err = errors.Join(err, fmt.Errorf("duplicate task name: %q", t.Name))
			}
			names[n] = struct{}{}
		}
	}

	return err
}

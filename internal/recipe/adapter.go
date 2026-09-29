// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/sony/niwashi/internal/logger"
)

type Adapter Spec

func (a *Adapter) SetDefaults() {
	if a.ExecutionUnit == "" {
		a.ExecutionUnit = "default"
	}
}

var allowdScopes = []string{
	KindHost,
	KindInfra,
	KindNode,
	KindCluster,
	"*", // wildcard
}

var executionUnits = []string{
	"default",
	"node",
}

func (a *Adapter) Validate() error {
	if a == nil {
		return fmt.Errorf("nil adapter")
	}

	s := (*Spec)(a)
	err := s.validateBase()

	// where check
	for _, t := range a.Tasks {
		if t.Where != "" {
			logger.Warn("where condition is not supported for adapter task, it will be ignored", "task", t.Name)
		}
	}

	if len(a.Commands) == 0 {
		err = errors.Join(err, fmt.Errorf("at least one command is required for adapter"))
	} else {
		for cmdName, c := range a.Commands {

			if !commandNamePattern.MatchString(cmdName) {
				err = errors.Join(err, fmt.Errorf("invalid command name %q: must match pattern %q", cmdName, commandNamePatternString))
			}

			if e := c.Validate(); e != nil {
				err = errors.Join(err, fmt.Errorf("invalid command %q: %w", cmdName, e))
			}
		}
	}

	if a.AllowedScope == "" {
		err = errors.Join(err, fmt.Errorf("allowedScope is required for adapter"))
	} else if !slices.Contains(allowdScopes, a.AllowedScope) {
		err = errors.Join(err, fmt.Errorf("invalid allowed scope: %q, allowed scope must be one of %q", a.AllowedScope, strings.Join(allowdScopes, ", ")))
	}

	if !slices.Contains(executionUnits, a.ExecutionUnit) {
		err = errors.Join(err, fmt.Errorf("invalid execution unit: %q", a.ExecutionUnit))
	} else if a.ExecutionUnit == "node" && a.AllowedScope != KindCluster {
		err = errors.Join(err, fmt.Errorf("invalid allowed scope: %q for adapter with execution unit %q, allowed scope must be \"cluster\"", a.AllowedScope, a.ExecutionUnit))
	}

	return err
}

func (a *Adapter) FindTask(name string) *Task {
	for _, t := range a.Tasks {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func (a *Adapter) FindTaskByCommand(cmdName string) (*Task, error) {
	if a == nil {
		return nil, fmt.Errorf("adapter is nil")
	}
	c, ok := a.Commands[cmdName]
	if !ok {
		return nil, fmt.Errorf("command %q not found", cmdName)
	}
	for _, t := range a.Tasks {
		if t.Name == c.Task {
			return t, nil
		}
	}
	return nil, fmt.Errorf("task %q not found for command %q", c.Task, cmdName)
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"errors"
	"fmt"
)

type Command struct {
	Task        string      `json:"task" yaml:"task"`
	Candidates  []Candidate `json:"candidates" yaml:"candidates"`
	Description string      `json:"description" yaml:"description"`
}

type Candidate struct {
	Task  string `json:"name" yaml:"name"`
	Where string `json:"where" yaml:"where"`
}

func (c *Command) Validate() error {
	var err error

	if c.Task != "" {
		// specifying task name
		if !taskNamePattern.MatchString(c.Task) {
			err = errors.Join(err, fmt.Errorf("command task=%q must match pattern %q", c.Task, taskNamePatternString))
		}

		if len(c.Candidates) > 0 {
			err = errors.Join(err, fmt.Errorf("candidates cannot be specified when task is specified for command"))
		}

	} else {
		if len(c.Candidates) == 0 {
			err = errors.Join(err, fmt.Errorf("either task or candidates must be specified for command"))
		} else {

			for i, candidate := range c.Candidates {
				if !taskNamePattern.MatchString(candidate.Task) {
					err = errors.Join(err, fmt.Errorf("candidate %d: task=%q must match pattern %q", i, candidate.Task, taskNamePatternString))
				}
			}
		}
	}

	return err
}

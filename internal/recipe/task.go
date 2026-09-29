// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/sony/niwashi/internal/action"
	"gopkg.in/yaml.v3"
)

const (
	OperationConstruct = "construct"
	OperationDestruct  = "destruct"
	OperationUpdate    = "update"
)

type EnvTemplate map[string]string

type Task struct {
	Name         string                           `json:"name" yaml:"name"`
	Operation    string                           `json:"operation,omitempty" yaml:"operation,omitempty"`
	DependsOn    []string                         `json:"dependsOn" yaml:"dependsOn"`
	Where        string                           `json:"where,omitempty" yaml:"where,omitempty"`
	Action       Action                           `json:"action" yaml:"action"`
	StateChanges map[string]*StateChangeOperation `json:"stateChanges" yaml:"stateChanges"`
}

type Action struct {
	ActionType string
	Spec       action.ActionSpec
}

func (a *Task) SetDefaults() {
	if a.Operation == "" {
		a.Operation = OperationConstruct
	}
}

func (t *Task) Validate() error {
	var err error

	if len(t.Name) == 0 {
		err = errors.Join(err, fmt.Errorf("task name is required"))
	}

	if t.Action.ActionType == "" {
		err = errors.Join(err, fmt.Errorf("task %q: action is required", t.Name))
	}

	if t.Action.Spec == nil {
		err = errors.Join(err, fmt.Errorf("task %q: action spec is required", t.Name))
	} else {
		if e := t.Action.Spec.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("task %q: invalid action spec: %w", t.Name, e))
		}
	}

	if t.Operation != "" && !slices.Contains(validOperations, t.Operation) {
		err = errors.Join(err, fmt.Errorf("task %q: invalid operation %q, must be one of %s", t.Name, t.Operation, strings.Join(validOperations, ", ")))
	}

	depSeen := make(map[string]struct{})
	for _, d := range t.DependsOn {
		if _, exists := depSeen[d]; exists {
			err = errors.Join(err, fmt.Errorf("task %q: duplicate dependsOn entry %q", t.Name, d))
		}
		depSeen[d] = struct{}{}
	}

	for _, sc := range t.StateChanges {
		if e := sc.Validate(); e != nil {
			err = errors.Join(err, fmt.Errorf("task %q: invalid state change operation: %w", t.Name, e))
		}
	}

	return err
}

func (a *Action) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("action must be a mapping")
	}
	if len(node.Content) != 2 {
		return fmt.Errorf("action mapping must have exactly one key-value pair")
	}

	keyNode := node.Content[0]
	valueNode := node.Content[1]

	if keyNode.Kind != yaml.ScalarNode {
		return fmt.Errorf("action key must be a string")
	}

	as, err := NewActionSpec(keyNode.Value)
	if err != nil {
		return err
	}

	if err := valueNode.Decode(as); err != nil {
		return fmt.Errorf("failed to decode action spec for %q: %w", keyNode.Value, err)
	}

	a.ActionType = keyNode.Value
	a.Spec = as

	return nil
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/sony/niwashi/internal/platform"
)

type Condition struct {
	prog cel.Program
}

func NewCondition(cond string) (*Condition, error) {

	env, err := cel.NewEnv(
		cel.Variable("node", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("host", cel.MapType(cel.StringType, cel.AnyType)),
	)
	if err != nil {
		return nil, err
	}

	ast, issues := env.Compile(cond)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}
	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	return &Condition{
		prog: prg,
	}, nil
}

func (c *Condition) Eval(node map[string]any) (bool, error) {
	out, _, err := c.prog.Eval(map[string]any{
		"node": node,
		"host": map[string]any{
			"os":   platform.HostOs,
			"arch": platform.HostArch,
		},
	})
	if err != nil {
		return false, err
	}
	b, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("condition did not evaluate to a boolean")
	}
	return b, nil
}

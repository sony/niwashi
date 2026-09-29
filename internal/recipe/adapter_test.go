// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_test

import (
	"testing"

	"github.com/sony/niwashi/internal/recipe"
)

func newMockAdapter() *recipe.Adapter {
	a := &recipe.Adapter{
		Tasks: []*recipe.Task{
			{
				Name: "task1",
				Action: recipe.Action{
					ActionType: "mock",
					Spec:       &mockActionSpec{},
				},
			},
		},
		Commands: map[string]*recipe.Command{
			"cmd1": &recipe.Command{
				Task: "task1",
			},
		},
		AllowedScope: "host",
	}
	a.SetDefaults()
	return a
}

func TestAdapter_Validate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		a       *recipe.Adapter
		wantErr bool
	}{
		{
			name:    "valid adapter",
			a:       newMockAdapter(),
			wantErr: false,
		},
		{
			name: "nil commands",
			a: func() *recipe.Adapter {
				a := newMockAdapter()
				a.Commands = nil
				return a
			}(),
			wantErr: true,
		},
		{
			name: "invalid command name",
			a: func() *recipe.Adapter {
				a := newMockAdapter()
				a.Commands = map[string]*recipe.Command{
					"invalid name": &recipe.Command{},
				}
				return a
			}(),
			wantErr: true,
		},
		{
			name: "nil allowed scope",
			a: func() *recipe.Adapter {
				a := newMockAdapter()
				a.AllowedScope = ""
				return a
			}(),
			wantErr: true,
		},
		{
			name: "wildcard allowed scope",
			a: func() *recipe.Adapter {
				a := newMockAdapter()
				a.AllowedScope = "*"
				return a
			}(),
			wantErr: false,
		},
		{
			name: "invalid allowed scope",
			a: func() *recipe.Adapter {
				a := newMockAdapter()
				a.AllowedScope = "invalid"
				return a
			}(),
			wantErr: true,
		},
		{
			name: "invalid execution unit",
			a: func() *recipe.Adapter {
				a := newMockAdapter()
				a.ExecutionUnit = "invalid"
				return a
			}(),
			wantErr: true,
		},
		{
			name: "invalid allowed scope for node execution unit",
			a: func() *recipe.Adapter {
				a := newMockAdapter()
				a.ExecutionUnit = "node"
				a.AllowedScope = "host"
				return a
			}(),
			wantErr: true,
		},
		{
			name: "with where condition in task",
			a: func() *recipe.Adapter {
				a := newMockAdapter()
				a.Tasks = []*recipe.Task{
					{
						Name:  "task1",
						Where: "some condition",
						Action: recipe.Action{
							ActionType: "mock",
							Spec:       &mockActionSpec{},
						},
					},
				}
				return a
			}(),
			wantErr: false, // where condition is not supported but it should not cause validation error
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.a.Validate()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}

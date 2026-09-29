// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_test

import (
	"testing"

	"github.com/sony/niwashi/internal/recipe"
)

func TestCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		c       *recipe.Command
		wantErr bool
	}{
		{
			name: "valid command",
			c: &recipe.Command{
				Task: "task1",
			},
			wantErr: false,
		},
		{
			name: "valid command with candidates",
			c: &recipe.Command{
				Candidates: []recipe.Candidate{
					{Task: "task1"},
					{Task: "task2"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid task name",
			c: &recipe.Command{
				Task: "$$ invalid task name",
			},
			wantErr: true,
		},
		{
			name: "task and candidates specified",
			c: &recipe.Command{
				Task: "task1",
				Candidates: []recipe.Candidate{
					{Task: "task2"},
				},
			},
			wantErr: true,
		},
		{
			name:    "neither task nor candidates specified",
			c:       &recipe.Command{},
			wantErr: true,
		},
		{
			name: "invalid candidate task name",
			c: &recipe.Command{
				Candidates: []recipe.Candidate{
					{Task: ""},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.c.Validate()
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

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_test

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/action/exec_local"
	"github.com/sony/niwashi/internal/action/exec_remote"
	"github.com/sony/niwashi/internal/recipe"
	"gopkg.in/yaml.v3"
)

func TestAction_UnmarshalYAML(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name    string
		args    args
		want    recipe.Action
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				in: `
xxxx:
`,
			},
			want: recipe.Action{
				ActionType: "",
				Spec:       nil,
			},
			wantErr: true,
		},
		{
			name: "exec.local full",
			args: args{
				in: `
exec.local:
  scriptTpl: here is script template
  envTpl: {}
`,
			},
			want: recipe.Action{
				ActionType: "exec.local",
				Spec: &exec_local.Spec{
					ScriptTemplate: "here is script template",
					EnvTemplate:    map[string]string{},
				},
			},
			wantErr: false,
		},
		{
			name: "exec.local full",
			args: args{
				in: `
exec.remote:
  scriptTpl: here is script template
  envTpl: {}
`,
			},
			want: recipe.Action{
				ActionType: "exec.remote",
				Spec: &exec_remote.Spec{
					ScriptTemplate: "here is script template",
					EnvTemplate:    map[string]string{},
				},
			},
			wantErr: false,
		},
		{
			name: "tool.run full",
			args: args{
				in: `
tool.run:
  toolRef: abcde
  command: start
  argvTpl:
    - a
    - b
    - c
  envTpl:
    TEST: 0
    TEST2: true
`,
			},
			want: recipe.Action{
				ActionType: "tool.run",
				Spec: &recipe.ToolRunSpec{
					ToolRef:      "abcde",
					Command:      "start",
					ArgvTemplate: []string{"a", "b", "c"},
					EnvTemplate: map[string]string{
						"TEST":  "0",
						"TEST2": "true",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//dt := task{}
			dt := recipe.Action{}
			if err := yaml.Unmarshal([]byte(tt.args.in), &dt); (err != nil) != tt.wantErr {
				t.Errorf("Action.UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(dt, tt.want) {
				t.Errorf("Action.UnmarshalYAML() got = %v, want %v", dt, tt.want)
			}
		})
	}
}

func TestTask_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    recipe.Task
		wantErr bool
	}{
		{
			name: "test",
			args: `
name: test task
dependsOn: [xxx,yyy]
action:
  tool.run:
    toolRef: abcde
    command: start
    argvTpl:
      - a
      - b
      - c
    envTpl:
      TEST: 0
      TEST2: true
stateChanges:
  add-tool-info:
    op: set
    path: /runtime/tool/git
    valueFromFile: "{{ .Outputs.git_info }}"
`,
			want: recipe.Task{
				Name:      "test task",
				DependsOn: []string{"xxx", "yyy"},
				Action: recipe.Action{
					ActionType: "tool.run",
					Spec: &recipe.ToolRunSpec{
						ToolRef:      "abcde",
						Command:      "start",
						ArgvTemplate: []string{"a", "b", "c"},
						EnvTemplate: map[string]string{
							"TEST":  "0",
							"TEST2": "true",
						},
					},
				},
				StateChanges: map[string]*recipe.StateChangeOperation{
					"add-tool-info": {
						Op:            recipe.OpSet,
						Path:          "/runtime/tool/git",
						ValueFromFile: "{{ .Outputs.git_info }}",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := recipe.Task{}
			if err := yaml.Unmarshal([]byte(tt.args), &dt); (err != nil) != tt.wantErr {
				t.Errorf("Action.UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(dt, tt.want) {
				t.Errorf("Action.UnmarshalYAML() got = %v, want %v", dt, tt.want)
			}
		})
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package action

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sony/niwashi/internal/types"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		input   string
		val     any
		want    string
		wantErr bool
	}{
		{
			name:  "case 1: normal",
			input: `Hello, {{ .Name }}!`,
			val: map[string]string{
				"Name": "World",
			},
			want:    `Hello, World!`,
			wantErr: false,
		},
		{
			name:  "case 2: Outputs transform",
			input: `The output path is {{ .Outputs.path.to.output }}`,
			val: map[string]string{
				"Outputs": "./custom/root",
			},
			want:    `The output path is ./custom/root/path/to/output`,
			wantErr: false,
		},
		{
			name:  "case 3: empty input",
			input: "",
			val: map[string]string{
				"Name": "World",
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewTemplate(NewFuncMap(nil)).Render(tt.input, tt.val)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("RenderTemplate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("RenderTemplate() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("RenderTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunction(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		input   string
		val     any
		want    string
		wantErr bool
	}{
		{
			name:  "toJson function",
			input: `{{ .Name | toJson }}`,
			val: map[string]string{
				"Name": `{"a":1}`,
			},
			want:    `"{\"a\":1}"`,
			wantErr: false,
		},
		{
			name:  "default function",
			input: `{{ .UnknownKey | default "xyz" }}`,
			val: map[string]string{
				"Name": "World",
			},
			want:    `xyz`,
			wantErr: false,
		},
		{
			name:  "render function",
			input: `{{ .Name | render }}`,
			val: map[string]string{
				"Name": "{{ .Test }}",
				"Test": "xyz",
			},
			want:    `xyz`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewTemplate(NewFuncMap(tt.val)).Render(tt.input, tt.val)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("RenderTemplate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("RenderTemplate() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("RenderTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransformOutputs(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		text    string
		want    string
		wantErr bool
	}{
		{
			name:    "case 1: normal",
			text:    "{{ .Outputs.xxx.yyy.zzz }}",
			want:    "{{ .Outputs }}/xxx/yyy/zzz",
			wantErr: false,
		},
		{
			name:    "case 2: normal 2",
			text:    "{{.Outputs.a.b}}",
			want:    "{{ .Outputs }}/a/b",
			wantErr: false,
		},
		{
			name:    "case 3: single case",
			text:    "{{ .Outputs.single }}",
			want:    "{{ .Outputs }}/single",
			wantErr: false,
		},
		{
			name:    "case 4: no path",
			text:    "{{ .Outputs }}",
			want:    "{{ .Outputs }}",
			wantErr: false,
		},
		{
			name:    "case 5: other",
			text:    "{{ .Other.xxx }}",
			want:    "{{ .Other.xxx }}",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := transformOutputs(tt.text)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("transformOutputs() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("transformOutputs() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("transformOutputs() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockTaskDir struct {
	ephemeralWorkDir, rootDirPath,
	inputDirPath, outputDirPath,
	logDirPath, workDirPath string
}

func (td *mockTaskDir) GetEphemeralWorkDirPath() string {
	return td.ephemeralWorkDir
}

func (td *mockTaskDir) GetRootDirPath() string {
	return td.rootDirPath
}

func (td *mockTaskDir) GetInputDirPath() string {
	return td.inputDirPath
}

func (td *mockTaskDir) GetOutputDirPath() string {
	return td.outputDirPath
}

func (td *mockTaskDir) GetLogDirPath() string {
	return td.logDirPath
}

func (td *mockTaskDir) GetWorkDirPath() string {
	return td.workDirPath
}

type mockRuntime struct {
	toolAlias         map[string]string
	capabilityBinding map[string]string
	tools             types.Dict
	service           types.Dict
}

func (r *mockRuntime) GetToolAlias() map[string]string {
	return r.toolAlias
}

func (r *mockRuntime) GetCapabilityBinding() map[string]string {
	return r.capabilityBinding
}

func (r *mockRuntime) GetTools() types.Dict {
	return r.tools
}

func (r *mockRuntime) GetService() types.Dict {
	return r.service
}

func TestTemplateParam(t *testing.T) {
	type args struct {
		params  map[string]any
		store   map[string]any
		stores  map[string]types.Dict
		ws      TaskDirs
		target  string
		runtime Runtime
	}
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		args args
		want *TemplateParams
	}{
		{
			name: "normal",
			args: args{
				params: map[string]any{"key": "value"},
				store:  map[string]any{"storeKey": "storeValue"},
				stores: map[string]types.Dict{
					"store1": {"key1": "value1"},
					"store2": {"key2": "value2"},
				},
				ws:     &mockTaskDir{outputDirPath: "/path/to/output"},
				target: "targetValue",
				runtime: &mockRuntime{
					tools:   map[string]any{"aliasKey": "aliasValue"},
					service: map[string]any{"bindingKey": "bindingValue"},
				},
			},
			want: &TemplateParams{
				Params: map[string]any{"key": "value"},
				Store:  map[string]any{"storeKey": "storeValue"},
				Stores: map[string]types.Dict{
					"store1": {"key1": "value1"},
					"store2": {"key2": "value2"},
				},
				Target:  "targetValue",
				Outputs: "/path/to/output",
				Paths: map[string]string{
					"cwd":        func() string { dir, _ := os.Getwd(); return dir }(),
					"workspace":  "",
					"work_dir":   "",
					"input_dir":  "",
					"output_dir": "/path/to/output",
					"log_dir":    "",
				},
				Runtime: types.Dict{
					"tool":    types.Dict{"aliasKey": "aliasValue"},
					"service": types.Dict{"bindingKey": "bindingValue"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTemplateParams(
				WithTaskDirs(tt.args.ws),
				WithParams(tt.args.params),
				WithStore(tt.args.store),
				WithStores(tt.args.stores),
				WithTarget(tt.args.target),
				WithRuntime(tt.args.runtime),
			)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewTemplateParams() mismatch (-want +got):\n%s", diff)
				//				t.Errorf("TransformOutputs() = %v, want %v", got, tt.want)
			}
		})
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package types

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDict_Get(t *testing.T) {
	type args struct {
		keys []string
	}
	defaultDict := Dict{
		"a": map[string]any{
			"a1": map[string]any{
				"a2": "test",
			},
		},
		"b": "true",
	}
	tests := []struct {
		name  string
		dict  *Dict
		args  args
		want  any
		want1 bool
	}{
		{
			name: "simple1",
			dict: &defaultDict,
			args: args{
				keys: []string{"a"},
			},
			want: map[string]any{
				"a1": map[string]any{
					"a2": "test",
				},
			},
			want1: true,
		},
		{
			name: "simple2",
			dict: &defaultDict,
			args: args{
				keys: []string{"a", "a1"},
			},
			want: map[string]any{
				"a2": "test",
			},
			want1: true,
		},
		{
			name: "simple3",
			dict: &defaultDict,
			args: args{
				keys: []string{"a", "a1", "a2"},
			},
			want:  "test",
			want1: true,
		},
		{
			name: "simple4",
			dict: &defaultDict,
			args: args{
				keys: []string{"b"},
			},
			want:  "true",
			want1: true,
		},
		{
			name: "not found1",
			dict: &defaultDict,
			args: args{
				keys: []string{"unknown key"},
			},
			want:  nil,
			want1: false,
		},
		{
			name: "not found2",
			dict: &defaultDict,
			args: args{
				keys: []string{"a", "a2", "a1"},
			},
			want:  nil,
			want1: false,
		},
		{
			name: "not found3",
			dict: &defaultDict,
			args: args{
				keys: []string{},
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.dict.Get(tt.args.keys...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dict.Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Dict.Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		a       Dict
		b       Dict
		want    Dict
		wantErr error
	}{
		{
			name: "success",
			a: Dict{
				"a": 1,
				"b": 2,
			},
			b: Dict{
				"b": 3,
				"c": 4,
			},
			want: Dict{
				"a": 1,
				"b": 3,
				"c": 4,
			},
		},
		{
			name: "success: dict value",
			a: Dict{
				"a": 1,
				"b": 3,
			},
			b: Dict{
				"b": Dict{
					"x": 10,
					"y": 20,
				},
				"c": 4,
			},
			want: Dict{
				"a": 1,
				"b": Dict{
					"x": 10,
					"y": 20,
				},
				"c": 4,
			},
		},
		{
			name: "success: nested dict value",
			a: Dict{
				"a": Dict{
					"x": 10,
					"y": 20,
				},
			},
			b: Dict{
				"a": Dict{
					"y": 30,
				},
				"c": 4,
			},
			want: Dict{
				"a": Dict{
					"x": 10,
					"y": 30,
				},
				"c": 4,
			},
		},
		{
			name: "success: array value(override)",
			a: Dict{
				"a": []int{1, 2, 3},
			},
			b: Dict{
				"a": []int{4, 5},
				"c": 4,
			},
			want: Dict{
				"a": []int{4, 5},
				"c": 4,
			},
		},

		{
			name: "error: both nil",
			a:    nil,
			b:    nil,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := Merge(tt.a, tt.b)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge() = %v, want %v", got, tt.want)
			}
			if gotErr != tt.wantErr {
				t.Errorf("Merge() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestDict_Clone(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		a    Dict
		want Dict
	}{
		{
			name: "simple 1: string value",
			a: Dict{
				"key1": "value1",
			},
			want: Dict{
				"key1": "value1",
			},
		},
		{
			name: "simple 2: array value",
			a: Dict{
				"key1": []string{"value1", "value2"},
			},
			want: Dict{
				"key1": []string{"value1", "value2"},
			},
		},
		{
			name: "simple 3: map value",
			a: Dict{
				"key1": map[string]any{
					"subkey1": "subvalue1",
					"subkey2": 42,
				},
			},
			want: Dict{
				"key1": map[string]any{
					"subkey1": "subvalue1",
					"subkey2": 42,
				},
			},
		},
		{
			name: "simple 2",
			a: Dict{
				"key1": "value1",
				"key2": map[string]any{
					"subkey1": "subvalue1",
					"subkey2": 42,
				},
				"key3": []string{"listitem1", "listitem2"},
			},
			want: Dict{
				"key1": "value1",
				"key2": map[string]any{
					"subkey1": "subvalue1",
					"subkey2": 42,
				},
				"key3": []string{"listitem1", "listitem2"},
			},
		},
		{
			name: "nil dict",
			a:    nil,
			want: nil,
		},
		{
			name: "empty dict",
			a:    Dict{},
			want: Dict{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Clone()
			if &got == &tt.a {
				t.Errorf("Clone() = %v, got should be a new instance", got)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Clone() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDict_ValidateAsTemplateVar(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		a       Dict
		wantErr bool
	}{
		{
			name: "nested valid dict",
			a: Dict{
				"valid_key": "value",
				"another_valid_key": map[string]any{
					"nested_valid_key": []any{
						map[string]any{
							"deeply_nested_valid_key": "value",
						},
						"just a string",
						42,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nested invalid dict",
			a: Dict{
				"valid_key_here": map[string]any{
					"nested-invalid-key": []any{}, // nd
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.a.ValidateAsTemplateVar()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateAsTemplateVar() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateAsTemplateVar() succeeded unexpectedly")
			}
		})
	}
}

func TestDict_ValidateAsLabelKey(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		a       Dict
		wantErr bool
	}{
		{
			name: "nested valid dict",
			a: Dict{
				"valid_key": "value",
				"another_valid_key": map[string]any{
					"nested_valid_key": []any{
						map[string]any{
							"deeply_nested_valid_key": "value",
						},
						"just a string",
						42,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nested invalid dict",
			a: Dict{
				"valid_key_here": map[string]any{
					"nested-invalid-key": []any{}, // nd
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.a.ValidateAsLabelKey()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateAsLabelKey() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateAsLabelKey() succeeded unexpectedly")
			}
		})
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package node_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/tmpl"
	"github.com/sony/niwashi/internal/types"
)

func TestNodeTemplate_Merge(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		self    *node.Template
		other   *node.Template
		want    *node.Template
		wantErr bool
	}{
		{
			name: "Test case 1",
			self: &node.Template{},
			other: &node.Template{
				Os:     "OverriddenOS",
				From:   "OverriddenFrom",
				Labels: map[string]string{"key1": "value1"},
				Capabilities: cap.CapabilityList{
					"cap1": &cap.Capability{
						Params: types.Params{
							"param1": "value1",
							"param2": map[string]any{
								"subparam1": "subvalue1",
								"subparam2": 42,
							},
							"param3": []string{"listitem1", "listitem2"},
						},
					},
					"cap2": &cap.Capability{},
				},
			},
			want: &node.Template{
				Os:     "OverriddenOS",
				From:   "OverriddenFrom",
				Labels: map[string]string{"key1": "value1"},
				Capabilities: cap.CapabilityList{
					"cap1": &cap.Capability{
						Params: types.Params{
							"param1": "value1",
							"param2": map[string]any{
								"subparam1": "subvalue1",
								"subparam2": 42,
							},
							"param3": []string{"listitem1", "listitem2"},
						},
					},
					"cap2": &cap.Capability{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.self.Merge(tt.other)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Merge() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Merge() succeeded unexpectedly")
			}

			if got == tt.self || got == tt.other {
				t.Errorf("Merge() = %v, got should be a new instance", got)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Merge() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNodeTemplate_Resolve(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		tpl     map[string]*node.Template
		self    *node.Template
		want    *node.Template
		wantErr bool
	}{
		{
			name: "Test case 1",
			tpl: map[string]*node.Template{
				"base": {
					Os: "Base",
					Labels: map[string]string{
						"env": "production",
					},
				},
			},
			self: &node.Template{
				From: "base",
				Labels: map[string]string{
					"role": "webserver",
				},
			},
			want: &node.Template{
				From: "base",
				Os:   "Base",
				Labels: map[string]string{
					"env":  "production",
					"role": "webserver",
				},
			},
			wantErr: false,
		},
		{
			name: "Test case 2: circular reference",
			tpl: map[string]*node.Template{
				"tpl1": {
					From: "tpl2",
				},
				"tpl2": {
					From: "tpl1",
				},
			},
			self: &node.Template{
				From: "tpl1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 3: missing template",
			tpl: map[string]*node.Template{
				"existing": {
					Os: "ExistingOS",
				},
			},
			self: &node.Template{
				From: "nonexistent",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test case 4: multiple inheritance",
			tpl: map[string]*node.Template{
				"base": {
					Os: "BaseOS",
					Labels: map[string]string{
						"tier": "backend",
						"tpl":  "base",
					},
				},
				"intermediate": {
					From: "base",
					Labels: map[string]string{
						"env": "staging",
						"tpl": "intermediate",
					},
				},
			},
			self: &node.Template{
				From: "intermediate",
				Labels: map[string]string{
					"role": "database",
					"tpl":  "self",
				},
			},
			want: &node.Template{
				From: "intermediate",
				Os:   "BaseOS",
				Labels: map[string]string{
					"tier": "backend",
					"env":  "staging",
					"role": "database",
					"tpl":  "self", // self overrides all previous
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tmpl.Resolve(tt.tpl, tt.self) // --- IGNORE ---
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Resolve() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Resolve() succeeded unexpectedly")
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Resolve() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNode_ApplyTemplate(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		tpl     map[string]*node.Template
		self    *node.Node
		wantErr bool
		want    *node.Node
	}{
		{
			name: "Not found template",
			self: &node.Node{
				Templates: []string{"nonexistent"},
			},
			tpl:     map[string]*node.Template{},
			wantErr: true,
		},
		{
			name: "Apply single template",
			self: &node.Node{
				Templates: []string{"tpl1"},
			},
			tpl: map[string]*node.Template{
				"tpl1": {
					Os: "TemplateOS",
					Labels: map[string]string{
						"key1": "value1",
					},
					Capabilities: cap.CapabilityList{
						"cap1": &cap.Capability{},
					},
				},
			},
			wantErr: false,
			want: &node.Node{
				Labels: map[string]string{
					"key1": "value1",
				},
				Capabilities: cap.CapabilityList{
					"cap1": &cap.Capability{},
				},
				Templates: []string{}, // cleared after apply
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.self.ApplyTemplate(tt.tpl)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ApplyTemplate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ApplyTemplate() succeeded unexpectedly")
			}

			if diff := cmp.Diff(tt.want, tt.self); diff != "" {
				t.Errorf("Resolve() mismatch (-want +tt.self):\n%s", diff)
			}
		})
	}
}

func TestNode_Validate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		node    *node.Node
		wantErr bool
	}{
		{
			name:    "empty node",
			node:    &node.Node{},
			wantErr: false,
		},
		// { // todo: add test case for id
		// 	name: "invalid capability id",
		// 	node: &node.Node{
		// 		Capabilities: cap.CapabilityList{
		// 			"invalid-id": &cap.Capability{},
		// 		},
		// 	},
		// 	wantErr: true,
		// },
		{
			name: "invalid capability store",
			node: &node.Node{
				Capabilities: cap.CapabilityList{
					"cap1": &cap.Capability{
						Store: map[string]any{
							"invalid-name": "value", // ng
							"invalid.name": "value", // ng
							"invalid name": "value", // ng
							"0":            "value", // ng
							".":            "value", // ng
							"a":            "value", // ok
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid label key",
			node: &node.Node{
				Labels: map[string]string{
					"invalid-key": "value", // ng
					"invalid.key": "value", // ng
					"invalid key": "value", // ng
					"0":           "value", // ng
					".":           "value", // ng
					"a":           "value", // ok
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.node.Validate()
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

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state_test

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/types"
)

func TestCapability_Merge(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		self    *cap.Capability
		other   *cap.Capability
		want    *cap.Capability
		wantErr bool
	}{
		{
			name: "normal 1",
			self: nil,
			other: &cap.Capability{
				Params: types.Params{
					"key1": "value1",
					"key2": "value2",
				},
			},
			want: &cap.Capability{
				Params: types.Params{
					"key1": "value1",
					"key2": "value2",
				},
			},
			wantErr: false,
		},
		{
			name: "normal 2",
			self: &cap.Capability{
				Params: types.Params{
					"key1": "value1",
					"key2": "value2",
				},
			},
			other: nil,
			want: &cap.Capability{
				Params: types.Params{
					"key1": "value1",
					"key2": "value2",
				},
			},
			wantErr: false,
		},
		{
			name: "normal 3",
			self: &cap.Capability{
				Params: types.Params{
					"key1": "value1",
					"key2": "value2",
				},
			},
			other: &cap.Capability{
				Params: types.Params{
					"key2": map[string]any{
						"subkey1": "subvalue1",
					},
				},
			},
			want: &cap.Capability{
				Params: types.Params{
					"key1": "value1",
					"key2": map[string]any{
						"subkey1": "subvalue1",
					},
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge() = %v, want %v", got, tt.want)
			}
		})
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package profile_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/profile"
	"github.com/sony/niwashi/internal/types"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		p       string
		want    *profile.Profile
		wantErr bool
	}{
		{
			name:    "empty",
			p:       "testdata/empty.yaml",
			want:    &profile.Profile{},
			wantErr: true,
		},
		{
			name:    "compact",
			p:       "testdata/compact.yaml",
			want:    profile.NewProfile(),
			wantErr: false,
		},
		{
			name: "fullset",
			p:    "testdata/fullset.yaml",
			want: &profile.Profile{
				Version:           profile.Version,
				Name:              "test",
				ToolAlias:         map[string]string{"tool1": "alias1"},
				CapabilityBinding: map[string]string{"cap1": "binding1"},
				Params: &profile.Params{
					Clusters: map[string]types.Dict{
						"cluster1": {
							"key1": "value1",
							"key2": types.Dict{
								"key2-1": "value2-1",
								"key2-2": "value2-2",
							},
						},
					},
					Capabilities: map[string]types.Dict{
						"cap1": {
							"key2": "value2",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := profile.Load(tt.p, file.NewDefaultFileSystem())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Load() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Load() succeeded unexpectedly")
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Load() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

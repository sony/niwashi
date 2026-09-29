// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package profile_test

import (
	"testing"

	"github.com/sony/niwashi/internal/profile"
	"github.com/sony/niwashi/internal/types"
)

func TestProfile_Validate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		profile *profile.Profile
		wantErr bool
	}{
		{
			name:    "empty",
			profile: profile.NewProfile(),
			wantErr: false,
		},
		{
			name: "invalid capability params",
			profile: &profile.Profile{
				Params: &profile.Params{
					Capabilities: map[string]types.Dict{
						"cap1": {
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
			name: "invalid cluster params",
			profile: &profile.Profile{
				Params: &profile.Params{
					Clusters: map[string]types.Dict{
						"cap2": {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.profile.Validate()
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

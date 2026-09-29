// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package instance_test

import (
	"testing"

	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/platform"
	"github.com/sony/niwashi/internal/types"
)

func TestInstance_Validate(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		instance *instance.Instance
		wantErr  bool
	}{
		{
			name:     "empty instance",
			instance: &instance.Instance{},
			wantErr:  false,
		},
		{
			name: "invalid system info",
			instance: &instance.Instance{
				SystemInfo: &platform.SystemInfo{
					Data: types.Dict{
						"invalid-key": "value", // ng
						"invalid.key": "value", // ng
						"invalid key": "value", // ng
						"0":           "value", // ng
						".":           "value", // ng
						"a":           "value", // ok
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.instance.Validate()
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

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_remote_test

import (
	"testing"

	"github.com/sony/niwashi/internal/action/exec_remote"
)

func TestSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		spec    *exec_remote.Spec
		wantErr bool
	}{
		{
			name:    "empty",
			spec:    &exec_remote.Spec{},
			wantErr: false,
		},
		{
			name: "invalid env name",
			spec: &exec_remote.Spec{
				EnvTemplate: map[string]string{
					"":             "value", // ng
					"invalid-name": "value", // ng
					"invalid.name": "value", // ng
					"invalid name": "value", // ng
					"0":            "value", // ng
					".":            "value", // ng
					"a":            "value", // ok
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotErr := tt.spec.Validate()
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

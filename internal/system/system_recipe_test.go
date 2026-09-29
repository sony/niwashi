// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package system_test

import (
	"slices"
	"testing"

	"github.com/sony/niwashi/internal/system"
)

func TestGetSystemRecipeNames(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want []string
	}{
		{
			name: "default",
			want: []string{
				"external-instance",
				"external-instance-from-ssh_config",
				"external-instance-from-file",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.GetSystemRecipeNames()
			if len(got) != len(tt.want) {
				t.Errorf("GetSystemRecipeNames() = %v, want %v", got, tt.want)
			}

			for _, v := range got {
				if !slices.Contains(tt.want, v) {
					t.Errorf("%v is not in want list %v", v, tt.want)
				}
			}
		})
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package platform_test

import (
	"testing"

	"github.com/sony/niwashi/internal/platform"
)

func TestNormalizeArch(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		arch string
		want string
	}{
		{
			name: "no change",
			arch: "amd64",
			want: "amd64",
		},
		{
			name: "x86_64 is normalized to amd64",
			arch: "x86_64",
			want: "amd64",
		},
		{
			name: "AMD64 is normalized to amd64",
			arch: "AMD64",
			want: "amd64",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := platform.NormalizeArch(tt.arch)
			if got != tt.want {
				t.Errorf("NormalizeArch() = %v, want %v", got, tt.want)
			}
		})
	}
}

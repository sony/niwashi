// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package clone_test

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/clone"
)

func TestStringMap(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		m    map[string]string
		want map[string]string
	}{
		{
			name: "case 1: normal",
			m: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clone.StringMap(tt.m)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
